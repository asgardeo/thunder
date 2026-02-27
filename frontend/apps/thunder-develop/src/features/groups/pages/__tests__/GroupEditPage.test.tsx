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
import {screen, waitFor, fireEvent} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {renderWithProviders} from '@thunder/test-utils';
import GroupEditPage from '../GroupEditPage';

const mockNavigate = vi.fn();
vi.mock('react-router', async () => {
  const actual = await vi.importActual<typeof import('react-router')>('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => ({groupId: 'g1'}),
  };
});

const mockUseGetGroup = vi.fn();
vi.mock('../../api/useGetGroup', () => ({
  default: (...args: unknown[]): unknown => mockUseGetGroup(...args),
}));

const mockMutateAsync = vi.fn();
let mockIsPending = false;
vi.mock('../../api/useUpdateGroup', () => ({
  default: () => ({
    mutateAsync: mockMutateAsync,
    get isPending() {
      return mockIsPending;
    },
  }),
}));

vi.mock('../../components/GroupDeleteDialog', () => ({
  default: ({
    open,
    onClose,
    onSuccess,
  }: {
    open: boolean;
    groupId: string | null;
    onClose: () => void;
    onSuccess?: () => void;
  }) =>
    open ? (
      <div data-testid="delete-dialog">
        <button type="button" data-testid="close-delete-dialog" onClick={onClose}>
          Close
        </button>
        <button
          type="button"
          data-testid="delete-success"
          onClick={() => {
            onClose();
            onSuccess?.();
          }}
        >
          Confirm Delete
        </button>
      </div>
    ) : null,
}));

vi.mock('../../components/edit-group/general-settings/EditGeneralSettings', () => ({
  default: ({group, onDeleteClick}: {group: {id: string; name: string}; onDeleteClick: () => void}) => (
    <div data-testid="general-settings">
      <span>{group.name}</span>
      <button type="button" data-testid="delete-click" onClick={onDeleteClick}>
        Delete
      </button>
    </div>
  ),
}));

vi.mock('../../components/edit-group/members-settings/EditMembersSettings', () => ({
  default: ({group}: {group: {id: string; name: string}}) => (
    <div data-testid="members-settings">
      <span>Members of {group.name}</span>
    </div>
  ),
}));

const mockGroup = {
  id: 'g1',
  name: 'Test Group',
  description: 'A test group',
  organizationUnitId: 'ou-1',
  members: [],
};

