/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

import vitestPlugin from '@vitest/eslint-plugin';
import type {Linter} from 'eslint';

const vitestConfig: Linter.Config[] = [
  {
    ...vitestPlugin.configs.recommended,
    files: ['**/*.test.ts', '**/*.test.tsx', '**/*.spec.ts', '**/*.spec.tsx', '**/test/**'],
    rules: {
      ...vitestPlugin.configs.recommended.rules,
      'vitest/expect-expect': ['error', {assertFunctionNames: ['expect', 'expectTypeOf']}],
    },
  },
];

export default vitestConfig;
