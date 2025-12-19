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
 * Authentication flow handles that map to pre-configured authentication flows
 * in the identity management platform.
 *
 * These handles correspond to flow definitions in the backend database that
 * orchestrate the authentication process based on the selected identity providers
 * and authentication mechanisms. The actual flow UUIDs are resolved at runtime
 * by querying the flows API using these handles.
 *
 * @remarks
 * Each property represents a unique authentication flow handle:
 * - `BASIC` - Standard username and password authentication flow
 * - `GOOGLE` - Google social login authentication flow
 * - `GITHUB` - GitHub social login authentication flow
 * - `BASIC_GOOGLE` - Combined username/password and Google authentication with user choice
 * - `BASIC_GITHUB` - Combined username/password and GitHub authentication with user choice
 * - `GOOGLE_GITHUB` - Combined Google and GitHub social authentication with user choice
 * - `BASIC_GOOGLE_GITHUB` - Combined username/password, Google, and GitHub authentication with user choice
 * - `BASIC_GOOGLE_GITHUB_SMS` - Multi-factor authentication combining basic, social, and SMS verification
 * - `BASIC_WITH_PROMPT` - Username/password authentication with additional user prompts or verification steps
 * - `SMS` - SMS-based passwordless authentication using one-time codes
 *
 * @example
 * Resolve flow handle to UUID:
 * ```tsx
 * import { AUTH_FLOW_HANDLES } from './auth-flow-graphs';
 * import useAuthenticationFlows from '../api/useAuthenticationFlows';
 *
 * function MyComponent() {
 *   const { getFlowIdByHandle } = useAuthenticationFlows();
 *   const flowUUID = getFlowIdByHandle(AUTH_FLOW_HANDLES.BASIC);
 *   // flowUUID is now the actual UUID from the database
 * }
 * ```
 */
export const AUTH_FLOW_HANDLES = {
  BASIC: 'default-basic-flow',
  GOOGLE: 'default-google-flow',
  GITHUB: 'default-github-flow',
  BASIC_GOOGLE: 'basic-google-flow',
  BASIC_GITHUB: 'basic-github-flow',
  GOOGLE_GITHUB: 'google-github-flow',
  BASIC_GOOGLE_GITHUB: 'basic-google-github-flow',
  BASIC_GOOGLE_GITHUB_SMS: 'basic-google-github-sms-flow',
  BASIC_WITH_PROMPT: 'basic-with-prompt-flow',
  SMS: 'default-sms-otp-flow',
} as const;

/**
 * Registration flow handles that map to pre-configured user registration flows
 * in the Thunder identity management platform.
 *
 * These handles correspond to flow definitions in the backend database that
 * orchestrate the user onboarding process based on the selected registration methods
 * and identity providers.
 *
 * @remarks
 * Each property represents a unique registration flow handle:
 * - `BASIC` - Standard email/username and password registration flow
 * - `BASIC_GOOGLE_GITHUB` - Registration with email/password or social login (Google/GitHub) options
 * - `BASIC_GOOGLE_GITHUB_SMS` - Registration with email/password, social login, and SMS verification
 * - `SMS` - SMS-based registration using mobile number verification
 *
 * @example
 * Resolve flow handle to UUID:
 * ```tsx
 * import { REGISTRATION_FLOW_HANDLES } from './auth-flow-graphs';
 *
 * const flowHandle = REGISTRATION_FLOW_HANDLES.BASIC;
 * ```
 */
export const REGISTRATION_FLOW_HANDLES = {
  BASIC: 'default-basic-flow',
  BASIC_GOOGLE_GITHUB: 'basic-google-github-flow',
  BASIC_GOOGLE_GITHUB_SMS: 'basic-google-github-sms-flow',
  SMS: 'default-sms-otp-flow',
} as const;

/**
 * System flow handles that should be excluded from user selection.
 *
 * These are internal flows used for system purposes and should not
 * be shown in user-facing dropdowns or selection lists.
 *
 * @remarks
 * - `DEVELOP_APP` - Internal flow used for the develop app registration process
 */
export const SYSTEM_FLOW_HANDLES = {
  DEVELOP_APP: 'develop-app-flow',
} as const;
