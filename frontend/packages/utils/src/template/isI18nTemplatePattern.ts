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
 * Regular expression to match the i18n pattern `{{t(key)}}` (exact, full-string match).
 * Allows optional whitespace around `t(key)` to handle both `{{t(key)}}` and `{{ t(key) }}`.
 */
export const I18N_PATTERN = /^\{\{\s*t\([^)]+\)\s*\}\}$/;

/**
 * Regular expression to extract the key from an i18n pattern `{{t(key)}}`.
 * Allows optional whitespace around `t(key)`.
 */
export const I18N_KEY_PATTERN = /^\{\{\s*t\(([^)]+)\)\s*\}\}$/;

/**
 * Check if a value matches the i18n template pattern `{{t(key)}}`.
 *
 * @param value - The string to test.
 * @returns `true` if the trimmed value matches the pattern, `false` otherwise.
 *
 * @example
 * ```typescript
 * isI18nTemplatePattern('{{t(signin:heading)}}') // true
 * isI18nTemplatePattern('hello world')           // false
 * ```
 */
export default function isI18nTemplatePattern(value: string): boolean {
  return I18N_PATTERN.test(value.trim());
}
