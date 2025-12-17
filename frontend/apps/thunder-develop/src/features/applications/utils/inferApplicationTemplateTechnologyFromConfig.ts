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

import type {OAuth2Config} from '../models/oauth';
import {OAuth2GrantTypes} from '../models/oauth';
import {TechnologyApplicationTemplate} from '../models/application-templates';

/**
 * Infers the application template technology type from an OAuth2 configuration.
 *
 * This function analyzes OAuth2 configuration properties to determine the most
 * appropriate application template technology. It uses patterns such as client
 * type (public/confidential) and grant types to make the inference.
 *
 * @param config - The OAuth2 configuration to analyze, or null if no config exists.
 * @returns The inferred technology-based application template type.
 *
 * @remarks
 * The inference logic:
 * - Public clients → REACT (typical for SPAs)
 * - Confidential clients with authorization code → NEXTJS (typical for server-side apps)
 * - No config or other patterns → OTHER (fallback)
 *
 * @example
 * ```typescript
 * // Public client (SPA)
 * const spaConfig = { public_client: true, grant_types: ['authorization_code'] };
 * inferApplicationTemplateTechnologyFromConfig(spaConfig);
 * // Returns: 'REACT'
 *
 * // Confidential client (SSR)
 * const ssrConfig = { public_client: false, grant_types: ['authorization_code'] };
 * inferApplicationTemplateTechnologyFromConfig(ssrConfig);
 * // Returns: 'NEXTJS'
 *
 * // No config
 * inferApplicationTemplateTechnologyFromConfig(null);
 * // Returns: 'OTHER'
 * ```
 */
export default function inferApplicationTemplateTechnologyFromConfig(
  config: OAuth2Config | null,
): TechnologyApplicationTemplate {
  if (!config) return TechnologyApplicationTemplate.OTHER;

  if (config.public_client) {
    return TechnologyApplicationTemplate.REACT;
  }

  if (config.grant_types.includes(OAuth2GrantTypes.AUTHORIZATION_CODE)) {
    return TechnologyApplicationTemplate.NEXTJS;
  }

  return TechnologyApplicationTemplate.OTHER;
}
