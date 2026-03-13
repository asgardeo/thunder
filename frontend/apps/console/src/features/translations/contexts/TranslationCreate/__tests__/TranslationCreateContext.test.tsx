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

import {describe, expect, it, vi} from 'vitest';
import {render, screen} from '@testing-library/react';
import {useContext, useMemo} from 'react';
import TranslationCreateContext, {type TranslationCreateContextType} from '../TranslationCreateContext';

// Test component to consume the context directly
function TestConsumer() {
  const context = useContext(TranslationCreateContext);

  if (!context) {
    return <div data-testid="context">undefined</div>;
  }

  return (
    <div>
      <div data-testid="context">defined</div>
      <div data-testid="context-type">{typeof context}</div>
      <div data-testid="current-step">{context.currentStep}</div>
      <div data-testid="locale-code">{context.localeCode}</div>
      <div data-testid="selected-country">{context.selectedCountry?.name ?? 'null'}</div>
    </div>
  );
}

// Test component with a mock context value
function TestWithMockValue() {
  const mockContextValue: TranslationCreateContextType = useMemo(
    () => ({
      currentStep: 'LANGUAGE',
      setCurrentStep: vi.fn(),
      selectedCountry: {name: 'France', regionCode: 'FR', flag: '🇫🇷'},
      setSelectedCountry: vi.fn(),
      selectedLocale: null,
      setSelectedLocale: vi.fn(),
      localeCodeOverride: '',
      setLocaleCodeOverride: vi.fn(),
      localeCode: 'fr-FR',
      populateFromEnglish: true,
      setPopulateFromEnglish: vi.fn(),
      isCreating: false,
      setIsCreating: vi.fn(),
      progress: 0,
      setProgress: vi.fn(),
      error: null,
      setError: vi.fn(),
      reset: vi.fn(),
    }),
    [],
  );

  return (
    <TranslationCreateContext.Provider value={mockContextValue}>
      <TestConsumer />
    </TranslationCreateContext.Provider>
  );
}

describe('TranslationCreateContext', () => {
  it('provides undefined value when used without provider', () => {
    render(<TestConsumer />);

    expect(screen.getByTestId('context')).toHaveTextContent('undefined');
  });

  it('provides context value when used with provider', () => {
    render(<TestWithMockValue />);

    expect(screen.getByTestId('context')).toHaveTextContent('defined');
    expect(screen.getByTestId('context-type')).toHaveTextContent('object');
  });

  it('provides correct context properties when used with provider', () => {
    render(<TestWithMockValue />);

    expect(screen.getByTestId('current-step')).toHaveTextContent('LANGUAGE');
    expect(screen.getByTestId('locale-code')).toHaveTextContent('fr-FR');
    expect(screen.getByTestId('selected-country')).toHaveTextContent('France');
  });

  it('has correct TypeScript interface definition', () => {
    const mockContext: TranslationCreateContextType = {
      currentStep: 'COUNTRY',
      setCurrentStep: () => {},
      selectedCountry: null,
      setSelectedCountry: () => {},
      selectedLocale: null,
      setSelectedLocale: () => {},
      localeCodeOverride: '',
      setLocaleCodeOverride: () => {},
      localeCode: '',
      populateFromEnglish: true,
      setPopulateFromEnglish: () => {},
      isCreating: false,
      setIsCreating: () => {},
      progress: 0,
      setProgress: () => {},
      error: null,
      setError: () => {},
      reset: () => {},
    };

    expect(mockContext).toBeDefined();
    expect(typeof mockContext.currentStep).toBe('string');
    expect(typeof mockContext.setCurrentStep).toBe('function');
    expect(mockContext.selectedCountry).toBeNull();
    expect(mockContext.selectedLocale).toBeNull();
    expect(typeof mockContext.localeCode).toBe('string');
    expect(typeof mockContext.reset).toBe('function');
  });

  it('allows null values for nullable properties', () => {
    const mockContext: TranslationCreateContextType = {
      currentStep: 'COUNTRY',
      setCurrentStep: () => {},
      selectedCountry: null,
      setSelectedCountry: () => {},
      selectedLocale: null,
      setSelectedLocale: () => {},
      localeCodeOverride: '',
      setLocaleCodeOverride: () => {},
      localeCode: '',
      populateFromEnglish: true,
      setPopulateFromEnglish: () => {},
      isCreating: false,
      setIsCreating: () => {},
      progress: 0,
      setProgress: () => {},
      error: null,
      setError: () => {},
      reset: () => {},
    };

    expect(mockContext.selectedCountry).toBeNull();
    expect(mockContext.selectedLocale).toBeNull();
    expect(mockContext.error).toBeNull();
  });

  it('creates context with expected default value (undefined)', () => {
    expect(TranslationCreateContext).toBeDefined();
    expect(typeof TranslationCreateContext).toBe('object');
  });
});
