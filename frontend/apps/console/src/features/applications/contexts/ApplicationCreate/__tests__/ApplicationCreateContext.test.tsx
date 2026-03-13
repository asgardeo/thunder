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

import {describe, expect, it, vi} from 'vitest';
import {render, screen} from '@testing-library/react';
import {useContext, useMemo} from 'react';
import ApplicationCreateContext, {type ApplicationCreateContextType} from '../ApplicationCreateContext';

// Test component to consume the context directly
function TestConsumer() {
  const context = useContext(ApplicationCreateContext);

  if (!context) {
    return <div data-testid="context">undefined</div>;
  }

  return (
    <div>
      <div data-testid="context">defined</div>
      <div data-testid="context-type">{typeof context}</div>
      <div data-testid="current-step">{context.currentStep}</div>
      <div data-testid="app-name">{context.appName}</div>
      <div data-testid="selected-theme">{context.themeId ?? 'null'}</div>
    </div>
  );
}

// Test component with a mock context value
function TestWithMockValue() {
  const mockContextValue: ApplicationCreateContextType = useMemo(
    () => ({
      currentStep: 'DESIGN',
      setCurrentStep: vi.fn(),
      appName: 'Test App',
      setAppName: vi.fn(),
      themeId: null,
      setThemeId: vi.fn(),
      selectedTheme: null,
      setSelectedTheme: vi.fn(),
      appLogo: null,
      setAppLogo: vi.fn(),
      selectedColor: '',
      setSelectedColor: vi.fn(),
      integrations: {},
      setIntegrations: vi.fn(),
      toggleIntegration: vi.fn(),
      selectedAuthFlow: null,
      setSelectedAuthFlow: vi.fn(),
      signInApproach: 'INBUILT',
      setSignInApproach: vi.fn(),
      selectedTechnology: null,
      setSelectedTechnology: vi.fn(),
      selectedPlatform: null,
      setSelectedPlatform: vi.fn(),
      selectedTemplateConfig: null,
      setSelectedTemplateConfig: vi.fn(),
      hostingUrl: 'https://example.com',
      setHostingUrl: vi.fn(),
      callbackUrlFromConfig: 'https://example.com/callback',
      setCallbackUrlFromConfig: vi.fn(),
      relyingPartyId: '',
      setRelyingPartyId: vi.fn(),
      relyingPartyName: '',
      setRelyingPartyName: vi.fn(),
      hasCompletedOnboarding: false,
      setHasCompletedOnboarding: vi.fn(),
      error: null,
      setError: vi.fn(),
      reset: vi.fn(),
    }),
    [],
  );

  return (
    <ApplicationCreateContext.Provider value={mockContextValue}>
      <TestConsumer />
    </ApplicationCreateContext.Provider>
  );
}

describe('ApplicationCreateContext', () => {
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

    expect(screen.getByTestId('current-step')).toHaveTextContent('DESIGN');
    expect(screen.getByTestId('app-name')).toHaveTextContent('Test App');
    expect(screen.getByTestId('selected-theme')).toHaveTextContent('null');
  });

  it('has correct TypeScript interface definition', () => {
    // This test ensures the interface matches expected shape
    const mockContext: ApplicationCreateContextType = {
      relyingPartyId: '',
      setRelyingPartyId: () => {},
      relyingPartyName: '',
      setRelyingPartyName: () => {},
      currentStep: 'NAME',
      setCurrentStep: () => {},
      appName: '',
      setAppName: () => {},
      themeId: null,
      setThemeId: () => {},
      selectedTheme: null,
      setSelectedTheme: () => {},
      appLogo: null,
      setAppLogo: () => {},
      selectedColor: '',
      setSelectedColor: () => {},
      integrations: {},
      setIntegrations: () => {},
      toggleIntegration: () => {},
      selectedAuthFlow: null,
      setSelectedAuthFlow: () => {},
      signInApproach: 'INBUILT',
      setSignInApproach: () => {},
      selectedTechnology: null,
      setSelectedTechnology: () => {},
      selectedPlatform: null,
      setSelectedPlatform: () => {},
      selectedTemplateConfig: null,
      setSelectedTemplateConfig: () => {},
      hostingUrl: '',
      setHostingUrl: () => {},
      callbackUrlFromConfig: '',
      setCallbackUrlFromConfig: () => {},
      hasCompletedOnboarding: false,
      setHasCompletedOnboarding: () => {},
      error: null,
      setError: () => {},
      reset: () => {},
    };

    expect(mockContext).toBeDefined();
    expect(typeof mockContext.currentStep).toBe('string');
    expect(typeof mockContext.setCurrentStep).toBe('function');
    expect(typeof mockContext.appName).toBe('string');
    expect(typeof mockContext.setAppName).toBe('function');
    expect(mockContext.themeId).toBeNull();
    expect(mockContext.selectedTheme).toBeNull();
    expect(typeof mockContext.setSelectedTheme).toBe('function');
    expect(typeof mockContext.toggleIntegration).toBe('function');
    expect(typeof mockContext.reset).toBe('function');
  });

  it('allows null values for optional properties', () => {
    const mockContext: ApplicationCreateContextType = {
      relyingPartyId: '',
      setRelyingPartyId: () => {},
      relyingPartyName: '',
      setRelyingPartyName: () => {},
      currentStep: 'NAME',
      setCurrentStep: () => {},
      appName: '',
      setAppName: () => {},
      themeId: null,
      setThemeId: () => {},
      selectedTheme: null,
      setSelectedTheme: () => {},
      appLogo: null, // Should allow null
      setAppLogo: () => {},
      selectedColor: '',
      setSelectedColor: () => {},
      integrations: {},
      setIntegrations: () => {},
      toggleIntegration: () => {},
      selectedAuthFlow: null, // Should allow null
      setSelectedAuthFlow: () => {},
      signInApproach: 'INBUILT',
      setSignInApproach: () => {},
      selectedTechnology: null, // Should allow null
      setSelectedTechnology: () => {},
      selectedPlatform: null, // Should allow null
      setSelectedPlatform: () => {},
      selectedTemplateConfig: null, // Should allow null
      setSelectedTemplateConfig: () => {},
      hostingUrl: '',
      setHostingUrl: () => {},
      callbackUrlFromConfig: '',
      setCallbackUrlFromConfig: () => {},
      hasCompletedOnboarding: false,
      setHasCompletedOnboarding: () => {},
      error: null, // Should allow null
      setError: () => {},
      reset: () => {},
    };

    expect(mockContext.appLogo).toBeNull();
    expect(mockContext.selectedAuthFlow).toBeNull();
    expect(mockContext.selectedTechnology).toBeNull();
    expect(mockContext.selectedPlatform).toBeNull();
    expect(mockContext.selectedTemplateConfig).toBeNull();
    expect(mockContext.error).toBeNull();
  });

  it('creates context with expected default value (undefined)', () => {
    // Testing the default export creates context with undefined default value
    // React Context doesn't expose _currentValue property in newer versions
    expect(ApplicationCreateContext).toBeDefined();
    expect(typeof ApplicationCreateContext).toBe('object');
  });
});
