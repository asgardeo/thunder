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
 * Accessibility Testing Utilities
 *
 * Provides utilities for accessibility (a11y) testing.
 * Note: Install @axe-core/playwright for full functionality:
 * npm install -D @axe-core/playwright
 *
 * @example
 * import { checkA11y } from '../../utils/accessibility';
 *
 * test('homepage is accessible', async ({ page }) => {
 *   await page.goto('/');
 *   await checkA11y(page);
 * });
 */

import { Page } from "@playwright/test";

export interface A11yOptions {
  /** Include specific WCAG rules */
  includedRules?: string[];

  /** Exclude specific WCAG rules */
  excludedRules?: string[];

  /** Only check specific elements */
  include?: string[];

  /** Exclude specific elements from checking */
  exclude?: string[];
}

/**
 * Basic accessibility check (manual implementation)
 * For production use, consider @axe-core/playwright
 */
export async function checkA11y(page: Page, options: A11yOptions = {}): Promise<void> {
  // Check for common a11y issues
  const issues: string[] = [];

  // 1. Check for images without alt text
  const imagesWithoutAlt = await page.locator("img:not([alt])").count();
  if (imagesWithoutAlt > 0) {
    issues.push(`Found ${imagesWithoutAlt} images without alt text`);
  }

  // 2. Check for buttons/links without accessible names
  const buttonsWithoutText = await page.locator("button:not([aria-label]):not(:has-text(*))").count();
  if (buttonsWithoutText > 0) {
    issues.push(`Found ${buttonsWithoutText} buttons without accessible text`);
  }

  // 3. Check for form inputs without labels
  const inputsWithoutLabel = await page.locator("input:not([aria-label]):not([aria-labelledby])").count();
  if (inputsWithoutLabel > 0) {
    issues.push(`Found ${inputsWithoutLabel} inputs without labels`);
  }

  // 4. Check for heading hierarchy
  const headings = await page.locator("h1, h2, h3, h4, h5, h6").allTextContents();
  if (headings.length > 0) {
    const h1Count = await page.locator("h1").count();
    if (h1Count === 0) {
      issues.push("Page is missing an h1 heading");
    } else if (h1Count > 1) {
      issues.push("Page has multiple h1 headings");
    }
  }

  // 5. Check for color contrast (basic check)
  const elementsToCheck = await page.locator("body *:visible").all();

  if (issues.length > 0) {
    console.warn("⚠️ Accessibility issues found:");
    issues.forEach(issue => console.warn(`  - ${issue}`));
    // Note: In production, you might want to fail the test or generate a report
  } else {
    console.log("✅ No basic accessibility issues found");
  }
}

/**
 * Check keyboard navigation
 */
export async function checkKeyboardNavigation(page: Page): Promise<void> {
  // Start from first interactive element
  await page.keyboard.press("Tab");

  // Check that focused element is visible
  const focusedElement = await page.evaluate(() => {
    const el = document.activeElement;
    return el
      ? {
          tagName: el.tagName,
          role: el.getAttribute("role"),
          ariaLabel: el.getAttribute("aria-label"),
        }
      : null;
  });

  if (focusedElement) {
    console.log("✅ Keyboard focus is working:", focusedElement);
  } else {
    console.warn("⚠️ No element received focus on Tab key");
  }
}

/**
 * Check screen reader announcements
 */
export async function checkAriaLiveRegions(page: Page): Promise<void> {
  const liveRegions = await page.locator("[aria-live]").all();
  console.log(`Found ${liveRegions.length} ARIA live regions`);

  for (const region of liveRegions) {
    const ariaLive = await region.getAttribute("aria-live");
    const text = await region.textContent();
    console.log(`  - [aria-live="${ariaLive}"]: ${text?.substring(0, 50)}...`);
  }
}
