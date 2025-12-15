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
 * Developer Portal Authentication Fixture
 * 
 * Provides an `authenticatedPage` fixture that automatically ensures
 * the test page is authenticated before the test runs.
 * 
 * This replaces the manual `setupAuthentication` call in beforeEach.
 * 
 * @example
 * import { test } from '../fixtures';
 * 
 * test('authenticated test', async ({ authenticatedPage }) => {
 *   // authenticatedPage is already logged in
 *   await authenticatedPage.goto(routes.users);
 * });
 */

import { test as base, Page } from '@playwright/test';
import { setupAuthentication } from '../../utils/authentication';

const baseUrl = process.env.BASE_URL as string;

type AuthFixtures = {
    authenticatedPage: Page;
};

export const test = base.extend<AuthFixtures>({
    authenticatedPage: async ({ page }, use) => {
        // Setup authentication before test usage
        await setupAuthentication(page, baseUrl, { debug: true });

        // Provide the authenticated page to the test
        await use(page);
    },
});
