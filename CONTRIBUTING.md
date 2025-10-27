# Contributing to WSO2 Thunder ⚡

Thank you for your interest in contributing to WSO2 Thunder! This guide will help you set up your development environment and understand the contribution process.

## 📋 Table of Contents

- [Prerequisites](#prerequisites)
- [Development Setup](#development-setup)

## 📦 Prerequisites

Before you begin, ensure you have the following installed on your system:

### Required Tools (Core Development)

- **[Git](https://git-scm.com/downloads)** - Version control system
- **[Go](https://golang.org/doc/install)** - Version 1.25 or higher
- **[Node.js](https://nodejs.org/en/download/)** - Version 22 or higher (`LTS` is recommended ✅)
- **[pnpm](https://pnpm.io/installation)** - Version 9 or higher (`LTS` is recommended ✅)

### Required Tools (Frontend Development)

- **[ESLint VSCode Extension](https://marketplace.visualstudio.com/items?itemName=dbaeumer.vscode-eslint)** - For linting support in VSCode
- **[Prettier VSCode Extension](https://marketplace.visualstudio.com/items?itemName=esbenp.prettier-vscode)** - For code formatting in VSCode

## 🛠️ Development Setup

1.**Fork the Repository**: Start by forking the WSO2 Thunder repository to your GitHub account.

2.**Clone the Repository**: Clone your forked repository to your local machine.

```bash
git clone https://github.com/<your-username>/thunder.git
cd thunder
```

3.**Build the Project**: Navigate to the project root directory and run the build command.

```bash
make build
```

4.**Run the Project**: Start the Thunder server.

```bash
make run
```
