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
 * Server configuration interface that defines connection parameters
 * for the Thunder backend server.
 *
 * @public
 */
export interface ServerConfig {
  /**
   * Server hostname or IP address
   * @example "localhost", "api.example.com", "192.168.1.100"
   */
  hostname: string;

  /**
   * Server port number
   * @example 8090, 3000, 8080
   */
  port: number;

  /**
   * Whether to use HTTP only (no HTTPS). When true, connections will use HTTP protocol.
   * When false, HTTPS will be used for secure connections.
   */
  http_only: boolean;
}

/**
 * Client configuration interface that defines authentication and client-specific settings.
 *
 * @public
 */
export interface ClientConfig {
  /**
   * Base path for the client application.
   * @example "/develop", "/admin", "/my-app"
   */
  base: string;

  /**
   * Unique identifier for the client application, used for authentication
   * and authorization with identity providers like Asgardeo.
   * @example "DEVELOP", "thunder-admin", "my-app-client-id"
   */
  client_id: string;

  /**
   * Server hostname or IP address
   * @example "localhost", "api.example.com", "192.168.1.100"
   */
  hostname?: string;

  /**
   * Server port number
   * @example 8090, 3000, 8080
   */
  port?: number;

  /**
   * Whether to use HTTP only (no HTTPS). When true, connections will use HTTP protocol.
   * When false, HTTPS will be used for secure connections.
   */
  http_only?: boolean;
}

/**
 * Thunder runtime configuration interface that contains all configuration
 * settings for Thunder applications.
 *
 * This interface defines the complete structure of the runtime configuration
 * that can be loaded from `window.__THUNDER_RUNTIME_CONFIG__` or provided
 * as default values.
 *
 * @public
 */
export interface ThunderConfig {
  /** Client-specific configuration including authentication settings */
  client: ClientConfig;

  /** Server connection configuration */
  server: ServerConfig;
}

/**
 * Global window interface extension for Thunder runtime configuration.
 *
 * This declaration extends the global Window interface to include the
 * Thunder runtime configuration object. The configuration is typically
 * loaded from a config.js file in the public directory and made available
 * globally on the window object.
 *
 * @example
 * ```javascript
 * // In public/config.js
 * window.__THUNDER_RUNTIME_CONFIG__ = {
 *   client: {
 *     client_id: 'DEVELOP'
 *   },
 *   server: {
 *     hostname: 'localhost',
 *     port: 8090,
 *     http_only: false
 *   }
 * };
 * ```
 *
 * @public
 */
declare global {
  interface Window {
    /** Thunder runtime configuration loaded from config.js */
    __THUNDER_RUNTIME_CONFIG__?: ThunderConfig;
  }
}
