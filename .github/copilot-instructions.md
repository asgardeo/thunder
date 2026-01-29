# GitHub Copilot Custom Instructions

## Project Overview

This repository is a lightweight user and identity management product written in Go. The primary focus is to provide authentication and authorization capabilities for applications and also to allow managing users, roles, and permissions. The server allows three different approaches to authenticate users: by using standard protocols such as OAuth2, OIDC, by defining a flexible orchestration flow, or by using individual authentication mechanisms such as password, passwordless, and social login.

The project is structured as a monorepo to manage the backend, frontend, and sample applications in a single repository. The backend code is located in the `/backend` directory. There are optional frontend apps in `/frontend` and sample React Vite apps in `/samples/apps`.

## Tech Stack

### Backend

- Go (Golang) latest stable version is used.
- Recommended database is PostgreSQL. SQLite is by default packaged for testing and local development purposes.

### Frontend

- React with Vite and TypeScript is used for the frontend.
- Frontend consists of two parts: 
  - `thunder-gate`: The login, registration and recovery UI.
  - `thunder-develop`: The main console application.

### Testing

- `stretchr/testify` is used to write unit tests.
- `mockery` is used to generate mocks for unit tests.
- `DATA-DOG/go-sqlmock` is used to mock database operations in unit tests.

## Project Structure

- backend/: Server backend implementation.
  - cmd/server/: Main server application.
    - repository/: Configurations and other resource files.
  - dbscripts/: Database scripts.
  - internal/: Internal packages for various functionalities.
    - {package-name}/: Code for the {package-name}.
    - system/: Common utilities, services, and configurations.
  - scripts/: Utility scripts such as init scripts.
  - tests/: Common unit tests related files.
    - mocks/: Generated mocks for unit tests.
    - resources/: Test resource files.
  - .mockery.private.yml: Mockery configurations for private interfaces.
  - .mockery.public.yml: Mockery configurations for public interfaces.
- frontend/: Individual frontend application code.
  - apps/thunder-gate/: UIs for login, registration and recovery. 
  - apps/thunder-develop/: UIs for the console application.
  - packages/: Common frontend packages such as UI components, services, and contexts.
- install/helm/: Helm charts for deployment.
- samples/apps/: Sample applications demonstrating the usage of the product.
- tests/integration/: Integration tests for the backend.
- tests/e2e/: End-to-end tests for the product.
- docs/: Documentation related files.
  - contributing/: Contribution guidelines.
  - guides/: Product documentation files.

## General Guidelines

- Follow general coding best practices, design patterns, and security recommendations.
- Ensure all identity-related code aligns with relevant RFC specifications.
- Follow https://wso2.com/whitepapers/wso2-rest-apis-design-guidelines/ for RESTful API design.
- Follow https://security.docs.wso2.com/en/latest/security-guidelines/secure-engineering-guidelines/secure-coding-guidlines/general-recommendations-for-secure-coding/ for secure coding practices.
- Promote code reusability and define constants where applicable.
- Ensure proper error handling and logging.
- Write unit tests to achieve at least 80% coverage and integration tests where applicable.
- Go source file names must follow snake_case (for example, `error_constants.go`)
- Pure typescript source file names (i.e with the extension .ts) must follow camelCase (for example, `useCreateApplication.ts`)
- React component source file names (i.e with the extension .tsx) must follow PascalCase (for example, `MenuButton.tsx`)

- Refer project README for other general instructions such as build the product, run the server, run tests, etc.

## Backend Specific Guidelines

### General
- Reuse common utilities from the `internal/system` packages.
- Define interfaces for services to enable dependency injection and testability.

### Package Structure and Organization
- Follow a modular package structure where each domain/feature lives in its own package under `internal/`.
- Follow a flat directory structure within a package. Avoid nested packages unless absolutely necessary for complex domains.
- Each domain package typically contains related components organized by responsibility (not all files are required):
  - `service.go`: Service interface and implementation (business logic layer)
  - `handler.go`: HTTP handlers (presentation layer) - only if the package exposes HTTP endpoints
  - `store.go`: Data access layer (persistence) - only if the package needs database operations
  - `model.go`: Domain models and DTOs - only if the package has domain-specific models
  - `constants.go`: Package-specific constants (e.g., default values, configuration constants, business logic constants)
  - `error_constants.go`: Define service and API error messages, error codes, and error-related constants
  - `store_constants.go`: Define database queries, table names, column names, and database-related constants
  - `utils.go`: Define package-specific utility functions in this file
  - `init.go`: Package initialization and route registration. Route registration is only required if the package exposes HTTP endpoints.
  - `file_based_store.go`: File-based storage implementation - only if the package needs to support declarative configs
  - `composite_store.go`: Composite storage implementation - only if the package needs to support both database and declarative configs simultaneously
  - `declarative_resource.go`: Declarative resource implementation - only if the package needs to support declarative configs
