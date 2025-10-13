/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

/**
 * @fileoverview ESLint config to be used in TypeScript based projects.
 */

import eslint from "@eslint/js";
import tseslint from "typescript-eslint";
import { FlatCompat } from "@eslint/eslintrc";
import path from "path";
import { fileURLToPath } from "url";
import fs from "fs";
import globals from "globals";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const compat = new FlatCompat({
  baseDirectory: __dirname
});

// Utility to resolve tsconfig
function resolveTsconfig() {
  const cwd = process.cwd();
  const eslintTsconfig = path.resolve(cwd, "tsconfig.eslint.json");
  const appTsconfig = path.resolve(cwd, "tsconfig.app.json");
  const defaultTsconfig = path.resolve(cwd, "tsconfig.json");

  if (fs.existsSync(eslintTsconfig)) {
    console.log("[eslint-config] Using tsconfig.eslint.json");
    return eslintTsconfig;
  }

  if (fs.existsSync(appTsconfig)) {
    console.log("[eslint-config] Using tsconfig.app.json");
    return appTsconfig;
  }

  if (fs.existsSync(defaultTsconfig)) {
    console.log("[eslint-config] Using tsconfig.json");
    return defaultTsconfig;
  }

  console.warn(
    "[eslint-config] No tsconfig found. Type checking will be limited."
  );

  return undefined;
}

const tsconfigPath = resolveTsconfig();

export const reactConfig = [
  // React-specific configs for all files
  ...compat.extends('airbnb'),
  ...compat.extends('airbnb/hooks'),
  
  // TypeScript React rules only for TypeScript files
  ...compat.extends("@kesills/airbnb-typescript").map(config => ({
    ...config,
    files: ["**/*.{ts,tsx}"],
  })),
  
  // React JSX runtime configuration
  {
    files: ["**/*.{js,jsx,ts,tsx}"],
    settings: {
      react: {
        version: "detect"
      }
    },
    languageOptions: {
      globals: globals.browser,
      parserOptions: {
        ecmaFeatures: {
          jsx: true
        },
        jsx: true
      }
    },
    rules: {
      // Use new JSX transform (React 17+)
      'react/jsx-uses-react': 'off',
      'react/react-in-jsx-scope': 'off'
    }
  },
  
  // TypeScript parser configuration for React (exclude config files)
  {
    files: ["**/*.{ts,tsx}"],
    languageOptions: {
      parserOptions: {
        project: tsconfigPath ? [tsconfigPath] : undefined,
        tsconfigRootDir: process.cwd()
      }
    },
    rules: {
      // TypeScript-specific rule overrides here
    }
  },
  
  // Explicitly disable type-checking for JavaScript files
  {
    files: ["**/*.{js,jsx,cjs,mjs}"],
    ...tseslint.configs.disableTypeChecked
  },
  
  // Allow devDependencies in JavaScript config files
  {
    files: ['*.config.js', '*.config.mjs', 'eslint.config.js'],
    rules: {
      'import/no-extraneous-dependencies': ['error', { devDependencies: true }],
    },
  }
];
