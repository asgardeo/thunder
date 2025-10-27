# Thunder Frontend ⚡

Frontend workspace for **WSO2 Thunder** - a modern identity management suite. This workspace is built with [Nx](https://nx.dev) and contains React applications and shared packages for the Thunder platform.

## Prerequisites

- [Git](https://git-scm.com/downloads) - Open source distributed version control system. For install instructions, refer [this](https://www.atlassian.com/git/tutorials/install-git).
- [Node.js](https://nodejs.org/en/download/) - JavaScript runtime. `LTS` is recommended ✅ (`v22 or higher`)
- [pnpm](https://pnpm.io/) - Fast, disk space efficient package manager. `LTS` is recommended ✅ (`v9 or higher`)
- Thunder server running (see [main README](../README.md))

### Installation

To set up the frontend workspace, run the following command:

```sh
cd frontend
pnpm install
```

### Applications

#### Thunder Gate

Authentication gateway application providing login, registration, and recovery UIs.

##### Run Thunder Gate

```bash
# From anywhere inside the frontend directory
pnpm --filter @thunder/gate dev
```

or

```bash
cd apps/thunder-gate
pnpm dev
```

> [!Note]
> The application will run on port `5190` by default.
> [https://localhost:5190/signin](https://localhost:5190/signin)

#### Thunder Develop

Development application for managing Thunder platform configurations.

##### Run Thunder Develop

```bash
# From anywhere inside the frontend directory
pnpm --filter @thunder/develop dev
```

or

```bash
cd apps/thunder-develop
pnpm dev
```

> [!Note]
> The application will run on port `5191` by default.
> [https://localhost:5191/signin](https://localhost:5191/signin)

### Packages

- **`thunder-ui`** - Shared UI components library
- **`thunder-logger`** - Logging utilities
- **`thunder-eslint-plugin`** - Custom ESLint plugin for Thunder projects  
- **`thunder-prettier-config`** - Shared Prettier configuration

### Development

Start all applications in development mode:

```sh
# Start all apps
pnpm dev

# Or start specific apps
npx nx dev thunder-gate
npx nx dev thunder-develop
```

### Building

Build all applications:

```sh
# Build everything
pnpm build

# Build specific app
npx nx build thunder-gate
```

## 🛠️ Available Scripts

### Workspace Level

```sh
pnpm build          # Build all applications and packages
pnpm clean          # Clean all build artifacts
pnpm lint           # Lint all projects
pnpm lint:fix       # Fix linting issues
pnpm test           # Run all tests
pnpm typecheck      # Run TypeScript checks
```

## 📊 Project Graph

Visualize the project dependencies:

```sh
npx nx graph
```

This will open an interactive visualization showing how your applications and packages depend on each other.

## 🧪 Testing

```sh
# Run all tests
pnpm test

# Run tests for specific project
pnpm --filter <project-name> test
```

### Code Quality

- **ESLint**: Configured with Thunder-specific rules
- **Prettier**: Consistent code formatting
- **TypeScript**: Strict type checking

### Nx Console

Install the Nx Console extension for VSCode or IntelliJ to enhance your development experience with:

- Task running from the UI
- Project graph visualization
- Code generation assistance

[Install Nx Console](https://nx.dev/getting-started/editor-setup)

## 🚀 Deployment

### Production Build

```sh
# Build for production
pnpm build

# Preview production build
npx nx preview thunder-gate
```

## 🤝 Contributing

1. Follow the coding standards defined in the ESLint and Prettier configurations
2. Ensure all tests pass before submitting changes
3. Update documentation when adding new features
4. Use conventional commit messages

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](../LICENSE) file for details.
