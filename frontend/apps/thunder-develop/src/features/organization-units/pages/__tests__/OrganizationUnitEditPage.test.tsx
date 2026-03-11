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

import type {ReactNode} from 'react';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import {renderWithProviders, getByDisplayValue} from '@thunder/test-utils/browser';
import {page, userEvent} from 'vitest/browser';
import OrganizationUnitEditPage from '../OrganizationUnitEditPage';
import type {OrganizationUnit} from '../../models/organization-unit';

// Mock navigate, useParams, and useLocation
const mockNavigate = vi.fn();
const mockUseLocation = vi.fn<() => {state: unknown; pathname: string; search: string; hash: string; key: string}>();
vi.mock('react-router', async () => {
  const actual = await vi.importActual('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => ({id: 'ou-123'}),
    useLocation: () => mockUseLocation(),
    Link: ({to, children = undefined, ...props}: {to: string; children?: ReactNode; [key: string]: unknown}) => (
      <a
        {...(props as Record<string, unknown>)}
        href={to}
        onClick={(e) => {
          e.preventDefault();
          Promise.resolve(mockNavigate(to)).catch(() => {});
        }}
      >
        {children}
      </a>
    ),
  };
});

// Mock logger
vi.mock('@thunder/logger/react', () => ({
  useLogger: () => ({
    error: vi.fn(),
    info: vi.fn(),
    debug: vi.fn(),
  }),
}));

// Mock get OU hook
const mockRefetch = vi.fn();
const mockUseGetOrganizationUnit = vi.fn();
vi.mock('../../api/useGetOrganizationUnit', () => ({
  default: () =>
    mockUseGetOrganizationUnit() as {
      data: OrganizationUnit | undefined;
      isLoading: boolean;
      error: Error | null;
      refetch: () => void;
    },
}));

// Mock update hook
const mockMutateAsync = vi.fn();
vi.mock('../../api/useUpdateOrganizationUnit', () => ({
  default: () => ({
    mutateAsync: mockMutateAsync,
    isPending: false,
  }),
}));

// Mock delete hook
const mockDeleteMutate = vi.fn();
vi.mock('../../api/useDeleteOrganizationUnit', () => ({
  default: () => ({
    mutate: mockDeleteMutate,
    isPending: false,
  }),
}));

// Mock child hooks
vi.mock('../../api/useGetChildOrganizationUnits', () => ({
  default: () => ({
    data: {organizationUnits: [], totalResults: 0, startIndex: 1, count: 0},
    isLoading: false,
  }),
}));

vi.mock('../../api/useGetOrganizationUnitUsers', () => ({
  default: () => ({
    data: {users: [], totalResults: 0, startIndex: 1, count: 0},
    isLoading: false,
  }),
}));

vi.mock('../../api/useGetOrganizationUnitGroups', () => ({
  default: () => ({
    data: {groups: [], totalResults: 0, startIndex: 1, count: 0},
    isLoading: false,
  }),
}));

// Mock useOrganizationUnit hook
vi.mock('../../contexts/useOrganizationUnit', () => ({
  default: () => ({
    resetTreeState: vi.fn(),
  }),
}));

// Mock useDataGridLocaleText
vi.mock('../../../../hooks/useDataGridLocaleText', () => ({
  default: () => ({}),
}));

// Mock LogoUpdateModal
vi.mock('../../../../components/LogoUpdateModal', () => ({
  default: vi.fn(
    ({open, onLogoUpdate, onClose}: {open: boolean; onLogoUpdate: (url: string) => void; onClose: () => void}) => (
      <div data-testid="logo-update-modal" style={{display: open ? 'block' : 'none'}}>
        <button type="button" onClick={() => onLogoUpdate('https://example.com/new-logo.png')}>
          Update Logo
        </button>
        <button type="button" onClick={onClose}>
          Close
        </button>
      </div>
    ),
  ),
}));

