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

import type {EmojiCategory} from '../EmojiPicker';
import EMOJI_DATA from '../emojis.json';

export const EMOJI_CATEGORIES: EmojiCategory[] = EMOJI_DATA as unknown as EmojiCategory[];

const ALL_EMOJIS: string[] = EMOJI_CATEGORIES.flatMap((c) => c.emojis.map((e) => e.char));

/**
 * Generates a specified number of random emoji icon suggestions.
 *
 * @param count - The number of random emoji icons to return.
 * @returns An array of emoji character strings.
 *
 * @example
 * ```typescript
 * const icons = generateIconSuggestions(8);
 * // Returns: ['🐼', '🚀', '💎', ...]
 * ```
 */
export default function generateIconSuggestions(count: number): string[] {
  const shuffled: string[] = [...ALL_EMOJIS].sort((): number => Math.random() - 0.5);

  return shuffled.slice(0, count);
}
