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
 * Converts a string to kebab-case.
 *
 * Splits on whitespace, hyphens, underscores, and camelCase boundaries,
 * lowercases each word, and joins them with hyphens. Non-alphanumeric
 * characters (other than the separators above) are stripped.
 *
 * @example
 * kebabCase('Acrylic Orange') // => 'acrylic-orange'
 * kebabCase('fooBar')         // => 'foo-bar'
 * kebabCase('FOO_BAR')        // => 'foo-bar'
 * kebabCase('  hello world ') // => 'hello-world'
 *
 * @param value - The string to convert.
 * @returns The kebab-cased string.
 */
export default function kebabCase(value: string): string {
  return (
    value
      // Insert a separator before uppercase letters that follow a lowercase letter or digit (camelCase)
      .replace(/([a-z\d])([A-Z])/g, '$1 $2')
      // Insert a separator before a run of uppercase letters followed by a lowercase letter (e.g. XMLParser → XML-Parser)
      .replace(/([A-Z]+)([A-Z][a-z])/g, '$1 $2')
      // Replace any sequence of non-alphanumeric characters with a single space
      .replace(/[^a-zA-Z0-9]+/g, ' ')
      .trim()
      .toLowerCase()
      .split(' ')
      .filter(Boolean)
      .join('-')
  );
}
