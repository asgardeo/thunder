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

import {defineConfig} from 'rolldown';
import {readFileSync} from 'fs';
import {join} from 'path';

const pkg = JSON.parse(readFileSync('./package.json', 'utf8'));

const external = [
  ...Object.keys(pkg.dependencies || {}),
  ...Object.keys(pkg.peerDependencies || {}),
  'react/jsx-runtime',
  // Needed to avoid hook ordering issues.
  /^@mui\//,
  /^@thunder\//,
  /^@wso2\//,
  /^@tanstack\//,
  /^@testing-library\//,
  /^@vitest\//,
  /^vitest/,
  'i18next',
  'react-i18next',
  'react-router',
];

// Common input files for jsdom mode (existing)
const jsdomInput = {
  index: join('src', 'index.ts'),
  setup: join('src', 'setup.ts'),
  'mocks/index': join('src', 'mocks', 'index.ts'),
};

// Browser mode input files (new)
const browserInput = {
  browser: join('src', 'browser.ts'),
  'setup-browser': join('src', 'setup-browser.ts'),
};

const commonOptions = {
  external,
  target: 'es2020',
  sourcemap: true,
};

export default defineConfig([
  // ESM build for jsdom mode (existing)
  {
    ...commonOptions,
    input: jsdomInput,
    platform: 'browser',
    output: {
      dir: 'dist',
      format: 'esm',
    },
  },
  // CommonJS build for jsdom mode (existing)
  {
    ...commonOptions,
    input: jsdomInput,
    platform: 'node',
    output: {
      dir: join('dist', 'cjs'),
      entryFileNames: '[name].cjs',
      format: 'cjs',
    },
  },
  // ESM build for browser mode - Vitest Browser Mode runs tests in real browsers
  // which natively support ES modules, so CJS is not needed for this build target
  {
    ...commonOptions,
    input: browserInput,
    platform: 'browser',
    output: {
      dir: 'dist',
      format: 'esm',
    },
  },
]);
