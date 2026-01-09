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
 * Developer Portal Routes Configuration
 *
 * Centralized route definitions for the WSO2 Developer Portal.
 * All route paths should be defined here to ensure consistency across tests.
 *
 * @example
 * import { routes } from '../../fixtures';
 * await page.goto(`${baseUrl}${routes.home}`);
 *
 * @example
 * import { DeveloperPortalRoutes } from '../configs/routes/developer-portal-routes';
 * await page.goto(`${baseUrl}${DeveloperPortalRoutes.applications}`);
 */
export const DeveloperPortalRoutes = {
  /** Sign-in page route */
  signin: "/gate/signin",

  /** Sign-out page route */
  signout: "/gate/signout",

  /** Developer portal home page */
  home: "/develop",

  /** Dashboard page */
  dashboard: "/develop/dashboard",

  /** Applications list page */
  applications: "/develop/applications",

  /** Create new application page */
  applicationCreate: "/develop/applications/create",

  /**
   * Application details page
   * @param appId - The application identifier
   */
  applicationDetails: (appId: string) => `/develop/applications/${appId}`,

  /** APIs list page */
  apis: "/develop/apis",

  /**
   * API details page
   * @param apiId - The API identifier
   */
  apiDetails: (apiId: string) => `/develop/apis/${apiId}`,

  /** Users list page */
  users: "/develop/users",

  /** Create new user page */
  userCreate: "/develop/users/create",

  /**
   * User details page
   * @param userId - The user identifier
   */
  userDetails: (userId: string) => `/develop/users/${userId}`,

  /** Settings page */
  settings: "/develop/settings",

  /** User profile settings page */
  profile: "/develop/settings/profile",
} as const;

export type DeveloperPortalRoute = (typeof DeveloperPortalRoutes)[keyof typeof DeveloperPortalRoutes];

export default DeveloperPortalRoutes;
