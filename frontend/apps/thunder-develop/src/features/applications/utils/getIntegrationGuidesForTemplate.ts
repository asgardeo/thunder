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

import TechnologyBasedApplicationTemplateMetadata from '../config/TechnologyBasedApplicationTemplateMetadata';
import PlatformBasedApplicationTemplateMetadata from '../config/PlatformBasedApplicationTemplateMetadata';
import type {IntegrationGuides} from '../models/application-templates';
import {normalizeTemplateId} from './template-utils';

/**
 * Gets the integration guides for a given template ID
 * @param templateId - The template ID (e.g., 'react', 'react-embedded', 'nextjs', 'browser')
 * @returns Integration guides object, or null if not found
 */
export default function getIntegrationGuidesForTemplate(templateId: string | undefined): IntegrationGuides | null {
  if (!templateId) {
    return null;
  }

  // Normalize the template ID to handle embedded variants (e.g., 'react-embedded' -> 'react')
  const normalizedTemplateId = normalizeTemplateId(templateId) ?? templateId;

  // Search in technology-based templates
  const techTemplate = TechnologyBasedApplicationTemplateMetadata.find(
    (metadata) => metadata.template.id === normalizedTemplateId,
  );

  if (techTemplate?.template.integration_guides) {
    return techTemplate.template.integration_guides;
  }

  // Search in platform-based templates
  const platformTemplate = PlatformBasedApplicationTemplateMetadata.find(
    (metadata) => metadata.template.id === normalizedTemplateId,
  );

  if (platformTemplate?.template.integration_guides) {
    return platformTemplate.template.integration_guides;
  }

  return null;
}
