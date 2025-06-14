# SLAR - Smart Live Alert & Response System

## 📱 Overview

**SLAR** is a smart alert management and response system, including:
- **Backend API** (Go + Gin + PostgreSQL + Redis) - Handles alerts, user management, on-call scheduling
- **Mobile App** (Flutter) - User interface for iOS/Android
- **Worker System** - Processes FCM notifications and automatic escalation

## 🏗️ System Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Flutter App   │    │   Backend API   │    │   Worker Pool   │
│                 │    │                 │    │                 │
│ • Dashboard     │◄──►│ • REST API      │◄──►│ • FCM Push      │
│ • Alerts List   │    │ • User Mgmt     │    │ • Escalation    │
│ • Incident Mgmt │    │ • On-call Mgmt  │    │ • Redis Queue   │
│ • Uptime Mon    │    │ • Alert Routing │    │ • Auto-assign   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   Database      │
                    │                 │
                    │ • PostgreSQL    │
                    │ • Redis Cache   │
                    │ • Migration     │
                    └─────────────────┘
```

## 🚀 Key Features

### 📊 Dashboard & Monitoring
- **Real-time dashboard** with alert statistics
- **Uptime monitoring** for services
- **On-call schedule** showing current on-duty personnel
- **Alert trends** and analytics

### 🚨 Alert Management
- **Auto-assignment** of alerts to on-call users
- **Multi-level escalation** with Redis TTL
- **FCM push notifications** in real-time
- **Alert lifecycle**: New → Acknowledged → Escalated → Closed
- **Severity levels**: Critical, High, Medium, Low

### 👥 User & Team Management
- **User CRUD** with roles (Engineer, Manager, Admin)
- **Team organization** (Platform, Backend, DevOps)
- **On-call scheduling** with time slots
- **FCM token management** for notifications

### 🔄 Worker System
- **Background processing** with Goroutines
- **Redis queue** for alert processing
- **Auto-escalation** after 5 minutes if not acknowledged
- **Concurrent processing** of multiple alerts

## 📁 Project Structure

```
slar/
├── api/                    # Backend API (Go)
│   ├── cmd/               # Main application
│   ├── db/                # Database models & connection
│   ├── handlers/          # HTTP request handlers
│   ├── services/          # Business logic
│   ├── workers/           # Background workers
│   ├── router/            # API routing
│   ├── migrations/        # Database migrations
│   ├── mg.sh             # Migration script
│   └── docker-compose.yaml
│
└── slarapp/               # Flutter Mobile App
    ├── lib/
    │   ├── screens/       # UI screens
    │   ├── models/        # Data models
    │   ├── widgets/       # Reusable widgets
    │   └── main.dart
    ├── android/
    ├── ios/
    └── pubspec.yaml
```

## 🛠️ Setup & Installation

### Prerequisites
- **Go 1.21+**
- **PostgreSQL 15+**
- **Redis 7+**
- **Flutter 3.0+** (for mobile app)

### 1. Backend Setup

```bash
# Clone repository
git clone <repository-url>
cd slar/api

# Install dependencies
go mod download

# Setup database
docker-compose up -d postgres redis

# Run migrations
chmod +x mg.sh
./mg.sh up

# Start API server
go run cmd/main.go
```

### 2. Mobile App Setup

```bash
cd ../slarapp

# Install Flutter dependencies
flutter pub get

# Run on device/emulator
flutter run
```

### 3. Environment Configuration

```bash
# Backend (.env)
DB_HOST=localhost
DB_PORT=5432
DB_USER=slar
DB_NAME=slar
DB_PASSWORD=slar
REDIS_URL=localhost:6379

# Flutter (lib/config.dart)
const API_BASE_URL = 'http://localhost:8080';
```

## 📡 API Endpoints

### Alerts
```
GET    /alerts              # List all alerts
POST   /alerts              # Create new alert (auto-assigned)
GET    /alerts/:id          # Get alert details
POST   /alerts/:id/ack      # Acknowledge alert
POST   /alerts/:id/unack    # Un-acknowledge alert
POST   /alerts/:id/close    # Close alert
```

### Users
```
GET    /users               # List all users
POST   /users               # Create new user
GET    /users/:id           # Get user details
PUT    /users/:id           # Update user
DELETE /users/:id           # Delete user (soft delete)
```

### On-Call Management
```
GET    /oncall/current      # Get current on-call user
GET    /oncall/schedules    # List all schedules
POST   /oncall/schedules    # Create new schedule
```

### Dashboard
```
GET    /dashboard           # Dashboard data
GET    /uptime              # Uptime statistics
```

## 🧪 Testing

### Backend API Testing
```bash
# Use REST Client in VS Code
# Open file: api/services/test.http

# Or use curl
curl -X GET http://localhost:8080/alerts
curl -X POST http://localhost:8080/alerts \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Alert","severity":"high"}'
```

### Sample Data Setup
```bash
# Run sample data script
# Open file: api/services/sample_data.http
# Execute each request to create sample users and schedules
```

## 🗄️ Database Schema

### Core Tables
- **users** - User information and FCM tokens
- **alerts** - Alert data with assignment
- **on_call_schedules** - On-call time slots
- **schema_migrations** - Migration tracking

### Key Relationships
```sql
alerts.assigned_to → users.id
on_call_schedules.user_id → users.id
```

## 🔧 Migration Management

```bash
# Check migration status
./mg.sh status

# Run all migrations
./mg.sh up

# Run specific migration
./mg.sh up 001_create_alerts

# Create new migration
./mg.sh create add_new_feature

# Reset database (be careful!)
./mg.sh reset
```

## 📱 Mobile App Features

### Screens
- **Dashboard** - Alert overview and on-call info
- **Incidents List** - Alert list with filters
- **Incident Detail** - Alert details with actions
- **Uptime Monitor** - Service status monitoring

### Key Components
- **Real-time updates** with API polling
- **Push notifications** from FCM
- **Dark theme** UI design
- **Responsive layout** for tablet/phone

## 🔄 Workflow

### 1. Normal Alert Flow
```
1. Alert is created (manual/API/monitoring)
2. Auto-assign to current on-call user
3. Push FCM notification
4. User acknowledges in app
5. User resolves and closes alert
```

### 2. Escalation Flow
```
1. Alert is created and assigned
2. Push notification sent
3. If not acknowledged after 5 minutes
4. Alert status → "escalated"
5. Send escalation notification
```

## 🚀 Deployment

### Docker Deployment
```bash
# Build and run with Docker
docker-compose up -d

# Scale workers
docker-compose up -d --scale worker=3
```

### Production Considerations
- **Environment variables** for configuration
- **SSL/TLS** for API endpoints
- **Database backup** strategy
- **Monitoring** with Prometheus/Grafana
- **Log aggregation** with ELK stack

## 🤝 Contributing

1. Fork repository
2. Create feature branch
3. Commit changes
4. Push to branch
5. Create Pull Request

## 📄 License

MIT License - see LICENSE file for more details.

## 📞 Support

- **Issues**: GitHub Issues
- **Documentation**: Wiki pages
- **API Docs**: Postman collection in `/docs`

---

**SLAR** - Keeping your systems monitored and your team responsive! 🚨📱
