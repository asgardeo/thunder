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

import normalizeTemplateId from './normalizeTemplateId';
import PlatformBasedApplicationTemplateMetadata from '../config/PlatformBasedApplicationTemplateMetadata';
import TechnologyBasedApplicationTemplateMetadata from '../config/TechnologyBasedApplicationTemplateMetadata';
import type {ApplicationTemplate} from '../models/application-templates';

/**
 * Gets the field constraints for a given template ID.
 * Automatically normalizes template IDs by removing the '-embedded' suffix,
 * so 'react-embedded' will match the 'react' template constraints.
 *
 * @param templateId - The template ID (e.g., 'react', 'react-embedded', 'browser')
 * @returns Template field constraints, or null if not found
 */
export default function getTemplateFieldConstraints(
  templateId: string | undefined,
): ApplicationTemplate['fieldConstraints'] | null {
  if (!templateId) return null;

  const normalizedId = normalizeTemplateId(templateId);
  if (!normalizedId) return null;

  const techTemplate = TechnologyBasedApplicationTemplateMetadata.find(
    (metadata) => metadata.template.id === normalizedId,
  );
  if (techTemplate) return techTemplate.template.fieldConstraints ?? null;

  const platformTemplate = PlatformBasedApplicationTemplateMetadata.find(
    (metadata) => metadata.template.id === normalizedId,
  );
  if (platformTemplate) return platformTemplate.template.fieldConstraints ?? null;

  return null;
}
