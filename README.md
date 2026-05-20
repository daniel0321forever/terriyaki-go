# Terriyaki Backend

Go backend service for the Terriyaki application.

## Architecture
This backend follows **Domain-Driven Design (DDD)** principles with a clean architecture structure. The codebase is organized into distinct layers, each with clear responsibilities.

```bash
internal/
├── domain/ # Domain Layer (Core Business Logic)
│ ├── entities/ # Domain entities (business objects)
│ └── repositories/ # Repository interfaces (abstractions)
│
├── application/ # Application Layer (Use Cases)
│ ├── dto/ # Data Transfer Objects (API contracts)
│ ├── mappers/ # Entity to DTO converters
│ └── services/ # Application services (orchestration)
│
├── infrastructure/ # Infrastructure Layer (Technical Details)
│ └── db/postgres/ # Repository implementations (GORM)
│
└── interface/api/ # HTTP controllers (Gin handlers)
```
### Layer Responsibilities
#### 1. Domain Layer ([internal/domain/](internal/domain/))

> **Purpose**: Core business logic and domain concepts.

##### **Entities** ([internal/domain/entities](internal/domain/entities)):
- Pure business objects with no infrastructure dependencies
- Validation can be placed
- Use factory pattern constructors (e.g., `NewUser()`, `NewGrind()`)
- [Example](internal/domain/entities/user.go):
   ```go
   type User struct {
         ID             string
         Username       string
         Email          string
         Avatar         string
         HashedPassword string
   }

   func NewUser(username, email, hashedPassword, avatar string) (*User, error) {
         // Validation logic here
         return &User{...}, nil
   }
   ```
##### **Repository Interfaces** ([internal/domain/repositories/](internal/domain/repositories/)):
- Define *contracts* for data access
- No implementation details (interfaces only)
- [Example](internal/domain/repositories/user_repository.go):
  ```go
  type UserRepository interface {
      FindById(id string) (*entities.User, error)
      FindByEmail(email string) (*entities.User, error)
      Create(user *entities.User) error
      Delete(id string) error
  }
  ```
#### 2. Application Layer (`internal/application/`)

> **Purpose**: Orchestrates domain objects to fulfill use cases.
##### **Services** [internal/application/services/](internal/application/services/):
- Orchestrate domain objects
- Accept and return **DTOs** instead of entities
- Handle business workflows
- Example:
   ```go
  type UserService struct {
      userRepo repositories.UserRepository
  }

  func (s *UserService) CreateUser(request dto.CreateUserDTO) (*dto.UserDTO, error) {
      // 1. Business logic
      // 2. Create entity
      // 3. Save via repository
      // 4. Return DTO
  }
  ```

##### **DTOs** ([internal/application/dto/](internal/application/dto/)):
- Data Transfer Objects for *service*-level input/output
- [Example](internal/application/dto/user_dto.go):
  ```go
  /*DTO*/
  // Input DTO
  type CreateUserDTO struct {
      Username string `json:"username" validate:"required"`
      Email    string `json:"email" validate:"required,email"`
      Password string `json:"password" validate:"required,min=6"`
      Avatar   string `json:"avatar,omitempty"`
  }

  // Output DTO
  type UserDTO struct {
      ID       string `json:"id"`
      Username string `json:"username"`
      Email    string `json:"email"`
      Avatar   string `json:"avatar"`
  }
  /*service*/
   func (s *UserService) GetUser(request dto.GetUserDTO) (*dto.UserDTO, error) {
      user, err := s.userRepo.FindById(request.UserID)
      if err != nil {
         return nil, errors.New("user not found")
      }
      return mappers.UserToUserDTO(user), nil
   }

  ```
##### **Mappers** [internal/application/mappers/](internal/application/mappers/):
- Convert *entities* to *DTOs*
- Centralized conversion logic makes entity to DTO transfer easier for nested entities such as [`Grind`](internal/domain/entities/grind.go)
- [Example](internal/application/mappers/user_mapper.go):
   ```go
  func UserToUserDTO(user *entities.User) *dto.UserDTO {
      return &dto.UserDTO{
          ID:       user.ID,
          Username: user.Username,
          Email:    user.Email,
          Avatar:   user.Avatar,
      }
  }
  ```

#### 3. Infrastructure Layer [internal/infrastructure](internal/infrastructure/)

> **Purpose**: Implements technical details (database, external APIs, etc.)

