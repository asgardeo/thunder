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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {waitFor} from '@testing-library/react';
import {useTranslation} from 'react-i18next';
import {renderHook} from '../../test/test-utils';

// Mock @thunder/i18n
vi.mock('@thunder/i18n', () => ({
  LANGUAGE_CONFIGS: {
    'en-US': {
      code: 'en-US',
      name: 'English (US)',
      nativeName: 'English (US)',
      direction: 'ltr',
    },
  },
}));

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: vi.fn(),
}));

// Import after mocking
const {useLanguage} = await import('../useLanguage');

describe('useLanguage', () => {
  const mockChangeLanguage = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    mockChangeLanguage.mockResolvedValue(undefined);

    vi.mocked(useTranslation).mockReturnValue({
      i18n: {
        language: 'en-US',
        changeLanguage: mockChangeLanguage,
      },
      // Add other useTranslation return properties as needed
      t: vi.fn(),
      ready: true,
    } as unknown as ReturnType<typeof useTranslation>);
  });

  it('should return current language from i18n', () => {
    const {result} = renderHook(() => useLanguage());

    expect(result.current.currentLanguage).toBe('en-US');
  });

  it('should return available languages from LANGUAGE_CONFIGS', () => {
    const {result} = renderHook(() => useLanguage());

    expect(result.current.availableLanguages).toEqual([
      {
        code: 'en-US',
        name: 'English (US)',
        nativeName: 'English (US)',
        direction: 'ltr',
      },
    ]);
  });

  it('should provide setLanguage function', () => {
    const {result} = renderHook(() => useLanguage());

    expect(typeof result.current.setLanguage).toBe('function');
  });

  it('should call i18n.changeLanguage when setLanguage is invoked', async () => {
    const {result} = renderHook(() => useLanguage());

    await result.current.setLanguage('en-US');

    expect(mockChangeLanguage).toHaveBeenCalledWith('en-US');
    expect(mockChangeLanguage).toHaveBeenCalledTimes(1);
  });

  it('should handle setLanguage as an async function', async () => {
    const {result} = renderHook(() => useLanguage());

    const promise = result.current.setLanguage('en-US');

    expect(promise).toBeInstanceOf(Promise);

    await waitFor(() => {
      expect(mockChangeLanguage).toHaveBeenCalled();
    });
  });

  it('should propagate errors from i18n.changeLanguage', async () => {
    const mockError = new Error('Language change failed');
    mockChangeLanguage.mockRejectedValueOnce(mockError);

    const {result} = renderHook(() => useLanguage());

    await expect(result.current.setLanguage('en-US')).rejects.toThrow('Language change failed');
  });

  it('should return all values with correct types', () => {
    const {result} = renderHook(() => useLanguage());

    expect(result.current).toHaveProperty('currentLanguage');
    expect(result.current).toHaveProperty('availableLanguages');
    expect(result.current).toHaveProperty('setLanguage');

    expect(typeof result.current.currentLanguage).toBe('string');
    expect(Array.isArray(result.current.availableLanguages)).toBe(true);
    expect(typeof result.current.setLanguage).toBe('function');
  });

  it('should handle different language values from i18n', () => {
    vi.mocked(useTranslation).mockReturnValue({
      i18n: {
        language: 'en-US',
        changeLanguage: mockChangeLanguage,
      },
      t: vi.fn(),
      ready: true,
    } as unknown as ReturnType<typeof useTranslation>);

    const {result} = renderHook(() => useLanguage());

    expect(result.current.currentLanguage).toBe('en-US');
  });
});
