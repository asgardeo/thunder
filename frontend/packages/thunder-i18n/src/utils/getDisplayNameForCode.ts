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
 * Resolve a human-readable English display name for a BCP 47 locale code using
 * {@link Intl.DisplayNames}. Returns `null` when the code is empty, invalid,
 * or when the resolved name equals the raw code (i.e. Intl has no data for it).
 *
 * Uses {@link Intl.getCanonicalLocales} internally to validate the tag before
 * attempting resolution.
 *
 * @example getDisplayNameForCode('fr-FR') // 'French (France)'
 * @example getDisplayNameForCode('xyz')   // null
 */
export default function getDisplayNameForCode(code: string): string | null {
  if (!code.trim()) return null;
  try {
    Intl.getCanonicalLocales(code);
    const dn = new Intl.DisplayNames(['en'], {type: 'language'});
    const name = dn.of(code);
    return name && name !== code ? name : null;
  } catch {
    return null;
  }
}
