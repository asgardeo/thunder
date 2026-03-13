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

import type {IdentityProvider} from './identity-provider';

/**
 * Identity Provider Request Model
 *
 * Data structure used when creating or updating an identity provider.
 * This model is used for POST and PUT operations.
 *
 * @public
 * @example
 * ```typescript
 * const createGoogleIdp: IdentityProviderRequest = {
 *   name: 'Google',
 *   description: 'Login with Google',
 *   type: IdentityProviderTypes.GOOGLE,
 *   properties: [
 *     { name: 'client_id', value: 'your-client-id', is_secret: true },
 *     { name: 'client_secret', value: 'your-client-secret', is_secret: true },
 *     { name: 'redirect_uri', value: 'https://localhost:5091/signin', is_secret: false },
 *     { name: 'scopes', value: 'openid,email,profile', is_secret: false }
 *   ]
 * };
 * ```
 */
export type IdentityProviderRequest = Omit<IdentityProvider, 'id'>;
