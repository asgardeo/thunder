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
import {render, screen, waitFor} from '@thunderid/test-utils';
import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';
import * as useAddRoleAssignmentsModule from '../../../../api/useAddRoleAssignments';
import * as useRemoveRoleAssignmentsModule from '../../../../api/useRemoveRoleAssignments';
import type {RoleAssignment} from '../../../../models/role';
import EditAssignmentsSettings from '../EditAssignmentsSettings';

vi.mock('../../../../api/useAddRoleAssignments');
vi.mock('../../../../api/useRemoveRoleAssignments');

vi.mock('../ManageAssignmentsSection', () => ({
  default: ({
    onRemoveAssignment,
    headerAction,
    activeAssignmentTab,
    onAssignmentTabChange,
  }: {
    roleId: string;
    onRemoveAssignment: (a: RoleAssignment) => void;
    headerAction?: React.ReactNode;
    activeAssignmentTab: number;
    onAssignmentTabChange: (tab: number) => void;
  }) => (
    <div data-testid="manage-section">
      {headerAction}
      <button type="button" onClick={() => onRemoveAssignment({id: 'user-1', type: 'user'})} data-testid="remove-btn">
        Remove
      </button>
      <button type="button" onClick={() => onAssignmentTabChange(1)} data-testid="switch-tab-btn">
        Switch Tab
      </button>
      <span data-testid="active-tab">{activeAssignmentTab}</span>
    </div>
  ),
}));

vi.mock('../AddAssignmentDialog', () => ({
  default: ({
    open,
    onClose,
    onAdd,
    initialTab,
  }: {
    open: boolean;
    roleId: string;
    onClose: () => void;
    onAdd: (assignments: RoleAssignment[]) => void;
    initialTab?: number;
  }) =>
    open ? (
      <div data-testid="add-dialog" role="dialog">
        <span data-testid="initial-tab">{initialTab}</span>
        <button type="button" onClick={() => onAdd([{id: 'user-3', type: 'user'}])} data-testid="confirm-add">
          Confirm Add
        </button>
        <button type="button" onClick={onClose} data-testid="close-dialog">
          Cancel
        </button>
      </div>
    ) : null,
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'roles:edit.assignments.sections.manage.addAssignment': 'Add Assignment',
        'roles:assignments.add.error': 'Failed to add assignments',
        'roles:assignments.remove.error': 'Failed to remove assignment',
      };
      return translations[key] || key;
    },
  }),
}));

