# Sample Client App

This is a sample React application that demonstrates the Thunder login flow. The application allows users to log in using their credentials and retrieves an access token from the server.

## Hosting Options

This sample application includes everything you need to run the Thunder login demo. The built package contains the compiled application (`dist` folder), along with a simple Node.js server implementation (`server.js`) for your convenience. You can choose to run the application using the provided server or host it on your preferred web server.

### Option 1: Using the Provided Node Server

The sample application comes with a built-in Node.js server that serves the React app over HTTPS. Follow the steps below to set it up:

**Prerequisites:**
- Node.js 20+

1. **Install dependencies:**
   ```bash
   npm i
   ```

2. **Start the server:**
   ```bash
   npm start
   ```

3. **Access the application:**
   Navigate to `https://localhost:3000`

### Option 2: Using Your Own Web Server

The `dist` folder contains the built application that can be hosted on any web server. Configure your server to serve these static files and ensure proper HTTPS setup.

**Generate Certificates:**

Generate a self-signed SSL certificate by running the following command:

```bash
openssl req -nodes -new -x509 -keyout server.key -out server.cert
```

**Configure the application:**

There are two ways to configure the application depending on your use case:

**For Development (using `npm run dev`):**

Create or update the `.env` file in the root directory with your configuration. Replace `<your-app-id>` with your actual application ID:

```env
VITE_REACT_APPLICATIONS_ENDPOINT=https://localhost:8090/applications
VITE_REACT_APP_SERVER_FLOW_ENDPOINT=https://localhost:8090/flow
VITE_REACT_APP_AUTH_APP_ID=<your-app-id>
```

**For Production (using built app with `npm start` or custom web server):**

Update the `public/runtime.json` file with your configuration:

```json
{
    "applicationID": "<your-app-id>",
    "flowEndpoint": "https://localhost:8090/flow",
    "applicationsEndpoint": "https://localhost:8090/applications"
}
```

**Note:** In development mode, `.env` values take precedence. In production mode, `runtime.json` values take precedence. Placeholder values (e.g., `{your-application-id}`) in `runtime.json` are automatically ignored.

## License

Licenses this source under the Apache License, Version 2.0 LICENSE, You may not use this file except in compliance with the License.

---------------------------------------------------------------------------
(c) Copyright 2025 WSO2 LLC.
