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

import type {Linter} from 'eslint';
import {createTypeScriptImportResolver} from 'eslint-import-resolver-typescript';
import tseslint from 'typescript-eslint';
import createParserOptions from '../utils/tsconfig-resolver';

const typescriptConfig: Linter.Config[] = [
  ...tseslint.configs.recommendedTypeChecked,
  ...tseslint.configs.stylisticTypeChecked,
  {
    languageOptions: {
      ecmaVersion: 2020,
      parserOptions: createParserOptions(),
    },
  },
  {
    files: ['**/*.{js,jsx,cjs,mjs}'],
    ...tseslint.configs.disableTypeChecked,
  },
  {
    name: 'thunderid/typescript-resolver',
    settings: {
      'import-x/resolver-next': [createTypeScriptImportResolver({alwaysTryTypes: true})],
    },
  },
  {
    name: 'thunderid/typescript-overrides',
    rules: {
      // Disallow the use of the `any` type to encourage more precise typings.
      // https://typescript-eslint.io/rules/no-explicit-any/
      '@typescript-eslint/no-explicit-any': 'error',
      'object-curly-spacing': ['error', 'never'],
      // Allow imports without file extensions for TypeScript files
      // This is especially useful for path aliases and modern module resolution
      'import-x/extensions': [
        'error',
        'ignorePackages',
        {
          js: 'never',
          jsx: 'never',
          ts: 'never',
          tsx: 'never',
        },
      ],
    },
  },
];

export default typescriptConfig;
