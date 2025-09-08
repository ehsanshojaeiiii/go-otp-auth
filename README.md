# Go OTP Auth - Modern Authentication Service

A secure and scalable backend service implementing OTP-based authentication and user management, built with Go, PostgreSQL, and Redis.

## Features

- 🔐 OTP-based authentication (SMS-less, console logging)
- 🚦 Rate limiting (3 OTP requests per phone per 10 minutes)
- 🔑 JWT token-based session management
- 👥 User management with pagination and search
- 📊 RESTful API with Swagger documentation
- 🐳 Fully containerized with Docker
- 🏗️ Clean architecture implementation
- 🔍 Phone number validation (E.164 format)
- ⚡ Redis for OTP storage and rate limiting
- 🗃️ PostgreSQL for user data persistence

## Tech Stack

- **Language**: Go 1.25.0
- **Framework**: Fiber v2.52.9
- **Database**: PostgreSQL 15
- **Cache/Storage**: Redis 7
- **Authentication**: JWT
- **Documentation**: Swagger/OpenAPI
- **Containerization**: Docker & Docker Compose

## Architecture

```
├── cmd/                    # Application entry point
├── internal/
│   ├── config/            # Configuration management
│   ├── handler/           # HTTP handlers (controllers)
│   ├── service/           # Business logic
│   ├── repository/        # Data access layer
│   ├── model/             # Data models and DTOs
│   └── middleware/        # HTTP middleware
├── pkg/                   # Reusable packages
│   ├── jwt/               # JWT utilities
│   └── utils/             # General utilities
└── docs/                  # API documentation
```

## Database Choice Justification

**PostgreSQL + Redis Combination:**

- **PostgreSQL**: Perfect for persistent user data with ACID compliance, excellent for relational data and provides strong consistency for user records
- **Redis**: Ideal for temporary OTP storage with TTL support, fast in-memory operations for rate limiting, and automatic expiration handling
- **Benefits**: Best of both worlds - reliability for important data, speed for temporary data

## Quick Start

### Prerequisites

- Docker and Docker Compose installed
- Go 1.21+ (for local development)
- Make (optional, for convenience commands)

### 1. Clone and Setup

```bash
git clone <repository-url>
cd golang-test-dekamond
```

### 2. Run with Docker (Recommended)

```bash
# Start all services (PostgreSQL, Redis, and the application)
docker-compose up -d

# View logs
docker-compose logs -f app
```

The API will be available at `http://localhost:8080`

### 3. Run Locally (Development)

```bash
# Start databases only
docker-compose up -d postgres redis

# Install dependencies
go mod download

# Run the application
go run cmd/main.go
```

## API Documentation

Once the service is running, access the Swagger documentation at:
- **Swagger UI**: http://localhost:8080/swagger/index.html

## API Endpoints

### Authentication
- `POST /api/v1/auth/send-otp` - Send OTP to phone number
- `POST /api/v1/auth/verify-otp` - Verify OTP and get JWT token

### User Management (Requires Authentication)
- `GET /api/v1/users/profile` - Get current user profile
- `GET /api/v1/users` - Get paginated list of users with search
- `GET /api/v1/users/{id}` - Get specific user by ID

### Health Check
- `GET /health` - Service health status

## Example Usage

### 1. Send OTP

```bash
curl -X POST http://localhost:8080/api/v1/auth/send-otp \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "+1234567890"}'
```

**Response:**
```json
{
  "message": "OTP sent successfully"
}
```

**Console Output:**
```
OTP for +1234567890: 123456
```

### 2. Verify OTP

```bash
curl -X POST http://localhost:8080/api/v1/auth/verify-otp \
  -H "Content-Type: application/json" \
  -d '{
    "phone_number": "+1234567890",
    "otp_code": "123456"
  }'
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "phone_number": "+1234567890",
    "registered_at": "2024-01-15T10:30:00Z"
  }
}
```

### 3. Get Users (with Authentication)

```bash
curl -X GET "http://localhost:8080/api/v1/users?page=1&page_size=10" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Response:**
```json
{
  "users": [
    {
      "id": 1,
      "phone_number": "+1234567890",
      "registered_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 10,
  "total_pages": 1
}
```

## Configuration

Environment variables can be set in `.env` file (copy from `.env.example`):

```env
# Server
SERVER_HOST=localhost
SERVER_PORT=8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=otp_service

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRY_HOURS=24

# OTP
OTP_LENGTH=6
OTP_EXPIRY_MINUTES=2
OTP_MAX_ATTEMPTS=3
OTP_RATE_LIMIT_MINUTES=10
```

## Development Commands

Using Make (recommended):

```bash
# Setup development environment
make dev-setup

# Install dependencies
make deps

# Run locally
make run

# Build application
make build

# Generate Swagger docs
make swagger

# Start databases only
make db-up

# Docker operations
make docker-build
make docker-up
make docker-down
make docker-logs
```

Using Go directly:

```bash
# Run
go run cmd/main.go

# Build
go build -o bin/otp-service cmd/main.go

# Test
go test ./...
```

## Security Features

- **Rate Limiting**: Max 3 OTP requests per phone number per 10 minutes
- **OTP Expiry**: OTP expires after 2 minutes
- **JWT Security**: Secure token-based authentication
- **Input Validation**: Phone number format validation (E.164)
- **Attempt Limiting**: Max 3 verification attempts per OTP

## Error Handling

The API returns consistent error responses:

```json
{
  "error": "error_code",
  "message": "Human readable error message"
}
```

Common error codes:
- `rate_limit_exceeded` - Too many OTP requests
- `invalid_otp` - Wrong OTP code
- `otp_expired` - OTP has expired
- `unauthorized` - Invalid/missing JWT token
- `invalid_phone_number` - Invalid phone format

## Testing

```bash
# Run all tests
make test

# Or with go
go test -v ./...
```

## Security Features

- **🛡️ Multi-Layer Rate Limiting**: 
  - OTP requests: 3 per phone per 10 minutes
  - Global API: 100 requests per IP per minute
  - Verification attempts: 3 per OTP
- **🔒 Timing Attack Prevention**: Constant-time OTP comparison
- **🚫 Input Validation**: Enhanced phone number validation with DoS protection
- **🛡️ Security Headers**: Helmet middleware for XSS/CSRF protection
- **🔑 JWT Security**: HS256 signing with proper token validation
- **📝 Input Sanitization**: All inputs sanitized and length-validated
- **⏱️ Context Timeouts**: All database operations have timeout protection
- **🔄 Graceful Shutdown**: Production-ready server lifecycle management

## Production Considerations

1. **Environment Variables**: Set strong JWT secret and database passwords
2. **HTTPS**: Use TLS/SSL in production
3. **Database**: Use connection pooling and proper indexing
4. **Monitoring**: Add logging and monitoring solutions
5. **Rate Limiting**: Additional rate limiting at API gateway level recommended
6. **SMS Integration**: Replace console logging with actual SMS service
7. **Security**: All security features are production-ready

## Docker Commands

```bash
# Build and run everything
docker-compose up --build

# Run in background
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Clean up volumes
docker-compose down -v
```

## Health Check

```bash
curl http://localhost:8080/health
```

Response:
```json
{"status": "healthy"}
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes with tests
4. Submit a pull request

## License

MIT License
