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

import {render, screen, waitFor} from '@thunder/test-utils';
import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';
import userEvent from '@testing-library/user-event';
import type {MutateOptions, MutationFunctionContext} from '@tanstack/react-query';
import RegenerateSecretDialog from '../RegenerateSecretDialog';
import type {RegenerateSecretDialogProps} from '../RegenerateSecretDialog';
import type {RegenerateSecretVariables, RegenerateSecretResult} from '../../api/useRegenerateClientSecret';

// Mock the logger
vi.mock('@thunder/logger', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@thunder/logger')>();
  return {
    ...actual,
    useLogger: () => ({
      info: vi.fn(),
      error: vi.fn(),
      warn: vi.fn(),
      debug: vi.fn(),
    }),
  };
});

// Create a mock mutate function
const mockMutate = vi.fn();
const mockRegenerateClientSecret = {
  mutate: mockMutate,
  isPending: false,
};

// Mock useRegenerateClientSecret hook
vi.mock('../../api/useRegenerateClientSecret', () => ({
  default: () => mockRegenerateClientSecret,
}));

// Mock translations
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'applications:regenerateSecret.dialog.title': 'Regenerate Client Secret',
        'applications:regenerateSecret.dialog.message':
          'Are you sure you want to regenerate the client secret for this application? This will regenerate the client secret.',
        'applications:regenerateSecret.dialog.disclaimer':
          'This action will regenerate the client secret. All existing access tokens will be invalidated and the application will stop working until the new client secret is updated in your application configuration.',
        'applications:regenerateSecret.dialog.confirmButton': 'Regenerate',
        'applications:regenerateSecret.dialog.regenerating': 'Regenerating...',
        'applications:regenerateSecret.dialog.error': 'Failed to regenerate client secret. Please try again.',
        'common:actions.cancel': 'Cancel',
      };
      return translations[key] || key;
    },
  }),
}));

describe('RegenerateSecretDialog', () => {
  const mockOnClose = vi.fn();
  const mockOnSuccess = vi.fn();
  const mockOnError = vi.fn();

  const defaultProps: RegenerateSecretDialogProps = {
    open: true,
    applicationId: 'test-app-id',
    onClose: mockOnClose,
    onSuccess: mockOnSuccess,
    onError: mockOnError,
  };

  const renderDialog = (props: RegenerateSecretDialogProps = defaultProps) =>
    render(<RegenerateSecretDialog {...props} />);

  beforeEach(() => {
    vi.clearAllMocks();
    mockRegenerateClientSecret.isPending = false;
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render the dialog when open is true', () => {
      renderDialog();

      expect(screen.getByRole('dialog')).toBeInTheDocument();
      expect(screen.getByText('Regenerate Client Secret')).toBeInTheDocument();
      expect(
        screen.getByText(
          'Are you sure you want to regenerate the client secret for this application? This will regenerate the client secret.',
        ),
      ).toBeInTheDocument();
    });

    it('should show warning disclaimer', () => {
      renderDialog();

      expect(
        screen.getByText(
          'This action will regenerate the client secret. All existing access tokens will be invalidated and the application will stop working until the new client secret is updated in your application configuration.',
        ),
      ).toBeInTheDocument();
    });

    it('should not render dialog content when open is false', () => {
      renderDialog({...defaultProps, open: false});

      expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });

    it('should render Cancel and Regenerate buttons', () => {
      renderDialog();

      expect(screen.getByRole('button', {name: 'Cancel'})).toBeInTheDocument();
      expect(screen.getByRole('button', {name: 'Regenerate'})).toBeInTheDocument();
    });
  });

  describe('User Interactions', () => {
    it('should call onClose when Cancel button is clicked', async () => {
      const user = userEvent.setup();
      renderDialog();

      const cancelButton = screen.getByRole('button', {name: 'Cancel'});
      await user.click(cancelButton);

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('should call onClose when Escape key is pressed', async () => {
      const user = userEvent.setup();
      renderDialog();

      await user.keyboard('{Escape}');

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('should call mutate when Regenerate button is clicked', async () => {
      const user = userEvent.setup();
      renderDialog();

      const regenerateButton = screen.getByRole('button', {name: 'Regenerate'});
      await user.click(regenerateButton);

      expect(mockMutate).toHaveBeenCalledWith(
        {applicationId: 'test-app-id'},
        expect.objectContaining({
          // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
          onSuccess: expect.any(Function),
          // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
          onError: expect.any(Function),
        }),
      );
    });

    it('should not initiate regeneration when applicationId is null', async () => {
      renderDialog({...defaultProps, applicationId: null});

      const regenerateButton = screen.getByRole('button', {name: 'Regenerate'});

      expect(regenerateButton).toBeDisabled();
    });
  });

  describe('Success Flow', () => {
    it('should call onSuccess with new client secret after successful regeneration', async () => {
      // Mock mutate to immediately call onSuccess
      mockMutate.mockImplementation(
        (vars: RegenerateSecretVariables, options?: MutateOptions<RegenerateSecretResult, Error, RegenerateSecretVariables>) => {
          const mockContext = {} as MutationFunctionContext;
          options?.onSuccess?.({clientSecret: 'new-test-secret-123'} as RegenerateSecretResult, vars, undefined, mockContext);
        },
      );

      const user = userEvent.setup();
      renderDialog();

      const regenerateButton = screen.getByRole('button', {name: 'Regenerate'});
      await user.click(regenerateButton);

      await waitFor(() => {
        expect(mockOnClose).toHaveBeenCalled();
        expect(mockOnSuccess).toHaveBeenCalledWith('new-test-secret-123');
      });
    });
  });

  describe('Error Handling', () => {
    it('should display error message when regeneration fails', async () => {
      // Mock mutate to immediately call onError
      mockMutate.mockImplementation(
        (vars: RegenerateSecretVariables, options?: MutateOptions<RegenerateSecretResult, Error, RegenerateSecretVariables>) => {
          const mockContext = {} as MutationFunctionContext;
          options?.onError?.(new Error('Failed to regenerate client secret. Please try again.'), vars, undefined, mockContext);
        },
      );

      const user = userEvent.setup();
      renderDialog();

      const regenerateButton = screen.getByRole('button', {name: 'Regenerate'});
      await user.click(regenerateButton);

      await waitFor(() => {
        expect(screen.getByText('Failed to regenerate client secret. Please try again.')).toBeInTheDocument();
      });
    });

    it('should call onError callback when regeneration fails', async () => {
      // Mock mutate to immediately call onError
      mockMutate.mockImplementation(
        (vars: RegenerateSecretVariables, options?: MutateOptions<RegenerateSecretResult, Error, RegenerateSecretVariables>) => {
          const mockContext = {} as MutationFunctionContext;
          options?.onError?.(new Error('Failed to regenerate client secret. Please try again.'), vars, undefined, mockContext);
        },
      );

      const user = userEvent.setup();
      renderDialog();

      const regenerateButton = screen.getByRole('button', {name: 'Regenerate'});
      await user.click(regenerateButton);

      await waitFor(() => {
        expect(mockOnError).toHaveBeenCalledWith('Failed to regenerate client secret. Please try again.');
      });
    });
  });
});
