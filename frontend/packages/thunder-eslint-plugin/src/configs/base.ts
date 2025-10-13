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

import javascriptConfig from './javascript.js';
import prettierConfig from './prettier.js';
import typescriptConfig from './typescript.js';

/**
 * @fileoverview Base ESLint config that includes header, JavaScript, Prettier, and TypeScript configurations.
 */

const baseConfig = [
  // Thunder copyright header rule will be added by the main plugin
  {
    name: 'thunder/copyright-header',
    rules: {'@thunder/copyright-header': 'error'},
  },
  ...javascriptConfig,
  ...typescriptConfig,
  ...prettierConfig,
  {
    files: [
      '*.config.js',
      '*.config.mjs',
      '*.config.ts',
      '**/eslint.config.js',
      '**/eslint.config.mjs',
      '**/eslint.config.ts',
      'vite.config.ts',
      'vitest.config.ts',
      '**/prettier.config.js',
      '**/prettier.config.mjs',
      '**/prettier.config.ts',
    ],
    rules: {
      'import/no-extraneous-dependencies': ['error', {devDependencies: true}],
    },
  },
];

export default baseConfig;
