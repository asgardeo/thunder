---
title: Development Guide
description: Learn how to contribute code to Thunder, including setup, development workflow, testing, and pull request processes.
hide_table_of_contents: false
---

# Development Guide - Contributing Code

This guide explains how to contribute code to Thunder.

---

## 🎯 When to Use This Guide

Use this guide when:

* ✅ You're implementing an **approved design proposal**
* ✅ You're working on a **Fast Track contribution** (bug fix, improvement)
* ✅ You need to understand **code standards and workflow**

---

## 📋 Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Development Commands](#development-commands)
  - [Building](#building)
  - [Running](#running)
  - [Testing](#testing)
  - [Linting](#linting)
  - [Generating Mocks](#generating-mocks)
  - [Other Commands](#other-commands)
- [Advanced Setup (Manual Mode)](#advanced-setup-manual-mode)
- [Development Workflow](#development-workflow)

---

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

---

## Quick Start

### 1. Fork the Repository

Start by forking the WSO2 Thunder repository to your GitHub account.

### 2. Clone the Repository

Clone your forked repository to your local machine.

```bash
git clone https://github.com/<your-username>/thunder.git
cd thunder
```

### 3. Configure the Deployment

Open `backend/cmd/server/repository/conf/deployment.yaml` in your favorite text editor and make the following changes:

**Add CORS allowed origins:**

```yaml
cors:
  allowed_origins:
    - "https://localhost:5190"
    - "https://localhost:5191"
```

> [!TIP]
> This configuration allows the Thunder Gate and Thunder Develop applications to communicate with the backend server during development.

**Configure the Gate client:**

```yaml
gate_client:
  port: 5190
```

> [!TIP]
> This configuration points the backend server to the local Thunder Gate application for authentication.

### 4. Run the Project

Start the Thunder server & apps in development mode.

```bash
make run
```

> [!Note]
> This command will automatically set up your complete development environment:
>
> **What gets created:**
>
> - `Develop` application with redirect URI pointing to the local Thunder Develop application
> - Default admin user with credentials: `admin` / `admin`
>
> **Services that start:**
>
> - **Backend server**: [https://localhost:8090](https://localhost:8090)
> - **Thunder Gate** (Login/Register): [https://localhost:5190/gate](https://localhost:5190/gate)
> - **Thunder Develop** (Admin Console): [https://localhost:5191/develop](https://localhost:5191/develop)

💡 **Alternatively**, if you would rather run the backend and frontend servers separately, run the following commands in separate terminal windows:

**Backend Server:**

```bash
make run_backend
```

**Frontend Server:**

```bash
make run_frontend
```

---

## Development Commands

This section provides a reference for all available make commands.

### Building

**Build everything (backend + frontend + samples):**
```bash
make build
```

**Build only the backend:**
```bash
make build_backend
```

**Build only the frontend:**
```bash
make build_frontend
```

**Build with test coverage instrumentation:**
```bash
make build_with_coverage
```

This will run unit tests, build with coverage flags, run integration tests, and generate a combined coverage report.

### Running

**Run everything (backend + frontend):**
```bash
make run
```

This automatically sets up the complete development environment with backend, frontend apps, and seed data.

**Run only the backend:**
```bash
make run_backend
```

**Run only the frontend:**
```bash
make run_frontend
```

### Testing

**Run all tests (unit + integration):**
```bash
make test
```

**Run unit tests only:**
```bash
make test_unit
```

**Run integration tests only:**
```bash
make test_integration
```

**Code Coverage Requirements:**
- **Minimum**: at least 80% code coverage for new code
- **Encouraged**: 100% coverage

### Linting

**Lint backend code:**
```bash
make lint_backend
```

This uses `golangci-lint` to check code quality and style.

**Lint frontend code:**
```bash
make lint_frontend
```

**Lint both backend and frontend:**
```bash
make lint
```

### Generating Mocks

Thunder uses `mockery` to generate mocks for unit tests.

**Generate mocks:**
```bash
make mockery
```

This will generate mocks based on the configurations in:
- `.mockery.public.yml` - For public interfaces
- `.mockery.private.yml` - For private interfaces

Generated mocks are placed in `backend/tests/mocks/`.

### Other Commands

**Clean build artifacts:**
```bash
make clean
```

**Build Docker images:**
```bash
# Single-arch image with version tag
make docker-build

# Multi-arch image (amd64 + arm64)
make docker-build-multiarch
```

**View all available commands:**
```bash
make help
```

---

## Advanced Setup (Manual Mode)

<details>
<summary><strong>📖 Click to expand - Manual Frontend Setup</strong></summary>

For developers who want to run frontend components separately without using `make run` commands.

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

### Seed Data (Optional)

💡 **Note**: This step is only necessary if you are running the backend server manually and have not yet set up the initial data.

If you have not already created the `Develop` application and the default admin user, you can do so by running the following command:

```bash
THUNDER_API_BASE="https://localhost:8090" \
  backend/cmd/server/bootstrap/01-default-resources.sh \
  --develop-redirect-uris "https://localhost:5191/develop"
```

### Setting up the Thunder Gate Application

1. Point the `gate_client` in `deployment.yaml` to the local Thunder Gate application.

   - If you are running `make run`, change the `gate_client` section in `backend/cmd/server/repository/conf/deployment.yaml`
   - If you are running the backend server manually, change the `gate_client` section in `<THUNDER_HOME>/repository/conf/deployment.yaml`

```yaml
gate_client:
  port: 5190
```

2. Add the local development origin of the Thunder Gate application (https://localhost:5190) to the CORS allowed origins in `<THUNDER_HOME>/repository/conf/deployment.yaml`.

```yaml
cors:
  allowed_origins:
    - "https://localhost:5190"
```

3. Run the Thunder Gate application.

```bash
cd frontend
pnpm --filter @thunder/gate dev
```

### Setting up the Thunder Develop Application

💡 **IMPORTANT**: This section assumes that you have already created the `Develop` application using the initial data setup script. If not, please refer to the [Seed Data](#seed-data-optional) section above.

1. First, retrieve the application ID of the **Develop** application from the Thunder server. This will be the application with the `client_id` **DEVELOP**.

```bash
curl -k -X GET "https://localhost:8090/applications"
```

2. Then, get the current **Develop** application configuration:

```bash
curl -k -X GET "https://localhost:8090/applications/<develop-application-id>"
```

💡 **Note**: Replace `<develop-application-id>` with the actual application ID (e.g., `6100bc91-ba99-4ce9-87dd-6d4d80178c38`) obtained from the previous step. The `-k` flag allows curl to work with self-signed SSL certificates in development.

3. Copy the response from step 2 and update the `redirect_uris` in the JSON object to include the local development URL (ex: https://localhost:5191/develop). Locate the `inbound_auth_config > config` section and modify the `redirect_uris` array:

```json
"redirect_uris": [
  "https://localhost:8090/develop",
  "https://localhost:5191/develop"
]
```

4. Update the **Develop** application with the modified configuration by passing the updated JSON directly:

```bash
curl -k -X PUT "https://localhost:8090/applications/<develop-application-id>" \
  -H "Content-Type: application/json" \
  -d '<paste-the-modified-json-here>'
```

5. Add the local development origin of the thunder-develop application (https://localhost:5191) to the CORS allowed origins in `<THUNDER_HOME>/repository/conf/deployment.yaml`.

```yaml
cors:
  allowed_origins:
    - "https://localhost:5191"
```

6. Run the Thunder Develop application.

```bash
pnpm --filter @thunder/develop dev
```

This will run the Thunder Develop application on `https://localhost:5191/develop`.

</details>

---

## Development Workflow

### 1. Make Your Changes

Follow coding guidelines and best practices:

- **Backend**: Review `.github/instructions/` for package structure, services, data access, API handlers, and testing patterns
- **Frontend**: See `.github/instructions/frontend.instructions.md`
- **Security**: Avoid SQL injection, XSS, command injection, and OWASP top 10 vulnerabilities

**Best practices:**
- Keep changes focused on a single issue
- Use meaningful variable/function names
- Add comments for complex logic
- Maintain consistency with existing code

### 2. Test and Validate Your Changes

Before committing, ensure code quality and correctness:

**Run tests:**
```bash
make test  # Run all tests (unit + integration)
```

**Run linting:**
```bash
make lint_backend  # Backend only
# or
make lint  # Both backend and frontend
```

**Requirements:**
- ✅ All tests must pass
- ✅ Code coverage: at least 80% (higher encouraged)
- ✅ No linting errors
- ✅ Unit tests for business logic
- ✅ Integration tests for API endpoints
- ✅ Test edge cases and error scenarios

> **Tip**: See [Testing](#testing) and [Linting](#linting) sections for more command options.

### 3. Commit Your Changes

Use clear, descriptive commit messages:

```bash
git commit -m "<commit-message>"
```

**Best practices:**
- Make atomic commits (one logical change per commit)
- Write commit messages that explain **why**, not just **what**
- Reference the issue number using `Fixes #<number>` or `Closes #<number>`

### 4. Push and Create PR

Push your changes to your forked repository:

```bash
git push origin <branch-name>
```

Then create a Pull Request to the `thunder/main` branch on GitHub:

1. Go to your fork on GitHub
2. Click **"Compare & pull request"**
3. Provide a clear, descriptive title (e.g., "Add TOTP-based MFA support")
4. Fill in the PR templates
5. Link to the related issue
6. Ensure all CI checks pass

### 5. Address Review Feedback

- Respond to reviewer comments promptly
- Make requested changes
- Keep commit history clean by rebasing or squashing review fixes into the appropriate commit, rather than adding separate "fix review comments" commits.
- If your PR has multiple commits, consolidate review feedback into a single meaningful commit
- Push updates to the same branch
- Re-request review when ready

**Tips:**
- Ask questions if feedback is unclear
- Explain your approach if you disagree (constructively)
- Mark resolved conversations as resolved
- Keep the discussion focused and professional

### 6. Merge

Once approved by maintainers and all checks pass a maintainer will merge your PR.

---

## 🎉 After Your PR is Merged

Your contribution will be:

* ✅ Included in the next release
* ✅ Mentioned in release notes
* ✅ Celebrated in the community! 🎊

**Thank you for contributing to Thunder!** 🚀⚡

---

**Ready to start coding?** Pick an issue, create a branch, and start building!
