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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render} from '@thunder/test-utils/browser';
import {page} from 'vitest/browser';
import {useContext} from 'react';
import ValidationProvider from '../ValidationProvider';
import {ValidationContext} from '../ValidationContext';
import Notification, {NotificationType} from '../../models/notification';

// Mock useFlowBuilderCore
const mockSetIsOpenResourcePropertiesPanel = vi.fn();
const mockRegisterCloseValidationPanel = vi.fn();

vi.mock('../../hooks/useFlowBuilderCore', () => ({
  default: () => ({
    setIsOpenResourcePropertiesPanel: mockSetIsOpenResourcePropertiesPanel,
    registerCloseValidationPanel: mockRegisterCloseValidationPanel,
  }),
}));

// Test component to access context
function TestConsumer() {
  const context = useContext(ValidationContext);
  return (
    <div>
      <span data-testid="is-valid">{context.isValid.toString()}</span>
      <span data-testid="notifications-count">{context.notifications.length}</span>
      <span data-testid="open-validation-panel">{context.openValidationPanel?.toString()}</span>
      <span data-testid="current-tab">{context.currentActiveTab}</span>
      <button
        type="button"
        data-testid="add-notification"
        onClick={() => {
          const notification = new Notification('test-id', 'Test Error', NotificationType.ERROR);
          context.addNotification?.(notification);
        }}
      >
        Add Notification
      </button>
      <button
        type="button"
        data-testid="add-warning"
        onClick={() => {
          const notification = new Notification('warning-id', 'Test Warning', NotificationType.WARNING);
          context.addNotification?.(notification);
        }}
      >
        Add Warning
      </button>
      <button type="button" data-testid="remove-notification" onClick={() => context.removeNotification?.('test-id')}>
        Remove Notification
      </button>
      <button type="button" data-testid="open-panel" onClick={() => context.setOpenValidationPanel?.(true)}>
        Open Panel
      </button>
      <button type="button" data-testid="close-panel" onClick={() => context.setOpenValidationPanel?.(false)}>
        Close Panel
      </button>
      <button type="button" data-testid="set-tab" onClick={() => context.setCurrentActiveTab?.(2)}>
        Set Tab
      </button>
    </div>
  );
}

