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
 * Developer Portal Routes Fixture
 * 
 * Provides centralized route definitions as a Playwright fixture.
 * 
 * @example
 * import { test } from '../fixtures/developer-portal-routes.fixture';
 * 
 * test('navigate', async ({ page, routes }) => {
 *   await page.goto(`${baseUrl}${routes.users}`);
 * });
 */

import { test as base } from '@playwright/test';

export const DeveloperPortalRoutes = {
    /** Sign-in page route */
    signin: '/gate/signin',

    /** Sign-out page route */
    signout: '/gate/signout',

    /** Developer portal home page */
    home: '/develop',

    /** Dashboard page */
    dashboard: '/develop/dashboard',

    /** Applications list page */
    applications: '/develop/applications',

    /** Create new application page */
    applicationCreate: '/develop/applications/create',

    /** Application details page */
    applicationDetails: (appId: string) => `/develop/applications/${appId}`,

    /** APIs list page */
    apis: '/develop/apis',

    /** API details page */
    apiDetails: (apiId: string) => `/develop/apis/${apiId}`,

    /** Users list page */
    users: '/develop/users',

    /** Create new user page */
    userCreate: '/develop/users/create',

    /** User details page */
    userDetails: (userId: string) => `/develop/users/${userId}`,

    /** Settings page */
    settings: '/develop/settings',

    /** User profile settings page */
    profile: '/develop/settings/profile',
} as const;

type RoutesFixture = {
    routes: typeof DeveloperPortalRoutes;
};

export const test = base.extend<RoutesFixture>({
    routes: async ({ }, use) => {
        await use(DeveloperPortalRoutes);
    },
});

export { expect } from '@playwright/test';
export const routes = DeveloperPortalRoutes;
