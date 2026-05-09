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
 * Drop-in replacement for lodash `isEmpty`.
 *
 * Returns `true` for:
 * - `null` / `undefined`
 * - strings, arrays, and `arguments` objects with `length === 0`
 * - `Map` / `Set` with `size === 0`
 * - plain objects with no own enumerable keys
 * - numbers, booleans, and `Symbol` values (lodash considers primitives empty)
 *
 * @param value - The value to check.
 * @returns `true` if `value` is empty, `false` otherwise.
 */
export default function isEmpty(value: unknown): boolean {
  if (value == null) return true;

  if (typeof value === 'boolean' || typeof value === 'number' || typeof value === 'symbol') {
    return true;
  }

  if (typeof value === 'string' || Array.isArray(value)) {
    return value.length === 0;
  }

  if (value instanceof Map || value instanceof Set) {
    return value.size === 0;
  }

  if (typeof value === 'object') {
    // arguments objects and array-like objects
    if ('length' in value && typeof (value as {length: unknown}).length === 'number') {
      return (value as {length: number}).length === 0;
    }
    return Object.keys(value).length === 0;
  }

  return true;
}
