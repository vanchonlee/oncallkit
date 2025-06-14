# Admin User Setup Guide

## 🔐 Admin User Information

### Default Admin Credentials
- **Email**: `admin@slar.com`
- **Password**: `admin123`
- **User ID**: `admin-user-id-001`
- **Role**: `admin`
- **Team**: `System Admin`

## 🚀 Setup Instructions

### 1. Run Database Migration
```bash
cd api
./mg.sh up
```

This will:
- Add `password_hash` column to users table
- Create the admin user with hashed password
- Create necessary indexes

### 2. Start API Server
```bash
go run cmd/main.go
```

### 3. Test Admin Login
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@slar.com","password":"admin123"}'
```

Expected response:
```json
{
  "user": {
    "id": "admin-user-id-001",
    "name": "Admin User",
    "email": "admin@slar.com",
    "phone": "+1234567890",
    "role": "admin",
    "team": "System Admin",
    "is_active": true,
    "created_at": "2025-01-XX...",
    "updated_at": "2025-01-XX..."
  },
  "message": "Login successful"
}
```

## 📋 Available Test Files

### 1. `admin_test.http` - Comprehensive Tests
- Complete admin authentication flow
- User management tasks
- System administration
- Error handling scenarios
- Workflow examples

### 2. `quick_admin_test.http` - Quick Tests
- Basic login test
- Password change test
- Simple validation

### 3. `auth_test.http` - Authentication Focus
- All authentication endpoints
- Error cases
- Security testing

## 🔧 Admin Capabilities

### Authentication
- ✅ Login with email/password
- ✅ Change password
- ✅ Secure password hashing (bcrypt)

### User Management
- ✅ Create new users
- ✅ Update user roles
- ✅ View all users
- ✅ Deactivate users

### System Administration
- ✅ View all alerts
- ✅ Manage on-call schedules
- ✅ Create system alerts
- ✅ Monitor uptime services
- ✅ Access dashboard data

## 🛡️ Security Features

### Password Security
- Passwords are hashed using bcrypt
- Password hashes are never exposed in API responses
- Secure password validation

### Authentication Flow
1. User submits email/password
2. System looks up user by email
3. Compares submitted password with stored hash
4. Returns user info (without password) on success

## 🧪 Testing Workflow

### Using VS Code REST Client
1. Install "REST Client" extension
2. Open `admin_test.http` or `quick_admin_test.http`
3. Click "Send Request" above each test

### Using curl
```bash
# Login
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@slar.com","password":"admin123"}'

# Change password
curl -X POST http://localhost:8080/auth/change-password \
  -H "Content-Type: application/json" \
  -d '{"user_id":"admin-user-id-001","old_password":"admin123","new_password":"newpass"}'

# Get user info
curl http://localhost:8080/users/admin-user-id-001
```

## 🔄 Password Management

### Change Admin Password
```bash
curl -X POST http://localhost:8080/auth/change-password \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "admin-user-id-001",
    "old_password": "admin123",
    "new_password": "your-new-secure-password"
  }'
```

### Reset Password (Database)
If you forget the admin password, you can reset it directly in the database:

```sql
-- Generate new hash for password "newpassword123"
-- Use online bcrypt generator or Go code

UPDATE users 
SET password_hash = '$2a$10$NEW_HASH_HERE', updated_at = NOW() 
WHERE id = 'admin-user-id-001';
```

## 🚨 Troubleshooting

### Migration Issues
```bash
# Check migration status
./mg.sh status

# Run specific migration
./mg.sh up 003_add_user_password

# Reset and re-run all migrations
./mg.sh reset
```

### Database Connection
```bash
# Test database connection
./mg.sh check

# Manual connection
psql -U slar -d slar -h localhost
```

### Authentication Errors
- **"invalid email or password"**: Check credentials
- **"user not found"**: Run migration to create admin user
- **Database connection error**: Check PostgreSQL is running

## 📝 Next Steps

1. **Change Default Password**: For security, change the default admin password
2. **Create Additional Users**: Use admin account to create other users
3. **Setup On-Call Schedules**: Configure who's on-call when
4. **Configure Monitoring**: Setup uptime monitoring for your services
5. **Test Alert Flow**: Create test alerts to verify the system works

## 🔗 Related Files

- `migrations/003_add_user_password.sql` - Database migration
- `services/auth.go` - Authentication service
- `handlers/auth.go` - Authentication endpoints
- `admin_test.http` - Comprehensive tests
- `quick_admin_test.http` - Quick tests 