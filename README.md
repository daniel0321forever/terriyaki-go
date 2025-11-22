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
