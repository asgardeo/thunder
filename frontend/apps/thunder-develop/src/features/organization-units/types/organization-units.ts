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

/**
 * Organization Unit type definition
 */
export interface OrganizationUnit {
  id: string;
  handle: string;
  name: string;
  description?: string | null;
  parent?: string | null;
}

/**
 * List response for organization units
 */
export interface OrganizationUnitListResponse {
  totalResults: number;
  startIndex: number;
  count: number;
  organizationUnits: OrganizationUnit[];
  links?: {
    rel: string;
    href: string;
  }[];
}

/**
 * Query parameters for listing organization units
 */
export interface OrganizationUnitListParams {
  limit?: number;
  offset?: number;
}

/**
 * Standard API error structure
 */
export interface ApiError {
  code: string;
  message: string;
  description: string;
}

/**
 * Request body for creating an organization unit
 */
export interface CreateOrganizationUnitRequest {
  handle: string;
  name: string;
  description?: string | null;
  parent?: string | null;
}

/**
 * Request body for updating an organization unit
 */
export interface UpdateOrganizationUnitRequest {
  handle: string;
  name: string;
  description?: string | null;
  parent?: string | null;
}

/**
 * Group type definition for OU groups
 */
export interface Group {
  id: string;
  name: string;
  organizationUnit: string;
}

/**
 * List response for groups
 */
export interface GroupListResponse {
  totalResults: number;
  startIndex: number;
  count: number;
  groups: Group[];
  links?: {
    rel: string;
    href: string;
  }[];
}

/**
 * Navigation state passed when navigating between organization units.
 * Used to track the source OU for proper back navigation.
 */
export interface OUNavigationState {
  fromOU: {
    id: string;
    name: string;
  };
}
