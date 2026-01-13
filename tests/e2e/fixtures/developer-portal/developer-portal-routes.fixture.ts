/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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
 * import { test } from '../fixtures/developer-portal';
 *
 * test('navigate', async ({ page, routes }) => {
 *   await page.goto(`${baseUrl}${routes.users}`);
 * });
 */

import { test as base } from "@playwright/test";

import { DeveloperPortalRoutes } from "../../configs/routes/developer-portal-routes";
export { DeveloperPortalRoutes };

type RoutesFixture = {
  routes: typeof DeveloperPortalRoutes;
};

export const test = base.extend<RoutesFixture>({
  routes: async ({}, use) => {
    await use(DeveloperPortalRoutes);
  },
});

export { expect } from "@playwright/test";
export const routes = DeveloperPortalRoutes;
