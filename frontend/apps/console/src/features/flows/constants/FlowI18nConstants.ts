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

import {NamespaceConstants} from '@thunder/i18n';

/**
 * i18n namespace constants for flow feature translations.
 *
 * @public
 * @remarks
 * These constants define the translation namespaces used for
 * flow-related UI and runtime translations. Use these to reference
 * the correct i18n namespace when rendering or processing flow content.
 *
 * @example
 * // Using in a translation function
 * t(`${FlowI18nConstants.CUSTOM_TRANSLATIONS_NAMESPACE}:form.title`)
 */
const FlowI18nConstants = {
  /**
   * Namespace for custom flow translations (e.g., user-defined or dynamic flows)
   */
  CUSTOM_TRANSLATIONS_NAMESPACE: NamespaceConstants.CUSTOM_NAMESPACE,
  /**
   * Namespaces for built-in flow translations (e.g., sign-in, sign-up, onboarding flows)
   */
  FLOW_TRANSLATIONS_NAMESPACES: [],
  /**
   * Namespace for default flow translations (e.g., built-in or system flows)
   */
  DEFAULT_LANGUAGE: 'en-US',
} as const;

export default FlowI18nConstants;
