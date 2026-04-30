# React Vanilla Sample Application

This sample React application demonstrates integrating Thunder authentication into your application. It supports two authentication approaches:

- **Native Flow**: App-native authentication using Thunder's flow orchestration API
- **OAuth 2.0 / OIDC**: Standards-based authentication using OAuth 2.0 Authorization Code flow

## Prerequisites

- Node.js 20+
- A running Thunder server instance (default: `https://localhost:8090`)
- An application registered in Thunder

## Quick Start (Pre-built Application)

If you have the pre-built distribution, you can run it directly:

### 1. Configure the Application

Open `app/runtime.json` and configure the settings based on your preferred authentication approach.

#### Option A: Native Flow Configuration

For app-native authentication using Thunder's flow API:

```json
{
    "flowEndpoint": "https://localhost:8090/flow",
    "applicationsEndpoint": "https://localhost:8090/applications",
    "redirectBasedLogin": false,
    "applicationID": "{your-application-id}"
}
```

| Property | Description |
|----------|-------------|
| `flowEndpoint` | Thunder's flow orchestration endpoint |
| `applicationsEndpoint` | Thunder's applications API endpoint |
| `redirectBasedLogin` | Set to `false` for native flow |
| `applicationID` | The application ID registered in Thunder (obtained during Thunder setup) |

#### Option B: OAuth 2.0 / OIDC Configuration

For standards-based OAuth 2.0 authentication with redirect-based login:

```json
{
    "redirectBasedLogin": true,
    "authorizationEndpoint": "https://localhost:8090/oauth2/authorize",
    "tokenEndpoint": "https://localhost:8090/oauth2/token",
    "clientId": "{your-oauth-client-id}",
    "redirectUri": "https://localhost:3000",
    "scope": "openid"
}
```

| Property | Description |
|----------|-------------|
| `redirectBasedLogin` | Set to `true` for OAuth redirect flow |
| `authorizationEndpoint` | OAuth 2.0 authorization endpoint |
| `tokenEndpoint` | OAuth 2.0 token endpoint |
| `clientId` | OAuth client ID for the application |
| `redirectUri` | Callback URL where Thunder redirects after authentication |
| `scope` | OAuth scopes (e.g., `openid`, `profile`, `email`) |

#### Expected Flow Node IDs

When using Native Flow, the sample app UI renders based on `nextNode` values in the flow definition. Your flow should use these node IDs for proper UI rendering:

| Node ID | Purpose |
|---------|---------|
| `basic_auth` | Username/password authentication |
| `github_auth` | GitHub OAuth |
| `google_auth` | Google OAuth |
| `prompt_mobile` or `mobile_prompt_username` | SMS OTP authentication |

### 2. Start the Application

**Linux/macOS:**
```bash
sh start.sh
```

**Windows:**
```powershell
.\start.ps1
```

### 3. Access the Application

Open your browser and navigate to [https://localhost:3000](https://localhost:3000)

## Development

To run the application in development mode with hot reloading:

### 1. Install Dependencies

```bash
npm install
```

### 2. Set Up SSL Certificates

For HTTPS support, copy the SSL certificates from your Thunder distribution to the project root:

```bash
# From Thunder distribution
cp /path/to/thunder/repository/resources/security/server.key .
cp /path/to/thunder/repository/resources/security/server.cert .

# Or from build output (if building from source)
cp ../../target/out/.cert/server.key .
cp ../../target/out/.cert/server.cert .
```

Or generate self-signed certificates:

```bash
openssl req -nodes -new -x509 -keyout server.key -out server.cert
```

### 3. Start Development Server

```bash
npm run dev
```

The application will be available at [https://localhost:3000](https://localhost:3000)

### Available Scripts

| Command | Description |
|---------|-------------|
| `npm run dev` | Start development server with hot reloading |
| `npm run build` | Build for production (outputs to `dist/` and prepares server) |
| `npm run preview` | Preview the production build locally |
| `npm run lint` | Run ESLint to check code quality |
| `npm start` | Build and preview the production application |

## Hosting Options

This sample includes a pre-built application with a simple Node.js server. You can also host the application on your own web server.

### Using the Provided Node Server

The sample comes with a built-in Node.js server that serves the React app over HTTPS.

1. Install dependencies and build:
   ```bash
   npm install
   npm run build
   ```

2. Start the server:
   ```bash
   cd server
   npm start
   ```

### Using Your Own Web Server

The `app` folder (or `dist` after building) contains the built application that can be hosted on any web server. Configure your server to:

1. Serve the static files from the `app` or `dist` folder
2. Set up HTTPS with valid certificates
3. Ensure `runtime.json` is accessible and editable for configuration

## License

Licensed under the Apache License, Version 2.0. You may not use this file except in compliance with the License.

---------------------------------------------------------------------------
(c) Copyright 2025 WSO2 LLC.
