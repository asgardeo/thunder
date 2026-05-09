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

/**
 * Convert a BCP 47 locale code or ISO 3166-1 alpha-2 region code to a flag
 * emoji using regional indicator symbol letters (Unicode range 0x1F1E6–0x1F1FF).
 *
 * When given a full locale code the region part is extracted first.
 * Language-only codes with no region part return 🌐.
 *
 * @example toFlagEmoji('de-CH') // '🇨🇭'
 * @example toFlagEmoji('en-US') // '🇺🇸'
 * @example toFlagEmoji('en')    // '🌐'
 */
export default function toFlagEmoji(localeOrRegionCode: string): string {
  const parts = localeOrRegionCode.split('-');
  const lastPart = parts.at(-1)!;
  const regionCode = /^[A-Z]{2}$/.test(lastPart) ? lastPart : null;

  if (!regionCode) {
    return '🌐';
  }

  return [...regionCode.toUpperCase()].map((c) => String.fromCodePoint(0x1f1e6 + c.charCodeAt(0) - 65)).join('');
}
