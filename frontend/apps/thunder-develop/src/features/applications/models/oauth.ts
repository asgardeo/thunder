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

import type {TokenConfig} from './token';

/**
 * OAuth2 Grant Type
 *
 * Supported OAuth2 grant types in the Thunder platform.
 *
 * @public
 */
export type OAuth2GrantType = 'authorization_code' | 'refresh_token' | 'client_credentials' | 'password' | 'implicit';

/**
 * OAuth2 Grant Type Constants
 *
 * Constant values for OAuth2 grant types.
 * Use these constants instead of hardcoding strings.
 *
 * @public
 * @example
 * ```typescript
 * const config = {
 *   grant_types: [OAuth2GrantTypes.AUTHORIZATION_CODE, OAuth2GrantTypes.REFRESH_TOKEN]
 * };
 * ```
 */
export const OAuth2GrantTypes = {
  /** Authorization Code Flow - Most secure flow for web applications */
  AUTHORIZATION_CODE: 'authorization_code',
  /** Refresh Token - Used to obtain new access tokens */
  REFRESH_TOKEN: 'refresh_token',
  /** Client Credentials - For machine-to-machine authentication */
  CLIENT_CREDENTIALS: 'client_credentials',
  /** Resource Owner Password Credentials - Direct username/password exchange */
  PASSWORD: 'password',
  /** Implicit Flow - Deprecated, for legacy client-side apps */
  IMPLICIT: 'implicit',
} as const;

/**
 * OAuth2 Response Type
 *
 * Supported OAuth2 response types for authorization requests.
 *
 * @public
 */
export type OAuth2ResponseType = 'code' | 'token' | 'id_token' | 'code token' | 'code id_token' | 'token id_token';

/**
 * OAuth2 Response Type Constants
 *
 * Constant values for OAuth2 response types.
 *
 * @public
 */
export const OAuth2ResponseTypes = {
  /** Authorization code response */
  CODE: 'code',
  /** Access token response (implicit flow) */
  TOKEN: 'token',
  /** ID token response (implicit flow) */
  ID_TOKEN: 'id_token',
  /** Code and token response */
  CODE_TOKEN: 'code token',
  /** Code and ID token response */
  CODE_ID_TOKEN: 'code id_token',
  /** Token and ID token response */
  TOKEN_ID_TOKEN: 'token id_token',
} as const;

/**
 * Token Endpoint Authentication Method
 *
 * Methods for authenticating the client at the token endpoint.
 *
 * @public
 */
export type TokenEndpointAuthMethod =
  | 'client_secret_basic'
  | 'client_secret_post'
  | 'client_secret_jwt'
  | 'private_key_jwt'
  | 'none';

/**
 * Token Endpoint Authentication Method Constants
 *
 * @public
 */
export const TokenEndpointAuthMethods = {
  /** HTTP Basic Authentication with client credentials */
  CLIENT_SECRET_BASIC: 'client_secret_basic',
  /** Client credentials in POST body */
  CLIENT_SECRET_POST: 'client_secret_post',
  /** JWT signed with client secret */
  CLIENT_SECRET_JWT: 'client_secret_jwt',
  /** JWT signed with private key */
  PRIVATE_KEY_JWT: 'private_key_jwt',
  /** No authentication (public clients) */
  NONE: 'none',
} as const;

/**
 * Scope Claims Mapping
 *
 * Maps OAuth2 scopes to the claims (user attributes) they should include.
 * Used in ID tokens to control which user information is exposed for each scope.
 *
 * @public
 * @remarks
 * Standard OIDC scopes include:
 * - `profile`: Name, family name, given name, middle name, nickname, preferred username, picture, website, gender, birthdate, zoneinfo, locale, updated_at
 * - `email`: Email address and email_verified flag
 * - `phone`: Phone number and phone_number_verified flag
 * - `address`: Formatted address, street address, locality, region, postal code, country
 * - `group`: Group memberships
 *
 * @example
 * ```typescript
 * const scopeClaims: ScopeClaims = {
 *   profile: ['given_name', 'family_name', 'picture'],
 *   email: ['email', 'email_verified'],
 *   phone: ['phone_number'],
 *   group: ['groups']
 * };
 * ```
 */
export interface ScopeClaims {
  /**
   * Claims included when 'profile' scope is requested
   * Typically includes name, picture, and other profile information
   */
  profile?: string[];

  /**
   * Claims included when 'email' scope is requested
   * Typically includes email address and verification status
   */
  email?: string[];

