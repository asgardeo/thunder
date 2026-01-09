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

import type LogLevel from './log-level';
import type Transport from './transport';

/**
 * Configuration for creating a transport instance.
 */
export interface TransportConfig {
  /**
   * The type of transport to create.
   * Built-in types: 'console', 'stdout', 'http'
   * Custom types can be registered via the transport factory.
   */
  type: string;

  /**
   * Minimum log level this transport should handle.
   * @default LogLevel.DEBUG
   */
  level?: LogLevel;

  /**
   * Transport-specific configuration options.
   */
  options?: Record<string, unknown>;
}

/**
 * Factory function type for creating custom transports.
 */
export type TransportFactory = (config: TransportConfig) => Transport;

export default TransportConfig;