##### **Repository Implementations** [internal/infrastructure/db/postgres/](internal/infrastructure/db/postgres/):
- Implement [*repository interfaces*](internal/domain/repositories/)
- Handle database-specific code (GORM)
- Map between database schemas and entities
- [Example](internal/infrastructure/db/postgres/user_repository_impl.go):
  ```go
  type GormUserRepository struct {
      db *gorm.DB
  }

  func (r *GormUserRepository) FindById(id string) (*entities.User, error) {
      // GORM implementation
  }
  ```

#### 4. Interface Layer [internal/interface](internal/interface/)

> **Purpose**: Handles HTTP requests and responses.

##### **Controllers** [internal/interface/api](internal/interface/api/):
- Handle HTTP routing
- Parse request bodies into DTOs
- Call services with DTOs
- Return JSON responses
- [Example](internal/interface/api/user_controller.go):
   ```go
  func (ctrl *UserController) RegisterAPI(c *gin.Context) {
      var body dto.CreateUserDTO
      if err := c.ShouldBindJSON(&body); err != nil {
          c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
          return
      }

      userDTO, err := ctrl.userService.CreateUser(body)
      if err != nil {
          c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
          return
      }

      c.JSON(http.StatusOK, gin.H{"user": userDTO})
  }
  ```

### Payment Reliability and Boundary Layers

The payment flow gained a few dedicated layers because payments are the part of the system where retry safety and provider boundaries matter most.

#### Payment data flow

```text
HTTP request
  -> payment controller
  -> payment service
  -> idempotency claim
  -> provider adapter
  -> settlement persistence
  -> cached replay or fresh response
```

Example flow for `CreateStripeCollectionIntent`:

```go
func (ctrl *PaymentController) PaymentIntentAPI(c *gin.Context) {
  idempotencyKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
  if idempotencyKey == "" {
    RespondBadRequest(c, "Missing required Idempotency-Key header")
    return
  }

  clientSecret, replayed, err := ctrl.stripeService.CreateStripeCollectionIntent(dto.StripeCreateIntentDTO{AmountCents: amount}, idempotencyKey)
  if err != nil {
    RespondInternalServerError(c, "Internal Server Error")
    return
  }

  c.JSON(200, gin.H{
    "clientSecret":      clientSecret,
    "idempotent_replay": replayed,
  })
}
```

#### Why the payment reliability layer exists

- **Idempotency repository**: protects payment operations from duplicate execution when clients retry requests or when the network fails after the provider has already accepted the call.
- **Settlement repository**: records the lifecycle of each payment operation so the backend can reconcile pending, failed, captured, and refunded states over time.
- **Provider-neutral payment model**: keeps payment method and settlement data independent from any single provider so Stripe, Solana, or another backend can fit behind the same application contract.
- **Safe-by-default service API**: public payment write methods now expose only idempotent variants, which removes the accidental footgun of calling an unsafe payment path directly.

#### Why the Rust boundary document exists

- **Keeps chain logic behind the adapter boundary**: the Go application service should not know about chain SDK details or smart-contract execution steps.
- **Defines the contract once**: the Rust-facing behavior is documented in one place so adapter implementations can stay consistent.
- **Supports future on-chain providers**: the boundary document makes it easier to add or swap chain execution backends without reshaping the rest of the payment flow.

#### What this buys the system

- **Retry safety**: repeated requests reuse the first successful result instead of creating duplicate side effects.
- **Reconciliation**: failed or stale settlements can be revisited and repaired later.
- **Architectural clarity**: the codebase now separates request handling, payment orchestration, persistence, and provider-specific execution more explicitly.

### Others
#### Aggregate Roots
- **Grind** is the main aggregate root
- Deleting a Grind should probably also delete Tasks and Participations

## Quick Start
### Prerequisites

