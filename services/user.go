package services

import (
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/vanchonlee/oncallkit/db"
)

type UserService struct {
	PG    *sql.DB
	Redis *redis.Client
}

func NewUserService(pg *sql.DB, redis *redis.Client) *UserService {
	return &UserService{PG: pg, Redis: redis}
}

// User CRUD operations
func (s *UserService) ListUsers() ([]db.User, error) {
	rows, err := s.PG.Query(`SELECT id, name, email, COALESCE(phone, '') as phone, role, team, COALESCE(fcm_token, '') as fcm_token, is_active, created_at, updated_at FROM users WHERE is_active = true ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []db.User
	for rows.Next() {
		var u db.User
		err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.Role, &u.Team, &u.FCMToken, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			continue
		}
		users = append(users, u)
	}
	return users, nil
}

func (s *UserService) GetUser(id string) (db.User, error) {
	var u db.User
	err := s.PG.QueryRow(`SELECT id, name, email, COALESCE(phone, '') as phone, role, team, COALESCE(fcm_token, '') as fcm_token, is_active, created_at, updated_at FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.Role, &u.Team, &u.FCMToken, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (s *UserService) CreateUser(c *gin.Context) (db.User, error) {
	var user db.User
	if err := c.ShouldBindJSON(&user); err != nil {
		return user, err
	}

	user.ID = uuid.New().String()
	user.IsActive = true
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// If password is provided, hash it
	if user.PasswordHash != "" {
		authService := NewAuthService(s.PG, s.Redis)
		hashedPassword, err := authService.HashPassword(user.PasswordHash)
		if err != nil {
			return user, err
		}
		user.PasswordHash = hashedPassword
	}

	// Ensure empty strings for optional fields to avoid NULL issues
	if user.Phone == "" {
		user.Phone = ""
	}
	if user.FCMToken == "" {
		user.FCMToken = ""
	}
	if user.PasswordHash == "" {
		user.PasswordHash = ""
	}

	_, err := s.PG.Exec(`INSERT INTO users (id, name, email, phone, role, team, fcm_token, password_hash, is_active, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		user.ID, user.Name, user.Email, user.Phone, user.Role, user.Team, user.FCMToken, user.PasswordHash, user.IsActive, user.CreatedAt, user.UpdatedAt)

	return user, err
}

func (s *UserService) UpdateUser(id string, c *gin.Context) (db.User, error) {
	var user db.User
	if err := c.ShouldBindJSON(&user); err != nil {
		return user, err
	}

	user.ID = id
	user.UpdatedAt = time.Now()

	// If password is provided, hash it
	var err error
	if user.PasswordHash != "" {
		authService := NewAuthService(s.PG, s.Redis)
		hashedPassword, hashErr := authService.HashPassword(user.PasswordHash)
		if hashErr != nil {
			return user, hashErr
		}
		user.PasswordHash = hashedPassword

		_, err = s.PG.Exec(`UPDATE users SET name=$2, email=$3, phone=$4, role=$5, team=$6, fcm_token=$7, password_hash=$8, updated_at=$9 WHERE id=$1`,
			user.ID, user.Name, user.Email, user.Phone, user.Role, user.Team, user.FCMToken, user.PasswordHash, user.UpdatedAt)
	} else {
		_, err = s.PG.Exec(`UPDATE users SET name=$2, email=$3, phone=$4, role=$5, team=$6, fcm_token=$7, updated_at=$8 WHERE id=$1`,
			user.ID, user.Name, user.Email, user.Phone, user.Role, user.Team, user.FCMToken, user.UpdatedAt)
	}

	return user, err
}

func (s *UserService) DeleteUser(id string) error {
	_, err := s.PG.Exec(`UPDATE users SET is_active = false, updated_at = $1 WHERE id = $2`, time.Now(), id)
	return err
}

// On-call schedule operations
func (s *UserService) GetCurrentOnCallUser() (db.User, error) {
	var u db.User
	now := time.Now()

	err := s.PG.QueryRow(`
		SELECT u.id, u.name, u.email, COALESCE(u.phone, '') as phone, u.role, u.team, COALESCE(u.fcm_token, '') as fcm_token, u.is_active, u.created_at, u.updated_at 
		FROM users u 
		JOIN on_call_schedules ocs ON u.id = ocs.user_id 
		WHERE ocs.start_time <= $1 AND ocs.end_time >= $1 AND ocs.is_active = true AND u.is_active = true
		ORDER BY ocs.start_time DESC 
		LIMIT 1`, now).
		Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.Role, &u.Team, &u.FCMToken, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)

	return u, err
}

func (s *UserService) CreateOnCallSchedule(c *gin.Context) (db.OnCallSchedule, error) {
	var schedule db.OnCallSchedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		return schedule, err
	}

	schedule.ID = uuid.New().String()
	schedule.IsActive = true
	schedule.CreatedAt = time.Now()

	_, err := s.PG.Exec(`INSERT INTO on_call_schedules (id, user_id, start_time, end_time, is_active, created_at) VALUES ($1,$2,$3,$4,$5,$6)`,
		schedule.ID, schedule.UserID, schedule.StartTime, schedule.EndTime, schedule.IsActive, schedule.CreatedAt)

	return schedule, err
}

func (s *UserService) ListOnCallSchedules() ([]db.OnCallSchedule, error) {
	rows, err := s.PG.Query(`SELECT id, user_id, start_time, end_time, is_active, created_at FROM on_call_schedules WHERE is_active = true ORDER BY start_time DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []db.OnCallSchedule
	for rows.Next() {
		var s db.OnCallSchedule
		err := rows.Scan(&s.ID, &s.UserID, &s.StartTime, &s.EndTime, &s.IsActive, &s.CreatedAt)
		if err != nil {
			continue
		}
		schedules = append(schedules, s)
	}
	return schedules, nil
}
