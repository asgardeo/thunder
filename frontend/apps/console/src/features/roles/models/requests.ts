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

import type {ResourcePermissions, RoleAssignment} from './role';

/**
 * Request payload for creating a new role.
 */
export interface CreateRoleRequest {
  /** Name of the role */
  name: string;
  /** Optional description */
  description?: string;
  /** ID of the organization unit this role belongs to */
  ouId: string;
  /** Optional initial permissions */
  permissions?: ResourcePermissions[];
  /** Optional initial assignments */
  assignments?: RoleAssignment[];
}

/**
 * Request payload for updating a role.
 */
export interface UpdateRoleRequest {
  /** Name of the role */
  name: string;
  /** Optional description */
  description?: string;
  /** ID of the organization unit */
  ouId: string;
  /** Full permissions list (replaces existing) */
  permissions: ResourcePermissions[];
}

/**
 * Request payload for adding or removing role assignments.
 */
export interface AssignmentsRequest {
  /** List of assignments to add or remove */
  assignments: RoleAssignment[];
}

/**
 * Pagination parameters for role list queries.
 */
export interface RoleListParams {
  /** Maximum number of records to return */
  limit?: number;
  /** Number of records to skip */
  offset?: number;
}

/**
 * Pagination parameters for role assignment list queries.
 */
export interface RoleAssignmentListParams {
  /** Maximum number of records to return */
  limit?: number;
  /** Number of records to skip */
  offset?: number;
  /** Set to "display" to resolve display names */
  include?: 'display';
  /** Filter by assignment type */
  type?: 'user' | 'group' | 'app' | 'agent';
}
