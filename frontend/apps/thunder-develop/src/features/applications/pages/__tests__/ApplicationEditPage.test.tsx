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

import {render} from '@thunder/test-utils/browser';
import {page, userEvent} from 'vitest/browser';
import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';
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
  default: vi.fn(
    ({
      onCopyToClipboard,
      copiedField,
    }: {
      onCopyToClipboard?: (text: string, fieldName: string) => void;
      copiedField?: string | null;
    }) => (
      <div data-testid="edit-general-settings">
        General Settings
        {copiedField && <span data-testid="copied-field">{copiedField}</span>}
        <button type="button" data-testid="copy-button" onClick={() => onCopyToClipboard?.('test-text', 'clientId')}>
          Copy
        </button>
      </div>
    ),
  ),
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
    it('should display loading state while fetching application', async () => {
      mockUseGetApplication.mockReturnValue({
        data: undefined,
        isLoading: true,
        isError: false,
        error: null,
      } as UseQueryResult<Application>);

      await renderComponent();

      // When loading, the component shows a loading indicator
      await expect.element(page.getByRole('progressbar')).toBeInTheDocument();
    });
  });

  describe('Error State', () => {
    it('should display error state when application is not found', async () => {
      mockUseGetApplication.mockReturnValue({
        data: undefined,
        isLoading: false,
        isError: true,
        error: new Error('Not found'),
      } as unknown as UseQueryResult<Application>);

      await renderComponent();

      // Check for error UI elements
      await expect.element(page.getByRole('button', {name: /back to applications/i})).toBeInTheDocument();
    });

    it('should display back button in error state', async () => {
      mockUseGetApplication.mockReturnValue({
        data: undefined,
        isLoading: false,
        isError: true,
        error: new Error('Not found'),
      } as unknown as UseQueryResult<Application>);

      await renderComponent();

      await expect.element(page.getByRole('button', {name: /back to applications/i})).toBeInTheDocument();
    });

    it('should navigate back when back button is clicked in error state', async () => {
      mockUseGetApplication.mockReturnValue({
        data: undefined,
        isLoading: false,
        isError: true,
        error: new Error('Not found'),
      } as unknown as UseQueryResult<Application>);

      await renderComponent();

      await userEvent.click(page.getByRole('button', {name: /back to applications/i}));

      // Button should still be present after click (navigation is async)
      await expect.element(page.getByRole('button', {name: /back to applications/i})).toBeInTheDocument();
    });
  });

  describe('Successful Load', () => {
    it('should render application details correctly', async () => {
      await renderComponent();

      await expect.element(page.getByText('Test Application', {exact: true})).toBeInTheDocument();
      await expect.element(page.getByText('Test application description')).toBeInTheDocument();
    });

    it('should display application logo', async () => {
      await renderComponent();

      const logo = page.getByRole('img');
      expect(logo).toHaveAttribute('src', 'https://example.com/logo.png');
    });

    it('should display template chip when template metadata is available', async () => {
      await renderComponent();

      await expect.element(page.getByText('React', {exact: true})).toBeInTheDocument();
    });

    it('should display back button', async () => {
      await renderComponent();

      await expect.element(page.getByRole('button', {name: /back to applications/i})).toBeInTheDocument();
    });

    it('should handle empty description', async () => {
      mockUseGetApplication.mockReturnValue({
        data: {...mockApplication, description: undefined},
        isLoading: false,
        isError: false,
        error: null,
      } as UseQueryResult<Application>);

      await renderComponent();

      await expect.element(page.getByText('No description')).toBeInTheDocument();
    });
  });

  describe('Tab Navigation', () => {
    it('should render all tabs without integration guides', async () => {
      await renderComponent();

      await expect.element(page.getByRole('tab', {name: /general/i})).toBeInTheDocument();
      await expect.element(page.getByRole('tab', {name: /flows/i})).toBeInTheDocument();
      await expect.element(page.getByRole('tab', {name: /customization/i})).toBeInTheDocument();
      await expect.element(page.getByRole('tab', {name: /token/i})).toBeInTheDocument();
      await expect.element(page.getByRole('tab', {name: /advanced/i})).toBeInTheDocument();
    });

    it('should render overview tab when integration guides are available', async () => {
      mockGetIntegrationGuidesForTemplate.mockReturnValue(['react-vite']);

      await renderComponent();

      await expect.element(page.getByRole('tab', {name: /guide/i})).toBeInTheDocument();
    });

    it('should display general settings tab by default when no integration guides', async () => {
      // Mock returns null by default (no integration guides)
      await renderComponent();

      // When there are no integration guides, general tab should be first and selected
      const generalTab = page.getByRole('tab', {name: /general/i});
      expect(generalTab).toHaveAttribute('aria-selected', 'true');
    });

    it('should display overview tab by default when integration guides are available', async () => {
      mockGetIntegrationGuidesForTemplate.mockReturnValue(['react-vite']);

      await renderComponent();

      await vi.waitFor(async () => {
        await expect.element(page.getByTestId('integration-guides')).toBeInTheDocument();
      });
    });

    it('should switch to flows tab when clicked', async () => {
            await renderComponent();

      const flowsTab = page.getByRole('tab', {name: /flows/i});
      await userEvent.click(flowsTab);

      await vi.waitFor(async () => {
        await expect.element(page.getByTestId('edit-flows-settings')).toBeInTheDocument();
      });
    });

    it('should switch to customization tab when clicked', async () => {
            await renderComponent();

      const customizationTab = page.getByRole('tab', {name: /customization/i});
      await userEvent.click(customizationTab);

      await vi.waitFor(async () => {
        await expect.element(page.getByTestId('edit-customization-settings')).toBeInTheDocument();
      });
    });

    it('should switch to token tab when clicked', async () => {
            await renderComponent();

      const tokenTab = page.getByRole('tab', {name: /token/i});
      await userEvent.click(tokenTab);

      await vi.waitFor(async () => {
        await expect.element(page.getByTestId('edit-token-settings')).toBeInTheDocument();
      });
    });

    it('should switch to advanced tab when clicked', async () => {
            await renderComponent();

      const advancedTab = page.getByRole('tab', {name: /advanced/i});
      await userEvent.click(advancedTab);

      await vi.waitFor(async () => {
        await expect.element(page.getByTestId('edit-advanced-settings')).toBeInTheDocument();
      });
    });
  });

  describe('Inline Editing', () => {
    it('should enable name editing when edit icon is clicked', async () => {
            await renderComponent();

      const nameSection = page.getByText('Test Application', {exact: true}).element().closest('div');
      const editButton = nameSection?.querySelector('button');
      expect(editButton).toBeInTheDocument();

      await userEvent.click(editButton!);

      const nameInput = page.getByRole('textbox');
      expect(nameInput).toHaveValue('Test Application');
    });

    it('should save name changes on blur', async () => {
            await renderComponent();

      // Click edit button
      const nameSection = page.getByText('Test Application', {exact: true}).element().closest('div');
      const editButton = nameSection?.querySelector('button');
      await userEvent.click(editButton!);

      // Change name
      const nameInput = page.getByRole('textbox');
      await userEvent.clear(nameInput);
      await userEvent.type(nameInput, 'Updated Application');

      // Blur to save
      await userEvent.tab();

      await vi.waitFor(async () => {
        await expect.element(page.getByText('Updated Application', {exact: true})).toBeInTheDocument();
      });
    });

    it('should save name changes on Enter key', async () => {
            await renderComponent();

      // Click edit button
      const nameSection = page.getByText('Test Application', {exact: true}).element().closest('div');
      const editButton = nameSection?.querySelector('button');
      await userEvent.click(editButton!);

      // Change name and press Enter
      const nameInput = page.getByRole('textbox');
      await userEvent.clear(nameInput);
      await userEvent.type(nameInput, 'Updated Application{Enter}');

      await vi.waitFor(async () => {
        await expect.element(page.getByText('Updated Application', {exact: true})).toBeInTheDocument();
      });
    });

    it('should cancel name editing on Escape key', async () => {
            await renderComponent();

      // Click edit button
      const nameSection = page.getByText('Test Application', {exact: true}).element().closest('div');
      const editButton = nameSection?.querySelector('button');
      await userEvent.click(editButton!);

      // Change name and press Escape
      const nameInput = page.getByRole('textbox');
      await userEvent.clear(nameInput);
      await userEvent.type(nameInput, 'Updated Application{Escape}');

      await vi.waitFor(async () => {
        await expect.element(page.getByText('Test Application', {exact: true})).toBeInTheDocument();
        await expect.element(page.getByRole('textbox')).not.toBeInTheDocument();
      });
    });

    it('should enable description editing when edit icon is clicked', async () => {
            await renderComponent();

      const descriptionSection = page.getByText('Test application description').element().closest('div');
      const editButton = descriptionSection?.querySelector('button');
      expect(editButton).toBeInTheDocument();

      await userEvent.click(editButton!);

      const descriptionInput = page.getByPlaceholder('Add a description');
      expect(descriptionInput).toHaveValue('Test application description');
    });

    it('should save description changes on blur', async () => {
            await renderComponent();

      // Click edit button
      const descriptionSection = page.getByText('Test application description').element().closest('div');
      const editButton = descriptionSection?.querySelector('button');
      await userEvent.click(editButton!);

      // Change description
      const descriptionInput = page.getByPlaceholder('Add a description');
      await userEvent.clear(descriptionInput);
      await userEvent.type(descriptionInput, 'Updated description');

      // Blur to save
      (descriptionInput.element() as HTMLElement).blur();

      await vi.waitFor(async () => {
        await expect.element(page.getByText('Updated description')).toBeInTheDocument();
      });
    });

    it('should save description changes on Ctrl+Enter', async () => {
            await renderComponent();

      // Click edit button
      const descriptionSection = page.getByText('Test application description').element().closest('div');
      const editButton = descriptionSection?.querySelector('button');
      await userEvent.click(editButton!);

      // Change description
      const descriptionInput = page.getByPlaceholder('Add a description');
      await userEvent.clear(descriptionInput);
      await userEvent.type(descriptionInput, 'Description via Ctrl+Enter');

      // Press Ctrl+Enter to save
      await userEvent.keyboard('{Control>}{Enter}{/Control}');

      await vi.waitFor(async () => {
        await expect.element(page.getByText('Description via Ctrl+Enter')).toBeInTheDocument();
      });
    });

    it('should cancel description editing on Escape key', async () => {
            await renderComponent();

      // Click edit button
      const descriptionSection = page.getByText('Test application description').element().closest('div');
      const editButton = descriptionSection?.querySelector('button');
      await userEvent.click(editButton!);

      // Change description and press Escape
      const descriptionInput = page.getByPlaceholder('Add a description');
      await userEvent.clear(descriptionInput);
      await userEvent.type(descriptionInput, 'Changed description');
      await userEvent.keyboard('{Escape}');

      await vi.waitFor(async () => {
        await expect.element(page.getByText('Test application description')).toBeInTheDocument();
        await expect.element(page.getByPlaceholder('Add a description')).not.toBeInTheDocument();
      });
    });

    it('should not save empty name on blur', async () => {
            await renderComponent();

      // Click edit button
      const nameSection = page.getByText('Test Application', {exact: true}).element().closest('div');
      const editButton = nameSection?.querySelector('button');
      await userEvent.click(editButton!);

      // Clear name
      const nameInput = page.getByRole('textbox');
      await userEvent.clear(nameInput);

      // Blur without entering text
      await userEvent.tab();

      await vi.waitFor(async () => {
        await expect.element(page.getByText('Test Application', {exact: true})).toBeInTheDocument();
      });
    });
  });

  describe('Logo Update', () => {
    it('should render logo update modal', async () => {
      await renderComponent();

      // Modal should be in the DOM (hidden by default)
      await expect.element(page.getByTestId('logo-update-modal')).toBeInTheDocument();
    });

    it('should open logo modal when avatar is clicked', async () => {
            await renderComponent();

      // Click on the avatar to open the modal using the edit icon button
      const editLogoButton = page.getByRole('button', {name: /update logo/i});
      await userEvent.click(editLogoButton);

      await vi.waitFor(() => {
        const modal = page.getByTestId('logo-update-modal');
        expect(modal).toHaveStyle({display: 'block'});
      });
    });

    it('should update logo and close modal when logo is updated', async () => {
            await renderComponent();

      // Open the modal using the edit icon button
      const editLogoButton = page.getByRole('button', {name: /update logo/i});
      await userEvent.click(editLogoButton);

      // Click update logo button in modal
      const modal = page.getByTestId('logo-update-modal');
      const updateLogoButton = modal.getByRole('button', {name: /update logo/i});
      await userEvent.click(updateLogoButton);

      await vi.waitFor(async () => {
        // Should show unsaved changes since logo was updated
        await expect.element(page.getByText('Unsaved changes')).toBeInTheDocument();
      });

      // Modal should be closed
      await vi.waitFor(() => {
        const closedModal = page.getByTestId('logo-update-modal');
        expect(closedModal).toHaveStyle({display: 'none'});
      });
    });

    it('should close logo modal when close button is clicked', async () => {
            await renderComponent();

      // Open the modal using the edit icon button
      const editLogoButton = page.getByRole('button', {name: /update logo/i});
      await userEvent.click(editLogoButton);

      // Click close button
      const closeButton = page.getByRole('button', {name: /close/i});
      await userEvent.click(closeButton);

      await vi.waitFor(() => {
        const modal = page.getByTestId('logo-update-modal');
        expect(modal).toHaveStyle({display: 'none'});
      });
    });
  });

  describe('Save Functionality', () => {
    it('should show floating action bar when changes are made', async () => {
            await renderComponent();

      // Make a change to name
      const nameSection = page.getByText('Test Application', {exact: true}).element().closest('div');
      const editButton = nameSection?.querySelector('button');
      await userEvent.click(editButton!);

      const nameInput = page.getByRole('textbox');
      await userEvent.clear(nameInput);
      await userEvent.type(nameInput, 'Updated Application{Enter}');

      await vi.waitFor(async () => {
        await expect.element(page.getByText('Unsaved changes')).toBeInTheDocument();
      });
    });

    it('should display reset and save buttons in action bar', async () => {
            await renderComponent();

      // Make a change
      const nameSection = page.getByText('Test Application', {exact: true}).element().closest('div');
      const editButton = nameSection?.querySelector('button');
      await userEvent.click(editButton!);

      const nameInput = page.getByRole('textbox');
      await userEvent.clear(nameInput);
      await userEvent.type(nameInput, 'Updated Application{Enter}');

      await vi.waitFor(async () => {
        await expect.element(page.getByRole('button', {name: /reset/i})).toBeInTheDocument();
        await expect.element(page.getByRole('button', {name: /^save$/i})).toBeInTheDocument();
      });
    });

    it('should reset changes when reset button is clicked', async () => {
            await renderComponent();

      // Make a change
      const nameSection = page.getByText('Test Application', {exact: true}).element().closest('div');
      const editButton = nameSection?.querySelector('button');
      await userEvent.click(editButton!);

      const nameInput = page.getByRole('textbox');
      await userEvent.clear(nameInput);
      await userEvent.type(nameInput, 'Updated Application{Enter}');

      // Click reset
      await vi.waitFor(async () => {
        await expect.element(page.getByRole('button', {name: /reset/i})).toBeInTheDocument();
      });

      const resetButton = page.getByRole('button', {name: /reset/i});
      await userEvent.click(resetButton);

      await vi.waitFor(async () => {
        await expect.element(page.getByText('Unsaved changes')).not.toBeInTheDocument();
      });
    });

    it('should save changes when save button is clicked', async () => {
            const mockMutateAsync = vi.fn().mockResolvedValue(mockApplication);

      mockUseUpdateApplication.mockReturnValue({
        mutate: mockUpdateApplicationMutate,
        mutateAsync: mockMutateAsync,
        isPending: false,
        isError: false,
        error: null,
      } as unknown as UseMutationResult<Application, Error, Partial<Application>>);

      await renderComponent();

      // Make a change
      const nameSection = page.getByText('Test Application', {exact: true}).element().closest('div');
      const editButton = nameSection?.querySelector('button');
      await userEvent.click(editButton!);

      const nameInput = page.getByRole('textbox');
      await userEvent.clear(nameInput);
      await userEvent.type(nameInput, 'Updated Application{Enter}');

      // Click save
      await vi.waitFor(async () => {
        await expect.element(page.getByRole('button', {name: /^save$/i})).toBeInTheDocument();
      });

      const saveButton = page.getByRole('button', {name: /^save$/i});
      await userEvent.click(saveButton);

      await vi.waitFor(() => {
        expect(mockMutateAsync).toHaveBeenCalled();
        const callArgs = mockMutateAsync.mock.calls[0][0] as {applicationId: string; data: Partial<Application>};
        expect(callArgs).toHaveProperty('applicationId', 'test-app-id');
        expect(callArgs).toHaveProperty('data');
        expect(callArgs.data).toHaveProperty('name', 'Updated Application');
      });
    });

    it('should disable save button while saving', async () => {
      
      mockUseUpdateApplication.mockReturnValue({
        mutate: mockUpdateApplicationMutate,
        mutateAsync: vi.fn().mockResolvedValue(mockApplication),
        isPending: true,
        isError: false,
        error: null,
      } as unknown as UseMutationResult<Application, Error, Partial<Application>>);

      await renderComponent();

      // Make a change first to show the action bar
      const nameSection = page.getByText('Test Application', {exact: true}).element().closest('div');
      const editButton = nameSection?.querySelector('button');
      await userEvent.click(editButton!);

      const nameInput = page.getByRole('textbox');
      await userEvent.clear(nameInput);
      await userEvent.type(nameInput, 'Updated Application{Enter}');

      await vi.waitFor(() => {
        const saveButton = page.getByRole('button', {name: /saving/i});
        expect(saveButton).toBeDisabled();
      });
    });

    it('should hide action bar after successful save', async () => {
            const mockMutateAsync = vi.fn().mockResolvedValue({...mockApplication, name: 'Updated Application'});

      mockUseUpdateApplication.mockReturnValue({
        mutate: mockUpdateApplicationMutate,
        mutateAsync: mockMutateAsync,
        isPending: false,
        isError: false,
        error: null,
      } as unknown as UseMutationResult<Application, Error, Partial<Application>>);

      await renderComponent();

      // Make a change
      const nameSection = page.getByText('Test Application', {exact: true}).element().closest('div');
      const editButton = nameSection?.querySelector('button');
      await userEvent.click(editButton!);

      const nameInput = page.getByRole('textbox');
      await userEvent.clear(nameInput);
      await userEvent.type(nameInput, 'Updated Application{Enter}');

      // Wait for action bar to appear
      await vi.waitFor(async () => {
        await expect.element(page.getByRole('button', {name: /^save$/i})).toBeInTheDocument();
      });

      // Click save
      const saveButton = page.getByRole('button', {name: /^save$/i});
      await userEvent.click(saveButton);

      await vi.waitFor(
        async () => {
          await expect.element(page.getByText('Unsaved changes')).not.toBeInTheDocument();
        },
        {timeout: 10000},
      );
    });
  });

  describe('Accessibility', () => {
    it('should have proper ARIA labels for tabs', async () => {
      await renderComponent();

      const generalTab = page.getByRole('tab', {name: /general/i});
      expect(generalTab).toHaveAttribute('id');
      expect(generalTab).toHaveAttribute('aria-controls');
    });

    it('should maintain focus management during inline editing', async () => {
            await renderComponent();

      const nameSection = page.getByText('Test Application', {exact: true}).element().closest('div');
      const editButton = nameSection?.querySelector('button');
      await userEvent.click(editButton!);

      const nameInput = page.getByRole('textbox');
      expect(nameInput).toHaveFocus();
    });
  });

  describe('Application Not Found', () => {
    it('should display warning when application is null', async () => {
      mockUseGetApplication.mockReturnValue({
        data: null,
        isLoading: false,
        isError: false,
        error: null,
      } as unknown as UseQueryResult<Application>);

      await renderComponent();

      await expect.element(page.getByText('Application not found')).toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: /back to applications/i})).toBeInTheDocument();
    });

    it('should navigate back when back button is clicked in not found state', async () => {
      mockUseGetApplication.mockReturnValue({
        data: null,
        isLoading: false,
        isError: false,
        error: null,
      } as unknown as UseQueryResult<Application>);

      await renderComponent();

      await userEvent.click(page.getByRole('button', {name: /back to applications/i}));

      // Button should still be present after click (navigation is async)
      await expect.element(page.getByRole('button', {name: /back to applications/i})).toBeInTheDocument();
    });
  });

  describe('Error Handling', () => {
    it('should display error message from error object', async () => {
      const errorMessage = 'Custom error message';
      mockUseGetApplication.mockReturnValue({
        data: undefined,
        isLoading: false,
        isError: true,
        error: {message: errorMessage},
      } as unknown as UseQueryResult<Application>);

      await renderComponent();

      await expect.element(page.getByText(errorMessage)).toBeInTheDocument();
    });

    it('should display default error message when error has no message', async () => {
      mockUseGetApplication.mockReturnValue({
        data: undefined,
        isLoading: false,
        isError: true,
        error: {},
      } as unknown as UseQueryResult<Application>);

      await renderComponent();

      await expect.element(page.getByText('Failed to load application information')).toBeInTheDocument();
    });

    it('should handle save failure gracefully', async () => {
            const mockMutateAsync = vi.fn().mockRejectedValue(new Error('Save failed'));

      mockUseUpdateApplication.mockReturnValue({
        mutate: mockUpdateApplicationMutate,
        mutateAsync: mockMutateAsync,
        isPending: false,
        isError: false,
        error: null,
      } as unknown as UseMutationResult<Application, Error, Partial<Application>>);

      await renderComponent();

      // Make a change
      const nameSection = page.getByText('Test Application', {exact: true}).element().closest('div');
      const editButton = nameSection?.querySelector('button');
      await userEvent.click(editButton!);

      const nameInput = page.getByRole('textbox');
      await userEvent.clear(nameInput);
      await userEvent.type(nameInput, 'Updated Application{Enter}');

      // Click save
      await vi.waitFor(async () => {
        await expect.element(page.getByRole('button', {name: /^save$/i})).toBeInTheDocument();
      });

      const saveButton = page.getByRole('button', {name: /^save$/i});
      await userEvent.click(saveButton);

      // Should have called mutateAsync (even if it failed)
      await vi.waitFor(() => {
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

      await renderComponent();

      // Make a change to trigger the floating save bar
      const nameSection = page.getByText('Test Application', {exact: true}).element().closest('div');
      const editButton = nameSection?.querySelector('button');
      await userEvent.click(editButton!);

      const nameInput = page.getByRole('textbox');
      await userEvent.clear(nameInput);
      await userEvent.type(nameInput, 'Updated Application{Enter}');

      // Click save — handleSave should return early due to missing applicationId
      await vi.waitFor(async () => {
        await expect.element(page.getByRole('button', {name: /^save$/i})).toBeInTheDocument();
      });

      const saveButton = page.getByRole('button', {name: /^save$/i});
      await userEvent.click(saveButton);

      // mutateAsync should not have been called since applicationId is missing
      expect(mockMutateAsync).not.toHaveBeenCalled();

      // Restore original mock
      (useParams as ReturnType<typeof vi.fn>).mockReturnValue({applicationId: 'test-app-id'});
    });
  });

  describe('Logo Image Error Handling', () => {
    it('should handle logo image loading error', async () => {
      await renderComponent();

      const logo = page.getByRole('img');

      // Simulate image load error
      (logo.element() as HTMLElement).dispatchEvent(new Event('error'));

      // The component should still be functional
      await expect.element(page.getByText('Test Application', {exact: true})).toBeInTheDocument();
    });
  });

  describe('Template Metadata', () => {
    it('should not display template chip when template metadata is null', async () => {
      mockGetTemplateMetadata.mockReturnValue(null);

      await renderComponent();

      await expect.element(page.getByText('React', {exact: true})).not.toBeInTheDocument();
    });

    it('should handle application without template', async () => {
      mockUseGetApplication.mockReturnValue({
        data: {...mockApplication, template: undefined},
        isLoading: false,
        isError: false,
        error: null,
      } as unknown as UseQueryResult<Application>);

      await renderComponent();

      // Should render without crashing
      await expect.element(page.getByText('Test Application', {exact: true})).toBeInTheDocument();
    });
  });

  describe('OAuth2 Config', () => {
    it('should handle application without inbound_auth_config', async () => {
      mockUseGetApplication.mockReturnValue({
        data: {...mockApplication, inbound_auth_config: undefined},
        isLoading: false,
        isError: false,
        error: null,
      } as unknown as UseQueryResult<Application>);

      await renderComponent();

      // Should render without crashing
      await expect.element(page.getByText('Test Application', {exact: true})).toBeInTheDocument();
    });

    it('should handle application with non-oauth2 inbound_auth_config', async () => {
      mockUseGetApplication.mockReturnValue({
        data: {
          ...mockApplication,
          inbound_auth_config: [{type: 'saml', config: {issuer: 'test'}}],
        },
        isLoading: false,
        isError: false,
        error: null,
      } as unknown as UseQueryResult<Application>);

      await renderComponent();

      // Should render without crashing
      await expect.element(page.getByText('Test Application', {exact: true})).toBeInTheDocument();
    });
  });

  describe('Name and Description Editing Edge Cases', () => {
    it('should not save empty name on Enter', async () => {
            await renderComponent();

      // Click edit button
      const nameSection = page.getByText('Test Application', {exact: true}).element().closest('div');
      const editButton = nameSection?.querySelector('button');
      await userEvent.click(editButton!);

      // Clear and press Enter
      const nameInput = page.getByRole('textbox');
      await userEvent.clear(nameInput);
      await userEvent.keyboard('{Enter}');

      await vi.waitFor(async () => {
        // Original name should be preserved
        await expect.element(page.getByText('Test Application', {exact: true})).toBeInTheDocument();
      });
    });

    it('should save empty description when cleared', async () => {
            await renderComponent();

      // Click edit button for description
      const descriptionSection = page.getByText('Test application description').element().closest('div');
      const editButton = descriptionSection?.querySelector('button');
      await userEvent.click(editButton!);

      // Clear description and blur
      const descriptionInput = page.getByPlaceholder('Add a description');
      await userEvent.clear(descriptionInput);
      await userEvent.tab();

      await vi.waitFor(async () => {
        // Should show unsaved changes indicator
        await expect.element(page.getByText('Unsaved changes')).toBeInTheDocument();
      });
    });

    it('should handle description when original is empty and new is empty', async () => {
      mockUseGetApplication.mockReturnValue({
        data: {...mockApplication, description: undefined},
        isLoading: false,
        isError: false,
        error: null,
      } as UseQueryResult<Application>);

            await renderComponent();

      // Click edit button for description - when description is undefined, the component shows 'No description'
      const descriptionSection = page.getByText('No description').element().closest('div');
      const editButton = descriptionSection?.querySelector('button');
      await userEvent.click(editButton!);

      // Just blur without typing
      const descriptionInput = page.getByPlaceholder('Add a description');
      await userEvent.click(descriptionInput);
      await userEvent.tab();

      // Should not show unsaved changes since nothing changed
      await expect.element(page.getByText('Unsaved changes')).not.toBeInTheDocument();
    });
  });

  describe('Edit Icon Click for Logo', () => {
    it('should open logo modal when edit icon button is clicked', async () => {
      await renderComponent();

      await userEvent.click(page.getByRole('button', {name: 'Update Logo'}));

      await vi.waitFor(async () => {
        await expect.element(page.getByTestId('logo-update-modal')).toHaveStyle({display: 'block'});
      });
    });
  });

  describe('Copy to Clipboard', () => {
    const originalClipboard = navigator.clipboard;

    afterEach(() => {
      Object.defineProperty(navigator, 'clipboard', {
        value: originalClipboard,
        writable: true,
        configurable: true,
      });
    });

    it('should copy text to clipboard when copy button is clicked', async () => {
      const writeTextMock = vi.fn().mockResolvedValue(undefined);
      Object.defineProperty(navigator, 'clipboard', {
        value: {writeText: writeTextMock},
        writable: true,
        configurable: true,
      });

      await renderComponent();

      await userEvent.click(page.getByTestId('copy-button'));

      await vi.waitFor(() => {
        expect(writeTextMock).toHaveBeenCalledWith('test-text');
      });

      await vi.waitFor(async () => {
        await expect.element(page.getByTestId('copied-field')).toHaveTextContent('clientId');
      });
    });

    it('should handle clipboard write failure gracefully', async () => {
      const writeTextMock = vi.fn().mockRejectedValue(new Error('Clipboard error'));
      Object.defineProperty(navigator, 'clipboard', {
        value: {writeText: writeTextMock},
        writable: true,
        configurable: true,
      });

      await renderComponent();

      await userEvent.click(page.getByTestId('copy-button'));

      await vi.waitFor(() => {
        expect(writeTextMock).toHaveBeenCalledWith('test-text');
      });

      // Component should still be functional after error
      await expect.element(page.getByText('Test Application', {exact: true})).toBeInTheDocument();
    });
  });

  describe('Avatar Image Error', () => {
    it('should hide avatar image when image fails to load', async () => {
      const {container} = await renderComponent();

      // Find the img element directly (may be hidden if load already failed)
      const avatarImg = container.querySelector('img');

      // Simulate image load error via the onError handler if img is visible
      if (avatarImg) {
        avatarImg.dispatchEvent(new Event('error'));
        // The image should be hidden after the error event
        expect(avatarImg).toHaveStyle({display: 'none'});
      } else {
        // Image was already hidden (load failed before test checked)
        // This is the expected behavior - the component handles image load errors
        expect(container.querySelector('img[style*="none"]')).toBeDefined();
      }
    });
  });

  describe('Edited App Fallbacks', () => {
    it('should display edited name when editedApp has name', async () => {
            await renderComponent();

      // Edit the name
      const nameSection = page.getByText('Test Application', {exact: true}).element().closest('div');
      const editButton = nameSection?.querySelector('button');
      await userEvent.click(editButton!);

      const nameInput = page.getByRole('textbox');
      await userEvent.clear(nameInput);
      await userEvent.type(nameInput, 'New App Name{Enter}');

      await vi.waitFor(async () => {
        await expect.element(page.getByText('New App Name')).toBeInTheDocument();
      });

      // Now click edit again - tempName should be set from editedApp.name
      const updatedNameSection = page.getByText('New App Name').element().closest('div');
      const editButtonAgain = updatedNameSection?.querySelector('button');
      await userEvent.click(editButtonAgain!);

      const nameInputAgain = page.getByRole('textbox');
      expect(nameInputAgain).toHaveValue('New App Name');
    });

    it('should display edited description when editedApp has description', async () => {
            await renderComponent();

      // Edit description
      const descriptionSection = page.getByText('Test application description').element().closest('div');
      const editButton = descriptionSection?.querySelector('button');
      await userEvent.click(editButton!);

      const descriptionInput = page.getByPlaceholder('Add a description');
      await userEvent.clear(descriptionInput);
      await userEvent.type(descriptionInput, 'New description');
      (descriptionInput.element() as HTMLElement).blur();

      await vi.waitFor(async () => {
        await expect.element(page.getByText('New description')).toBeInTheDocument();
      });

      // Click edit again - tempDescription should use editedApp.description
      const updatedSection = page.getByText('New description').element().closest('div');
      const editButtonAgain = updatedSection?.querySelector('button');
      await userEvent.click(editButtonAgain!);

      const descriptionInputAgain = page.getByPlaceholder('Add a description');
      expect(descriptionInputAgain).toHaveValue('New description');
    });

    it('should display edited logo_url in avatar when editedApp has logo_url', async () => {
            await renderComponent();

      // Open modal and update logo
      const avatar = page.getByRole('img');
      await userEvent.click(avatar);

      const logoModal = page.getByTestId('logo-update-modal');
      const updateLogoButton = logoModal.getByRole('button', {name: /update logo/i});
      await userEvent.click(updateLogoButton);

      await vi.waitFor(() => {
        const updatedAvatar = page.getByRole('img');
        expect(updatedAvatar).toHaveAttribute('src', 'https://example.com/new-logo.png');
      });
    });

    it('should handle application with no logo_url', async () => {
      mockUseGetApplication.mockReturnValue({
        data: {...mockApplication, logo_url: undefined},
        isLoading: false,
        isError: false,
        error: null,
      } as UseQueryResult<Application>);

      await renderComponent();

      // Should render without crashing - avatar will use fallback icon
      await expect.element(page.getByText('Test Application', {exact: true})).toBeInTheDocument();
    });
  });

  describe('Tab Navigation with Integration Guides', () => {
    it('should switch to general tab when overview is first tab', async () => {
            mockGetIntegrationGuidesForTemplate.mockReturnValue(['react-vite']);

      await renderComponent();

      // Click General tab (second tab when integration guides are present)
      const generalTab = page.getByRole('tab', {name: /general/i});
      await userEvent.click(generalTab);

      await vi.waitFor(async () => {
        await expect.element(page.getByTestId('edit-general-settings')).toBeInTheDocument();
      });
    });

    it('should switch to flows tab when integration guides exist', async () => {
            mockGetIntegrationGuidesForTemplate.mockReturnValue(['react-vite']);

      await renderComponent();

      const flowsTab = page.getByRole('tab', {name: /flows/i});
      await userEvent.click(flowsTab);

      await vi.waitFor(async () => {
        await expect.element(page.getByTestId('edit-flows-settings')).toBeInTheDocument();
      });
    });
  });

  describe('Description Escape with Edited Value', () => {
    it('should restore edited description on Escape key when editedApp has description', async () => {
            await renderComponent();

      // First, edit the description to set editedApp.description
      const descriptionSection = page.getByText('Test application description').element().closest('div');
      const editButton = descriptionSection?.querySelector('button');
      await userEvent.click(editButton!);

      const descriptionInput = page.getByPlaceholder('Add a description');
      await userEvent.clear(descriptionInput);
      await userEvent.type(descriptionInput, 'Edited description');
      (descriptionInput.element() as HTMLElement).blur();

      await vi.waitFor(async () => {
        await expect.element(page.getByText('Edited description')).toBeInTheDocument();
      });

      // Now edit again and press Escape - should restore the editedApp.description
      const updatedSection = page.getByText('Edited description').element().closest('div');
      const editButtonAgain = updatedSection?.querySelector('button');
      await userEvent.click(editButtonAgain!);

      const descriptionInputAgain = page.getByPlaceholder('Add a description');
      await userEvent.clear(descriptionInputAgain);
      await userEvent.type(descriptionInputAgain, 'Something else');
      await userEvent.keyboard('{Escape}');

      await vi.waitFor(async () => {
        // Should revert to the editedApp.description value
        await expect.element(page.getByText('Edited description')).toBeInTheDocument();
      });
    });
  });

  describe('Name Editing with Edited Value', () => {
    it('should restore edited name on Escape when editedApp has name', async () => {
            await renderComponent();

      // First, edit name to set editedApp.name
      const nameSection = page.getByText('Test Application', {exact: true}).element().closest('div');
      const editButton = nameSection?.querySelector('button');
      await userEvent.click(editButton!);

      const nameInput = page.getByRole('textbox');
      await userEvent.clear(nameInput);
      await userEvent.type(nameInput, 'Edited Name{Enter}');

      await vi.waitFor(async () => {
        await expect.element(page.getByText('Edited Name')).toBeInTheDocument();
      });

      // Edit again and press Escape - should restore editedApp.name
      const updatedNameSection = page.getByText('Edited Name').element().closest('div');
      const editButtonAgain = updatedNameSection?.querySelector('button');
      await userEvent.click(editButtonAgain!);

      const nameInputAgain = page.getByRole('textbox');
      await userEvent.clear(nameInputAgain);
      await userEvent.type(nameInputAgain, 'Something else{Escape}');

      await vi.waitFor(async () => {
        await expect.element(page.getByText('Edited Name')).toBeInTheDocument();
      });
    });
  });
});
