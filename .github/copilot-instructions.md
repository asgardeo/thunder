# GitHub Copilot Custom Instructions

## Project Overview

This repository is a lightweight user and identity management product written in Go. The primary focus is to provide authentication and authorization capabilities for applications and also to allow managing users, roles, and permissions. The server allows three different approaches to authenticate users: by using standard protocols such as OAuth2, OIDC, by defining a flexible orchestration flow, or by using individual authentication mechanisms such as password, passwordless, and social login.

The project is structured as a monorepo to manage the backend, frontend, and sample applications in a single repository. The backend code is located in the `/backend` directory. There is an optional Next.js frontend in `/frontend` and a sample React Vite app in `/samples/apps`.

## Tech Stack

### Backend

- Go (Golang) latest stable version is used.
- Recommended database is PostgreSQL. SQLite is by default packaged for testing and local development purposes.

### Frontend

- Next.js with TypeScript is used for the frontend.
- React with Vite and TypeScript is used for the sample app.

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
    - authn/: Individual authentication-related code.
    - oauth/: OAuth related codes such as OAuth, OAuth2, AuthZ, OIDC, etc.
    - flow/: Flow orchestration engine and related code.
    - executor/: Individual flow executor implementations.
    - system/: Common utilities, services, and configurations.
  - scripts/: Utility scripts such as init scripts.
  - tests/: Common unit tests related files.
    - mocks/: Generated mocks for unit tests.
    - resources/: Test resource files.
  - .mockery.private.yml: Mockery configurations for private interfaces.
  - .mockery.public.yml: Mockery configurations for public interfaces.
- frontend/: Individual frontend application code.
  - apps/gate/: Gate app implementation which serves UIs for login, registration and recovery. 
  - packages/: Common frontend packages such as UI components, services, and contexts.
- install/helm/: Helm charts for deployment.
- samples/apps/: Sample applications demonstrating the usage of the product.
  - oauth/: Sample React Vite app implementing authentication with OAuth2 and flow execution APIs.
- tests/integration/: Integration tests for the backend.
- docs/: Documentation related files.
  - apis/: Swagger definitions for APIs.
  - content/: Other documentation files.

## General Guidelines

- Follow general coding best practices, design patterns, and security recommendations.
- Ensure all identity-related code aligns with relevant RFC specifications.
- Follow https://wso2.com/whitepapers/wso2-rest-apis-design-guidelines/ for RESTful API design.
- Follow https://security.docs.wso2.com/en/latest/security-guidelines/secure-engineering-guidelines/secure-coding-guidlines/general-recommendations-for-secure-coding/ for secure coding practices.
- Promote code reusability and define constants where applicable.
- Ensure proper error handling and logging.
- Write unit tests to achieve at least 80% coverage and integration tests where applicable.
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
  - `errorconstants.go`: Define service and API error messages, error codes, and error-related constants
  - `storeconstants.go`: Define database queries, table names, column names, and database-related constants
  - `utils.go`: Define package-specific utility functions in this file
  - `init.go`: Package initialization and route registration - only for packages with HTTP endpoints
- Adjust the file structure based on actual requirements. For example:
  - No HTTP layer? Skip `handler.go` and `init.go`
  - File-based or cache-backed storage? Add additional storage implementation files
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
- Constructor functions should accept all dependencies as parameters when the service needs external dependencies.
  - Example without dependencies: `func newIDPService() IDPServiceInterface`
  - Example with dependencies: `func newGroupService(ouService OrganizationUnitServiceInterface) GroupServiceInterface`
- Services should depend on interfaces, not concrete implementations, to enable testing with mocks.
- Keep constructors private (unexported) - external packages should only interact through the `Initialize()` function.

### Service Initialization and Dependency Management
- Service initialization should happen **once** during application startup in the `init.go` file of each package.
- The `Initialize(mux, deps)` function in `init.go` should:
  1. Create the store instances using the constructor functions (Only if the package requires database operations)
  2. Create the service instances using the constructor functions, passing created store and required dependencies which are passed in as parameters(`deps`). If there are multiple services in the package, initialize all services according to their dependency requirements.
  3. Create handlers and inject the service instance into them
  4. Register routes with the mux
  5. Return the created service interfaces to be used as dependencies by other packages
- Keep all initialized service instances and pass them to dependent services during their initialization.
- Example initialization flow:
  ```go
  // In internal/user/init.go
  func Initialize(mux *http.ServeMux, ouService ou.OrganizationUnitServiceInterface) UserServiceInterface {
      userStore := newUserStore() // No dependencies
      userService := newUserService(ouService, userStore) // Inject dependency via private constructor
      userHandler := newUserHandler(userService)
      registerRoutes(mux, userHandler)
      return userService // Return the created service interfaces for dependency injection
  }
  ```
- The main service manager in `cmd/server/servicemanager.go` should orchestrate all initializations in the correct order, passing dependencies as needed.
  ```go
  // In cmd/server/servicemanager.go
  func registerServices(mux *http.ServeMux) {
      ouService := ou.Initialize(mux) // No dependencies
      userService := user.Initialize(mux, ouService) // Pass dependencies
      groupService := group.Initialize(mux, ouService, groupService) // Pass dependency
      ...
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
