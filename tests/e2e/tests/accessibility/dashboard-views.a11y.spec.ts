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
 * Dashboard & Main Views — Accessibility Tests
 *
 * Validates WCAG 2.1 AA compliance on authenticated pages:
 * dashboard, navigation, and user management.
 *
 * These tests use the stored authentication state from the setup project,
 * so they run as an authenticated admin user.
 *
 * @see https://www.w3.org/WAI/WCAG22/quickref/
 */

import { test, expect } from "@playwright/test";
import {
  expectNoA11yViolations,
  checkAriaLiveRegions,
  A11Y_RULE_SETS,
} from "../../utils/accessibility";

/**
 * Known accessibility violations in the current Thunder app.
 * TODO: Remove these exclusions as the product fixes each issue.
 */
const KNOWN_VIOLATIONS = ["document-title", "html-has-lang"];

test.describe("Accessibility — Dashboard & Main Views @accessibility", () => {
  test.describe("Dashboard Home", () => {
    test.beforeEach(async ({ page }) => {
      await page.goto("/", { waitUntil: "networkidle" });
    });

    test(
      "TC-A11Y-DASH-001: Dashboard page meets WCAG 2.1 AA standards",
      async ({ page }, testInfo) => {
        await test.step("Run axe-core WCAG 2.1 AA audit on dashboard", async () => {
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
      "TC-A11Y-DASH-002: Dashboard has proper landmark regions",
      async ({ page }, testInfo) => {
        await test.step("Verify ARIA landmarks and regions", async () => {
          await expectNoA11yViolations(
            page,
            {
              includeRules: [
                "landmark-banner-is-top-level",
                "landmark-contentinfo-is-top-level",
                "landmark-main-is-top-level",
                "landmark-no-duplicate-banner",
                "landmark-no-duplicate-contentinfo",
                "landmark-no-duplicate-main",
                "landmark-one-main",
                "landmark-unique",
                "region",
              ],
            },
            testInfo,
          );
        });
      },
    );

    test(
      "TC-A11Y-DASH-003: Dashboard has valid heading hierarchy",
      async ({ page }, testInfo) => {
        await test.step("Check heading structure across dashboard", async () => {
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
      "TC-A11Y-DASH-004: Dashboard ARIA live regions are properly configured",
      async ({ page }) => {
        await test.step("Check ARIA live regions for dynamic content", async () => {
          const liveRegions = await checkAriaLiveRegions(page);

          // Validate that any live regions found have valid politeness values
          for (const region of liveRegions) {
            expect(["polite", "assertive", "off"]).toContain(region.politeness);
          }
        });
      },
    );
  });

  test.describe("Navigation", () => {
    test.beforeEach(async ({ page }) => {
      await page.goto("/", { waitUntil: "networkidle" });
    });

    test(
      "TC-A11Y-DASH-005: Navigation/sidebar is accessible",
      async ({ page }, testInfo) => {
        await test.step("Verify navigation accessibility", async () => {
          await expectNoA11yViolations(
            page,
            {
              includeRules: [
                "link-name",
                "link-in-text-block",
                "aria-required-attr",
                "aria-valid-attr",
              ],
              excludeRules: KNOWN_VIOLATIONS,
            },
            testInfo,
          );
        });
      },
    );

    test(
      "TC-A11Y-DASH-006: Navigation links have descriptive accessible names",
      async ({ page }, testInfo) => {
        await test.step("Verify link accessibility", async () => {
          await expectNoA11yViolations(
            page,
            {
              includeRules: [
                "link-name",
                "link-in-text-block",
              ],
            },
            testInfo,
          );
        });
      },
    );
  });

  test.describe("User Management Page", () => {
    test.beforeEach(async ({ page }) => {
      await page.goto("/manage/users", { waitUntil: "networkidle" });
    });

    test(
      "TC-A11Y-DASH-007: User management page meets WCAG 2.1 AA standards",
      async ({ page }, testInfo) => {
        await test.step("Run full WCAG 2.1 AA audit on user management page", async () => {
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
      "TC-A11Y-DASH-008: User management tables are accessible",
      async ({ page }, testInfo) => {
        await test.step("Verify table accessibility", async () => {
          await expectNoA11yViolations(
            page,
            {
              includeRules: [
                "table-duplicate-name",
                "td-headers-attr",
                "th-has-data-cells",
                "td-has-header",
                "scope-attr-valid",
              ],
              includeSelectors: ["table", "[role='table']", "[role='grid']"],
            },
            testInfo,
          );
        });
      },
    );
  });

  test.describe("Comprehensive WCAG 2.2 AA", () => {
    test(
      "TC-A11Y-DASH-009: Dashboard passes WCAG 2.2 AA audit",
      async ({ page }, testInfo) => {
        await page.goto("/", { waitUntil: "networkidle" });

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
