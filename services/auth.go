package services

import (
	"database/sql"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/vanchonlee/oncallkit/db"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	PG         *sql.DB
	Redis      *redis.Client
	JWTService *JWTService
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	User    db.User `json:"user"`
	Token   string  `json:"token,omitempty"`
	Message string  `json:"message"`
}

func NewAuthService(pg *sql.DB, redis *redis.Client) *AuthService {
	jwtService := NewJWTService("") // Use default secret for now
	return &AuthService{
		PG:         pg,
		Redis:      redis,
		JWTService: jwtService,
	}
}

// Login authenticates user with email and password
func (s *AuthService) Login(c *gin.Context) (LoginResponse, error) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return LoginResponse{}, err
	}

	// Get user by email
	var user db.User
	err := s.PG.QueryRow(`
		SELECT id, name, email, COALESCE(phone, '') as phone, role, team, COALESCE(fcm_token, '') as fcm_token, password_hash, is_active, created_at, updated_at 
		FROM users 
		WHERE email = $1 AND is_active = true
	`, req.Email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Phone, &user.Role, &user.Team,
		&user.FCMToken, &user.PasswordHash, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return LoginResponse{}, errors.New("invalid email or password")
		}
		return LoginResponse{}, err
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return LoginResponse{}, errors.New("invalid email or password")
	}

	// Generate JWT token
	token, err := s.JWTService.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return LoginResponse{}, errors.New("failed to generate token")
	}

	response := LoginResponse{
		User:    user,
		Token:   token,
		Message: "Login successful",
	}

	return response, nil
}

// HashPassword creates a bcrypt hash of the password
func (s *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// ChangePassword allows user to change their password
func (s *AuthService) ChangePassword(userID, oldPassword, newPassword string) error {
	// Get current password hash
	var currentHash string
	err := s.PG.QueryRow(`SELECT password_hash FROM users WHERE id = $1`, userID).Scan(&currentHash)
	if err != nil {
		return err
	}

	// Verify old password
	err = bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(oldPassword))
	if err != nil {
		return errors.New("current password is incorrect")
	}

	// Hash new password
	newHash, err := s.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password
	_, err = s.PG.Exec(`UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3`,
		newHash, time.Now(), userID)

	return err
}

// SetupAdminUser creates admin user if not exists
func (s *AuthService) SetupAdminUser() error {
	// Check if admin user already exists
	var count int
	err := s.PG.QueryRow(`SELECT COUNT(*) FROM users WHERE email = 'admin@slar.com'`).Scan(&count)
	if err != nil {
		return err
	}

	// If admin user already exists, return success
	if count > 0 {
		return nil
	}

	// Hash the default password
	hashedPassword, err := s.HashPassword("admin123")
	if err != nil {
		return err
	}

	// Create admin user
	_, err = s.PG.Exec(`
		INSERT INTO users (id, name, email, phone, role, team, fcm_token, password_hash, is_active, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
	`, "admin-user-id-001", "Admin User", "admin@slar.com", "+1234567890", "admin", "System Admin", "", hashedPassword, true)

	return err
}
