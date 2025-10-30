# Contributing to WSO2 Thunder ⚡

Thank you for your interest in contributing to WSO2 Thunder! This guide will help you set up your development environment and understand the contribution process.

## 📋 Table of Contents

- [Prerequisites](#prerequisites)
  - [Required Tools (Core Development)](#required-tools-core-development)
  - [Required Tools (Frontend Development)](#required-tools-frontend-development)
- [Development Setup](#development-setup)
- [Development Setup (Frontend)](#development-setup-frontend)
  - [Installing Dependencies](#installing-dependencies)
  - [Building the Project](#building-the-project)
  - [Setting up the Thunder Gate Application](#setting-up-the-thunder-gate-application)
  - [Setting up the Thunder Develop Application](#setting-up-the-thunder-develop-application)

## Prerequisites

Before you begin, ensure you have the following installed on your system:

### Required Tools (Core Development)

- **[Git](https://git-scm.com/downloads)** - Version control system
- **[Go](https://golang.org/doc/install)** - Version 1.25 or higher
- **[Node.js](https://nodejs.org/en/download/)** - Version 22 or higher (`LTS` is recommended ✅)
- **[pnpm](https://pnpm.io/installation)** - Version 9 or higher (`LTS` is recommended ✅)

### Required Tools (Frontend Development)

- **[ESLint VSCode Extension](https://marketplace.visualstudio.com/items?itemName=dbaeumer.vscode-eslint)** - For linting support in VSCode
- **[Prettier VSCode Extension](https://marketplace.visualstudio.com/items?itemName=esbenp.prettier-vscode)** - For code formatting in VSCode

## Development Setup

1. **Fork the Repository**: Start by forking the WSO2 Thunder repository to your GitHub account.

2. **Clone the Repository**: Clone your forked repository to your local machine.

```bash
git clone https://github.com/<your-username>/thunder.git
cd thunder
```

3. **Run the Project**: Start the Thunder server.

```bash
make run
```

This will run the backend server on `http://localhost:8090`.
If you want to run the frontend applications as well, follow the instructions in the [Development Setup (Frontend)](#development-setup-frontend) section below.

## Development Setup (Frontend)

### Installing Dependencies

1. Navigate to the Thunder frontend directory.

```bash
cd frontend
```

2. Install the dependencies using `pnpm`.

```bash
pnpm install
```

### Building the Project

Execute the build command to compile the project. This will build all the necessary packages and applications.

```bash
pnpm build
```

### Setting up the Thunder Gate Application

1. Point the `gate_client` in `thunder-home/repository/conf/deployment.yaml` to the local Thunder Gate application.

```yaml
gate_client:
  port: 5190
  scheme: "https"
  login_path: "/signin"
```

2. Run the Thunder Gate application.

```bash
cd thunder-home/frontend
pnpm --filter @thunder/gate dev
```

### Setting up the Thunder Develop Application

1. First, retrieve the application ID of the **Develop** application from the Thunder server. This will be the application with the `client_id` **DEVELOP**.

```bash
curl -k -X GET "https://localhost:8090/applications"
```

2. Then, get the current **Develop** application configuration:

```bash
curl -k -X GET "https://localhost:8090/applications/<develop-application-id>"
```

> [!Note]
> - Replace `<develop-application-id>` with the actual application ID (e.g., `6100bc91-ba99-4ce9-87dd-6d4d80178c38`) obtained from the previous step.
> - The `-k` flag allows curl to work with self-signed SSL certificates in development.

The response will look similar to the following:

```json
{
    "id": "6100bc91-ba99-4ce9-87dd-6d4d80178c38",
    "name": "Develop",
    "description": "Developer application for Thunder",
    "client_id": "DEVELOP",
    "auth_flow_graph_id": "auth_flow_config_basic",
    "registration_flow_graph_id": "registration_flow_config_basic",
    "is_registration_flow_enabled": true,
    "url": "https://localhost:8090/develop",
    "logo_url": "https://localhost:8090/develop/assets/images/asgardeo-trifacta.svg",
    "token": {
        "issuer": "thunder",
        "validity_period": 3600,
        "user_attributes": null
    },
    "certificate": { "type": "NONE", "value": "" },
    "inbound_auth_config": [
        {
            "type": "oauth2",
            "config": {
                "client_id": "DEVELOP",
                "redirect_uris": [
                    "https://localhost:8090/develop"
                ],
                "grant_types": ["authorization_code"],
                "response_types": ["code"],
                "token_endpoint_auth_method": "none",
                "pkce_required": false,
                "public_client": true,
                "token": {
                    "issuer": "https://localhost:8090/oauth2/token",
                    "access_token": {
                        "issuer": "",
                        "validity_period": 3600,
                        "user_attributes": [
                            "given_name",
                            "family_name",
                            "email",
                            "groups",
                            "name"
                        ]
                    },
                    "id_token": {
                        "validity_period": 3600,
                        "user_attributes": [
                            "given_name",
                            "family_name",
                            "email",
                            "groups",
                            "name"
                        ],
                        "scope_claims": {
                            "email": ["email", "email_verified"],
                            "group": ["groups"],
                            "phone": ["phone_number", "phone_number_verified"],
                            "profile": [
                                "name",
                                "given_name",
                                "family_name",
                                "picture"
                            ]
                        }
                    }
                }
            }
        }
    ]
}
```

2. Copy the response from step 1 and update the `redirect_uris` in the JSON object to include the local development URL (ex: https://localhost:5191/develop). Locate the `inbound_auth_config > config` section and modify the `redirect_uris` array:

```json
"redirect_uris": [
  "https://localhost:8090/develop",
  "https://localhost:5191/develop"
]
```

3. Update the **Develop** application with the modified configuration by passing the updated JSON directly:

```bash
curl -k -X PUT "https://localhost:8090/applications/<develop-application-id>" \
  -H "Content-Type: application/json" \
  -d '<paste-the-modified-json-here>'
```

4. Add the local development origin of the thunder-develop application (https://localhost:5191) to the CORS allowed origins in `thunder-home/repository/conf/deployment.yaml`.

```yaml
cors:
  allowed_origins:
    - "https://localhost:3000"
    - "https://localhost:5191"
```

5. Run the Thunder Develop application.

```bash
pnpm --filter @thunder/develop dev
```

This will run the Thunder Develop application on `https://localhost:5191/develop`.
