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
 * Regular expression to match the meta pattern `{{meta(key)}}` (exact, full-string match).
 */
export const META_PATTERN = /^\{\{meta\([^)]+\)\}\}$/;

/**
 * Regular expression to extract the key from a meta pattern `{{meta(key)}}`.
 */
export const META_KEY_PATTERN = /^\{\{meta\(([^)]+)\)\}\}$/;

/**
 * Check if a value matches the meta template pattern `{{meta(key)}}`.
 *
 * @param value - The string to test.
 * @returns `true` if the trimmed value matches the pattern, `false` otherwise.
 *
 * @example
 * ```typescript
 * isMetaTemplatePattern('{{meta(user:name)}}') // true
 * isMetaTemplatePattern('hello world')         // false
 * ```
 */
export default function isMetaTemplatePattern(value: string): boolean {
  return META_PATTERN.test(value.trim());
}
