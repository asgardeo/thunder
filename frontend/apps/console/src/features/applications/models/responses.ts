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

import type { BasicApplication } from "./application";

/**
 * Application List Response
 *
 * Response structure for paginated application list queries.
 * Contains pagination metadata along with the list of applications.
 *
 * @public
 * @remarks
 * This is the response structure from GET /applications endpoint.
 * Includes total count and current page count for pagination support.
 *
 * @example
 * ```typescript
 * const response: ApplicationListResponse = {
 *   totalResults: 25,
 *   count: 10,
 *   applications: [
 *     {
 *       id: '550e8400-e29b-41d4-a716-446655440000',
 *       name: 'My Web App',
 *       description: 'Customer portal',
 *       client_id: 'my_client_id',
 *       is_registration_flow_enabled: true
 *     }
 *   ]
 * };
 * ```
 */
export interface ApplicationListResponse {
  /**
   * Total number of applications available
   * @example 25
   */
  totalResults: number;

  /**
   * Number of applications in the current response
   * @example 10
   */
  count: number;

  /**
   * Array of basic application information
   */
  applications: BasicApplication[];
}
