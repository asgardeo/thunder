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

/* eslint-disable @typescript-eslint/naming-convention */
/* eslint-disable no-underscore-dangle */

import type {ESLint} from 'eslint';
import {readFileSync} from 'fs';
import {fileURLToPath} from 'url';
import {dirname, join} from 'path';
import copyrightHeaderRule from './rules/copyright-header.js';
import baseConfig from './configs/base.js';
import reactConfig from './configs/react.js';
import javascriptConfig from './configs/javascript.js';
import prettierConfig from './configs/prettier.js';
import typescriptConfig from './configs/typescript.js';

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
    // @ts-expect-error TODO: Update to the latest ESLint and remove `@types/eslint`.
    namespace,
  },
  configs: {},
  rules: {
    // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
    'copyright-header': copyrightHeaderRule,
  },
  processors: {},
};

// Assign configs here so we can reference `plugin`
// @ts-expect-error TODO: Update to the latest ESLint and remove `@types/eslint`.
// eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
Object.assign(plugin.configs, {
  base: [
    {
      name: 'thunder/plugin-setup',
      plugins: {
        // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
        '@thunder': plugin,
      },
    },
    ...baseConfig,
    // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
    ...javascriptConfig,
    // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
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
        // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
        '@thunder': plugin,
      },
    },
    ...baseConfig,
    // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
    ...javascriptConfig,
    // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
    ...typescriptConfig,
    // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
    ...reactConfig,
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
