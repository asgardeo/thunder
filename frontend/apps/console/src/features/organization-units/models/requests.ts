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
 * Request body for creating an organization unit.
 *
 * Derived from {@link OrganizationUnit} by picking only the fields
 * required for creation. Server-generated fields (id, theme_id, layout_id, logo_url)
 * are excluded.
 *
 * @public
 * @remarks
 * Used with POST /organization-units endpoint.
 *
 * @example
 * ```typescript
 * const request: CreateOrganizationUnitRequest = {
 *   handle: 'engineering',
 *   name: 'Engineering Department',
 *   description: 'Software engineering team',
 *   parent: 'root-ou-id'
 * };
 * ```
 */
export type CreateOrganizationUnitRequest = Pick<OrganizationUnit, 'handle' | 'name' | 'description' | 'parent'>;

/**
 * Request body for updating an organization unit.
 *
 * Derived from {@link OrganizationUnit} by omitting only the server-managed
 * `id` and `layout_id` fields. All other fields can be updated.
 *
 * @public
 * @remarks
 * Used with PUT /organization-units/:id endpoint.
 *
 * @example
 * ```typescript
 * const request: UpdateOrganizationUnitRequest = {
 *   handle: 'engineering',
 *   name: 'Engineering Department (Updated)',
 *   description: 'Updated description',
 *   parent: 'root-ou-id',
 *   theme_id: '96c62e6d-9297-4295-8195-d28dfe0c9ff7',
 *   logo_url: 'https://example.com/new-logo.png'
 * };
 * ```
 */
export type UpdateOrganizationUnitRequest = Omit<OrganizationUnit, 'id' | 'layout_id'>;

/**
 * Query parameters for listing organization units.
 *
 * Supports pagination through limit and offset parameters.
 *
 * @public
 * @remarks
 * Used with GET /organization-units endpoint.
 *
 * @example
 * ```typescript
 * const params: OrganizationUnitListParams = {
 *   limit: 10,
 *   offset: 0
 * };
 * ```
 */
export interface OrganizationUnitListParams {
  /**
   * Maximum number of results to return
   * @example 10
   */
  limit?: number;

  /**
   * Number of results to skip for pagination
   * @example 0
   */
  offset?: number;
}
