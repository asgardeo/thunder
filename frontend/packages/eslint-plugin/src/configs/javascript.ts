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

import eslint from '@eslint/js';
import type {Linter} from 'eslint';
import {importX} from 'eslint-plugin-import-x';

const javascriptConfig: Linter.Config[] = [
  eslint.configs.recommended,
  importX.flatConfigs.recommended,
  {
    name: 'thunderid/javascript-overrides',
    rules: {
      // Disallow the use of console in-favor of the Logger.
      // https://eslint.org/docs/latest/rules/no-console
      'no-console': 'error',
      // Disallow new operators outside of assignments or comparisons
      // https://eslint.org/docs/latest/rules/no-new
      'no-new': 'error',
      'object-curly-spacing': ['error', 'never'],
      // Modify the order a bit to make the imports more readable.
      // https://github.com/un-ts/eslint-plugin-import-x/blob/master/docs/rules/order.md
      'import-x/order': [
        'warn',
        {
          alphabetize: {
            caseInsensitive: true,
            order: 'asc',
          },
          groups: ['builtin', 'external', 'index', 'sibling', 'parent', 'internal'],
        },
      ],
      // Allow imports without file extensions for JavaScript files
      // This is especially useful for path aliases and modern module resolution
      'import-x/extensions': [
        'error',
        'ignorePackages',
        {
          js: 'never',
          jsx: 'never',
        },
      ],
      // Enforce no cycles in imports to prevent circular dependencies
      // https://github.com/un-ts/eslint-plugin-import-x/blob/master/docs/rules/no-cycle.md
      'import-x/no-cycle': 'error',
    },
  },
];

export default javascriptConfig;
