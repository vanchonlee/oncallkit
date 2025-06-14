# Authentication Methods Comparison

## 🔍 So sánh các phương pháp xác thực

### 1. Current System (Email/Password)
```
Endpoint: POST /auth/login
Method: Email + Password → User Session
```

**Ưu điểm:**
- ✅ Đơn giản, dễ hiểu
- ✅ Kiểm soát hoàn toàn
- ✅ Không phụ thuộc third-party
- ✅ Phù hợp cho admin interface

**Nhược điểm:**
- ❌ Không có session management
- ❌ Không có token expiration
- ❌ Khó scale cho multiple clients
- ❌ Không phù hợp cho webhook/API

### 2. Firebase Authentication
```
Endpoint: Firebase SDK → JWT Token → API
Method: OAuth/Email → Firebase → Custom Claims
```

**Ưu điểm:**
- ✅ OAuth providers (Google, GitHub, etc.)
- ✅ JWT tokens với auto-refresh
- ✅ 2FA built-in
- ✅ Email verification
- ✅ Password reset
- ✅ Mobile SDK integration
- ✅ Offline support

**Nhược điểm:**
- ❌ Phụ thuộc Firebase
- ❌ Chi phí khi scale
- ❌ Phức tạp hơn
- ❌ Vendor lock-in

### 3. API Key Authentication
```
Endpoint: POST /alert/webhook?apikey=abc123
Method: API Key → User Attribution → Alert Creation
```

**Ưu điểm:**
- ✅ Perfect cho webhooks/automation
- ✅ User-specific attribution
- ✅ Rate limiting per user
- ✅ Permissions granular
- ✅ Usage analytics
- ✅ Easy integration với external systems
- ✅ Stateless authentication

**Nhược điểm:**
- ❌ Không phù hợp cho interactive users
- ❌ Key management complexity
- ❌ Security risks nếu key bị leak

## 🏗️ Recommended Hybrid Architecture

### Multi-Authentication System
```
┌─────────────────────────────────────────────────────────────┐
│                    Authentication Layer                     │
├─────────────────┬─────────────────┬─────────────────────────┤
│  Web Dashboard  │   Mobile App    │   External Systems      │
│                 │                 │                         │
│ Firebase Auth   │ Firebase Auth   │ API Key Auth            │
│ - OAuth login   │ - Social login  │ - Webhook alerts        │
│ - JWT tokens    │ - Biometric     │ - Monitoring systems    │
│ - 2FA           │ - Push auth     │ - CI/CD pipelines       │
└─────────────────┴─────────────────┴─────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    Unified Auth Middleware                  │
├─────────────────┬─────────────────┬─────────────────────────┤
│ Firebase Token  │ Session Cookie  │ API Key Validation      │
│ Verification    │ Validation      │                         │
│                 │                 │ - Rate limiting         │
│ - JWT decode    │ - Legacy auth   │ - Permission check      │
│ - Custom claims │ - Backward      │ - Usage logging         │
│ - Role mapping  │   compatibility │ - User attribution      │
└─────────────────┴─────────────────┴─────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                      Business Logic                         │
│                                                             │
│ - Alert Management    - User Management    - On-call Logic  │
│ - Service Monitoring  - Dashboard Data    - Notifications  │
└─────────────────────────────────────────────────────────────┘
```

## 📊 Use Case Matrix

| Use Case | Current Auth | Firebase Auth | API Key Auth | Recommended |
|----------|-------------|---------------|--------------|-------------|
| **Web Dashboard Login** | ✅ Works | ✅ Better UX | ❌ Not suitable | Firebase |
| **Mobile App Login** | ⚠️ Basic | ✅ Native | ❌ Not suitable | Firebase |
| **Admin Management** | ✅ Works | ✅ Enhanced | ❌ Not suitable | Firebase |
| **Webhook Alerts** | ❌ No auth | ❌ Complex | ✅ Perfect | API Key |
| **Monitoring Integration** | ❌ No auth | ❌ Overkill | ✅ Perfect | API Key |
| **CI/CD Automation** | ❌ No auth | ❌ Complex | ✅ Perfect | API Key |
| **Third-party Services** | ❌ No auth | ❌ Complex | ✅ Perfect | API Key |
| **Social Login** | ❌ None | ✅ Built-in | ❌ Not suitable | Firebase |
| **2FA Security** | ❌ None | ✅ Built-in | ❌ Not suitable | Firebase |
| **Password Reset** | ❌ Manual | ✅ Automatic | ❌ Not applicable | Firebase |

## 🎯 Implementation Strategy

### Phase 1: API Key System (Immediate Need)
```
Priority: HIGH
Timeline: 2-3 weeks
Reason: Enable webhook authentication for monitoring systems

Implementation:
├── Database schema for API keys
├── API key generation/validation
├── Webhook endpoint with API key auth
├── Basic rate limiting
└── Usage logging
```

### Phase 2: Firebase Integration (Medium Term)
```
Priority: MEDIUM
Timeline: 4-6 weeks
Reason: Improve user experience for web/mobile

Implementation:
├── Firebase project setup
├── OAuth provider configuration
├── JWT token verification middleware
├── Custom claims for roles
└── Gradual migration from current auth
```

