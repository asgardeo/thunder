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

import REGION_LOCALES from './regionLocales';
import toFlagEmoji from './toFlagEmoji';

/**
 * Represents a country option for selection in language/region pickers.
 *
 * @property regionCode - ISO 3166-1 alpha-2 region code, e.g. "FR".
 * @property name - English display name resolved via Intl.DisplayNames, e.g. "France".
 * @property flag - Flag emoji derived from the region code.
 *
 * @public
 */
export interface CountryOption {
  regionCode: string;
  name: string;
  flag: string;
}

/**
 * Builds a sorted list of {@link CountryOption} objects derived from {@link REGION_LOCALES},
 * with display names resolved via {@link Intl.DisplayNames} and flag emojis.
 *
 * @returns Sorted array of country options for use in pickers and forms.
 *
 * @example
 * ```ts
 * const options = buildCountryOptions();
 * // [{ regionCode: 'FR', name: 'France', flag: '🇫🇷' }, ...]
 * ```
 *
 * @public
 */
export default function buildCountryOptions(): CountryOption[] {
  const dn = new Intl.DisplayNames(['en'], {type: 'region'});

  return Object.keys(REGION_LOCALES)
    .map((regionCode) => ({
      regionCode,
      name: dn.of(regionCode) ?? regionCode,
      flag: toFlagEmoji(regionCode),
    }))
    .sort((a, b) => a.name.localeCompare(b.name));
}
