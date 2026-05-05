# pkg/ — ThunderID Public Packages

This directory contains packages that are intentionally importable by external Go modules.
Packages under `internal/` are restricted to this module; packages here are not.

---

## Why packages were moved here

Go's `internal/` directory rule prevents external modules from importing those packages.
ThunderID's OAuth 2.0 / OIDC protocol layer, flow engine, application model, and authentication
provider stack are battle-tested implementations that external adopters should be able to reuse
without forking the entire repository.

The goal is to let an external module do this:

```go
import (
    appbuilder "github.com/asgardeo/thunder/pkg/app"
    "github.com/asgardeo/thunder/pkg/entityprovider"
    "github.com/asgardeo/thunder/pkg/flow/core"
)

app, err := appbuilder.New().
    WithEntityProvider(myEntityProvider).   // plug in your own user store
    WithAuthnProvider(myAuthnProvider).     // plug in your own authn stack
    WithConsentService(myConsentService).   // plug in your own consent logic
    WithExecutor("my_step", myExecutor).    // register a custom flow executor
    WithObservabilityService(myObs).        // plug in your own observability (OTEL, etc.)
    WithEmailClient(myEmailClient).         // plug in your own email provider
    Build(mux)
```

Everything not injected falls back to ThunderID's built-in defaults.

---

## What lives here and why

### `pkg/app` — Orchestrator / builder

The entry point for external consumers. `Builder` wires all services together and returns a
`ThunderApp` that exposes:

- `Mux()` — the HTTP multiplexer with all routes registered
- `SecurityMiddleware()` — the pre-wired JWT security middleware as a standard
  `func(http.Handler) http.Handler`
- `Shutdown()` — graceful teardown of background services

The builder accepts optional dependency overrides via `With*` methods. Dependencies that are
not overridden default to ThunderID's standard implementations.

### `pkg/oauth` — OAuth 2.0 / OIDC protocol layer

The complete OAuth 2.0 / OIDC server implementation: authorization, token issuance,
introspection, userinfo, JWKS, PAR, DCR, and discovery. This is the most
reuse-valuable package — adopters get a fully RFC-compliant protocol layer without
reimplementing it.

### `pkg/flow` — Flow engine

The orchestration engine that powers ThunderID's login sequences. Sub-packages:

| Package | Role |
|---|---|
| `flow/core` | `ExecutorInterface`, flow graph model |
| `flow/executor` | Built-in executor implementations (password, OTP, passkey, social, etc.) |
| `flow/mgt` | Flow definition CRUD and MCP tool registration |
| `flow/flowexec` | Runtime flow execution |
| `flow/flowmeta` | Flow metadata for login page rendering |
| `flow/common` | Shared types |

External consumers can add custom steps by implementing `ExecutorInterface` and passing them
via `WithExecutor`.

### `pkg/application` — Application management

CRUD and business logic for OAuth clients / applications registered in ThunderID.

### `pkg/inboundclient` — Inbound client resolution

Resolves an inbound OAuth client from a request, validates its credentials, and provides the
flow/theme/layout configuration bound to that client.

### `pkg/authnprovider` — Authentication provider

The interface and default implementations for authenticating a principal during a flow step.

| Sub-package | Role |
|---|---|
| `authnprovider/provider` | `AuthnProviderInterface` + default (password, passkey, OTP) |
| `authnprovider/manager` | `AuthnProviderManagerInterface` — routes step requests to providers |
| `authnprovider/common` | Shared request/response types |

### `pkg/entityprovider` — Entity provider

`EntityProviderInterface` defines how ThunderID looks up and persists identity entities
(users, service accounts). Implement this to connect an external user store.

Default implementations are included:
- `InitializeEntityProvider` — reads/writes from ThunderID's own database
- `DisabledEntityProvider` — rejects all lookups (useful when an external IdP owns all identities)

### `pkg/consent` — Consent service

`ConsentServiceInterface` defines how the server records and checks consent grants. Implement
this to store consent in an external system.

---

## The two rules that govern what goes where

**Rule 1: domain / protocol logic → `pkg/`**

If an external adopter would reasonably want to reuse an implementation as-is, or override it
with their own, it belongs in `pkg/`. This covers the OAuth protocol, the flow engine, the
application model, and the extension-point interfaces (entity provider, authn provider,
consent, executors).

**Rule 2: infrastructure → `internal/system/`**

Low-level plumbing that is tightly coupled to ThunderID's specific configuration format
(key paths in `deployment.yaml`, database connection parameters, PKI certificate layout)
stays in `internal/system/`. Examples: JWT service, JWE service, PKI, database provider,
observability, email client, i18n, MCP server.

External adopters replace infrastructure by implementing the corresponding port interface, not
by importing Thunder's implementation.

---

## Port interfaces — the bridge for infrastructure

Each `pkg/` package that depends on an infrastructure service declares a minimal port interface
(`ports.go`) covering only the methods it actually calls. The `internal/` implementations
satisfy these ports structurally (Go duck-typing — no changes needed on the infrastructure
side). The builder in `pkg/app` wires the concrete implementations to the port interfaces at
startup.

Example: `pkg/oauth/ports.go` declares `JWTServicePort` with five methods. The oauth package
depends only on that interface — it never imports `internal/system/jose/jwt` directly.

### Current limitation

Some port interface signatures still reference `internal/` data-model types
(`ou.OrganizationUnit`, `authz.GetAuthorizedPermissionsRequest`, `resource.ResourceServer`,
`event.Event`, `email.EmailData`). These types are simple value structs with no internal
dependencies of their own. Moving them to `pkg/` is the planned next step; until then,
external consumers who want to implement a port that uses these types must import the
corresponding `internal/` package.

---

## Roadmap

The following `internal/` packages are candidates to move to `pkg/` in subsequent phases,
which will eliminate the limitation above and make the full port surface externally
implementable:

1. `internal/ou` → `pkg/ou` — unlocks `OUServicePort` in `pkg/oauth/ports.go`
2. `internal/authz` → `pkg/authz` — unlocks `AuthZServicePort`
3. `internal/resource` → `pkg/resource` — unlocks `ResourceServicePort`
4. `internal/attributecache` → `pkg/attributecache` — unlocks `AttributeCacheServicePort`
5. `internal/entity` → `pkg/entity` — core identity model
6. `internal/idp` → `pkg/idp` — identity provider management
7. `internal/userschema`, `internal/group`, `internal/role`, `internal/user` — user management layer
8. `internal/authn` (and sub-packages) — full authentication handler stack
