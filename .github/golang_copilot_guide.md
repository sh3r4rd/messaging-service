# Go Project - AI Coding Guide for GitHub Copilot

**Owner: Sherard Bailey | Echo Framework | PostgreSQL | **

---

## 📆 Project Overview

**Core Pattern**: REST API backend using Echo + PostgreSQL + standard library

- **Routing**: Echo router with middleware support
- **Database**: PostgreSQL using `database/sql`
- **Data Modeling**: Custom Go types with explicit marshaling
- **Deployment**: Docker + CI/CD ready

Project reference: [messaging-service](https://github.com/sh3r4rd/messaging-service)

---

## ⚖️ Architecture Principles

1. **Use Database Driver**: with SQL statements and transaction control
2. **Strict typing**: Define models and request types explicitly
3. **Clean separation** of concerns: `server`, `repository`, `service`, `integrationtests`
4. **Centralized error formatting and logging**
5. **Request validation** via `go-playground/validator`
6. **Consistent integration testing** using a real DB container

---

## 📁 Directory Structure

```
/messaging-service
|-- cmd                   # Entrypoint for `migrate` and `server` commands
|-- internal/
|   |-- app/                # Application logic and server
|   |-- app/server          # API handlers and validation
|   |-- app/integrationtests/ # Full integration test suite
|   |-- pkg/                # Internal packages used by the `migrate` and `server` commands
|   |-- pkg/apperrors/      # Defines application errors: service errors, database errors, HTTP errors, etc
|   |-- pkg/migration/      # Migration logic for running and rolling back database migrations
|   |-- pkg/repositiory/    # Database interactions and queries
|   |-- pkg/service/        # Services for interacting with external resources, eg cache, external APIs
|-- migrations/           # SQL migration scripts
|-- Dockerfile
```

---

## 🚀 Development Workflow

### Run Locally

#### Server
Run the server
```bash
make server
```

#### Migrate
Run migrations
```bash
make migrate.up // up migrations
make migrate.down // rollback migrations
```

### Run Integration Tests

```bash
make integrations.test
```

---

## ⚙️ Adding New Features (Copilot Prompts)

1. **Add new API endpoint**

   - Define handler in `internal/app/server/`
   - Add service logic to `internal/pkg/service/`
   - Use database interaction in `internal/pkg/repository/`
   - Update integration tests in `internal/app/integrationtests/`

2. **Add new database table**

   - Write SQL in `migrations/*.up.sql` AND `migrations/*.down.sql`
   - Add queries in `internal/pkg/repository/`
   - Reflect in models and service layer

3. **Add validation**
   - Use struct tags to input models in `internal/app/server/input.go`

---

## 📊 Database Access

### With `database/sql`

```go
stmt := `SELECT id, body FROM messages WHERE id = $1`
row := db.QueryRowContext(ctx, stmt, id)
var msg Message
err := row.Scan(&msg.ID, &msg.Body)
```

### Transactions

```go
tx, err := db.BeginTx(ctx, nil)
defer tx.Rollback()
// perform tx.ExecContext or tx.QueryRowContext
err = tx.Commit()
```

### Avoid ORM & sqlx

- Use native types and scanning for full control
- Define all SQL explicitly in `migrations/`

---

## 🔒 Middleware & Validation

- **Logging**: Use structured loggers with request context
- **CORS/Auth**: Register middleware in `internal/app/server/main.go`
- **Validator**: Initialize and use `go-playground/validator`

---

## 🎉 Testing Strategy

### Integration Tests (Real DB)

- Located in: [`internal/app/integrationtests/integration_test.go`](https://github.com/sh3r4rd/messaging-service/blob/main/internal/app/integrationtests/integration_test.go)
- Use Docker container for Postgres setup
- Reset database between tests

```go
func TestSendSMS(t *testing.T) {
  cleaner.Acquire(tables...)
  defer cleaner.Clean(tables...)

    // seed data, send request, assert response
}
```

📌 **Rules:**

- Always use a real database for integration tests
- Use setup/teardown helpers to isolate test cases
- Validate both HTTP response and DB state

---

## 🔧 Tooling

- `go fmt` and `goimports`
- Lint with `golangci-lint`
- Use Makefile or shell scripts for test/migrate/dev tasks

---

## 💡 Copilot Prompt Examples

- "Create a new Echo handler to send a message"
- "Write a Postgres query to get conversations with latest message"
- "Add validation to ensure body is not empty"
- "Add an integration test for a failed email message"
- "Write a transaction to create a message and link to a conversation"

---

## 🚨 Final Thoughts

1. **Use Copilot to scaffold**, then verify correctness
2. **Keep handlers thin**, push logic into `app/`
3. **Use transactions where needed**, but keep them tight
4. **Use integration tests to catch real-world failures**
5. **Avoid premature abstraction**, favor clear readable code

---
