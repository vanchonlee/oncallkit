# API Testing Guide

## Overview
File `test.http` chứa tất cả test cases cho API endpoints. Sử dụng VS Code REST Client extension để chạy tests.

## Setup

### 1. Install VS Code Extension
- Cài đặt extension "REST Client" trong VS Code
- Hoặc sử dụng curl/Postman với các request tương tự

### 2. Start API Server
```bash
go run cmd/main.go
```

### 3. Run Database Migration
```bash
psql -U slar -d slar -h localhost -f migrations/001_create_alerts.sql
```

## Testing Workflow

### Step 1: Create Users
Chạy các request tạo user trước:
```http
POST http://localhost:8080/users
Content-Type: application/json

{
    "name": "Evgenii Druzhinin",
    "email": "evgenii.druzhinin@company.com",
    "phone": "+1234567890",
    "role": "engineer",
    "team": "Platform Team",
    "fcm_token": "fcm_token_evgenii_123"
}
```

**Response sẽ trả về user ID, copy ID này để dùng cho các bước tiếp theo.**

### Step 2: Create On-Call Schedule
Thay `{{user_id}}` bằng ID thực từ Step 1:
```http
POST http://localhost:8080/oncall/schedules
Content-Type: application/json

{
    "user_id": "PASTE_USER_ID_HERE",
    "start_time": "2025-06-14T00:00:00Z",
    "end_time": "2025-06-15T23:59:59Z"
}
```

### Step 3: Create Alert (Auto-Assignment)
```http
POST http://localhost:8080/alerts
Content-Type: application/json

{
    "title": "[Datadog] [P1] High CPU usage on production server",
    "description": "CPU usage has exceeded 90% for the last 5 minutes on prod-web-01",
    "severity": "critical",
    "source": "datadog"
}
```

**Alert sẽ tự động được gán cho user đang on-call.**

### Step 4: Verify Assignment
```http
GET http://localhost:8080/alerts
```

Check field `assigned_to` trong response.

## API Endpoints Reference

### Alerts
- `GET /alerts` - List all alerts
- `POST /alerts` - Create new alert (auto-assigned)
- `GET /alerts/:id` - Get alert details
- `POST /alerts/:id/ack` - Acknowledge alert
- `POST /alerts/:id/unack` - Un-acknowledge alert
- `POST /alerts/:id/close` - Close alert

### Users
- `GET /users` - List all users
- `POST /users` - Create new user
- `GET /users/:id` - Get user details
- `PUT /users/:id` - Update user
- `DELETE /users/:id` - Delete user (soft delete)

### On-Call Management
- `GET /oncall/current` - Get current on-call user
- `GET /oncall/schedules` - List all schedules
- `POST /oncall/schedules` - Create new schedule

## Sample Data

### User Roles
- `engineer` - Software Engineer
- `senior_engineer` - Senior Engineer
- `manager` - Team Lead/Manager
- `admin` - System Administrator

### Alert Severities
- `critical` - P1, immediate response required
- `high` - P2, response within 1 hour
- `medium` - P3, response within 4 hours
- `low` - P4, response within 24 hours

### Alert Sources
- `datadog` - Datadog monitoring
- `prometheus` - Prometheus alerts
- `grafana` - Grafana alerts
- `manual` - Manually created
- `api` - External API integration

## Testing Scenarios

### Scenario 1: Normal Flow
1. Create user
2. Create on-call schedule
3. Create alert → Auto-assigned
4. Acknowledge alert
5. Close alert

### Scenario 2: No On-Call User
1. Create alert when no one is on-call
2. Alert should be created but not assigned
3. Manually assign later

### Scenario 3: Multiple Users
1. Create multiple users
2. Create overlapping schedules
3. Test which user gets assigned (latest schedule wins)

### Scenario 4: Escalation
1. Create alert
2. Don't acknowledge within 5 minutes
3. Check if alert status becomes "escalated"

## Error Handling Tests

### Invalid Requests
- Missing required fields
- Invalid JSON format
- Non-existent IDs
- Invalid date formats

### Expected Responses
- `400 Bad Request` - Invalid input
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

## Tips

1. **Copy IDs**: Always copy real IDs from responses to use in subsequent requests
2. **Time Zones**: Use UTC format for timestamps
3. **Order Matters**: Create users before schedules, schedules before alerts
4. **Check Logs**: Monitor server logs for worker activity and FCM notifications

## Troubleshooting

### Common Issues
1. **Port 8080 in use**: Kill existing process or change port
2. **Database connection**: Check PostgreSQL is running
3. **Redis connection**: Check Redis is running
4. **Migration not run**: Run SQL migration first

### Debug Commands
```bash
# Check if services are running
lsof -i :8080  # API server
lsof -i :5432  # PostgreSQL
lsof -i :6379  # Redis

# Check database
psql -U slar -d slar -h localhost -c "SELECT * FROM users;"
psql -U slar -d slar -h localhost -c "SELECT * FROM alerts;"
``` 