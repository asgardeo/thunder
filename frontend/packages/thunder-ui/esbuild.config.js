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

import {readFileSync} from 'fs';
import {build} from 'esbuild';
import {preserveDirectivesPlugin} from 'esbuild-plugin-preserve-directives';
import {resolve, dirname} from 'path';
import {fileURLToPath} from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const pkg = JSON.parse(readFileSync('./package.json', 'utf8'));

const commonOptions = {
  bundle: true,
  entryPoints: ['src/index.ts'],
  external: [...Object.keys(pkg.dependencies || {}), ...Object.keys(pkg.peerDependencies || {})],
  metafile: true,
  platform: 'browser',
  plugins: [
    dtsPlugin({
      entry: 'src/index.ts',
      output: 'dist/index.d.ts',
    }),
    preserveDirectivesPlugin({
      directives: ['use client', 'use strict'],
      include: /\.(js|ts|jsx|tsx)$/,
      exclude: /node_modules/,
    }),
  ],
  target: ['es2020'],
};

await build({
  ...commonOptions,
  format: 'esm',
  outfile: 'dist/index.js',
  sourcemap: true,
});

await build({
  ...commonOptions,
  format: 'cjs',
  outfile: 'dist/cjs/index.js',
  sourcemap: true,
});

import {generateDtsBundle} from 'dts-bundle-generator';
import {writeFileSync} from 'fs';
import path from 'path';

export function dtsPlugin(options) {
  return {
    name: 'dts-plugin',
    setup(build) {
      build.onEnd(() => {
        const output = generateDtsBundle([
          {
            filePath: options.entry,
            output: {exportReferencedTypes: false},
          },
        ]);

        const outPath = path.resolve(options.output);
        writeFileSync(outPath, output[0]);
        console.log(`✅ Types bundled into ${outPath}`);
      });
    },
  };
}
