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

import React, {useMemo, PropsWithChildren} from 'react';
import {ThunderConfig} from './types';
import ConfigContext, {ConfigContextType} from './ConfigContext';

/* eslint-disable no-underscore-dangle */

/**
 * Props for the ConfigProvider component.
 *
 * @public
 */
export type ConfigProviderProps = PropsWithChildren;

/**
 * Loads configuration from window.__THUNDER_RUNTIME_CONFIG__ or uses default values.
 *
 * This function safely accesses the global window object and merges any runtime
 * configuration with the default configuration values. It performs a deep merge
 * to ensure all configuration properties are properly set.
 *
 * @returns The merged Thunder configuration object
 *
 * @internal
 */
function loadConfig(): ThunderConfig {
  if (typeof window !== 'undefined' && window.__THUNDER_RUNTIME_CONFIG__) {
    return window.__THUNDER_RUNTIME_CONFIG__;
  }

  throw new Error('Thunder runtime configuration is not available on window.__THUNDER_RUNTIME_CONFIG__');
}

/**
 * React context provider component that provides Thunder runtime configuration
 * to all child components.
 *
 * This component loads configuration from `window.__THUNDER_RUNTIME_CONFIG__` at
 * initialization time and provides it through React context. If the global
 * configuration is not available, it falls back to default values.
 *
 * The provider creates utility methods for common configuration operations
 * such as getting the server URL, hostname, port, and checking HTTP-only mode.
 *
 * @param props - The component props
 * @param props.children - React children to be wrapped with the configuration context
 *
 * @returns JSX element that provides configuration context to children
 *
 * @example
 * ```tsx
 * import ConfigProvider from './ConfigProvider';
 * import App from './App';
 *
 * function Root() {
 *   return (
 *     <ConfigProvider>
 *       <App />
 *     </ConfigProvider>
 *   );
 * }
 * ```
 *
 * @public
 */
export default function ConfigProvider({children}: ConfigProviderProps) {
  const config = useMemo(() => loadConfig(), []);

  const contextValue: ConfigContextType = useMemo(
    () => ({
      config,
      getServerUrl: () => {
        const {hostname, port, http_only: httpOnly} = config.server;
        const protocol: string = httpOnly ? 'http' : 'https';
        return `${protocol}://${hostname}:${port}`;
      },
      getServerHostname: () => config.server.hostname,
      getServerPort: () => config.server.port,
      isHttpOnly: () => config.server.http_only,
      getClientId: () => config.client.client_id,
      getClientUrl: () => {
        const {hostname, port, http_only: httpOnly, base} = config.client;

        // If client has its own hostname/port/protocol config, use that
        if (hostname && port !== undefined && httpOnly !== undefined) {
          const protocol: string = httpOnly ? 'http' : 'https';
          const baseUrl = `${protocol}://${hostname}:${port}`;
          return base ? `${baseUrl}${base}` : baseUrl;
        }

        // Otherwise, use window.location.origin and add base if it exists
        const origin: string = typeof window !== 'undefined' ? window.location.origin : '';
        return base ? `${origin}${base}` : origin;
      },
    }),
    [config],
  );

  return <ConfigContext.Provider value={contextValue}>{children}</ConfigContext.Provider>;
}
