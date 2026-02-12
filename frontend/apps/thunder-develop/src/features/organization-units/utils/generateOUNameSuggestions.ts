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

import humanId from 'human-id';

/**
 * Generates a list of human-readable organization unit name suggestions.
 *
 * This function creates 10 unique, randomly generated name suggestions using the human-id library.
 * Each name is formatted with proper capitalization (title case).
 *
 * @returns An array of 10 generated organization unit name suggestions.
 *
 * @example
 * ```typescript
 * const suggestions = generateOUNameSuggestions();
 * // Returns: ['Blue Falcon', 'Red Dragon', 'Green Phoenix', ...]
 * ```
 */
export default function generateOUNameSuggestions(): string[] {
  return Array.from({length: 10}, () =>
    humanId({
      separator: ' ',
      capitalize: true,
      adjectiveCount: 0,
      addAdverb: false,
    }),
  );
}
