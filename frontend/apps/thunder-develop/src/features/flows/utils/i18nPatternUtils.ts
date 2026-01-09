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
 * Regular expression pattern to match i18n format: {{t(key)}}
 */
const I18N_PATTERN = /^\{\{t\([^)]+\)\}\}$/;

/**
 * Regular expression pattern to extract the key from i18n format: {{t(key)}}
 */
const I18N_KEY_PATTERN = /^\{\{t\(([^)]+)\)\}\}$/;

/**
 * Strip HTML tags from a string using a simple regex approach.
 * Note: This uses a basic regex pattern suitable for well-formed HTML from the Lexical editor.
 * It is not intended for sanitizing arbitrary or malformed HTML content.
 * @param value - The string to strip HTML tags from.
 * @returns The string without HTML tags.
 */
function stripHtmlTags(value: string): string {
  return value.replace(/<[^>]*>/g, '').trim();
}

/**
 * Check if a value matches the i18n pattern {{t(key)}}.
 * @param value - The value to check.
 * @param stripHtml - Whether to strip HTML tags before checking. Default is false.
 * @returns True if the value matches the i18n pattern, false otherwise.
 */
export function isI18nPattern(value: string | undefined, stripHtml = false): boolean {
  if (!value) return false;

  const textContent = stripHtml ? stripHtmlTags(value) : value.trim();

  return I18N_PATTERN.test(textContent);
}

/**
 * Extract the i18n key from a pattern like {{t(key)}}.
 * @param value - The value containing the pattern.
 * @param stripHtml - Whether to strip HTML tags before extraction. Default is false.
 * @returns The extracted key or null if no match is found.
 */
export function extractI18nKey(value: string | undefined, stripHtml = false): string | null {
  if (!value) return null;

  const textContent = stripHtml ? stripHtmlTags(value) : value.trim();
  const match = I18N_KEY_PATTERN.exec(textContent);

  return match?.[1] ?? null;
}

/**
 * Resolve the i18n value by extracting and translating the key.
 * @param value - The value containing the pattern.
 * @param translateFn - The translation function.
 * @param stripHtml - Whether to strip HTML tags. Default is false.
 * @returns The translated value or an empty string if no match is found.
 */
export function resolveI18nValue(
  value: string | undefined,
  translateFn: (key: string) => string,
  stripHtml = false,
): string {
  const i18nKey = extractI18nKey(value, stripHtml);

  if (i18nKey) {
    return translateFn(i18nKey);
  }

  return '';
}
