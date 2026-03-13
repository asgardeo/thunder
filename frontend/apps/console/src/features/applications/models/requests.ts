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

import type {Application} from './application';

/**
 * Application Request Model
 *
 * Data structure used when creating or updating an application.
 * This model is used for POST and PUT operations to the /applications endpoint.
 *
 * @public
 * @remarks
 * Applications in Thunder represent OAuth2/OIDC client applications that can
 * authenticate users and access protected resources. Each application can be
 * configured with:
 * - Basic metadata (name, description, logo, URLs)
 * - Authentication and registration flows
 * - OAuth2/OIDC inbound authentication settings
 * - User attributes to include in tokens
 *
 * The server will generate additional fields (id, client_id, timestamps) upon creation.
 *
 * @example
 * ```typescript
 * // Create a basic web application with OAuth2 authentication
 * const createWebApp: CreateApplicationRequest = {
 *   name: 'My Web Application',
 *   description: 'Customer portal application',
 *   url: 'https://myapp.com',
 *   logo_url: 'https://myapp.com/logo.png',
 *   tos_uri: 'https://myapp.com/terms',
 *   policy_uri: 'https://myapp.com/privacy',
 *   contacts: ['admin@myapp.com', 'support@myapp.com'],
 *   auth_flow_id: 'edc013d0-e893-4dc0-990c-3e1d203e005b',
 *   registration_flow_id: '80024fb3-29ed-4c33-aa48-8aee5e96d522',
 *   is_registration_flow_enabled: true,
 *   user_attributes: ['email', 'username', 'roles'],
 *   inbound_auth_config: [{
 *     type: 'oauth2',
 *     config: {
 *       redirect_uris: ['https://myapp.com/callback'],
 *       grant_types: ['authorization_code', 'refresh_token'],
 *       response_types: ['code'],
 *       scopes: ['openid', 'profile', 'email'],
 *       token: {
 *         access_token: {
 *           validity_period: 3600,
 *           user_attributes: ['email', 'username']
 *         },
 *         id_token: {
 *           validity_period: 3600,
 *           user_attributes: ['sub', 'email', 'name'],
 *           scope_claims: {
 *             profile: ['name', 'picture'],
 *             email: ['email', 'email_verified']
 *           }
 *         }
 *       }
 *     }
 *   }]
 * };
 * ```
 *
 * @example
 * ```typescript
 * // Create a minimal SPA application
 * const createSPA: CreateApplicationRequest = {
 *   name: 'My SPA',
 *   url: 'http://localhost:3000',
 *   inbound_auth_config: [{
 *     type: 'oauth2',
 *     config: {
 *       redirect_uris: ['http://localhost:3000/callback'],
 *       grant_types: ['authorization_code', 'refresh_token'],
 *       response_types: ['code'],
 *       pkce_required: true,
 *       public_client: true,
 *       scopes: ['openid', 'profile'],
 *       token: {
 *         access_token: { validity_period: 3600, user_attributes: [] },
 *         id_token: { validity_period: 3600, user_attributes: ['sub'], scope_claims: {} }
 *       }
 *     }
 *   }]
 * };
 * ```
 */
export type CreateApplicationRequest = Omit<Application, 'id' | 'created_at' | 'updated_at'>;
