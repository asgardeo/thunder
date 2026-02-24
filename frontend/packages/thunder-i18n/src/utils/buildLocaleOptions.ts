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

import COMMON_LOCALES from './commonLocales';
import REGION_LOCALES from './regionLocales';
import toFlagEmoji from './toFlagEmoji';

export type LocaleOption = {
  /** Full BCP 47 locale code, e.g. "fr-FR". */
  code: string;
  /** English display name resolved via Intl.DisplayNames, e.g. "French (France)". */
  displayName: string;
  /** Flag emoji for the locale's region. */
  flag: string;
};

/**
 * Build a sorted list of {@link LocaleOption}.
 *
 * When `regionCode` is provided the list is scoped to locales that belong to
 * that region (from {@link REGION_LOCALES}); otherwise all
 * {@link COMMON_LOCALES} are returned.
 *
 * @param regionCode - Optional ISO 3166-1 alpha-2 region code to filter by.
 */
export default function buildLocaleOptions(regionCode?: string): LocaleOption[] {
  const dn = new Intl.DisplayNames(['en'], {type: 'language'});
  const codes = regionCode ? (REGION_LOCALES[regionCode] ?? []) : COMMON_LOCALES;
  return codes
    .map((code) => ({
      code,
      displayName: dn.of(code) ?? code,
      flag: toFlagEmoji(code.split('-')[1]?.toUpperCase() ?? ''),
    }))
    .sort((a, b) => a.displayName.localeCompare(b.displayName));
}