### Phase 3: Unified Auth Middleware (Long Term)
```
Priority: LOW
Timeline: 2-3 weeks
Reason: Support multiple auth methods seamlessly

Implementation:
├── Unified authentication middleware
├── Route-specific auth requirements
├── Backward compatibility
└── Performance optimization
```

## 🔧 Technical Implementation

### API Key Middleware
```go
func APIKeyAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        apiKey := c.Query("apikey")
        if apiKey == "" {
            c.JSON(401, gin.H{"error": "API key required"})
            c.Abort()
            return
        }
        
        key, err := ValidateAPIKey(apiKey)
        if err != nil {
            c.JSON(401, gin.H{"error": err.Error()})
            c.Abort()
            return
        }
        
        // Check permissions
        if !HasPermission(key, PermissionCreateAlerts) {
            c.JSON(403, gin.H{"error": "insufficient permissions"})
            c.Abort()
            return
        }
        
        // Rate limiting
        if err := CheckRateLimit(key.ID); err != nil {
            c.JSON(429, gin.H{"error": err.Error()})
            c.Abort()
            return
        }
        
        // Set user context
        c.Set("api_key", key)
        c.Set("user_id", key.UserID)
        c.Next()
        
        // Log usage
        LogAPIKeyUsage(key.ID, c)
    }
}
```

### Unified Auth Middleware
```go
func UnifiedAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Try API key first (for webhooks)
        if apiKey := c.Query("apikey"); apiKey != "" {
            if key, err := ValidateAPIKey(apiKey); err == nil {
                c.Set("auth_method", "api_key")
                c.Set("user_id", key.UserID)
                c.Set("api_key", key)
                c.Next()
                return
            }
        }
        
        // Try Firebase JWT token
        if token := extractBearerToken(c); token != "" {
            if claims, err := verifyFirebaseToken(token); err == nil {
                c.Set("auth_method", "firebase")
                c.Set("user_id", claims.UID)
                c.Set("firebase_claims", claims)
                c.Next()
                return
            }
        }
        
        // Try legacy session (backward compatibility)
        if sessionID := c.GetHeader("X-Session-ID"); sessionID != "" {
            if user, err := validateSession(sessionID); err == nil {
                c.Set("auth_method", "session")
                c.Set("user_id", user.ID)
                c.Set("user", user)
                c.Next()
                return
            }
        }
        
        c.JSON(401, gin.H{"error": "authentication required"})
        c.Abort()
    }
}
```

### Route Configuration
```go
func SetupRoutes(r *gin.Engine) {
    // Public endpoints
    r.POST("/auth/login", authHandler.Login)
    r.POST("/auth/setup-admin", authHandler.SetupAdmin)
    
    // API key protected endpoints
    apiKeyRoutes := r.Group("/alert")
    apiKeyRoutes.Use(APIKeyAuthMiddleware())
    {
        apiKeyRoutes.POST("/webhook", alertHandler.CreateAlertFromWebhook)
    }
    
    // Firebase protected endpoints
    firebaseRoutes := r.Group("/api")
    firebaseRoutes.Use(FirebaseAuthMiddleware())
    {
        firebaseRoutes.GET("/dashboard", dashboardHandler.GetDashboard)
        firebaseRoutes.GET("/alerts", alertHandler.ListAlerts)
    }
    
    // Unified auth endpoints (supports multiple methods)
    unifiedRoutes := r.Group("/v2")
    unifiedRoutes.Use(UnifiedAuthMiddleware())
    {
        unifiedRoutes.GET("/users", userHandler.ListUsers)
        unifiedRoutes.POST("/api-keys", apiKeyHandler.CreateAPIKey)
    }
}
```

## 📈 Migration Path

### Current → API Key (Immediate)
```
1. Implement API key system
2. Create webhook endpoint with API key auth
3. Migrate monitoring systems to use API keys
4. Deprecate unauthenticated webhook endpoints
```

### Current → Firebase (Gradual)
```
1. Setup Firebase project
2. Implement Firebase auth alongside current auth
3. Add Firebase login option to web/mobile
4. Gradually migrate users to Firebase
5. Deprecate current email/password auth
```

### Final State: Hybrid System
```
- Web/Mobile: Firebase Authentication
- Webhooks/APIs: API Key Authentication  
- Admin: Either Firebase or API Key
- Legacy: Backward compatibility during transition
```

## 🔐 Security Considerations

### API Key Security
- Store hashed versions in database
- Implement key rotation
- Monitor for unusual usage patterns
- Rate limiting per key
- IP whitelisting option
- Audit logging

### Firebase Security
- Custom claims for role management
- Security rules for data access
- Token validation on backend
- Refresh token rotation
- Multi-factor authentication

### General Security
- HTTPS everywhere
- Request signing for sensitive operations
- Anomaly detection
- Security headers
- CORS configuration
- Input validation

## 💰 Cost Analysis

### Current System: $0/month
- Self-hosted authentication
- No external dependencies
- Full control over costs

### Firebase Auth: ~$0.02/user/month
- Free tier: 50,000 MAU
- Paid tier: $0.0055/verification
- Additional costs for phone auth

### API Key System: ~$5-20/month
- Database storage costs
- Logging infrastructure
- Monitoring tools
- Development time

### Recommendation
Start with API Key system (immediate need), then evaluate Firebase based on user growth and requirements. 