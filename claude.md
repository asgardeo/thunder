# CLAUDE.md - Thunder Identity Management Suite

> **For AI Assistants**: This document helps you understand and work with the Thunder project effectively.

## 🎯 Project Overview

**Thunder** is a modern, open-source Identity and Access Management (IAM) solution built by WSO2. It's designed for teams building secure, customizable authentication experiences across applications, services, and AI agents.

**Key Characteristics:**
- **Language**: Go (backend) + React + TypeScript (frontend)
- **Purpose**: OAuth 2.0/OIDC identity provider with flexible flow designer
- **Architecture**: Microservices-ready, containerized deployment, monorepo with Nx
- **License**: Apache 2.0
- **Repository**: https://github.com/asgardeo/thunder

## 🏗️ Architecture & Tech Stack

### Backend
- **Language**: Go 1.24+ (1.25+ for builds)
- **Standards**: OAuth 2.0, OpenID Connect (OIDC)
- **Supported Flows**: Client Credentials, Authorization Code, Refresh Token
- **Database**: SQLite (default) / PostgreSQL (configurable)
- **Server**: Runs on `https://localhost:8090`

### Frontend Architecture

Thunder uses a **monorepo architecture** managed by **Nx** (v21.6+) with multiple React applications and shared packages.

#### Core Technologies
- **Build Tool**: Vite 6.4+ (some apps use Rolldown-Vite 7.1+)
- **Language**: TypeScript 5.8+
- **Framework**: React 19.2+
- **Package Manager**: pnpm (with workspace support)
- **Monorepo Tool**: Nx 21.6+
- **Styling**: Emotion (@emotion/react, @emotion/styled)
- **UI Library**: Material-UI (MUI) 7.1+ & WSO2 Oxygen UI

#### Frontend Applications

**1. Developer Console (`thunder-develop`)**
- **Path**: `frontend/apps/thunder-develop/`
- **Purpose**: Admin interface for managing applications, users, flows, and identity providers
- **URL**: `https://localhost:8090/develop`
- **Login**: `admin` / `admin` (default)
- **Key Features**:
  - Visual flow designer for authentication/registration flows (using @xyflow/react)
  - Application management (OAuth2/OIDC client configuration)
  - User management
  - Identity provider configuration (Google, GitHub, SMS OTP)
  - Rich text editing with Lexical editor
  - Drag-and-drop flow orchestration (@dnd-kit)
  - Form validation with React Hook Form + Zod
  - State management with @tanstack/react-query
  - Internationalization with i18next

**Tech Stack (Developer Console)**:
```json
{
  "ui": ["@mui/material", "@wso2/oxygen-ui"],
  "routing": ["@asgardeo/react-router", "react-router"],
  "auth": ["@asgardeo/react"],
  "flow-design": ["@xyflow/react", "@dnd-kit/*"],
  "editor": ["lexical"],
  "forms": ["react-hook-form", "@hookform/resolvers", "zod"],
  "data": ["@tanstack/react-query"],
  "i18n": ["i18next", "react-i18next"],
  "utilities": ["lodash-es", "dayjs", "dompurify"]
}
```

**2. Thunder Gate (`thunder-gate`)**
- **Path**: `frontend/apps/thunder-gate/`
- **Purpose**: End-user authentication interface (login/registration flows)
- **URL**: Served by backend at runtime

**3. Sample Applications**

**React Vanilla Sample** (`samples/apps/react-vanilla-sample/`)
- Direct API integration without SDK
- Supports both Native Flow API and OAuth/OIDC redirect flows
- Built with React 19, Vite, Material-UI
- Configuration via `.env` (dev) or `runtime.json` (production)
- Demonstrates `axios` for API calls
- Runs on `https://localhost:3000`

**React SDK Sample** (`samples/apps/react-sdk-sample/`)
- SDK-based integration using `@asgardeo/react`
- Standard OAuth 2.0/OIDC flow
- Uses WSO2 Oxygen UI components
- Built with Rolldown-Vite for faster builds
- Runs on `https://localhost:3000`

