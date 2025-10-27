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

import {useContext} from 'react';
import ConfigContext, {ConfigContextType} from './ConfigContext';

/**
 * React hook for accessing Thunder runtime configuration throughout the application.
 *
 * This hook provides access to the configuration loaded from `window.__THUNDER_RUNTIME_CONFIG__`
 * or falls back to default values. It must be used within a component tree wrapped by
 * `ConfigProvider`, otherwise it will throw an error.
 *
 * The hook returns a context object containing the complete configuration and utility
 * methods for common operations like getting server URLs, hostnames, ports, and
 * checking HTTP-only mode.
 *
 * @returns The configuration context containing config data and utility methods
 *
 * @throws {Error} Throws an error if used outside of ConfigProvider
 *
 * @example
 * Basic usage:
 * ```tsx
 * import useConfig from './useConfig';
 *
 * function MyComponent() {
 *   const { config, getServerUrl, isHttpOnly } = useConfig();
 *
 *   return (
 *     <div>
 *       <p>Server: {getServerUrl()}</p>
 *       <p>Protocol: {isHttpOnly() ? 'HTTP' : 'HTTPS'}</p>
 *       <p>Port: {config.server.port}</p>
 *     </div>
 *   );
 * }
 * ```
 *
 * @example
 * Using with error boundary:
 * ```tsx
 * import useConfig from './useConfig';
 *
 * function ServerStatus() {
 *   try {
 *     const { getServerHostname, getServerPort } = useConfig();
 *     return <span>{getServerHostname()}:{getServerPort()}</span>;
 *   } catch (error) {
 *     return <span>Configuration not available</span>;
 *   }
 * }
 * ```
 *
 * @public
 */
export default function useConfig(): ConfigContextType {
  const context = useContext(ConfigContext);

  if (context === undefined) {
    throw new Error('useConfig must be used within a ConfigProvider');
  }

  return context;
}
