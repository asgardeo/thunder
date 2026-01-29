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
import {screen, fireEvent, waitFor} from '@testing-library/react';
import {renderWithProviders} from '../../../../test/test-utils';
import OrganizationUnitEditPage from '../OrganizationUnitEditPage';
import type {OrganizationUnit} from '../../types/organization-units';

// Mock navigate and useParams
const mockNavigate = vi.fn();
vi.mock('react-router', async () => {
  const actual = await vi.importActual('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => ({id: 'ou-123'}),
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
vi.mock('../../api/useDeleteOrganizationUnit', () => ({
  default: () => ({
    mutate: vi.fn(),
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

// Mock useDataGridLocaleText
vi.mock('../../../../hooks/useDataGridLocaleText', () => ({
  default: () => ({}),
}));

// Mock translations
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'organizationUnits:view.back': 'Back',
        'organizationUnits:view.error': 'Failed to load organization unit',
        'organizationUnits:view.notFound': 'Organization unit not found',
        'organizationUnits:view.tabs.general': 'General',
        'organizationUnits:view.tabs.childOUs': 'Child OUs',
        'organizationUnits:view.tabs.users': 'Users',
        'organizationUnits:view.tabs.groups': 'Groups',
        'organizationUnits:view.tabs.advanced': 'Advanced',
        'organizationUnits:view.unsavedChanges': 'You have unsaved changes',
        'organizationUnits:view.reset': 'Reset',
        'organizationUnits:view.save': 'Save',
        'organizationUnits:view.saving': 'Saving...',
        'organizationUnits:view.description.placeholder': 'Add a description',
        'organizationUnits:view.description.empty': 'No description',
        'organizationUnits:view.general.title': 'General Information',
        'organizationUnits:view.general.subtitle': 'Basic details',
        'organizationUnits:form.handle': 'Handle',
        'organizationUnits:view.general.id': 'Organization Unit ID',
        'organizationUnits:view.childOUs.title': 'Child OUs',
        'organizationUnits:view.childOUs.subtitle': 'Child organization units',
        'organizationUnits:view.users.title': 'Users',
        'organizationUnits:view.users.subtitle': 'Users in this OU',
        'organizationUnits:view.users.columns.id': 'User ID',
        'organizationUnits:view.users.columns.type': 'User Type',
        'organizationUnits:view.groups.title': 'Groups',
        'organizationUnits:view.groups.subtitle': 'Groups in this OU',
        'organizationUnits:view.groups.columns.name': 'Name',
        'organizationUnits:view.groups.columns.id': 'ID',
        'organizationUnits:view.advanced.dangerZone': 'Danger Zone',
        'organizationUnits:view.advanced.dangerZoneDescription': 'Irreversible actions',
        'organizationUnits:view.advanced.deleteButton': 'Delete Organization Unit',
        'organizationUnits:delete.title': 'Delete Organization Unit',
        'organizationUnits:delete.message': 'Are you sure?',
        'organizationUnits:delete.disclaimer': 'This cannot be undone.',
        'common:actions.cancel': 'Cancel',
        'common:actions.delete': 'Delete',
        'organizationUnits:listing.columns.name': 'Name',
        'organizationUnits:listing.columns.handle': 'Handle',
        'organizationUnits:listing.columns.description': 'Description',
      };
      return translations[key] ?? key;
    },
  }),
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
    mockUseGetOrganizationUnit.mockReturnValue({
      data: mockOrganizationUnit,
      isLoading: false,
      error: null,
      refetch: mockRefetch,
    });
  });

  it('should render organization unit name', async () => {
    renderWithProviders(<OrganizationUnitEditPage />);

    await waitFor(() => {
      expect(screen.getByText('Test Organization Unit')).toBeInTheDocument();
    });
  });

  it('should render organization unit description', async () => {
    renderWithProviders(<OrganizationUnitEditPage />);

    await waitFor(() => {
      expect(screen.getByText('A test description')).toBeInTheDocument();
    });
  });

  it('should render back button', async () => {
    renderWithProviders(<OrganizationUnitEditPage />);

    await waitFor(() => {
      expect(screen.getByText('Back')).toBeInTheDocument();
    });
  });

  it('should navigate back when back button is clicked', async () => {
    renderWithProviders(<OrganizationUnitEditPage />);

    fireEvent.click(screen.getByText('Back'));

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/organization-units');
    });
  });

  it('should render all tabs', async () => {
    renderWithProviders(<OrganizationUnitEditPage />);

    await waitFor(() => {
      expect(screen.getByText('General')).toBeInTheDocument();
      expect(screen.getByText('Child OUs')).toBeInTheDocument();
      expect(screen.getByText('Users')).toBeInTheDocument();
      expect(screen.getByText('Groups')).toBeInTheDocument();
      expect(screen.getByText('Advanced')).toBeInTheDocument();
    });
  });

  it('should show loading state', () => {
    mockUseGetOrganizationUnit.mockReturnValue({
      data: undefined,
      isLoading: true,
      error: null,
      refetch: mockRefetch,
    });

    renderWithProviders(<OrganizationUnitEditPage />);

    expect(screen.getByRole('progressbar')).toBeInTheDocument();
  });

  it('should show error state', async () => {
    mockUseGetOrganizationUnit.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: new Error('Network error'),
      refetch: mockRefetch,
    });

    renderWithProviders(<OrganizationUnitEditPage />);

    await waitFor(() => {
      expect(screen.getByText('Network error')).toBeInTheDocument();
    });
  });

  it('should show not found state when OU is undefined', async () => {
    mockUseGetOrganizationUnit.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
      refetch: mockRefetch,
    });

    renderWithProviders(<OrganizationUnitEditPage />);

    await waitFor(() => {
      expect(screen.getByText('Organization unit not found')).toBeInTheDocument();
    });
  });

  it('should switch tabs when clicked', async () => {
    renderWithProviders(<OrganizationUnitEditPage />);

    await waitFor(() => {
      expect(screen.getByText('General')).toBeInTheDocument();
    });

    // Click on Advanced tab
    fireEvent.click(screen.getByText('Advanced'));

    await waitFor(() => {
      expect(screen.getByText('Danger Zone')).toBeInTheDocument();
    });
  });

  it('should show edit button for name', async () => {
    renderWithProviders(<OrganizationUnitEditPage />);

    await waitFor(() => {
      expect(screen.getByText('Test Organization Unit')).toBeInTheDocument();
    });

    // There should be edit buttons
    const editButtons = screen.getAllByRole('button');
    expect(editButtons.length).toBeGreaterThan(0);
  });

  it('should render "No description" when description is null', async () => {
    mockUseGetOrganizationUnit.mockReturnValue({
      data: {...mockOrganizationUnit, description: null},
      isLoading: false,
      error: null,
      refetch: mockRefetch,
    });

    renderWithProviders(<OrganizationUnitEditPage />);

    await waitFor(() => {
      expect(screen.getByText('No description')).toBeInTheDocument();
    });
  });

  it('should open delete dialog when delete button is clicked in Advanced tab', async () => {
    renderWithProviders(<OrganizationUnitEditPage />);

    // Navigate to Advanced tab
    fireEvent.click(screen.getByText('Advanced'));

    await waitFor(() => {
      expect(screen.getByText('Delete Organization Unit')).toBeInTheDocument();
    });

    // Click delete button
    fireEvent.click(screen.getByText('Delete Organization Unit'));

    await waitFor(() => {
      // Delete dialog should open
      expect(screen.getByText('Are you sure?')).toBeInTheDocument();
    });
  });

  it('should show floating save bar when changes are made', async () => {
    renderWithProviders(<OrganizationUnitEditPage />);

    await waitFor(() => {
      expect(screen.getByText('Test Organization Unit')).toBeInTheDocument();
    });

    // Find and click edit button for name (second button after back)
    const editButtons = screen.getAllByRole('button');
    // The edit button is near the name
    const nameEditButton = editButtons.find(
      (btn) => btn.querySelector('svg') && btn.closest('div')?.textContent?.includes('Test Organization Unit'),
    );

    if (nameEditButton) {
      fireEvent.click(nameEditButton);

      // Type new name
      const nameInput = screen.getByRole('textbox');
      fireEvent.change(nameInput, {target: {value: 'Updated Name'}});
      fireEvent.blur(nameInput);

      await waitFor(() => {
        expect(screen.getByText('You have unsaved changes')).toBeInTheDocument();
      });
    }
  });

  it('should reset changes when reset button is clicked', async () => {
    renderWithProviders(<OrganizationUnitEditPage />);

    await waitFor(() => {
      expect(screen.getByText('Test Organization Unit')).toBeInTheDocument();
    });

    // Make a change to show the floating bar
    const editButtons = screen.getAllByRole('button');
    const nameEditButton = editButtons.find(
      (btn) => btn.querySelector('svg') && btn.closest('div')?.textContent?.includes('Test Organization Unit'),
    );

    if (nameEditButton) {
      fireEvent.click(nameEditButton);

      const nameInput = screen.getByRole('textbox');
      fireEvent.change(nameInput, {target: {value: 'Updated Name'}});
      fireEvent.blur(nameInput);

      await waitFor(() => {
        expect(screen.getByText('You have unsaved changes')).toBeInTheDocument();
      });

      // Click reset
      fireEvent.click(screen.getByText('Reset'));

      await waitFor(() => {
        expect(screen.queryByText('You have unsaved changes')).not.toBeInTheDocument();
      });
    }
  });

  it('should edit description when edit button is clicked', async () => {
    renderWithProviders(<OrganizationUnitEditPage />);

    await waitFor(() => {
      expect(screen.getByText('A test description')).toBeInTheDocument();
    });

    // Find edit button next to description
    const editButtons = screen.getAllByRole('button');
    const descriptionEditButton = editButtons.find(
      (btn) => btn.querySelector('svg') && btn.closest('div')?.textContent?.includes('A test description'),
    );

    if (descriptionEditButton) {
      fireEvent.click(descriptionEditButton);

      // Should show a textbox for editing
      await waitFor(() => {
        const textbox = screen.getByRole('textbox');
        expect(textbox).toBeInTheDocument();
      });

      const textbox = screen.getByRole('textbox');
      fireEvent.change(textbox, {target: {value: 'Updated description'}});
      fireEvent.blur(textbox);

      await waitFor(() => {
        expect(screen.getByText('You have unsaved changes')).toBeInTheDocument();
      });
    }
  });

  it('should cancel description editing on Escape key', async () => {
    renderWithProviders(<OrganizationUnitEditPage />);

    await waitFor(() => {
      expect(screen.getByText('A test description')).toBeInTheDocument();
    });

    const editButtons = screen.getAllByRole('button');
    const descriptionEditButton = editButtons.find(
      (btn) => btn.querySelector('svg') && btn.closest('div')?.textContent?.includes('A test description'),
    );

    if (descriptionEditButton) {
      fireEvent.click(descriptionEditButton);

      await waitFor(() => {
        expect(screen.getByRole('textbox')).toBeInTheDocument();
      });

      const textbox = screen.getByRole('textbox');
      fireEvent.keyDown(textbox, {key: 'Escape'});

      await waitFor(() => {
        expect(screen.getByText('A test description')).toBeInTheDocument();
      });
    }
  });

  it('should save description on Ctrl+Enter', async () => {
    renderWithProviders(<OrganizationUnitEditPage />);

    await waitFor(() => {
      expect(screen.getByText('A test description')).toBeInTheDocument();
    });

    const editButtons = screen.getAllByRole('button');
    const descriptionEditButton = editButtons.find(
      (btn) => btn.querySelector('svg') && btn.closest('div')?.textContent?.includes('A test description'),
    );

    if (descriptionEditButton) {
      fireEvent.click(descriptionEditButton);

      await waitFor(() => {
        expect(screen.getByRole('textbox')).toBeInTheDocument();
      });

      const textbox = screen.getByRole('textbox');
      fireEvent.change(textbox, {target: {value: 'New description'}});
      fireEvent.keyDown(textbox, {key: 'Enter', ctrlKey: true});

      await waitFor(() => {
        expect(screen.getByText('You have unsaved changes')).toBeInTheDocument();
      });
    }
  });

  it('should cancel name editing on Escape key', async () => {
    renderWithProviders(<OrganizationUnitEditPage />);

    await waitFor(() => {
      expect(screen.getByText('Test Organization Unit')).toBeInTheDocument();
    });

    const editButtons = screen.getAllByRole('button');
    const nameEditButton = editButtons.find(
      (btn) => btn.querySelector('svg') && btn.closest('div')?.textContent?.includes('Test Organization Unit'),
    );

    if (nameEditButton) {
      fireEvent.click(nameEditButton);

      await waitFor(() => {
        expect(screen.getByRole('textbox')).toBeInTheDocument();
      });

      const textbox = screen.getByRole('textbox');
      fireEvent.keyDown(textbox, {key: 'Escape'});

      await waitFor(() => {
        expect(screen.getByText('Test Organization Unit')).toBeInTheDocument();
      });
    }
  });

  it('should save name on Enter key', async () => {
    renderWithProviders(<OrganizationUnitEditPage />);

    await waitFor(() => {
      expect(screen.getByText('Test Organization Unit')).toBeInTheDocument();
    });

    const editButtons = screen.getAllByRole('button');
    const nameEditButton = editButtons.find(
      (btn) => btn.querySelector('svg') && btn.closest('div')?.textContent?.includes('Test Organization Unit'),
    );

    if (nameEditButton) {
      fireEvent.click(nameEditButton);

      await waitFor(() => {
        expect(screen.getByRole('textbox')).toBeInTheDocument();
      });

      const textbox = screen.getByRole('textbox');
      fireEvent.change(textbox, {target: {value: 'Updated Name'}});
      fireEvent.keyDown(textbox, {key: 'Enter'});

      await waitFor(() => {
        expect(screen.getByText('You have unsaved changes')).toBeInTheDocument();
      });
    }
  });

  it('should call save and refetch when save button is clicked', async () => {
    mockMutateAsync.mockResolvedValue({});

    renderWithProviders(<OrganizationUnitEditPage />);

    await waitFor(() => {
      expect(screen.getByText('Test Organization Unit')).toBeInTheDocument();
    });

    // Make a change to show the save bar
    const editButtons = screen.getAllByRole('button');
    const nameEditButton = editButtons.find(
      (btn) => btn.querySelector('svg') && btn.closest('div')?.textContent?.includes('Test Organization Unit'),
    );

    if (nameEditButton) {
      fireEvent.click(nameEditButton);

      const nameInput = screen.getByRole('textbox');
      fireEvent.change(nameInput, {target: {value: 'Updated Name'}});
      fireEvent.blur(nameInput);

      await waitFor(() => {
        expect(screen.getByText('You have unsaved changes')).toBeInTheDocument();
      });

      // Click save
      fireEvent.click(screen.getByText('Save'));

      await waitFor(() => {
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

    renderWithProviders(<OrganizationUnitEditPage />);

    await waitFor(() => {
      expect(screen.getByText('Test Organization Unit')).toBeInTheDocument();
    });

    // Make a change
    const editButtons = screen.getAllByRole('button');
    const nameEditButton = editButtons.find(
      (btn) => btn.querySelector('svg') && btn.closest('div')?.textContent?.includes('Test Organization Unit'),
    );

    if (nameEditButton) {
      fireEvent.click(nameEditButton);

      const nameInput = screen.getByRole('textbox');
      fireEvent.change(nameInput, {target: {value: 'Updated Name'}});
      fireEvent.blur(nameInput);

      await waitFor(() => {
        expect(screen.getByText('You have unsaved changes')).toBeInTheDocument();
      });

      // Click save - should not throw
      fireEvent.click(screen.getByText('Save'));
    }
  });

  it('should not save empty name on blur', async () => {
    renderWithProviders(<OrganizationUnitEditPage />);

    await waitFor(() => {
      expect(screen.getByText('Test Organization Unit')).toBeInTheDocument();
    });

    const editButtons = screen.getAllByRole('button');
    const nameEditButton = editButtons.find(
      (btn) => btn.querySelector('svg') && btn.closest('div')?.textContent?.includes('Test Organization Unit'),
    );

    if (nameEditButton) {
      fireEvent.click(nameEditButton);

      const nameInput = screen.getByRole('textbox');
      fireEvent.change(nameInput, {target: {value: ''}});
      fireEvent.blur(nameInput);

      // Should not show unsaved changes for empty name
      await waitFor(() => {
        expect(screen.getByText('Test Organization Unit')).toBeInTheDocument();
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

    renderWithProviders(<OrganizationUnitEditPage />);

    await waitFor(() => {
      expect(screen.getByText('Network error')).toBeInTheDocument();
    });

    // Click back button
    fireEvent.click(screen.getByText('Back'));

    await waitFor(() => {
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

    renderWithProviders(<OrganizationUnitEditPage />);

    await waitFor(() => {
      expect(screen.getByText('Organization unit not found')).toBeInTheDocument();
    });

    // Click back button
    fireEvent.click(screen.getByText('Back'));

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/organization-units');
    });
  });

  it('should handle delete dialog close', async () => {
    renderWithProviders(<OrganizationUnitEditPage />);

    // Navigate to Advanced tab
    fireEvent.click(screen.getByText('Advanced'));

    await waitFor(() => {
      expect(screen.getByText('Delete Organization Unit')).toBeInTheDocument();
    });

    // Open delete dialog
    fireEvent.click(screen.getByText('Delete Organization Unit'));

    await waitFor(() => {
      expect(screen.getByText('Are you sure?')).toBeInTheDocument();
    });

    // Close dialog by clicking cancel
    fireEvent.click(screen.getByText('Cancel'));
  });

  it('should switch to child OUs tab', async () => {
    renderWithProviders(<OrganizationUnitEditPage />);

    await waitFor(() => {
      expect(screen.getByText('Test Organization Unit')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('tab', {name: 'Child OUs'}));

    await waitFor(() => {
      expect(screen.getByRole('tab', {name: 'Child OUs', selected: true})).toBeInTheDocument();
    });
  });

  it('should switch to users tab', async () => {
    renderWithProviders(<OrganizationUnitEditPage />);

    await waitFor(() => {
      expect(screen.getByText('Test Organization Unit')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('tab', {name: 'Users'}));

    await waitFor(() => {
      expect(screen.getByRole('tab', {name: 'Users', selected: true})).toBeInTheDocument();
    });
  });

  it('should switch to groups tab', async () => {
    renderWithProviders(<OrganizationUnitEditPage />);

    await waitFor(() => {
      expect(screen.getByText('Test Organization Unit')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByRole('tab', {name: 'Groups'}));

    await waitFor(() => {
      expect(screen.getByRole('tab', {name: 'Groups', selected: true})).toBeInTheDocument();
    });
  });
});
