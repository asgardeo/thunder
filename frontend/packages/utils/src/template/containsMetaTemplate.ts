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
 * Build a regex that matches `{{ meta(key) }}` (with optional whitespace) anywhere
 * within a string, escaping any special regex characters in `key`.
 */
function buildMetaTemplateRegex(key: string): RegExp {
  const escapedKey = key.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');

  return new RegExp(`\\{\\{\\s*meta\\(${escapedKey}\\)\\s*\\}\\}`);
}

/**
 * Check whether a string contains a `{{ meta(key) }}` template anywhere within it.
 *
 * Unlike {@link isMetaTemplatePattern}, which requires the whole string to be the
 * template, this function detects the pattern embedded inside a larger string such
 * as an HTML label.
 *
 * Whitespace around `{{` / `}}` is allowed, e.g. `{{ meta(application.sign_up_url) }}`.
 *
 * @param str - The string to search (may be a plain value or an HTML fragment).
 * @param key - The meta key to look for, e.g. `"application.sign_up_url"`.
 * @returns `true` if the pattern is found anywhere in `str`, `false` otherwise.
 *
 * @example
 * ```typescript
 * containsMetaTemplate('<a href="{{meta(application.sign_up_url)}}">Sign up</a>', 'application.sign_up_url')
 * // true
 *
 * containsMetaTemplate('<a href="https://example.com">Sign up</a>', 'application.sign_up_url')
 * // false
 * ```
 */
export default function containsMetaTemplate(str: string, key: string): boolean {
  return buildMetaTemplateRegex(key).test(str);
}

/**
 * Replace all occurrences of `{{ meta(key) }}` (with optional whitespace) in `str`
 * with `replacement`.
 *
 * @param str - The source string.
 * @param key - The meta key to replace, e.g. `"application.sign_up_url"`.
 * @param replacement - The value to substitute for each match.
 * @returns A new string with all occurrences replaced.
 */
export function replaceMetaTemplate(str: string, key: string, replacement: string): string {
  const escapedKey = key.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  const regex = new RegExp(`\\{\\{\\s*meta\\(${escapedKey}\\)\\s*\\}\\}`, 'g');

  return str.replace(regex, replacement);
}
