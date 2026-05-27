# Trachios - Order Management System

A robust order management system built with Go, featuring automated order lifecycle tracking, Redis caching, and JWT-based authentication.

## Features

- 🔐 **JWT Authentication** - Secure user authentication with role-based access control
- 📦 **Order Lifecycle Management** - Automated status progression with background tracking
- 🚀 **Redis Caching** - High-performance order caching with automatic invalidation
- 👑 **Admin Dashboard** - Administrative functions for order management
- 🔄 **Concurrent Processing** - Safe handling of concurrent order operations
- 🐳 **Docker Support** - Full containerization with Docker Compose

## Architecture

```
Trachios/
├── cmd/server/          # Application entry point
├── internal/
│   ├── api/
│   │   ├── handler/     # HTTP request handlers
│   │   └── middleware/  # Authentication & authorization middleware
│   ├── config/          # Configuration management
│   ├── domain/          # Business entities (User, Order)
│   ├── repository/      # Data access layer
│   ├── service/         # Business logic layer
│   └── worker/          # Background processing
├── pkg/util/            # Shared utilities
├── test/                # Integration tests
└── configs/             # Configuration files
```

## Quick Start

### Prerequisites

- Go 1.24+
- Docker & Docker Compose
- PostgreSQL 15+
- Redis 7+

### 1. Clone the Repository

```bash
git clone https://github.com/varun-karmikanda/trachios.git
cd trachios
```

### 2. Set Up Environment

```bash
cp .env.example .env
cp configs/config.yaml.example configs/config.yaml
```

Edit `.env` with your configuration:

```env
SERVER_PORT=8080

POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=trachios_user
POSTGRES_PASSWORD=your_secure_password
POSTGRES_DB=trachios_db

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

JWT_SECRET=your_jwt_secret_key_here

ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=admin_secure_password
```

### 3. Run with Docker Compose

```bash
docker-compose up -d
```

### 4. Run Locally (Development)

```bash
# Start PostgreSQL and Redis
docker-compose up -d postgres redis

# Run migrations (if using golang-migrate)
migrate -path internal/repository/migrations -database "postgres://user:password@localhost:5432/db?sslmode=disable" up

# Start the server
go run cmd/server/main.go
```

## API Endpoints

### Base URL
```
http://localhost:8080/api/v1
```

### 1. Health Check

Check if the server is running.

```bash
curl -X GET http://localhost:8080/api/v1/health
```

### 2. User Registration

Register a new user account.

```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "securepassword123"
  }'
```

### 3. User Login

Authenticate and get JWT token.

```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "securepassword123"
  }'
```

Save the returned token for authenticated requests:
```bash
export TOKEN="your_jwt_token_here"
```

### 4. Get Current User

Get the currently authenticated user's information.

```bash
curl -X GET http://localhost:8080/api/v1/user \
  -H "Authorization: Bearer $TOKEN"
```

### 5. Create Order

Create a new order (authenticated users only).

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "item_description": "MacBook Pro 16-inch M3"
  }'
```

### 6. Get My Orders

Retrieve all orders for the authenticated user.

```bash
curl -X GET http://localhost:8080/api/v1/my-orders \
  -H "Authorization: Bearer $TOKEN"
```

### 7. Get Order by ID

Retrieve a specific order by ID (user can only access their own orders, admins can access any).

```bash
curl -X GET http://localhost:8080/api/v1/orders/1 \
  -H "Authorization: Bearer $TOKEN"
```

### 8. Cancel Order

Cancel an order (user can only cancel their own orders).

```bash
curl -X PUT http://localhost:8080/api/v1/orders/1/cancel \
  -H "Authorization: Bearer $TOKEN"
```

### Admin Endpoints

These endpoints require admin privileges.

### 9. Get All Orders (Admin Only)

Retrieve all orders in the system (admin only).

```bash
curl -X GET http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

### 10. Update Order Status (Admin Only)

Manually update an order's status (admin only).

```bash
curl -X PUT http://localhost:8080/api/v1/orders/1/status \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "delivered"
  }'
```

Valid status values:
- `created`
- `dispatched` 
- `in_transit`
- `delivered`
- `cancelled`

### Authentication Examples

#### Admin Login
```bash
# First, login as admin to get admin token
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin_password"
  }'

# Save admin token
export ADMIN_TOKEN="your_admin_jwt_token_here"
```

