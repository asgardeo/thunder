/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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
 * Developer Portal Sign-in Page Object Model
 * 
 * Encapsulates all locators and actions for the Developer Portal login page.
 * 
 * @example
 * const signinPage = new DeveloperPortalSigninPage(page, baseUrl);
 * await signinPage.goto();
 * await signinPage.login('admin', 'password');
 */

import { Page, Locator, expect } from '@playwright/test';
import { DeveloperPortalRoutes } from '../../configs/routes/developer-portal-routes';
import { BasePage } from '../base.page';
import { Timeouts } from '../../constants/timeouts';

export class DeveloperPortalSigninPage extends BasePage {
    readonly baseUrl: string;

    // Locators
    readonly usernameInput: Locator;
    readonly passwordInput: Locator;
    readonly signInButton: Locator;
    readonly errorMessage: Locator;

    constructor(page: Page, baseUrl: string) {
        super(page);
        this.baseUrl = baseUrl;

        // Username field
        this.usernameInput = page.locator('input[name="username"]')
            .or(page.locator('input[type="text"]'))
            .or(page.locator('input[id*="username"]'));

        // Password field
        this.passwordInput = page.locator('input[name="password"]')
            .or(page.locator('input[type="password"]'));

        // Sign in button
        this.signInButton = page.locator('button[type="submit"]')
            .or(page.getByRole('button', { name: /sign in|login|submit/i }));

        // Error message
        this.errorMessage = page.locator('[class*="error"], [role="alert"], .error-message');
    }

    /** Navigate to the login page */
    async goto() {
        await this.page.goto(`${this.baseUrl}${DeveloperPortalRoutes.signin}`, {
            waitUntil: 'networkidle'
        });
    }

    /** Navigate to home page (redirects to login if not authenticated) */
    async gotoHome() {
        await this.page.goto(`${this.baseUrl}${DeveloperPortalRoutes.home}`, {
            waitUntil: 'networkidle'
        });
    }

    /** Check if currently on login page */
    async isOnLoginPage(): Promise<boolean> {
        const url = this.page.url();
        return url.includes(DeveloperPortalRoutes.signin) ||
            url.includes('/auth') ||
            url.includes('/login');
    }

    /** Wait for login form to be visible */
    async waitForLoginForm() {
        await this.usernameInput.first().waitFor({ state: 'visible', timeout: Timeouts.ELEMENT_VISIBILITY });
    }

    /** Fill username field */
    async fillUsername(username: string) {
        await this.usernameInput.first().fill(username);
    }

    /** Fill password field */
    async fillPassword(password: string) {
        await this.passwordInput.first().fill(password);
    }

    /** Click the sign in button */
    async clickSignIn() {
        await this.signInButton.first().click();
    }

    /** Perform complete login flow */
    async login(username: string, password: string) {
        await this.waitForLoginForm();
        await this.fillUsername(username);
        await this.fillPassword(password);
        await this.clickSignIn();
    }

    /** Wait for successful login */
    async waitForLoginSuccess() {
        await this.page.waitForURL(`**${DeveloperPortalRoutes.home}/**`, { timeout: Timeouts.PAGE_LOAD });
        await this.page.waitForLoadState('networkidle');
    }

    /** Verify login was successful */
    async verifyLoginSuccess() {
        const url = this.page.url();
        expect(url).toContain(DeveloperPortalRoutes.home);
        expect(url).not.toContain(DeveloperPortalRoutes.signin);
    }
}
