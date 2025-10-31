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

import {FlatCompat} from '@eslint/eslintrc';
import type {Linter} from 'eslint';
import path from 'path';
import {fileURLToPath} from 'url';

const __filename: string = fileURLToPath(import.meta.url);
const __dirname: string = path.dirname(__filename);

const compat: FlatCompat = new FlatCompat({
  baseDirectory: __dirname,
});

const javascriptConfig: Linter.Config[] = [
  // FlatCompat.extends() returns Config[] but @eslint/eslintrc doesn't provide proper types
  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
  ...compat.extends('airbnb-base'),
  {
    name: 'thunder/javascript-overrides',
    rules: {
      'object-curly-spacing': ['error', 'never'],
      // Modify the order a bit to make the imports more readable.
      // https://github.com/import-js/eslint-plugin-import/blob/main/docs/rules/order.md
      'import/order': [
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
      'import/extensions': [
        'error',
        'ignorePackages',
        {
          js: 'never',
          jsx: 'never',
        },
      ],
    },
  },
];

export default javascriptConfig;