describe('EditAssignmentsSettings', () => {
  const mockAddMutate = vi.fn();
  const mockRemoveMutate = vi.fn();

  const baseMutationState = {
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
    status: 'idle' as const,
    submittedAt: 0,
    variables: undefined,
  };

  const renderComponent = () => render(<EditAssignmentsSettings roleId="role-1" />);

  beforeEach(() => {
    vi.mocked(useAddRoleAssignmentsModule.default).mockReturnValue({
      ...baseMutationState,
      mutate: mockAddMutate,
    } as unknown as ReturnType<typeof useAddRoleAssignmentsModule.default>);

    vi.mocked(useRemoveRoleAssignmentsModule.default).mockReturnValue({
      ...baseMutationState,
      mutate: mockRemoveMutate,
    } as unknown as ReturnType<typeof useRemoveRoleAssignmentsModule.default>);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render ManageAssignmentsSection', () => {
      renderComponent();

      expect(screen.getByTestId('manage-section')).toBeInTheDocument();
    });

    it('should render Add Assignment button in headerAction', () => {
      renderComponent();

      expect(screen.getByRole('button', {name: 'Add Assignment'})).toBeInTheDocument();
    });

    it('should not render AddAssignmentDialog initially', () => {
      renderComponent();

      expect(screen.queryByTestId('add-dialog')).not.toBeInTheDocument();
    });

    it('should not render error alert initially', () => {
      renderComponent();

      expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    });
  });

  describe('Add Dialog Interactions', () => {
    it('should open AddAssignmentDialog when Add Assignment button is clicked', async () => {
      const user = userEvent.setup();
      renderComponent();

      await user.click(screen.getByRole('button', {name: 'Add Assignment'}));

      expect(screen.getByTestId('add-dialog')).toBeInTheDocument();
    });

    it('should pass activeAssignmentTab as initialTab to dialog', async () => {
      const user = userEvent.setup();
      renderComponent();

      await user.click(screen.getByTestId('switch-tab-btn'));
      await user.click(screen.getByRole('button', {name: 'Add Assignment'}));

      expect(screen.getByTestId('initial-tab')).toHaveTextContent('1');
    });

    it('should call addRoleAssignments.mutate when dialog confirms', async () => {
      const user = userEvent.setup();
      renderComponent();

      await user.click(screen.getByRole('button', {name: 'Add Assignment'}));
      await user.click(screen.getByTestId('confirm-add'));

      expect(mockAddMutate).toHaveBeenCalledTimes(1);
      expect(mockAddMutate).toHaveBeenCalledWith(
        {roleId: 'role-1', assignments: [{id: 'user-3', type: 'user'}]},
        expect.any(Object),
      );
    });

    it('should close dialog on successful add', async () => {
      const user = userEvent.setup();
      mockAddMutate.mockImplementation(
        (_variables: unknown, options: {onSuccess?: () => void; onError?: (error: Error) => void}) => {
          options?.onSuccess?.();
        },
      );

      renderComponent();

      await user.click(screen.getByRole('button', {name: 'Add Assignment'}));
      await user.click(screen.getByTestId('confirm-add'));

      await waitFor(() => {
        expect(screen.queryByTestId('add-dialog')).not.toBeInTheDocument();
      });
    });

    it('should show error alert on add failure', async () => {
      const user = userEvent.setup();
      mockAddMutate.mockImplementation(
        (_variables: unknown, options: {onSuccess?: () => void; onError?: (error: Error) => void}) => {
          options?.onError?.(new Error('Network error'));
        },
      );

      renderComponent();

      await user.click(screen.getByRole('button', {name: 'Add Assignment'}));
      await user.click(screen.getByTestId('confirm-add'));

      await waitFor(() => {
        expect(screen.getByText('Network error')).toBeInTheDocument();
      });
    });

    it('should close dialog when Cancel is clicked', async () => {
      const user = userEvent.setup();
      renderComponent();

      await user.click(screen.getByRole('button', {name: 'Add Assignment'}));
      expect(screen.getByTestId('add-dialog')).toBeInTheDocument();

      await user.click(screen.getByTestId('close-dialog'));

      await waitFor(() => {
        expect(screen.queryByTestId('add-dialog')).not.toBeInTheDocument();
      });
    });
  });

  describe('Remove Assignment', () => {
    it('should call removeRoleAssignments.mutate when remove is triggered', async () => {
      const user = userEvent.setup();
      renderComponent();

      await user.click(screen.getByTestId('remove-btn'));

      expect(mockRemoveMutate).toHaveBeenCalledTimes(1);
      expect(mockRemoveMutate).toHaveBeenCalledWith(
        {roleId: 'role-1', assignments: [{id: 'user-1', type: 'user'}]},
        expect.any(Object),
      );
    });

    it('should clear error on successful remove', async () => {
      const user = userEvent.setup();

      mockAddMutate.mockImplementation(
        (_variables: unknown, options: {onSuccess?: () => void; onError?: (error: Error) => void}) => {
          options?.onError?.(new Error('Some error'));
        },
      );

      renderComponent();

      await user.click(screen.getByRole('button', {name: 'Add Assignment'}));
      await user.click(screen.getByTestId('confirm-add'));

      await waitFor(() => {
        expect(screen.getByText('Some error')).toBeInTheDocument();
      });

      mockRemoveMutate.mockImplementation(
        (_variables: unknown, options: {onSuccess?: () => void; onError?: (error: Error) => void}) => {
          options?.onSuccess?.();
        },
      );

      await user.click(screen.getByTestId('remove-btn'));

      await waitFor(() => {
        expect(screen.queryByRole('alert')).not.toBeInTheDocument();
      });
    });

    it('should show error alert on remove failure', async () => {
      const user = userEvent.setup();
      mockRemoveMutate.mockImplementation(
        (_variables: unknown, options: {onSuccess?: () => void; onError?: (error: Error) => void}) => {
          options?.onError?.(new Error('Remove failed'));
        },
      );

      renderComponent();

      await user.click(screen.getByTestId('remove-btn'));

      await waitFor(() => {
        expect(screen.getByText('Remove failed')).toBeInTheDocument();
      });
    });
  });

  describe('Error Handling', () => {
    it('should clear error alert when close button is clicked', async () => {
      const user = userEvent.setup();
      mockRemoveMutate.mockImplementation(
        (_variables: unknown, options: {onSuccess?: () => void; onError?: (error: Error) => void}) => {
          options?.onError?.(new Error('Some error'));
        },
      );

      renderComponent();

      await user.click(screen.getByTestId('remove-btn'));

      await waitFor(() => {
        expect(screen.getByRole('alert')).toBeInTheDocument();
      });

      const closeButton = screen.getByRole('alert').querySelector('button');
      if (closeButton) {
        await user.click(closeButton);
      }

      await waitFor(() => {
        expect(screen.queryByRole('alert')).not.toBeInTheDocument();
      });
    });

    it('should clear error on successful add', async () => {
      const user = userEvent.setup();

      mockRemoveMutate.mockImplementation(
        (_variables: unknown, options: {onSuccess?: () => void; onError?: (error: Error) => void}) => {
          options?.onError?.(new Error('Previous error'));
        },
      );

      renderComponent();

      await user.click(screen.getByTestId('remove-btn'));

      await waitFor(() => {
        expect(screen.getByText('Previous error')).toBeInTheDocument();
      });

      mockAddMutate.mockImplementation(
        (_variables: unknown, options: {onSuccess?: () => void; onError?: (error: Error) => void}) => {
          options?.onSuccess?.();
        },
      );

      await user.click(screen.getByRole('button', {name: 'Add Assignment'}));
      await user.click(screen.getByTestId('confirm-add'));

      await waitFor(() => {
        expect(screen.queryByRole('alert')).not.toBeInTheDocument();
      });
    });
  });
});
