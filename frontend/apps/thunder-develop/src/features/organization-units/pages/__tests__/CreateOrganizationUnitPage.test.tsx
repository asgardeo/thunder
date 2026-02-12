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
import {screen, fireEvent, waitFor, renderWithProviders} from '@thunder/test-utils';
import CreateOrganizationUnitPage from '../CreateOrganizationUnitPage';
import type {OrganizationUnitListResponse} from '../../types/organization-units';

// Mock navigate
const mockNavigate = vi.fn();
vi.mock('react-router', async () => {
  const actual = await vi.importActual('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
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

// Mock create hook
const mockMutate = vi.fn();
vi.mock('../../api/useCreateOrganizationUnit', () => ({
  default: () => ({
    mutate: mockMutate,
    isPending: false,
  }),
}));

// Mock get OUs hook
const mockOUsData: OrganizationUnitListResponse = {
  totalResults: 2,
  startIndex: 1,
  count: 2,
  organizationUnits: [
    {id: 'ou-1', handle: 'parent-one', name: 'Parent One', description: null, parent: null},
    {id: 'ou-2', handle: 'parent-two', name: 'Parent Two', description: null, parent: null},
  ],
};

vi.mock('../../api/useGetOrganizationUnits', () => ({
  default: () => ({
    data: mockOUsData,
    isLoading: false,
    error: null,
  }),
}));

// Mock name suggestions utility
vi.mock('../../utils/generateOUNameSuggestions', () => ({
  default: () => ['Suggested Name One', 'Suggested Name Two', 'Suggested Name Three'],
}));

// Mock translations
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'organizationUnits:create.title': 'Create Organization Unit',
        'organizationUnits:create.heading': 'Create a new organization unit',
        'organizationUnits:create.suggestions.label': 'Try these suggestions:',
        'organizationUnits:create.error': 'Failed to create organization unit',
        'organizationUnits:form.name': 'Name',
        'organizationUnits:form.namePlaceholder': 'Enter organization unit name',
        'organizationUnits:form.handle': 'Handle',
        'organizationUnits:form.handlePlaceholder': 'Enter handle',
        'organizationUnits:form.handleHelperText': 'A unique identifier for this organization unit',
        'organizationUnits:form.description': 'Description',
        'organizationUnits:form.descriptionPlaceholder': 'Enter description',
        'organizationUnits:form.parent': 'Parent Organization Unit',
        'organizationUnits:form.parentPlaceholder': 'Select parent',
        'organizationUnits:form.parentHelperText': 'Optional parent organization unit',
        'common:actions.create': 'Create',
        'common:status.saving': 'Creating...',
      };
      return translations[key] ?? key;
    },
  }),
}));

