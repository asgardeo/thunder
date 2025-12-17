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

import type {OAuth2Config} from './oauth';

/**
 * Inbound Authentication Protocol Type
 *
 * Supported inbound authentication protocols in the Thunder platform.
 * Currently supports OAuth2/OIDC for application authentication.
 *
 * @public
 */
export type InboundAuthType = 'oauth2';

/**
 * Inbound Authentication Type Constants
 *
 * Constant values for inbound authentication protocol types.
 * Use these constants instead of hardcoding strings.
 *
 * @public
 * @example
 * ```typescript
 * const authConfig = {
 *   type: InboundAuthTypes.OAUTH2
 * };
 * ```
 */
export const InboundAuthTypes = {
  /** OAuth 2.0 / OpenID Connect authentication */
  OAUTH2: 'oauth2',
} as const;

/**
 * Inbound Authentication Configuration
 *
 * Defines the inbound authentication protocol and its configuration for an application.
 * Inbound authentication controls how external clients authenticate to access the application's resources.
 *
 * @public
 * @remarks
 * Currently, Thunder supports OAuth2/OIDC as the primary inbound authentication protocol.
 * This configuration is used when creating or updating an application to define:
 * - The authentication protocol type
 * - Protocol-specific configuration (OAuth2 settings)
 *
 * In the future, additional authentication protocols may be supported (e.g., SAML, WS-Federation).
 *
 * @example
 * ```typescript
 * // OAuth2 inbound authentication for a web application
 * const inboundAuth: InboundAuthConfig = {
 *   type: InboundAuthTypes.OAUTH2,
 *   config: {
 *     client_id: 'my-web-app',
 *     client_secret: 'super-secret',
 *     redirect_uris: ['https://myapp.com/callback'],
 *     grant_types: ['authorization_code', 'refresh_token'],
 *     response_types: ['code'],
 *     scopes: ['openid', 'profile', 'email'],
 *     token: {
 *       access_token: {
 *         validity_period: 3600,
 *         user_attributes: ['email', 'username']
 *       },
 *       id_token: {
 *         validity_period: 3600,
 *         user_attributes: ['sub', 'email', 'name'],
 *         scope_claims: {
 *           profile: ['name', 'picture'],
 *           email: ['email', 'email_verified']
 *         }
 *       }
 *     }
 *   }
 * };
 * ```
 *
 * @example
 * ```typescript
 * // OAuth2 inbound authentication for a SPA with PKCE
 * const spaInboundAuth: InboundAuthConfig = {
 *   type: InboundAuthTypes.OAUTH2,
 *   config: {
 *     redirect_uris: ['http://localhost:3000/callback'],
 *     grant_types: ['authorization_code', 'refresh_token'],
 *     response_types: ['code'],
 *     pkce_required: true,
 *     public_client: true,
 *     scopes: ['openid', 'profile', 'email'],
 *     token: {
 *       access_token: {
 *         validity_period: 3600,
 *         user_attributes: ['email']
 *       },
 *       id_token: {
 *         validity_period: 3600,
 *         user_attributes: ['sub', 'email'],
 *         scope_claims: {
 *           email: ['email', 'email_verified']
 *         }
 *       }
 *     }
 *   }
 * };
 * ```
 */
export interface InboundAuthConfig {
  /**
   * The authentication protocol type
   * Currently only 'oauth2' is supported
   * @example InboundAuthTypes.OAUTH2
   */
  type: string;

  /**
   * Protocol-specific configuration
   * For OAuth2/OIDC, this contains client credentials, allowed flows,
   * redirect URIs, scopes, and token settings
   * @see OAuth2Config
   */
  config: OAuth2Config;
}
