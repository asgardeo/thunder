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
 * Browser mode test setup for Vitest Browser Mode
 * This file is used instead of setup.ts when running tests in real browsers
 * via Vitest Browser Mode with Playwright.
 *
 * Unlike the jsdom setup, this does not need:
 * - @testing-library/jest-dom (browser mode has built-in matchers)
 * - cleanup() calls (vitest-browser-react handles this automatically)
 * - jsdom workarounds (IntersectionObserver, ResizeObserver, etc.)
 */

import {beforeAll, vi} from 'vitest';
import i18n from 'i18next';
import {initReactI18next} from 'react-i18next';
import enUS from '@thunder/i18n/locales/en-US';

// Initialize i18n for tests
beforeAll(async () => {
  await i18n.use(initReactI18next).init({
    resources: {
      'en-US': {
        common: enUS.common,
        navigation: enUS.navigation,
        users: enUS.users,
        userTypes: enUS.userTypes,
        integrations: enUS.integrations,
        applications: enUS.applications,
        dashboard: enUS.dashboard,
        auth: enUS.auth,
        mfa: enUS.mfa,
        social: enUS.social,
        consent: enUS.consent,
        errors: enUS.errors,
      },
    },
    lng: 'en-US',
    fallbackLng: 'en-US',
    defaultNS: 'common',
    interpolation: {
      escapeValue: false,
    },
    // Disable Suspense in tests for faster execution
    react: {
      useSuspense: false,
    },
  });
});

// Mock global for Node.js built-ins used by @asgardeo packages
if (typeof window !== 'undefined') {
  (window as unknown as {global: Window}).global = window;
}

// Mock @asgardeo/react to avoid buffer import issues in tests
vi.mock('@asgardeo/react', () => ({
  useAsgardeo: vi.fn(() => ({
    http: {
      request: vi.fn(),
    },
    signIn: vi.fn(),
    signOut: vi.fn(),
    getAccessToken: vi.fn(),
    getIDToken: vi.fn(),
    getDecodedIDToken: vi.fn(),
    isAuthenticated: false,
    isLoading: false,
  })),
  AsgardeoProvider: ({children}: {children: React.ReactNode}) => children,
}));