- Adjust the file structure based on actual requirements. For example:
  - No HTTP layer? Skip `handler.go`
  - Need cache-backed storage? Add additional storage implementation files
  - Declarative resource support not required? Skip `file_based_store.go` and `composite_store.go` and `declarative_resource.go`
  - Complex domain? Use subdirectories for further organize related functionality (e.g., `internal/oauth/oauth2/`, `internal/oauth/jwks/`).

### Package Exports
- Only export the service interface (e.g., `XServiceInterface`) and models that are used in the service interface from a package.
- Keep all internal implementations (service structs, store interfaces, store implementations, handlers) unexported (lowercase).
- Keep internal constants such as database queries, error codes, and other implementation details unexported (private).
- This ensures proper encapsulation and prevents external packages from depending on internal implementation details.
- Example: Export `UserServiceInterface` and `User` model, but keep `userService`, `userStore`, `userHandler`, and internal query constants unexported.

### Logging
- Use the `log` package in `internal/system` for logging.
- Add minimal info logs and ensure server errors are logged for debugging.
- Avoid logging PII. Use `MaskString` from `internal/system/log` to mask sensitive information.
- Add debug logs where necessary, but avoid excessive logging.
- Use `IsDebugEnabled` from `internal/system/log` if excessive handling is done for debugging log construction.

### Database
- Use `DBClient` in `internal/system/database` for database operations.
- Use `DBQuery` from `internal/system/database/model` to define queries with a unique ID. This allows for DB-specific queries where needed.
  - Define each query with a unique identifier for traceability
  - Support database-specific query variations when necessary (e.g., SQLite vs PostgreSQL)
- All tables must support multi-deployment scenarios by including `DEPLOYMENT_ID` as part of the composite primary key
  - When defining SQL queries, always include `DEPLOYMENT_ID` and follow these patterns:
    - **INSERT queries**: Add `DEPLOYMENT_ID` as the **last column** in the column list and the **last parameter** in VALUES
      - Example: `INSERT INTO TABLE_NAME (COL1, COL2, ..., DEPLOYMENT_ID) VALUES ($1, $2, ..., $N)`
    - **UPDATE queries**: Add `AND DEPLOYMENT_ID = $N` to the WHERE clause, where `$N` is the **last parameter**
      - Example: `UPDATE TABLE_NAME SET COL1 = $2, COL2 = $3 WHERE ID = $1 AND DEPLOYMENT_ID = $4`
    - **DELETE queries**: Add `AND DEPLOYMENT_ID = $N` to the WHERE clause, where `$N` is the **last parameter**
      - Example: `DELETE FROM TABLE_NAME WHERE ID = $1 AND DEPLOYMENT_ID = $2`
    - **SELECT queries**: Add `AND DEPLOYMENT_ID = $N` to the WHERE clause for single-table queries
      - Example: `SELECT * FROM TABLE_NAME WHERE ID = $1 AND DEPLOYMENT_ID = $2`
    - **JOIN queries**: Include `DEPLOYMENT_ID` in JOIN conditions and WHERE clauses
      - Example: `SELECT * FROM T1 LEFT JOIN T2 ON T1.ID = T2.ID AND T1.DEPLOYMENT_ID = T2.DEPLOYMENT_ID WHERE T1.ID = $1 AND T1.DEPLOYMENT_ID = $2`
  - Always ensure `DEPLOYMENT_ID` is the last parameter in the query to maintain consistency across the codebase

### Store Layer (Data Access)
- Define store interfaces (e.g., `xStoreInterface`) and implementations (e.g., `xStore` struct) in `store.go`.
- Store layer handles all database interactions and should be used by the service layer.
- Use private constructors (e.g., `newXStore()`) to create store instances.
- Store initialization should use `DBProvider` to obtain database client. Individual store methods should use the created client.
- Keep store methods focused on data access operations without business logic.

### HTTP
- Use `HTTPClient` in `internal/system/http` for sending external requests.

### Cache
- Extend `BaseCache` in `internal/system/cache` for caching requirements.

### Config
- Use `ThunderRuntime` in `internal/system/config` to read system configs.

### Server Constants
- Use constants defined in `internal/system/constants` for reusable global values.

### Error Handling
- Use `ServiceError` from `internal/system/error/serviceerror` to return errors from service layer.
- Use `ErrorResponse` from `internal/system/error/apierror` to define and return API layer errors.
- Avoid logging the same error twice. Return a Go error or `ServiceError` from internal components and log at the service layer.
- Avoid returning unnecessary details from the API layer for server-side errors. Log and return a generic message like "Internal server error" or "Something went wrong" where applicable.

### Defining APIs
- Return JSON responses from APIs where applicable.
- Return JSON errors as per the server `ErrorResponse` definition. For 500 internal server errors, a generic message may be returned.
- Define API handlers in a `handler.go` file within the domain package.
- For packages with HTTP endpoints, use an `init.go` file to register routes with the mux and initialize dependencies.
- Define CORS policies using `middleware.WithCORS` from `internal/system/middleware` where applicable.

