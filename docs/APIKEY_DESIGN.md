# API Key Authentication System Design

## 🎯 Overview

Thiết kế hệ thống API Key cho phép users tạo alerts thông qua webhook với authentication dựa trên API key cá nhân.

## 🗄️ Database Schema

### API Keys Table
```sql
CREATE TABLE api_keys (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES users(id),
    name TEXT NOT NULL,                    -- Friendly name: "Production Monitoring", "Dev Alerts"
    api_key TEXT UNIQUE NOT NULL,          -- Generated key: "slar_abc123xyz456def789"
    permissions TEXT[] DEFAULT '{}',       -- ["create_alerts", "read_alerts", "manage_oncall"]
    is_active BOOLEAN DEFAULT true,
    last_used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP,                  -- Optional expiration
    
    -- Rate limiting
    rate_limit_per_hour INTEGER DEFAULT 1000,
    rate_limit_per_day INTEGER DEFAULT 10000,
    
    -- Usage tracking
    total_requests INTEGER DEFAULT 0,
    total_alerts_created INTEGER DEFAULT 0
);

-- Indexes
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_key ON api_keys(api_key);
CREATE INDEX idx_api_keys_active ON api_keys(is_active);
```

### API Key Usage Logs
```sql
CREATE TABLE api_key_usage_logs (
    id TEXT PRIMARY KEY,
    api_key_id TEXT NOT NULL REFERENCES api_keys(id),
    endpoint TEXT NOT NULL,               -- "/alert/webhook"
    method TEXT NOT NULL,                 -- "POST"
    ip_address TEXT,
    user_agent TEXT,
    request_size INTEGER,
    response_status INTEGER,
    response_time_ms INTEGER,
    created_at TIMESTAMP NOT NULL,
    
    -- Alert specific
    alert_id TEXT,                        -- If alert was created
    alert_title TEXT,
    alert_severity TEXT
);

-- Indexes for analytics
CREATE INDEX idx_usage_logs_api_key ON api_key_usage_logs(api_key_id);
CREATE INDEX idx_usage_logs_created_at ON api_key_usage_logs(created_at);
```

## 🔧 API Key Format

### Key Structure
```
Format: slar_{environment}_{random_string}
Examples:
- slar_prod_abc123xyz456def789ghi012
- slar_dev_mno345pqr678stu901vwx234
- slar_test_yz567abc890def123ghi456

Components:
├── Prefix: "slar_" (identifies our service)
├── Environment: "prod", "dev", "test" 
└── Random: 24-character alphanumeric string
```

### Key Generation Algorithm
```go
func GenerateAPIKey(environment string) string {
    const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
    const keyLength = 24
    
    random := make([]byte, keyLength)
    for i := range random {
        random[i] = charset[rand.Intn(len(charset))]
    }
    
    return fmt.Sprintf("slar_%s_%s", environment, string(random))
}
```

## 🛡️ Security Features

### 1. **Key Validation**
```go
func ValidateAPIKey(key string) (*APIKey, error) {
    // Format validation
    if !strings.HasPrefix(key, "slar_") {
        return nil, errors.New("invalid key format")
    }
    
    parts := strings.Split(key, "_")
    if len(parts) != 3 {
        return nil, errors.New("invalid key structure")
    }
    
    // Database lookup
    apiKey, err := GetAPIKeyByKey(key)
    if err != nil {
        return nil, errors.New("invalid key")
    }
    
    // Active check
    if !apiKey.IsActive {
        return nil, errors.New("key is disabled")
    }
    
    // Expiration check
    if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
        return nil, errors.New("key has expired")
    }
    
    return apiKey, nil
}
```

### 2. **Rate Limiting**
```go
func CheckRateLimit(apiKeyID string) error {
    // Check hourly limit
    hourlyCount := GetUsageCount(apiKeyID, time.Hour)
    if hourlyCount >= apiKey.RateLimitPerHour {
        return errors.New("hourly rate limit exceeded")
    }
    
    // Check daily limit
    dailyCount := GetUsageCount(apiKeyID, 24*time.Hour)
    if dailyCount >= apiKey.RateLimitPerDay {
        return errors.New("daily rate limit exceeded")
    }
    
    return nil
}
```

### 3. **Permissions System**
```go
type Permission string

const (
    PermissionCreateAlerts   Permission = "create_alerts"
    PermissionReadAlerts     Permission = "read_alerts"
    PermissionManageOnCall   Permission = "manage_oncall"
    PermissionViewDashboard  Permission = "view_dashboard"
    PermissionManageServices Permission = "manage_services"
)

func HasPermission(apiKey *APIKey, permission Permission) bool {
    for _, p := range apiKey.Permissions {
        if p == string(permission) {
            return true
        }
    }
    return false
}
```

## 🔌 API Endpoints

### 1. **Webhook Endpoint**
```
POST /alert/webhook?apikey={api_key}
Content-Type: application/json

{
    "title": "High CPU Usage",
    "description": "CPU usage exceeded 90%",
    "severity": "critical",
    "source": "monitoring_system",
    "metadata": {
        "server": "prod-web-01",
        "cpu_usage": "95%"
    }
}
```

