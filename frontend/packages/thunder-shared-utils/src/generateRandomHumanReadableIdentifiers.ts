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

import humanId from 'human-id';

/**
 * Generates a list of human-readable theme name suggestions.
 *
 * @param length - The number of theme name suggestions to generate.
 * @returns An array of generated theme name suggestions.
 */
export default function generateThemeNameSuggestions(length: number = 5): string[] {
  return Array.from({length}, () => {
    const id: string = humanId({
      separator: ' ',
      capitalize: true,
      adjectiveCount: 1,
      addAdverb: false,
    });

    return id
      .split(' ')
      .map((word: string): string => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
      .join(' ');
  });
}
