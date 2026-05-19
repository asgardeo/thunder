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

import {render, screen} from '@thunderid/test-utils';
import {describe, expect, it, vi} from 'vitest';
import AgentCreateProvider from '../AgentCreateProvider';
import useAgentCreate from '../useAgentCreate';

function TestConsumer() {
  const context = useAgentCreate();
  return <div data-testid="context-available">{typeof context}</div>;
}

function TestConsumerWithoutProvider() {
  const context = useAgentCreate();
  return <div data-testid="context">{JSON.stringify(context)}</div>;
}

describe('useAgentCreate', () => {
  it('returns context when used within AgentCreateProvider', () => {
    render(
      <AgentCreateProvider>
        <TestConsumer />
      </AgentCreateProvider>,
    );

    expect(screen.getByTestId('context-available')).toHaveTextContent('object');
  });

  it('throws error when used outside AgentCreateProvider', () => {
    const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => null);

    expect(() => {
      render(<TestConsumerWithoutProvider />);
    }).toThrow('useAgentCreate must be used within AgentCreateProvider');

    errorSpy.mockRestore();
  });

  it('provides all required context properties', () => {
    function TestContextProperties() {
      const context = useAgentCreate();

      const requiredProperties = [
        'currentStep',
        'setCurrentStep',
        'selectedSchema',
        'setSelectedSchema',
        'selectedOuId',
        'setSelectedOuId',
        'agentName',
        'setAgentName',
        'formValues',
        'setFormValues',
        'selectedOwnerId',
        'setSelectedOwnerId',
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
      <AgentCreateProvider>
        <TestContextProperties />
      </AgentCreateProvider>,
    );

    expect(screen.getByTestId('has-all-properties')).toHaveTextContent('true');
    expect(screen.getByTestId('missing-properties')).toHaveTextContent('[]');
  });

  it('returns the same context reference across multiple hook calls', () => {
    function TestMultipleHookCalls() {
      const context1 = useAgentCreate();
      const context2 = useAgentCreate();

      return <div data-testid="same-reference">{(context1 === context2).toString()}</div>;
    }

    render(
      <AgentCreateProvider>
        <TestMultipleHookCalls />
      </AgentCreateProvider>,
    );

    expect(screen.getByTestId('same-reference')).toHaveTextContent('true');
  });

  it('provides functions with correct types', () => {
    function TestFunctionTypes() {
      const {
        setCurrentStep,
        setSelectedSchema,
        setSelectedOuId,
        setAgentName,
        setFormValues,
        setSelectedOwnerId,
        setError,
        reset,
      } = useAgentCreate();

      return (
        <div>
          <div data-testid="setCurrentStep-type">{typeof setCurrentStep}</div>
          <div data-testid="setSelectedSchema-type">{typeof setSelectedSchema}</div>
          <div data-testid="setSelectedOuId-type">{typeof setSelectedOuId}</div>
          <div data-testid="setAgentName-type">{typeof setAgentName}</div>
          <div data-testid="setFormValues-type">{typeof setFormValues}</div>
          <div data-testid="setSelectedOwnerId-type">{typeof setSelectedOwnerId}</div>
          <div data-testid="setError-type">{typeof setError}</div>
          <div data-testid="reset-type">{typeof reset}</div>
        </div>
      );
    }

    render(
      <AgentCreateProvider>
        <TestFunctionTypes />
      </AgentCreateProvider>,
    );

    expect(screen.getByTestId('setCurrentStep-type')).toHaveTextContent('function');
    expect(screen.getByTestId('setSelectedSchema-type')).toHaveTextContent('function');
    expect(screen.getByTestId('setSelectedOuId-type')).toHaveTextContent('function');
    expect(screen.getByTestId('setAgentName-type')).toHaveTextContent('function');
    expect(screen.getByTestId('setFormValues-type')).toHaveTextContent('function');
    expect(screen.getByTestId('setSelectedOwnerId-type')).toHaveTextContent('function');
    expect(screen.getByTestId('setError-type')).toHaveTextContent('function');
    expect(screen.getByTestId('reset-type')).toHaveTextContent('function');
  });

  it('throws a descriptive error message when used outside provider', () => {
    const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => null);

    let thrownError: Error | null = null;
    try {
      render(<TestConsumerWithoutProvider />);
    } catch (error) {
      thrownError = error as Error;
    }

    expect(thrownError).toBeInstanceOf(Error);
    expect(thrownError?.message).toBe('useAgentCreate must be used within AgentCreateProvider');

    errorSpy.mockRestore();
  });
});
