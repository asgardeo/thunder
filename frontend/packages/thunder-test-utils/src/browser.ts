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
 * Browser mode test utilities for Vitest Browser Mode
 * Import from '@thunder/test-utils/browser' instead of '@thunder/test-utils'
 * when using Vitest Browser Mode with Playwright
 *
 * @example
 * ```typescript
 * import {render} from '@thunder/test-utils/browser';
 * import {page, userEvent} from 'vitest/browser';
 *
 * await render(<MyComponent />);
 *
 * // Use native Vitest Browser Mode page object for queries
 * const button = page.getByRole('button', {name: 'Submit'});
 * await userEvent.click(button);
 *
 * const input = page.getByLabelText('Email');
 * await userEvent.fill(input, 'test@example.com');
 * ```
 */

export {render, renderWithProviders, renderHook, getByTranslationKey, configureTestUtils, resetTestUtils} from './browser-test-utils';
export type {ThunderTestConfig, ThunderRenderOptions, ThunderRenderHookOptions} from './browser-test-utils';
