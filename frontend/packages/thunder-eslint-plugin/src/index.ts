
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

import type { ESLint, Rule } from 'eslint';
import { readFileSync } from 'fs';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';
import { copyrightHeaderRule } from './rules/copyright-header.js';
import { baseConfig } from './configs/base.js';
import { javascriptConfig } from './configs/javascript.js';
import { prettierConfig } from './configs/prettier.js';
import { reactConfig } from './configs/react.js';
import { typescriptConfig } from './configs/typescript.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const pkgPath = join(__dirname, '../package.json');
const pkg = JSON.parse(readFileSync(pkgPath, 'utf8'));
const namespace = 'thunder';

const plugin: ESLint.Plugin = {
  meta: {
    name: pkg.name,
    version: pkg.version,
    namespace
  },
  configs: {},
  rules: {
    'copyright-header': copyrightHeaderRule
  },
  processors: {},
};

// Assign configs here so we can reference `plugin`
Object.assign(plugin.configs, {
  base: [
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
  ],
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
  ],
});

export default plugin;
