---
title: Scaffolding Tool
sidebar_position: 1
description: Generate feature packages and shared packages with Thunder Create CLI tool
---

# Scaffolding Tool

Thunder Create (`@thunder/create`) is a CLI scaffolding tool that automates the creation of feature packages and shared packages in the Thunder frontend workspace. The tool generates complete, production-ready package structures with pre-configured build tools, testing infrastructure, and best practices.

## Overview

Thunder Create provides:

- Full TypeScript support with type definitions
- Pre-configured test setup with Vitest and React Testing Library
- React components with Oxygen UI and Material-UI integration
- Build tools: ESLint, Prettier, and Rolldown
- Workspace dependencies and npm scripts
- Optional automatic dependency installation and build

## Prerequisites

Before using Thunder Create:

- Run the tool from within the Thunder monorepo
- Ensure pnpm is installed for package management
- Execute from the frontend directory or any subdirectory

:::info Time Estimate
Package creation and build typically completes in approximately 5 minutes.
:::

## Usage

### Linking the Package

Since we haven't published `@thunder/create` to npm yet, link it locally:

```bash
cd frontend/packages/thunder-create
pnpm link --global
```

### Running the Package

Execute from anywhere in the Thunder workspace:

```bash
thunder-create
```

### Creating Feature Packages

Feature packages are domain-specific modules containing business logic, UI components, and API integrations with a structured CRUD pattern.

To create a feature package, follow the interactive prompts:

```bash
thunder-create feature
```

#### Step 1: Select Feature Type

Choose the feature type:

```bash
◆  Feature type:
│  ● Admin feature (thunder-admin-xxx)
│  ○ Gate feature (gate-xxx)
└
```

- **Admin feature**: Administrative console features
- **Gate feature**: Authentication and registration features

#### Step 2: Provide Feature Name

Enter a descriptive name using kebab-case:

```bash
◆  Feature name:
│  user-management
└
```

:::info Naming Requirements

- Use lowercase with hyphens (e.g., `user-management`, `oauth2-integration`)
- Use singular nouns where applicable (e.g., `application`, `role`)
- Start with a letter
- Contain only letters, numbers, underscores, and hyphens
- Maximum 50 characters
:::

#### Step 3: Install Dependencies

Choose whether to install dependencies and build:

```bash
◆  Would you like to install dependencies and
build the feature now?
│  ● Yes, install and build
│  ○ No, I will do it later
└
```

- **Yes**: Executes `pnpm install && pnpm build` automatically
- **No**: Displays manual setup instructions

#### Generated Package Structure

The tool generates a feature package (e.g., `thunder-admin-user-management`) with the following structure:

```bash
thunder-admin-user-management/
├── package.json                 # Package manifest (@thunder/admin-user-management)
├── tsconfig.json               # Base TypeScript configuration
├── tsconfig.lib.json           # Library build configuration
├── tsconfig.spec.json          # Test configuration
├── eslint.config.js            # Linting rules
├── vitest.config.ts            # Test runner configuration
├── rolldown.config.js          # Bundler configuration
├── prettier.config.js          # Code formatting rules
├── .editorconfig               # Editor configuration
├── .gitignore                  # Git ignore patterns
├── .prettierignore             # Prettier ignore patterns
└── src/
    ├── index.ts                # Main entry point
    ├── api/                    # API integration layer
    │   ├── useGetUserManagement.ts
    │   ├── useGetUserManagements.ts
    │   └── __tests__/
    ├── components/             # React components
    │   ├── UserManagementList.tsx
    │   └── __tests__/
    ├── config/                 # Feature configuration
    │   ├── UserManagementConfig.ts
    │   └── __tests__/
    ├── constants/              # Constants and query keys
    │   ├── user-management-query-keys.ts
    │   └── __tests__/
    ├── contexts/               # React context providers
    │   └── UserManagementContext/
    │       ├── UserManagementContext.tsx
    │       ├── UserManagementProvider.tsx
    │       ├── useUserManagement.tsx
    │       └── __tests__/
    ├── data/                   # Static data and fixtures
    │   └── .gitkeep
    ├── hooks/                  # Custom React hooks
    │   ├── useUserManagementSearch.ts
    │   └── __tests__/
    ├── models/                 # TypeScript interfaces and types
    │   ├── user-management.ts
    │   └── __tests__/
    ├── pages/                  # Page-level components
    │   ├── UserManagementListPage.tsx
    │   └── __tests__/
    └── utils/                  # Utility functions
        ├── userManagementUtils.ts
        └── __tests__/
```