#### Frontend Workspace Structure
```
frontend/
├── apps/
│   ├── thunder-develop/        # Developer Console (Admin UI)
│   └── thunder-gate/            # Authentication UI (Login/Register)
├── packages/                    # Shared libraries
│   ├── thunder-logger/          # Logging utility
│   ├── thunder-eslint-plugin/   # Custom ESLint rules
│   ├── thunder-test-utils/      # Testing utilities
│   ├── thunder-shared-contexts/ # React contexts
│   ├── thunder-shared-hooks/    # Custom React hooks
│   ├── thunder-shared-branding/ # Branding/theming
│   └── thunder-i18n/            # Internationalization
├── scripts/                     # Build scripts
├── pnpm-workspace.yaml          # pnpm workspace config
├── nx.json                      # Nx configuration
├── package.json                 # Root package.json
└── README.md                    # Frontend docs
```

### Authentication Features
- Username/Password (Basic Auth)
- Social Login: Google, GitHub
- SMS OTP
- Self-registration with similar methods
- Custom authentication flows via visual designer

## 📁 Directory Structure

```
thunder/
├── backend/                      # Go backend source
│   ├── cmd/server/               # Main server entry point
│   │   └── repository/conf/      # Configuration files (deployment.yaml)
│   ├── pkg/                      # Shared packages
│   └── internal/                 # Internal application code
├── frontend/                     # Frontend monorepo (Nx workspace)
│   ├── apps/
│   │   ├── thunder-develop/      # Developer Console (Admin UI)
│   │   └── thunder-gate/         # Authentication UI
│   ├── packages/                 # Shared React packages
│   ├── pnpm-workspace.yaml       # pnpm workspace config
│   └── nx.json                   # Nx monorepo config
├── samples/apps/                 # Sample React applications
│   ├── react-vanilla-sample/     # Direct API integration sample
│   │   ├── .env.example          # Configuration template
│   │   ├── app/runtime.json      # Runtime config (for builds)
│   │   └── server/               # Express server for hosting
│   └── react-sdk-sample/         # SDK-based sample (@asgardeo/react)
├── install/
│   ├── quick-start/              # Docker Compose files
│   └── local-development/        # PostgreSQL setup
├── docs/                         # Documentation
├── resources/images/             # Screenshots and assets
├── Makefile                      # Build and run commands
└── README.md                     # Main documentation
```

## 🚀 Common Workflows

### Quick Start (User Testing)
1. **Download release** from GitHub releases
2. **Run setup**: `./setup.sh` (or `.ps1` on Windows) — generates sample app ID
3. **Start server**: `./start.sh` — runs on https://localhost:8090
4. **Access console**: https://localhost:8090/develop (admin/admin)
5. **Run sample app**: Configure with sample app ID, start, test at https://localhost:3000

### Development Workflow - Backend
```bash
# Clone and start backend
git clone https://github.com/asgardeo/thunder
cd thunder
make run  # Starts on https://localhost:8090
```

### Development Workflow - Frontend (Developer Console)
```bash
# From project root
cd frontend

# Install dependencies
pnpm install

# Start Developer Console in dev mode
pnpm --filter @thunder/develop dev

# Or using Nx directly
nx serve thunder-develop

# Build all frontend apps
pnpm build

# Run tests across all apps
pnpm test

# Lint all code
pnpm lint
```

### Development Workflow - Sample Apps
```bash
# React Vanilla Sample
cd samples/apps/react-vanilla-sample
cp .env.example .env
# Edit .env with VITE_REACT_APP_AUTH_APP_ID and endpoints
npm install
npm run dev  # Starts on http://localhost:3000

# React SDK Sample
cd samples/apps/react-sdk-sample
npm install
npm run dev  # Starts on https://localhost:3000 (with SSL)
```

### Frontend Configuration Files

**Developer Console** (runtime, embedded in backend):
- No external config needed
- Served by Go backend at `/develop`
- Uses OAuth redirect flow with backend

**Sample Apps** (standalone):

