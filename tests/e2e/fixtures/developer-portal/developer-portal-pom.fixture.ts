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
 * Developer Portal Page Object Model Fixture
 * 
 * Provides page object models as Playwright fixtures.
 * 
 * - `signinPage`: Uses standard `page` (no auth required)
 * - `usersPage`: Uses `authenticatedPage` (enforces auth)
 */

import { test as base } from './developer-portal-auth.fixture';
import { DeveloperPortalSigninPage } from '../../pages/authentication';
import { UsersPage } from '../../pages/user-management';

const baseUrl = process.env.BASE_URL || '';

type POMFixtures = {
    signinPage: DeveloperPortalSigninPage;
    usersPage: UsersPage;
};

export const test = base.extend<POMFixtures>({
    // Signin page does NOT need auth, uses raw page
    signinPage: async ({ page }, use) => {
        await use(new DeveloperPortalSigninPage(page, baseUrl));
    },

    // Users page requires auth, uses authenticatedPage fixture
    usersPage: async ({ authenticatedPage }, use) => {
        await use(new UsersPage(authenticatedPage, baseUrl));
    },
});

export { expect } from '@playwright/test';
export { DeveloperPortalSigninPage } from '../../pages/authentication';
export { UsersPage, type UserFormData } from '../../pages/user-management';
