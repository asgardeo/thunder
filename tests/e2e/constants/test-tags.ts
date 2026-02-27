/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

/**
 * Test Tags for organizing and filtering tests
 *
 * Usage in tests:
 * test('my test', { tag: [TestTags.SMOKE, TestTags.USER_MANAGEMENT] }, async ({ page }) => {
 *   // test code
 * });
 *
 * Run specific tags:
 * npx playwright test --grep @smoke
 * npx playwright test --grep-invert @slow
 */

export const TestTags = {
  /** Critical path tests that must pass */
  SMOKE: "@smoke",

  /** Tests that cover happy path scenarios */
  HAPPY_PATH: "@happy-path",

  /** Tests that cover error scenarios */
  ERROR_HANDLING: "@error-handling",

  /** Tests that are known to be slow */
  SLOW: "@slow",

  /** Tests that are flaky and need investigation */
  FLAKY: "@flaky",

  /** User management related tests */
  USER_MANAGEMENT: "@user-management",

  /** Authentication related tests */
  AUTHENTICATION: "@authentication",

  /** Application management tests */
  APPLICATIONS: "@applications",

  /** API related tests */
  API: "@api",

  /** Accessibility (a11y) related tests */
  ACCESSIBILITY: "@accessibility",
} as const;

export type TestTag = (typeof TestTags)[keyof typeof TestTags];