`.env` file (development):
```env
# Application ID from setup.sh output
VITE_REACT_APP_AUTH_APP_ID={your-application-id}

# Thunder endpoints
VITE_REACT_APP_SERVER_FLOW_ENDPOINT=https://localhost:8090/flow
VITE_REACT_APPLICATIONS_ENDPOINT=https://localhost:8090/applications

# Login mode: false = native flow API, true = OAuth redirect
VITE_REACT_APP_REDIRECT_BASED_LOGIN=false

# For OAuth redirect mode (additional):
VITE_REACT_APP_CLIENT_ID={client-id}
VITE_REACT_APP_CLIENT_SECRET={client-secret}
VITE_REACT_APP_SERVER_AUTHORIZATION_ENDPOINT=https://localhost:8090/oauth2/authorize
VITE_REACT_APP_SERVER_TOKEN_ENDPOINT=https://localhost:8090/oauth2/token
```

`runtime.json` file (production builds):
```json
{
  "applicationID": "{your-application-id}"
}
```

### Build from Source
```bash
# Full build (backend + frontend + samples)
make all

# Backend only
make build_backend

# Frontend only
make build_frontend
```

### Testing
```bash
# Backend tests
make test_unit                       # Unit tests only
make test_integration                # Integration tests (requires built product)
make build_backend test_integration  # Build + integration tests
make build_with_coverage             # Build with coverage, run tests, generate reports

# Frontend tests
cd frontend
pnpm test                            # Run all tests
pnpm test:watch                      # Watch mode
nx test thunder-develop              # Test specific app
nx test thunder-develop --coverage   # With coverage
```

### Debugging

**Backend Debugging:**
```bash
# Start in debug mode (requires Delve)
./start.sh --debug  # Listens on localhost:2345

# VS Code: Use .vscode/launch.json and press F5
# GoLand: Run → Edit Configurations → Go Remote (127.0.0.1:2345)
```

**Frontend Debugging:**
- Use browser DevTools
- React DevTools extension recommended
- Nx Console extension for VS Code
- Vite HMR provides instant feedback

## 🔑 Key API Flows

### 1. Authentication Flow (Native API)
```bash
# Step 1: Initialize flow
curl -k -X POST 'https://localhost:8090/flow/execute' \
  -d '{"applicationId":"<app_id>","flowType":"AUTHENTICATION"}'
# Returns: {"flowId":"<flow_id>","flowStatus":"INCOMPLETE",...}

# Step 2: Submit credentials
curl -k -X POST 'https://localhost:8090/flow/execute' \
  -d '{"flowId":"<flow_id>","inputs":{"username":"admin","password":"admin"},"action":"action_001"}'
# Returns: {"flowId":"<flow_id>","flowStatus":"COMPLETE","assertion":"<token>",...}
```

### 2. System API Token (for admin operations)
Use the authentication flow above with `"requested_permissions":"system"` in inputs.

### 3. Client Credentials Flow (M2M)
```bash
# Step 1: Create OAuth2 client app (requires system token)
curl -kL -X POST https://localhost:8090/applications \
  -H "Authorization: Bearer <system_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "M2M Client",
    "inbound_auth_config": [{
      "type": "oauth2",
      "config": {
        "client_id": "<your_client_id>",
        "client_secret": "<your_secret>",
        "grant_types": ["client_credentials"],
        "scopes": ["api:read", "api:write"]
      }
    }]
  }'

# Step 2: Get access token
curl -k -X POST https://localhost:8090/oauth2/token \
  -d 'grant_type=client_credentials' \
  -u '<client_id>:<client_secret>'
```

### 4. OAuth2/OIDC Endpoints
- **Authorization**: `/oauth2/authorize`
- **Token**: `/oauth2/token`
- **Flow Execution**: `/flow/execute`
- **Applications**: `/applications` (requires system token)
- **Users**: `/users` (requires system token)
- **Identity Providers**: `/identity-providers` (requires system token)
- **Developer Console**: `/develop` (web UI)

## 🗄️ Database Configuration

### Default (SQLite)
Configured out-of-the-box, suitable for development.

### PostgreSQL (Production)
```yaml
# backend/cmd/server/repository/conf/deployment.yaml
database:
  identity:
    type: "postgres"
    hostname: "localhost"
    port: 5432
    name: "thunderdb"
    username: "asgthunder"
    password: "asgthunder"
    sslmode: "disable"
  runtime:
    type: "postgres"
    # ... similar config
  user:
    type: "postgres"
    # ... similar config
```

```bash
# Start PostgreSQL via Docker
cd install/local-development
docker compose up -d

# Run Thunder with PostgreSQL
make run
```

## 🎨 Frontend Development Tips

