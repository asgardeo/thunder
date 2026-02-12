/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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

/* eslint-disable max-classes-per-file */

import '@testing-library/jest-dom';
import {cleanup} from '@testing-library/react';
import {afterEach, beforeAll, vi} from 'vitest';
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

// Cleanup after each test
afterEach(() => {
  cleanup();
});

// Patch CSSStyleDeclaration.setProperty to handle cssstyle errors with CSS variables in shorthand properties
// This is a known issue with jsdom/cssstyle when using CSS variables like `var(--rowBorderColor)` in border shorthand
// eslint-disable-next-line @typescript-eslint/unbound-method
const originalSetProperty = window.CSSStyleDeclaration.prototype.setProperty;
// eslint-disable-next-line func-names
window.CSSStyleDeclaration.prototype.setProperty = function (
  this: CSSStyleDeclaration,
  property: string,
  value: string | null,
  priority?: string,
) {
  try {
    originalSetProperty.call(this, property, value, priority ?? '');
  } catch {
    // Silently ignore cssstyle errors for CSS variables in shorthand properties
  }
};

// Mock HTMLMediaElement methods that don't exist in jsdom
Object.defineProperty(window.HTMLMediaElement.prototype, 'play', {
  configurable: true,
  value: () => Promise.resolve(),
});

Object.defineProperty(window.HTMLMediaElement.prototype, 'pause', {
  configurable: true,
  value: () => {
    // Intentionally empty
  },
});

Object.defineProperty(window.HTMLMediaElement.prototype, 'load', {
  configurable: true,
  value: () => {
    // Intentionally empty
  },
});

// Mock IntersectionObserver
global.IntersectionObserver = class IntersectionObserver {
  readonly root = null;

  readonly rootMargin = '';

  readonly thresholds = [];

  observe() {
    return this;
  }

  disconnect() {
    return this;
  }

  unobserve() {
    return this;
  }

  // eslint-disable-next-line @typescript-eslint/class-methods-use-this
  takeRecords() {
    return [];
  }
} as unknown as typeof IntersectionObserver;

// Mock ResizeObserver
global.ResizeObserver = class ResizeObserver {
  observe() {
    return this;
  }

  disconnect() {
    return this;
  }

  unobserve() {
    return this;
  }
} as unknown as typeof ResizeObserver;

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
