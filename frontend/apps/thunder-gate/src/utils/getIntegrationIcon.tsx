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

import type {JSX} from 'react';
import {Google, GitHub} from '@wso2/oxygen-ui-icons-react';

/**
 * TODO: Move this to a shared place: i.e thunder-shared-integrations
 * Get the identity provider icon component based on the label or image URL/path.
 *
 * Returns the appropriate icon component by analyzing either the label text or image path/URL.
 * Supports common social login providers like Google and GitHub.
 *
 * @param label - The label text that identifies the identity provider (e.g., 'Continue with Google', 'Google')
 * @param image - The image URL or path that identifies the identity provider (e.g., 'assets/images/icons/google.svg')
 * @returns The corresponding JSX icon component, or `null` if the provider cannot be identified
 *
 * @public
 * @example
 * ```tsx
 * const icon = getIntegrationIcon('Continue with Google', 'assets/images/icons/google.svg'); // Returns <Google />
 * const icon2 = getIntegrationIcon('Sign in with GitHub', 'github-icon.png'); // Returns <GitHub />
 * const unknownIcon = getIntegrationIcon('Unknown Provider', 'unknown.svg'); // Returns null
 * ```
 */
const getIntegrationIcon = (label: string, image: string): JSX.Element | null => {
  if (label.includes('google') || image.includes('google')) return <Google />;
  if (label.includes('github') || image.includes('github')) return <GitHub />;

  return null;
};

export default getIntegrationIcon;
