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

// AlertResponse includes user information for API responses
type AlertResponse struct {
	ID              string     `json:"id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Status          string     `json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	Severity        string     `json:"severity"`
	Source          string     `json:"source"`
	AckedBy         string     `json:"acked_by,omitempty"`
	AckedAt         *time.Time `json:"acked_at,omitempty"`
	AssignedTo      string     `json:"assigned_to,omitempty"`       // User ID
	AssignedToName  string     `json:"assigned_to_name,omitempty"`  // User Name
	AssignedToEmail string     `json:"assigned_to_email,omitempty"` // User Email
	AssignedAt      *time.Time `json:"assigned_at,omitempty"`
}

// Uptime Monitoring Models
type Service struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Type      string    `json:"type"`     // http, https, tcp, ping
	Method    string    `json:"method"`   // GET, POST, HEAD
	Interval  int       `json:"interval"` // Check interval in seconds
	Timeout   int       `json:"timeout"`  // Timeout in seconds
	IsActive  bool      `json:"is_active"`
	IsEnabled bool      `json:"is_enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Expected response
	ExpectedStatus int    `json:"expected_status,omitempty"` // Expected HTTP status code
	ExpectedBody   string `json:"expected_body,omitempty"`   // Expected response body content

	// Headers for HTTP requests
	Headers map[string]string `json:"headers,omitempty"`
}

type ServiceCheck struct {
	ID           string    `json:"id"`
	ServiceID    string    `json:"service_id"`
	Status       string    `json:"status"`        // up, down, timeout, error
	ResponseTime int       `json:"response_time"` // Response time in milliseconds
	StatusCode   int       `json:"status_code,omitempty"`
	ResponseBody string    `json:"response_body,omitempty"`
	ErrorMessage string    `json:"error_message,omitempty"`
	CheckedAt    time.Time `json:"checked_at"`

	// SSL Certificate info (for HTTPS)
	SSLExpiry   *time.Time `json:"ssl_expiry,omitempty"`
	SSLIssuer   string     `json:"ssl_issuer,omitempty"`
	SSLDaysLeft int        `json:"ssl_days_left,omitempty"`
}

type UptimeStats struct {
	ServiceID        string    `json:"service_id"`
	Period           string    `json:"period"` // 1h, 24h, 7d, 30d
	UptimePercentage float64   `json:"uptime_percentage"`
	TotalChecks      int       `json:"total_checks"`
	SuccessfulChecks int       `json:"successful_checks"`
	FailedChecks     int       `json:"failed_checks"`
	AvgResponseTime  float64   `json:"avg_response_time"`
	MinResponseTime  int       `json:"min_response_time"`
	MaxResponseTime  int       `json:"max_response_time"`
	LastUpdated      time.Time `json:"last_updated"`
}

type ServiceIncident struct {
	ID          string     `json:"id"`
	ServiceID   string     `json:"service_id"`
	Type        string     `json:"type"`   // downtime, slow_response, ssl_expiry
	Status      string     `json:"status"` // ongoing, resolved
	StartedAt   time.Time  `json:"started_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	Duration    int        `json:"duration,omitempty"` // Duration in seconds
	Description string     `json:"description"`
	AlertID     string     `json:"alert_id,omitempty"` // Related alert ID
}
