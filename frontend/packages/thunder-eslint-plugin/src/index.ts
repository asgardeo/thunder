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

import {readFileSync} from 'fs';
import {dirname, join} from 'path';
import {fileURLToPath} from 'url';
import type {ESLint} from 'eslint';
import baseConfig from './configs/base';
import javascriptConfig from './configs/javascript';
import playwrightConfig from './configs/playwright';
import prettierConfig from './configs/prettier';
import reactConfig from './configs/react';
import typescriptConfig from './configs/typescript';
import vitestConfig from './configs/vitest';
import vueConfig from './configs/vue';
import copyrightHeaderRule from './rules/copyright-header';
import createParserOptions from './utils/tsconfig-resolver';

interface PackageJson {
  name: string;
  version: string;
}

const __filename: string = fileURLToPath(import.meta.url);
const __dirname: string = dirname(__filename);
const pkgPath: string = join(__dirname, '../package.json');

const pkg: PackageJson = JSON.parse(readFileSync(pkgPath, 'utf8')) as PackageJson;
const namespace = 'thunder';

const DEV_DEPENDENCIES_ALLOWED_FILES: string[] = [
  '*.config.js',
  '*.config.mjs',
  '*.config.ts',
  '**/eslint.config.js',
  '**/eslint.config.mjs',
  '**/eslint.config.ts',
  '**/rolldown.config.js',
  '**/rolldown.config.ts',
  'vite.config.ts',
  'vitest.config.ts',
  '**/prettier.config.js',
  '**/prettier.config.mjs',
  '**/prettier.config.ts',
  '**/test/**',
];

const plugin: ESLint.Plugin = {
  meta: {
    name: pkg.name,
    version: pkg.version,
    namespace,
  },
  configs: {},
  rules: {
    'copyright-header': copyrightHeaderRule,
  },
  processors: {},
};

// Assign configs here so we can reference `plugin`
// @ts-expect-error TODO: Update to the latest ESLint and remove `@types/eslint`.

Object.assign(plugin.configs, {
  javascript: [
    {
      name: 'thunder/plugin-setup',
      plugins: {
        '@thunder': plugin,
      },
    },
    ...baseConfig,

    ...javascriptConfig,
    ...prettierConfig,
    {
      files: DEV_DEPENDENCIES_ALLOWED_FILES,
      rules: {
        'import/no-extraneous-dependencies': ['error', {devDependencies: true}],
      },
    },
  ],
  typescript: [
    {
      name: 'thunder/plugin-setup',
      plugins: {
        '@thunder': plugin,
      },
    },
    ...baseConfig,

    ...javascriptConfig,

    ...typescriptConfig,
    ...prettierConfig,
    {
      files: DEV_DEPENDENCIES_ALLOWED_FILES,
      rules: {
        'import/no-extraneous-dependencies': ['error', {devDependencies: true}],
      },
    },
  ],
  react: [
    {
      name: 'thunder/plugin-setup',
      plugins: {
        '@thunder': plugin,
      },
    },
    ...baseConfig,

    ...javascriptConfig,

    ...typescriptConfig,

    ...reactConfig,
    ...prettierConfig,
    {
      files: DEV_DEPENDENCIES_ALLOWED_FILES,
      rules: {
        'import/no-extraneous-dependencies': ['error', {devDependencies: true}],
      },
    },
  ],
  // Overlay config for Playwright e2e test files — spread alongside base/react.
  playwright: [
    {
      name: 'thunder/plugin-setup',
      plugins: {
        '@thunder': plugin,
      },
    },
    ...playwrightConfig,
  ],
  // Overlay config for Vitest unit/integration test files — spread alongside base/react.
  vitest: [
    {
      name: 'thunder/plugin-setup',
      plugins: {
        '@thunder': plugin,
      },
    },
    ...vitestConfig,
  ],
  // Full project config for Vue applications.
  vue: [
    {
      name: 'thunder/plugin-setup',
      plugins: {
        '@thunder': plugin,
      },
    },
    ...baseConfig,

    ...javascriptConfig,

    ...typescriptConfig,

    ...vueConfig,
    ...prettierConfig,
    {
      files: DEV_DEPENDENCIES_ALLOWED_FILES,
      rules: {
        'import/no-extraneous-dependencies': ['error', {devDependencies: true}],
      },
    },
  ],
});

export default plugin;
export {createParserOptions};
