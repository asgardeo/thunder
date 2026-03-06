/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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
 * i18n namespace constants for Thunder i18n translations.
 *
 * @public
 * @remarks
 * These constants define the translation namespaces used for
 * Thunder i18n-related translations. Use these to reference
 * the correct i18n namespace when rendering or processing
 * Thunder i18n content.
 *
 * @example
 * // Using in a translation function
 * t(`${NamespaceConstants.CUSTOM_NAMESPACE}:form.title`)
 */
const NamespaceConstants = {
  /**
   * Namespace for custom flow translations (e.g., user-defined or dynamic flows)
   */
  CUSTOM_NAMESPACE: 'custom',
} as const;

export default NamespaceConstants;