1. **Go** (version 1.25.3 or later) - [Install Go](https://golang.org/dl/)
2. **PostgreSQL** (version 15 or later) - [Install PostgreSQL](https://www.postgresql.org/download/)

## Database Setup

1. **Start PostgreSQL service:**
   ```bash
   brew services start postgresql@15
   ```
   Verify it's running:
   ```bash
   brew services list | grep postgresql
   ```

2. **Find your PostgreSQL username:**
   When PostgreSQL is installed via Homebrew, it creates a superuser with your system username. To find it:
   ```bash
   whoami
   ```
   Or check existing PostgreSQL roles:
   ```bash
   /opt/homebrew/opt/postgresql@15/bin/psql -U $(whoami) -d postgres -c "\du"
   ```
   The default superuser is typically your macOS username (e.g., `li-yangtseng`).

3. **Create a database for the application:**
   ```bash
   createdb terriyaki
   ```
   Or using psql:
   ```bash
   /opt/homebrew/opt/postgresql@15/bin/psql -U $(whoami) -d postgres -c "CREATE DATABASE terriyaki;"
   ```

## Backend Setup

1. **Install Go dependencies:**
   ```bash
   go mod download
   ```

2. **Create a `.env` file from the example:**
   ```bash
   cp .env.example .env
   ```

3. **Edit the `.env` file with your PostgreSQL credentials:**

   **Option A: Use your system username (recommended for local development):**
   ```
   POSTGRES_HOST=localhost
   POSTGRES_PORT=5432
   POSTGRES_USER=user
   POSTGRES_DB=terriyaki
   POSTGRES_PASSWORD=
   ```
   Replace `user` with your actual system username (run `whoami` to find it).

## Running the Backend Server

### Start the Backend

```bash
go run internal/cmd/api_server/main.go
```

The backend will start on **http://localhost:8080**

### Check Server Health

Verify the server is running by visiting:

- **Health Check (v1)**: [http://localhost:8080/api/v1/ping](http://localhost:8080/api/v1/ping)
- **OpenAPI Documentation**: See [`openapi.yaml`](../openapi.yaml) in the project root for complete API contract

### Test Authentication & API

1. **Register a new user:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/register \
     -H "Content-Type: application/json" \
     -d '{"username":"testuser","email":"test@example.com","password":"password123"}'
   ```

2. **Verify response contains JWT token** — use this token for authenticated requests:
   ```bash
   curl http://localhost:8080/api/v1/verify-token \
     -H "Authorization: Bearer YOUR_JWT_TOKEN"
   ```

3. **View all available endpoints** in `openapi.yaml`

## Testing Strategy

This project uses four testing levels. Keep each level in its owning layer to keep responsibilities clear and avoid mixing intent.

### 1. Entity Tests (Domain)

- **Scope**: Domain invariants and constructors only.
- **Location**: `internal/domain/entities/*_test.go`
- **Dependencies**: No database, no HTTP.
- **Command**:
  ```bash
  go test -v ./internal/domain/entities/...
  ```

### 2. Application Tests (Service Layer)

- **Scope**: Use-case orchestration and error mapping using repository mocks.
- **Location**: `internal/application/services/*_test.go`
- **Dependencies**: Mock repositories from `internal/domain/mocks`.
- **Command**:
  ```bash
  go test -v ./internal/application/services/...
  ```
- **Payment-specific focus**: idempotency replay, settlement transitions, and reconciliation behavior should be covered here because those rules belong to the service layer.

### 3. Repository Integration Tests (Infrastructure Adapter)

- **Scope**: Validate GORM/Postgres adapter behavior with real database semantics.
- **Location**: `internal/infrastructure/db/postgres/*_integration_test.go`
- **Build tag**: `integration`
- **Runtime**: Docker/Testcontainers required.
- **Command**:
  ```bash
  go test -tags=integration -v ./internal/infrastructure/db/postgres/...
  ```

### 4. End-to-End API Tests (Frontend-facing contract)

- **Scope**: Full request path from HTTP handlers through services to persistence.
- **Location**: `tests/integration/*_test.go`
- **Dependencies**: Real DB connection (test env) and router wiring.
- **Command**:
  ```bash
  go test -v ./tests/integration/...
  ```

### Why This Separation?

- **Fast feedback**: Entity and service tests run quickly and isolate logic regressions.
- **SQL confidence**: Repository integration tests catch schema/query issues early.
- **Frontend stability**: End-to-end tests protect response contracts and status/error behaviors.
- **DDD alignment**: Test code sits with the layer that owns the behavior.

## Running Tests

### Option 1: Local commands by level

```bash
# Domain
go test -v ./internal/domain/entities/...

# Application
go test -v ./internal/application/services/...

# Repository integration (requires Docker)
go test -tags=integration -v ./internal/infrastructure/db/postgres/...

# End-to-end API
go test -v ./tests/integration/...
```

### Option 2: Docker Compose for E2E harness

From backend root:

```bash
cd tests
docker-compose up --abort-on-container-exit
docker-compose down
```

## Local Integration Test
- The problem
  - Frontend → ElevenLabs: works (browser can reach ElevenLabs servers)
  - ElevenLabs → Backend: fails (ElevenLabs servers can't reach localhost:8080)
- Use ngrok to test the request from ElevenLabs to backend running on local.
   - Ngrok creates a public URL that tunnels to your localhost:
   ```text
   Your Backend (localhost:8080)
       ↓
   Ngrok Tunnel
       ↓
   Public URL: https://abc123.ngrok.io
       ↓
   ElevenLabs servers can now reach it!
   ```