  /**
   * Claims included when 'phone' scope is requested
   * Typically includes phone number and verification status
   */
  phone?: string[];

  /**
   * Claims included when 'group' scope is requested
   * Typically includes user's group memberships
   */
  group?: string[];

  /**
   * Custom scope mappings
   * Allows defining claims for custom scopes beyond the standard OIDC scopes
   */
  [key: string]: string[] | undefined;
}

/**
 * ID Token Configuration
 *
 * Configuration specific to OpenID Connect ID tokens.
 * Extends the base TokenConfig with scope-to-claims mappings.
 *
 * @public
 * @remarks
 * ID tokens are JWT tokens that contain user identity information.
 * The scope_claims mapping controls which user attributes are included
 * in the ID token based on the requested OAuth2 scopes.
 *
 * @example
 * ```typescript
 * const idTokenConfig: IDTokenConfig = {
 *   validity_period: 3600,
 *   user_attributes: ['sub', 'email', 'name'],
 *   scope_claims: {
 *     profile: ['name', 'given_name', 'family_name', 'picture'],
 *     email: ['email', 'email_verified'],
 *     phone: ['phone_number', 'phone_number_verified']
 *   }
 * };
 * ```
 */
export interface IDTokenConfig extends TokenConfig {
  /**
   * Mapping of OAuth2 scopes to their corresponding claims
   * Defines which user attributes should be included in the ID token
   * for each requested scope
   */
  scope_claims: ScopeClaims;
}

/**
 * OAuth2 Token Settings
 *
 * Complete token configuration for both access tokens and ID tokens.
 *
 * @public
 * @remarks
 * This configuration is used in the OAuth2 inbound authentication settings
 * to define how tokens should be generated and what information they contain.
 *
 * @example
 * ```typescript
 * const tokenSettings: OAuth2Token = {
 *   access_token: {
 *     validity_period: 3600,
 *     user_attributes: ['email', 'username']
 *   },
 *   id_token: {
 *     validity_period: 3600,
 *     user_attributes: ['sub', 'email', 'name'],
 *     scope_claims: {
 *       profile: ['name', 'picture'],
 *       email: ['email', 'email_verified']
 *     }
 *   }
 * };
 * ```
 */
export interface OAuth2Token {
  /**
   * Access token configuration
   * Defines the validity period and included user attributes for access tokens
   */
  access_token: TokenConfig;

  /**
   * ID token configuration
   * Defines the validity period, user attributes, and scope-to-claims mapping for ID tokens
   */
  id_token: IDTokenConfig;
}

/**
 * OAuth2 Configuration
 *
 * Complete OAuth2/OIDC configuration for an application's inbound authentication.
 * This includes client credentials, allowed OAuth2 flows, redirect URIs,
 * security settings (PKCE, public client), scopes, and token configuration.
 *
 * @public
 * @remarks
 * This configuration is used when creating or updating an application
 * to define how OAuth2/OIDC authentication should work for that application.
 *
 * Key security considerations:
 * - Use `pkce_required: true` for mobile and SPA applications
 * - Set `public_client: true` only for applications that cannot securely store credentials
 * - Validate all redirect_uris to prevent open redirect vulnerabilities
 * - Use authorization_code grant with PKCE for the most secure flow
 *
 * @example
 * ```typescript
 * // Secure web application configuration
 * const webAppConfig: OAuth2Config = {
 *   client_id: 'my-web-app',
 *   client_secret: 'super-secret-value',
 *   redirect_uris: ['https://myapp.com/callback'],
 *   grant_types: [OAuth2GrantTypes.AUTHORIZATION_CODE, OAuth2GrantTypes.REFRESH_TOKEN],
 *   response_types: [OAuth2ResponseTypes.CODE],
 *   token_endpoint_auth_method: TokenEndpointAuthMethods.CLIENT_SECRET_BASIC,
 *   pkce_required: false,
 *   public_client: false,
 *   scopes: ['openid', 'profile', 'email'],
 *   token: {
 *     access_token: {
 *       validity_period: 3600,
 *       user_attributes: ['email', 'username']
 *     },
 *     id_token: {
 *       validity_period: 3600,
 *       user_attributes: ['sub', 'email', 'name'],
 *       scope_claims: {
 *         profile: ['name', 'picture'],
 *         email: ['email', 'email_verified']
 *       }
 *     }
 *   }
 * };
 *
 * // SPA or mobile app configuration (with PKCE)
 * const spaConfig: OAuth2Config = {
 *   redirect_uris: ['http://localhost:3000/callback'],
 *   grant_types: [OAuth2GrantTypes.AUTHORIZATION_CODE, OAuth2GrantTypes.REFRESH_TOKEN],
 *   response_types: [OAuth2ResponseTypes.CODE],
 *   pkce_required: true,
 *   public_client: true,
 *   scopes: ['openid', 'profile', 'email'],
 *   token: {
 *     access_token: { validity_period: 3600, user_attributes: ['email'] },
 *     id_token: {
 *       validity_period: 3600,
 *       user_attributes: ['sub', 'email'],
 *       scope_claims: { email: ['email'] }
 *     }
 *   }
 * };
 * ```
 */
