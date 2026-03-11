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
import UserCreateProvider from '../UserCreateProvider';
import useUserCreate from '../useUserCreate';
import {UserCreateFlowStep} from '../../../models/user-create-flow';

// Test component to consume the context
function TestConsumer() {
  const context = useUserCreate();

  return (
    <div>
      <div data-testid="current-step">{context.currentStep}</div>
      <div data-testid="selected-schema">{context.selectedSchema?.name ?? 'null'}</div>
      <div data-testid="form-values">{JSON.stringify(context.formValues)}</div>
      <div data-testid="error">{context.error ?? 'null'}</div>

      <button type="button" onClick={() => context.setCurrentStep(UserCreateFlowStep.USER_DETAILS)}>
        Set Details Step
      </button>
      <button
        type="button"
        onClick={() => context.setSelectedSchema({id: 'schema-1', name: 'Employee', ouId: 'ou-1'})}
      >
        Set Schema
      </button>
      <button type="button" onClick={() => context.setFormValues({username: 'john'})}>
        Set Form Values
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

describe('UserCreateProvider', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('provides initial state values', async () => {
    await renderWithProviders(
      <UserCreateProvider>
        <TestConsumer />
      </UserCreateProvider>,
    );

    await expect.element(page.getByTestId('current-step')).toHaveTextContent(UserCreateFlowStep.USER_TYPE);
    await expect.element(page.getByTestId('selected-schema')).toHaveTextContent('null');
    await expect.element(page.getByTestId('form-values')).toHaveTextContent('{}');
    await expect.element(page.getByTestId('error')).toHaveTextContent('null');
  });

  it('updates current step when setCurrentStep is called', async () => {
    await renderWithProviders(
      <UserCreateProvider>
        <TestConsumer />
      </UserCreateProvider>,
    );

    await userEvent.click(page.getByText('Set Details Step'));

    await expect.element(page.getByTestId('current-step')).toHaveTextContent(UserCreateFlowStep.USER_DETAILS);
  });

  it('updates selected schema when setSelectedSchema is called', async () => {
    await renderWithProviders(
      <UserCreateProvider>
        <TestConsumer />
      </UserCreateProvider>,
    );

    await userEvent.click(page.getByText('Set Schema'));

    await expect.element(page.getByTestId('selected-schema')).toHaveTextContent('Employee');
  });

  it('updates form values when setFormValues is called', async () => {
    await renderWithProviders(
      <UserCreateProvider>
        <TestConsumer />
      </UserCreateProvider>,
    );

    await userEvent.click(page.getByText('Set Form Values'));

    await expect.element(page.getByTestId('form-values')).toHaveTextContent(JSON.stringify({username: 'john'}));
  });

  it('updates error when setError is called', async () => {
    await renderWithProviders(
      <UserCreateProvider>
        <TestConsumer />
      </UserCreateProvider>,
    );

    await userEvent.click(page.getByText('Set Error'));

    await expect.element(page.getByTestId('error')).toHaveTextContent('Test error');
  });

  it('resets all state when reset is called', async () => {
    await renderWithProviders(
      <UserCreateProvider>
        <TestConsumer />
      </UserCreateProvider>,
    );

    // Set some values
    await userEvent.click(page.getByText('Set Details Step'));
    await userEvent.click(page.getByText('Set Schema'));
    await userEvent.click(page.getByText('Set Form Values'));
    await userEvent.click(page.getByText('Set Error'));

    // Verify values are set
    await expect.element(page.getByTestId('current-step')).toHaveTextContent(UserCreateFlowStep.USER_DETAILS);
    await expect.element(page.getByTestId('selected-schema')).toHaveTextContent('Employee');
    await expect.element(page.getByTestId('error')).toHaveTextContent('Test error');

    // Reset
    await userEvent.click(page.getByText('Reset'));

    // Verify back to initial state
    await expect.element(page.getByTestId('current-step')).toHaveTextContent(UserCreateFlowStep.USER_TYPE);
    await expect.element(page.getByTestId('selected-schema')).toHaveTextContent('null');
    await expect.element(page.getByTestId('form-values')).toHaveTextContent('{}');
    await expect.element(page.getByTestId('error')).toHaveTextContent('null');
  });

  it('memoizes context value to prevent unnecessary re-renders', async () => {
    const renderSpy = vi.fn();

    function TestRenderer() {
      renderSpy();
      return <TestConsumer />;
    }

    const {rerender} = await renderWithProviders(
      <UserCreateProvider>
        <TestRenderer />
      </UserCreateProvider>,
    );

    expect(renderSpy).toHaveBeenCalledTimes(1);

    // Re-render with same props
    await rerender(
      <UserCreateProvider>
        <TestRenderer />
      </UserCreateProvider>,
    );

    // Should only render once more due to memoization
    expect(renderSpy).toHaveBeenCalledTimes(2);
  });

  it('throws error when useUserCreate is used outside provider', async () => {
    // Suppress console.error for this test since React will log the error
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => undefined);

    await expect(
      renderWithProviders(<TestConsumer />),
    ).rejects.toThrow('useUserCreate must be used within a UserCreateProvider');

    consoleSpy.mockRestore();
  });
});
