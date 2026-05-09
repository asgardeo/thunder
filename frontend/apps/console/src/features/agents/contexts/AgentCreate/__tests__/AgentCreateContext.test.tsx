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
import {useContext, useMemo} from 'react';
import {describe, expect, it, vi} from 'vitest';
import {AgentCreateFlowStep} from '../../../models/agent-create-flow';
import AgentCreateContext, {type AgentCreateContextType} from '../AgentCreateContext';

function TestConsumer() {
  const context = useContext(AgentCreateContext);

  if (!context) {
    return <div data-testid="context">undefined</div>;
  }

  return (
    <div>
      <div data-testid="context">defined</div>
      <div data-testid="context-type">{typeof context}</div>
      <div data-testid="current-step">{context.currentStep}</div>
      <div data-testid="agent-name">{context.agentName}</div>
      <div data-testid="selected-schema">{context.selectedSchema?.id ?? 'null'}</div>
    </div>
  );
}

function TestWithMockValue() {
  const mockContextValue: AgentCreateContextType = useMemo(
    () => ({
      currentStep: AgentCreateFlowStep.NAME,
      setCurrentStep: vi.fn(),
      selectedSchema: {id: 'schema-1', name: 'default', ouId: 'ou-1'},
      setSelectedSchema: vi.fn(),
      selectedOuId: null,
      setSelectedOuId: vi.fn(),
      agentName: 'Test Agent',
      setAgentName: vi.fn(),
      formValues: {},
      setFormValues: vi.fn(),
      selectedOwnerId: null,
      setSelectedOwnerId: vi.fn(),
      error: null,
      setError: vi.fn(),
      reset: vi.fn(),
    }),
    [],
  );

  return (
    <AgentCreateContext.Provider value={mockContextValue}>
      <TestConsumer />
    </AgentCreateContext.Provider>
  );
}

describe('AgentCreateContext', () => {
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

    expect(screen.getByTestId('current-step')).toHaveTextContent(AgentCreateFlowStep.NAME);
    expect(screen.getByTestId('agent-name')).toHaveTextContent('Test Agent');
    expect(screen.getByTestId('selected-schema')).toHaveTextContent('schema-1');
  });

  it('has the expected TypeScript interface shape', () => {
    const mockContext: AgentCreateContextType = {
      currentStep: AgentCreateFlowStep.NAME,
      setCurrentStep: () => null,
      selectedSchema: null,
      setSelectedSchema: () => null,
      selectedOuId: null,
      setSelectedOuId: () => null,
      agentName: '',
      setAgentName: () => null,
      formValues: {},
      setFormValues: () => null,
      selectedOwnerId: null,
      setSelectedOwnerId: () => null,
      error: null,
      setError: () => null,
      reset: () => null,
    };

    expect(mockContext).toBeDefined();
    expect(typeof mockContext.currentStep).toBe('string');
    expect(typeof mockContext.setCurrentStep).toBe('function');
    expect(typeof mockContext.reset).toBe('function');
    expect(mockContext.selectedSchema).toBeNull();
    expect(mockContext.error).toBeNull();
  });

  it('exports a context with the expected default value', () => {
    expect(AgentCreateContext).toBeDefined();
    expect(typeof AgentCreateContext).toBe('object');
  });
});
