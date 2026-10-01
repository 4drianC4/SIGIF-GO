# SIGIF-GO - Inventory & Sales Management System

A multi-tenant inventory and sales management system built with Go, Fiber, and GORM. Designed with Vertical Slice + Hexagonal Architecture for scalability and maintainability.

## Features

- **Multi-tenancy**: Support for multiple businesses (minimarket, hardware, pharmacy)
- **Inventory Management**: Products, categories, brands, units, taxes, stock movements, warehouses, batches, serial numbers
- **Sales Management**: Sales, customers, payments, returns, promotions, loyalty programs
- **Pharmacy Module**: Prescriptions, controlled substances, doctor management
- **Hardware Module**: Serial number tracking, warranties, service orders
- **Authentication**: JWT with access/refresh tokens, role-based access control
- **Dashboard & Analytics**: Metrics, reports, scheduled reports
- **Event-Driven**: Domain events for cross-module communication

## Tech Stack

- **Language**: Go 1.23
- **Framework**: Fiber v2
- **ORM**: GORM with PostgreSQL
- **Cache**: Redis
- **Auth**: JWT (HS256) with Argon2id password hashing
- **DI**: Uber FX
- **Config**: Viper (YAML + Environment variables)
- **Logging**: Zap
- **Validation**: Go Playground Validator v10
- **API Client**: Bruno

## Project Structure

```
.
├── cmd/
│   ├── server/          # Main HTTP server
│   └── migrate/         # Database migration tool
├── configs/
│   └── config.yaml      # Configuration file
├── deployments/
│   └── docker/          # Docker Compose, Dockerfiles
├── docs/
│   └── api/             # API documentation
├── internal/
│   ├── modules/         # Vertical slice modules
│   │   ├── tenant/
│   │   ├── company/
│   │   ├── user/
│   │   ├── auth/
│   │   ├── product/
│   │   ├── inventory/
│   │   ├── sales/
│   │   ├── pharmacy/
│   │   ├── minimarket/
│   │   ├── hardware/
│   │   └── dashboard/
│   └── shared/          # Shared kernel
│       ├── config/
│       ├── database/
│       ├── errors/
│       ├── logger/
│       ├── middleware/
│       ├── validator/
│       ├── pagination/
│       ├── response/
│       ├── security/
│       ├── jwt/
│       ├── events/
│       ├── clock/
│       └── transaction/
├── pkg/
│   └── container/       # FX container wiring
├── bruno/               # Bruno API collections
├── go.mod
├── go.sum
├── Makefile
└── .env.example
```

## Quick Start

### Prerequisites

- Go 1.26+
- Docker & Docker Compose
- Bruno (for API testing)

### Development Setup

```bash
# Clone and enter project
cd SIGIF-GO

# Install development tools
make install-tools

# Start infrastructure (PostgreSQL, Redis)
make docker-up

# Wait for database to be ready (or run migrations manually)
sleep 10

# Run database migrations
make migrate-up

# Start the application
make run
```

The server will be available at `http://localhost:8080`

### Using Docker Compose (Full Stack)

```bash
# Build and start all services
make docker-build
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down
```

## Configuration

Copy `.env.example` to `.env` and adjust values:

```bash
cp .env.example .env
```

Key environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_PORT` | HTTP server port | 8080 |
| `DATABASE_HOST` | PostgreSQL host | localhost |
| `DATABASE_PORT` | PostgreSQL port | 5432 |
| `DATABASE_USER` | PostgreSQL user | sigif |
| `DATABASE_PASSWORD` | PostgreSQL password | sigif |
| `DATABASE_NAME` | Database name | sigif |
| `JWT_SECRET` | JWT signing secret | (required) |
| `REDIS_HOST` | Redis host | localhost |
| `REDIS_PORT` | Redis port | 6379 |

## Available Commands

```bash
# Development
make run              # Run server locally on port 8080
make run PORT=3000    # Run server locally on port 3000
make build            # Build binaries to bin/
make test             # Run tests with coverage
make lint             # Run golangci-lint
make fmt              # Format code
make vet              # Go vet
make check            # Run fmt, vet, lint, test

