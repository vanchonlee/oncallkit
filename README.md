# SLAR - Smart Live Alert & Response System

## 📱 Tổng quan

**SLAR** là hệ thống quản lý cảnh báo và phản hồi thông minh, bao gồm:
- **Backend API** (Go + Gin + PostgreSQL + Redis) - Xử lý alerts, user management, on-call scheduling
- **Mobile App** (Flutter) - Giao diện người dùng cho iOS/Android
- **Worker System** - Xử lý FCM notifications và escalation tự động

## 🏗️ Kiến trúc hệ thống

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

## 🚀 Tính năng chính

### 📊 Dashboard & Monitoring
- **Real-time dashboard** với alert statistics
- **Uptime monitoring** cho các services
- **On-call schedule** hiển thị người đang trực
- **Alert trends** và analytics

### 🚨 Alert Management
- **Auto-assignment** alerts cho user đang on-call
- **Multi-level escalation** với Redis TTL
- **FCM push notifications** real-time
- **Alert lifecycle**: New → Acknowledged → Escalated → Closed
- **Severity levels**: Critical, High, Medium, Low

### 👥 User & Team Management
- **User CRUD** với roles (Engineer, Manager, Admin)
- **Team organization** (Platform, Backend, DevOps)
- **On-call scheduling** với time slots
- **FCM token management** cho notifications

### 🔄 Worker System
- **Background processing** với Goroutines
- **Redis queue** cho alert processing
- **Auto-escalation** sau 5 phút nếu không ACK
- **Concurrent processing** nhiều alerts đồng thời

## 📁 Cấu trúc dự án

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
- **Flutter 3.0+** (cho mobile app)

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
# Sử dụng REST Client trong VS Code
# Mở file: api/services/test.http

# Hoặc sử dụng curl
curl -X GET http://localhost:8080/alerts
curl -X POST http://localhost:8080/alerts \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Alert","severity":"high"}'
```

### Sample Data Setup
```bash
# Chạy sample data script
# Mở file: api/services/sample_data.http
# Chạy từng request để tạo users và schedules mẫu
```

## 🗄️ Database Schema

### Core Tables
- **users** - User information và FCM tokens
- **alerts** - Alert data với assignment
- **on_call_schedules** - On-call time slots
- **schema_migrations** - Migration tracking

### Key Relationships
```sql
alerts.assigned_to → users.id
on_call_schedules.user_id → users.id
```

## 🔧 Migration Management

```bash
# Xem trạng thái migrations
./mg.sh status

# Chạy tất cả migrations
./mg.sh up

# Chạy migration cụ thể
./mg.sh up 001_create_alerts

# Tạo migration mới
./mg.sh create add_new_feature

# Reset database (cẩn thận!)
./mg.sh reset
```

## 📱 Mobile App Features

### Screens
- **Dashboard** - Tổng quan alerts và on-call info
- **Incidents List** - Danh sách alerts với filter
- **Incident Detail** - Chi tiết alert với actions
- **Uptime Monitor** - Monitoring services status

### Key Components
- **Real-time updates** với API polling
- **Push notifications** từ FCM
- **Dark theme** UI design
- **Responsive layout** cho tablet/phone

## 🔄 Workflow

### 1. Normal Alert Flow
```
1. Alert được tạo (manual/API/monitoring)
2. Auto-assign cho user đang on-call
3. Push FCM notification
4. User acknowledge trong app
5. User resolve và close alert
```

### 2. Escalation Flow
```
1. Alert được tạo và assigned
2. Push notification gửi đi
3. Nếu không ACK sau 5 phút
4. Alert status → "escalated"
5. Gửi escalation notification
```

## 🚀 Deployment

### Docker Deployment
```bash
# Build và run với Docker
docker-compose up -d

# Scale workers
docker-compose up -d --scale worker=3
```

### Production Considerations
- **Environment variables** cho config
- **SSL/TLS** cho API endpoints
- **Database backup** strategy
- **Monitoring** với Prometheus/Grafana
- **Log aggregation** với ELK stack

## 🤝 Contributing

1. Fork repository
2. Create feature branch
3. Commit changes
4. Push to branch
5. Create Pull Request

## 📄 License

MIT License - xem file LICENSE để biết thêm chi tiết.

## 📞 Support

- **Issues**: GitHub Issues
- **Documentation**: Wiki pages
- **API Docs**: Postman collection trong `/docs`

---

**SLAR** - Keeping your systems monitored and your team responsive! 🚨📱
