# CLAUDE.md - Thunder Identity Management Suite

> **For AI Assistants**: This document helps you understand and work with the Thunder project effectively.

## 🎯 Project Overview

**Thunder** is a modern, open-source Identity and Access Management (IAM) solution built by WSO2. It's designed for teams building secure, customizable authentication experiences across applications, services, and AI agents.

**Key Characteristics:**
- **Language**: Go (backend) + React (frontend/samples)
- **Purpose**: OAuth 2.0/OIDC identity provider with flexible flow designer
- **Architecture**: Microservices-ready, containerized deployment
- **License**: Apache 2.0
- **Repository**: https://github.com/asgardeo/thunder

## 🏗️ Architecture & Tech Stack

### Backend
- **Language**: Go 1.24+ (1.25+ for builds)
- **Standards**: OAuth 2.0, OpenID Connect (OIDC)
- **Supported Flows**: Client Credentials, Authorization Code, Refresh Token
- **Database**: SQLite (default) / PostgreSQL (configurable)
- **Server**: Runs on `https://localhost:8090`

### Frontend
- **Framework**: React 18+ with Vite
- **Node.js**: 24+
- **Sample Apps**: React Vanilla (direct API) and React SDK (@asgardeo/react)
- **Dev Server**: Runs on `http://localhost:3000`

### Authentication Features
- Username/Password (Basic Auth)
- Social Login: Google, GitHub
- SMS OTP
- Self-registration with similar methods

## 📁 Directory Structure

```
thunder/
├── backend/                      # Go backend source
│   ├── cmd/server/               # Main server entry point
│   │   └── repository/conf/      # Configuration files (deployment.yaml)
│   ├── pkg/                      # Shared packages
│   └── internal/                 # Internal application code
├── samples/apps/                 # Sample React applications
│   ├── react-vanilla-sample/     # Direct API integration sample
│   │   ├── .env.example          # Configuration template
│   │   └── app/runtime.json      # Runtime config (for builds)
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

### Development Workflow
```bash
# Clone and start backend
git clone https://github.com/asgardeo/thunder
cd thunder
make run  # Starts on https://localhost:8090

# In another terminal: start sample app
cd samples/apps/react-vanilla-sample
cp .env.example .env
# Edit .env with VITE_REACT_APP_AUTH_APP_ID
npm install
npm run dev  # Starts on http://localhost:3000
```

### Build from Source
```bash
make all  # Builds with tests
```

### Testing
```bash
make test_unit                       # Unit tests only
make test_integration                # Integration tests (requires built product)
make build_backend test_integration  # Build + integration tests
make build_with_coverage             # Build with coverage, run tests, generate reports
```

### Debugging
```bash
# Start in debug mode (requires Delve)
./start.sh --debug  # Listens on localhost:2345

# VS Code: Use .vscode/launch.json and press F5
# GoLand: Run → Edit Configurations → Go Remote (127.0.0.1:2345)
```

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

## 🎨 Sample App Configuration

### Development (.env file)
```env
# Application ID from setup.sh output
VITE_REACT_APP_AUTH_APP_ID={your-application-id}

# Thunder endpoints
VITE_REACT_APP_SERVER_FLOW_ENDPOINT=https://localhost:8090/flow
VITE_REACT_APPLICATIONS_ENDPOINT=https://localhost:8090/applications

# Login mode: false = native flow API, true = OAuth redirect
VITE_REACT_APP_REDIRECT_BASED_LOGIN=false
```

### Production (runtime.json)
```json
{
  "applicationID": "{your-application-id}"
}
```

## 🛠️ Tips for AI Assistants

### When helping with Thunder development:

1. **Authentication Context**: Users often need help with OAuth flows. Thunder supports both native flow APIs (simpler, app-native) and standard OAuth redirect flows.

2. **Sample App ID**: Always generated during `setup.sh` — users must note this value for sample app configuration.

3. **HTTPS vs HTTP**: Backend runs HTTPS (8090), sample apps run HTTP (3000) in dev mode. Use `-k` flag in curl for self-signed certs.

4. **Configuration Layers**: Thunder has multiple configs:
   - `deployment.yaml` (server config)
   - `.env` (dev-time sample config)
   - `runtime.json` (runtime sample config)

5. **Testing Strategy**:
   - Unit tests: Fast, no dependencies
   - Integration tests: Requires built product
   - Always build backend before integration tests if code changed

6. **Common Issues**:
   - Missing sample app ID → run setup again, note the ID
   - Certificate errors → use `-k` flag or trust the cert
   - PostgreSQL connection → ensure Docker container is running
   - Port conflicts → check if 8090/3000 are already in use

7. **Code Contributions**: Follow `docs/content/community/contributing/README.md` guidelines.

### Helpful Makefile Targets
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

## 📚 Additional Resources

- **Main README**: Comprehensive quickstart and feature list
- **Contributing Guide**: `docs/content/community/contributing/README.md`
- **Issues**: https://github.com/asgardeo/thunder/issues
- **Releases**: https://github.com/asgardeo/thunder/releases
- **License**: Apache 2.0

## 🎯 Project Goals & Philosophy

Thunder aims to be:
- **Developer-Friendly**: Easy to set up, clear APIs, good documentation
- **Flexible**: Customizable authentication flows via flow designer
- **Cloud-Native**: Container-ready, microservices-compatible
- **Standards-Compliant**: Full OAuth 2.0/OIDC support
- **Extensible**: Plugin architecture for identity providers, notification senders

When suggesting changes or helping users, prioritize simplicity, security, and standards compliance.

---

**Version**: Based on Thunder v0.16.0+ codebase  
**Last Updated**: 2026-02-03  
**Maintained by**: WSO2 / Asgardeo Team
