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
 * Authentication Setup for Developer Portal E2E Tests
 * 
 * This setup file performs admin login and saves the authentication state
 * (cookies, localStorage, sessionStorage) to a JSON file. Other tests can
 * then reuse this session via Playwright's storageState feature.
 * 
 * Required environment variables:
 * - ADMIN_USERNAME: Admin user login name
 * - ADMIN_PASSWORD: Admin user password
 * - BASE_URL: Developer portal base URL
 * 
 * @see https://playwright.dev/docs/auth
 */

import { setup, expect, routes } from '../../fixtures/developer-portal';
import path from 'path';
import fs from 'fs';

/** Path to save authentication state */
const AUTH_FILE = path.join(__dirname, '../../playwright/.auth/devportal-admin.json');

setup('Admin login test', async ({ page, context, signinPage }) => {
  const authDir = path.dirname(AUTH_FILE);
  if (!fs.existsSync(authDir)) {
    fs.mkdirSync(authDir, { recursive: true });
  }

  const username = process.env.ADMIN_USERNAME as string;
  const password = process.env.ADMIN_PASSWORD as string;
  const baseUrl = process.env.BASE_URL as string;

  console.log('🔐 Starting authentication process...');
  console.log('Base URL:', baseUrl);
  console.log('Username:', username);

  let capturedCookies: any[] = [];
  let tokenRequests: any[] = [];

  // Capture auth-related network requests for debugging
  page.on('response', response => {
    if (response.url().includes('/token') || response.url().includes('/auth') || response.url().includes('/session')) {
      tokenRequests.push({ url: response.url(), status: response.status(), headers: response.headers() });
    }
  });

  // Navigate to develop page (redirects to login if not authenticated)
  console.log('🌐 Navigating to develop page for authentication...');
  await signinPage.gotoHome();

  console.log('⏳ Checking if redirected to authentication...');
  if (await signinPage.isOnLoginPage()) {
    console.log('✅ Detected authentication page, ready to login');
  } else {
    console.log('ℹ️ Already on target page, checking if authentication needed');
  }

  await signinPage.screenshot('debug-login-page');

  // Use POM to login
  console.log('📝 Filling login credentials...');
  await signinPage.waitForLoginForm();
  await signinPage.fillUsername(username);
  await signinPage.fillPassword(password);

  await signinPage.screenshot('debug-before-signin');

  await signinPage.clickSignIn();

  console.log('⏳ Waiting for authentication to complete...');

  try {
    await signinPage.waitForLoginSuccess();
    console.log('✅ Redirected to develop page');
  } catch (error) {
    console.log('❌ Failed to redirect to develop page');
    await signinPage.screenshot('debug-after-signin-failed');
    throw error;
  }

  await signinPage.verifyLoginSuccess();
  console.log('✅ Login verified via URL check');

  console.log('⏳ Waiting for authentication to be fully established...');
  await page.waitForTimeout(3000);

  // Check storage availability (using raw page evaluate as this is somewhat internal state check)
  await page.evaluate(() => {
    if (window.localStorage) console.log('LocalStorage available');
    if (window.sessionStorage) console.log('SessionStorage available');
  });

  const extendedCookieCheck = await context.cookies();
  console.log('🔍 Checking all cookies after extended wait...');
  extendedCookieCheck.forEach(cookie => {
    console.log(`  Cookie: ${cookie.name} = ${cookie.value.substring(0, 30)}... (domain: ${cookie.domain}, httpOnly: ${cookie.httpOnly})`);
  });

  const httpOnlyCookies = extendedCookieCheck.filter(c => c.httpOnly);
  console.log('🔒 HTTP-only cookies:', httpOnlyCookies.length);

  await page.waitForTimeout(2000);
  await signinPage.screenshot('debug-after-signin-success');

  console.log('✅ Authentication successful!');
  console.log('🔍 Token/Auth requests captured:', tokenRequests.length);
  tokenRequests.forEach(req => {
    console.log(`  - ${req.url}: ${req.status}`);
  });

  console.log('🧪 Testing navigation to protected page...');
  try {
    await page.goto(`${baseUrl}${routes.applications}`, { waitUntil: 'domcontentloaded', timeout: 10000 });
    if (!page.url().includes(routes.signin)) {
      console.log('✅ Can access protected applications page - auth working');
    } else {
      console.log('❌ Redirected to signin when accessing applications page');
    }
  } catch (error) {
    console.log('⚠️ Error testing protected page:', String(error));
  }

  await signinPage.gotoHome();

  console.log('🍪 Capturing cookies from current session...');

  const allCookies = await context.cookies();
  console.log('All cookies from context:', allCookies.length);

  const urlCookies = await context.cookies(baseUrl);
  console.log('URL-specific cookies:', urlCookies.length);

  const pageCookies = await page.evaluate(() => {
    return document.cookie.split(';').map(cookie => {
      const [name, ...rest] = cookie.split('=');
      return { name: name.trim(), value: rest.join('=').trim() };
    }).filter(cookie => cookie.name && cookie.value);
  });
  console.log('Page-accessible cookies:', pageCookies.length);

  capturedCookies = allCookies.length > 0 ? allCookies : urlCookies;

  console.log('🍪 Final cookie summary:');
  console.log(`  - Context cookies: ${allCookies.length}`);
  console.log(`  - URL cookies: ${urlCookies.length}`);
  console.log(`  - Page cookies: ${pageCookies.length}`);

  if (capturedCookies.length > 0) {
    capturedCookies.forEach(cookie => {
      console.log(`  ✓ ${cookie.name}: ${cookie.domain} (httpOnly: ${cookie.httpOnly}, secure: ${cookie.secure})`);
    });
  } else {
    console.warn('⚠️ No cookies captured! This will likely cause authentication issues.');
    await page.screenshot({ path: 'test-results/debug/debug-no-cookies-captured.png', fullPage: true });
  }

  const storageState = {
    cookies: capturedCookies,
    origins: [{
      origin: baseUrl,
      localStorage: await page.evaluate(() => {
        const items = [];
        for (let i = 0; i < localStorage.length; i++) {
          const key = localStorage.key(i);
          if (key) items.push({ name: key, value: localStorage.getItem(key) });
        }
        return items;
      }),
      sessionStorage: await page.evaluate(() => {
        const items = [];
        for (let i = 0; i < sessionStorage.length; i++) {
          const key = sessionStorage.key(i);
          if (key) items.push({ name: key, value: sessionStorage.getItem(key) });
        }
        return items;
      })
    }]
  };

  fs.writeFileSync(AUTH_FILE, JSON.stringify(storageState, null, 2));
  console.log('💾 Enhanced authentication state saved to:', AUTH_FILE);

  const workingAuthFile = path.join(authDir, 'working-login.json');
  fs.writeFileSync(workingAuthFile, JSON.stringify(storageState, null, 2));
  console.log('💾 Working authentication state also saved to:', workingAuthFile);

  console.log('📋 Saved state summary:');
  console.log(`  - Cookies: ${storageState.cookies.length}`);
  console.log(`  - Origins: ${storageState.origins.length}`);
  console.log(`  - LocalStorage items: ${storageState.origins[0]?.localStorage?.length || 0}`);
  console.log(`  - SessionStorage items: ${storageState.origins[0]?.sessionStorage?.length || 0}`);
});
