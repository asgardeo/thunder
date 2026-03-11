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

import {render} from '@thunder/test-utils/browser';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import {page, userEvent} from 'vitest/browser';
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

  beforeEach(() => {
    vi.clearAllMocks();
    mockRegenerateClientSecret.isPending = false;
  });

  describe('Rendering', () => {
    it('should render the dialog when open is true', async () => {
      await render(<RegenerateSecretDialog {...defaultProps} />);

      await expect.element(page.getByRole('dialog')).toBeInTheDocument();
      await expect.element(page.getByText('Regenerate Client Secret')).toBeInTheDocument();
      await expect.element(
        page.getByText(
          'Are you sure you want to regenerate the client secret for this application? This will immediately invalidate the current client secret and generate a new one.',
        ),
      ).toBeInTheDocument();
    });

    it('should show warning disclaimer', async () => {
      await render(<RegenerateSecretDialog {...defaultProps} />);

      await expect.element(
        page.getByText(
          'Warning: Regenerating the client secret will invalidate the current secret and the application may stop working until the new client secret is updated in its configuration.',
        ),
      ).toBeInTheDocument();
    });

    it('should not render dialog content when open is false', async () => {
      await render(<RegenerateSecretDialog {...defaultProps} open={false} />);

      await expect.element(page.getByRole('dialog')).not.toBeInTheDocument();
    });

    it('should render Cancel and Regenerate buttons', async () => {
      await render(<RegenerateSecretDialog {...defaultProps} />);

      await expect.element(page.getByRole('button', {name: 'Cancel'})).toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: 'Regenerate'})).toBeInTheDocument();
    });
  });

  describe('User Interactions', () => {
    it('should call onClose when Cancel button is clicked', async () => {
      await render(<RegenerateSecretDialog {...defaultProps} />);

      await userEvent.click(page.getByRole('button', {name: 'Cancel'}));

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('should call onClose when Escape key is pressed', async () => {
      await render(<RegenerateSecretDialog {...defaultProps} />);

      await userEvent.keyboard('{Escape}');

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('should call mutate when Regenerate button is clicked', async () => {
      await render(<RegenerateSecretDialog {...defaultProps} />);

      await userEvent.click(page.getByRole('button', {name: 'Regenerate'}));

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
      await render(<RegenerateSecretDialog {...defaultProps} applicationId={null} />);

      const regenerateButton = page.getByRole('button', {name: 'Regenerate'});
      expect(regenerateButton.element()).toBeDisabled();
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

      await render(<RegenerateSecretDialog {...defaultProps} />);

      await userEvent.click(page.getByRole('button', {name: 'Regenerate'}));

      await vi.waitFor(() => {
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

      await render(<RegenerateSecretDialog {...defaultProps} />);

      await userEvent.click(page.getByRole('button', {name: 'Regenerate'}));

      await vi.waitFor(async () => {
        await expect.element(page.getByText('Failed to regenerate client secret. Please try again.')).toBeInTheDocument();
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

      await render(<RegenerateSecretDialog {...defaultProps} />);

      await userEvent.click(page.getByRole('button', {name: 'Regenerate'}));

      await vi.waitFor(() => {
        expect(mockOnError).toHaveBeenCalledWith('Failed to regenerate client secret. Please try again.');
      });
    });
  });
});
