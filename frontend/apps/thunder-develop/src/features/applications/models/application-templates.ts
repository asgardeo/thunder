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
import type {Application} from './application';

/**
 * Technology-based application template identifiers.
 * Used for framework-specific application configurations (React, Next.js, etc.).
 *
 * @public
 */
export const TechnologyApplicationTemplate = {
  REACT: 'REACT',
  NEXTJS: 'NEXTJS',
  OTHER: 'OTHER',
} as const;

/**
 * Platform-based application template identifiers.
 * Used for platform-specific application configurations (Browser, Mobile, etc.).
 *
 * @public
 */
export const PlatformApplicationTemplate = {
  BACKEND: 'BACKEND',
  BROWSER: 'BROWSER',
  MOBILE: 'MOBILE',
  SERVER: 'SERVER',
} as const;

export type ApplicationTemplate = Pick<
  Application,
  'name' | 'description' | 'inbound_auth_config' | 'allowed_user_types'
>;

export interface ApplicationTemplateMetadata<T = TechnologyApplicationTemplate | PlatformApplicationTemplate> {
  value: T;
  icon: JSX.Element;
  titleKey: string;
  descriptionKey: string;
  template: ApplicationTemplate;
}

export type TechnologyApplicationTemplate = keyof typeof TechnologyApplicationTemplate;

export type PlatformApplicationTemplate = keyof typeof PlatformApplicationTemplate;
