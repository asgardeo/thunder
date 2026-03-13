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
import {render, screen} from '@thunder/test-utils';
import React from 'react';
import useTranslationCreate from '../useTranslationCreate';
import TranslationCreateProvider from '../TranslationCreateProvider';

// Test component to consume the hook successfully
function TestConsumer() {
  const context = useTranslationCreate();

  return <div data-testid="context-available">{typeof context}</div>;
}

// Test component to trigger the error path
function TestConsumerWithoutProvider() {
  const context = useTranslationCreate();

  return <div data-testid="context">{JSON.stringify(context)}</div>;
}

function TestWrapper({children}: {children: React.ReactNode}) {
  return children;
}

describe('useTranslationCreate', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('returns context when used within TranslationCreateProvider', () => {
    render(
      <TestWrapper>
        <TranslationCreateProvider>
          <TestConsumer />
        </TranslationCreateProvider>
      </TestWrapper>,
    );

    expect(screen.getByTestId('context-available')).toHaveTextContent('object');
  });

  it('throws error when used outside TranslationCreateProvider', () => {
    const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    expect(() => {
      render(<TestConsumerWithoutProvider />);
    }).toThrow('useTranslationCreate must be used within TranslationCreateProvider');

    errorSpy.mockRestore();
  });

  it('throws descriptive error message when used outside provider', () => {
    const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    let thrownError: Error | null = null;

    try {
      render(<TestConsumerWithoutProvider />);
    } catch (error) {
      thrownError = error as Error;
    }

    expect(thrownError).toBeInstanceOf(Error);
    expect(thrownError?.message).toBe('useTranslationCreate must be used within TranslationCreateProvider');

    errorSpy.mockRestore();
  });

  it('provides all required context properties', () => {
    function TestContextProperties() {
      const context = useTranslationCreate();

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

      const missingProperties = requiredProperties.filter((prop) => !(prop in context));

      return (
        <div>
          <div data-testid="missing-properties">{JSON.stringify(missingProperties)}</div>
          <div data-testid="has-all-properties">{missingProperties.length === 0 ? 'true' : 'false'}</div>
        </div>
      );
    }

    render(
      <TestWrapper>
        <TranslationCreateProvider>
          <TestContextProperties />
        </TranslationCreateProvider>
      </TestWrapper>,
    );

    expect(screen.getByTestId('has-all-properties')).toHaveTextContent('true');
    expect(screen.getByTestId('missing-properties')).toHaveTextContent('[]');
  });

  it('returns same context reference across multiple hook calls', () => {
    function TestMultipleHookCalls() {
      const context1 = useTranslationCreate();
      const context2 = useTranslationCreate();

      return (
        <div>
          <div data-testid="same-reference">{(context1 === context2).toString()}</div>
        </div>
      );
    }

    render(
      <TestWrapper>
        <TranslationCreateProvider>
          <TestMultipleHookCalls />
        </TranslationCreateProvider>
      </TestWrapper>,
    );

    expect(screen.getByTestId('same-reference')).toHaveTextContent('true');
  });

  it('provides functions that are properly typed', () => {
    function TestFunctionTypes() {
      const {
        setCurrentStep,
        setSelectedCountry,
        setSelectedLocale,
        setLocaleCodeOverride,
        setPopulateFromEnglish,
        setIsCreating,
        setProgress,
        setError,
        reset,
      } = useTranslationCreate();

      return (
        <div>
          <div data-testid="setCurrentStep-type">{typeof setCurrentStep}</div>
          <div data-testid="setSelectedCountry-type">{typeof setSelectedCountry}</div>
          <div data-testid="setSelectedLocale-type">{typeof setSelectedLocale}</div>
          <div data-testid="setLocaleCodeOverride-type">{typeof setLocaleCodeOverride}</div>
          <div data-testid="setPopulateFromEnglish-type">{typeof setPopulateFromEnglish}</div>
          <div data-testid="setIsCreating-type">{typeof setIsCreating}</div>
          <div data-testid="setProgress-type">{typeof setProgress}</div>
          <div data-testid="setError-type">{typeof setError}</div>
          <div data-testid="reset-type">{typeof reset}</div>
        </div>
      );
    }

    render(
      <TestWrapper>
        <TranslationCreateProvider>
          <TestFunctionTypes />
        </TranslationCreateProvider>
      </TestWrapper>,
    );

    expect(screen.getByTestId('setCurrentStep-type')).toHaveTextContent('function');
    expect(screen.getByTestId('setSelectedCountry-type')).toHaveTextContent('function');
    expect(screen.getByTestId('setSelectedLocale-type')).toHaveTextContent('function');
    expect(screen.getByTestId('setLocaleCodeOverride-type')).toHaveTextContent('function');
    expect(screen.getByTestId('setPopulateFromEnglish-type')).toHaveTextContent('function');
    expect(screen.getByTestId('setIsCreating-type')).toHaveTextContent('function');
    expect(screen.getByTestId('setProgress-type')).toHaveTextContent('function');
    expect(screen.getByTestId('setError-type')).toHaveTextContent('function');
    expect(screen.getByTestId('reset-type')).toHaveTextContent('function');
  });

  it('has exactly 18 properties in the context interface', () => {
    function TestContextProperties() {
      const context = useTranslationCreate();

      return (
        <div>
          <div data-testid="property-count">{Object.keys(context).length}</div>
        </div>
      );
    }

    render(
      <TestWrapper>
        <TranslationCreateProvider>
          <TestContextProperties />
        </TranslationCreateProvider>
      </TestWrapper>,
    );

    expect(screen.getByTestId('property-count')).toHaveTextContent('18');
  });
});
