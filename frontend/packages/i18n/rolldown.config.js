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
import {join, dirname} from 'path';
import {fileURLToPath} from 'url';

const currentDir = dirname(fileURLToPath(import.meta.url));
const pkg = JSON.parse(readFileSync(join(currentDir, 'package.json'), 'utf8'));

const external = [...Object.keys(pkg.dependencies || {}), ...Object.keys(pkg.peerDependencies || {})];

const commonOptions = {
  input: {
    index: join('src', 'index.ts'),
    'locales/en-US': join('src', 'locales', 'en-US.ts'),
  },
  external,
  target: 'es2020',
  sourcemap: true,
};

export default defineConfig([
  // ✅ ESM build (for browsers/bundlers)
  {
    ...commonOptions,
    platform: 'browser',
    output: {
      dir: 'dist',
      format: 'esm',
      entryFileNames: '[name].js',
    },
  },
  // ✅ CommonJS build (for Node/SSR/testing)
  {
    ...commonOptions,
    platform: 'node',
    output: {
      dir: 'dist',
      format: 'cjs',
      entryFileNames: '[name].cjs',
    },
  },
]);