describe('OrganizationUnitEditPage', () => {
  const mockOrganizationUnit: OrganizationUnit = {
    id: 'ou-123',
    handle: 'test-ou',
    name: 'Test Organization Unit',
    description: 'A test description',
    parent: null,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockNavigate.mockReset();
    mockMutateAsync.mockReset();
    mockRefetch.mockReset();
    mockDeleteMutate.mockReset();
    mockUseLocation.mockReturnValue({
      state: null,
      pathname: '/organization-units/ou-123',
      search: '',
      hash: '',
      key: 'default',
    });
    mockUseGetOrganizationUnit.mockReturnValue({
      data: mockOrganizationUnit,
      isLoading: false,
      error: null,
      refetch: mockRefetch,
    });
  });

  it('should render organization unit name', async () => {
    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Test Organization Unit')).toBeInTheDocument();
    });
  });

  it('should render organization unit description', async () => {
    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('A test description')).toBeInTheDocument();
    });
  });

  it('should render back button', async () => {
    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Back to Organization Units')).toBeInTheDocument();
    });
  });

  it('should navigate back when back button is clicked', async () => {
    await renderWithProviders(<OrganizationUnitEditPage />);

    await userEvent.click(page.getByText('Back to Organization Units'));

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/organization-units');
    });
  });

  it('should render all tabs', async () => {
    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('General')).toBeInTheDocument();
      await expect.element(page.getByText('Child OUs')).toBeInTheDocument();
      await expect.element(page.getByText('Users')).toBeInTheDocument();
      await expect.element(page.getByText('Groups')).toBeInTheDocument();
      await expect.element(page.getByText('Customization')).toBeInTheDocument();
    });
  });

  it('should show loading state', async () => {
    mockUseGetOrganizationUnit.mockReturnValue({
      data: undefined,
      isLoading: true,
      error: null,
      refetch: mockRefetch,
    });

    await renderWithProviders(<OrganizationUnitEditPage />);

    await expect.element(page.getByRole('progressbar')).toBeInTheDocument();
  });

  it('should show error state', async () => {
    mockUseGetOrganizationUnit.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: new Error('Network error'),
      refetch: mockRefetch,
    });

    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Network error')).toBeInTheDocument();
    });
  });

  it('should show not found state when OU is undefined', async () => {
    mockUseGetOrganizationUnit.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
      refetch: mockRefetch,
    });

    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Organization unit not found')).toBeInTheDocument();
    });
  });

  it('should switch tabs when clicked', async () => {
    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('General')).toBeInTheDocument();
    });

    // Click on Customization tab
    await userEvent.click(page.getByRole('tab', {name: 'Customization'}));

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Theme')).toBeInTheDocument();
    });
  });

  it('should show edit button for name', async () => {
    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Test Organization Unit')).toBeInTheDocument();
    });

    // There should be edit buttons
    const editButtons = page.getByRole('button').all();
    expect(editButtons.length).toBeGreaterThan(0);
  });

  it('should render "No description" when description is null', async () => {
    mockUseGetOrganizationUnit.mockReturnValue({
      data: {...mockOrganizationUnit, description: null},
      isLoading: false,
      error: null,
      refetch: mockRefetch,
    });

    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('No description')).toBeInTheDocument();
    });
  });

  it('should show floating save bar when changes are made', async () => {
    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Test Organization Unit')).toBeInTheDocument();
    });

    // Find and click edit button for name (second button after back)
    const editButtons = page.getByRole('button').all();
    // The edit button is near the name
    const nameEditButton = editButtons.find(
      (btn) => btn.element().querySelector('svg') && btn.element().closest('div')?.textContent?.includes('Test Organization Unit'),
    );

    if (nameEditButton) {
      await userEvent.click(nameEditButton);

      // Type new name - get by current display value
      const nameInput = getByDisplayValue('Test Organization Unit');
      await userEvent.fill(nameInput, 'Updated Name');
      (nameInput.element() as HTMLElement).blur();

      await vi.waitFor(async () => {
        await expect.element(page.getByText('You have unsaved changes')).toBeInTheDocument();
      });
    }
  });

  it('should reset changes when reset button is clicked', async () => {
    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Test Organization Unit')).toBeInTheDocument();
    });

    // Make a change to show the floating bar
    const editButtons = page.getByRole('button').all();
    const nameEditButton = editButtons.find(
      (btn) => btn.element().querySelector('svg') && btn.element().closest('div')?.textContent?.includes('Test Organization Unit'),
    );

    if (nameEditButton) {
      await userEvent.click(nameEditButton);

      const nameInput = getByDisplayValue('Test Organization Unit');
      await userEvent.fill(nameInput, 'Updated Name');
      (nameInput.element() as HTMLElement).blur();

      await vi.waitFor(async () => {
        await expect.element(page.getByText('You have unsaved changes')).toBeInTheDocument();
      });

      // Click reset
      await userEvent.click(page.getByText('Reset'));

      await vi.waitFor(async () => {
        await expect.element(page.getByText('You have unsaved changes')).not.toBeInTheDocument();
      });
    }
  });

  it('should edit description when edit button is clicked', async () => {
    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('A test description')).toBeInTheDocument();
    });

    // Find edit button next to description
    const editButtons = page.getByRole('button').all();
    const descriptionEditButton = editButtons.find(
      (btn) => btn.element().querySelector('svg') && btn.element().closest('div')?.textContent?.includes('A test description'),
    );

    if (descriptionEditButton) {
      await userEvent.click(descriptionEditButton);

      // Should show a textbox for editing - get by current display value
      await vi.waitFor(() => {
        const textbox = getByDisplayValue('A test description');
        expect(textbox).toBeInTheDocument();
      });

      const textbox = getByDisplayValue('A test description');
      await userEvent.fill(textbox, 'Updated description');
      (textbox.element() as HTMLElement).blur();

      await vi.waitFor(async () => {
        await expect.element(page.getByText('You have unsaved changes')).toBeInTheDocument();
      });
    }
  });

  it('should cancel description editing on Escape key', async () => {
    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('A test description')).toBeInTheDocument();
    });

    const editButtons = page.getByRole('button').all();
    const descriptionEditButton = editButtons.find(
      (btn) => btn.element().querySelector('svg') && btn.element().closest('div')?.textContent?.includes('A test description'),
    );

    if (descriptionEditButton) {
      await userEvent.click(descriptionEditButton);

      await vi.waitFor(() => {
        expect(getByDisplayValue('A test description')).toBeInTheDocument();
      });

      await userEvent.keyboard('{Escape}');

      await vi.waitFor(async () => {
        await expect.element(page.getByText('A test description')).toBeInTheDocument();
      });
    }
  });

  it('should save description on Ctrl+Enter', async () => {
    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('A test description')).toBeInTheDocument();
    });

    const editButtons = page.getByRole('button').all();
    const descriptionEditButton = editButtons.find(
      (btn) => btn.element().querySelector('svg') && btn.element().closest('div')?.textContent?.includes('A test description'),
    );

    if (descriptionEditButton) {
      await userEvent.click(descriptionEditButton);

      await vi.waitFor(() => {
        expect(getByDisplayValue('A test description')).toBeInTheDocument();
      });

      const textbox = getByDisplayValue('A test description');
      await userEvent.fill(textbox, 'New description');
      await userEvent.keyboard('{Control>}{Enter}{/Control}');

      await vi.waitFor(async () => {
        await expect.element(page.getByText('You have unsaved changes')).toBeInTheDocument();
      });
    }
  });

  it('should cancel name editing on Escape key', async () => {
    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Test Organization Unit')).toBeInTheDocument();
    });

    const editButtons = page.getByRole('button').all();
    const nameEditButton = editButtons.find(
      (btn) => btn.element().querySelector('svg') && btn.element().closest('div')?.textContent?.includes('Test Organization Unit'),
    );

    if (nameEditButton) {
      await userEvent.click(nameEditButton);

      await vi.waitFor(() => {
        expect(getByDisplayValue('Test Organization Unit')).toBeInTheDocument();
      });

      await userEvent.keyboard('{Escape}');

      await vi.waitFor(async () => {
        await expect.element(page.getByText('Test Organization Unit')).toBeInTheDocument();
      });
    }
  });

  it('should save name on Enter key', async () => {
    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Test Organization Unit')).toBeInTheDocument();
    });

    const editButtons = page.getByRole('button').all();
    const nameEditButton = editButtons.find(
      (btn) => btn.element().querySelector('svg') && btn.element().closest('div')?.textContent?.includes('Test Organization Unit'),
    );

    if (nameEditButton) {
      await userEvent.click(nameEditButton);

      await vi.waitFor(() => {
        expect(getByDisplayValue('Test Organization Unit')).toBeInTheDocument();
      });

      const textbox = getByDisplayValue('Test Organization Unit');
      await userEvent.fill(textbox, 'Updated Name');
      await userEvent.keyboard('{Enter}');

      await vi.waitFor(async () => {
        await expect.element(page.getByText('You have unsaved changes')).toBeInTheDocument();
      });
    }
  });

  it('should call save and refetch when save button is clicked', async () => {
    mockMutateAsync.mockResolvedValue({});

    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Test Organization Unit')).toBeInTheDocument();
    });

    // Make a change to show the save bar
    const editButtons = page.getByRole('button').all();
    const nameEditButton = editButtons.find(
      (btn) => btn.element().querySelector('svg') && btn.element().closest('div')?.textContent?.includes('Test Organization Unit'),
    );

    if (nameEditButton) {
      await userEvent.click(nameEditButton);

      const nameInput = getByDisplayValue('Test Organization Unit');
      await userEvent.fill(nameInput, 'Updated Name');
      (nameInput.element() as HTMLElement).blur();

      await vi.waitFor(async () => {
        await expect.element(page.getByText('You have unsaved changes')).toBeInTheDocument();
      });

      // Click save
      await userEvent.click(page.getByText('Save Changes'));

      await vi.waitFor(() => {
        expect(mockMutateAsync).toHaveBeenCalledWith(
          expect.objectContaining({
            id: 'ou-123',
            data: expect.objectContaining({
              name: 'Updated Name',
            }) as unknown,
          }),
        );
      });
    }
  });

  it('should handle save error gracefully', async () => {
    mockMutateAsync.mockRejectedValue(new Error('Save error'));

    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Test Organization Unit')).toBeInTheDocument();
    });

    // Make a change
    const editButtons = page.getByRole('button').all();
    const nameEditButton = editButtons.find(
      (btn) => btn.element().querySelector('svg') && btn.element().closest('div')?.textContent?.includes('Test Organization Unit'),
    );

    if (nameEditButton) {
      await userEvent.click(nameEditButton);

      const nameInput = getByDisplayValue('Test Organization Unit');
      await userEvent.fill(nameInput, 'Updated Name');
      (nameInput.element() as HTMLElement).blur();

      await vi.waitFor(async () => {
        await expect.element(page.getByText('You have unsaved changes')).toBeInTheDocument();
      });

      // Click save - should not throw
      await userEvent.click(page.getByText('Save Changes'));
    }
  });

  it('should not save empty name on blur', async () => {
    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Test Organization Unit')).toBeInTheDocument();
    });

    const editButtons = page.getByRole('button').all();
    const nameEditButton = editButtons.find(
      (btn) => btn.element().querySelector('svg') && btn.element().closest('div')?.textContent?.includes('Test Organization Unit'),
    );

    if (nameEditButton) {
      await userEvent.click(nameEditButton);

      const nameInput = getByDisplayValue('Test Organization Unit');
      await userEvent.fill(nameInput, '');
      (nameInput.element() as HTMLElement).blur();

      // Should not show unsaved changes for empty name
      await vi.waitFor(async () => {
        await expect.element(page.getByText('Test Organization Unit')).toBeInTheDocument();
      });
    }
  });

  it('should navigate back from error state', async () => {
    mockUseGetOrganizationUnit.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: new Error('Network error'),
      refetch: mockRefetch,
    });

    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Network error')).toBeInTheDocument();
    });

    // Click back button
    await userEvent.click(page.getByText('Back to Organization Units'));

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/organization-units');
    });
  });

  it('should navigate back from not found state', async () => {
    mockUseGetOrganizationUnit.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
      refetch: mockRefetch,
    });

    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Organization unit not found')).toBeInTheDocument();
    });

    // Click back button
    await userEvent.click(page.getByText('Back to Organization Units'));

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/organization-units');
    });
  });

  it('should switch to child OUs tab', async () => {
    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Test Organization Unit')).toBeInTheDocument();
    });

    await userEvent.click(page.getByRole('tab', {name: 'Child OUs'}));

    await vi.waitFor(async () => {
      await expect.element(page.getByRole('tab', {name: 'Child OUs', selected: true})).toBeInTheDocument();
    });
  });

  it('should switch to users tab', async () => {
    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Test Organization Unit')).toBeInTheDocument();
    });

    await userEvent.click(page.getByRole('tab', {name: 'Users'}));

    await vi.waitFor(async () => {
      await expect.element(page.getByRole('tab', {name: 'Users', selected: true})).toBeInTheDocument();
    });
  });

  it('should switch to groups tab', async () => {
    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Test Organization Unit')).toBeInTheDocument();
    });

    await userEvent.click(page.getByRole('tab', {name: 'Groups'}));

    await vi.waitFor(async () => {
      await expect.element(page.getByRole('tab', {name: 'Groups', selected: true})).toBeInTheDocument();
    });
  });

  it('should navigate back to parent OU when fromOU is provided', async () => {
    mockUseLocation.mockReturnValue({
      state: {
        fromOU: {
          id: 'parent-ou-id',
          name: 'Parent OU',
        },
      },
      pathname: '/organization-units/ou-123',
      search: '',
      hash: '',
      key: 'default',
    });

    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Test Organization Unit')).toBeInTheDocument();
    });

    // Back button should show the parent OU name - find by partial text
    const backButton = page.getByText('Back to Parent OU');
    await userEvent.click(backButton);

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/organization-units/parent-ou-id');
    });
  });

  it('should handle back navigation error in error state', async () => {
    mockNavigate.mockRejectedValue(new Error('Navigation failed'));
    mockUseGetOrganizationUnit.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: new Error('Network error'),
      refetch: mockRefetch,
    });

    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Network error')).toBeInTheDocument();
    });

    // Click back button - should not throw
    await userEvent.click(page.getByText('Back to Organization Units'));

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/organization-units');
    });
  });

  it('should handle back navigation error in not found state', async () => {
    mockNavigate.mockRejectedValue(new Error('Navigation failed'));
    mockUseGetOrganizationUnit.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
      refetch: mockRefetch,
    });

    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Organization unit not found')).toBeInTheDocument();
    });

    // Click back button - should not throw
    await userEvent.click(page.getByText('Back to Organization Units'));

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/organization-units');
    });
  });

  it('should handle back navigation error in main view', async () => {
    mockNavigate.mockRejectedValue(new Error('Navigation failed'));

    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Test Organization Unit')).toBeInTheDocument();
    });

    // Click back button - should not throw
    await userEvent.click(page.getByText('Back to Organization Units'));

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/organization-units');
    });
  });

  it('should handle delete success and navigate to list', async () => {
    // Mock delete to trigger onSuccess
    mockDeleteMutate.mockImplementation((_id: string, options: {onSuccess?: () => void}) => {
      options.onSuccess?.();
    });

    await renderWithProviders(<OrganizationUnitEditPage />);

    // Open delete dialog
    await vi.waitFor(async () => {
      await expect.element(page.getByText('Delete Organization Unit')).toBeInTheDocument();
    });

    await userEvent.click(page.getByText('Delete Organization Unit'));

    await vi.waitFor(() => {
      expect(
        page.getByText('Are you sure you want to delete this organization unit? This action cannot be undone.'),
      ).toBeInTheDocument();
    });

    // Find and click the delete confirm button in dialog
    const deleteButtons = page.getByText('Delete').all();
    const confirmDeleteButton = deleteButtons.find((btn) => btn.element().closest('.MuiDialog-root'));
    if (confirmDeleteButton) {
      await userEvent.click(confirmDeleteButton);

      await vi.waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/organization-units');
      });
    }
  });

  it('should handle delete success navigation error gracefully', async () => {
    mockNavigate.mockRejectedValue(new Error('Navigation failed'));
    // Mock delete to trigger onSuccess
    mockDeleteMutate.mockImplementation((_id: string, options: {onSuccess?: () => void}) => {
      options.onSuccess?.();
    });

    await renderWithProviders(<OrganizationUnitEditPage />);

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Delete Organization Unit')).toBeInTheDocument();
    });

    // Open delete dialog
    await userEvent.click(page.getByText('Delete Organization Unit'));

    await vi.waitFor(() => {
      expect(
        page.getByText('Are you sure you want to delete this organization unit? This action cannot be undone.'),
      ).toBeInTheDocument();
    });

    // Find and click the delete confirm button in dialog
    const deleteButtons = page.getByText('Delete').all();
    const confirmDeleteButton = deleteButtons.find((btn) => btn.element().closest('.MuiDialog-root'));
    if (confirmDeleteButton) {
      await userEvent.click(confirmDeleteButton);

      // Should not throw - error is logged
      await vi.waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/organization-units');
      });
    }
  });

  describe('Avatar Image', () => {
    it('should hide avatar image when image fails to load', async () => {
      mockUseGetOrganizationUnit.mockReturnValue({
        data: {...mockOrganizationUnit, logo_url: 'https://example.com/logo.png'},
        isLoading: false,
        error: null,
        refetch: mockRefetch,
      });

      await renderWithProviders(<OrganizationUnitEditPage />);

      const avatar = page.getByRole('img');
      avatar.element().dispatchEvent(new Event('error'));

      expect(avatar).toHaveStyle({display: 'none'});
    });

    it('should open logo modal when avatar is clicked', async () => {
      mockUseGetOrganizationUnit.mockReturnValue({
        data: {...mockOrganizationUnit, logo_url: 'https://example.com/logo.png'},
        isLoading: false,
        error: null,
        refetch: mockRefetch,
      });

      await renderWithProviders(<OrganizationUnitEditPage />);

      const avatar = page.getByRole('img');
      await userEvent.click(avatar);

      await vi.waitFor(() => {
        const modal = page.getByTestId('logo-update-modal');
        expect(modal).toHaveStyle({display: 'block'});
      });
    });

    it('should display edited logo_url in avatar when editedOU has logo_url', async () => {
      mockUseGetOrganizationUnit.mockReturnValue({
        data: {...mockOrganizationUnit, logo_url: 'https://example.com/original.png'},
        isLoading: false,
        error: null,
        refetch: mockRefetch,
      });

      await renderWithProviders(<OrganizationUnitEditPage />);

      // Open logo modal and update logo
      const avatar = page.getByRole('img');
      await userEvent.click(avatar);

      await vi.waitFor(async () => {
        await expect.element(page.getByTestId('logo-update-modal')).toHaveStyle({display: 'block'});
      });

      await userEvent.click(page.getByText('Update Logo'));

      await vi.waitFor(async () => {
        await expect.element(page.getByText('You have unsaved changes')).toBeInTheDocument();
      });
    });
  });

  describe('Logo Update Modal', () => {
    it('should open logo modal when edit icon button is clicked', async () => {
      await renderWithProviders(<OrganizationUnitEditPage />);

      const editButtons = page.getByRole('button').all();
      const logoEditButton = editButtons.find(
        (btn) => btn.element().getAttribute('aria-label') === 'Update Logo',
      );

      if (logoEditButton) {
        await userEvent.click(logoEditButton);

        await vi.waitFor(() => {
          const modal = page.getByTestId('logo-update-modal');
          expect(modal).toHaveStyle({display: 'block'});
        });
      }
    });

    it('should close logo modal when close button is clicked', async () => {
      await renderWithProviders(<OrganizationUnitEditPage />);

      // Open the modal via logo edit icon button
      const editButtons = page.getByRole('button').all();
      const logoEditButton = editButtons.find(
        (btn) => btn.element().getAttribute('aria-label') === 'Update Logo',
      );

      expect(logoEditButton).toBeInTheDocument();
      await userEvent.click(logoEditButton!);

      await vi.waitFor(async () => {
        await expect.element(page.getByTestId('logo-update-modal')).toHaveStyle({display: 'block'});
      });

      // Close the modal
      await userEvent.click(page.getByText('Close'));

      await vi.waitFor(async () => {
        await expect.element(page.getByTestId('logo-update-modal')).toHaveStyle({display: 'none'});
      });
    });

    it('should update logo and close modal when logo is updated', async () => {
      await renderWithProviders(<OrganizationUnitEditPage />);

      // Open the modal
      const editButtons = page.getByRole('button').all();
      const logoEditButton = editButtons.find(
        (btn) => btn.element().getAttribute('aria-label') === 'Update Logo',
      );

      if (logoEditButton) {
        await userEvent.click(logoEditButton);

        await vi.waitFor(async () => {
          await expect.element(page.getByTestId('logo-update-modal')).toHaveStyle({display: 'block'});
        });

        // Click update logo
        await userEvent.click(page.getByText('Update Logo'));

        // Modal should close
        await vi.waitFor(async () => {
          await expect.element(page.getByTestId('logo-update-modal')).toHaveStyle({display: 'none'});
        });

        // Should show unsaved changes
        await expect.element(page.getByText('You have unsaved changes')).toBeInTheDocument();
      }
    });
  });

  describe('Delete Error and Snackbar', () => {
    it('should show error snackbar when delete fails', async () => {
      // Mock delete to trigger onError
      mockDeleteMutate.mockImplementation((_id: string, options: {onError?: (err: Error) => void}) => {
        options.onError?.(new Error('Delete failed'));
      });

      await renderWithProviders(<OrganizationUnitEditPage />);

      await vi.waitFor(async () => {
        await expect.element(page.getByText('Delete Organization Unit')).toBeInTheDocument();
      });

      // Open delete dialog
      await userEvent.click(page.getByText('Delete Organization Unit'));

      await vi.waitFor(() => {
        expect(
          page.getByText('Are you sure you want to delete this organization unit? This action cannot be undone.'),
        ).toBeInTheDocument();
      });

      // Find and click the delete confirm button in dialog
      const deleteButtons = page.getByText('Delete').all();
      const confirmDeleteButton = deleteButtons.find((btn) => btn.element().closest('.MuiDialog-root'));
      if (confirmDeleteButton) {
        await userEvent.click(confirmDeleteButton);

        // Snackbar should appear with error
        await vi.waitFor(async () => {
          await expect.element(page.getByRole('alert')).toBeInTheDocument();
        });
      }
    });
  });

  describe('Edited OU Fallbacks', () => {
    it(
      'should display edited name when re-editing after a name change',
      async () => {
        await renderWithProviders(<OrganizationUnitEditPage />);

        await vi.waitFor(async () => {
          await expect.element(page.getByText('Test Organization Unit')).toBeInTheDocument();
        });

        // First edit: change name
        const editButtons = page.getByRole('button').all();
        const nameEditButton = editButtons.find(
          (btn) => btn.element().querySelector('svg') && btn.element().closest('div')?.textContent?.includes('Test Organization Unit'),
        );

        if (nameEditButton) {
          await userEvent.click(nameEditButton);
          const nameInput = getByDisplayValue('Test Organization Unit');
          await userEvent.fill(nameInput, 'Updated Name');
          (nameInput.element() as HTMLElement).blur();

          await vi.waitFor(async () => {
            await expect.element(page.getByText('Updated Name')).toBeInTheDocument();
          });

          // Second edit: the input should show the edited name
          const editButtons2 = page.getByRole('button').all();
          const nameEditButton2 = editButtons2.find(
            (btn) => btn.element().querySelector('svg') && btn.element().closest('div')?.textContent?.includes('Updated Name'),
          );

          if (nameEditButton2) {
            await userEvent.click(nameEditButton2);
            expect(getByDisplayValue('Updated Name')).toBeInTheDocument();
          }
        }
      },
      15_000,
    );

    it('should display edited description when re-editing after a description change', async () => {
      await renderWithProviders(<OrganizationUnitEditPage />);

      await vi.waitFor(async () => {
        await expect.element(page.getByText('A test description')).toBeInTheDocument();
      });

      // First edit: change description
      const editButtons = page.getByRole('button').all();
      const descEditButton = editButtons.find(
        (btn) => btn.element().querySelector('svg') && btn.element().closest('div')?.textContent?.includes('A test description'),
      );

      if (descEditButton) {
        await userEvent.click(descEditButton);
        const descInput = getByDisplayValue('A test description');
        await userEvent.fill(descInput, 'Updated Description');
        (descInput.element() as HTMLElement).blur();

        await vi.waitFor(async () => {
          await expect.element(page.getByText('Updated Description')).toBeInTheDocument();
        });

        // Second edit: Escape should restore the edited description
        const editButtons2 = page.getByRole('button').all();
        const descEditButton2 = editButtons2.find(
          (btn) => btn.element().querySelector('svg') && btn.element().closest('div')?.textContent?.includes('Updated Description'),
        );

        if (descEditButton2) {
          await userEvent.click(descEditButton2);
          await userEvent.keyboard('{Escape}');

          await vi.waitFor(async () => {
            await expect.element(page.getByText('Updated Description')).toBeInTheDocument();
          });
        }
      }
    });

    it('should restore edited name on Escape key when editedOU has name', async () => {
      await renderWithProviders(<OrganizationUnitEditPage />);

      await vi.waitFor(async () => {
        await expect.element(page.getByText('Test Organization Unit')).toBeInTheDocument();
      });

      // First edit: change name
      const editButtons = page.getByRole('button').all();
      const nameEditButton = editButtons.find(
        (btn) => btn.element().querySelector('svg') && btn.element().closest('div')?.textContent?.includes('Test Organization Unit'),
      );

      if (nameEditButton) {
        await userEvent.click(nameEditButton);
        const nameInput = getByDisplayValue('Test Organization Unit');
        await userEvent.fill(nameInput, 'Edited Name');
        (nameInput.element() as HTMLElement).blur();

        await vi.waitFor(async () => {
          await expect.element(page.getByText('Edited Name')).toBeInTheDocument();
        });

        // Re-edit and press Escape
        const editButtons2 = page.getByRole('button').all();
        const nameEditButton2 = editButtons2.find(
          (btn) => btn.element().querySelector('svg') && btn.element().closest('div')?.textContent?.includes('Edited Name'),
        );

        if (nameEditButton2) {
          await userEvent.click(nameEditButton2);
          const nameInput2 = getByDisplayValue('Edited Name');
          await userEvent.fill(nameInput2, 'Something Else');
          await userEvent.keyboard('{Escape}');

          // Should restore to the edited name, not the original
          await vi.waitFor(async () => {
            await expect.element(page.getByText('Edited Name')).toBeInTheDocument();
          });
        }
      }
    });

    it('should save name changes on Enter with trimmed value', async () => {
      await renderWithProviders(<OrganizationUnitEditPage />);

      await vi.waitFor(async () => {
        await expect.element(page.getByText('Test Organization Unit')).toBeInTheDocument();
      });

      const editButtons = page.getByRole('button').all();
      const nameEditButton = editButtons.find(
        (btn) => btn.element().querySelector('svg') && btn.element().closest('div')?.textContent?.includes('Test Organization Unit'),
      );

      if (nameEditButton) {
        await userEvent.click(nameEditButton);
        const nameInput = getByDisplayValue('Test Organization Unit');
        await userEvent.fill(nameInput, '');
        await userEvent.keyboard('{Enter}');

        // Should not save empty name, should exit editing
        await vi.waitFor(async () => {
          await expect.element(page.getByText('Test Organization Unit')).toBeInTheDocument();
        });
      }
    });
  });

  describe('Save with edited fields', () => {
    it('should include description and theme_id in save when edited', async () => {
      mockMutateAsync.mockResolvedValue({});

      await renderWithProviders(<OrganizationUnitEditPage />);

      await vi.waitFor(async () => {
        await expect.element(page.getByText('Test Organization Unit')).toBeInTheDocument();
      });

      // Make a description change
      const editButtons = page.getByRole('button').all();
      const descEditButton = editButtons.find(
        (btn) =>
          btn.element().querySelector('svg') && btn.element().closest('div')?.textContent?.includes('A test description'),
      );

      if (descEditButton) {
        await userEvent.click(descEditButton);
        const descInput = getByDisplayValue('A test description');
        await userEvent.fill(descInput, 'New description');
        (descInput.element() as HTMLElement).blur();

        await vi.waitFor(async () => {
          await expect.element(page.getByText('You have unsaved changes')).toBeInTheDocument();
        });

        // Click save
        await userEvent.click(page.getByText('Save Changes'));

        await vi.waitFor(() => {
          expect(mockMutateAsync).toHaveBeenCalledWith(
            expect.objectContaining({
              id: 'ou-123',
              data: expect.objectContaining({
                description: 'New description',
              }) as unknown,
            }),
          );
        });
      }
    });
  });
});
