# Integration Test Suite

This directory contains comprehensive integration tests for all API endpoints in the Terriyaki Go application.

## Test Structure

### Helper Functions (`test/helper.go`)

1. **`AssertMapValues`** - Recursive map value checker that validates response structures
   - Supports exact value matching
   - Supports wildcard matching with `"*"` (checks key exists with any value)
   - Supports nested map validation
   - Supports array/slice validation
   - Provides detailed error messages with path information

2. **`ParseResponseMap`** - Parses JSON response body into map[string]interface{}

3. **`MakeAuthenticatedRequest`** - Makes HTTP requests with custom bearer token authentication

### Test Files

#### 1. `auth_api_test.go`
Tests for authentication endpoints:
- `POST /api/v1/register` - User registration (successful, duplicate email, invalid body, etc.)
- `POST /api/v1/login` - User login (successful, invalid credentials, etc.)
- `POST /api/v1/logout` - User logout
- `GET /api/v1/verify-token` - Token verification

#### 2. `auth_v2_api_test.go`
Tests for v2 authentication endpoints:
- `POST /api/v2/login` - V2 login with grinds map
- `GET /api/v2/verify-token` - V2 token verification

#### 3. `ping_api_test.go`
Tests for health check:
- `GET /api/v1/ping` - Ping endpoint

#### 4. `user_api_test.go`
Tests for user management:
- `GET /api/v1/users/exists` - Check user existence by email
- `DELETE /api/v1/users/delete` - Delete user account

#### 5. `grind_api_test.go`
Tests for grind (challenge) management:
- `POST /api/v1/grinds` - Create new grind
- `GET /api/v1/grinds` - Get all user grinds
- `GET /api/v1/grinds/current` - Get current active grind
- `GET /api/v1/grinds/:id` - Get specific grind by ID
- `POST /api/v1/grinds/:id/quit` - Quit a grind
- `GET /api/v1/grinds/:id/progress` - Get progress records
- `DELETE /api/v1/grinds/delete-all` - Delete all grinds

#### 6. `task_api_test.go`
Tests for task management:
- `GET /api/v1/tasks/today` - Get today's task
- `POST /api/v1/tasks/finish` - Mark task as finished
- `GET /api/v1/tasks/:id` - Get specific task

#### 7. `message_api_test.go`
Tests for messaging system:
- `GET /api/v1/messages` - Get received messages
- `GET /api/v1/messages/sent` - Get sent messages
- `POST /api/v1/messages/:id/read` - Mark message as read
- `POST /api/v1/messages/invitation` - Create grind invitation
- `POST /api/v1/messages/:id/invitation/accept` - Accept invitation
- `POST /api/v1/messages/:id/invitation/reject` - Reject invitation

#### 8. `profile_api_test.go`
Tests for user profile:
- `PATCH /api/v1/profile` - Update user profile (username, avatar)

#### 9. `interview_api_test.go`
Tests for interview system:
- `POST /api/v1/interviews/start` - Start interview session
- `POST /api/v1/interviews/:id/end` - End interview session
- `POST /api/v1/interviews/:id/response` - Save agent response
- `POST /api/v1/interviews/llm` - LLM webhook endpoint

#### 10. `payment_api_test.go`
Tests for payment system (Stripe integration):
- `POST /api/v1/payments/payment-intent` - Create payment intent
- `POST /api/v1/payments/save-card-intent` - Create save card intent
- `POST /api/v1/payments/save-card` - Save card to user account
- `GET /api/v1/payments/methods` - Get available payment methods
- `POST /api/v1/payments/methods/select-default` - Select default payment method
- `POST /api/v1/charges/dued` - Force investigate dued penalties
- `POST /api/v1/payments/force-charging` - Test force charging

## Running Tests

### Run all integration tests:
```bash
cd test/integration
go test -v
```

### Run specific test file:
```bash
go test -v -run TestPingAPI
```

### Run specific test case:
```bash
go test -v -run TestRegisterController/Successful_Registration
```

## Test Patterns

### 1. Using AssertMapValues for Response Validation

```go
response := test.ParseResponseMap(t, rr.Body.Bytes())
expected := map[string]interface{}{
    "message": "Login successful",
    "user":    map[string]interface{}{"id": "*", "email": testEmail},
    "token":   "*",
}
test.AssertMapValues(t, response, expected, "")
```

### 2. Making Authenticated Requests

```go
// Get token first
loginRr := test.MakeRequest("POST", "/api/v1/login", loginReq, false)
loginResponse := test.ParseResponseMap(t, loginRr.Body.Bytes())
token := loginResponse["token"].(string)

// Make authenticated request
rr := test.MakeAuthenticatedRequest("GET", "/api/v1/verify-token", nil, token)
```

### 3. Test Data Cleanup

Always clean up test data using defer:

```go
testEmail := "test@example.com"
user, _ := models.CreateUser("testuser", testEmail, "password", "avatar.jpg")
defer database.Db.Delete(&user)
```

## Coverage

This test suite covers:
- ✅ All authentication endpoints (v1 and v2)
- ✅ User management endpoints
- ✅ Grind (challenge) CRUD operations
- ✅ Task management
- ✅ Messaging and invitations
- ✅ User profile updates
- ✅ Interview system
- ✅ Payment and charging system
- ✅ Health check endpoint

Each endpoint is tested for:
- Successful operations
- Unauthorized access (where applicable)
- Invalid input handling
- Not found scenarios
- Error conditions

## Notes

- Tests use a test database configured via `database.ConnectTestDB()`
- Database migrations are run before tests via `migrate.MigrateDatabase()`
- Each test creates its own test users with unique emails to avoid conflicts
- Cleanup is performed using defer statements to ensure test data is removed
