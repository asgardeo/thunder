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
 * Users Page Object Model
 *
 * Encapsulates all locators and actions for the User Management page.
 *
 * @example
 * const usersPage = new UsersPage(page, baseUrl);
 * await usersPage.goto();
 * await usersPage.createUser({ username: 'test', email: 'test@test.com' });
 */

import { Page, Locator, expect } from "@playwright/test";
import { DeveloperPortalRoutes } from "../../configs/routes/developer-portal-routes";
import { BasePage } from "../base.page";
import { Timeouts } from "../../constants/timeouts";

export type UserFormData = {
  username: string;
  email: string;
  firstName?: string;
  lastName?: string;
};

export class UsersPage extends BasePage {
  readonly baseUrl: string;

  // Page Locators
  readonly addUserButton: Locator;
  readonly userTable: Locator;
  readonly searchInput: Locator;

  // Form Locators
  readonly usernameInput: Locator;
  readonly emailInput: Locator;
  readonly firstNameInput: Locator;
  readonly lastNameInput: Locator;
  readonly submitButton: Locator;
  readonly cancelButton: Locator;
  readonly formHeading: Locator;

  // Messages
  readonly successMessage: Locator;
  readonly errorMessage: Locator;

  constructor(page: Page, baseUrl: string) {
    super(page);
    this.baseUrl = baseUrl;

    // Add User button
    this.addUserButton = page
      .getByRole("button", { name: /add user/i })
      .or(page.locator('button:has-text("Add User")'))
      .or(page.locator('button:has-text("+ Add User")'))
      .or(page.locator('[data-testid*="add"][data-testid*="user"]'))
      .or(page.locator('a:has-text("Add User")'));

    // User table
    this.userTable = page.locator('table, [role="table"], [data-testid*="user-list"]');

    // Search input
    this.searchInput = page.locator('input[placeholder*="search" i], input[type="search"]');

    // Form fields
    this.usernameInput = page.locator('input[name="username"]').or(page.getByLabel(/username/i));

    this.emailInput = page.locator('input[name="email"]').or(page.getByLabel(/email/i));

    this.firstNameInput = page
      .locator('input[name="given_name"], input[name="firstName"]')
      .or(page.getByLabel(/first.*name|given.*name/i));

    this.lastNameInput = page
      .locator('input[name="family_name"], input[name="lastName"]')
      .or(page.getByLabel(/last.*name|family.*name/i));

    // Form buttons
    this.submitButton = page.getByRole("button", { name: /create.*user|add.*user|submit|save/i });
    this.cancelButton = page.getByRole("button", { name: /cancel|close/i });

    // Form heading
    this.formHeading = page.locator("h1, h2, h3, h4, h5, h6").filter({ hasText: /create.*user|add.*user|new.*user/i });

    // Messages
    this.successMessage = page.locator('[class*="success"], [role="status"]');
    this.errorMessage = page.locator('[class*="error"], [role="alert"]');
  }

  /** Navigate to users management page */
  async goto() {
    await this.page.goto(`${this.baseUrl}${DeveloperPortalRoutes.users}`, {
      waitUntil: "networkidle",
      timeout: Timeouts.PAGE_LOAD,
    });
  }

  /** Check if currently on users page */
  async isOnUsersPage(): Promise<boolean> {
    const url = this.page.url();
    return url.includes(DeveloperPortalRoutes.users) && !url.includes(DeveloperPortalRoutes.signin);
  }

  /** Verify page loaded successfully */
  async verifyPageLoaded() {
    const url = this.page.url();
    if (url.includes(DeveloperPortalRoutes.signin)) {
      throw new Error("Authentication failed: Redirected to signin page");
    }
    expect(url).toContain(DeveloperPortalRoutes.users);
  }

  /** Click the Add User button */
  async clickAddUser() {
    await this.addUserButton.first().waitFor({ state: "visible", timeout: Timeouts.ELEMENT_VISIBILITY });
    await this.addUserButton.first().scrollIntoViewIfNeeded();
    await this.addUserButton.first().click();
  }

  /** Wait for user creation form */
  async waitForUserForm() {
    await expect(this.formHeading.first()).toBeVisible({ timeout: Timeouts.FORM_LOAD });
  }

  /** Fill the user form */
  async fillUserForm(data: UserFormData) {
    await this.usernameInput.first().fill(data.username);
    await this.emailInput.first().fill(data.email);
    if (data.firstName) await this.firstNameInput.first().fill(data.firstName);
    if (data.lastName) await this.lastNameInput.first().fill(data.lastName);
  }

  /** Submit the form */
  async submitForm() {
    await this.submitButton.first().click();
  }

  /** Cancel the form */
  async cancelForm() {
    await this.cancelButton.first().click();
  }

  /** Create a new user (complete flow) */
  async createUser(data: UserFormData) {
    await this.clickAddUser();
    await this.waitForUserForm();
    await this.fillUserForm(data);
    await this.submitForm();
  }

  /** Search for a user */
  async searchUser(query: string) {
    await this.searchInput.first().fill(query);
    // Using network idle after triggering search.
    // This is acceptable here because the users page is expected not to keep long-lived
    // connections (e.g., websockets) and search is the primary network activity.
    // If additional long-running requests are introduced, prefer a more targeted wait
    // such as page.waitForResponse() for the search API or waiting for the results
    // table locator to update instead of relying on 'networkidle'.
    await this.page.waitForLoadState("networkidle");
  }

  /** Get user count */
  async getUserCount(): Promise<number> {
    const rows = this.page.locator('table tbody tr, [role="row"]');
    return await rows.count();
  }
}
