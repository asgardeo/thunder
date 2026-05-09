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

import type {Element} from './elements';
import type {Step} from './steps';
import type {Template} from './templates';
import type {Widget} from './widget';

export type Resource = Element | Step | Widget;

export const ResourceTypes = {
  Step: 'STEP',
  Element: 'ELEMENT',
  Widget: 'WIDGET',
  Template: 'TEMPLATE',
} as const;

export type ResourceTypes = (typeof ResourceTypes)[keyof typeof ResourceTypes];

/**
 * Interface for the entire JSON structure.
 */
export interface Resources {
  /**
   * List of Elements.
   */
  elements: Element[];
  /**
   * List of Steps.
   */
  steps: Step[];
  /**
   * List of widgets.
   */
  widgets: Widget[];
  /**
   * List of templates.
   */
  templates: Template[];
  /**
   * List of executors (EXECUTION step resources).
   */
  executors: Step[];
}