### Working with the Developer Console

**Key Features to Understand:**
1. **Flow Designer**: Visual drag-and-drop interface for creating authentication flows
   - Uses `@xyflow/react` for flow visualization
   - Nodes represent authentication steps (username/password, social login, MFA)
   - Edges represent flow transitions

2. **Application Management**: Configure OAuth2/OIDC clients
   - Set redirect URIs, grant types, scopes
   - Generate client credentials
   - Test with sample apps

3. **User Management**: CRUD operations for users
   - Create users manually or via self-registration
   - View user attributes and authentication history

4. **Identity Providers**: Configure social logins
   - Google, GitHub OAuth configurations
   - SMS OTP provider settings

### Frontend Architecture Patterns

**Shared Packages** (in `frontend/packages/`):
- `@thunder/shared-contexts`: React context providers
- `@thunder/shared-hooks`: Custom React hooks (auth, API calls, etc.)
- `@thunder/shared-branding`: Theming and branding utilities
- `@thunder/i18n`: Internationalization resources
- `@thunder/logger`: Logging abstraction
- `@thunder/test-utils`: Testing utilities and mocks

**State Management:**
- `@tanstack/react-query` for server state
- React Context for app state
- React Hook Form for form state

**Styling Approach:**
- Emotion for CSS-in-JS
- Material-UI components with custom theming
- WSO2 Oxygen UI design system
- SASS for additional styling

### Common Frontend Tasks

**Add a new feature to Developer Console:**
1. Create feature components in `frontend/apps/thunder-develop/src/`
2. Use shared hooks from `@thunder/shared-hooks`
3. Integrate with React Query for API calls
4. Add i18n keys to `@thunder/i18n`
5. Write tests using `@thunder/test-utils`

**Create a custom authentication flow:**
1. Open Developer Console → Flows
2. Use visual designer to add/connect nodes
3. Configure node properties (authenticators, validators)
4. Test with sample app

**Customize sample app UI:**
1. Modify components in `samples/apps/react-vanilla-sample/src/`
2. Use Material-UI components for consistency
3. Update `.env` for configuration
4. Hot reload with `npm run dev`

## 🛠️ Tips for AI Assistants

### When helping with Thunder development:

1. **Authentication Context**: Users often need help with OAuth flows. Thunder supports both native flow APIs (simpler, app-native) and standard OAuth redirect flows.

2. **Sample App ID**: Always generated during `setup.sh` — users must note this value for sample app configuration.

3. **HTTPS vs HTTP**: Backend runs HTTPS (8090), Developer Console embedded in backend (HTTPS), sample apps can run HTTP (3000) in dev mode. Use `-k` flag in curl for self-signed certs.

4. **Configuration Layers**: Thunder has multiple configs:
   - `deployment.yaml` (server config)
   - `.env` (dev-time sample config)
   - `runtime.json` (runtime sample config)
   - `nx.json` (frontend workspace config)

5. **Frontend Workspace**: Thunder uses Nx monorepo. Commands use `pnpm` or `nx` CLI. Always run frontend commands from `frontend/` directory or use workspace filters.

6. **Testing Strategy**:
   - Backend: Unit tests (Go), integration tests (requires built product)
   - Frontend: Vitest for unit/integration, React Testing Library
   - Always build backend before integration tests if code changed

7. **Common Issues**:
   - Missing sample app ID → run setup again, note the ID
   - Certificate errors → use `-k` flag or trust the cert
   - PostgreSQL connection → ensure Docker container is running
   - Port conflicts → check if 8090/3000 are already in use
   - Frontend build issues → clear `node_modules`, reinstall with `pnpm install`
   - Nx cache issues → run `nx reset` to clear cache

8. **Code Contributions**: Follow `docs/content/community/contributing/README.md` guidelines.

### Frontend-Specific Tips

**When helping with React/TypeScript:**
- Thunder uses **React 19** (latest) with new features
- Strict TypeScript configuration
- Functional components with hooks
- Prefer composition over inheritance
- Use shared packages for common functionality

**When helping with UI/UX:**
- Follow Material Design principles (MUI)
- Use Oxygen UI components where available
- Maintain accessibility (ARIA labels, keyboard navigation)
- Support internationalization (i18next)
- Test responsive design (mobile, tablet, desktop)

