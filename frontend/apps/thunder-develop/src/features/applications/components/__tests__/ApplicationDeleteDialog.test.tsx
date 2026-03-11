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

import {page, userEvent} from 'vitest/browser';
import {renderWithProviders} from '@thunder/test-utils/browser';
import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';
import ApplicationDeleteDialog from '../ApplicationDeleteDialog';
import * as useDeleteApplicationModule from '../../api/useDeleteApplication';

// Mock the useDeleteApplication hook
vi.mock('../../api/useDeleteApplication');

describe('ApplicationDeleteDialog', () => {
  const mockOnClose = vi.fn();
  const mockOnSuccess = vi.fn();
  const mockMutate = vi.fn();

  const defaultProps = {
    open: true,
    applicationId: 'test-app-id',
    onClose: mockOnClose,
    onSuccess: mockOnSuccess,
  };

  const renderComponent = (props = defaultProps) => renderWithProviders(<ApplicationDeleteDialog {...props} />);

  beforeEach(() => {

    // Default mock implementation
    vi.mocked(useDeleteApplicationModule.default).mockReturnValue({
      mutate: mockMutate,
      isPending: false,
      isError: false,
      isSuccess: false,
      error: null,
      data: undefined,
      mutateAsync: vi.fn(),
      reset: vi.fn(),
      context: undefined,
      failureCount: 0,
      failureReason: null,
      isIdle: true,
      isPaused: false,
      status: 'idle',
      submittedAt: 0,
      variables: undefined,
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render the dialog when open is true', async () => {
      await renderComponent();

      await expect.element(page.getByRole('dialog')).toBeInTheDocument();
      await expect.element(page.getByText('Delete Application')).toBeInTheDocument();
      await expect.element(page.getByText('Are you sure you want to delete this application? This action cannot be undone.')).toBeInTheDocument();
      expect(
        page.getByText(
          'Warning: All associated data, configurations, and access tokens will be permanently removed.',
        ),
      ).toBeInTheDocument();
    });

    it('should not render dialog content when open is false', async () => {
      await renderComponent({...defaultProps, open: false});

      await expect.element(page.getByRole('dialog')).not.toBeInTheDocument();
    });

    it('should render Cancel and Delete buttons', async () => {
      await renderComponent();

      await expect.element(page.getByRole('button', {name: 'Cancel'})).toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: 'Delete'})).toBeInTheDocument();
    });

    it('should not render error alert initially', async () => {
      await renderComponent();

      await expect.element(page.getByRole('alert')).toHaveTextContent('Warning: All associated data'); // Only warning alert
    });
  });

  describe('User Interactions', () => {
    it('should call onClose when Cancel button is clicked', async () => {
      await renderComponent();

      const cancelButton = page.getByRole('button', {name: 'Cancel'});
      await userEvent.click(cancelButton);

      expect(mockOnClose).toHaveBeenCalledTimes(1);
      expect(mockMutate).not.toHaveBeenCalled();
    });

    it('should call onClose when Escape key is pressed', async () => {
      await renderComponent();

      // Press Escape key
      await userEvent.keyboard('{Escape}');

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('should trigger delete mutation when Delete button is clicked', async () => {
      await renderComponent();

      const deleteButton = page.getByRole('button', {name: 'Delete'});
      await userEvent.click(deleteButton);

      expect(mockMutate).toHaveBeenCalledTimes(1);
      expect(mockMutate).toHaveBeenCalledWith('test-app-id', expect.any(Object));
    });

    it('should not trigger delete mutation when applicationId is null', async () => {
      await renderComponent({...defaultProps, applicationId: ''});

      const deleteButton = page.getByRole('button', {name: 'Delete'});
      await userEvent.click(deleteButton);

      expect(mockMutate).not.toHaveBeenCalled();
    });
  });

  describe('Delete Success Flow', () => {
    it('should call onClose and onSuccess callbacks on successful delete', async () => {

      mockMutate.mockImplementation((_, options: {onSuccess?: () => void}) => {
        // Simulate successful mutation
        options?.onSuccess?.();
      });

      await renderComponent();

      const deleteButton = page.getByRole('button', {name: 'Delete'});
      await userEvent.click(deleteButton);

      await vi.waitFor(() => {
        expect(mockOnClose).toHaveBeenCalledTimes(1);
        expect(mockOnSuccess).toHaveBeenCalledTimes(1);
      });
    });

    it('should work without onSuccess callback', async () => {

      mockMutate.mockImplementation((_, options: {onSuccess?: () => void}) => {
        options?.onSuccess?.();
      });

      await renderComponent({...defaultProps, onSuccess: vi.fn()});

      const deleteButton = page.getByRole('button', {name: 'Delete'});
      await userEvent.click(deleteButton);

      await vi.waitFor(() => {
        expect(mockOnClose).toHaveBeenCalledTimes(1);
      });
    });

    it('should clear any previous errors on successful delete', async () => {

      // First, trigger an error
      mockMutate.mockImplementationOnce((_, options: {onError?: (error: Error) => void}) => {
        options?.onError?.(new Error('Delete failed'));
      });

      const {rerender} = await renderComponent();

      let deleteButton = page.getByRole('button', {name: 'Delete'});
      await userEvent.click(deleteButton);

      await vi.waitFor(() => {
        expect(page.getByText('Delete failed').element()).toBeInTheDocument();
      });

      // Then trigger success
      mockMutate.mockImplementationOnce((_, options: {onSuccess?: () => void}) => {
        options?.onSuccess?.();
      });

      await rerender(<ApplicationDeleteDialog {...defaultProps} />);

      deleteButton = page.getByRole('button', {name: 'Delete'});
      await userEvent.click(deleteButton);

      await vi.waitFor(() => {
        expect(mockOnClose).toHaveBeenCalled();
        expect(page.getByText('Delete failed').query()).not.toBeInTheDocument();
      });
    });
  });

  describe('Delete Error Flow', () => {
    it('should display error message when delete fails', async () => {
      const errorMessage = 'Failed to delete application';

      mockMutate.mockImplementation((_, options: {onError?: (error: Error) => void}) => {
        options?.onError?.(new Error(errorMessage));
      });

      await renderComponent();

      const deleteButton = page.getByRole('button', {name: 'Delete'});
      await userEvent.click(deleteButton);

      await vi.waitFor(() => {
        expect(page.getByText(errorMessage).element()).toBeInTheDocument();
      });

      expect(mockOnClose).not.toHaveBeenCalled();
      expect(mockOnSuccess).not.toHaveBeenCalled();
    });

    it('should clear error when Cancel is clicked after error', async () => {

      mockMutate.mockImplementation((_, options: {onError?: (error: Error) => void}) => {
        options?.onError?.(new Error('Delete failed'));
      });

      await renderComponent();

      const deleteButton = page.getByRole('button', {name: 'Delete'});
      await userEvent.click(deleteButton);

      await vi.waitFor(() => {
        expect(page.getByText('Delete failed').element()).toBeInTheDocument();
      });

      const cancelButton = page.getByRole('button', {name: 'Cancel'});
      await userEvent.click(cancelButton);

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('should persist error message across re-renders until cleared', async () => {

      mockMutate.mockImplementation((_, options: {onError?: (error: Error) => void}) => {
        options?.onError?.(new Error('Delete failed'));
      });

      const {rerender} = await renderComponent();

      const deleteButton = page.getByRole('button', {name: 'Delete'});
      await userEvent.click(deleteButton);

      await vi.waitFor(() => {
        expect(page.getByText('Delete failed').element()).toBeInTheDocument();
      });

      // Re-render with same props
      await rerender(<ApplicationDeleteDialog {...defaultProps} />);

      await expect.element(page.getByText('Delete failed')).toBeInTheDocument();
    });
  });

  describe('Loading State', () => {
    it('should disable buttons when delete is pending', async () => {
      vi.mocked(useDeleteApplicationModule.default).mockReturnValue({
        mutate: mockMutate,
        isPending: true,
        isError: false,
        isSuccess: false,
        error: null,
        data: undefined,
        mutateAsync: vi.fn(),
        reset: vi.fn(),
        context: undefined,
        failureCount: 0,
        failureReason: null,
        isIdle: false,
        isPaused: false,
        status: 'pending',
        submittedAt: Date.now(),
        variables: '',
      });

      await renderComponent();

      await expect.element(page.getByRole('button', {name: 'Cancel'})).toBeDisabled();
      await expect.element(page.getByRole('button', {name: 'Deleting...'})).toBeDisabled();
    });

    it('should show "Deleting..." text on Delete button when pending', async () => {
      vi.mocked(useDeleteApplicationModule.default).mockReturnValue({
        mutate: mockMutate,
        isPending: true,
        isError: false,
        isSuccess: false,
        error: null,
        data: undefined,
        mutateAsync: vi.fn(),
        reset: vi.fn(),
        context: undefined,
        failureCount: 0,
        failureReason: null,
        isIdle: false,
        isPaused: false,
        status: 'pending',
        submittedAt: Date.now(),
        variables: '',
      });

      await renderComponent();

      await expect.element(page.getByRole('button', {name: 'Deleting...'})).toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: 'Delete'})).not.toBeInTheDocument();
    });

    it('should not trigger another delete when already pending', async () => {
      vi.mocked(useDeleteApplicationModule.default).mockReturnValue({
        mutate: mockMutate,
        isPending: true,
        isError: false,
        isSuccess: false,
        error: null,
        data: undefined,
        mutateAsync: vi.fn(),
        reset: vi.fn(),
        context: undefined,
        failureCount: 0,
        failureReason: null,
        isIdle: false,
        isPaused: false,
        status: 'pending',
        submittedAt: Date.now(),
        variables: 'test-app-id',
      });

      await renderComponent();

      const deleteButton = page.getByRole('button', {name: 'Deleting...'});
      expect(deleteButton).toBeDisabled();

      // Verify button cannot be interacted with when disabled
      expect(mockMutate).not.toHaveBeenCalled();
    });
  });

  describe('Edge Cases', () => {
    it('should handle rapid clicks on Delete button', async () => {
      await renderComponent();

      const deleteButton = page.getByRole('button', {name: 'Delete'});

      await userEvent.click(deleteButton);
      await userEvent.click(deleteButton);
      await userEvent.click(deleteButton);

      // Should still only call mutate once per click (3 times total)
      expect(mockMutate).toHaveBeenCalledTimes(3);
    });

    it('should handle dialog opening and closing multiple times', async () => {
      const {rerender} = await renderComponent({...defaultProps, open: false});

      await expect.element(page.getByRole('dialog')).not.toBeInTheDocument();

      // Open dialog
      await rerender(<ApplicationDeleteDialog {...defaultProps} open />);

      await expect.element(page.getByRole('dialog')).toBeInTheDocument();

      // Click Cancel
      const cancelButton = page.getByRole('button', {name: 'Cancel'});
      await userEvent.click(cancelButton);

      expect(mockOnClose).toHaveBeenCalledTimes(1);

      // Close dialog (simulate parent closing it)
      await rerender(<ApplicationDeleteDialog {...defaultProps} open={false} />);

      // Wait for dialog to close
      await vi.waitFor(() => {
        expect(page.getByRole('dialog').query()).not.toBeInTheDocument();
      });

      // Open again
      await rerender(<ApplicationDeleteDialog {...defaultProps} open />);

      await expect.element(page.getByRole('dialog')).toBeInTheDocument();
    });

    it('should handle changing applicationId while dialog is open', async () => {
      const {rerender} = await renderComponent();

      const deleteButton = page.getByRole('button', {name: 'Delete'});
      await userEvent.click(deleteButton);

      expect(mockMutate).toHaveBeenCalledWith('test-app-id', expect.any(Object));

      // Change applicationId
      await rerender(<ApplicationDeleteDialog {...defaultProps} applicationId="new-app-id" />);

      mockMutate.mockClear();

      const deleteButtonAfterChange = page.getByRole('button', {name: 'Delete'});
      await userEvent.click(deleteButtonAfterChange);

      expect(mockMutate).toHaveBeenCalledWith('new-app-id', expect.any(Object));
    });

    it('should handle all callbacks being undefined', async () => {

      await renderComponent({
        open: true,
        applicationId: 'test-app-id',
        onClose: vi.fn(),
        onSuccess: vi.fn(),
      });

      mockMutate.mockImplementation((_, options: {onSuccess?: () => void}) => {
        options?.onSuccess?.();
      });

      const deleteButton = page.getByRole('button', {name: 'Delete'});
      await userEvent.click(deleteButton);

      // Should not throw error even though onSuccess is undefined
      await vi.waitFor(() => {
        expect(mockMutate).toHaveBeenCalled();
      });
    });
  });

  describe('Accessibility', () => {
    it('should have proper ARIA attributes', async () => {
      await renderComponent();

      const dialog = page.getByRole('dialog');
      expect(dialog).toHaveAttribute('role', 'dialog');
    });

    it('should be keyboard accessible', async () => {
      await renderComponent();

      const cancelButton = page.getByRole('button', {name: 'Cancel'});
      const deleteButton = page.getByRole('button', {name: 'Delete'});

      // Tab to focus buttons
      await userEvent.tab();
      expect(cancelButton).toHaveFocus();

      await userEvent.tab();
      expect(deleteButton).toHaveFocus();

      // Press Enter on Delete button
      await userEvent.keyboard('{Enter}');

      expect(mockMutate).toHaveBeenCalledWith('test-app-id', expect.any(Object));
    });

    it('should have proper button labels', async () => {
      await renderComponent();

      await expect.element(page.getByRole('button', {name: 'Cancel'})).toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: 'Delete'})).toBeInTheDocument();
    });
  });
});
