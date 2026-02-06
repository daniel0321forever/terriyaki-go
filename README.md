# Terriyaki Backend

Go backend service for the Terriyaki application.

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

4. **Run the backend server:**
   ```bash
   go run main.go
   ```

   The backend will start on **http://localhost:8080**

## Running Tests

### Option 1: Run Tests Locally

1. **Create a test database:**
   ```bash
   createdb terriyaki_test
   ```
   Or using psql:
   ```bash
   /opt/homebrew/opt/postgresql@15/bin/psql -U $(whoami) -d postgres -c "CREATE DATABASE terriyaki_test;"
   ```

2. **Update your `.env` file with test database credentials:**
   Add these lines to your `.env` file:
   ```
   TEST_POSTGRES_HOST=localhost
   TEST_POSTGRES_PORT=5432
   TEST_POSTGRES_USER=your_username
   TEST_POSTGRES_DB=terriyaki_test
   TEST_POSTGRES_PASSWORD=
   TEST_POSTGRES_SSLMODE=disable
   GORM_SILENT=true
   ```
   Replace `your_username` with your system username.

3. **Run all integration tests:**
   ```bash
   go test -v ./test/integration/...
   ```

4. **Run specific test file:**
   ```bash
   go test -v ./test/integration -run TestCreateGrindAPI
   ```

### Option 2: Run Tests in Docker

This approach creates isolated test containers with their own PostgreSQL database, ensuring a clean test environment.

1. **Navigate to the test directory:**
   ```bash
   cd test
   ```

2. **Build and run tests:**
   ```bash
   docker-compose up --abort-on-container-exit
   ```
   This will:
   - Start a PostgreSQL test database container
   - Build the test container with your code
   - Run all integration tests
   - Stop automatically when tests complete

3. **Clean up containers:**
   ```bash
   docker-compose down
   ```

4. **Rebuild after code changes:**
   ```bash
   docker-compose build
   docker-compose up --abort-on-container-exit
   ```

### Test Results

All tests should pass. You should see output ending with:
```
PASS
ok      github.com/daniel0321forever/terriyaki-go/test/integration  [time]
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
