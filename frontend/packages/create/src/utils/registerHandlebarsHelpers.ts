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

import Handlebars from 'handlebars';
import type {HelperOptions} from 'handlebars';

/**
 * Registers custom Handlebars helpers for use in template rendering.
 *
 * @example
 * registerHandlebarsHelpers();
 * // Enables custom helpers for templates
 *
 * @public
 */
export default function registerHandlebarsHelpers(): void {
  // Helper to convert to PascalCase
  function pascalCaseHelper(str: string): string {
    return str
      .replace(/([a-z])([A-Z])/g, '$1 $2')
      .replace(/[_-]/g, ' ')
      .split(' ')
      .filter(Boolean)
      .map((word) => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
      .join('');
  }
  Handlebars.registerHelper('pascalCase', pascalCaseHelper);

  // Helper to convert to camelCase
  function camelCaseHelper(str: string): string {
    const pascalCase = pascalCaseHelper(str);
    return pascalCase.charAt(0).toLowerCase() + pascalCase.slice(1);
  }
  Handlebars.registerHelper('camelCase', camelCaseHelper);

  // Helper to convert to kebab-case
  function kebabCaseHelper(str: string): string {
    return str
      .replace(/([a-z])([A-Z])/g, '$1-$2')
      .replace(/[_\s]+/g, '-')
      .toLowerCase();
  }
  Handlebars.registerHelper('kebabCase', kebabCaseHelper);

  // Helper to convert to CONSTANT_CASE
  function constantCaseHelper(str: string): string {
    return str
      .replace(/([a-z])([A-Z])/g, '$1_$2')
      .replace(/[_-\s]+/g, '_')
      .toUpperCase();
  }
  Handlebars.registerHelper('constantCase', constantCaseHelper);

  // Helper for conditional inclusion
  function ifEqHelper(this: unknown, a: unknown, b: unknown, options: HelperOptions) {
    return a === b ? options.fn(this) : options.inverse(this);
  }
  Handlebars.registerHelper('if_eq', ifEqHelper);

  // Helper for array inclusion
  function ifIncludesHelper(this: unknown, array: unknown, item: unknown, options: HelperOptions) {
    const arr = Array.isArray(array) ? array : [];
    const found = arr.some((el: unknown) => el === item);
    return found ? options.fn(this) : options.inverse(this);
  }
  Handlebars.registerHelper('if_includes', ifIncludesHelper);
}
