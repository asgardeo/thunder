/*
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
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

/**
 * Playwright E2E Test Configuration
 * 
 * This configuration sets up test projects for Chromium, Firefox, and Webkit.
 * All projects depend on the `setup` project for authentication.
 * 
 * Reports are generated in both HTML and Blob format (for merging).
 * 
 * @see https://playwright.dev/docs/test-configuration
 */

import { defineConfig, devices } from '@playwright/test';
import dotenv from 'dotenv';
import path from 'path';

const envPath = path.resolve(__dirname, '.env');
dotenv.config({ path: envPath });

export default defineConfig({
  /** Directory containing test files */
  testDir: './tests',

  /** Run tests sequentially to avoid auth conflicts */
  fullyParallel: false,

  /** Fail CI builds if test.only() is accidentally committed */
  forbidOnly: !!process.env.CI,

  /** Retry failed tests (more on CI) */
  retries: process.env.CI ? 2 : 1,

  /** Single worker for sequential execution */
  workers: 1,

  /** Generate HTML report, Console list, and Blob report for merging */
  reporter: [
    ['html'],
    ['list'],
    ['blob']
  ],

  /** Global test timeout */
  timeout: 60000,

  /** Shared settings for all projects */
  use: {
    trace: 'retain-on-failure',
    ignoreHTTPSErrors: true,
    screenshot: 'only-on-failure',
    video: 'on',
    actionTimeout: 15000,
    baseURL: process.env.BASE_URL,
  },

  projects: [
    /** Setup project - only runs auth.setup.ts */
    {
      name: 'setup',
      testMatch: '**/*.setup.ts',
      use: { ...devices['Desktop Chrome'] },
    },

    /** Main test project - runs .spec.ts files with authenticated session */
    {
      name: 'chromium',
      testMatch: '**/*.spec.ts',
      use: {
        ...devices['Desktop Chrome'],
        storageState: 'playwright/.auth/devportal-admin.json',
      },
      dependencies: ['setup'],
    },

    {
      name: 'firefox',
      testMatch: '**/*.spec.ts',
      use: {
        ...devices['Desktop Firefox'],
        storageState: 'playwright/.auth/devportal-admin.json',
      },
      dependencies: ['setup'],
    },

    {
      name: 'webkit',
      testMatch: '**/*.spec.ts',
      use: {
        ...devices['Desktop Safari'],
        storageState: 'playwright/.auth/devportal-admin.json',
      },
      dependencies: ['setup'],
    },
  ],
});
