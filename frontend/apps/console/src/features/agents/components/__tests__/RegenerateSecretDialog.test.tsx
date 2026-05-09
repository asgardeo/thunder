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

import type {MutateOptions, MutationFunctionContext} from '@tanstack/react-query';
import userEvent from '@testing-library/user-event';
import {render, screen, waitFor} from '@thunderid/test-utils';
import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';
import type {RegenerateAgentSecretVariables, RegenerateAgentSecretResult} from '../../api/useRegenerateAgentSecret';
import RegenerateSecretDialog from '../RegenerateSecretDialog';
import type {RegenerateSecretDialogProps} from '../RegenerateSecretDialog';

// Mock the logger
vi.mock('@thunderid/logger', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@thunderid/logger')>();
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

const mockMutate = vi.fn();
const mockRegenerateAgentSecret = {
  mutate: mockMutate,
  isPending: false,
};

// Mock useRegenerateAgentSecret hook
vi.mock('../../api/useRegenerateAgentSecret', () => ({
  default: () => mockRegenerateAgentSecret,
}));

// Mock translations
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, fallback?: string) => {
      const translations: Record<string, string> = {
        'agents:regenerateSecret.dialog.title': 'Regenerate client secret?',
        'agents:regenerateSecret.dialog.message':
          'A new client secret will be generated for this agent. Any service using the current client secret will stop working immediately.',
        'agents:regenerateSecret.dialog.disclaimer':
          'This action cannot be undone. The current client secret will be invalidated as soon as you confirm.',
        'agents:regenerateSecret.dialog.confirmButton': 'Regenerate',
        'agents:regenerateSecret.dialog.regenerating': 'Regenerating…',
        'agents:regenerateSecret.dialog.error': 'Failed to regenerate client secret',
        'common:actions.cancel': 'Cancel',
      };
      return translations[key] ?? fallback ?? key;
    },
  }),
}));

describe('RegenerateSecretDialog', () => {
  const mockOnClose = vi.fn();
  const mockOnSuccess = vi.fn();
  const mockOnError = vi.fn();

  const defaultProps: RegenerateSecretDialogProps = {
    open: true,
    agentId: 'test-agent-id',
    onClose: mockOnClose,
    onSuccess: mockOnSuccess,
    onError: mockOnError,
  };

  const renderDialog = (props: RegenerateSecretDialogProps = defaultProps) =>
    render(<RegenerateSecretDialog {...props} />);

  beforeEach(() => {
    vi.clearAllMocks();
    mockRegenerateAgentSecret.isPending = false;
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render the dialog when open is true', () => {
      renderDialog();

      expect(screen.getByRole('dialog')).toBeInTheDocument();
      expect(screen.getByText('Regenerate client secret?')).toBeInTheDocument();
    });

    it('should show warning disclaimer', () => {
      renderDialog();

      expect(
        screen.getByText(
          'This action cannot be undone. The current client secret will be invalidated as soon as you confirm.',
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
        {agentId: 'test-agent-id'},
        expect.objectContaining({
          // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
          onSuccess: expect.any(Function),
          // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
          onError: expect.any(Function),
        }),
      );
    });

    it('should disable Regenerate button when agentId is null', () => {
      renderDialog({...defaultProps, agentId: null});

      const regenerateButton = screen.getByRole('button', {name: 'Regenerate'});
      expect(regenerateButton).toBeDisabled();
    });
  });

  describe('Success Flow', () => {
    it('should call onSuccess with new client secret after successful regeneration', async () => {
      mockMutate.mockImplementation(
        (
          vars: RegenerateAgentSecretVariables,
          options?: MutateOptions<RegenerateAgentSecretResult, Error, RegenerateAgentSecretVariables>,
        ) => {
          const mockContext = {} as MutationFunctionContext;
          options?.onSuccess?.(
            {clientSecret: 'new-test-secret-123'} as RegenerateAgentSecretResult,
            vars,
            undefined,
            mockContext,
          );
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
      mockMutate.mockImplementation(
        (
          vars: RegenerateAgentSecretVariables,
          options?: MutateOptions<RegenerateAgentSecretResult, Error, RegenerateAgentSecretVariables>,
        ) => {
          const mockContext = {} as MutationFunctionContext;
          options?.onError?.(new Error('Failed to regenerate client secret'), vars, undefined, mockContext);
        },
      );

      const user = userEvent.setup();
      renderDialog();

      const regenerateButton = screen.getByRole('button', {name: 'Regenerate'});
      await user.click(regenerateButton);

      await waitFor(() => {
        expect(screen.getByText('Failed to regenerate client secret')).toBeInTheDocument();
      });
    });

    it('should call onError callback when regeneration fails', async () => {
      mockMutate.mockImplementation(
        (
          vars: RegenerateAgentSecretVariables,
          options?: MutateOptions<RegenerateAgentSecretResult, Error, RegenerateAgentSecretVariables>,
        ) => {
          const mockContext = {} as MutationFunctionContext;
          options?.onError?.(new Error('Failed to regenerate client secret'), vars, undefined, mockContext);
        },
      );

      const user = userEvent.setup();
      renderDialog();

      const regenerateButton = screen.getByRole('button', {name: 'Regenerate'});
      await user.click(regenerateButton);

      await waitFor(() => {
        expect(mockOnError).toHaveBeenCalledWith('Failed to regenerate client secret');
      });
    });

    it('should fall back to default error message for non-Error rejections', async () => {
      mockMutate.mockImplementation(
        (
          vars: RegenerateAgentSecretVariables,
          options?: MutateOptions<RegenerateAgentSecretResult, Error, RegenerateAgentSecretVariables>,
        ) => {
          const mockContext = {} as MutationFunctionContext;
          // Simulate a non-Error rejection
          options?.onError?.('string error' as unknown as Error, vars, undefined, mockContext);
        },
      );

      const user = userEvent.setup();
      renderDialog();

      const regenerateButton = screen.getByRole('button', {name: 'Regenerate'});
      await user.click(regenerateButton);

      await waitFor(() => {
        expect(screen.getByText('Failed to regenerate client secret')).toBeInTheDocument();
      });
    });

    it('should clear error when Cancel is clicked after error', async () => {
      mockMutate.mockImplementation(
        (
          vars: RegenerateAgentSecretVariables,
          options?: MutateOptions<RegenerateAgentSecretResult, Error, RegenerateAgentSecretVariables>,
        ) => {
          const mockContext = {} as MutationFunctionContext;
          options?.onError?.(new Error('Some error'), vars, undefined, mockContext);
        },
      );

      const user = userEvent.setup();
      renderDialog();

      await user.click(screen.getByRole('button', {name: 'Regenerate'}));

      await waitFor(() => {
        expect(screen.getByText('Some error')).toBeInTheDocument();
      });

      await user.click(screen.getByRole('button', {name: 'Cancel'}));

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });
  });
});