describe('ValidationProvider', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Initial State', () => {
    it('should provide initial valid state as true', async () => {
      await render(
        <ValidationProvider>
          <TestConsumer />
        </ValidationProvider>,
      );

      await expect.element(page.getByTestId('is-valid')).toHaveTextContent('true');
    });

    it('should provide empty notifications list initially', async () => {
      await render(
        <ValidationProvider>
          <TestConsumer />
        </ValidationProvider>,
      );

      await expect.element(page.getByTestId('notifications-count')).toHaveTextContent('0');
    });

    it('should have validation panel closed initially', async () => {
      await render(
        <ValidationProvider>
          <TestConsumer />
        </ValidationProvider>,
      );

      await expect.element(page.getByTestId('open-validation-panel')).toHaveTextContent('false');
    });

    it('should have current active tab set to 0 initially', async () => {
      await render(
        <ValidationProvider>
          <TestConsumer />
        </ValidationProvider>,
      );

      await expect.element(page.getByTestId('current-tab')).toHaveTextContent('0');
    });
  });

  describe('Validation Config', () => {
    it('should use default validation config when not provided', async () => {
      await render(
        <ValidationProvider>
          <TestConsumer />
        </ValidationProvider>,
      );

      await expect.element(page.getByTestId('is-valid')).toBeInTheDocument();
    });

    it('should accept custom validation config', async () => {
      await render(
        <ValidationProvider validationConfig={{isOTPValidationEnabled: true, isRecoveryFactorValidationEnabled: true}}>
          <TestConsumer />
        </ValidationProvider>,
      );

      await expect.element(page.getByTestId('is-valid')).toBeInTheDocument();
    });
  });

  describe('Notification Management', () => {
    it('should add notification', async () => {
      await render(
        <ValidationProvider>
          <TestConsumer />
        </ValidationProvider>,
      );

      await page.getByTestId('add-notification').click();;

      await expect.element(page.getByTestId('notifications-count')).toHaveTextContent('1');
    });

    it('should set isValid to false when error notification is added', async () => {
      await render(
        <ValidationProvider>
          <TestConsumer />
        </ValidationProvider>,
      );

      await expect.element(page.getByTestId('is-valid')).toHaveTextContent('true');

      await page.getByTestId('add-notification').click();;

      await expect.element(page.getByTestId('is-valid')).toHaveTextContent('false');
    });

    it('should remain valid when only warning notification is added', async () => {
      await render(
        <ValidationProvider>
          <TestConsumer />
        </ValidationProvider>,
      );

      await page.getByTestId('add-warning').click();;

      await expect.element(page.getByTestId('is-valid')).toHaveTextContent('true');
    });

    it('should remove notification', async () => {
      await render(
        <ValidationProvider>
          <TestConsumer />
        </ValidationProvider>,
      );

      await page.getByTestId('add-notification').click();;

      await expect.element(page.getByTestId('notifications-count')).toHaveTextContent('1');

      await page.getByTestId('remove-notification').click();;

      await expect.element(page.getByTestId('notifications-count')).toHaveTextContent('0');
    });

    it('should restore isValid to true when error notification is removed', async () => {
      await render(
        <ValidationProvider>
          <TestConsumer />
        </ValidationProvider>,
      );

      await page.getByTestId('add-notification').click();;

      await expect.element(page.getByTestId('is-valid')).toHaveTextContent('false');

      await page.getByTestId('remove-notification').click();;

      await expect.element(page.getByTestId('is-valid')).toHaveTextContent('true');
    });
  });

  describe('Validation Panel', () => {
    it('should open validation panel', async () => {
      await render(
        <ValidationProvider>
          <TestConsumer />
        </ValidationProvider>,
      );

      await page.getByTestId('open-panel').click();;

      await expect.element(page.getByTestId('open-validation-panel')).toHaveTextContent('true');
    });

    it('should close resource properties panel when opening validation panel', async () => {
      await render(
        <ValidationProvider>
          <TestConsumer />
        </ValidationProvider>,
      );

      await page.getByTestId('open-panel').click();;

      expect(mockSetIsOpenResourcePropertiesPanel).toHaveBeenCalledWith(false);
    });

    it('should close validation panel', async () => {
      await render(
        <ValidationProvider>
          <TestConsumer />
        </ValidationProvider>,
      );

      await page.getByTestId('open-panel').click();;

      await page.getByTestId('close-panel').click();;

      await expect.element(page.getByTestId('open-validation-panel')).toHaveTextContent('false');
    });
  });

  describe('Tab Management', () => {
    it('should update current active tab', async () => {
      await render(
        <ValidationProvider>
          <TestConsumer />
        </ValidationProvider>,
      );

      await page.getByTestId('set-tab').click();;

      await expect.element(page.getByTestId('current-tab')).toHaveTextContent('2');
    });
  });

  describe('Registration with FlowBuilderCore', () => {
    it('should register close validation panel callback', async () => {
      await render(
        <ValidationProvider>
          <TestConsumer />
        </ValidationProvider>,
      );

      expect(mockRegisterCloseValidationPanel).toHaveBeenCalledWith(expect.any(Function));
    });

    it('should unregister callback on unmount', async () => {
      const {unmount} = await render(
        <ValidationProvider>
          <TestConsumer />
        </ValidationProvider>,
      );

      await unmount();

      // The cleanup function should register an empty callback
      expect(mockRegisterCloseValidationPanel).toHaveBeenLastCalledWith(expect.any(Function));
    });
  });

  describe('Children Rendering', () => {
    it('should render children', async () => {
      await render(
        <ValidationProvider>
          <div data-testid="child">Child Content</div>
        </ValidationProvider>,
      );

      await expect.element(page.getByTestId('child')).toHaveTextContent('Child Content');
    });

    it('should render without children', async () => {
      const {container} = await render(<ValidationProvider />);

      expect(container).toBeInTheDocument();
    });
  });
});