#### Complete Workflow Example
```bash
# 1. Register a new user
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"name": "Test User", "email": "test@example.com", "password": "password123"}'

# 2. Login and get token
RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "password123"}')
TOKEN=$(echo $RESPONSE | jq -r '.token')

# 3. Create an order
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"item_description": "iPhone 15 Pro"}'

# 4. Check my orders
curl -X GET http://localhost:8080/api/v1/my-orders \
  -H "Authorization: Bearer $TOKEN"

# 5. Get specific order (replace 1 with actual order ID)
curl -X GET http://localhost:8080/api/v1/orders/1 \
  -H "Authorization: Bearer $TOKEN"
```

## Order Lifecycle

Orders automatically progress through the following statuses:

1. **created** → *15 seconds (dev)* → **dispatched**
2. **dispatched** → *15 seconds (dev)* → **in_transit** 
3. **in_transit** → *15 seconds (dev)* → **delivered**

### Status Management

- **Automatic Progression**: Orders move through lifecycle stages automatically every 15 seconds (configured for development)
- **Admin Override**: Admins can manually update order status at any time using the `/orders/{orderID}/status` endpoint
- **Cancellation**: Users can cancel orders (except delivered/cancelled orders) using the `/orders/{orderID}/cancel` endpoint
- **Background Tracking**: Each order has its own goroutine managing lifecycle progression

### Monitoring Order Progression

You can monitor an order's automatic progression:

```bash
# Create order and note the returned ID
ORDER_ID=$(curl -s -X POST http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"item_description": "Test Item"}' | jq -r '.id')

# Check status every 16 seconds to see progression (dev timing)
for i in {1..4}; do
  echo "Check $i:"
  curl -s -X GET http://localhost:8080/api/v1/orders/$ORDER_ID \
    -H "Authorization: Bearer $TOKEN" | jq '.status'
  sleep 16
done
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | HTTP server port | `8080` |
| `POSTGRES_HOST` | PostgreSQL host | `localhost` |
| `POSTGRES_PORT` | PostgreSQL port | `5432` |
| `POSTGRES_USER` | Database user | - |
| `POSTGRES_PASSWORD` | Database password | - |
| `POSTGRES_DB` | Database name | - |
| `REDIS_HOST` | Redis host | `localhost` |
| `REDIS_PORT` | Redis port | `6379` |
| `REDIS_PASSWORD` | Redis password | `""` |
| `JWT_SECRET` | JWT signing secret | - |
| `ADMIN_EMAIL` | Admin user email | - |
| `ADMIN_PASSWORD` | Admin user password | - |

### Configuration Files

The application uses both environment variables and YAML configuration:

```yaml
# configs/config.yaml
server:
  port: 8080

postgres:
  host: "localhost"
  port: 5432
  user: "trachios_user"
  password: "secure_password"
  dbname: "trachios_db"
  sslmode: "disable"

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

jwt:
  secret: "your-jwt-secret-here"

admin:
  email: "admin@example.com"
  password: "admin_password"
```

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role user_role NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TYPE user_role AS ENUM ('customer', 'admin');
```

### Orders Table
```sql
CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    customer_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    item_description TEXT NOT NULL,
    status order_status NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TYPE order_status AS ENUM (
    'created', 'dispatched', 'in_transit', 'delivered', 'cancelled'
);
```

## Testing

### Running Tests

The project includes comprehensive integration tests that validate all functionality:

```bash
# Make sure the server is running first
go run cmd/server/main.go

# In another terminal, run tests
go test -v ./test/...
```

### Test Coverage

The tests validate:
- ✅ User authentication (registration, login, JWT validation)
- ✅ Order creation and retrieval
- ✅ Automated status progression (waits 16 seconds for progression)
- ✅ Order cancellation functionality
- ✅ Admin-only operations and access control
- ✅ Concurrent order processing (creates 5 orders simultaneously)

### Test Structure

The [`api_lifecycle_test.go`](test/api_lifecycle_test.go) includes:
- Authentication testing (user and admin)
- Order creation and management
- Status progression testing with 16-second wait
- Order cancellation validation
- Admin function testing
- Concurrent operation testing

## Development

### Project Structure

- **[`cmd/server/main.go`](cmd/server/main.go)**: Application entry point and routing
- **[`internal/service/order_service.go`](internal/service/order_service.go)**: Core business logic with Redis caching
- **[`internal/api/handler/`](internal/api/handler/)**: HTTP request handlers
- **[`internal/repository/`](internal/repository/)**: Data access layer
- **[`pkg/util/`](pkg/util/)**: Shared utilities (password hashing, JSON responses)

