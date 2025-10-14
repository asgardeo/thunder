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

import path from 'path';
import {fileURLToPath} from 'url';
import {FlatCompat} from '@eslint/eslintrc';
import type {Linter} from 'eslint';
import reactRefresh from 'eslint-plugin-react-refresh';
import globals from 'globals';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const compat = new FlatCompat({
  baseDirectory: __dirname,
});

const reactConfig: Linter.Config[] = [
  // React-specific configs for all files
  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
  ...compat.extends('airbnb'),
  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
  ...compat.extends('airbnb/hooks'),
  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
  ...compat.extends('@kesills/airbnb-typescript'),
  reactRefresh.configs.recommended,
  {
    languageOptions: {
      ecmaVersion: 2020,
      globals: globals.browser,
      parserOptions: {
        projectService: true,
        tsconfigRootDir: import.meta.dirname,
      },
    },
  },
  {
    name: 'thunder/react-overrides',
    rules: {
      'react/jsx-props-no-spreading': 'off',
    },
  },
];

export default reactConfig;
