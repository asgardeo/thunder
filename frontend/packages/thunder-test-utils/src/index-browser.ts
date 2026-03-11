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

import {page as _page} from 'vitest/browser';

export {default as render, renderWithProviders, renderHook, getByTranslationKey, configureTestUtils} from './test-utils-browser';
export type {ThunderTestConfig} from './test-utils';

// Re-export everything from vitest-browser-react for convenience
export * from 'vitest-browser-react';
// Re-export page and userEvent from vitest/browser
export {page, userEvent} from 'vitest/browser';

/**
 * Locate an input/textarea/select element by its current display value.
 * This bridges the gap from @testing-library/react's `getByDisplayValue`
 * to Vitest Browser Mode, which doesn't have a built-in equivalent.
 *
 * Checks both the HTML `value` attribute and the DOM `.value` property
 * to handle React controlled inputs which only update the property.
 */
export function getByDisplayValue(value: string) {
  // First try the attribute selector (fast path)
  const byAttr = document.querySelector<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>(
    `input[value="${CSS.escape(value)}"], textarea[value="${CSS.escape(value)}"]`,
  );
  if (byAttr) {
    return _page.elementLocator(byAttr);
  }

  // Fall back to checking the DOM .value property (React controlled inputs)
  const allEls = document.querySelectorAll<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>(
    'input, textarea, select',
  );
  const match = Array.from(allEls).find((e) => e.value === value);
  if (!match) {
    throw new Error(`Unable to find an element with display value: ${value}`);
  }
  return _page.elementLocator(match);
}
