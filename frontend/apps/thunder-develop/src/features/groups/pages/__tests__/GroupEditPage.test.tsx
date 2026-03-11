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
import {page, userEvent} from 'vitest/browser';
import {renderWithProviders} from '@thunder/test-utils/browser';
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

/**
 * Helper to click the edit button adjacent to the name heading (h3).
 * The name text appears in both the h3 heading and the mocked general-settings span,
 * so we need to specifically find the h3 to locate the edit button next to it.
 */
async function clickNameEditButton() {
  const nameHeadings = page.getByText('Test Group').all();
  const h3Heading = nameHeadings.find((el) => el.element().tagName === 'H3');
  expect(h3Heading).toBeTruthy();
  const nameEditBtn = h3Heading!.element().parentElement?.querySelector('button');
  expect(nameEditBtn).toBeTruthy();
  await userEvent.click(nameEditBtn!);
}

/**
 * Helper to click the edit button adjacent to the description text.
 */
async function clickDescriptionEditButton() {
  const descText = page.getByText('A test group');
  const descEditBtn = descText.element().parentElement?.querySelector('button');
  expect(descEditBtn).toBeTruthy();
  await userEvent.click(descEditBtn!);
}

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

  it('should render loading state', async () => {
    mockUseGetGroup.mockReturnValue({
      data: null,
      isLoading: true,
      error: null,
      refetch: vi.fn(),
    });
    await renderWithProviders(<GroupEditPage />);

    await expect.element(page.getByRole('progressbar')).toBeInTheDocument();
  });

  it('should render error state', async () => {
    mockUseGetGroup.mockReturnValue({
      data: null,
      isLoading: false,
      error: new Error('Fetch failed'),
      refetch: vi.fn(),
    });
    await renderWithProviders(<GroupEditPage />);

    await expect.element(page.getByText('Fetch failed')).toBeInTheDocument();
    await expect.element(page.getByText('Back to Groups')).toBeInTheDocument();
  });

  it('should render not found state when no group', async () => {
    mockUseGetGroup.mockReturnValue({
      data: null,
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    });
    await renderWithProviders(<GroupEditPage />);

    await expect.element(page.getByText('Group not found')).toBeInTheDocument();
  });

  it('should render group name and description', async () => {
    await renderWithProviders(<GroupEditPage />);

    await expect.element(page.getByRole('heading', {name: 'Test Group', level: 3})).toBeInTheDocument();
    await expect.element(page.getByText('A test group')).toBeInTheDocument();
  });

  it('should render group ID', async () => {
    await renderWithProviders(<GroupEditPage />);

    await expect.element(page.getByText('g1')).toBeInTheDocument();
  });

  it('should render back button and navigate on click', async () => {
    await renderWithProviders(<GroupEditPage />);

    // PageTitle.BackButton renders as a link element; use getByRole to avoid strict mode violation
    // from multiple elements containing "Back to Groups" text
    const backLinks = page.getByText('Back to Groups').all();
    expect(backLinks.length).toBeGreaterThan(0);
    await userEvent.click(backLinks[0]);

    expect(mockNavigate).toHaveBeenCalledWith('/groups');
  });

  it('should render tabs for general and members', async () => {
    await renderWithProviders(<GroupEditPage />);

    await expect.element(page.getByRole('tab', {name: 'General'})).toBeInTheDocument();
    await expect.element(page.getByRole('tab', {name: 'Members'})).toBeInTheDocument();
  });

  it('should show general settings by default', async () => {
    await renderWithProviders(<GroupEditPage />);

    await expect.element(page.getByTestId('general-settings')).toBeInTheDocument();
  });

  it('should switch to members tab', async () => {
    await renderWithProviders(<GroupEditPage />);

    await userEvent.click(page.getByRole('tab', {name: 'Members'}));

    await expect.element(page.getByTestId('members-settings')).toBeInTheDocument();
  });

  it('should open delete dialog from general settings', async () => {
    await renderWithProviders(<GroupEditPage />);

    await userEvent.click(page.getByTestId('delete-click'));

    await expect.element(page.getByTestId('delete-dialog')).toBeInTheDocument();
  });

  it('should close delete dialog', async () => {
    await renderWithProviders(<GroupEditPage />);

    await userEvent.click(page.getByTestId('delete-click'));
    await expect.element(page.getByTestId('delete-dialog')).toBeInTheDocument();

    await userEvent.click(page.getByTestId('close-delete-dialog'));
    await expect.element(page.getByTestId('delete-dialog')).not.toBeInTheDocument();
  });

  it('should navigate on successful delete', async () => {
    await renderWithProviders(<GroupEditPage />);

    await userEvent.click(page.getByTestId('delete-click'));
    await expect.element(page.getByTestId('delete-dialog')).toBeInTheDocument();

    await userEvent.click(page.getByTestId('delete-success'));
    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/groups');
    });
  });

  it('should not show floating action bar initially', async () => {
    await renderWithProviders(<GroupEditPage />);

    await expect.element(page.getByText('You have unsaved changes')).not.toBeInTheDocument();
  });

  it('should call useGetGroup with the groupId from params', async () => {
    await renderWithProviders(<GroupEditPage />);

    expect(mockUseGetGroup).toHaveBeenCalledWith('g1');
  });

  it('should navigate back from error state', async () => {
    mockUseGetGroup.mockReturnValue({
      data: null,
      isLoading: false,
      error: new Error('Fetch failed'),
      refetch: vi.fn(),
    });
    await renderWithProviders(<GroupEditPage />);

    await userEvent.click(page.getByText('Back to Groups'));

    expect(mockNavigate).toHaveBeenCalledWith('/groups');
  });

  it('should show empty description placeholder when no description', async () => {
    mockUseGetGroup.mockReturnValue({
      data: {...mockGroup, description: undefined},
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    });
    await renderWithProviders(<GroupEditPage />);

    await expect.element(page.getByText('No description')).toBeInTheDocument();
  });

  it('should enter name editing mode and save on Enter', async () => {
    await renderWithProviders(<GroupEditPage />);

    await clickNameEditButton();

    // The autoFocused textbox should appear with the current name
    const nameInput = page.getByRole('textbox');
    await expect.element(nameInput).toBeInTheDocument();

    // Clear and type new name
    await userEvent.clear(nameInput);
    await userEvent.fill(nameInput, 'Updated Name');
    await userEvent.keyboard('{Enter}');

    // Should show floating action bar
    await expect.element(page.getByText('You have unsaved changes')).toBeInTheDocument();
  });

  it('should cancel name editing on Escape', async () => {
    await renderWithProviders(<GroupEditPage />);

    await clickNameEditButton();

    const nameInput = page.getByRole('textbox');
    await expect.element(nameInput).toBeInTheDocument();

    await userEvent.clear(nameInput);
    await userEvent.fill(nameInput, 'Updated Name');
    await userEvent.keyboard('{Escape}');

    // Should revert to original name and exit editing mode
    await expect.element(page.getByRole('heading', {name: 'Test Group', level: 3})).toBeInTheDocument();
  });

  it('should save name on blur', async () => {
    await renderWithProviders(<GroupEditPage />);

    await clickNameEditButton();

    const nameInput = page.getByRole('textbox');
    await expect.element(nameInput).toBeInTheDocument();

    await userEvent.clear(nameInput);
    await userEvent.fill(nameInput, 'Blur Name');
    await userEvent.tab(); // trigger blur

    await expect.element(page.getByText('You have unsaved changes')).toBeInTheDocument();
  });

  it('should enter description editing mode and save on Ctrl+Enter', async () => {
    await renderWithProviders(<GroupEditPage />);

    await clickDescriptionEditButton();

    // The autoFocused multiline textbox should appear with the current description
    const descInput = page.getByRole('textbox');
    await expect.element(descInput).toBeInTheDocument();

    await userEvent.clear(descInput);
    await userEvent.fill(descInput, 'Updated description');

    // Focus the input and press Ctrl+Enter
    await userEvent.click(descInput);
    await userEvent.keyboard('{Control>}{Enter}{/Control}');

    await expect.element(page.getByText('You have unsaved changes')).toBeInTheDocument();
  });

  it('should cancel description editing on Escape', async () => {
    await renderWithProviders(<GroupEditPage />);

    await clickDescriptionEditButton();

    const descInput = page.getByRole('textbox');
    await expect.element(descInput).toBeInTheDocument();

    await userEvent.clear(descInput);
    await userEvent.fill(descInput, 'Some new text');
    await userEvent.keyboard('{Escape}');

    // Should revert — the textbox should be gone and original text shown
    await expect.element(page.getByText('A test group')).toBeInTheDocument();
  });

  it('should save description on blur', async () => {
    await renderWithProviders(<GroupEditPage />);

    await clickDescriptionEditButton();

    const descInput = page.getByRole('textbox');
    await expect.element(descInput).toBeInTheDocument();

    await userEvent.clear(descInput);
    await userEvent.fill(descInput, 'Blurred desc');
    await userEvent.tab();

    await expect.element(page.getByText('You have unsaved changes')).toBeInTheDocument();
  });

  it('should show empty placeholder after clearing description', async () => {
    await renderWithProviders(<GroupEditPage />);

    await clickDescriptionEditButton();

    const descInput = page.getByRole('textbox');
    await expect.element(descInput).toBeInTheDocument();

    await userEvent.clear(descInput);
    await userEvent.tab();

    await expect.element(page.getByText('No description')).toBeInTheDocument();
  });

  it('should copy group ID to clipboard on click', async () => {
    await renderWithProviders(<GroupEditPage />);

    const groupIdElement = page.getByText('g1').element();
    const copyButton = groupIdElement.closest('[role="button"]');
    expect(copyButton).toBeTruthy();
    await userEvent.click(copyButton!);

    await vi.waitFor(() => {
      expect(mockWriteText).toHaveBeenCalledWith('g1');
    });
  });

  it('should copy group ID on keyboard Enter', async () => {
    await renderWithProviders(<GroupEditPage />);

    const groupIdElement = page.getByText('g1').element();
    const copyButton = groupIdElement.closest('[role="button"]');
    await userEvent.click(copyButton!);
    await userEvent.keyboard('{Enter}');

    expect(mockWriteText).toHaveBeenCalledWith('g1');
  });

  it('should copy group ID on keyboard Space', async () => {
    await renderWithProviders(<GroupEditPage />);

    const groupIdElement = page.getByText('g1').element();
    const copyButton = groupIdElement.closest('[role="button"]');
    await userEvent.click(copyButton!);
    await userEvent.keyboard(' ');

    expect(mockWriteText).toHaveBeenCalledWith('g1');
  });

  it('should save changes when save button is clicked', async () => {
    mockMutateAsync.mockResolvedValue(undefined);
    await renderWithProviders(<GroupEditPage />);

    // Edit the name to trigger hasChanges
    await clickNameEditButton();

    const nameInput = page.getByRole('textbox');
    await expect.element(nameInput).toBeInTheDocument();
    await userEvent.clear(nameInput);
    await userEvent.fill(nameInput, 'New Name');
    await userEvent.keyboard('{Enter}');

    await expect.element(page.getByText('You have unsaved changes')).toBeInTheDocument();

    await userEvent.click(page.getByText('Save Changes'));

    await vi.waitFor(() => {
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
    await renderWithProviders(<GroupEditPage />);

    // Edit the name to trigger hasChanges
    await clickNameEditButton();

    const nameInput = page.getByRole('textbox');
    await expect.element(nameInput).toBeInTheDocument();
    await userEvent.clear(nameInput);
    await userEvent.fill(nameInput, 'New Name');
    await userEvent.keyboard('{Enter}');

    await expect.element(page.getByText('Save Changes')).toBeInTheDocument();

    await userEvent.click(page.getByText('Save Changes'));

    await expect.element(page.getByText('Save failed')).toBeInTheDocument();
  });

  it('should reset changes when reset button is clicked', async () => {
    await renderWithProviders(<GroupEditPage />);

    // Edit the name
    await clickNameEditButton();

    const nameInput = page.getByRole('textbox');
    await expect.element(nameInput).toBeInTheDocument();
    await userEvent.clear(nameInput);
    await userEvent.fill(nameInput, 'New Name');
    await userEvent.keyboard('{Enter}');

    await expect.element(page.getByText('You have unsaved changes')).toBeInTheDocument();

    await userEvent.click(page.getByText('Reset'));

    await expect.element(page.getByText('You have unsaved changes')).not.toBeInTheDocument();
  });

  it('should navigate back from not found state', async () => {
    mockUseGetGroup.mockReturnValue({
      data: null,
      isLoading: false,
      error: null,
      refetch: vi.fn(),
    });
    await renderWithProviders(<GroupEditPage />);

    await userEvent.click(page.getByText('Back to Groups'));

    expect(mockNavigate).toHaveBeenCalledWith('/groups');
  });

  it('should handle navigate rejection from back button gracefully', async () => {
    mockNavigate.mockRejectedValue(new Error('Nav failed'));
    await renderWithProviders(<GroupEditPage />);

    const backLinks = page.getByText('Back to Groups').all();
    expect(backLinks.length).toBeGreaterThan(0);
    await userEvent.click(backLinks[0]);

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/groups');
    });
  });

  it('should close error snackbar when close button is clicked', async () => {
    mockMutateAsync.mockRejectedValue(new Error('Save failed'));
    await renderWithProviders(<GroupEditPage />);

    // Edit the name to trigger hasChanges
    await clickNameEditButton();

    const nameInput = page.getByRole('textbox');
    await expect.element(nameInput).toBeInTheDocument();
    await userEvent.clear(nameInput);
    await userEvent.fill(nameInput, 'New Name');
    await userEvent.keyboard('{Enter}');

    await expect.element(page.getByText('Save Changes')).toBeInTheDocument();

    await userEvent.click(page.getByText('Save Changes'));

    await expect.element(page.getByText('Save failed')).toBeInTheDocument();

    // Close the snackbar via the Alert's close button
    await userEvent.click(page.getByRole('button', {name: /close/i}));

    await expect.element(page.getByText('Save failed')).not.toBeInTheDocument();
  });
});
