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
 * Known camelCase React prop names that map to different HTML attribute names.
 */
const ATTR_MAP: Record<string, string> = {
  charSet: 'charset',
  crossOrigin: 'crossorigin',
  httpEquiv: 'http-equiv',
  noModule: 'nomodule',
  referrerPolicy: 'referrerpolicy',
};

/**
 * Applies a React props object as HTML attributes on a DOM element,
 * translating camelCase prop names to their HTML attribute equivalents.
 */
export default function applyAttributes(el: Element, props: Record<string, unknown>): void {
  for (const [key, value] of Object.entries(props)) {
    if (key === 'children') continue;

    if (value === undefined || value === null) continue;

    const attrName = ATTR_MAP[key] ?? key;

    if (typeof value === 'boolean') {
      if (value) el.setAttribute(attrName, '');
    } else if (typeof value === 'string' || typeof value === 'number' || typeof value === 'bigint') {
      el.setAttribute(attrName, String(value));
    }
  }
}
