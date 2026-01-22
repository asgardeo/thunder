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

import TemplateConstants from '../constants/template-constants';

/**
 * Normalizes a template ID by removing the embedded suffix.
 * This allows looking up template metadata for both base and embedded templates.
 *
 * @param templateId - The template ID (e.g., 'react', 'react-embedded')
 * @returns The normalized template ID without the embedded suffix (e.g., 'react')
 *
 * @example
 * ```ts
 * normalizeTemplateId('react-embedded') // Returns: 'react'
 * normalizeTemplateId('react') // Returns: 'react'
 * normalizeTemplateId('nextjs') // Returns: 'nextjs'
 * ```
 */
export function normalizeTemplateId(templateId: string | undefined): string | undefined {
  if (!templateId) {
    return templateId;
  }

  return templateId.endsWith(TemplateConstants.EMBEDDED_SUFFIX)
    ? templateId.slice(0, -TemplateConstants.EMBEDDED_SUFFIX.length)
    : templateId;
}
