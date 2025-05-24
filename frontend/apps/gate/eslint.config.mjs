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

const LICENSE_HEADER_DEFAULT_PATTERN = new RegExp(
  `\\*\\*\\n` +
  ` \\* Copyright \\(c\\) \\d{4}, WSO2 LLC\\. \\(https://www\\.wso2\\.com\\)\\.\\n` +
  ` \\*\\n` +
  ` \\* WSO2 LLC\\. licenses this file to you under the Apache License,\\n` +
  ` \\* Version 2\\.0 \\(the "License"\\); you may not use this file except\\n` +
  ` \\* in compliance with the License\\.\\n` +
  ` \\* You may obtain a copy of the License at\\n` +
  ` \\*\\n` +
  ` \\* http://www\\.apache\\.org/licenses/LICENSE-2\\.0\\n` +
  ` \\*\\n` +
  ` \\* Unless required by applicable law or agreed to in writing,\\n` +
  ` \\* software distributed under the License is distributed on an\\n` +
  ` \\* "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY\\n` +
  ` \\* KIND, either express or implied\\. See the License for the\\n` +
  ` \\* specific language governing permissions and limitations\\n` +
  ` \\* under the License\\.\\n` +
  ` \\*/`,
  'm'
);

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
    rules: {
      "headers/header-format": [
        "error",
        {
          source: LICENSE_HEADER_DEFAULT_PATTERN,
          style: "jsdoc",
          patterns: {
            year: {
              pattern: "\\d{4}",
              defaultValue: "2025",
            },
          },
          preservePragmas: true,
          enableVueSupport: false,
        },
      ],
      '@typescript-eslint/typedef': 'off',
    },
  },
];

const eslintConfig = fixupConfigRules(baseConfigs);

export default eslintConfig;