**When helping with flow designer:**
- Understand `@xyflow/react` concepts (nodes, edges, handles)
- Flow data stored in backend, visualized in frontend
- Nodes represent authentication steps
- Support drag-and-drop, zoom, pan interactions

### Helpful Commands

**Backend (Makefile):**
```bash
make all                  # Build everything with tests
make run                  # Start in dev mode
make build_backend        # Build backend only
make build_frontend       # Build frontend only
make test_unit            # Run unit tests
make test_integration     # Run integration tests
make build_with_coverage  # Build with coverage + run tests
make clean                # Clean build artifacts
```

**Frontend (pnpm/Nx):**
```bash
pnpm install              # Install all dependencies
pnpm build                # Build all apps
pnpm test                 # Run all tests
pnpm lint                 # Lint all code
pnpm lint:fix             # Auto-fix linting issues

# Nx commands (run from frontend/)
nx serve thunder-develop           # Start Developer Console
nx build thunder-develop           # Build Developer Console
nx test thunder-develop            # Test Developer Console
nx lint thunder-develop            # Lint Developer Console
nx run-many --target=build --all  # Build all apps
nx graph                           # Visualize dependency graph
nx reset                           # Clear Nx cache
```

**Sample Apps:**
```bash
npm run dev               # Start dev server
npm run build             # Build for production
npm run preview           # Preview production build
npm run lint              # Lint code
```

## 📚 Additional Resources

- **Main README**: Comprehensive quickstart and feature list
- **Contributing Guide**: `docs/content/community/contributing/README.md`
- **Frontend README**: `frontend/README.md`
- **Issues**: https://github.com/asgardeo/thunder/issues
- **Releases**: https://github.com/asgardeo/thunder/releases
- **License**: Apache 2.0
- **Nx Documentation**: https://nx.dev
- **Asgardeo React SDK**: https://github.com/asgardeo/asgardeo-auth-react-sdk

## 🎯 Project Goals & Philosophy

Thunder aims to be:
- **Developer-Friendly**: Easy to set up, clear APIs, good documentation
- **Flexible**: Customizable authentication flows via visual flow designer
- **Cloud-Native**: Container-ready, microservices-compatible
- **Standards-Compliant**: Full OAuth 2.0/OIDC support
- **Extensible**: Plugin architecture for identity providers, notification senders
- **Modern**: Latest React, TypeScript, Go best practices
- **Monorepo**: Efficient code sharing, unified tooling with Nx

When suggesting changes or helping users, prioritize simplicity, security, standards compliance, and modern development practices.

## 🔧 Advanced Topics

### Frontend Architecture Decisions

**Why Nx?**
- Efficient builds with computation caching
- Consistent tooling across apps and packages
- Dependency graph visualization
- Parallel task execution
- Integrated testing and linting

**Why pnpm?**
- Faster installs (disk space efficient)
- Strict dependency resolution
- Better monorepo support than npm/yarn
- Workspace protocol for internal packages

**Why Vite?**
- Fast HMR (Hot Module Replacement)
- Modern build tool (ESBuild/Rollup)
- Native ES modules support
- Better developer experience than Webpack

**Why Material-UI + Oxygen UI?**
- Material Design compliance
- Comprehensive component library
- Accessibility built-in
- WSO2 branding integration (Oxygen UI)
- Active maintenance and community

### Integration Patterns

**Backend ↔ Developer Console:**
- Developer Console served by Go backend
- API calls authenticated with OAuth tokens
- WebSocket for real-time updates (if applicable)
- Shared configuration (deployment.yaml)

**Sample App ↔ Thunder:**
- Two integration modes:
  1. **Native Flow API**: Direct REST API calls, no OAuth redirect
  2. **OAuth/OIDC**: Standard authorization code flow with redirects
- SDK abstracts OAuth complexity (`@asgardeo/react`)

**Flow Designer ↔ Backend:**
- Flow definitions stored in backend database
- Frontend visualizes and edits flows
- Backend executes flows at runtime
- Visual editor generates JSON flow configuration

---

**Version**: Based on Thunder v0.16.0+ codebase  
**Last Updated**: 2026-02-03  
**Maintained by**: WSO2 / Asgardeo Team
