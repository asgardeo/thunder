# @thunder/eslint-plugin

ESLint plugin for ⚡️ Thunder projects with custom rules and configurations.

## Installation

```bash
npm install --save-dev @thunder/eslint-plugin eslint
# or
pnpm add -D @thunder/eslint-plugin eslint
```

## Usage

### ESLint Flat Config (ESLint 9+)

```js
// eslint.config.js
import thunder from '@thunder/eslint-plugin';

export default [
  {
    plugins: {
      '@thunder': thunder,
    },
    rules: {
      '@thunder/copyright-header': 'error',
      '@thunder/no-internal-imports': 'error',
      '@thunder/prefer-system-utils': 'warn',
    },
  },
  // Or use a predefined config
  ...thunder.configs.recommended,
];
```

### Predefined Configurations

#### `recommended`

General rules for all Thunder frontend projects:

- `@thunder/copyright-header`: Enforces WSO2 Apache 2.0 copyright headers
- `@thunder/no-internal-imports`: Prevents importing from internal paths

#### `typescript`

Rules for TypeScript projects:

- Extends `@typescript-eslint/recommended`
- `@thunder/copyright-header`: Enforces copyright headers  
- `@thunder/no-internal-imports`: Prevents internal imports

#### `react`

Rules for React projects:

- Extends `plugin:react/recommended` and `plugin:react-hooks/recommended`
- `@thunder/copyright-header`: Enforces copyright headers  
- `@thunder/no-internal-imports`: Prevents internal imports
- React-specific rules and settings

#### `next`

Rules for Next.js projects:

- Extends `next/core-web-vitals`
- `@thunder/copyright-header`: Enforces copyright headers  
- `@thunder/no-internal-imports`: Prevents internal imports

Rules for TypeScript frontend projects:

- `@thunder/copyright-header`: Enforces copyright headers  
- `@thunder/no-internal-imports`: Prevents internal imports

## Rules

### `@thunder/copyright-header`

Enforces the presence of WSO2 Apache 2.0 copyright header in all source files.

**Options:**

- `excludePatterns`: Array of regex patterns for files to exclude
- `template`: Custom copyright header template

**Examples:**

```js
// ❌ Incorrect - missing header
export const foo = 'bar';

// ✅ Correct - has proper header
/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 * ...rest of Apache 2.0 header...
 */
export const foo = 'bar';
```

### `@thunder/no-internal-imports`

Prevents importing from internal directories of other modules to maintain proper module boundaries.

**Options:**

- `allowedPatterns`: Array of regex patterns for allowed internal imports
- `forbiddenPatterns`: Array of regex patterns for forbidden imports

**Examples:**

```js
// ❌ Incorrect - importing from internal path
import { secret } from 'package/internal/secret';
import { build } from 'package/lib/build';

// ✅ Correct - using public API
import { publicApi } from 'package';
import { helper } from './helper'; // relative imports are OK
```

## Development

### Building

```bash
pnpm build
```

### Testing

```bash
pnpm test
```

### Linting

```bash
pnpm lint
```

## License

Apache 2.0 - see LICENSE file for details.