### 2. **API Key Management**
```
# List user's API keys
GET /api-keys

# Create new API key
POST /api-keys
{
    "name": "Production Monitoring",
    "permissions": ["create_alerts"],
    "expires_at": "2024-12-31T23:59:59Z",
    "rate_limit_per_hour": 500
}

# Get API key details
GET /api-keys/{id}

# Update API key
PUT /api-keys/{id}
{
    "name": "Updated Name",
    "is_active": false
}

# Delete API key
DELETE /api-keys/{id}

# Regenerate API key
POST /api-keys/{id}/regenerate
```

### 3. **Usage Analytics**
```
# Get API key usage stats
GET /api-keys/{id}/usage?period=7d

# Get usage logs
GET /api-keys/{id}/logs?limit=100&offset=0
```

## 🔄 Request Flow

### Webhook Authentication Flow
```
1. Client Request:
   POST /alert/webhook?apikey=slar_prod_abc123xyz456
   
2. API Key Extraction:
   - Extract from query parameter
   - Validate format
   
3. Authentication:
   - Lookup key in database
   - Check if active and not expired
   - Verify permissions
   
4. Rate Limiting:
   - Check hourly/daily limits
   - Update usage counters
   
5. Alert Creation:
   - Create alert with user attribution
   - Log usage
   - Return response
```

### Error Responses
```json
// Invalid API key
{
    "error": "invalid_api_key",
    "message": "The provided API key is invalid or expired",
    "code": 401
}

// Rate limit exceeded
{
    "error": "rate_limit_exceeded", 
    "message": "Hourly rate limit of 1000 requests exceeded",
    "code": 429,
    "retry_after": 3600
}

// Insufficient permissions
{
    "error": "insufficient_permissions",
    "message": "API key does not have 'create_alerts' permission",
    "code": 403
}
```

## 📊 Usage Analytics

### Dashboard Metrics
```
API Key Analytics:
├── Total Requests (24h, 7d, 30d)
├── Success Rate (%)
├── Average Response Time
├── Top Error Types
├── Alerts Created
├── Rate Limit Hits
└── Geographic Distribution
```

### Usage Tracking
```go
type UsageStats struct {
    APIKeyID        string    `json:"api_key_id"`
    Period          string    `json:"period"`          // "1h", "24h", "7d", "30d"
    TotalRequests   int       `json:"total_requests"`
    SuccessRequests int       `json:"success_requests"`
    ErrorRequests   int       `json:"error_requests"`
    AlertsCreated   int       `json:"alerts_created"`
    AvgResponseTime float64   `json:"avg_response_time_ms"`
    RateLimitHits   int       `json:"rate_limit_hits"`
    LastUsed        time.Time `json:"last_used"`
}
```

## 🔐 Security Best Practices

### 1. **Key Storage**
- Store hashed version in database
- Use bcrypt or similar for hashing
- Never log full API keys

### 2. **Transmission Security**
- HTTPS only for all API calls
- Consider API key in header instead of URL
- Implement request signing for sensitive operations

### 3. **Monitoring & Alerting**
```
Security Alerts:
├── Unusual usage patterns
├── Multiple failed authentication attempts
├── API key used from new IP/location
├── Rate limit violations
└── Suspicious request patterns
```

## 🚀 Implementation Phases

### Phase 1: Core API Key System (1-2 weeks)
- Database schema creation
- API key generation/validation
- Basic CRUD operations
- Webhook authentication

### Phase 2: Security & Rate Limiting (1 week)
- Rate limiting implementation
- Permission system
- Usage logging
- Security monitoring

### Phase 3: Analytics & Management (1 week)
- Usage analytics
- Management dashboard
- API key lifecycle management
- Advanced security features

### Phase 4: Integration & Testing (1 week)
- Integration with existing alert system
- Comprehensive testing
- Documentation
- Production deployment

## 📝 Example Usage

### Creating API Key
```bash
# Create API key for monitoring system
curl -X POST http://localhost:8080/api-keys \
  -H "Authorization: Bearer {user_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production Monitoring",
    "permissions": ["create_alerts"],
    "rate_limit_per_hour": 1000
  }'

# Response
{
    "id": "key_123",
    "name": "Production Monitoring", 
    "api_key": "slar_prod_abc123xyz456def789",
    "permissions": ["create_alerts"],
    "created_at": "2024-01-15T10:00:00Z"
}
```

### Using API Key for Alerts
```bash
# Send alert using API key
curl -X POST "http://localhost:8080/alert/webhook?apikey=slar_prod_abc123xyz456" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Database Connection Failed",
    "description": "Unable to connect to primary database",
    "severity": "critical",
    "source": "database_monitor"
  }'

# Response
{
    "alert_id": "alert_789",
    "status": "created",
    "assigned_to": "user_456",
    "message": "Alert created successfully"
}
```

## 🔗 Integration Points

### With Existing Systems
- User management (API keys belong to users)
- Alert system (alerts created via API key)
- On-call system (auto-assignment still works)
- Notification system (FCM notifications)
- Dashboard (show API key usage)

### External Integrations
- Prometheus AlertManager
- Grafana webhooks
- Custom monitoring systems
- CI/CD pipelines
- Third-party services 