describe('GroupEditPage', () => {
  let mockRefetch: ReturnType<typeof vi.fn>;
  let mockWriteText: ReturnType<typeof vi.fn>;
  const originalClipboard = navigator.clipboard;

  beforeEach(() => {
    vi.clearAllMocks();
    mockIsPending = false;
    mockNavigate.mockResolvedValue(undefined);
    mockRefetch = vi.fn().mockResolvedValue(undefined);
    mockWriteText = vi.fn().mockResolvedValue(undefined);
    mockUseGetGroup.mockReturnValue({
      data: mockGroup,
      isLoading: false,
      error: null,
      refetch: mockRefetch,
    });
    Object.defineProperty(navigator, 'clipboard', {
      value: {writeText: mockWriteText},
      writable: true,
      configurable: true,
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
    Object.defineProperty(navigator, 'clipboard', {
      value: originalClipboard,
      writable: true,
      configurable: true,
    });
  });

  it('should render loading state', () => {
    mockUseGetGroup.mockReturnValue({
      data: null,
      isLoading: true,
      error: null,
      refetch: vi.fn(),
    });
    renderWithProviders(<GroupEditPage />);

    expect(screen.getByRole('progressbar')).toBeInTheDocument();
  });

  it('should render error state', () => {
    mockUseGetGroup.mockReturnValue({
      data: null,
      isLoading: false,
      error: new Error('Fetch failed'),
      refetch: vi.fn(),
    });
    renderWithProviders(<GroupEditPage />);

    expect(screen.getByText('Fetch failed')).toBeInTheDocument();
    expect(screen.getByText('edit.page.back')).toBeInTheDocument();
  });

  it('should render not found state when no group', () => {
    mockUseGetGroup.mockReturnValue({
      data: null,
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    });
    renderWithProviders(<GroupEditPage />);

    expect(screen.getByText('edit.page.notFound')).toBeInTheDocument();
  });

  it('should render group name and description', () => {
    renderWithProviders(<GroupEditPage />);

    expect(screen.getAllByText('Test Group').length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText('A test group')).toBeInTheDocument();
  });

  it('should render group ID', () => {
    renderWithProviders(<GroupEditPage />);

    expect(screen.getByText('g1')).toBeInTheDocument();
  });

  it('should render back button and navigate on click', async () => {
    const user = userEvent.setup();
    renderWithProviders(<GroupEditPage />);

    await user.click(screen.getByText('edit.page.back'));

    expect(mockNavigate).toHaveBeenCalledWith('/groups');
  });

  it('should render tabs for general and members', () => {
    renderWithProviders(<GroupEditPage />);

    expect(screen.getByText('edit.page.tabs.general')).toBeInTheDocument();
    expect(screen.getByText('edit.page.tabs.members')).toBeInTheDocument();
  });

  it('should show general settings by default', () => {
    renderWithProviders(<GroupEditPage />);

    expect(screen.getByTestId('general-settings')).toBeInTheDocument();
  });

  it('should switch to members tab', async () => {
    const user = userEvent.setup();
    renderWithProviders(<GroupEditPage />);

    await user.click(screen.getByText('edit.page.tabs.members'));

    expect(screen.getByTestId('members-settings')).toBeInTheDocument();
  });

  it('should open delete dialog from general settings', async () => {
    const user = userEvent.setup();
    renderWithProviders(<GroupEditPage />);

    await user.click(screen.getByTestId('delete-click'));

    await waitFor(() => {
      expect(screen.getByTestId('delete-dialog')).toBeInTheDocument();
    });
  });

  it('should close delete dialog', async () => {
    const user = userEvent.setup();
    renderWithProviders(<GroupEditPage />);

    await user.click(screen.getByTestId('delete-click'));
    await waitFor(() => {
      expect(screen.getByTestId('delete-dialog')).toBeInTheDocument();
    });

    await user.click(screen.getByTestId('close-delete-dialog'));
    await waitFor(() => {
      expect(screen.queryByTestId('delete-dialog')).not.toBeInTheDocument();
    });
  });

  it('should navigate on successful delete', async () => {
    const user = userEvent.setup();
    renderWithProviders(<GroupEditPage />);

    await user.click(screen.getByTestId('delete-click'));
    await waitFor(() => {
      expect(screen.getByTestId('delete-dialog')).toBeInTheDocument();
    });

    await user.click(screen.getByTestId('delete-success'));
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/groups');
    });
  });

  it('should not show floating action bar initially', () => {
    renderWithProviders(<GroupEditPage />);

    expect(screen.queryByText('edit.page.unsavedChanges')).not.toBeInTheDocument();
  });

  it('should call useGetGroup with the groupId from params', () => {
    renderWithProviders(<GroupEditPage />);

    expect(mockUseGetGroup).toHaveBeenCalledWith('g1');
  });

  it('should navigate back from error state', async () => {
    mockUseGetGroup.mockReturnValue({
      data: null,
      isLoading: false,
      error: new Error('Fetch failed'),
      refetch: vi.fn(),
    });
    const user = userEvent.setup();
    renderWithProviders(<GroupEditPage />);

    await user.click(screen.getByText('edit.page.back'));

    expect(mockNavigate).toHaveBeenCalledWith('/groups');
  });

  it('should show empty description placeholder when no description', () => {
    mockUseGetGroup.mockReturnValue({
      data: {...mockGroup, description: undefined},
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    });
    renderWithProviders(<GroupEditPage />);

    expect(screen.getByText('edit.page.description.empty')).toBeInTheDocument();
  });

  it('should enter name editing mode and save on Enter', async () => {
    const user = userEvent.setup();
    renderWithProviders(<GroupEditPage />);

    // Find the name heading (h3) and the adjacent edit button
    const nameHeadings = screen.getAllByText('Test Group');
    // The h3 heading is the one rendered by GroupEditPage (not the mock)
    const h3Heading = nameHeadings.find((el) => el.tagName === 'H3');
    expect(h3Heading).toBeTruthy();
    const nameEditBtn = h3Heading!.parentElement?.querySelector('button');
    expect(nameEditBtn).toBeTruthy();
    await user.click(nameEditBtn!);

    // Should show a text field with the current name
    const nameInput = screen.getByDisplayValue('Test Group');
    expect(nameInput).toBeInTheDocument();

    // Clear and type new name
    await user.clear(nameInput);
    await user.type(nameInput, 'Updated Name');
    await user.keyboard('{Enter}');

    // Should show floating action bar
    await waitFor(() => {
      expect(screen.getByText('edit.page.unsavedChanges')).toBeInTheDocument();
    });
  });

  it('should cancel name editing on Escape', async () => {
    const user = userEvent.setup();
    renderWithProviders(<GroupEditPage />);

    const h3Heading = screen.getAllByText('Test Group').find((el) => el.tagName === 'H3');
    const nameEditBtn = h3Heading!.parentElement?.querySelector('button');
    await user.click(nameEditBtn!);

    const nameInput = screen.getByDisplayValue('Test Group');
    await user.clear(nameInput);
    await user.type(nameInput, 'Updated Name');
    await user.keyboard('{Escape}');

    // Should revert to original name and exit editing mode
    expect(screen.queryByDisplayValue('Updated Name')).not.toBeInTheDocument();
    expect(screen.getAllByText('Test Group').length).toBeGreaterThanOrEqual(1);
  });

  it('should save name on blur', async () => {
    const user = userEvent.setup();
    renderWithProviders(<GroupEditPage />);

    const h3Heading = screen.getAllByText('Test Group').find((el) => el.tagName === 'H3');
    const nameEditBtn = h3Heading!.parentElement?.querySelector('button');
    await user.click(nameEditBtn!);

    const nameInput = screen.getByDisplayValue('Test Group');
    await user.clear(nameInput);
    await user.type(nameInput, 'Blur Name');
    await user.tab(); // trigger blur

    await waitFor(() => {
      expect(screen.getByText('edit.page.unsavedChanges')).toBeInTheDocument();
    });
  });

  it('should enter description editing mode and save on Ctrl+Enter', async () => {
    const user = userEvent.setup();
    renderWithProviders(<GroupEditPage />);

    const descText = screen.getByText('A test group');
    const descEditBtn = descText.parentElement?.querySelector('button');
    expect(descEditBtn).toBeTruthy();
    await user.click(descEditBtn!);

    const descInput = screen.getByDisplayValue('A test group');
    expect(descInput).toBeInTheDocument();

    await user.clear(descInput);
    await user.type(descInput, 'Updated description');

    fireEvent.keyDown(descInput, {key: 'Enter', ctrlKey: true});

    await waitFor(() => {
      expect(screen.getByText('edit.page.unsavedChanges')).toBeInTheDocument();
    });
  });

  it('should cancel description editing on Escape', async () => {
    const user = userEvent.setup();
    renderWithProviders(<GroupEditPage />);

    const descText = screen.getByText('A test group');
    const descEditBtn = descText.parentElement?.querySelector('button');
    await user.click(descEditBtn!);

    const descInput = screen.getByDisplayValue('A test group');
    await user.clear(descInput);
    await user.type(descInput, 'Some new text');
    await user.keyboard('{Escape}');

    expect(screen.queryByDisplayValue('Some new text')).not.toBeInTheDocument();
  });

  it('should save description on blur', async () => {
    const user = userEvent.setup();
    renderWithProviders(<GroupEditPage />);

    const descText = screen.getByText('A test group');
    const descEditBtn = descText.parentElement?.querySelector('button');
    await user.click(descEditBtn!);

    const descInput = screen.getByDisplayValue('A test group');
    await user.clear(descInput);
    await user.type(descInput, 'Blurred desc');
    await user.tab();

    await waitFor(() => {
      expect(screen.getByText('edit.page.unsavedChanges')).toBeInTheDocument();
    });
  });

  it('should show empty placeholder after clearing description', async () => {
    const user = userEvent.setup();
    renderWithProviders(<GroupEditPage />);

    const descText = screen.getByText('A test group');
    const descEditBtn = descText.parentElement?.querySelector('button');
    await user.click(descEditBtn!);

    const descInput = screen.getByDisplayValue('A test group');
    await user.clear(descInput);
    await user.tab();

    await waitFor(() => {
      expect(screen.getByText('edit.page.description.empty')).toBeInTheDocument();
    });
  });

  it('should copy group ID to clipboard on click', async () => {
    renderWithProviders(<GroupEditPage />);

    const groupIdElement = screen.getByText('g1');
    const copyButton = groupIdElement.closest('[role="button"]');
    expect(copyButton).toBeTruthy();
    fireEvent.click(copyButton!);

    await waitFor(() => {
      expect(mockWriteText).toHaveBeenCalledWith('g1');
    });
  });

  it('should copy group ID on keyboard Enter', () => {
    renderWithProviders(<GroupEditPage />);

    const groupIdElement = screen.getByText('g1');
    const copyButton = groupIdElement.closest('[role="button"]');
    fireEvent.keyDown(copyButton!, {key: 'Enter'});

    expect(mockWriteText).toHaveBeenCalledWith('g1');
  });

  it('should copy group ID on keyboard Space', () => {
    renderWithProviders(<GroupEditPage />);

    const groupIdElement = screen.getByText('g1');
    const copyButton = groupIdElement.closest('[role="button"]');
    fireEvent.keyDown(copyButton!, {key: ' '});

    expect(mockWriteText).toHaveBeenCalledWith('g1');
  });

  it('should save changes when save button is clicked', async () => {
    mockMutateAsync.mockResolvedValue(undefined);
    const user = userEvent.setup();
    renderWithProviders(<GroupEditPage />);

    // Edit the name to trigger hasChanges
    const h3Heading = screen.getAllByText('Test Group').find((el) => el.tagName === 'H3');
    const nameEditBtn = h3Heading!.parentElement?.querySelector('button');
    await user.click(nameEditBtn!);
    const nameInput = screen.getByDisplayValue('Test Group');
    await user.clear(nameInput);
    await user.type(nameInput, 'New Name');
    await user.keyboard('{Enter}');

    await waitFor(() => {
      expect(screen.getByText('edit.page.unsavedChanges')).toBeInTheDocument();
    });

    await user.click(screen.getByText('edit.page.save'));

    await waitFor(() => {
      expect(mockMutateAsync).toHaveBeenCalledWith({
        groupId: 'g1',
        data: {
          name: 'New Name',
          description: 'A test group',
          organizationUnitId: 'ou-1',
        },
      });
    });

    expect(mockRefetch).toHaveBeenCalled();
  });

  it('should show error snackbar when save fails', async () => {
    mockMutateAsync.mockRejectedValue(new Error('Save failed'));
    const user = userEvent.setup();
    renderWithProviders(<GroupEditPage />);

    // Edit the name to trigger hasChanges
    const h3Heading = screen.getAllByText('Test Group').find((el) => el.tagName === 'H3');
    const nameEditBtn = h3Heading!.parentElement?.querySelector('button');
    await user.click(nameEditBtn!);
    const nameInput = screen.getByDisplayValue('Test Group');
    await user.clear(nameInput);
    await user.type(nameInput, 'New Name');
    await user.keyboard('{Enter}');

    await waitFor(() => {
      expect(screen.getByText('edit.page.save')).toBeInTheDocument();
    });

    await user.click(screen.getByText('edit.page.save'));

    await waitFor(() => {
      expect(screen.getByText('Save failed')).toBeInTheDocument();
    });
  });

  it('should reset changes when reset button is clicked', async () => {
    const user = userEvent.setup();
    renderWithProviders(<GroupEditPage />);

    // Edit the name
    const h3Heading = screen.getAllByText('Test Group').find((el) => el.tagName === 'H3');
    const nameEditBtn = h3Heading!.parentElement?.querySelector('button');
    await user.click(nameEditBtn!);
    const nameInput = screen.getByDisplayValue('Test Group');
    await user.clear(nameInput);
    await user.type(nameInput, 'New Name');
    await user.keyboard('{Enter}');

    await waitFor(() => {
      expect(screen.getByText('edit.page.unsavedChanges')).toBeInTheDocument();
    });

    await user.click(screen.getByText('edit.page.reset'));

    await waitFor(() => {
      expect(screen.queryByText('edit.page.unsavedChanges')).not.toBeInTheDocument();
    });
  });

  it('should navigate back from not found state', async () => {
    mockUseGetGroup.mockReturnValue({
      data: null,
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    });
    const user = userEvent.setup();
    renderWithProviders(<GroupEditPage />);

    await user.click(screen.getByText('edit.page.back'));

    expect(mockNavigate).toHaveBeenCalledWith('/groups');
  });

  it('should handle navigate rejection from back button gracefully', async () => {
    mockNavigate.mockRejectedValue(new Error('Nav failed'));
    const user = userEvent.setup();
    renderWithProviders(<GroupEditPage />);

    await user.click(screen.getByText('edit.page.back'));

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/groups');
    });
  });

  it('should close error snackbar when close button is clicked', async () => {
    mockMutateAsync.mockRejectedValue(new Error('Save failed'));
    const user = userEvent.setup();
    renderWithProviders(<GroupEditPage />);

    // Edit the name to trigger hasChanges
    const h3Heading = screen.getAllByText('Test Group').find((el) => el.tagName === 'H3');
    const nameEditBtn = h3Heading!.parentElement?.querySelector('button');
    await user.click(nameEditBtn!);
    const nameInput = screen.getByDisplayValue('Test Group');
    await user.clear(nameInput);
    await user.type(nameInput, 'New Name');
    await user.keyboard('{Enter}');

    await waitFor(() => {
      expect(screen.getByText('edit.page.save')).toBeInTheDocument();
    });

    await user.click(screen.getByText('edit.page.save'));

    await waitFor(() => {
      expect(screen.getByText('Save failed')).toBeInTheDocument();
    });

    // Close the snackbar via the Alert's close button
    const closeButton = screen.getByRole('button', {name: /close/i});
    await user.click(closeButton);

    await waitFor(() => {
      expect(screen.queryByText('Save failed')).not.toBeInTheDocument();
    });
  });
});
