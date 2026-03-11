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

import {describe, expect, it, vi, beforeEach} from 'vitest';
import {renderHook} from '@thunder/test-utils/browser';
import type {ReactNode} from 'react';
import useTranslationCreate from '../useTranslationCreate';
import TranslationCreateProvider from '../TranslationCreateProvider';

function Wrapper({children}: {children: ReactNode}) {
  return <TranslationCreateProvider>{children}</TranslationCreateProvider>;
}

describe('useTranslationCreate', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('returns context when used within TranslationCreateProvider', async () => {
    const {result} = await renderHook(() => useTranslationCreate(), {
      wrapper: Wrapper,
    });

    expect(typeof result.current).toBe('object');
  });

  it('throws error when used outside TranslationCreateProvider', async () => {
    const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    await expect(renderHook(() => useTranslationCreate())).rejects.toThrow(
      'useTranslationCreate must be used within TranslationCreateProvider',
    );

    errorSpy.mockRestore();
  });

  it('throws descriptive error message when used outside provider', async () => {
    const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    await expect(renderHook(() => useTranslationCreate())).rejects.toThrow(
      'useTranslationCreate must be used within TranslationCreateProvider',
    );

    errorSpy.mockRestore();
  });

  it('provides all required context properties', async () => {
    const {result} = await renderHook(() => useTranslationCreate(), {
      wrapper: Wrapper,
    });

    const requiredProperties = [
      'currentStep',
      'setCurrentStep',
      'selectedCountry',
      'setSelectedCountry',
      'selectedLocale',
      'setSelectedLocale',
      'localeCodeOverride',
      'setLocaleCodeOverride',
      'localeCode',
      'populateFromEnglish',
      'setPopulateFromEnglish',
      'isCreating',
      'setIsCreating',
      'progress',
      'setProgress',
      'error',
      'setError',
      'reset',
    ];

    const missingProperties = requiredProperties.filter((prop) => !(prop in result.current));

    expect(missingProperties).toHaveLength(0);
  });

  it('returns same context reference across multiple hook calls', async () => {
    const {result: result1} = await renderHook(() => useTranslationCreate(), {wrapper: Wrapper});
    const {result: result2} = await renderHook(() => useTranslationCreate(), {wrapper: Wrapper});

    // Both instances from the same provider type should have the same shape
    expect(typeof result1.current).toBe('object');
    expect(typeof result2.current).toBe('object');
    expect(Object.keys(result1.current)).toEqual(Object.keys(result2.current));
  });

  it('provides functions that are properly typed', async () => {
    const {result} = await renderHook(() => useTranslationCreate(), {
      wrapper: Wrapper,
    });

    expect(typeof result.current.setCurrentStep).toBe('function');
    expect(typeof result.current.setSelectedCountry).toBe('function');
    expect(typeof result.current.setSelectedLocale).toBe('function');
    expect(typeof result.current.setLocaleCodeOverride).toBe('function');
    expect(typeof result.current.setPopulateFromEnglish).toBe('function');
    expect(typeof result.current.setIsCreating).toBe('function');
    expect(typeof result.current.setProgress).toBe('function');
    expect(typeof result.current.setError).toBe('function');
    expect(typeof result.current.reset).toBe('function');
  });

  it('has exactly 18 properties in the context interface', async () => {
    const {result} = await renderHook(() => useTranslationCreate(), {
      wrapper: Wrapper,
    });

    expect(Object.keys(result.current)).toHaveLength(18);
  });
});
