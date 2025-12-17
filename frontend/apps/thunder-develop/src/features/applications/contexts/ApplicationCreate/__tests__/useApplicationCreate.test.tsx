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
import useApplicationCreate from '../useApplicationCreate';
import ApplicationCreateProvider from '../ApplicationCreateProvider';

// Test component to consume the hook
function TestConsumer() {
  const context = useApplicationCreate();

  return <div data-testid="context-available">{typeof context}</div>;
}

// Test component without provider
function TestConsumerWithoutProvider() {
  const context = useApplicationCreate();

  return <div data-testid="context">{JSON.stringify(context)}</div>;
}

describe('useApplicationCreate', () => {
  it('returns context when used within ApplicationCreateProvider', () => {
    render(
      <ApplicationCreateProvider>
        <TestConsumer />
      </ApplicationCreateProvider>,
    );

    expect(screen.getByTestId('context-available')).toHaveTextContent('object');
  });

  it('throws error when used outside ApplicationCreateProvider', () => {
    // Suppress error output in tests
    const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    expect(() => {
      render(<TestConsumerWithoutProvider />);
    }).toThrow('useApplicationCreate must be used within ApplicationCreateProvider');

    // Restore console.error
    errorSpy.mockRestore();
  });

  it('provides all required context properties', () => {
    function TestContextProperties() {
      const context = useApplicationCreate();

      const requiredProperties = [
        'currentStep',
        'setCurrentStep',
        'appName',
        'setAppName',
        'selectedColor',
        'setSelectedColor',
        'appLogo',
        'setAppLogo',
        'integrations',
        'setIntegrations',
        'toggleIntegration',
        'signInApproach',
        'setSignInApproach',
        'selectedTechnology',
        'setSelectedTechnology',
        'selectedPlatform',
        'setSelectedPlatform',
        'selectedTemplateConfig',
        'setSelectedTemplateConfig',
        'hostingUrl',
        'setHostingUrl',
        'callbackUrlFromConfig',
        'setCallbackUrlFromConfig',
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
      <ApplicationCreateProvider>
        <TestContextProperties />
      </ApplicationCreateProvider>,
    );

    expect(screen.getByTestId('has-all-properties')).toHaveTextContent('true');
    expect(screen.getByTestId('missing-properties')).toHaveTextContent('[]');
  });
});
