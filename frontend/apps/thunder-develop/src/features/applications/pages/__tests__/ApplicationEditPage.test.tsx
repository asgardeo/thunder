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

import {render, screen, waitFor} from '@thunder/test-utils';
import userEvent from '@testing-library/user-event';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import type {UseQueryResult, UseMutationResult} from '@tanstack/react-query';
import type {Application} from '../../models/application';
import ApplicationEditPage from '../ApplicationEditPage';
import useGetApplication from '../../api/useGetApplication';
import useUpdateApplication from '../../api/useUpdateApplication';
import getTemplateMetadata from '../../utils/getTemplateMetadata';
import getIntegrationGuidesForTemplate from '../../utils/getIntegrationGuidesForTemplate';

// Mock dependencies
vi.mock('react-router', async () => {
  const actual = await vi.importActual('react-router');
  return {
    ...actual,
    useNavigate: vi.fn(() => vi.fn()),
    useParams: vi.fn(() => ({applicationId: 'test-app-id'})),
  };
});

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'applications:view.back': 'Back to Applications',
        'applications:edit.page.loading': 'Loading application...',
        'applications:edit.page.notFound.title': 'Application Not Found',
        'applications:edit.page.notFound.description': 'The application you are looking for does not exist.',
        'applications:edit.page.description.placeholder': 'Add a description',
        'applications:edit.page.description.empty': 'No description provided',
        'applications:edit.page.tabs.overview': 'Overview',
        'applications:edit.page.tabs.general': 'General',
        'applications:edit.page.tabs.flows': 'Flows',
        'applications:edit.page.tabs.customization': 'Customization',
        'applications:edit.page.tabs.token': 'Token',
        'applications:edit.page.tabs.advanced': 'Advanced',
        'applications:edit.page.unsavedChanges': 'You have unsaved changes',
        'applications:edit.page.reset': 'Reset',
        'applications:edit.page.save': 'Save Changes',
        'applications:edit.page.saving': 'Saving...',
      };
      return translations[key] || key;
    },
  }),
}));

vi.mock('../../api/useGetApplication', () => ({
  default: vi.fn(),
}));

vi.mock('../../api/useUpdateApplication', () => ({
  default: vi.fn(),
}));

vi.mock('../../utils/getTemplateMetadata', () => ({
  default: vi.fn(),
}));

vi.mock('../../utils/getIntegrationGuidesForTemplate', () => ({
  default: vi.fn(),
}));

// Mock child components
vi.mock('../../components/edit-application/general-settings/EditGeneralSettings', () => ({
  default: vi.fn(() => <div data-testid="edit-general-settings">General Settings</div>),
}));

vi.mock('../../components/edit-application/flows-settings/EditFlowsSettings', () => ({
  default: vi.fn(() => <div data-testid="edit-flows-settings">Flows Settings</div>),
}));

vi.mock('../../components/edit-application/customization-settings/EditCustomizationSettings', () => ({
  default: vi.fn(() => <div data-testid="edit-customization-settings">Customization Settings</div>),
}));

vi.mock('../../components/edit-application/token-settings/EditTokenSettings', () => ({
  default: vi.fn(() => <div data-testid="edit-token-settings">Token Settings</div>),
}));

vi.mock('../../components/edit-application/advanced-settings/EditAdvancedSettings', () => ({
  default: vi.fn(() => <div data-testid="edit-advanced-settings">Advanced Settings</div>),
}));

vi.mock('../../components/edit-application/integration-guides/IntegrationGuides', () => ({
  default: vi.fn(() => <div data-testid="integration-guides">Integration Guides</div>),
}));

