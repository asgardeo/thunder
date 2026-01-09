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

import { Page } from "@playwright/test";
import path from "path";

/**
 * Base Page Object Model
 *
 * Provides common functionality for all page objects, such as
 * screenshot capabilities and shared locators.
 */
export class BasePage {
  constructor(readonly page: Page) {}

  /**
   * Take a full-page screenshot and save it to test-results/debug.
   *
   * The screenshot is saved as a PNG file in the `tests/e2e/test-results/debug` directory.
   * Useful for debugging visual states during test execution.
   *
   * @param name Name of the screenshot file (without extension)
   */
  async screenshot(name: string) {
    // Resolve path relative to this file's location (tests/e2e/pages)
    // We want to go up one level to e2e, then into test-results
    const screenshotPath = path.resolve(__dirname, "../test-results/debug", `${name}.png`);

    await this.page.screenshot({
      path: screenshotPath,
      fullPage: true,
    });
  }
}
