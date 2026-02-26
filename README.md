# WSO2 Thunder ⚡

### Identity Management Suite

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![GitHub last commit](https://img.shields.io/github/last-commit/asgardeo/thunder.svg)](https://github.com/asgardeo/thunder/commits/main)
[![GitHub issues](https://img.shields.io/github/issues/asgardeo/thunder.svg)](https://github.com/asgardeo/thunder/issues)
[![codecov.io](https://codecov.io/github/asgardeo/thunder/coverage.svg?branch=main)](https://codecov.io/github/asgardeo/thunder?branch=main)

Thunder is a modern, open-source identity management service designed for teams building secure, customizable authentication experiences across applications, services, and AI agents. It enables developers to design and orchestrate login, registration, and recovery flows using a flexible identity flow designer.

Designed for extensibility, scalability, and seamless containerized deployment, Thunder integrates naturally with microservices and DevOps environments—serving as the core identity layer for your cloud platform.

---

## 🚀 Features

- **Standards-Based**
  - OAuth 2/ OpenID Connect (OIDC): Client Credentials, Authorization Code, Refresh Token
- **Login Options:**
  - Basic Authentication (Username/Password)
  - Social Logins: Google, Github
  - SMS OTP
- **Registration Options:**
  - Username/Password
  - Social Registration: Google, Github
  - SMS OTP
- **RESTful APIs:**
  - App Native Login/Registration
  - User Management
  - Application Management
  - Identity Provider Management
  - Message Notification Sender Management

---

## ⚡ Quickstart

This Quickstart guide will help you get started with WSO2 Thunder quickly. It walks you through downloading and running the product, trying out the sample app, and exploring registering a user, logging in, and using the Client Credentials flow.

### Download and Run WSO2 Thunder

You can run WSO2 Thunder either by downloading the release artifact or using the official Docker image.

#### Option 1: Run from Release Artifact

Follow these steps to download the latest release of WSO2 Thunder and run it locally.

1. **Download the distribution from the latest release**

    Download `thunder-<version>-<os>-<arch>.zip` from the [latest release](https://github.com/asgardeo/thunder/releases/latest) for your operating system and architecture.

    For example, if you are using a MacOS machine with a Apple Silicon (ARM64) processor, you would download `thunder-<version>-macos-arm64.zip`.

2. **Unzip the product**

    Unzip the downloaded file using the following command:

    ```bash
    unzip thunder-<version>-<os>-<arch>.zip
    ```

    Navigate to the unzipped directory:

    ```bash
    cd thunder-<version>-<os>-<arch>/
    ```

3. **Setup the product**

    You need to setup the server with the initial configurations and data before starting the server for the first time.

    If you are using a Linux or macOS machine:

    ```bash
    ./setup.sh
    ```

    If you are using a Windows machine:

    ```powershell
    .\setup.ps1
    ```

    **Note the id of the sample app indicated with the log line `[INFO] Sample App ID: <id>`.** You'll need it for the sample app configuration.

4. **Start the product**

    If you are using a Linux or macOS machine:

    ```bash
    ./start.sh
    ```

    If you are using a Windows machine:

    ```powershell
    .\start.ps1
    ```

    The product will start on `https://localhost:8090`.

#### Option 2: Run with Docker Compose

Follow these steps to run WSO2 Thunder using Docker Compose.

1. **Download the Docker Compose file**

    Download the `docker-compose.yml` file using the following command:

    ```bash
    curl -o docker-compose.yml https://raw.githubusercontent.com/asgardeo/thunder/v0.23.0/install/quick-start/docker-compose.yml
    ```

2. **Start Thunder**

    Run the following command in the directory where you downloaded the `docker-compose.yml` file:

    ```bash
    docker compose up
    ```

    This will automatically:
    - Initialize the database
    - Run the setup process
    - Start the Thunder server

    **Note the id of the sample app indicated with the log line `[INFO] Sample App ID: <id>` in the setup logs.** You'll need it for the sample app configuration.

    The product will start on `https://localhost:8090`.

### Try Out the Product

#### Try out the Developer Console

Follow these steps to access the Developer Console:

1. Open your browser and navigate to [https://localhost:8090/develop](https://localhost:8090/develop).

2. Log in using the admin credentials created during the initial data setup (`admin` / `admin`).

#### Try Out with the Sample App

Thunder provides two sample applications to help you get started quickly:

- **React Vanilla Sample** — Sample React application demonstrating direct API integration without external SDKs. Supports Native Flow API or Standard OAuth/OIDC.
- **React SDK Sample** — Sample React application demonstrating SDK-based integration using `@asgardeo/react` for OAuth 2.0/OIDC authentication.

##### React Vanilla Sample

1. **Download the sample**

    Download `sample-app-react-vanilla-<version>-<os>-<arch>.zip` from the [latest release](https://github.com/asgardeo/thunder/releases/latest).

2. **Unzip and navigate to the sample app directory**

    ```bash
    unzip sample-app-react-vanilla-<version>-<os>-<arch>.zip
    cd sample-app-react-vanilla-<version>-<os>-<arch>/
    ```

3. **Configure the sample**

    Open `app/runtime.json` and set the `applicationID` to the sample app ID generated during "Setup the product":

    ```json
    {
        "applicationID": "{your-application-id}"
    }
    ```

4. **Start the sample**

    ```bash
    ./start.sh
    ```

    Open your browser and navigate to [https://localhost:3000](https://localhost:3000) to access the sample app.

    > 📖 Refer to the `README.md` inside the extracted sample app for detailed configuration options including OAuth redirect-based login.

##### React SDK Sample

1. **Download the sample**

    Download `sample-app-react-sdk-<version>-<os>-<arch>.zip` from the [latest release](https://github.com/asgardeo/thunder/releases/latest).

2. **Unzip and navigate to the sample app directory**

    ```bash
    unzip sample-app-react-sdk-<version>-<os>-<arch>.zip
    cd sample-app-react-sdk-<version>-<os>-<arch>/
    ```

3. **Start the sample**

    ```bash
    ./start.sh
    ```

    Open your browser and navigate to [https://localhost:3000](https://localhost:3000) to access the sample app.

    > 📖 Refer to the `README.md` inside the extracted sample app for detailed configuration and troubleshooting.

##### Self Register and Login (React Vanilla Sample)

The React Vanilla sample supports user self-registration and login:

1. Open [https://localhost:3000](https://localhost:3000) and click **"Sign up"** to register a new user.

    <p align="left">
        <img src="resources/images/sample-app-self-registration-basic.png" alt="Self Registration Username Password" width="400">
    </p>

2. After registration, use the same credentials to **"Sign In"**.

    <p align="left">
        <img src="resources/images/sample-app-login.png" alt="Login to Sample App" width="400">
    </p>

3. Upon successful login, you'll see the home page with your access token.


#### Obtain System API Token

To access the system APIs of Thunder, you need a token with system permissions. Follow the steps below to obtain a system API token.

1. Run the following command, replacing `<application_id>` with the sample app ID generated during "Setup the product."

```bash
curl -k -X POST 'https://localhost:8090/flow/execute' \
  -d '{"applicationId":"<application_id>","flowType":"AUTHENTICATION"}'
```
2. Extract the `flowId` value from the response.
```json
{"flowId":"<flow_id>","flowStatus":"INCOMPLETE", ...}
```

3. Run the following command, replacing `<flow_id>` with the `flowId` value you extracted above.
```bash
curl -k -X POST 'https://localhost:8090/flow/execute' \
  -d '{"flowId":"<flow_id>", "inputs":{"username":"admin","password":"admin", "requested_permissions":"system"},"action": "action_001"}'
```

4. Obtain the system API token by extracting the `assertion` value from the response.
```json
{"flowId":"<flow_id>","flowStatus":"COMPLETE","data":{},"assertion":"<assertion>"}
```

#### Try Out Client Credentials Flow

The Client Credentials flow is used to obtain an access token for machine-to-machine communication. This flow does not require user interaction and is typically used for server-to-server communication.

To try out the Client Credentials flow, follow these steps:

1. Create a Client Application

   Application creation is secured functionality, so you first need to obtain a system API token as mentioned in the "Obtain System API Token" section above.

   Run the following command, replacing `<assertion>` with the assertion value obtained from the previous step.

    ```bash
    curl -kL -X POST -H 'Content-Type: application/json' -H 'Accept: application/json' https://localhost:8090/applications \
    -H 'Authorization: Bearer <assertion>' \
    -d '{
        "name": "Test Sample App",
        "description": "Initial testing App",
        "inbound_auth_config": [
            {
                "type": "oauth2",
                "config": {
                    "client_id": "<client_id>",
                    "client_secret": "<client_secret>",
                    "redirect_uris": [
                        "https://localhost:3000"
                    ],
                    "grant_types": [
                        "client_credentials"
                    ],
                    "token_endpoint_auth_method": "client_secret_basic",
                    "pkce_required": false,
                    "public_client": false,
                    "scopes": ["api:read", "api:write"]
                }
            }
        ]
    }'
    ```

2. Obtain an Access Token

   Use the following cURL command to obtain an access token using the Client Credentials flow. Make sure to replace the `<client_id>` and `<client_secret>` with the values you used when creating the client application.

    ```bash
    curl -k -X POST https://localhost:8090/oauth2/token \
      -d 'grant_type=client_credentials' \
      -u '<client_id>:<client_secret>'
    ```

---

<details>
<summary><h2>🛠️ Build the Product from Source</h2></summary>

### Prerequisites

- Go 1.26+
- Node.js 24+

---

- Build the product with tests using the following command:

    ```bash
    make all
    ```

</details>

---

<details>
<summary><h2>⚙️ Development Setup</h2></summary>

### Prerequisites

- Go 1.26+
- Node.js 24+

### Start Thunder in Development Mode

- Clone the repository:

    ```bash
    git clone https://github.com/asgardeo/thunder
    cd thunder
    ```

- Run the following command to start the product in development mode:

    ```bash
    make run
    ```

- The product will start on `https://localhost:8090`.

### Start the Sample App in Development Mode

The sample apps support two configuration approaches:
- **`.env` file**: Used during development (values are bundled at build time)
- **`runtime.json` file**: Used for deployed/distributed apps (values loaded at runtime)

For development, use the `.env` file approach:

- Navigate to the sample app directory:

  ```bash
  cd samples/apps/react-vanilla-sample
  ```

- Create a file `.env` in the path `samples/apps/react-vanilla-sample/` by copying `.env.example`:

  ```bash
  cp .env.example .env
  ```

- Edit the `.env` file and configure the required values:

  ```env
  # Application ID registered in Thunder
  VITE_REACT_APP_AUTH_APP_ID={your-application-id}
  
  # Thunder server endpoints
  VITE_REACT_APP_SERVER_FLOW_ENDPOINT=https://localhost:8090/flow
  VITE_REACT_APPLICATIONS_ENDPOINT=https://localhost:8090/applications
  
  # Set to false for native flow, true for OAuth redirect flow
  VITE_REACT_APP_REDIRECT_BASED_LOGIN=false
  ```

  > **Note**: For OAuth redirect flow, additional configurations like `VITE_REACT_APP_CLIENT_ID`, `VITE_REACT_APP_SERVER_AUTHORIZATION_ENDPOINT`, and `VITE_REACT_APP_SERVER_TOKEN_ENDPOINT` are required. See `.env.example` for the complete list.

- Install the dependencies:

  ```bash
  npm install
  ```

- Run the sample app using the following command:

  ```bash
  npm run dev
  ```
  
- Open your browser and navigate to `http://localhost:3000` to see the sample app in action.

### Remote Debugging Setup

Thunder supports remote debugging using Delve debugger, enabling debugging from any IDE that supports the Debug Adapter Protocol (DAP).

#### Install Delve Debugger

Install Delve using Go:

```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

Add Delve to your PATH:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

#### Start Thunder in Debug Mode

From the distribution directory:

**Linux/macOS:**

```bash
./start.sh --debug
```

**Windows (PowerShell):**

```powershell
.\start.ps1 --debug
```

**Windows (Command Prompt):**

```cmd
start.bat --debug
```

The debugger will listen on `localhost:2345` by default.

#### Connect from IDE

**VS Code:**
- Use the provided `.vscode/launch.json` configuration
- Press `F5` or go to `Run and Debug`

**GoLand/IntelliJ:**
- Go to `Run → Edit Configurations → + → Go Remote`
- Set Host: `127.0.0.1`, Port: `2345`

**Other IDEs:**
- Configure DAP client to connect to `127.0.0.1:2345`

### Testing

Thunder includes both unit tests and integration tests:

#### Run Unit Tests

```bash
make test_unit
```

#### Run Integration Tests

```bash
make test_integration
```

**Note:** This command will run integration tests on an already built product. If you need to build the product before running integration tests, use:

```bash
make build_backend test_integration
```

#### Run Tests with Coverage

```bash
make build_with_coverage
```

This will build the server with coverage instrumentation, run tests, and generate coverage reports at `target/` directory.

</details>

---

<details>
<summary><h2>🔧 Advanced Setup & Configuration</h2></summary>

<details>
<summary><h3>Running with PostgreSQL Database</h3></summary>

#### Step 1: Start and Initialize PostgreSQL

1. Navigate to local-development directory

```bash
cd install/local-development
```

2. Start PostgreSQL Database in background

```bash
docker compose up -d 
```

3. View PostgreSQL Database logs

```bash
docker compose logs -f
```

4. Stop PostgreSQL Database

```bash
docker compose down
```

- Stop PostgreSQL Database and delete all data 

```bash
docker compose down -v
```

#### Step 2: Configure Thunder to Use PostgreSQL

1. Open the `backend/cmd/server/repository/conf/deployment.yaml` file.
2. Update the `database` section to point to the PostgreSQL database:

    ```yaml
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
            hostname: "localhost"
            port: 5432
            name: "runtimedb"
            username: "asgthunder"
            password: "asgthunder"
            sslmode: "disable"
        user:
            type: "postgres"
            hostname: "localhost"
            port: 5432
            name: "userdb"
            username: "asgthunder"
            password: "asgthunder"
            sslmode: "disable"
    ```

#### Step 3: Run the Product

   ```bash
   make run
   ```

The product will now use the PostgreSQL database for its operations.

</details>

</details>

---

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=asgardeo/thunder&type=date&legend=top-left)](https://www.star-history.com/#asgardeo/thunder&type=date&legend=top-left)

## 🤝 Contributing

Please refer to the [CONTRIBUTING.md](docs/content/community/contributing/README.md) for guidelines on how to contribute to this project.

## License

Licenses this source under the Apache License, Version 2.0 ([LICENSE](LICENSE)), You may not use this file except in compliance with the License.

---------------------------------------------------------------------------
(c) Copyright 2025 WSO2 LLC.
