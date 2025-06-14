# 🔑 Token Authentication Guide

## 📋 Overview

Hệ thống hiện tại đã được update để support JWT tokens cho authentication. Đây là hướng dẫn chi tiết cách lấy và sử dụng tokens.

## 🚀 Quick Start

### 1. **Lấy Admin Token**

```bash
# Step 1: Setup admin user (nếu chưa có)
curl -X POST http://localhost:8080/auth/setup-admin

# Step 2: Login để lấy token
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@slar.com",
    "password": "admin123"
  }'
```

**Response sẽ có dạng:**
```json
{
  "user": {
    "id": "admin-user-id-001",
    "name": "Admin User",
    "email": "admin@slar.com",
    "role": "admin",
    "team": "System Admin",
    "is_active": true,
    "created_at": "2024-01-15T10:00:00Z",
    "updated_at": "2024-01-15T10:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiYWRtaW4tdXNlci1pZC0wMDEiLCJlbWFpbCI6ImFkbWluQHNsYXIuY29tIiwicm9sZSI6ImFkbWluIiwiZXhwIjoxNzM3MDI0MDAwLCJpYXQiOjE3MzY5Mzc2MDB9.signature_here",
  "message": "Login successful"
}
```

### 2. **Sử dụng Token**

Copy giá trị `token` từ response và sử dụng trong header:

```bash
# Sử dụng token trong API calls
curl -X GET http://localhost:8080/api-keys \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

## 🔧 Token Structure

### JWT Token Format
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiYWRtaW4tdXNlci1pZC0wMDEiLCJlbWFpbCI6ImFkbWluQHNsYXIuY29tIiwicm9sZSI6ImFkbWluIiwiZXhwIjoxNzM3MDI0MDAwLCJpYXQiOjE3MzY5Mzc2MDB9.signature_here
```

### Token Parts
1. **Header**: `{"alg":"HS256","typ":"JWT"}`
2. **Payload**: `{"user_id":"admin-user-id-001","email":"admin@slar.com","role":"admin","exp":1737024000,"iat":1736937600}`
3. **Signature**: HMAC-SHA256 signature

### Token Claims
- `user_id`: User ID trong database
- `email`: Email của user
- `role`: Role của user (admin, engineer, manager)
- `exp`: Expiration timestamp (24 hours)
- `iat`: Issued at timestamp

## 👥 Lấy Token cho Different Users

### Admin User
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@slar.com",
    "password": "admin123"
  }'
```

### Regular User (nếu đã tạo)
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "user_password"
  }'
```

## 🔐 Authentication Methods

### 1. **JWT Token Authentication**
```bash
# For API key management endpoints
curl -X POST http://localhost:8080/api-keys \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My API Key",
    "environment": "prod",
    "permissions": ["create_alerts"]
  }'
```

### 2. **API Key Authentication**
```bash
# For webhook endpoints
curl -X POST "http://localhost:8080/alert/webhook?apikey=slar_prod_abc123" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Alert Title",
    "description": "Alert Description",
    "severity": "critical",
    "source": "monitoring"
  }'
```

## 📝 Step-by-Step Examples

### Example 1: Create API Key
```bash
# 1. Get admin token
ADMIN_TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@slar.com","password":"admin123"}' | \
  jq -r '.token')

# 2. Create API key
curl -X POST http://localhost:8080/api-keys \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production Monitoring",
    "environment": "prod",
    "permissions": ["create_alerts"],
    "rate_limit_per_hour": 1000
  }'
```

### Example 2: Use API Key for Webhook
```bash
# Use the API key from previous step
curl -X POST "http://localhost:8080/alert/webhook?apikey=slar_prod_abc123xyz456" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "High CPU Usage",
    "description": "CPU usage exceeded 90%",
    "severity": "critical",
    "source": "monitoring_system"
  }'
```

## 🛠️ Using with HTTP Clients

### VS Code REST Client
```http
### Get Admin Token
POST http://localhost:8080/auth/login
Content-Type: application/json

{
    "email": "admin@slar.com",
    "password": "admin123"
}

### Use Token (replace with actual token)
@token = eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

### Create API Key
POST http://localhost:8080/api-keys
Authorization: Bearer {{token}}
Content-Type: application/json

{
    "name": "My API Key",
    "environment": "prod",
    "permissions": ["create_alerts"]
}
```

### Postman
1. **Get Token**:
   - Method: POST
   - URL: `http://localhost:8080/auth/login`
   - Body: `{"email":"admin@slar.com","password":"admin123"}`
   - Copy `token` from response

2. **Use Token**:
   - Add Header: `Authorization: Bearer <your_token>`
   - Or use Postman's Auth tab → Bearer Token

### cURL with Variables
```bash
# Set token as environment variable
export ADMIN_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Use in requests
curl -X GET http://localhost:8080/api-keys \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

## ⚠️ Important Notes

### Token Expiration
- Tokens expire after **24 hours**
- You'll need to login again to get a new token
- Check `exp` claim in token payload for exact expiration

### Security
- Never share tokens publicly
- Store tokens securely
- Use HTTPS in production
- Tokens contain user information (not encrypted, just signed)

### Error Handling
Common authentication errors:
- `401 Unauthorized`: Invalid or missing token
- `403 Forbidden`: Valid token but insufficient permissions
- `Token has expired`: Need to login again

## 🔄 Token Refresh

Currently, tokens need to be refreshed by logging in again:

```bash
# When token expires, login again
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@slar.com",
    "password": "admin123"
  }'
```

## 🧪 Testing Tokens

### Validate Token
```bash
# Test if token is valid
curl -X GET http://localhost:8080/users \
  -H "Authorization: Bearer $TOKEN"
```

### Decode Token (for debugging)
```bash
# Decode JWT payload (base64)
echo "eyJ1c2VyX2lkIjoiYWRtaW4tdXNlci1pZC0wMDEi..." | base64 -d
```

## 📊 Token vs API Key Usage

| Use Case | Authentication Method | Example |
|----------|----------------------|---------|
| **Web Dashboard** | JWT Token | `Authorization: Bearer <token>` |
| **API Management** | JWT Token | Create/manage API keys |
| **Webhook Alerts** | API Key | `?apikey=slar_prod_abc123` |
| **External Systems** | API Key | Monitoring tools, CI/CD |
| **Mobile App** | JWT Token | User authentication |

## 🎯 Quick Reference

### Get Admin Token
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@slar.com","password":"admin123"}' | \
  jq -r '.token'
```

### Use Token in Header
```
Authorization: Bearer <your_jwt_token>
```

### API Key in URL
```
?apikey=<your_api_key>
``` 