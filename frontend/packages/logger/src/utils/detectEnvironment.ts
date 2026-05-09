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

/* eslint-disable no-console */

import type RuntimeEnvironment from '../models/runtime-environment';

interface ImportMetaEnv {
  DEV?: boolean;
  PROD?: boolean;
  [key: string]: unknown;
}

interface ImportMeta {
  env?: ImportMetaEnv;
}

/**
 * Detect the current runtime environment.
 * @returns Runtime environment information
 */
export default function detectEnvironment(): RuntimeEnvironment {
  const isBrowser = typeof window !== 'undefined' && typeof window.document !== 'undefined';
  const isNode = typeof process !== 'undefined' && process.versions?.node != null;

  // Detect development mode
  let isDevelopment = false;
  let isProduction = false;

  try {
    // Try Vite environment detection
    if (typeof import.meta !== 'undefined' && (import.meta as ImportMeta).env !== undefined) {
      isDevelopment = (import.meta as ImportMeta).env?.DEV === true;
      isProduction = (import.meta as ImportMeta).env?.PROD === true;
    }
  } catch {
    // import.meta may not be available in all contexts
  }

  // Fallback to Node.js environment detection
  if (!isDevelopment && !isProduction && isNode) {
    isDevelopment = process.env['NODE_ENV'] === 'development';
    isProduction = process.env['NODE_ENV'] === 'production';
  }

  return {
    isBrowser,
    isNode,
    isDevelopment,
    isProduction,
  };
}

/**
 * Safely check if console is available.
 * @returns True if console is available
 */
export function hasConsole(): boolean {
  return typeof console !== 'undefined' && typeof console.log === 'function';
}

/**
 * Safely check if process is available.
 * @returns True if process is available
 */
export function hasProcess(): boolean {
  return typeof process !== 'undefined' && typeof process.stdout !== 'undefined';
}
