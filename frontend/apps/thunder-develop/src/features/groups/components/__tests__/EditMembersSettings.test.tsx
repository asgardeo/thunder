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

import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest';
import {screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {renderWithProviders} from '@thunder/test-utils';
import EditMembersSettings from '../edit-group/members-settings/EditMembersSettings';
import type {Group} from '../../models/group';

vi.mock('../edit-group/members-settings/ManageMembersSection', () => ({
  default: ({
    headerAction,
    onRemoveMember,
  }: {
    groupId: string;
    onRemoveMember: (member: {id: string; type: string}) => void;
    headerAction?: React.ReactNode;
  }) => (
    <div data-testid="manage-members-section">
      {headerAction && <div data-testid="header-action">{headerAction}</div>}
      <button
        type="button"
        data-testid="remove-member-btn"
        onClick={() => onRemoveMember({id: 'u1', type: 'user'})}
      >
        Remove
      </button>
    </div>
  ),
}));

vi.mock('../edit-group/members-settings/AddMemberDialog', () => ({
  default: ({open, onClose, onAdd}: {open: boolean; onClose: () => void; onAdd: (members: {id: string; type: string}[]) => void}) =>
    open ? (
      <div data-testid="add-member-dialog">
        <button type="button" onClick={onClose}>
          Close
        </button>
        <button type="button" onClick={() => onAdd([{id: 'u1', type: 'user'}])}>
          Add
        </button>
      </div>
    ) : null,
}));

const mockAddMutate = vi.fn();
vi.mock('../../api/useAddGroupMembers', () => ({
  default: () => ({mutate: mockAddMutate}),
}));

const mockRemoveMutate = vi.fn();
vi.mock('../../api/useRemoveGroupMembers', () => ({
  default: () => ({mutate: mockRemoveMutate}),
}));

describe('EditMembersSettings', () => {
  const mockGroup: Group = {
    id: 'g1',
    name: 'Test Group',
    organizationUnitId: 'ou1',
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should render ManageMembersSection', () => {
    renderWithProviders(<EditMembersSettings group={mockGroup} />);

    expect(screen.getByTestId('manage-members-section')).toBeInTheDocument();
  });

  it('should render Add Member button in header action', () => {
    renderWithProviders(<EditMembersSettings group={mockGroup} />);

    expect(screen.getByText('edit.members.sections.manage.addMember')).toBeInTheDocument();
  });

  it('should open add member dialog when button clicked', async () => {
    const user = userEvent.setup();
    renderWithProviders(<EditMembersSettings group={mockGroup} />);

    await user.click(screen.getByText('edit.members.sections.manage.addMember'));

    await waitFor(() => {
      expect(screen.getByTestId('add-member-dialog')).toBeInTheDocument();
    });
  });

  it('should call addGroupMembers when members are added', async () => {
    const user = userEvent.setup();
    renderWithProviders(<EditMembersSettings group={mockGroup} />);

    await user.click(screen.getByText('edit.members.sections.manage.addMember'));

    await waitFor(() => {
      expect(screen.getByTestId('add-member-dialog')).toBeInTheDocument();
    });

    await user.click(screen.getByText('Add'));

    expect(mockAddMutate).toHaveBeenCalledWith(
      {groupId: 'g1', members: [{id: 'u1', type: 'user'}]},
      expect.objectContaining({onSuccess: expect.any(Function) as unknown, onError: expect.any(Function) as unknown}),
    );
  });

  it('should close dialog when close is clicked', async () => {
    const user = userEvent.setup();
    renderWithProviders(<EditMembersSettings group={mockGroup} />);

    await user.click(screen.getByText('edit.members.sections.manage.addMember'));
    await waitFor(() => {
      expect(screen.getByTestId('add-member-dialog')).toBeInTheDocument();
    });

    await user.click(screen.getByText('Close'));

    await waitFor(() => {
      expect(screen.queryByTestId('add-member-dialog')).not.toBeInTheDocument();
    });
  });

  it('should close dialog on successful add', async () => {
    mockAddMutate.mockImplementation(
      (_data: unknown, opts: {onSuccess: () => void}) => {
        opts.onSuccess();
      },
    );

    const user = userEvent.setup();
    renderWithProviders(<EditMembersSettings group={mockGroup} />);

    await user.click(screen.getByText('edit.members.sections.manage.addMember'));
    await waitFor(() => {
      expect(screen.getByTestId('add-member-dialog')).toBeInTheDocument();
    });

    await user.click(screen.getByText('Add'));

    await waitFor(() => {
      expect(screen.queryByTestId('add-member-dialog')).not.toBeInTheDocument();
    });
  });

  it('should show error when add fails', async () => {
    mockAddMutate.mockImplementation(
      (_data: unknown, opts: {onError: (err: Error) => void}) => {
        opts.onError(new Error('Add failed'));
      },
    );

    const user = userEvent.setup();
    renderWithProviders(<EditMembersSettings group={mockGroup} />);

    await user.click(screen.getByText('edit.members.sections.manage.addMember'));
    await waitFor(() => {
      expect(screen.getByTestId('add-member-dialog')).toBeInTheDocument();
    });

    await user.click(screen.getByText('Add'));

    await waitFor(() => {
      expect(screen.getByText('Add failed')).toBeInTheDocument();
    });
  });

  it('should call removeGroupMembers when remove button is clicked', async () => {
    const user = userEvent.setup();
    renderWithProviders(<EditMembersSettings group={mockGroup} />);

    await user.click(screen.getByTestId('remove-member-btn'));

    expect(mockRemoveMutate).toHaveBeenCalledWith(
      {groupId: 'g1', members: [{id: 'u1', type: 'user'}]},
      expect.objectContaining({onSuccess: expect.any(Function) as unknown, onError: expect.any(Function) as unknown}),
    );
  });

  it('should clear error on successful remove', async () => {
    // First trigger an error
    mockAddMutate.mockImplementation(
      (_data: unknown, opts: {onError: (err: Error) => void}) => {
        opts.onError(new Error('Some error'));
      },
    );

    const user = userEvent.setup();
    renderWithProviders(<EditMembersSettings group={mockGroup} />);

    await user.click(screen.getByText('edit.members.sections.manage.addMember'));
    await waitFor(() => {
      expect(screen.getByTestId('add-member-dialog')).toBeInTheDocument();
    });
    await user.click(screen.getByText('Add'));
    await waitFor(() => {
      expect(screen.getByText('Some error')).toBeInTheDocument();
    });

    // Now trigger successful remove which should clear the error
    mockRemoveMutate.mockImplementation(
      (_data: unknown, opts: {onSuccess: () => void}) => {
        opts.onSuccess();
      },
    );

    await user.click(screen.getByTestId('remove-member-btn'));

    await waitFor(() => {
      expect(screen.queryByText('Some error')).not.toBeInTheDocument();
    });
  });

  it('should show error when remove fails', async () => {
    mockRemoveMutate.mockImplementation(
      (_data: unknown, opts: {onError: (err: Error) => void}) => {
        opts.onError(new Error('Remove failed'));
      },
    );

    const user = userEvent.setup();
    renderWithProviders(<EditMembersSettings group={mockGroup} />);

    await user.click(screen.getByTestId('remove-member-btn'));

    await waitFor(() => {
      expect(screen.getByText('Remove failed')).toBeInTheDocument();
    });
  });

  it('should close error alert when dismiss is clicked', async () => {
    mockRemoveMutate.mockImplementation(
      (_data: unknown, opts: {onError: (err: Error) => void}) => {
        opts.onError(new Error('Remove failed'));
      },
    );

    const user = userEvent.setup();
    renderWithProviders(<EditMembersSettings group={mockGroup} />);

    await user.click(screen.getByTestId('remove-member-btn'));
    await waitFor(() => {
      expect(screen.getByText('Remove failed')).toBeInTheDocument();
    });

    const closeAlertButton = screen.getByRole('button', {name: /close/i});
    await user.click(closeAlertButton);

    await waitFor(() => {
      expect(screen.queryByText('Remove failed')).not.toBeInTheDocument();
    });
  });
});
