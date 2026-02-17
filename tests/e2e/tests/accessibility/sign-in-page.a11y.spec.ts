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
 * Authentication Flows — Accessibility Tests
 *
 * Validates WCAG 2.1 AA compliance on authentication-related pages.
 * These tests run against unauthenticated pages (sign-in, sign-up)
 * and do not require pre-existing session state.
 *
 * @see https://www.w3.org/WAI/WCAG22/quickref/
 */

import { test, expect } from "@playwright/test";
import {
  expectNoA11yViolations,
  checkKeyboardNavigation,
  A11Y_RULE_SETS,
} from "../../utils/accessibility";

const baseUrl = process.env.BASE_URL || "https://localhost:8090";

/**
 * Known accessibility violations in the current Thunder app.
 * TODO: Remove these exclusions as the product fixes each issue.
 */
const KNOWN_VIOLATIONS = ["document-title", "html-has-lang"];

/** Selector for visible, enabled interactive elements only. */
const VISIBLE_INTERACTIVE_SELECTOR =
  "input:visible:not([disabled]), " +
  "button:visible:not([disabled]), " +
  "a[href]:visible, " +
  "select:visible:not([disabled]), " +
  "textarea:visible:not([disabled]), " +
  "[tabindex]:not([tabindex='-1']):visible";

test.describe("Accessibility — Authentication Flows @accessibility", () => {
  test.describe("Sign-In Page", () => {
    test.beforeEach(async ({ page }) => {
      await page.goto(baseUrl, { waitUntil: "networkidle" });
    });

    test(
      "TC-A11Y-AUTH-001: Sign-in page meets WCAG 2.1 AA standards",
      async ({ page }, testInfo) => {
        await test.step("Run axe-core WCAG 2.1 AA audit", async () => {
          await expectNoA11yViolations(
            page,
            {
              tags: A11Y_RULE_SETS.WCAG_21_AA,
              excludeRules: KNOWN_VIOLATIONS,
            },
            testInfo,
          );
        });
      },
    );

    test(
      "TC-A11Y-AUTH-002: Sign-in form has proper labels and ARIA attributes",
      async ({ page }, testInfo) => {
        await test.step("Verify form elements are accessible", async () => {
          await expectNoA11yViolations(
            page,
            {
              includeRules: ["label", "label-title-only", "aria-input-field-name"],
            },
            testInfo,
          );
        });

        await test.step("Verify submit button is accessible", async () => {
          const submitButton = page.locator(
            "button[type='submit'], input[type='submit'], button:has-text('Sign'), button:has-text('Log')",
          ).first();

          if (await submitButton.isVisible()) {
            const accessibleName =
              (await submitButton.getAttribute("aria-label")) ||
              (await submitButton.textContent());
            expect(accessibleName).toBeTruthy();
          }
        });
      },
    );

    test(
      "TC-A11Y-AUTH-003: Sign-in page supports keyboard navigation",
      async ({ page }) => {
        await test.step("Verify Tab navigation through interactive elements", async () => {
          const interactiveCount = await page.locator(VISIBLE_INTERACTIVE_SELECTOR).count();

          const result = await checkKeyboardNavigation(page, interactiveCount);

          expect(result.focusedElements.length).toBeGreaterThanOrEqual(interactiveCount);
          expect(result.tabTrapDetected).toBe(false);
        });

        await test.step("Verify focus is received by interactive elements", async () => {
          const firstInput = page.locator("input").first();

          if (await firstInput.isVisible()) {
            await firstInput.focus();
            const isFocused = await page.evaluate(
              () => document.activeElement?.tagName.toLowerCase() === "input",
            );
            expect(isFocused).toBe(true);
          }
        });
      },
    );

    test(
      "TC-A11Y-AUTH-004: Sign-in page has valid heading hierarchy",
      async ({ page }, testInfo) => {
        await test.step("Check heading structure", async () => {
          await expectNoA11yViolations(
            page,
            {
              includeRules: [
                "heading-order",
                "page-has-heading-one",
                "empty-heading",
              ],
            },
            testInfo,
          );
        });
      },
    );

    test(
      "TC-A11Y-AUTH-005: Sign-in page has sufficient color contrast",
      async ({ page }, testInfo) => {
        await test.step("Validate WCAG AA color contrast ratios", async () => {
          await expectNoA11yViolations(
            page,
            {
              includeRules: ["color-contrast", "color-contrast-enhanced"],
            },
            testInfo,
          );
        });
      },
    );

    test(
      "TC-A11Y-AUTH-006: Sign-in error states are accessible",
      async ({ page }, testInfo) => {
        await test.step("Trigger validation error", async () => {
          const submitButton = page.locator(
            "button[type='submit'], input[type='submit'], button:has-text('Sign'), button:has-text('Log')",
          ).first();

          if (await submitButton.isVisible()) {
            await submitButton.click();

            // Wait for actual error UI to appear instead of arbitrary timeout
            await page.locator(
              "[role='alert'], [aria-invalid='true'], .error, .validation-error",
            ).first().waitFor({ state: "visible", timeout: 5000 }).catch(() => {
              // If no error element appears, proceed anyway — the page state is still valid to audit
            });
          }
        });

        await test.step("Verify error messages are accessible", async () => {
          await expectNoA11yViolations(
            page,
            {
              tags: A11Y_RULE_SETS.WCAG_21_AA,
              excludeRules: KNOWN_VIOLATIONS,
            },
            testInfo,
          );
        });
      },
    );
  });

  test.describe("Sign-In Page — Comprehensive WCAG 2.2 AA", () => {
    test(
      "TC-A11Y-AUTH-007: Sign-in page passes WCAG 2.2 AA audit",
      async ({ page }, testInfo) => {
        await page.goto(baseUrl, { waitUntil: "networkidle" });

        await expectNoA11yViolations(
          page,
          {
            tags: A11Y_RULE_SETS.WCAG_22_AA,
            excludeRules: KNOWN_VIOLATIONS,
          },
          testInfo,
        );
      },
    );
  });
});
