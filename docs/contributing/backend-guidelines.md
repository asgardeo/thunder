# Backend Coding Guidelines

Go backend guidelines for the Thunder project.

## Tech Stack

- Go (latest stable), PostgreSQL (recommended), SQLite (development/testing default).
- Testing: `stretchr/testify` (test suites), `mockery` (mock generation), `go-sqlmock` (DB mocking).
- Linting: `golangci-lint` with config in `backend/.golangci.yml`. Max line length: 120 chars.

## General Rules

- Reuse common utilities from the `internal/system` packages.
- Define interfaces for services to enable dependency injection and testability.

## Package Structure and Organization

- Each domain/feature lives in its own package under `internal/`.
- Flat directory structure within packages. Avoid nested packages unless absolutely necessary for complex domains.
- Each domain package typically contains related components organized by responsibility (not all files are required):
  - `service.go` — Service interface and implementation (business logic layer)
  - `handler.go` — HTTP handlers (presentation layer) — only if the package exposes HTTP endpoints
  - `store.go` — Data access layer (persistence) — only if the package needs database operations
  - `model.go` — Domain models and DTOs — only if the package has domain-specific models
  - `constants.go` — Package-specific constants (default values, configuration constants, business logic constants)
  - `error_constants.go` — Service and API error messages, error codes, and error-related constants
  - `store_constants.go` — Database queries, table names, column names, and database-related constants
  - `utils.go` — Package-specific utility functions
  - `init.go` — Package initialization and route registration (only if package exposes HTTP endpoints)
  - `file_based_store.go` — File-based storage implementation — only if the package needs declarative configurations
  - `composite_store.go` — Composite storage implementation — only if the package needs both database and declarative configurations
  - `declarative_resource.go` — Declarative resource implementation — only if the package needs declarative configurations
- Adjust based on actual requirements:
  - No HTTP layer? Skip `handler.go`
  - Need cache-backed storage? Add additional storage implementation files
  - No declarative resource support? Skip `file_based_store.go`, `composite_store.go`, and `declarative_resource.go`
  - Complex domain? Use subdirectories (e.g., `internal/oauth/oauth2/`, `internal/oauth/jwks/`)

## Package Exports

- Only export the service interface (e.g., `XServiceInterface`) and models used in that interface.
- Keep all internal implementations (service structs, store interfaces, store implementations, handlers) unexported (lowercase).
- Keep internal constants such as database queries, error codes, and other implementation details unexported (private).
- Example: Export `UserServiceInterface` and `User` model, but keep `userService`, `userStore`, `userHandler`, and internal query constants unexported.

## Database

- Use `DBClient` from `internal/system/database` for database operations.
- Use `DBQuery` from `internal/system/database/model` to define queries with a unique ID for traceability and DB-specific query support (e.g., SQLite vs PostgreSQL).
- All tables must include `DEPLOYMENT_ID` in composite primary keys for multi-deployment support.
- `DEPLOYMENT_ID` must always be the **last parameter** in queries. Follow these patterns:
  - **INSERT**: `INSERT INTO TABLE_NAME (COL1, COL2, ..., DEPLOYMENT_ID) VALUES ($1, $2, ..., $N)`
  - **UPDATE**: `UPDATE TABLE_NAME SET COL1 = $2, COL2 = $3 WHERE ID = $1 AND DEPLOYMENT_ID = $4`
  - **DELETE**: `DELETE FROM TABLE_NAME WHERE ID = $1 AND DEPLOYMENT_ID = $2`
  - **SELECT**: `SELECT * FROM TABLE_NAME WHERE ID = $1 AND DEPLOYMENT_ID = $2`
  - **JOIN**: `SELECT * FROM T1 LEFT JOIN T2 ON T1.ID = T2.ID AND T1.DEPLOYMENT_ID = T2.DEPLOYMENT_ID WHERE T1.ID = $1 AND T1.DEPLOYMENT_ID = $2`

## Store Layer (Data Access)

- Define store interfaces (e.g., `xStoreInterface`) and implementations (e.g., `xStore` struct) in `store.go`.
- Store layer handles all database interactions and should be used by the service layer.
- Use private constructors (e.g., `newXStore()`) to create store instances.
- Store initialization should use `DBProvider` to get the database client. Individual store methods should use the created client.
- Keep store methods focused on data access operations without business logic.

## Error Handling

- Use `ServiceError` from `internal/system/error/serviceerror` to return errors from the service layer.
- Use `ErrorResponse` from `internal/system/error/apierror` to define and return API layer errors.
- Do not log the same error twice. Return a Go error or `ServiceError` from internal components and log at the service layer.
- Do not expose internal error details in API responses for 5xx errors. Log and return a generic message like "Internal server error".

## Logging

- Use the `log` package from `internal/system`.
- Minimal info logs. Log server errors for debugging.
- Never log PII. Use `MaskString` from `internal/system/log` to mask sensitive information.
- Use `IsDebugEnabled` from `internal/system/log` before expensive debug log construction.

