/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import { dirname } from 'path';
import { fileURLToPath } from 'url';
import { FlatCompat } from '@eslint/eslintrc';
import { fixupConfigRules } from '@eslint/compat';
import headers from 'eslint-plugin-headers';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const LICENSE_HEADER_DEFAULT_PATTERN = `Copyright (c) {year}, {company}. ({url}).

{company}. licenses this file to you under the Apache License,
Version 2.0 (the "License"); you may not use this file except
in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the
specific language governing permissions and limitations
under the License.`;

const compat = new FlatCompat({
  baseDirectory: __dirname,
});

const baseConfigs = [
  ...compat.extends(
    'next/core-web-vitals',
    'next/typescript',
    'plugin:@wso2/strict',
    'plugin:prettier/recommended'
  ),
  {
    plugins: {
      headers: headers,
    },
    files: ['**/*.ts', '**/*.tsx'],
    rules: {
      'headers/header-format': [
        'error',
        {
          source: 'string',
          content: LICENSE_HEADER_DEFAULT_PATTERN,
          variables: {
            'year': '2025',
            'company': "WSO2 LLC",
            'url': "https://www.wso2.com"
          },
          trailingNewlines: 2
        }
      ],
      '@typescript-eslint/typedef': 'off',
      // TODO: Temporarily disable this rule until the plugin 
      // false positives are resolved
      'headers/header-format': 'off',
      'no-unused-vars': ['error', { 'argsIgnorePattern': '^_' }]
    }
  },
];

const eslintConfig = fixupConfigRules(baseConfigs);

export default eslintConfig;
