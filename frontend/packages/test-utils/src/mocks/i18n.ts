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

import {type Mock, vi} from 'vitest';

/**
 * Mock implementation of useTranslation hook for testing
 */
export const mockUseTranslation = (): {t: (key: string) => string; i18n: {changeLanguage: Mock; language: string}} => ({
  t: (key: string) => key,
  i18n: {
    changeLanguage: vi.fn(),
    language: 'en',
  },
});

/**
 * Mock implementation of useLanguage hook for testing
 */
export const mockUseLanguage = (): {currentLanguage: string; setLanguage: Mock; availableLanguages: string[]} => ({
  currentLanguage: 'en',
  setLanguage: vi.fn(),
  availableLanguages: ['en', 'si'],
});

/**
 * Mock implementation of useDataGridLocaleText hook for testing
 */
export const mockUseDataGridLocaleText = () => ({
  noRowsLabel: 'No rows',
  noResultsOverlayLabel: 'No results found.',
  paginationRowsPerPage: 'Rows per page:',
});