### Service Layer and Dependency Injection
- Define service interfaces (e.g., `XServiceInterface`) and implementations (e.g., `xService` struct) in `service.go`.
- Use private constructor functions (e.g., `newXService()`) to create service instances.
- If the service needs to interact with the database, accept the store interface as a parameter in the constructor.
- Constructor functions should accept all dependencies as parameters when the service needs external dependencies. The parameter should be in the following order: service dependencies, store, transactioner.
  - Example without dependencies: `func newIDPService() IDPServiceInterface`
  - Example with dependencies: `func newGroupService(ouService OrganizationUnitServiceInterface, store GroupStoreInterface, transactioner Transactioner) GroupServiceInterface`
- Services should depend on interfaces, not concrete implementations, to enable testing with mocks.
- Keep constructors private (unexported) - external packages should only interact through the `Initialize()` function.

### Service Initialization and Dependency Management
- Service initialization should happen **once** during application startup in the `init.go` file of each package.
- The `Initialize(mux, deps)` function in `init.go` should:
  1. Create the store instances using the constructor functions (Only if the package requires database operations)
  2. Create the service instances using the constructor functions, passing created store and required dependencies which are passed in as parameters(`deps`). If there are multiple services in the package, initialize all services according to their dependency requirements.
     - **Note**: If a service supports declarative configs, it should return a `ResourceExporter` during initialization.
  3. Create handlers and inject the service instance into them
  4. Register routes with the mux
  5. Return the created service interfaces (and `ResourceExporter` if applicable) to be used as dependencies by other packages
- Keep all initialized service instances and pass them to dependent services during their initialization.
- Example initialization flow:
  ```go
  // In internal/user/init.go
  func Initialize(mux *http.ServeMux, ouService ou.OrganizationUnitServiceInterface) (UserServiceInterface, error) {
      userStore, err := newUserStore() // No dependencies
      if err != nil {
          return nil, err
      }
      // Obtain transactioner from the provider
      transactioner, err := provider.GetDBProvider().GetUserDBTransactioner()
      if err != nil {
          return nil, err
      }

      userService := newUserService(ouService, userStore, transactioner) // Inject dependency via private constructor
      userHandler := newUserHandler(userService)
      registerRoutes(mux, userHandler)
      return userService, nil // Return the created service interfaces for dependency injection
  }
  ```
- The main service manager in `cmd/server/servicemanager.go` should orchestrate all initializations in the correct order, passing dependencies as needed.
  - It should also collect `ResourceExporter` interfaces from packages that support declarative configs and register them with the `export` service.
  ```go
  // In cmd/server/servicemanager.go
  // Note: Error handling is not shown for brevity.
  func registerServices(mux *http.ServeMux) {
      var exporters []immutableresource.ResourceExporter

      ouService, ouExporter, err := ou.Initialize(mux) // Returns service and exporter

      exporters = append(exporters, ouExporter)
      
      userService, err := user.Initialize(mux, ouService) // Pass dependencies
      
      // ... initialize other services ...

      // Initialize export service with collected exporters
      _ = export.Initialize(mux, exporters)
  }
  ```

### Transaction Management
- Use `transaction.Transactioner` from `internal/system/database/transaction` to handle database transactions within the service layer.
- Inject the `transactioner` into the service struct.
- Use `transactioner.Transact` to execute operations that require atomicity.
  - `Transact` accepts a context and a function. The function receives a new context `txCtx` which **must** be passed to all store operations within the closure to ensure they participate in the transaction.
  - If the closure returns an error, the transaction is automatically rolled back. If it returns `nil`, the transaction is committed.
- Example usage:
  ```go
  // In internal/user/service.go
  func (us *userService) CreateUser(ctx context.Context, user *User) (*User, *serviceerror.ServiceError) {
      // ... validation logic ...
  
      // Use transaction to ensure atomic user creation
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

### Testing

#### Unit Tests
- Ensure unit tests are written to achieve at least 80% coverage.
- Use `stretchr/testify` for tests and follow the test suite pattern.
- `mockery` is used to generate mocks; configurations for private and public interfaces are in `.mockery.private.yml` and `.mockery.public.yml` respectively. Mocks can be generated using `make mockery` command from the project root directory.
- Place generated mocks in the `/backend/tests/mocks/` directory.
- Unit tests can be run using `make test_unit` command from the project root directory. Alternatively `go test` command can also be used from the `/backend` directory with applicable flags.

#### Integration Tests
- Write integration tests in the `/tests/integration/` directory where applicable.
- Add unit and integrations tests for each new feature or bug fix to achieve a combined coverage of at least 80%.
- Integration tests can be run using `make all` command from the project root directory. This will build the project, package into a zip, unzip in a temp directory, and run the integration tests. So it will take some time to complete. Integration tests can be run on an already built product by executing the `make test_integration` command from the project root directory.

### Documentation
- Ensure applicable changes are documented in the `README` file or `/docs/content/` directory.
- Ensure each new feature or API is documented.
- Add Swagger definitions for the APIs to `/docs/apis/`.