### Key Components

#### Order Service
The [`OrderService`](internal/service/order_service.go) handles:
- Order lifecycle management with background goroutines
- Redis caching with automatic invalidation (5-minute cache duration)
- Concurrent order tracking with thread-safe operations
- Admin override capabilities with goroutine management

#### Authentication Middleware
The [`AuthMiddleware`](internal/api/middleware/auth.go) provides:
- JWT token validation
- User context injection
- Role-based access control (customer/admin)

#### Repository Layer
- **[`UserRepository`](internal/repository/user_repository.go)**: User CRUD operations with admin seeding
- **[`OrderRepository`](internal/repository/order_repository.go)**: Order data management
- **Database connections**: PostgreSQL and Redis clients with connection pooling

### Adding New Features

1. **Add new domain entity** in [`internal/domain/`](internal/domain/)
2. **Create repository** in [`internal/repository/`](internal/repository/)
3. **Implement service logic** in [`internal/service/`](internal/service/)
4. **Add HTTP handlers** in [`internal/api/handler/`](internal/api/handler/)
5. **Update routes** in [`cmd/server/main.go`](cmd/server/main.go)

## Deployment

### Docker Production Build

```bash
# Build and deploy with Docker Compose
docker-compose up -d

# Or build manually
docker build -t trachios:latest .
docker run -p 8080:8080 trachios:latest
```

### Environment-Specific Configuration

Create environment-specific configuration files:
- `configs/config.dev.yaml`
- `configs/config.prod.yaml`
- `configs/config.test.yaml`

## Performance & Monitoring

### Redis Caching Strategy

- **Cache Duration**: 5 minutes per order (configured in [`order_service.go`](internal/service/order_service.go))
- **Cache Keys**: `order:{id}` format
- **Invalidation**: Automatic on status updates and admin overrides
- **Hit/Miss Logging**: Enabled for monitoring cache performance

### Database Optimization

- Connection pooling configured in [`postgres.go`](internal/repository/postgres.go)
- Max idle connections: 25
- Max open connections: 25
- Connection max lifetime: 5 minutes
- Prepared statements for better performance

### Logging

The application provides structured logging for:
- Order lifecycle events and goroutine management
- Cache operations (hits/misses) for performance monitoring
- Authentication attempts and failures
- Admin overrides and goroutine stopping/starting

## Security Considerations

- **Password Hashing**: Uses bcrypt with default cost (configured in [`password.go`](pkg/util/password.go))
- **JWT Security**: Tokens expire in 72 hours (configured in [`user_handler.go`](internal/api/handler/user_handler.go))
- **SQL Injection**: Protected with parameterized queries throughout repositories
- **Authorization**: Role-based access control with middleware
- **Admin Seeding**: Automatic admin user creation on first startup

## Error Handling

The API returns appropriate HTTP status codes:
- `200` - Success
- `201` - Created (registration, order creation)
- `400` - Bad Request (validation errors, invalid order states)
- `401` - Unauthorized (invalid/missing token)
- `403` - Forbidden (insufficient privileges)
- `404` - Not Found (user/order not found)
- `500` - Internal Server Error

Error responses follow this format:
```json
{
  "error": "Error message description"
}
```

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   ```bash
   # Check PostgreSQL is running
   docker-compose ps postgres
   
   # Check connection
   psql -h localhost -p 5432 -U username -d database
   ```

2. **Redis Connection Failed**
   ```bash
   # Check Redis is running
   docker-compose ps redis
   
   # Test connection
   redis-cli -h localhost -p 6379 ping
   ```

3. **JWT Token Issues**
   - Ensure `JWT_SECRET` is set and consistent
   - Check token expiration (72 hours default)
   - Verify Authorization header format: `Bearer <token>`

4. **Order Status Not Progressing**
   - Check server logs for goroutine errors
   - Verify order is not cancelled or delivered
   - Confirm background tracking is active

5. **Admin Access Issues**
   - Verify admin credentials in `.env` file
   - Check admin user was created during server startup
   - Ensure admin token is used for admin endpoints

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and conventions
- Add tests for new functionality
- Update documentation for API changes
- Use meaningful commit messages
- Test all endpoints with curl before submitting

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support and questions:
- Create an issue on GitHub
- Check the troubleshooting section
- Review the test cases for usage examples
- Test endpoints using the provided curl commands

---

**Trachios** - Efficient Order Management for Modern Applications