# Database
make migrate-up       # Run migrations
make docker-up        # Start PostgreSQL & Redis
make docker-down      # Stop containers
make docker-logs      # View container logs

# Code Generation
make generate         # Run go generate
make deps             # Download & tidy dependencies

# Production
make prod-build       # Build optimized Linux binaries
```

## API Endpoints

All endpoints require `X-Tenant-ID` header.

### Authentication
```
POST   /api/v1/auth/login       # Login
POST   /api/v1/auth/refresh     # Refresh access token
POST   /api/v1/auth/logout      # Logout
POST   /api/v1/auth/logout-all  # Logout from all devices
```

### Tenants
```
POST   /api/v1/tenants          # Create tenant
GET    /api/v1/tenants          # List tenants
GET    /api/v1/tenants/:id      # Get tenant
GET    /api/v1/tenants/slug/:slug # Get tenant by slug
PUT    /api/v1/tenants/:id      # Update tenant
DELETE /api/v1/tenants/:id      # Delete tenant
POST   /api/v1/tenants/:id/activate
POST   /api/v1/tenants/:id/deactivate
```

### Companies
```
POST   /api/v1/companies        # Create company
GET    /api/v1/companies        # List companies
GET    /api/v1/companies/:id    # Get company
PUT    /api/v1/companies/:id    # Update company
DELETE /api/v1/companies/:id    # Delete company
```

### Users
```
POST   /api/v1/users            # Create user
GET    /api/v1/users            # List users
GET    /api/v1/users/:id        # Get user
GET    /api/v1/users/by-email   # Get user by email
PUT    /api/v1/users/:id        # Update user
PUT    /api/v1/users/:id/password # Change password
DELETE /api/v1/users/:id        # Delete user
```

### Products
```
POST   /api/v1/products         # Create product
GET    /api/v1/products         # List products (with filters)
GET    /api/v1/products/:id     # Get product
GET    /api/v1/products/sku/:sku # Get by SKU
GET    /api/v1/products/barcode/:barcode # Get by barcode
PUT    /api/v1/products/:id     # Update product
DELETE /api/v1/products/:id     # Delete product
POST   /api/v1/products/:id/activate
POST   /api/v1/products/:id/deactivate
POST   /api/v1/products/:id/stock/adjust  # Adjust stock
POST   /api/v1/products/:id/stock/set     # Set stock
```

### Inventory
```
POST   /api/v1/warehouses       # Create warehouse
GET    /api/v1/warehouses       # List warehouses
POST   /api/v1/batches          # Create batch
GET    /api/v1/batches          # List batches
POST   /api/v1/suppliers        # Create supplier
GET    /api/v1/suppliers        # List suppliers
POST   /api/v1/purchase-orders  # Create PO
GET    /api/v1/purchase-orders  # List POs
```

### Sales
```
POST   /api/v1/sales            # Create sale
GET    /api/v1/sales            # List sales
GET    /api/v1/sales/:id        # Get sale
POST   /api/v1/sales/:id/complete # Complete sale
POST   /api/v1/sales/:id/cancel   # Cancel sale
POST   /api/v1/customers        # Create customer
GET    /api/v1/customers        # List customers
POST   /api/v1/payments         # Create payment
POST   /api/v1/returns          # Create return
```

### Pharmacy (Prescriptions)
```
POST   /api/v1/prescriptions    # Create prescription
GET    /api/v1/prescriptions    # List prescriptions
GET    /api/v1/prescriptions/:id # Get prescription
POST   /api/v1/prescriptions/:id/verify  # Verify prescription
POST   /api/v1/prescriptions/:id/dispense # Dispense prescription
POST   /api/v1/prescriptions/:id/cancel  # Cancel prescription
POST   /api/v1/doctors          # Create doctor
GET    /api/v1/doctors          # List doctors
```

### Minimarket (Promotions & Loyalty)
```
POST   /api/v1/promotions       # Create promotion
GET    /api/v1/promotions       # List promotions
POST   /api/v1/loyalty/programs # Create loyalty program
GET    /api/v1/loyalty/programs # List programs
POST   /api/v1/loyalty/customers # Enroll customer
GET    /api/v1/loyalty/customers/:id # Get customer loyalty
```

### Hardware (Serial Numbers & Services)
```
POST   /api/v1/serial-numbers   # Create serial number
GET    /api/v1/serial-numbers   # List serial numbers
POST   /api/v1/serial-numbers/:id/sell   # Sell serial
POST   /api/v1/serial-numbers/:id/return # Return serial
POST   /api/v1/warranty-claims  # Create warranty claim
GET    /api/v1/warranty-claims  # List claims
POST   /api/v1/service-orders   # Create service order
GET    /api/v1/service-orders   # List service orders
```

### Dashboard & Reports
```
GET    /api/v1/dashboard/metrics     # Get dashboard metrics
GET    /api/v1/dashboard/sales       # Sales analytics
GET    /api/v1/dashboard/inventory   # Inventory analytics
POST   /api/v1/reports               # Create report
GET    /api/v1/reports               # List reports
POST   /api/v1/reports/:id/execute   # Execute report
GET    /api/v1/reports/:id/executions # List executions
```

## Bruno Collections

Import the collections from `bruno/` directory:

```bash
# In Bruno: File > Import > Folder > Select bruno/ directory
```

Collections included:
- **Auth** - Login, refresh, logout
- **Tenants** - CRUD operations
- **Companies** - CRUD operations
- **Users** - CRUD + password change
- **Products** - CRUD + stock management
- **Inventory** - Warehouses, batches, suppliers, POs
- **Sales** - Sales, customers, payments, returns
- **Pharmacy** - Prescriptions, doctors
- **Minimarket** - Promotions, loyalty
- **Hardware** - Serial numbers, warranties, services
- **Dashboard** - Metrics, reports

Each collection uses variables for `baseUrl`, `tenantId`, `accessToken`, etc.

## Testing

```bash
# Run all tests
make test

