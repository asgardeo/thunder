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

import merge from 'lodash-es/merge';
import type {Element} from '../models/elements';
import type {Resources} from '../models/resources';

/**
 * Type-safe wrapper for lodash merge function.
 *
 * @param target - The target object to merge into.
 * @param sources - Source objects to merge from.
 * @returns The merged object.
 */
const safeMerge = <T>(...sources: Partial<T>[]): T => (merge as (...args: Partial<T>[]) => T)(...sources);

const resolveComponentMetadata = (resources: Resources, components?: Element[]): Element[] => {
  if (!components) {
    return [];
  }

  const updateComponentResourceType = (component: Element): Element => {
    let updatedComponent = {...component};

    resources?.elements?.forEach((componentWithMeta: Element) => {
      // Match by type only - element types are unique across categories (e.g., TEXT_INPUT only exists
      // in FIELD category, ACTION only in ACTION category). This allows proper metadata resolution
      // when components from templates have correct type but different category values.
      if (component.type === componentWithMeta.type) {
        if (component.variant) {
          // If the component metadata has a variants array, merge.
          if (componentWithMeta.variants) {
            updatedComponent = safeMerge<Element>({}, componentWithMeta, updatedComponent);

            return;
          }

          // If the component metadata has a high level variant, merge.
          if (componentWithMeta.variant && component.variant === componentWithMeta.variant) {
            updatedComponent = safeMerge<Element>({}, componentWithMeta, updatedComponent);
          }
        } else {
          updatedComponent = safeMerge<Element>({}, componentWithMeta, updatedComponent);
        }
      }
    });

    if (updatedComponent?.components) {
      updatedComponent.components = updatedComponent.components.map(updateComponentResourceType);
    }

    return updatedComponent;
  };

  // Filter out non-element resources (like widgets, templates, steps) that may have been merged in
  // Only process actual elements (resourceType === 'ELEMENT' or undefined for template components)
  // Template components won't have resourceType set yet, so we include undefined values
  return components
    ?.filter((component) => !component.resourceType || component.resourceType === 'ELEMENT')
    .map(updateComponentResourceType);
};

export default resolveComponentMetadata;
