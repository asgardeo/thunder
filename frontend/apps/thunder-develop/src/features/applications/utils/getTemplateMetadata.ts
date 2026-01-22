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
import TechnologyBasedApplicationTemplateMetadata from '../config/TechnologyBasedApplicationTemplateMetadata';
import PlatformBasedApplicationTemplateMetadata from '../config/PlatformBasedApplicationTemplateMetadata';
import {normalizeTemplateId} from './template-utils';

interface TemplateMetadata {
  icon: JSX.Element;
  displayName: string;
}

/**
 * Gets the template metadata (icon and display name) for a given template ID.
 * Automatically normalizes template IDs by removing the '-embedded' suffix,
 * so 'react-embedded' will match the 'react' template metadata.
 *
 * @param templateId - The template ID (e.g., 'react', 'react-embedded', 'browser')
 * @returns Template metadata with icon and display name, or null if not found
 */
export default function getTemplateMetadata(templateId: string | undefined): TemplateMetadata | null {
  if (!templateId) {
    return null;
  }

  // Normalize the template ID by removing the embedded suffix
  const normalizedId = normalizeTemplateId(templateId);

  if (!normalizedId) {
    return null;
  }

  // Search in technology-based templates
  const techTemplate = TechnologyBasedApplicationTemplateMetadata.find(
    (metadata) => metadata.template.id === normalizedId,
  );

  if (techTemplate) {
    return {
      icon: techTemplate.icon,
      displayName: techTemplate.template.displayName ?? normalizedId,
    };
  }

  // Search in platform-based templates
  const platformTemplate = PlatformBasedApplicationTemplateMetadata.find(
    (metadata) => metadata.template.id === normalizedId,
  );

  if (platformTemplate) {
    return {
      icon: platformTemplate.icon,
      displayName: platformTemplate.template.displayName ?? templateId,
    };
  }

  return null;
}
