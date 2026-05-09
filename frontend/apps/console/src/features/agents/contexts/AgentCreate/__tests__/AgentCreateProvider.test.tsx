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

import userEvent from '@testing-library/user-event';
import {render, screen} from '@thunderid/test-utils';
import {describe, expect, it} from 'vitest';
import {AgentCreateFlowStep} from '../../../models/agent-create-flow';
import AgentCreateProvider from '../AgentCreateProvider';
import useAgentCreate from '../useAgentCreate';

function TestConsumer() {
  const context = useAgentCreate();

  return (
    <div>
      <div data-testid="current-step">{context.currentStep}</div>
      <div data-testid="agent-name">{context.agentName}</div>
      <div data-testid="selected-schema">{context.selectedSchema?.id ?? 'null'}</div>
      <div data-testid="selected-ouid">{context.selectedOuId ?? 'null'}</div>
      <div data-testid="form-values">{JSON.stringify(context.formValues)}</div>
      <div data-testid="selected-owner">{context.selectedOwnerId ?? 'null'}</div>
      <div data-testid="error">{context.error ?? 'null'}</div>

      <button type="button" onClick={() => context.setCurrentStep(AgentCreateFlowStep.PROFILE)}>
        Set Profile Step
      </button>
      <button type="button" onClick={() => context.setAgentName('Test Agent')}>
        Set Agent Name
      </button>
      <button type="button" onClick={() => context.setSelectedSchema({id: 'schema-1', name: 'default', ouId: 'ou-1'})}>
        Set Schema
      </button>
      <button type="button" onClick={() => context.setSelectedOuId('ou-2')}>
        Set OU Id
      </button>
      <button type="button" onClick={() => context.setFormValues({email: 'a@b.com'})}>
        Set Form Values
      </button>
      <button type="button" onClick={() => context.setSelectedOwnerId('user-1')}>
        Set Owner
      </button>
      <button type="button" onClick={() => context.setError('Boom')}>
        Set Error
      </button>
      <button type="button" onClick={() => context.reset()}>
        Reset
      </button>
    </div>
  );
}

describe('AgentCreateProvider', () => {
  const renderInProvider = () =>
    render(
      <AgentCreateProvider>
        <TestConsumer />
      </AgentCreateProvider>,
    );

  it('provides initial state values', () => {
    renderInProvider();

    expect(screen.getByTestId('current-step')).toHaveTextContent(AgentCreateFlowStep.NAME);
    expect(screen.getByTestId('agent-name')).toHaveTextContent('');
    expect(screen.getByTestId('selected-schema')).toHaveTextContent('null');
    expect(screen.getByTestId('selected-ouid')).toHaveTextContent('null');
    expect(screen.getByTestId('form-values')).toHaveTextContent('{}');
    expect(screen.getByTestId('selected-owner')).toHaveTextContent('null');
    expect(screen.getByTestId('error')).toHaveTextContent('null');
  });

  it('updates currentStep via setCurrentStep', async () => {
    const user = userEvent.setup();
    renderInProvider();

    await user.click(screen.getByText('Set Profile Step'));
    expect(screen.getByTestId('current-step')).toHaveTextContent(AgentCreateFlowStep.PROFILE);
  });

  it('updates agentName via setAgentName', async () => {
    const user = userEvent.setup();
    renderInProvider();

    await user.click(screen.getByText('Set Agent Name'));
    expect(screen.getByTestId('agent-name')).toHaveTextContent('Test Agent');
  });

  it('updates selectedSchema via setSelectedSchema', async () => {
    const user = userEvent.setup();
    renderInProvider();

    await user.click(screen.getByText('Set Schema'));
    expect(screen.getByTestId('selected-schema')).toHaveTextContent('schema-1');
  });

  it('updates selectedOuId via setSelectedOuId', async () => {
    const user = userEvent.setup();
    renderInProvider();

    await user.click(screen.getByText('Set OU Id'));
    expect(screen.getByTestId('selected-ouid')).toHaveTextContent('ou-2');
  });

  it('updates formValues via setFormValues', async () => {
    const user = userEvent.setup();
    renderInProvider();

    await user.click(screen.getByText('Set Form Values'));
    expect(screen.getByTestId('form-values')).toHaveTextContent('a@b.com');
  });

  it('updates selectedOwnerId via setSelectedOwnerId', async () => {
    const user = userEvent.setup();
    renderInProvider();

    await user.click(screen.getByText('Set Owner'));
    expect(screen.getByTestId('selected-owner')).toHaveTextContent('user-1');
  });

  it('updates error via setError', async () => {
    const user = userEvent.setup();
    renderInProvider();

    await user.click(screen.getByText('Set Error'));
    expect(screen.getByTestId('error')).toHaveTextContent('Boom');
  });

  it('resets all state when reset() is called', async () => {
    const user = userEvent.setup();
    renderInProvider();

    await user.click(screen.getByText('Set Agent Name'));
    await user.click(screen.getByText('Set Schema'));
    await user.click(screen.getByText('Set Owner'));
    await user.click(screen.getByText('Set Error'));

    expect(screen.getByTestId('agent-name')).toHaveTextContent('Test Agent');
    expect(screen.getByTestId('selected-schema')).toHaveTextContent('schema-1');
    expect(screen.getByTestId('selected-owner')).toHaveTextContent('user-1');
    expect(screen.getByTestId('error')).toHaveTextContent('Boom');

    await user.click(screen.getByText('Reset'));

    expect(screen.getByTestId('current-step')).toHaveTextContent(AgentCreateFlowStep.NAME);
    expect(screen.getByTestId('agent-name')).toHaveTextContent('');
    expect(screen.getByTestId('selected-schema')).toHaveTextContent('null');
    expect(screen.getByTestId('selected-owner')).toHaveTextContent('null');
    expect(screen.getByTestId('error')).toHaveTextContent('null');
    expect(screen.getByTestId('form-values')).toHaveTextContent('{}');
  });
});
