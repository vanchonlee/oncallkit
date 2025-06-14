package db

import "time"

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone,omitempty"`
	Role         string    `json:"role"` // admin, engineer, manager
	Team         string    `json:"team"` // Platform Team, Backend Team, etc.
	FCMToken     string    `json:"fcm_token,omitempty"`
	PasswordHash string    `json:"-"` // Don't expose password hash in JSON
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
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

// API Key Authentication Models
type APIKey struct {
	ID                 string     `json:"id"`
	UserID             string     `json:"user_id"`
	Name               string     `json:"name"`
	APIKey             string     `json:"api_key,omitempty"` // Only shown during creation
	APIKeyHash         string     `json:"-"`                 // Never expose hash
	Permissions        []string   `json:"permissions"`
	IsActive           bool       `json:"is_active"`
	LastUsedAt         *time.Time `json:"last_used_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	ExpiresAt          *time.Time `json:"expires_at,omitempty"`
	RateLimitPerHour   int        `json:"rate_limit_per_hour"`
	RateLimitPerDay    int        `json:"rate_limit_per_day"`
	TotalRequests      int        `json:"total_requests"`
	TotalAlertsCreated int        `json:"total_alerts_created"`
	Description        string     `json:"description"`
	Environment        string     `json:"environment"` // prod, dev, test
	CreatedBy          string     `json:"created_by,omitempty"`
}

type APIKeyUsageLog struct {
	ID             string    `json:"id"`
	APIKeyID       string    `json:"api_key_id"`
	Endpoint       string    `json:"endpoint"`
	Method         string    `json:"method"`
	IPAddress      string    `json:"ip_address,omitempty"`
	UserAgent      string    `json:"user_agent,omitempty"`
	RequestSize    int       `json:"request_size"`
	ResponseStatus int       `json:"response_status"`
	ResponseTimeMs int       `json:"response_time_ms"`
	CreatedAt      time.Time `json:"created_at"`
	AlertID        string    `json:"alert_id,omitempty"`
	AlertTitle     string    `json:"alert_title,omitempty"`
	AlertSeverity  string    `json:"alert_severity,omitempty"`
	RequestID      string    `json:"request_id,omitempty"`
	ErrorMessage   string    `json:"error_message,omitempty"`
}

type APIKeyRateLimit struct {
	ID           string    `json:"id"`
	APIKeyID     string    `json:"api_key_id"`
	WindowStart  time.Time `json:"window_start"`
	WindowType   string    `json:"window_type"` // hour, day
	RequestCount int       `json:"request_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// API Key Statistics (from view)
type APIKeyStats struct {
	ID                 string     `json:"id"`
	Name               string     `json:"name"`
	UserID             string     `json:"user_id"`
	UserName           string     `json:"user_name"`
	UserEmail          string     `json:"user_email"`
	Environment        string     `json:"environment"`
	IsActive           bool       `json:"is_active"`
	CreatedAt          time.Time  `json:"created_at"`
	LastUsedAt         *time.Time `json:"last_used_at,omitempty"`
	TotalRequests      int        `json:"total_requests"`
	TotalAlertsCreated int        `json:"total_alerts_created"`
	RateLimitPerHour   int        `json:"rate_limit_per_hour"`
	RateLimitPerDay    int        `json:"rate_limit_per_day"`
	RequestsLast24h    int        `json:"requests_last_24h"`
	AlertsLast24h      int        `json:"alerts_last_24h"`
	ErrorsLast24h      int        `json:"errors_last_24h"`
	AvgResponseTimeMs  float64    `json:"avg_response_time_ms"`
	Status             string     `json:"status"` // active, disabled, expired
}

// Request/Response DTOs
type CreateAPIKeyRequest struct {
	Name             string     `json:"name" binding:"required"`
	Description      string     `json:"description"`
	Environment      string     `json:"environment" binding:"required,oneof=prod dev test"`
	Permissions      []string   `json:"permissions" binding:"required"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
	RateLimitPerHour int        `json:"rate_limit_per_hour,omitempty"`
	RateLimitPerDay  int        `json:"rate_limit_per_day,omitempty"`
}

type CreateAPIKeyResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	APIKey      string     `json:"api_key"` // Only shown once
	Environment string     `json:"environment"`
	Permissions []string   `json:"permissions"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	Message     string     `json:"message"`
}

type UpdateAPIKeyRequest struct {
	Name             *string    `json:"name,omitempty"`
	Description      *string    `json:"description,omitempty"`
	IsActive         *bool      `json:"is_active,omitempty"`
	Permissions      []string   `json:"permissions,omitempty"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
	RateLimitPerHour *int       `json:"rate_limit_per_hour,omitempty"`
	RateLimitPerDay  *int       `json:"rate_limit_per_day,omitempty"`
}

type WebhookAlertRequest struct {
	Title       string                 `json:"title" binding:"required"`
	Description string                 `json:"description" binding:"required"`
	Severity    string                 `json:"severity" binding:"required,oneof=low medium high critical"`
	Source      string                 `json:"source" binding:"required"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type WebhookAlertResponse struct {
	AlertID    string `json:"alert_id"`
	Status     string `json:"status"`
	AssignedTo string `json:"assigned_to,omitempty"`
	Message    string `json:"message"`
}

// Permission constants
type Permission string

const (
	PermissionCreateAlerts   Permission = "create_alerts"
	PermissionReadAlerts     Permission = "read_alerts"
	PermissionManageOnCall   Permission = "manage_oncall"
	PermissionViewDashboard  Permission = "view_dashboard"
	PermissionManageServices Permission = "manage_services"
)

// Valid permissions list
var ValidPermissions = []Permission{
	PermissionCreateAlerts,
	PermissionReadAlerts,
	PermissionManageOnCall,
	PermissionViewDashboard,
	PermissionManageServices,
}

// Environment constants
const (
	EnvironmentProd = "prod"
	EnvironmentDev  = "dev"
	EnvironmentTest = "test"
)

// Rate limit window types
const (
	WindowTypeHour = "hour"
	WindowTypeDay  = "day"
)
