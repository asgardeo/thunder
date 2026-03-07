# Thunder React SDK Sample Application

This sample application demonstrates how to integrate Thunder authentication into a React application using the `@asgardeo/react` SDK. It showcases OAuth 2.0/OIDC based user authentication, token management, and user profile display.

## Features

- 🔐 OAuth 2.0/OIDC authentication with Thunder
- 👤 Display user profile information (name, username)
- 🎫 View access tokens and decoded JWT components (header, payload, signature)
- 🎨 Modern UI with Oxygen UI components
- 🔄 Token refresh and session management
- 📱 Responsive design

## Prerequisites

- Node.js 20+
- A running Thunder server instance (default: `https://localhost:8090`)
- An OAuth application registered in Thunder with appropriate redirect URIs

## Quick Start (Pre-built Application)

If you have the pre-built distribution, you can run it directly:

### 1. Configure the Application

Open `dist/runtime.json` and set your Thunder application credentials:

```json
{
  "clientId": "{your-client-id}",
  "baseUrl": "https://localhost:8090"
}
```

| Property | Description |
|----------|-------------|
| `clientId` | The OAuth client ID from your Thunder application |
| `baseUrl` | The base URL of your Thunder server |

### 2. Start the Application

```bash
./start.sh
```

The start script will:
- Serve the application on port 3000
- Use HTTPS if SSL certificates are present in the `dist` folder
- Fall back to HTTP if certificates are not found

### 3. Access the Application

Open your browser and navigate to [https://localhost:3000](https://localhost:3000) (or `http://localhost:3000` if running without SSL)

## Development

To run the application in development mode with hot reloading:

### 1. Install Dependencies

```bash
npm install
```

### 2. Set Up SSL Certificates

For HTTPS support during development, copy the SSL certificates from your Thunder distribution to the project root:

```bash
# From Thunder distribution
cp /path/to/thunder/repository/resources/security/server.key .
cp /path/to/thunder/repository/resources/security/server.cert .

# Or from build output (if building from source)
cp ../../target/out/.cert/server.key .
cp ../../target/out/.cert/server.cert .
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
| `npm run build` | Build for production (outputs to `dist/`) |
| `npm run preview` | Preview the production build locally |
| `npm run lint` | Run ESLint to check code quality |

## Configuration Reference

### Complete `runtime.json` Schema

```json
{
  "clientId": "string (required) - OAuth client ID",
  "baseUrl": "string (required) - Thunder server base URL"
}
```

### Thunder Application Setup

Before running the app, ensure your Thunder application is configured with:

1. **Authorized Redirect URLs**: Add your application URL (e.g., `https://localhost:3000`)
2. **Allowed Origins**: Add your application origin for CORS
3. **Grant Types**: Authorization Code (with PKCE required for SPAs)

## Troubleshooting

### Common Issues

**Issue**: "Failed to fetch token"
- Ensure Thunder server is running and accessible at the configured base URL
- Verify the client ID is correct
- Check that redirect URLs are properly configured in Thunder

**Issue**: "Invalid client" error
- Double-check the `clientId` in your `runtime.json`
- Ensure the application exists in Thunder and is enabled

**Issue**: CORS errors
- Add your application URL to "Allowed Origins" in Thunder's `deployment.yaml`:
  ```yaml
  cors:
    allowed_origins:
      - "https://localhost:3000"
  ```

## How It Works

### Authentication Flow

1. **SDK Provider Setup**: The app wraps components with `AsgardeoProvider` configured with Thunder's base URL and client ID
2. **Conditional Rendering**: Uses `SignedIn`/`SignedOut` components to show appropriate content based on auth state
3. **Token Management**: Retrieves and decodes JWT tokens to display user information

### Key Code Examples

**Provider Configuration:**
```tsx
<AsgardeoProvider
  baseUrl={config.baseUrl}
  clientId={config.clientId}
  platform="AsgardeoV2"
>
  <App />
</AsgardeoProvider>
```

**Using Authentication Hooks:**
```tsx
import { useAsgardeo } from "@asgardeo/react";

const { getAccessToken, signIn } = useAsgardeo();
const accessToken = await getAccessToken();
```

## License

Licensed under the Apache License, Version 2.0. You may not use this file except in compliance with the License.

---------------------------------------------------------------------------
(c) Copyright 2025 WSO2 LLC.

