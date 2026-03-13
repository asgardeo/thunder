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

/**
 * Group Model
 *
 * Represents a group belonging to an organization unit.
 * Groups are used to organize users within an OU for access control and permissions.
 *
 * @public
 * @remarks
 * Groups are scoped to a specific organization unit via the `organizationUnit` field.
 *
 * @example
 * ```typescript
 * const group: Group = {
 *   id: '7a1b2c3d-4e5f-6789-abcd-ef0123456789',
 *   name: 'Developers',
 *   organizationUnit: '550e8400-e29b-41d4-a716-446655440000'
 * };
 * ```
 */
export interface Group {
  /**
   * Unique identifier of the group
   * @example '7a1b2c3d-4e5f-6789-abcd-ef0123456789'
   */
  id: string;

  /**
   * Display name of the group
   * @example 'Developers'
   */
  name: string;

  /**
   * ID of the organization unit this group belongs to
   * @example '550e8400-e29b-41d4-a716-446655440000'
   */
  organizationUnit: string;
}

/**
 * Group List Response
 *
 * Response structure for paginated group list queries within an organization unit.
 *
 * @public
 * @remarks
 * This is the response structure from GET /organization-units/:id/groups endpoint.
 *
 * @example
 * ```typescript
 * const response: GroupListResponse = {
 *   totalResults: 5,
 *   startIndex: 0,
 *   count: 5,
 *   groups: [
 *     { id: 'group-1', name: 'Developers', organizationUnit: 'ou-id' }
 *   ]
 * };
 * ```
 */
export interface GroupListResponse {
  /**
   * Total number of groups available
   * @example 5
   */
  totalResults: number;

  /**
   * Starting index of the current page
   * @example 0
   */
  startIndex: number;

  /**
   * Number of groups in the current response
   * @example 5
   */
  count: number;

  /**
   * Array of groups in the current page
   */
  groups: Group[];

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
