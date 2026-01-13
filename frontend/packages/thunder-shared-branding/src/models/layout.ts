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

/**
 * Enumeration of supported layout types for UI components.
 * Used to specify how content should be positioned and arranged.
 */
export const LayoutType = {
  /** Centered layout with content in the middle of the viewport */
  CENTERED: 'CENTERED',

  /** Left-aligned layout with content positioned to the left */
  LEFT_ALIGNED: 'LEFT_ALIGNED',

  /** Right-aligned layout with content positioned to the right */
  RIGHT_ALIGNED: 'RIGHT_ALIGNED',
} as const;

/**
 * Union type representing the possible layout types.
 * @example 'CENTERED' | 'LEFT_ALIGNED' | 'RIGHT_ALIGNED' | 'FULL_WIDTH' | 'SPLIT'
 */
export type LayoutType = (typeof LayoutType)[keyof typeof LayoutType];

/**
 * Layout configuration for UI components
 */
export interface LayoutConfig {
  /**
   * The type of layout to apply
   * @example "CENTERED"
   */
  type: LayoutType;
}
