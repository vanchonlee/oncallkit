package db

import "time"

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone,omitempty"`
	Role      string    `json:"role"` // admin, engineer, manager
	Team      string    `json:"team"` // Platform Team, Backend Team, etc.
	FCMToken  string    `json:"fcm_token,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OnCallSchedule struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type Alert struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Severity    string     `json:"severity"`
	Source      string     `json:"source"`
	AckedBy     string     `json:"acked_by,omitempty"`
	AckedAt     *time.Time `json:"acked_at,omitempty"`
	AssignedTo  string     `json:"assigned_to,omitempty"` // User ID
	AssignedAt  *time.Time `json:"assigned_at,omitempty"`
}
