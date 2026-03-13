/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

import type {OrganizationUnit} from './organization-unit';

/**
 * Organization Unit List Response
 *
 * Response structure for paginated organization unit list queries.
 * Contains pagination metadata along with the list of organization units.
 *
 * @public
 * @remarks
 * This is the response structure from GET /organization-units endpoint.
 * Includes total count, start index, and current page count for pagination support.
 *
 * @example
 * ```typescript
 * const response: OrganizationUnitListResponse = {
 *   totalResults: 50,
 *   startIndex: 0,
 *   count: 10,
 *   organizationUnits: [
 *     {
 *       id: '550e8400-e29b-41d4-a716-446655440000',
 *       handle: 'engineering',
 *       name: 'Engineering Department',
 *       description: 'Software engineering team'
 *     }
 *   ],
 *   links: [{ rel: 'next', href: '/organization-units?offset=10&limit=10' }]
 * };
 * ```
 */
export interface OrganizationUnitListResponse {
  /**
   * Total number of organization units available
   * @example 50
   */
  totalResults: number;

  /**
   * Starting index of the current page
   * @example 0
   */
  startIndex: number;

  /**
   * Number of organization units in the current response
   * @example 10
   */
  count: number;

  /**
   * Array of organization units in the current page
   */
  organizationUnits: OrganizationUnit[];

  /**
   * Pagination links for navigating between pages
   */
  links?: {
    /** Link relation type (e.g., 'next', 'prev') */
    rel: string;
    /** Link URL */
    href: string;
  }[];
}
