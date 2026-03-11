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

import {beforeAll, vi} from 'vitest';
import i18n from 'i18next';
import {initReactI18next} from 'react-i18next';
import enUS from '@thunder/i18n/locales/en-US';

// Initialize i18n for tests
beforeAll(async () => {
  await i18n.use(initReactI18next).init({
    resources: {
      'en-US': enUS,
    },
    lng: 'en-US',
    fallbackLng: 'en-US',
    defaultNS: 'common',
    keySeparator: false,
    interpolation: {
      escapeValue: false,
    },
    // Disable Suspense in tests for faster execution
    react: {
      useSuspense: false,
    },
  });
});

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