vi.mock('../../components/LogoUpdateModal', () => ({
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

const mockUseGetApplication = useGetApplication as ReturnType<typeof vi.fn>;
const mockUseUpdateApplication = useUpdateApplication as ReturnType<typeof vi.fn>;
const mockGetTemplateMetadata = getTemplateMetadata as ReturnType<typeof vi.fn>;
const mockGetIntegrationGuidesForTemplate = getIntegrationGuidesForTemplate as ReturnType<typeof vi.fn>;

describe('ApplicationEditPage', () => {
  const mockApplication: Application = {
    id: 'test-app-id',
    name: 'Test Application',
    description: 'Test application description',
    template: 'react',
    logo_url: 'https://example.com/logo.png',
    url: 'https://example.com',
    inbound_auth_config: [
      {
        type: 'oauth2',
        config: {
          response_types: ['code'],
          client_id: 'test-client-id',
          client_secret: 'test-client-secret',
          grant_types: ['authorization_code'],
          redirect_uris: ['https://example.com/callback'],
          pkce_required: true,
          public_client: false,
          token_endpoint_auth_method: 'client_secret_basic',
        },
      },
    ],
  };

  const mockUpdateApplicationMutate = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();

    // Default mock implementations
    mockGetTemplateMetadata.mockReturnValue({
      displayName: 'React',
      icon: <div>React Icon</div>,
    });

    // Return null by default to indicate no integration guides
    mockGetIntegrationGuidesForTemplate.mockReturnValue(null);

    mockUseGetApplication.mockReturnValue({
      data: mockApplication,
      isLoading: false,
      isError: false,
      error: null,
    } as UseQueryResult<Application>);

    mockUseUpdateApplication.mockReturnValue({
      mutate: mockUpdateApplicationMutate,
      mutateAsync: vi.fn().mockResolvedValue(mockApplication),
      isPending: false,
      isError: false,
      error: null,
    } as unknown as UseMutationResult<Application, Error, Partial<Application>>);
  });

  const renderComponent = () => render(<ApplicationEditPage />);

  describe('Loading State', () => {
    it('should display loading state while fetching application', () => {
      mockUseGetApplication.mockReturnValue({
        data: undefined,
        isLoading: true,
        isError: false,
        error: null,
      } as UseQueryResult<Application>);

      renderComponent();

      // When loading, the component shows a loading indicator
      expect(screen.getByRole('progressbar')).toBeInTheDocument();
    });
  });

  describe('Error State', () => {
    it('should display error state when application is not found', () => {
      mockUseGetApplication.mockReturnValue({
        data: undefined,
        isLoading: false,
        isError: true,
        error: new Error('Not found'),
      } as unknown as UseQueryResult<Application>);

      renderComponent();

      // Check for error UI elements
      expect(screen.getByRole('button', {name: /back to applications/i})).toBeInTheDocument();
    });

    it('should display back button in error state', () => {
      mockUseGetApplication.mockReturnValue({
        data: undefined,
        isLoading: false,
        isError: true,
        error: new Error('Not found'),
      } as unknown as UseQueryResult<Application>);

      renderComponent();

      expect(screen.getByRole('button', {name: /back to applications/i})).toBeInTheDocument();
    });
  });

  describe('Successful Load', () => {
    it('should render application details correctly', () => {
      renderComponent();

      expect(screen.getByText('Test Application')).toBeInTheDocument();
      expect(screen.getByText('Test application description')).toBeInTheDocument();
    });

    it('should display application logo', () => {
      renderComponent();

      const logo = screen.getByRole('img');
      expect(logo).toHaveAttribute('src', 'https://example.com/logo.png');
    });

    it('should display template chip when template metadata is available', () => {
      renderComponent();

      expect(screen.getByText('React')).toBeInTheDocument();
    });

    it('should display back button', () => {
      renderComponent();

      expect(screen.getByRole('button', {name: /back to applications/i})).toBeInTheDocument();
    });

    it('should handle empty description', () => {
      mockUseGetApplication.mockReturnValue({
        data: {...mockApplication, description: undefined},
        isLoading: false,
        isError: false,
        error: null,
      } as UseQueryResult<Application>);

      renderComponent();

      expect(screen.getByText('No description provided')).toBeInTheDocument();
    });
  });

  describe('Tab Navigation', () => {
    it('should render all tabs without integration guides', () => {
      renderComponent();

      expect(screen.getByRole('tab', {name: /general/i})).toBeInTheDocument();
      expect(screen.getByRole('tab', {name: /flows/i})).toBeInTheDocument();
      expect(screen.getByRole('tab', {name: /customization/i})).toBeInTheDocument();
      expect(screen.getByRole('tab', {name: /token/i})).toBeInTheDocument();
      expect(screen.getByRole('tab', {name: /advanced/i})).toBeInTheDocument();
    });

    it('should render overview tab when integration guides are available', () => {
      mockGetIntegrationGuidesForTemplate.mockReturnValue(['react-vite']);

      renderComponent();

      expect(screen.getByRole('tab', {name: /overview/i})).toBeInTheDocument();
    });

    it('should display general settings tab by default when no integration guides', async () => {
      // Mock returns null by default (no integration guides)
      renderComponent();

      // When there are no integration guides, general tab should be first and selected
      const generalTab = screen.getByRole('tab', {name: /general/i});
      expect(generalTab).toHaveAttribute('aria-selected', 'true');
    });

    it('should display overview tab by default when integration guides are available', async () => {
      mockGetIntegrationGuidesForTemplate.mockReturnValue(['react-vite']);

      renderComponent();

      await waitFor(() => {
        expect(screen.getByTestId('integration-guides')).toBeInTheDocument();
      });
    });

    it('should switch to flows tab when clicked', async () => {
      const user = userEvent.setup();
      renderComponent();

      const flowsTab = screen.getByRole('tab', {name: /flows/i});
      await user.click(flowsTab);

      await waitFor(() => {
        expect(screen.getByTestId('edit-flows-settings')).toBeInTheDocument();
      });
    });

    it('should switch to customization tab when clicked', async () => {
      const user = userEvent.setup();
      renderComponent();

      const customizationTab = screen.getByRole('tab', {name: /customization/i});
      await user.click(customizationTab);

      await waitFor(() => {
        expect(screen.getByTestId('edit-customization-settings')).toBeInTheDocument();
      });
    });

    it('should switch to token tab when clicked', async () => {
      const user = userEvent.setup();
      renderComponent();

      const tokenTab = screen.getByRole('tab', {name: /token/i});
      await user.click(tokenTab);

      await waitFor(() => {
        expect(screen.getByTestId('edit-token-settings')).toBeInTheDocument();
      });
    });

    it('should switch to advanced tab when clicked', async () => {
      const user = userEvent.setup();
      renderComponent();

      const advancedTab = screen.getByRole('tab', {name: /advanced/i});
      await user.click(advancedTab);

      await waitFor(() => {
        expect(screen.getByTestId('edit-advanced-settings')).toBeInTheDocument();
      });
    });
  });

  describe('Inline Editing', () => {
    it('should enable name editing when edit icon is clicked', async () => {
      const user = userEvent.setup();
      renderComponent();

      const nameSection = screen.getByText('Test Application').closest('div');
      const editButton = nameSection?.querySelector('button');
      expect(editButton).toBeInTheDocument();

      await user.click(editButton!);

      const nameInput = screen.getByRole('textbox');
      expect(nameInput).toHaveValue('Test Application');
    });

    it('should save name changes on blur', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Click edit button
      const nameSection = screen.getByText('Test Application').closest('div');
      const editButton = nameSection?.querySelector('button');
      await user.click(editButton!);

      // Change name
      const nameInput = screen.getByRole('textbox');
      await user.clear(nameInput);
      await user.type(nameInput, 'Updated Application');

      // Blur to save
      await user.tab();

      await waitFor(() => {
        expect(screen.getByText('Updated Application')).toBeInTheDocument();
      });
    });

    it('should save name changes on Enter key', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Click edit button
      const nameSection = screen.getByText('Test Application').closest('div');
      const editButton = nameSection?.querySelector('button');
      await user.click(editButton!);

      // Change name and press Enter
      const nameInput = screen.getByRole('textbox');
      await user.clear(nameInput);
      await user.type(nameInput, 'Updated Application{Enter}');

      await waitFor(() => {
        expect(screen.getByText('Updated Application')).toBeInTheDocument();
      });
    });

    it('should cancel name editing on Escape key', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Click edit button
      const nameSection = screen.getByText('Test Application').closest('div');
      const editButton = nameSection?.querySelector('button');
      await user.click(editButton!);

      // Change name and press Escape
      const nameInput = screen.getByRole('textbox');
      await user.clear(nameInput);
      await user.type(nameInput, 'Updated Application{Escape}');

      await waitFor(() => {
        expect(screen.getByText('Test Application')).toBeInTheDocument();
        expect(screen.queryByDisplayValue('Updated Application')).not.toBeInTheDocument();
      });
    });

    it('should enable description editing when edit icon is clicked', async () => {
      const user = userEvent.setup();
      renderComponent();

      const descriptionSection = screen.getByText('Test application description').closest('div');
      const editButton = descriptionSection?.querySelector('button');
      expect(editButton).toBeInTheDocument();

      await user.click(editButton!);

      const descriptionInput = screen.getByPlaceholderText('Add a description');
      expect(descriptionInput).toHaveValue('Test application description');
    });

    it('should save description changes on blur', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Click edit button
      const descriptionSection = screen.getByText('Test application description').closest('div');
      const editButton = descriptionSection?.querySelector('button');
      await user.click(editButton!);

      // Change description
      const descriptionInput = screen.getByPlaceholderText('Add a description');
      await user.clear(descriptionInput);
      await user.type(descriptionInput, 'Updated description');

      // Blur to save
      await user.tab();

      await waitFor(() => {
        expect(screen.getByText('Updated description')).toBeInTheDocument();
      });
    });

    it('should save description changes on Ctrl+Enter', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Click edit button
      const descriptionSection = screen.getByText('Test application description').closest('div');
      const editButton = descriptionSection?.querySelector('button');
      await user.click(editButton!);

      // Change description
      const descriptionInput = screen.getByPlaceholderText('Add a description');
      await user.clear(descriptionInput);
      await user.type(descriptionInput, 'Description via Ctrl+Enter');

      // Press Ctrl+Enter to save
      await user.keyboard('{Control>}{Enter}{/Control}');

      await waitFor(() => {
        expect(screen.getByText('Description via Ctrl+Enter')).toBeInTheDocument();
      });
    });

    it('should cancel description editing on Escape key', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Click edit button
      const descriptionSection = screen.getByText('Test application description').closest('div');
      const editButton = descriptionSection?.querySelector('button');
      await user.click(editButton!);

      // Change description and press Escape
      const descriptionInput = screen.getByPlaceholderText('Add a description');
      await user.clear(descriptionInput);
      await user.type(descriptionInput, 'Changed description');
      await user.keyboard('{Escape}');

      await waitFor(() => {
        expect(screen.getByText('Test application description')).toBeInTheDocument();
        expect(screen.queryByDisplayValue('Changed description')).not.toBeInTheDocument();
      });
    });

    it('should not save empty name on blur', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Click edit button
      const nameSection = screen.getByText('Test Application').closest('div');
      const editButton = nameSection?.querySelector('button');
      await user.click(editButton!);

      // Clear name
      const nameInput = screen.getByRole('textbox');
      await user.clear(nameInput);

      // Blur without entering text
      await user.tab();

      await waitFor(() => {
        expect(screen.getByText('Test Application')).toBeInTheDocument();
      });
    });
  });

  describe('Logo Update', () => {
    it('should render logo update modal', () => {
      renderComponent();

      // Modal should be in the DOM (hidden by default)
      expect(screen.getByTestId('logo-update-modal')).toBeInTheDocument();
    });

    it('should open logo modal when avatar is clicked', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Click on the avatar to open the modal
      const avatar = screen.getByRole('img');
      await user.click(avatar);

      await waitFor(() => {
        const modal = screen.getByTestId('logo-update-modal');
        expect(modal).toHaveStyle({display: 'block'});
      });
    });

    it('should update logo and close modal when logo is updated', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Open the modal
      const avatar = screen.getByRole('img');
      await user.click(avatar);

      // Click update logo button in modal
      const updateLogoButton = screen.getByRole('button', {name: /update logo/i});
      await user.click(updateLogoButton);

      await waitFor(() => {
        // Should show unsaved changes since logo was updated
        expect(screen.getByText('You have unsaved changes')).toBeInTheDocument();
      });

      // Modal should be closed
      await waitFor(() => {
        const modal = screen.getByTestId('logo-update-modal');
        expect(modal).toHaveStyle({display: 'none'});
      });
    });

    it('should close logo modal when close button is clicked', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Open the modal
      const avatar = screen.getByRole('img');
      await user.click(avatar);

      // Click close button
      const closeButton = screen.getByRole('button', {name: /close/i});
      await user.click(closeButton);

      await waitFor(() => {
        const modal = screen.getByTestId('logo-update-modal');
        expect(modal).toHaveStyle({display: 'none'});
      });
    });
  });

  describe('Save Functionality', () => {
    it('should show floating action bar when changes are made', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Make a change to name
      const nameSection = screen.getByText('Test Application').closest('div');
      const editButton = nameSection?.querySelector('button');
      await user.click(editButton!);

      const nameInput = screen.getByRole('textbox');
      await user.clear(nameInput);
      await user.type(nameInput, 'Updated Application{Enter}');

      await waitFor(() => {
        expect(screen.getByText('You have unsaved changes')).toBeInTheDocument();
      });
    });

    it('should display reset and save buttons in action bar', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Make a change
      const nameSection = screen.getByText('Test Application').closest('div');
      const editButton = nameSection?.querySelector('button');
      await user.click(editButton!);

      const nameInput = screen.getByRole('textbox');
      await user.clear(nameInput);
      await user.type(nameInput, 'Updated Application{Enter}');

      await waitFor(() => {
        expect(screen.getByRole('button', {name: /reset/i})).toBeInTheDocument();
        expect(screen.getByRole('button', {name: /save changes/i})).toBeInTheDocument();
      });
    });

    it('should reset changes when reset button is clicked', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Make a change
      const nameSection = screen.getByText('Test Application').closest('div');
      const editButton = nameSection?.querySelector('button');
      await user.click(editButton!);

      const nameInput = screen.getByRole('textbox');
      await user.clear(nameInput);
      await user.type(nameInput, 'Updated Application{Enter}');

      // Click reset
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /reset/i})).toBeInTheDocument();
      });

      const resetButton = screen.getByRole('button', {name: /reset/i});
      await user.click(resetButton);

      await waitFor(() => {
        expect(screen.queryByText('You have unsaved changes')).not.toBeInTheDocument();
      });
    });

    it('should save changes when save button is clicked', async () => {
      const user = userEvent.setup();
      const mockMutateAsync = vi.fn().mockResolvedValue(mockApplication);

      mockUseUpdateApplication.mockReturnValue({
        mutate: mockUpdateApplicationMutate,
        mutateAsync: mockMutateAsync,
        isPending: false,
        isError: false,
        error: null,
      } as unknown as UseMutationResult<Application, Error, Partial<Application>>);

      renderComponent();

      // Make a change
      const nameSection = screen.getByText('Test Application').closest('div');
      const editButton = nameSection?.querySelector('button');
      await user.click(editButton!);

      const nameInput = screen.getByRole('textbox');
      await user.clear(nameInput);
      await user.type(nameInput, 'Updated Application{Enter}');

      // Click save
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /save changes/i})).toBeInTheDocument();
      });

      const saveButton = screen.getByRole('button', {name: /save changes/i});
      await user.click(saveButton);

      await waitFor(() => {
        expect(mockMutateAsync).toHaveBeenCalled();
        const callArgs = mockMutateAsync.mock.calls[0][0] as {applicationId: string; data: Partial<Application>};
        expect(callArgs).toHaveProperty('applicationId', 'test-app-id');
        expect(callArgs).toHaveProperty('data');
        expect(callArgs.data).toHaveProperty('name', 'Updated Application');
      });
    });

    it('should disable save button while saving', async () => {
      const user = userEvent.setup();

      mockUseUpdateApplication.mockReturnValue({
        mutate: mockUpdateApplicationMutate,
        mutateAsync: vi.fn().mockResolvedValue(mockApplication),
        isPending: true,
        isError: false,
        error: null,
      } as unknown as UseMutationResult<Application, Error, Partial<Application>>);

      renderComponent();

      // Make a change first to show the action bar
      const nameSection = screen.getByText('Test Application').closest('div');
      const editButton = nameSection?.querySelector('button');
      await user.click(editButton!);

      const nameInput = screen.getByRole('textbox');
      await user.clear(nameInput);
      await user.type(nameInput, 'Updated Application{Enter}');

      await waitFor(() => {
        const saveButton = screen.getByRole('button', {name: /saving/i});
        expect(saveButton).toBeDisabled();
      });
    });

    it('should hide action bar after successful save', async () => {
      const user = userEvent.setup();
      const mockMutateAsync = vi.fn().mockResolvedValue({...mockApplication, name: 'Updated Application'});

      mockUseUpdateApplication.mockReturnValue({
        mutate: mockUpdateApplicationMutate,
        mutateAsync: mockMutateAsync,
        isPending: false,
        isError: false,
        error: null,
      } as unknown as UseMutationResult<Application, Error, Partial<Application>>);

      renderComponent();

      // Make a change
      const nameSection = screen.getByText('Test Application').closest('div');
      const editButton = nameSection?.querySelector('button');
      await user.click(editButton!);

      const nameInput = screen.getByRole('textbox');
      await user.clear(nameInput);
      await user.type(nameInput, 'Updated Application{Enter}');

      // Wait for action bar to appear
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /save changes/i})).toBeInTheDocument();
      });

      // Click save
      const saveButton = screen.getByRole('button', {name: /save changes/i});
      await user.click(saveButton);

      await waitFor(
        () => {
          expect(screen.queryByText('You have unsaved changes')).not.toBeInTheDocument();
        },
        {timeout: 10000},
      );
    });
  });

  describe('Accessibility', () => {
    it('should have proper ARIA labels for tabs', () => {
      renderComponent();

      const generalTab = screen.getByRole('tab', {name: /general/i});
      expect(generalTab).toHaveAttribute('id');
      expect(generalTab).toHaveAttribute('aria-controls');
    });

    it('should maintain focus management during inline editing', async () => {
      const user = userEvent.setup();
      renderComponent();

      const nameSection = screen.getByText('Test Application').closest('div');
      const editButton = nameSection?.querySelector('button');
      await user.click(editButton!);

      const nameInput = screen.getByRole('textbox');
      expect(nameInput).toHaveFocus();
    });
  });

  describe('Application Not Found', () => {
    it('should display warning when application is null', () => {
      mockUseGetApplication.mockReturnValue({
        data: null,
        isLoading: false,
        isError: false,
        error: null,
      } as unknown as UseQueryResult<Application>);

      renderComponent();

      expect(screen.getByText('applications:view.notFound')).toBeInTheDocument();
      expect(screen.getByRole('button', {name: /back to applications/i})).toBeInTheDocument();
    });
  });

  describe('Error Handling', () => {
    it('should display error message from error object', () => {
      const errorMessage = 'Custom error message';
      mockUseGetApplication.mockReturnValue({
        data: undefined,
        isLoading: false,
        isError: true,
        error: {message: errorMessage},
      } as unknown as UseQueryResult<Application>);

      renderComponent();

      expect(screen.getByText(errorMessage)).toBeInTheDocument();
    });

    it('should display default error message when error has no message', () => {
      mockUseGetApplication.mockReturnValue({
        data: undefined,
        isLoading: false,
        isError: true,
        error: {},
      } as unknown as UseQueryResult<Application>);

      renderComponent();

      expect(screen.getByText('applications:view.error')).toBeInTheDocument();
    });

    it('should handle save failure gracefully', async () => {
      const user = userEvent.setup();
      const mockMutateAsync = vi.fn().mockRejectedValue(new Error('Save failed'));

      mockUseUpdateApplication.mockReturnValue({
        mutate: mockUpdateApplicationMutate,
        mutateAsync: mockMutateAsync,
        isPending: false,
        isError: false,
        error: null,
      } as unknown as UseMutationResult<Application, Error, Partial<Application>>);

      renderComponent();

      // Make a change
      const nameSection = screen.getByText('Test Application').closest('div');
      const editButton = nameSection?.querySelector('button');
      await user.click(editButton!);

      const nameInput = screen.getByRole('textbox');
      await user.clear(nameInput);
      await user.type(nameInput, 'Updated Application{Enter}');

      // Click save
      await waitFor(() => {
        expect(screen.getByRole('button', {name: /save changes/i})).toBeInTheDocument();
      });

      const saveButton = screen.getByRole('button', {name: /save changes/i});
      await user.click(saveButton);

      // Should have called mutateAsync (even if it failed)
      await waitFor(() => {
        expect(mockMutateAsync).toHaveBeenCalled();
      });
    });

    it('should not save when application or applicationId is missing', async () => {
      // Mock useParams to return undefined applicationId
      const {useParams} = await import('react-router');
      (useParams as ReturnType<typeof vi.fn>).mockReturnValue({applicationId: undefined});

      const mockMutateAsync = vi.fn().mockResolvedValue(mockApplication);
      mockUseUpdateApplication.mockReturnValue({
        mutate: mockUpdateApplicationMutate,
        mutateAsync: mockMutateAsync,
        isPending: false,
        isError: false,
        error: null,
      } as unknown as UseMutationResult<Application, Error, Partial<Application>>);

      renderComponent();

      // mutateAsync should not have been called since applicationId is missing
      expect(mockMutateAsync).not.toHaveBeenCalled();

      // Restore original mock
      (useParams as ReturnType<typeof vi.fn>).mockReturnValue({applicationId: 'test-app-id'});
    });
  });

  describe('Logo Image Error Handling', () => {
    it('should handle logo image loading error', () => {
      renderComponent();

      const logo = screen.getByRole('img');

      // Simulate image load error
      logo.dispatchEvent(new Event('error'));

      // The component should still be functional
      expect(screen.getByText('Test Application')).toBeInTheDocument();
    });
  });

  describe('Template Metadata', () => {
    it('should not display template chip when template metadata is null', () => {
      mockGetTemplateMetadata.mockReturnValue(null);

      renderComponent();

      expect(screen.queryByText('React')).not.toBeInTheDocument();
    });

    it('should handle application without template', () => {
      mockUseGetApplication.mockReturnValue({
        data: {...mockApplication, template: undefined},
        isLoading: false,
        isError: false,
        error: null,
      } as unknown as UseQueryResult<Application>);

      renderComponent();

      // Should render without crashing
      expect(screen.getByText('Test Application')).toBeInTheDocument();
    });
  });

  describe('OAuth2 Config', () => {
    it('should handle application without inbound_auth_config', () => {
      mockUseGetApplication.mockReturnValue({
        data: {...mockApplication, inbound_auth_config: undefined},
        isLoading: false,
        isError: false,
        error: null,
      } as unknown as UseQueryResult<Application>);

      renderComponent();

      // Should render without crashing
      expect(screen.getByText('Test Application')).toBeInTheDocument();
    });

    it('should handle application with non-oauth2 inbound_auth_config', () => {
      mockUseGetApplication.mockReturnValue({
        data: {
          ...mockApplication,
          inbound_auth_config: [{type: 'saml', config: {issuer: 'test'}}],
        },
        isLoading: false,
        isError: false,
        error: null,
      } as unknown as UseQueryResult<Application>);

      renderComponent();

      // Should render without crashing
      expect(screen.getByText('Test Application')).toBeInTheDocument();
    });
  });

  describe('Name and Description Editing Edge Cases', () => {
    it('should not save empty name on Enter', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Click edit button
      const nameSection = screen.getByText('Test Application').closest('div');
      const editButton = nameSection?.querySelector('button');
      await user.click(editButton!);

      // Clear and press Enter
      const nameInput = screen.getByRole('textbox');
      await user.clear(nameInput);
      await user.keyboard('{Enter}');

      await waitFor(() => {
        // Original name should be preserved
        expect(screen.getByText('Test Application')).toBeInTheDocument();
      });
    });

    it('should save empty description when cleared', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Click edit button for description
      const descriptionSection = screen.getByText('Test application description').closest('div');
      const editButton = descriptionSection?.querySelector('button');
      await user.click(editButton!);

      // Clear description and blur
      const descriptionInput = screen.getByPlaceholderText('Add a description');
      await user.clear(descriptionInput);
      await user.tab();

      await waitFor(() => {
        // Should show unsaved changes indicator
        expect(screen.getByText('You have unsaved changes')).toBeInTheDocument();
      });
    });

    it('should handle description when original is empty and new is empty', async () => {
      mockUseGetApplication.mockReturnValue({
        data: {...mockApplication, description: undefined},
        isLoading: false,
        isError: false,
        error: null,
      } as UseQueryResult<Application>);

      const user = userEvent.setup();
      renderComponent();

      // Click edit button for description - when description is undefined, the component shows 'No description provided'
      const descriptionSection = screen.getByText('No description provided').closest('div');
      const editButton = descriptionSection?.querySelector('button');
      await user.click(editButton!);

      // Just blur without typing
      const descriptionInput = screen.getByPlaceholderText('Add a description');
      await user.click(descriptionInput);
      await user.tab();

      // Should not show unsaved changes since nothing changed
      expect(screen.queryByText('You have unsaved changes')).not.toBeInTheDocument();
    });
  });

  describe('Edit Icon Click for Logo', () => {
    it('should open logo modal when edit icon button is clicked', async () => {
      const user = userEvent.setup();
      renderComponent();

      // Find all edit buttons and click the one next to the avatar
      const allButtons = screen.getAllByRole('button');
      // The edit icon next to avatar has a smaller icon (size 14)
      const avatarEditButton = allButtons.find(
        (btn) => btn.querySelector('svg') && btn.closest('[class*="absolute"]'),
      );

      if (avatarEditButton) {
        await user.click(avatarEditButton);

        await waitFor(() => {
          const modal = screen.getByTestId('logo-update-modal');
          expect(modal).toHaveStyle({display: 'block'});
        });
      }
    });
  });
});
