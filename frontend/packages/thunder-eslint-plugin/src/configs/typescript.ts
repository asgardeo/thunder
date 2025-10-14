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
import tseslint from 'typescript-eslint';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const compat = new FlatCompat({
  baseDirectory: __dirname,
});

const typescriptConfig: Linter.Config[] = [
  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
  ...compat.extends('airbnb-base'),
  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
  ...compat.extends('airbnb/hooks'),
  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
  ...compat.extends('@kesills/airbnb-typescript/base'),
  ...tseslint.configs.recommendedTypeChecked,
  ...tseslint.configs.stylisticTypeChecked,
  {
    languageOptions: {
      ecmaVersion: 2020,
      parserOptions: {
        projectService: true,
        tsconfigRootDir: import.meta.dirname,
      },
    },
  },
  {
    files: ['**/*.{js,jsx,cjs,mjs}'],
    ...tseslint.configs.disableTypeChecked,
  },
  {
    name: 'thunder/typescript-overrides',
    rules: {
      'object-curly-spacing': ['error', 'never'],
    },
  },
];

export default typescriptConfig;
