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
 * Flow type enum matching the API specification.
 */
export type FlowType = 'AUTHENTICATION' | 'REGISTRATION';

/**
 * Navigation link for pagination.
 */
export interface Link {
  /**
   * The URI of the link
   */
  href: string;
  /**
   * The link relation type (first, next, previous, last)
   */
  rel: string;
}

/**
 * Basic flow definition returned in list responses.
 */
export interface BasicFlowDefinition {
  /**
   * Unique identifier for the flow
   */
  id: string;
  /**
   * Type of flow (AUTHENTICATION or REGISTRATION)
   */
  flowType: FlowType;
  /**
   * Name of the flow
   */
  name: string;
  /**
   * URL-friendly handle for the flow (auto-generated from name)
   */
  handle: string;
  /**
   * The version number that is currently active
   */
  activeVersion: number;
  /**
   * Timestamp when the flow was initially created
   */
  createdAt: string;
  /**
   * Timestamp when the flow was last modified
   */
  updatedAt: string;
}

/**
 * Flow List Response
 *
 * Response structure for paginated flow list queries.
 * Contains pagination metadata along with the list of flows.
 */
export interface FlowListResponse {
  /**
   * Number of results that match the listing operation
   */
  totalResults: number;
  /**
   * Index of the first element of the page (offset + 1)
   */
  startIndex: number;
  /**
   * Number of elements in the returned page
   */
  count: number;
  /**
   * Array of basic flow information
   */
  flows: BasicFlowDefinition[];
  /**
   * Navigation links for pagination
   */
  links?: Link[];
}

/**
 * Node types for flow nodes.
 */
export type FlowNodeType = 'START' | 'PROMPT' | 'TASK_EXECUTION' | 'END';

/**
 * Layout information for a flow node.
 */
export interface FlowNodeLayout {
  /**
   * Size of the node
   */
  size: {
    width: number;
    height: number;
  };
  /**
   * Position of the node
   */
  position: {
    x: number;
    y: number;
  };
}

/**
 * UI metadata for PROMPT nodes.
 */
export interface FlowNodeMeta {
  /**
   * UI components to render
   */
  components?: unknown[];
}

/**
 * Input definition for flow nodes.
 */
export interface FlowNodeInput {
  /**
   * Reference to the input component ID
   */
  ref?: string;
  /**
   * Input type (TEXT_INPUT, PASSWORD_INPUT, OTP_INPUT, etc.)
   */
  type: string;
  /**
   * The mapped attribute identifier
   */
  identifier: string;
  /**
   * Whether this input is required
   */
  required: boolean;
}

/**
 * Action definition for PROMPT nodes.
 */
export interface FlowNodeAction {
  /**
   * Reference to the action component ID
   */
  ref: string;
  /**
   * ID of the next node to navigate to
   */
  nextNode: string;
  /**
   * Executor configuration for actions that trigger executors
   */
  executor?: FlowExecutor;
}

/**
 * Executor configuration for TASK_EXECUTION nodes.
 */
export interface FlowExecutor {
  /**
   * Name of the registered executor
   */
  name: string;
  /**
   * Additional executor properties
   */
  [key: string]: unknown;
}

/**
 * Flow node definition matching the API specification.
 */
export interface FlowNode {
  /**
   * Unique identifier for the node within the flow
   */
  id: string;
  /**
   * Type of node
   */
  type: FlowNodeType;
  /**
   * Layout information for the node (position and size)
   */
  layout?: FlowNodeLayout;
  /**
   * UI metadata for PROMPT nodes
   */
  meta?: FlowNodeMeta;
  /**
   * Input definitions
   */
  inputs?: FlowNodeInput[];
  /**
   * Action definitions for PROMPT nodes
   */
  actions?: FlowNodeAction[];
  /**
   * Node-level properties for configuration
   */
  properties?: Record<string, unknown>;
  /**
   * Executor configuration for TASK_EXECUTION nodes
   */
  executor?: FlowExecutor;
  /**
   * Next node ID on successful execution
   */
  onSuccess?: string;
  /**
   * Next node ID on failed execution
   */
  onFailure?: string;
}

/**
 * Request body for creating a new flow.
 */
export interface CreateFlowRequest {
  /**
   * Name of the flow
   */
  name: string;
  /**
   * URL-friendly handle for the flow (auto-generated from name)
   */
  handle: string;
  /**
   * Type of flow
   */
  flowType: FlowType;
  /**
   * List of nodes that define the flow graph
   */
  nodes: FlowNode[];
}

/**
 * Request body for updating an existing flow.
 */
export interface UpdateFlowRequest {
  /**
   * Name of the flow
   */
  name: string;
  /**
   * URL-friendly handle for the flow (auto-generated from name)
   */
  handle: string;
  /**
   * Type of flow
   */
  flowType: FlowType;
  /**
   * List of nodes that define the flow graph
   */
  nodes: FlowNode[];
}

/**
 * Full flow definition response from the API.
 */
export interface FlowDefinitionResponse {
  /**
   * Unique identifier for the flow
   */
  id: string;
  /**
   * Name of the flow
   */
  name: string;
  /**
   * URL-friendly handle for the flow (auto-generated from name)
   */
  handle: string;
  /**
   * Type of flow
   */
  flowType: FlowType;
  /**
   * The version number that is currently active
   */
  activeVersion: number;
  /**
   * List of nodes that define the flow graph
   */
  nodes: FlowNode[];
  /**
   * Timestamp when the flow was initially created
   */
  createdAt: string;
  /**
   * Timestamp when the flow was last modified
   */
  updatedAt: string;
}