export interface OAuth2Config {
  /**
   * OAuth2 client identifier
   * Unique identifier for the application
   * Generated by the server if not provided during creation
   * @example 'my-web-app-client-id'
   */
  client_id?: string;

  /**
   * OAuth2 client secret
   * Secret credential for authenticating the client
   * Required for confidential clients, not used for public clients
   * Should be securely stored and never exposed to end users
   * @example 'super-secret-value'
   */
  client_secret?: string;

  /**
   * Allowed redirect URIs
   * List of valid URIs where the authorization server can redirect the user after authentication
   * All URIs must be pre-registered to prevent open redirect attacks
   * @example ['https://myapp.com/callback', 'https://myapp.com/oauth/callback']
   */
  redirect_uris?: string[];

  /**
   * Allowed OAuth2 grant types
   * Defines which OAuth2 flows the application can use
   * @example [OAuth2GrantTypes.AUTHORIZATION_CODE, OAuth2GrantTypes.REFRESH_TOKEN]
   */
  grant_types: string[];

  /**
   * Allowed OAuth2 response types
   * Defines what the authorization endpoint should return
   * @example [OAuth2ResponseTypes.CODE]
   */
  response_types: string[];

  /**
   * Token endpoint authentication method
   * Defines how the client authenticates at the token endpoint
   * @defaultValue 'client_secret_basic'
   * @example TokenEndpointAuthMethods.CLIENT_SECRET_BASIC
   */
  token_endpoint_auth_method?: string;

  /**
   * Whether PKCE (Proof Key for Code Exchange) is required
   * Should be true for mobile and single-page applications
   * Provides additional security against authorization code interception attacks
   * @defaultValue false
   * @see https://oauth.net/2/pkce/
   */
  pkce_required?: boolean;

  /**
   * Whether this is a public client
   * Public clients cannot securely store credentials (e.g., SPAs, mobile apps)
   * If true, client_secret should not be used
   * @defaultValue false
   */
  public_client?: boolean;

  /**
   * OAuth2/OIDC scopes
   * List of scopes the application can request
   * Standard OIDC scopes: openid, profile, email, phone, address
   * @example ['openid', 'profile', 'email']
   */
  scopes?: string[];

  /**
   * Token configuration
   * Defines how access tokens and ID tokens are generated
   */
  token?: OAuth2Token & Partial<TokenConfig>;

  /**
   * User Info configuration
   * Defines which attributes are returned in the user info response
   */
  user_info?: UserInfoConfig;
}

/**
 * User Info Configuration
 *
 * Configuration specific to the User Info endpoint response.
 * Allows defining which user attributes should be returned in the user info response.
 *
 * @public
 */
export interface UserInfoConfig {
  /**
   * List of user attributes to include in the user info response
   */
  user_attributes: string[];
}

/**
 * Default OAuth2 configuration
 *
 * Returns a default OAuth2 configuration with sensible defaults.
 * Used as a starting point for OAuth2 configuration in application creation.
 *
 * @returns Default OAuth2 configuration
 * @public
 */
export function getDefaultOAuthConfig(): OAuth2Config {
  return {
    redirect_uris: [],
    grant_types: [OAuth2GrantTypes.CLIENT_CREDENTIALS],
    response_types: [], // Client credentials doesn't use response types
    token_endpoint_auth_method: TokenEndpointAuthMethods.CLIENT_SECRET_BASIC,
    pkce_required: false,
    public_client: false,
    scopes: ['openid', 'profile', 'email'],
    token: {
      access_token: {
        validity_period: 3600,
        user_attributes: ['email', 'username'],
      },
      id_token: {
        validity_period: 3600,
        user_attributes: ['sub', 'email', 'name'],
        scope_claims: {
          profile: ['name', 'given_name', 'family_name', 'picture'],
          email: ['email', 'email_verified'],
        },
      },
    },
  };
}
