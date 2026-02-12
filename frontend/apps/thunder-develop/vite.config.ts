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

import {defineConfig} from 'vitest/config';
import react from '@vitejs/plugin-react';
import {resolve, dirname} from 'path';
import {fileURLToPath} from 'url';
import basicSsl from '@vitejs/plugin-basic-ssl';
import {visualizer} from 'rollup-plugin-visualizer';

const currentDir = dirname(fileURLToPath(import.meta.url));
const PORT = process.env.PORT ? Number(process.env.PORT) : 5191;
const HOST = process.env.HOST ?? 'localhost';
const BASE_URL = process.env.BASE_URL ?? '/develop';

// https://vite.dev/config/
export default defineConfig({
  base: BASE_URL,
  plugins: [
    basicSsl(),
    react({
      babel: {
        plugins: [['babel-plugin-react-compiler']],
      },
    }),
    // Add visualizer plugin for bundle analysis (only when ANALYZE=true)
    ...(process.env.ANALYZE === 'true'
      ? [
          visualizer({
            filename: resolve(currentDir, 'dist', 'stats.html'),
            open: true,
            gzipSize: true,
            brotliSize: true,
          }),
        ]
      : []),
  ],
  server: {
    port: PORT,
    host: HOST,
  },
  resolve: {
    alias: {
      '@': resolve(currentDir, 'src'),
      '@/components': resolve(currentDir, 'src', 'components'),
      '@/layouts': resolve(currentDir, 'src', 'layouts'),
      '@/theme': resolve(currentDir, 'src', 'theme'),
      '@/contexts': resolve(currentDir, 'src', 'contexts'),
      '@/lib': resolve(currentDir, 'src', 'lib'),
      '@/hooks': resolve(currentDir, 'src', 'hooks'),
      '@/types': resolve(currentDir, 'src', 'types'),
      // Force using the same React instance to avoid "Invalid hook call" errors
      // when using linked packages
      react: resolve(__dirname, './node_modules/react'),
      'react-dom': resolve(__dirname, './node_modules/react-dom'),
      '@emotion/react': resolve(__dirname, './node_modules/@emotion/react'),
      '@emotion/styled': resolve(__dirname, './node_modules/@emotion/styled'),
    },
    conditions: ['browser', 'module', 'import', 'default'],
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['@thunder/test-utils/setup'],
    css: {
      modules: {
        classNameStrategy: 'non-scoped',
      },
    },
    // Added to fix test failures dues to css imports from oxygen-ui. Need to check if can fix from oxygen-ui library.
    server: {
      deps: {
        inline: [
          '@asgardeo/browser',
          '@asgardeo/react',
          '@wso2/oxygen-ui',
          '@wso2/oxygen-ui-icons-react',
          '@mui/x-data-grid',
          '@mui/x-date-pickers',
          '@mui/x-tree-view',
          '@mui/x-charts',
        ],
      },
    },
    coverage: {
      provider: 'istanbul',
      reporter: ['text', 'json', 'html', ['lcov', {projectRoot: resolve(currentDir, '..', '..', '..')}]],
      exclude: [
        'node_modules/',
        'dist/',
        'public/',
        'coverage/',
        'src/test/',
        '**/*.d.ts',
        '**/*.config.*',
        '**/mockData',
        '**/*.type.ts',
        '**/*.test.ts',
        '**/*.test.tsx',
        '**/*.spec.ts',
        '**/*.spec.tsx',
        '**/EditTokenSettings.tsx',
      ],
    },
  },
});
