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

import type {Base} from './base';
import type {Step} from './steps';

/**
 * Template placeholder replacer.
 */
export interface TemplateReplacer {
  placeholder: string;
  value: string;
  [key: string]: unknown;
}

/**
 * Generation metadata for template placeholders.
 */
export interface TemplateGenerationMeta {
  /**
   * Replacers for template placeholders.
   */
  replacers?: TemplateReplacer[];
  [key: string]: unknown;
}

/**
 * Template-specific configuration data.
 */
export interface TemplateConfigData {
  /**
   * Steps contained in the template.
   */
  steps: Step[];
  /**
   * Generation metadata for template placeholders.
   */
  __generationMeta__?: TemplateGenerationMeta;
}

/**
 * Template-specific configuration that extends the base config.
 */
export interface TemplateConfig {
  /**
   * Template data containing steps.
   */
  data: TemplateConfigData;
}

export type Template = Base<TemplateConfig>;

export const TemplateCategories = {
  Starter: 'STARTER',
} as const;

export type TemplateCategories = (typeof TemplateCategories)[keyof typeof TemplateCategories];

export const TemplateTypes = {
  Blank: 'BLANK',
  Basic: 'BASIC',
  BasicFederated: 'BASIC_FEDERATED',
  GeneratedWithAI: 'GENERATE_WITH_AI',
  PasskeyLogin: 'PASSKEY_LOGIN',
  Default: 'DEFAULT',
} as const;

export type TemplateTypes = (typeof TemplateTypes)[keyof typeof TemplateTypes];
