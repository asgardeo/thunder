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
 * Visual Regression Test Helpers
 *
 * Provides utilities for visual regression testing using Playwright's screenshot comparison.
 *
 * @example
 * import { expectVisualMatch } from '../../utils/visual-testing';
 *
 * test('homepage looks correct', async ({ page }) => {
 *   await page.goto('/');
 *   await expectVisualMatch(page, 'homepage');
 * });
 */

import { Page, expect } from "@playwright/test";

export interface VisualTestOptions {
  /** Maximum allowed pixel difference (0-1) */
  maxDiffPixelRatio?: number;

  /** Maximum allowed different pixels */
  maxDiffPixels?: number;

  /** Mask elements that might change (dates, random IDs) */
  mask?: Array<any>;

  /** Wait for animations to complete */
  animations?: "disabled" | "allow";

  /** Full page screenshot or viewport only */
  fullPage?: boolean;
}

/**
 * Compare page screenshot with baseline
 */
export async function expectVisualMatch(page: Page, screenshotName: string, options: VisualTestOptions = {}) {
  const { maxDiffPixelRatio = 0.01, maxDiffPixels, mask = [], animations = "disabled", fullPage = false } = options;

  await expect(page).toHaveScreenshot(`${screenshotName}.png`, {
    maxDiffPixelRatio,
    maxDiffPixels,
    mask,
    animations,
    fullPage,
  });
}

/**
 * Compare element screenshot with baseline
 */
export async function expectElementVisualMatch(
  element: any,
  screenshotName: string,
  options: Omit<VisualTestOptions, "fullPage"> = {}
) {
  const { maxDiffPixelRatio = 0.01, maxDiffPixels, mask = [], animations = "disabled" } = options;

  await expect(element).toHaveScreenshot(`${screenshotName}.png`, {
    maxDiffPixelRatio,
    maxDiffPixels,
    mask,
    animations,
  });
}