## Defining APIs

- Return JSON responses from APIs where applicable.
- Return JSON errors as per the server `ErrorResponse` definition. For 500 internal server errors, a generic message may be returned.
- Define API handlers in a `handler.go` file within the domain package.
- For packages with HTTP endpoints, use an `init.go` file to register routes with the mux and initialize dependencies.
- Define CORS policies using `middleware.WithCORS` from `internal/system/middleware` where applicable.

## Service Layer and Dependency Injection

- Define service interfaces (e.g., `XServiceInterface`) and implementations (e.g., `xService` struct) in `service.go`.
- Use private constructor functions (e.g., `newXService()`) to create service instances.
- If the service interacts with the database, accept the store interface as a parameter in the constructor.
- Constructor parameter order: service dependencies, store, transactioner.
  - Example without dependencies: `func newIDPService() IDPServiceInterface`
  - Example with dependencies: `func newGroupService(ouService OrganizationUnitServiceInterface, store GroupStoreInterface, transactioner Transactioner) GroupServiceInterface`
- Services should depend on interfaces, not concrete implementations, to enable testing with mocks.
- Keep constructors private (unexported) — external packages should only interact through the `Initialize()` function.

## Service Initialization and Dependency Management

- Service initialization should happen **once** during application startup in the `init.go` file of each package.
- The `Initialize(mux, deps)` function in `init.go` should:
  1. Create the store instances using the constructor functions (only if the package requires database operations)
  2. Create the service instances, passing created store and required dependencies. If a service supports declarative configurations, it should return a `ResourceExporter` during initialization.
  3. Create handlers and inject the service instance into them
  4. Register routes with the mux
  5. Return the created service interfaces (and `ResourceExporter` if applicable) for use as dependencies by other packages
- Example:
  ```go
  // In internal/user/init.go
  func Initialize(mux *http.ServeMux, ouService ou.OrganizationUnitServiceInterface) (UserServiceInterface, error) {
      userStore, err := newUserStore()
      if err != nil {
          return nil, err
      }
      transactioner, err := provider.GetDBProvider().GetUserDBTransactioner()
      if err != nil {
          return nil, err
      }
      userService := newUserService(ouService, userStore, transactioner)
      userHandler := newUserHandler(userService)
      registerRoutes(mux, userHandler)
      return userService, nil
  }
  ```
- The main service manager in `cmd/server/servicemanager.go` orchestrates all initializations in the correct order, passing dependencies as needed and collecting `ResourceExporter` interfaces for the `export` service.

## Transaction Management

- Use `transaction.Transactioner` from `internal/system/database/transaction` to handle database transactions within the service layer.
- Inject the `transactioner` into the service struct.
- Use `transactioner.Transact` to execute operations that require atomicity.
- `Transact` accepts a context and a function. The function receives a new context `txCtx` which **must** be passed to all store operations within the closure.
- If the closure returns an error, the transaction is rolled back. If it returns `nil`, the transaction is committed.
- Example:
  ```go
  func (us *userService) CreateUser(ctx context.Context, user *User) (*User, *serviceerror.ServiceError) {
      err = us.transactioner.Transact(ctx, func(txCtx context.Context) error {
          // IMPORTANT: Pass txCtx to store methods, not the original ctx
          return us.userStore.CreateUser(txCtx, *user, credentials)
      })
      if err != nil {
          return nil, logErrorAndReturnServerError(logger, "Failed to create user", err)
      }
      return user, nil
  }
  ```

## Common Utilities

- HTTP: `HTTPClient` from `internal/system/http`
- Cache: Extend `BaseCache` from `internal/system/cache`
- Config: `ThunderRuntime` from `internal/system/config`
- Constants: `internal/system/constants`
- Middleware: `middleware.WithCORS` from `internal/system/middleware`

## Testing

### Unit Tests

- Ensure unit tests achieve at least 80% coverage.
- Use `stretchr/testify` for tests and follow the test suite pattern.
- `mockery` is used to generate mocks; configurations are in `.mockery.private.yml` and `.mockery.public.yml`.
- **IMPORTANT**: After modifying any interface, regenerate mocks by running `make mockery` and commit the updated mock files. CI includes a `verify-mocks` job that will fail if mocks are out of sync.
- Place generated mocks in `/backend/tests/mocks/`. Exception: when a package's tests need a mock of an interface defined in the same package, a generated mock in a `_mock_test.go` file within that package is acceptable to avoid circular imports.
- Run unit tests: `make test_unit` from project root, or `go test` from the `backend/` directory with applicable flags.

### Integration Tests

- Write integration tests in `/tests/integration/` where applicable.
- Target combined unit + integration coverage of at least 80%.
- Run integration tests: `make test_integration` from project root (requires a built product), or `make all` to build and test end-to-end.
