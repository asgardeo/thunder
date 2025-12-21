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

export const FlowTypes = {
  REGISTRATION: 'REGISTRATION',
  PASSWORD_RECOVERY: 'PASSWORD_RECOVERY',
  LOGIN: 'LOGIN',
} as const;

export type FlowTypes = (typeof FlowTypes)[keyof typeof FlowTypes];

/**
 * Interface for common metadata.
 */
export interface MetadataInterface {
  /**
   * The type of the flow.
   */
  flowType: FlowTypes;
  /**
   * Supported executors for the flow.
   */
  supportedExecutors: string[];
  /**
   * Connector configuration for the flow.
   */
  connectorConfigs: ConnectorConfigs;
  /**
   * The default attribute profile to be used.
   */
  attributeProfile: string;
  /**
   * Supported flow completion configurations.
   */
  supportedFlowCompletionConfigs?: string[];
  /**
   * Metadata for attributes used in the flow.
   */
  attributeMetadata: AttributeMetadataInterface[];
  /**
   * List of executor connections.
   */
  executorConnections: ExecutorConnectionInterface[];
}

/**
 * Common connector configuration interface.
 */
export interface ConnectorConfigs {
  /**
   * Indicates if multi-attribute login is enabled.
   */
  multiAttributeLoginEnabled: boolean;
  /**
   * Indicates if account verification is enabled.
   */
  accountVerificationEnabled: boolean;
}

/**
 * Interface for attribute metadata.
 */
export interface AttributeMetadataInterface {
  /**
   * The name of the attribute.
   */
  name: string;
  /**
   * Claim URI of the attribute.
   */
  claimURI: string;
  /**
   * Indicates if the attribute is required.
   */
  required: boolean;
  /**
   * Indicates if the attribute is read-only.
   */
  readOnly: boolean;
  /**
   * List of validators for the attribute.
   */
  validators: string[];
}

/**
 * Interface for executor connection.
 */
export interface ExecutorConnectionInterface {
  /**
   * The name of the executor.
   */
  executorName: string;
  /**
   * List of connections for the executor.
   */
  connections: string[];
}

/**
 *  Captures claim management properties.
 */
export interface Claim {
  id: string;
  claimURI: string;
  dialectURI?: string;
  description: string;
  displayOrder: number;
  multiValued: boolean;
  dataType: string;
  subAttributes?: string[];
  canonicalValues?: unknown[];
  displayName: string;
  readOnly: boolean;
  regEx: string;
  required: boolean;
  supportedByDefault: boolean;
}
