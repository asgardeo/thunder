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

import {
  IdentityProviderTypes,
  type IdentityProvider,
  type IdentityProviderType,
} from '@/features/integrations/models/identity-provider';
import {AUTH_FLOW_HANDLES} from '../models/auth-flow-graphs';

/**
 * Options for resolving authentication flow graph configuration.
 */
interface ResolveAuthFlowOptions {
  /**
   * Whether username/password authentication is enabled
   */
  hasUsernamePassword: boolean;

  /**
   * Array of selected identity providers (e.g., Google, GitHub)
   */
  identityProviders: IdentityProvider[];
}

/**
 * Resolves the appropriate authentication flow handle based on selected sign-in options.
 *
 * The resolver follows these rules:
 * 1. If only username/password is selected -> default-basic-flow
 * 2. If only Google is selected -> default-google-flow
 * 3. If only GitHub is selected -> default-github-flow
 * 4. If username/password + Google -> basic-google-flow
 * 5. If username/password + GitHub -> basic-github-flow
 * 6. If Google + GitHub (no basic) -> google-github-flow
 * 7. If username/password + Google + GitHub -> basic-google-github-flow
 *
 * @param options - Configuration object with sign-in options
 * @param options.hasUsernamePassword - Whether username/password authentication is enabled
 * @param options.identityProviders - Array of selected identity providers
 * @returns The authentication flow handle that matches the selected options
 *
 * @example
 * ```tsx
 * // Username & Password only
 * const flowHandle = resolveAuthFlowHandle({
 *   hasUsernamePassword: true,
 *   identityProviders: []
 * });
 * // Returns: 'default-basic-flow'
 *
 * // Username & Password + Google
 * const flowHandle = resolveAuthFlowHandle({
 *   hasUsernamePassword: true,
 *   identityProviders: [{ id: '123', name: 'Google', type: 'GOOGLE' }]
 * });
 * // Returns: 'basic-google-flow'
 *
 * // Username & Password + Google + GitHub
 * const flowHandle = resolveAuthFlowHandle({
 *   hasUsernamePassword: true,
 *   identityProviders: [
 *     { id: '123', name: 'Google', type: 'GOOGLE' },
 *     { id: '456', name: 'GitHub', type: 'GITHUB' }
 *   ]
 * });
 * // Returns: 'basic-google-github-flow'
 * ```
 */
export default function resolveAuthFlowHandle({
  hasUsernamePassword,
  identityProviders,
}: ResolveAuthFlowOptions): string {
  const providerTypes: IdentityProviderType[] = identityProviders.map((idp: IdentityProvider) => idp.type);
  const hasGoogle: boolean = providerTypes.includes(IdentityProviderTypes.GOOGLE);
  const hasGitHub: boolean = providerTypes.includes(IdentityProviderTypes.GITHUB);

  // Only username/password
  if (hasUsernamePassword && !hasGoogle && !hasGitHub) {
    return AUTH_FLOW_HANDLES.BASIC;
  }

  // Only Google
  if (!hasUsernamePassword && hasGoogle && !hasGitHub) {
    return AUTH_FLOW_HANDLES.GOOGLE;
  }

  // Only GitHub
  if (!hasUsernamePassword && !hasGoogle && hasGitHub) {
    return AUTH_FLOW_HANDLES.GITHUB;
  }

  // Username/Password + Google
  if (hasUsernamePassword && hasGoogle && !hasGitHub) {
    return AUTH_FLOW_HANDLES.BASIC_GOOGLE;
  }

  // Username/Password + Google + GitHub
  if (hasUsernamePassword && hasGoogle && hasGitHub) {
    return AUTH_FLOW_HANDLES.BASIC_GOOGLE_GITHUB;
  }

  // Username/Password + GitHub
  if (hasUsernamePassword && !hasGoogle && hasGitHub) {
    return AUTH_FLOW_HANDLES.BASIC_GITHUB;
  }

  // Only Google + GitHub (no username/password)
  if (!hasUsernamePassword && hasGoogle && hasGitHub) {
    return AUTH_FLOW_HANDLES.GOOGLE_GITHUB;
  }

  // Default fallback to basic
  return AUTH_FLOW_HANDLES.BASIC;
}
