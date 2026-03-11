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
import {page, userEvent} from 'vitest/browser';
import {renderWithProviders} from '@thunder/test-utils/browser';
import GroupCreateProvider from '../GroupCreateProvider';
import useGroupCreate from '../useGroupCreate';
import {GroupCreateFlowStep} from '../../../models/group-create-flow';

function TestConsumer() {
  const context = useGroupCreate();

  return (
    <div>
      <div data-testid="current-step">{context.currentStep}</div>
      <div data-testid="name">{context.name || 'empty'}</div>
      <div data-testid="description">{context.description || 'empty'}</div>
      <div data-testid="organization-unit-id">{context.organizationUnitId || 'empty'}</div>
      <div data-testid="error">{context.error ?? 'null'}</div>

      <button type="button" onClick={() => context.setCurrentStep(GroupCreateFlowStep.ORGANIZATION_UNIT)}>
        Set OU Step
      </button>
      <button type="button" onClick={() => context.setName('Test Group')}>
        Set Name
      </button>
      <button type="button" onClick={() => context.setDescription('A test description')}>
        Set Description
      </button>
      <button type="button" onClick={() => context.setOrganizationUnitId('ou-123')}>
        Set OU Id
      </button>
      <button type="button" onClick={() => context.setError('Test error')}>
        Set Error
      </button>
      <button type="button" onClick={() => context.reset()}>
        Reset
      </button>
    </div>
  );
}

describe('GroupCreateProvider', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('provides initial state values', async () => {
    await renderWithProviders(
      <GroupCreateProvider>
        <TestConsumer />
      </GroupCreateProvider>,
    );

    await expect.element(page.getByTestId('current-step')).toHaveTextContent(GroupCreateFlowStep.NAME);
    await expect.element(page.getByTestId('name')).toHaveTextContent('empty');
    await expect.element(page.getByTestId('description')).toHaveTextContent('empty');
    await expect.element(page.getByTestId('organization-unit-id')).toHaveTextContent('empty');
    await expect.element(page.getByTestId('error')).toHaveTextContent('null');
  });

  it('updates current step when setCurrentStep is called', async () => {
    await renderWithProviders(
      <GroupCreateProvider>
        <TestConsumer />
      </GroupCreateProvider>,
    );

    await userEvent.click(page.getByText('Set OU Step'));

    await expect.element(page.getByTestId('current-step')).toHaveTextContent(GroupCreateFlowStep.ORGANIZATION_UNIT);
  });

  it('updates name when setName is called', async () => {
    await renderWithProviders(
      <GroupCreateProvider>
        <TestConsumer />
      </GroupCreateProvider>,
    );

    await userEvent.click(page.getByText('Set Name'));

    await expect.element(page.getByTestId('name')).toHaveTextContent('Test Group');
  });

  it('updates description when setDescription is called', async () => {
    await renderWithProviders(
      <GroupCreateProvider>
        <TestConsumer />
      </GroupCreateProvider>,
    );

    await userEvent.click(page.getByText('Set Description'));

    await expect.element(page.getByTestId('description')).toHaveTextContent('A test description');
  });

  it('updates organizationUnitId when setOrganizationUnitId is called', async () => {
    await renderWithProviders(
      <GroupCreateProvider>
        <TestConsumer />
      </GroupCreateProvider>,
    );

    await userEvent.click(page.getByText('Set OU Id'));

    await expect.element(page.getByTestId('organization-unit-id')).toHaveTextContent('ou-123');
  });

  it('updates error when setError is called', async () => {
    await renderWithProviders(
      <GroupCreateProvider>
        <TestConsumer />
      </GroupCreateProvider>,
    );

    await userEvent.click(page.getByText('Set Error'));

    await expect.element(page.getByTestId('error')).toHaveTextContent('Test error');
  });

  it('resets all state when reset is called', async () => {
    await renderWithProviders(
      <GroupCreateProvider>
        <TestConsumer />
      </GroupCreateProvider>,
    );

    // Set some values
    await userEvent.click(page.getByText('Set OU Step'));
    await userEvent.click(page.getByText('Set Name'));
    await userEvent.click(page.getByText('Set Description'));
    await userEvent.click(page.getByText('Set OU Id'));
    await userEvent.click(page.getByText('Set Error'));

    // Verify values are set
    await expect.element(page.getByTestId('current-step')).toHaveTextContent(GroupCreateFlowStep.ORGANIZATION_UNIT);
    await expect.element(page.getByTestId('name')).toHaveTextContent('Test Group');
    await expect.element(page.getByTestId('description')).toHaveTextContent('A test description');
    await expect.element(page.getByTestId('organization-unit-id')).toHaveTextContent('ou-123');
    await expect.element(page.getByTestId('error')).toHaveTextContent('Test error');

    // Reset
    await userEvent.click(page.getByText('Reset'));

    // Verify back to initial state
    await expect.element(page.getByTestId('current-step')).toHaveTextContent(GroupCreateFlowStep.NAME);
    await expect.element(page.getByTestId('name')).toHaveTextContent('empty');
    await expect.element(page.getByTestId('description')).toHaveTextContent('empty');
    await expect.element(page.getByTestId('organization-unit-id')).toHaveTextContent('empty');
    await expect.element(page.getByTestId('error')).toHaveTextContent('null');
  });

  it('memoizes context value to prevent unnecessary re-renders', async () => {
    const renderSpy = vi.fn();

    function TestRenderer() {
      renderSpy();
      return <TestConsumer />;
    }

    const {rerender} = await renderWithProviders(
      <GroupCreateProvider>
        <TestRenderer />
      </GroupCreateProvider>,
    );

    expect(renderSpy).toHaveBeenCalledTimes(1);

    await rerender(
      <GroupCreateProvider>
        <TestRenderer />
      </GroupCreateProvider>,
    );

    expect(renderSpy).toHaveBeenCalledTimes(2);
  });

  it('throws error when useGroupCreate is used outside provider', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => undefined);

    await expect(renderWithProviders(<TestConsumer />)).rejects.toThrow(
      'useGroupCreate must be used within a GroupCreateProvider',
    );

    consoleSpy.mockRestore();
  });
});