describe('CreateOrganizationUnitPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockNavigate.mockReset();
    mockMutate.mockReset();
  });

  it('should render page title and heading', () => {
    renderWithProviders(<CreateOrganizationUnitPage />);

    expect(screen.getByText('Create Organization Unit')).toBeInTheDocument();
    expect(screen.getByText('Create a new organization unit')).toBeInTheDocument();
  });

  it('should render name input field', () => {
    renderWithProviders(<CreateOrganizationUnitPage />);

    expect(screen.getByLabelText(/Name/i)).toBeInTheDocument();
  });

  it('should render handle input field', () => {
    renderWithProviders(<CreateOrganizationUnitPage />);

    expect(screen.getByLabelText(/Handle/i)).toBeInTheDocument();
  });

  it('should render description input field', () => {
    renderWithProviders(<CreateOrganizationUnitPage />);

    expect(screen.getByLabelText(/Description/i)).toBeInTheDocument();
  });

  it('should render name suggestions', () => {
    renderWithProviders(<CreateOrganizationUnitPage />);

    expect(screen.getByText('Suggested Name One')).toBeInTheDocument();
    expect(screen.getByText('Suggested Name Two')).toBeInTheDocument();
    expect(screen.getByText('Suggested Name Three')).toBeInTheDocument();
  });

  it('should auto-generate handle from name', async () => {
    renderWithProviders(<CreateOrganizationUnitPage />);

    const nameInput = screen.getByLabelText(/Name/i);
    fireEvent.change(nameInput, {target: {value: 'Test Organization'}});

    const handleInput = screen.getByLabelText(/Handle/i);
    expect(handleInput).toHaveValue('test-organization');
  });

  it('should fill name when suggestion is clicked', async () => {
    renderWithProviders(<CreateOrganizationUnitPage />);

    fireEvent.click(screen.getByText('Suggested Name One'));

    const nameInput = screen.getByLabelText(/Name/i);
    expect(nameInput).toHaveValue('Suggested Name One');
  });

  it('should auto-generate handle when suggestion is clicked', async () => {
    renderWithProviders(<CreateOrganizationUnitPage />);

    fireEvent.click(screen.getByText('Suggested Name One'));

    const handleInput = screen.getByLabelText(/Handle/i);
    expect(handleInput).toHaveValue('suggested-name-one');
  });

  it('should not auto-generate handle after manual edit', async () => {
    renderWithProviders(<CreateOrganizationUnitPage />);

    const handleInput = screen.getByLabelText(/Handle/i);
    fireEvent.change(handleInput, {target: {value: 'my-custom-handle'}});

    const nameInput = screen.getByLabelText(/Name/i);
    fireEvent.change(nameInput, {target: {value: 'Test Organization'}});

    expect(handleInput).toHaveValue('my-custom-handle');
  });

  it('should disable create button when form is invalid', () => {
    renderWithProviders(<CreateOrganizationUnitPage />);

    const createButton = screen.getByText('Create');
    expect(createButton).toBeDisabled();
  });

  it('should enable create button when form is valid', async () => {
    renderWithProviders(<CreateOrganizationUnitPage />);

    const nameInput = screen.getByLabelText(/Name/i);
    const handleInput = screen.getByLabelText(/Handle/i);

    fireEvent.change(nameInput, {target: {value: 'Test Organization'}});
    fireEvent.change(handleInput, {target: {value: 'test-org'}});

    // Wait for form validation to complete
    await waitFor(() => {
      const createButton = screen.getByText('Create');
      expect(createButton).not.toBeDisabled();
    });
  });

  it('should call mutate on form submit', async () => {
    renderWithProviders(<CreateOrganizationUnitPage />);

    const nameInput = screen.getByLabelText(/Name/i);
    fireEvent.change(nameInput, {target: {value: 'Test Organization'}});

    // Wait for form validation to complete
    await waitFor(() => {
      const createButton = screen.getByText('Create');
      expect(createButton).not.toBeDisabled();
    });

    const createButton = screen.getByText('Create');
    fireEvent.click(createButton);

    await waitFor(() => {
      expect(mockMutate).toHaveBeenCalledWith(
        expect.objectContaining({
          name: 'Test Organization',
          handle: 'test-organization',
        }),
        expect.any(Object),
      );
    });
  });

  it('should navigate back when close button is clicked', async () => {
    renderWithProviders(<CreateOrganizationUnitPage />);

    // Find the close button (X icon button)
    const closeButton = screen.getByRole('button', {name: ''});
    fireEvent.click(closeButton);

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/organization-units');
    });
  });

  it('should navigate on successful creation', async () => {
    mockMutate.mockImplementation((_data, options: {onSuccess: () => void}) => {
      options.onSuccess();
    });

    renderWithProviders(<CreateOrganizationUnitPage />);

    const nameInput = screen.getByLabelText(/Name/i);
    fireEvent.change(nameInput, {target: {value: 'Test Organization'}});

    // Wait for form validation to complete
    await waitFor(() => {
      const createButton = screen.getByText('Create');
      expect(createButton).not.toBeDisabled();
    });

    const createButton = screen.getByText('Create');
    fireEvent.click(createButton);

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/organization-units');
    });
  });

  it('should display error on creation failure', async () => {
    mockMutate.mockImplementation((_data, options: {onError: (err: Error) => void}) => {
      options.onError(new Error('Network error'));
    });

    renderWithProviders(<CreateOrganizationUnitPage />);

    const nameInput = screen.getByLabelText(/Name/i);
    fireEvent.change(nameInput, {target: {value: 'Test Organization'}});

    // Wait for form validation to complete
    await waitFor(() => {
      const createButton = screen.getByText('Create');
      expect(createButton).not.toBeDisabled();
    });

    const createButton = screen.getByText('Create');
    fireEvent.click(createButton);

    await waitFor(() => {
      expect(screen.getByText('Network error')).toBeInTheDocument();
    });
  });

  it('should close error alert when close button is clicked', async () => {
    mockMutate.mockImplementation((_data, options: {onError: (err: Error) => void}) => {
      options.onError(new Error('Network error'));
    });

    renderWithProviders(<CreateOrganizationUnitPage />);

    const nameInput = screen.getByLabelText(/Name/i);
    fireEvent.change(nameInput, {target: {value: 'Test Organization'}});

    // Wait for form validation to complete
    await waitFor(() => {
      const createButton = screen.getByText('Create');
      expect(createButton).not.toBeDisabled();
    });

    const createButton = screen.getByText('Create');
    fireEvent.click(createButton);

    await waitFor(() => {
      expect(screen.getByText('Network error')).toBeInTheDocument();
    });

    // Close the alert
    const alertCloseButton = screen.getByRole('button', {name: /close/i});
    fireEvent.click(alertCloseButton);

    await waitFor(() => {
      expect(screen.queryByText('Network error')).not.toBeInTheDocument();
    });
  });

  it('should include description in request when provided', async () => {
    renderWithProviders(<CreateOrganizationUnitPage />);

    const nameInput = screen.getByLabelText(/Name/i);
    const descriptionInput = screen.getByLabelText(/Description/i);

    fireEvent.change(nameInput, {target: {value: 'Test Organization'}});
    fireEvent.change(descriptionInput, {target: {value: 'A test description'}});

    // Wait for form validation to complete
    await waitFor(() => {
      const createButton = screen.getByText('Create');
      expect(createButton).not.toBeDisabled();
    });

    const createButton = screen.getByText('Create');
    fireEvent.click(createButton);

    await waitFor(() => {
      expect(mockMutate).toHaveBeenCalledWith(
        expect.objectContaining({
          description: 'A test description',
        }),
        expect.any(Object),
      );
    });
  });

  it('should set description to null when empty', async () => {
    renderWithProviders(<CreateOrganizationUnitPage />);

    const nameInput = screen.getByLabelText(/Name/i);
    fireEvent.change(nameInput, {target: {value: 'Test Organization'}});

    // Wait for form validation to complete
    await waitFor(() => {
      const createButton = screen.getByText('Create');
      expect(createButton).not.toBeDisabled();
    });

    const createButton = screen.getByText('Create');
    fireEvent.click(createButton);

    await waitFor(() => {
      expect(mockMutate).toHaveBeenCalledWith(
        expect.objectContaining({
          description: null,
        }),
        expect.any(Object),
      );
    });
  });

  it('should set parent to null when not selected', async () => {
    renderWithProviders(<CreateOrganizationUnitPage />);

    const nameInput = screen.getByLabelText(/Name/i);
    fireEvent.change(nameInput, {target: {value: 'Test Organization'}});

    // Wait for form validation to complete
    await waitFor(() => {
      const createButton = screen.getByText('Create');
      expect(createButton).not.toBeDisabled();
    });

    const createButton = screen.getByText('Create');
    fireEvent.click(createButton);

    await waitFor(() => {
      expect(mockMutate).toHaveBeenCalledWith(
        expect.objectContaining({
          parent: null,
        }),
        expect.any(Object),
      );
    });
  });

  it('should select parent OU from autocomplete', async () => {
    renderWithProviders(<CreateOrganizationUnitPage />);

    // Find the parent autocomplete input
    const parentInput = screen.getByLabelText(/Parent Organization Unit/i);

    // Open the autocomplete dropdown
    fireEvent.mouseDown(parentInput);
    fireEvent.click(parentInput);

    await waitFor(() => {
      expect(screen.getByText('Parent One')).toBeInTheDocument();
    });

    // Select a parent
    fireEvent.click(screen.getByText('Parent One'));

    // Fill required fields and submit
    const nameInput = screen.getByLabelText(/Name/i);
    fireEvent.change(nameInput, {target: {value: 'Test Organization'}});

    // Wait for form validation to complete
    await waitFor(() => {
      const createButton = screen.getByText('Create');
      expect(createButton).not.toBeDisabled();
    });

    const createButton = screen.getByText('Create');
    fireEvent.click(createButton);

    await waitFor(() => {
      expect(mockMutate).toHaveBeenCalledWith(
        expect.objectContaining({
          parent: 'ou-1',
        }),
        expect.any(Object),
      );
    });
  });

  it('should keep handle unchanged after manual edit when suggestion is clicked', async () => {
    renderWithProviders(<CreateOrganizationUnitPage />);

    const handleInput = screen.getByLabelText(/Handle/i);
    fireEvent.change(handleInput, {target: {value: 'my-custom-handle'}});

    fireEvent.click(screen.getByText('Suggested Name Two'));

    // Handle should not change after suggestion click since it was manually edited
    expect(handleInput).toHaveValue('my-custom-handle');
  });

  it('should handle error without message', async () => {
    mockMutate.mockImplementation((_data, options: {onError: (err: unknown) => void}) => {
      options.onError({});
    });

    renderWithProviders(<CreateOrganizationUnitPage />);

    const nameInput = screen.getByLabelText(/Name/i);
    fireEvent.change(nameInput, {target: {value: 'Test Organization'}});

    // Wait for form validation to complete
    await waitFor(() => {
      const createButton = screen.getByText('Create');
      expect(createButton).not.toBeDisabled();
    });

    const createButton = screen.getByText('Create');
    fireEvent.click(createButton);

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument();
    });
  });

  it('should handle close navigation error gracefully', async () => {
    mockNavigate.mockRejectedValue(new Error('Navigation failed'));

    renderWithProviders(<CreateOrganizationUnitPage />);

    const closeButton = screen.getByRole('button', {name: ''});
    fireEvent.click(closeButton);

    // Should not throw - error is logged
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/organization-units');
    });
  });

  it('should handle success navigation error gracefully', async () => {
    mockNavigate.mockRejectedValue(new Error('Navigation failed'));
    mockMutate.mockImplementation((_data, options: {onSuccess: () => void}) => {
      options.onSuccess();
    });

    renderWithProviders(<CreateOrganizationUnitPage />);

    const nameInput = screen.getByLabelText(/Name/i);
    fireEvent.change(nameInput, {target: {value: 'Test Organization'}});

    // Wait for form validation to complete
    await waitFor(() => {
      const createButton = screen.getByText('Create');
      expect(createButton).not.toBeDisabled();
    });

    const createButton = screen.getByText('Create');
    fireEvent.click(createButton);

    // Should not throw - error is logged
    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/organization-units');
    });
  });

  it('should render parent autocomplete with options', async () => {
    renderWithProviders(<CreateOrganizationUnitPage />);

    const parentInput = screen.getByLabelText(/Parent Organization Unit/i);
    fireEvent.focus(parentInput);
    fireEvent.mouseDown(parentInput);

    await waitFor(() => {
      expect(screen.getByText('Parent One')).toBeInTheDocument();
      expect(screen.getByText('Parent Two')).toBeInTheDocument();
    });
  });

  it('should trim whitespace from inputs on submit', async () => {
    renderWithProviders(<CreateOrganizationUnitPage />);

    const nameInput = screen.getByLabelText(/Name/i);
    const handleInput = screen.getByLabelText(/Handle/i);
    const descriptionInput = screen.getByLabelText(/Description/i);

    fireEvent.change(nameInput, {target: {value: '  Test Organization  '}});
    fireEvent.change(handleInput, {target: {value: '  test-org  '}});
    fireEvent.change(descriptionInput, {target: {value: '  A description  '}});

    // Wait for form validation to complete
    await waitFor(() => {
      const createButton = screen.getByText('Create');
      expect(createButton).not.toBeDisabled();
    });

    const createButton = screen.getByText('Create');
    fireEvent.click(createButton);

    await waitFor(() => {
      expect(mockMutate).toHaveBeenCalledWith(
        expect.objectContaining({
          name: 'Test Organization',
          handle: 'test-org',
          description: 'A description',
        }),
        expect.any(Object),
      );
    });
  });

  it('should render progress bar', () => {
    renderWithProviders(<CreateOrganizationUnitPage />);

    expect(screen.getByRole('progressbar')).toBeInTheDocument();
  });

  it('should render suggestions label', () => {
    renderWithProviders(<CreateOrganizationUnitPage />);

    expect(screen.getByText('Try these suggestions:')).toBeInTheDocument();
  });

  it('should compare autocomplete options by id', async () => {
    renderWithProviders(<CreateOrganizationUnitPage />);

    const parentInput = screen.getByLabelText(/Parent Organization Unit/i);

    // Open the dropdown
    fireEvent.mouseDown(parentInput);

    await waitFor(() => {
      expect(screen.getByText('Parent One')).toBeInTheDocument();
    });

    // Select Parent One
    fireEvent.click(screen.getByText('Parent One'));

    // Open the dropdown again
    fireEvent.mouseDown(parentInput);

    await waitFor(() => {
      // Parent One should still be selectable/visible as an option
      expect(screen.getByText('Parent One')).toBeInTheDocument();
    });

    // The isOptionEqualToValue (line 343) is used to compare options
    // This verifies the selected option is properly maintained
  });
});