# Run with coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run specific module tests
go test -v ./internal/modules/tenant/...
```

## Architecture Guidelines

### Adding a New Module

1. Create module structure:
```
internal/modules/newmodule/
├── domain/
│   ├── entity/
│   ├── repository/
│   ├── service/
│   └── event/
├── application/
│   ├── command/
│   ├── query/
│   ├── handler/
│   ├── dto/
│   └── port/
├── infrastructure/
│   └── persistence/
│       └── gorm/
├── interfaces/
│   └── http/
│       ├── handler/
│       ├── router/
│       └── dtos/
└── module.go
```

2. Implement domain entities with business logic
3. Define repository interfaces (ports)
4. Implement GORM repositories (adapters)
5. Create application commands/queries/handlers
6. Build HTTP handlers and routers
7. Wire in `module.go` with FX
8. Register in `cmd/server/main.go`

### Dependency Rules

- **Domain**: No external dependencies
- **Application**: Depends only on Domain + Shared
- **Infrastructure**: Implements Domain ports
- **Interfaces**: Depends on Application + Shared
- **Shared**: No dependencies on modules

### Cross-Module Communication

Use domain events via shared event bus:

```go
// Publish event
eventBus.Publish(ctx, events.NewOrderCreatedEvent(order))

// Subscribe in another module
eventBus.Subscribe("order.created", func(ctx context.Context, event Event) error {
    // Handle event
    return nil
})
```

## Database Migrations

Migrations run automatically on startup via GORM AutoMigrate. For production, use proper migration tools:

```bash
# Generate migration SQL
go run ./cmd/migrate
```

## Health Check

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "ok",
  "version": "1.0.0",
  "time": "2024-01-15T10:30:00Z"
}
```

## License

Proprietary - SIGIF Internal Use Only