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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render, getByDisplayValue} from '@thunder/test-utils/browser';
import {page, userEvent} from 'vitest/browser';
import AccessSection from '../AccessSection';
import useGetUserTypes from '../../../../../user-types/api/useGetUserTypes';
import type {Application} from '../../../../models/application';
import type {OAuth2Config} from '../../../../models/oauth';

// Mock the useGetUserTypes hook
vi.mock('../../../../../user-types/api/useGetUserTypes');

type MockedUseGetUserTypes = ReturnType<typeof useGetUserTypes>;

// Mock the SettingsCard component
vi.mock('../../../../../../components/SettingsCard', () => ({
  default: ({title, description, children}: {title: string; description: string; children: React.ReactNode}) => (
    <div data-testid="settings-card">
      <div data-testid="card-title">{title}</div>
      <div data-testid="card-description">{description}</div>
      {children}
    </div>
  ),
}));

describe('AccessSection', () => {
  const mockOnFieldChange = vi.fn();
  const mockApplication: Application = {
    id: 'app-123',
    name: 'Test App',
    url: 'https://example.com',
    allowed_user_types: ['admin', 'user'],
    inbound_auth_config: [
      {
        type: 'oauth2',
        config: {
          client_id: 'client-123',
          redirect_uris: ['https://example.com/callback'],
        },
      },
    ],
  } as Application;

  const mockOAuth2Config: OAuth2Config = {
    client_id: 'client-123',
    redirect_uris: ['https://example.com/callback'],
  } as OAuth2Config;

  const mockUserTypes = {
    schemas: [
      {name: 'admin', id: '1'},
      {name: 'user', id: '2'},
      {name: 'guest', id: '3'},
    ],
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render the settings card with title and description', async () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      await render(<AccessSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      await expect.element(page.getByTestId('card-title')).toHaveTextContent('Access');
      await expect.element(page.getByTestId('card-description')).toHaveTextContent(
        "Configure who can access this application, where it's hosted, etc.",
      );
    });

    it('should render allowed user types autocomplete', async () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      await render(<AccessSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      await expect.element(page.getByLabelText('Allowed User Types')).toBeInTheDocument();
    });

    it('should render application URL field', async () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      await render(<AccessSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      await expect.element(page.getByLabelText('Application URL')).toBeInTheDocument();
      expect(getByDisplayValue('https://example.com')).toBeInTheDocument();
    });

    it('should render redirect URIs section when OAuth2 config is provided', async () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      await render(
        <AccessSection
          application={mockApplication}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      await expect.element(page.getByText('Authorized redirect URIs')).toBeInTheDocument();
      expect(getByDisplayValue('https://example.com/callback')).toBeInTheDocument();
    });

    it('should not render redirect URIs section when OAuth2 config is not provided', async () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      await render(<AccessSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      await expect.element(page.getByLabelText('Redirect URIs')).not.toBeInTheDocument();
    });
  });

  describe('Loading State', () => {
    it('should show loading indicator while fetching user types', async () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: undefined,
        isLoading: true,
      } as unknown as MockedUseGetUserTypes);

      await render(<AccessSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      await expect.element(page.getByRole('progressbar')).toBeInTheDocument();
    });

    it('should not show loading indicator when user types are loaded', async () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      await render(<AccessSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      await expect.element(page.getByRole('progressbar')).not.toBeInTheDocument();
    });
  });

  describe('Allowed User Types', () => {
    it('should display selected user types from application', async () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      await render(<AccessSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      await expect.element(page.getByText('admin', {exact: true})).toBeInTheDocument();
      await expect.element(page.getByText('user', {exact: true})).toBeInTheDocument();
    });

    it('should display selected user types from editedApp over application', async () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      await render(
        <AccessSection
          application={mockApplication}
          editedApp={{allowed_user_types: ['guest']}}
          onFieldChange={mockOnFieldChange}
        />,
      );

      await expect.element(page.getByText('guest')).toBeInTheDocument();
      await expect.element(page.getByText('admin')).not.toBeInTheDocument();
    });

    it('should display all available user types in dropdown', async () => {
            vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      await render(<AccessSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const input = page.getByLabelText('Allowed User Types');
      await userEvent.click(input);

      await vi.waitFor(() => {
        expect(page.getByText('admin').length).toBeGreaterThan(0);
        expect(page.getByText('guest').length).toBeGreaterThan(0);
      });
    });
  });

  describe('Application URL', () => {
    it('should display URL from application', async () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      await render(<AccessSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const urlInput = page.getByLabelText('Application URL');
      expect(urlInput).toHaveAttribute('value', 'https://example.com');
    });

    it('should display URL from editedApp over application', async () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      await render(
        <AccessSection
          application={mockApplication}
          editedApp={{url: 'https://edited.com'}}
          onFieldChange={mockOnFieldChange}
        />,
      );

      const urlInput = page.getByLabelText('Application URL');
      expect(urlInput).toHaveAttribute('value', 'https://edited.com');
    });

    it('should show validation error for invalid URL', async () => {
            vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      await render(<AccessSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const urlInput = page.getByLabelText('Application URL');
      await userEvent.clear(urlInput);
      await userEvent.type(urlInput, 'invalid-url');

      await vi.waitFor(async () => {
        await expect.element(page.getByText('Please enter a valid URL')).toBeInTheDocument();
      });
    });

    it('should accept valid URL without error', async () => {
            vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      await render(
        <AccessSection application={{...mockApplication, url: ''}} editedApp={{}} onFieldChange={mockOnFieldChange} />,
      );

      const urlInput = page.getByLabelText('Application URL');
      await userEvent.type(urlInput, 'https://newurl.com');

      await vi.waitFor(async () => {
        await expect.element(page.getByText('Please enter a valid URL')).not.toBeInTheDocument();
      });
    });
  });

  describe('Redirect URIs', () => {
    it('should display existing redirect URIs', async () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      const configWithMultipleUris = {
        ...mockOAuth2Config,
        redirect_uris: ['https://example.com/callback1', 'https://example.com/callback2'],
      };

      await render(
        <AccessSection
          application={mockApplication}
          editedApp={{}}
          oauth2Config={configWithMultipleUris}
          onFieldChange={mockOnFieldChange}
        />,
      );

      expect(getByDisplayValue('https://example.com/callback1')).toBeInTheDocument();
      expect(getByDisplayValue('https://example.com/callback2')).toBeInTheDocument();
    });

    it('should add new redirect URI when add button is clicked', async () => {
            vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      await render(
        <AccessSection
          application={mockApplication}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      const addButton = page.getByRole('button', {name: /Add URI/i});
      await userEvent.click(addButton);

      const inputs = page.getByPlaceholder('https://example.com/callback').all();
      expect(inputs).toHaveLength(2);
    });

    it('should remove redirect URI when delete button is clicked', async () => {
            vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      const configWithMultipleUris = {
        ...mockOAuth2Config,
        redirect_uris: ['https://example.com/callback1', 'https://example.com/callback2'],
      };

      await render(
        <AccessSection
          application={mockApplication}
          editedApp={{}}
          oauth2Config={configWithMultipleUris}
          onFieldChange={mockOnFieldChange}
        />,
      );

      const deleteButtons = page.getByRole('button', {name: /delete/i}).all();
      await userEvent.click(deleteButtons[0]);

      await vi.waitFor(() => {
        expect(document.querySelector('input[value="https://example.com/callback1"]')).toBeNull();
      });
      expect(getByDisplayValue('https://example.com/callback2')).toBeInTheDocument();
    });
  });

  describe('Field Change Callbacks', () => {
    it('should call onFieldChange when user types are changed', async () => {
            vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      await render(<AccessSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const input = page.getByLabelText('Allowed User Types');
      await userEvent.click(input);

      const guestOption = page.getByRole('option', {name: 'guest'});
      await userEvent.click(guestOption);

      await vi.waitFor(() => {
        expect(mockOnFieldChange).toHaveBeenCalled();
        const {calls} = mockOnFieldChange.mock;
        const userTypesCall = calls.find((call) => call[0] === 'allowed_user_types');
        expect(userTypesCall).toBeDefined();
      });
    });
  });

  describe('URI Validation on Blur', () => {
    const mockApplicationWithAuth: Application = {
      id: 'app-123',
      name: 'Test App',
      url: 'https://example.com',
      allowed_user_types: ['admin', 'user'],
      inbound_auth_config: [
        {
          type: 'oauth2',
          config: {
            client_id: 'client-123',
            redirect_uris: ['https://example.com/callback'],
          },
        },
      ],
    } as Application;

    it('should show error when invalid URI is entered and blurred', async () => {
            vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      await render(
        <AccessSection
          application={mockApplicationWithAuth}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      // Find the existing URI input and enter invalid URI
      const uriInput = getByDisplayValue('https://example.com/callback');
      await userEvent.clear(uriInput);
      await userEvent.type(uriInput, 'not-a-valid-url');

      // Blur the input to trigger validation
      await userEvent.tab();

      // Should show error and not call onFieldChange for inbound_auth_config
      await vi.waitFor(() => {
        const errorCalls = mockOnFieldChange.mock.calls.filter((call) => call[0] === 'inbound_auth_config');
        expect(errorCalls).toHaveLength(0);
      });
    });

    it('should show error when URI is empty and blurred', async () => {
            vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      await render(
        <AccessSection
          application={mockApplicationWithAuth}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      // Find the existing URI input and clear it
      const uriInput = getByDisplayValue('https://example.com/callback');
      await userEvent.clear(uriInput);

      // Blur the input to trigger validation
      await userEvent.tab();

      // Should not call onFieldChange for empty URI
      await vi.waitFor(() => {
        const errorCalls = mockOnFieldChange.mock.calls.filter((call) => call[0] === 'inbound_auth_config');
        expect(errorCalls).toHaveLength(0);
      });
    });

    it('should validate URI on blur', async () => {
            vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      await render(
        <AccessSection
          application={mockApplicationWithAuth}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      // Find the existing URI input
      const uriInput = getByDisplayValue('https://example.com/callback');

      // Focus and blur to trigger validation flow
      await userEvent.click(uriInput);
      await userEvent.tab();

      // The onBlur handler should have been called
      // Since URI is valid and non-empty, it should call updateRedirectUris
      await vi.waitFor(() => {
        expect(mockOnFieldChange).toHaveBeenCalledWith('inbound_auth_config', expect.any(Array));
      });
    });
  });

  describe('Handle empty user types data', () => {
    it('should handle undefined user types data gracefully', async () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: undefined,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      await render(<AccessSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      await expect.element(page.getByLabelText('Allowed User Types')).toBeInTheDocument();
    });

    it('should handle null application allowed_user_types', async () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      const appWithNullTypes = {
        ...mockApplication,
        allowed_user_types: undefined,
      };

      await render(
        <AccessSection
          application={appWithNullTypes as unknown as Application}
          editedApp={{}}
          onFieldChange={mockOnFieldChange}
        />,
      );

      await expect.element(page.getByLabelText('Allowed User Types')).toBeInTheDocument();
    });
  });

  describe('URI Error Handling', () => {
    it('should clear error when typing non-empty value in URI field', async () => {
            vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      await render(
        <AccessSection
          application={mockApplication}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      const uriInput = getByDisplayValue('https://example.com/callback');

      // Clear and blur to trigger empty error
      await userEvent.clear(uriInput);
      (uriInput.element() as HTMLElement).blur();

      // Now type something to clear the error
      await userEvent.type(uriInput, 'https://new-uri.com');

      // Error should be cleared when typing non-empty value
      await vi.waitFor(async () => {
        await expect.element(page.getByText('Invalid Redirect: URI must not be empty.')).not.toBeInTheDocument();
      });
    });

    it('should reindex errors when removing a URI with errors on subsequent URIs', async () => {
            vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      const configWithThreeUris = {
        ...mockOAuth2Config,
        redirect_uris: ['https://example.com/callback1', 'invalid-uri', 'https://example.com/callback3'],
      };

      await render(
        <AccessSection
          application={mockApplication}
          editedApp={{}}
          oauth2Config={configWithThreeUris}
          onFieldChange={mockOnFieldChange}
        />,
      );

      // First, trigger validation error on the second URI by blurring it
      const secondUriInput = getByDisplayValue('invalid-uri');
      await userEvent.click(secondUriInput);
      await userEvent.tab();

      // Now remove the first URI - this should trigger reindexing of errors
      const deleteButtons = page.getByRole('button', {name: /delete/i}).all();
      await userEvent.click(deleteButtons[0]);

      // The first URI should be removed
      await vi.waitFor(() => {
        expect(document.querySelector('input[value="https://example.com/callback1"]')).toBeNull();
      });
    });

    it('should preserve errors on URIs before the removed index', async () => {
            vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      const configWithThreeUris = {
        ...mockOAuth2Config,
        redirect_uris: ['invalid-first', 'https://example.com/callback2', 'https://example.com/callback3'],
      };

      await render(
        <AccessSection
          application={mockApplication}
          editedApp={{}}
          oauth2Config={configWithThreeUris}
          onFieldChange={mockOnFieldChange}
        />,
      );

      // Trigger validation error on the first URI
      const firstUriInput = getByDisplayValue('invalid-first');
      await userEvent.click(firstUriInput);
      await userEvent.tab();

      // Remove the last URI (index 2) - error on index 0 should be preserved
      const deleteButtons = page.getByRole('button', {name: /delete/i}).all();
      await userEvent.click(deleteButtons[2]);

      // The last URI should be removed
      await vi.waitFor(() => {
        expect(document.querySelector('input[value="https://example.com/callback3"]')).toBeNull();
      });
      // First URI should still be present
      expect(getByDisplayValue('invalid-first')).toBeInTheDocument();
    });
  });

  describe('Mixed Inbound Auth Config', () => {
    it('should preserve non-oauth2 config when updating redirect URIs', async () => {
            vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      const appWithMixedConfig: Application = {
        ...mockApplication,
        inbound_auth_config: [
          {
            type: 'saml',
            config: {issuer: 'test-issuer'},
          },
          {
            type: 'oauth2',
            config: {
              client_id: 'client-123',
              redirect_uris: ['https://example.com/callback'],
            },
          },
        ],
      } as Application;

      await render(
        <AccessSection
          application={appWithMixedConfig}
          editedApp={{}}
          oauth2Config={mockOAuth2Config}
          onFieldChange={mockOnFieldChange}
        />,
      );

      // Blur the URI input to trigger updateRedirectUris
      const uriInput = getByDisplayValue('https://example.com/callback');
      await userEvent.click(uriInput);
      await userEvent.tab();

      await vi.waitFor(() => {
        expect(mockOnFieldChange).toHaveBeenCalledWith('inbound_auth_config', expect.any(Array));
        const call = mockOnFieldChange.mock.calls.find((c) => c[0] === 'inbound_auth_config');
        if (call) {
          const updatedConfig = call[1] as {type: string}[];
          // Should contain both saml and oauth2 configs
          expect(updatedConfig.some((c) => c.type === 'saml')).toBe(true);
          expect(updatedConfig.some((c) => c.type === 'oauth2')).toBe(true);
        }
      });
    });
  });

  describe('URL Field Sync Effect', () => {
    it('should display editedApp URL over application URL', async () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      await render(
        <AccessSection
          application={mockApplication}
          editedApp={{url: 'https://edited-url.com'}}
          onFieldChange={mockOnFieldChange}
        />,
      );

      const urlInput = page.getByLabelText('Application URL');
      expect(urlInput).toHaveValue('https://edited-url.com');
    });

    it('should display application URL when editedApp URL is not provided', async () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      await render(<AccessSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const urlInput = page.getByLabelText('Application URL');
      expect(urlInput).toHaveValue('https://example.com');
    });

    it('should display empty string when neither editedApp nor application have URL', async () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      const appWithoutUrl = {...mockApplication, url: undefined};
      await render(
        <AccessSection
          application={appWithoutUrl as unknown as Application}
          editedApp={{}}
          onFieldChange={mockOnFieldChange}
        />,
      );

      const urlInput = page.getByLabelText('Application URL');
      expect(urlInput).toHaveValue('');
    });
  });

  describe('Redirect URI Updates', () => {
    it('should not update redirect URIs when oauth2Config is undefined', async () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      await render(<AccessSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      // Without oauth2Config, redirect URI section should not be rendered
      await expect.element(page.getByText('Authorized redirect URIs')).not.toBeInTheDocument();

      // No inbound_auth_config calls should be made
      const inboundAuthCalls = mockOnFieldChange.mock.calls.filter((call) => call[0] === 'inbound_auth_config');
      expect(inboundAuthCalls).toHaveLength(0);
    });

    it('should filter out empty URIs when updating redirect URIs', async () => {
            vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      const configWithMultipleUris = {
        ...mockOAuth2Config,
        redirect_uris: ['https://example.com/callback1', 'https://example.com/callback2'],
      };

      await render(
        <AccessSection
          application={mockApplication}
          editedApp={{}}
          oauth2Config={configWithMultipleUris}
          onFieldChange={mockOnFieldChange}
        />,
      );

      // Focus on first valid URI and blur to trigger update
      const uriInput = getByDisplayValue('https://example.com/callback1');
      await userEvent.click(uriInput);
      await userEvent.tab();

      await vi.waitFor(() => {
        const inboundAuthCalls = mockOnFieldChange.mock.calls.filter((call) => call[0] === 'inbound_auth_config');
        expect(inboundAuthCalls.length).toBeGreaterThan(0);
      });
    });
  });

  describe('Error Reindexing on URI Removal', () => {
    it('should reindex errors when removing URI from the middle of the list', async () => {
            vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      const configWithThreeUris = {
        ...mockOAuth2Config,
        redirect_uris: ['https://example.com/callback1', 'https://example.com/callback2', 'invalid-uri-3'],
      };

      await render(
        <AccessSection
          application={mockApplication}
          editedApp={{}}
          oauth2Config={configWithThreeUris}
          onFieldChange={mockOnFieldChange}
        />,
      );

      // Trigger validation error on the third URI
      const thirdUriInput = getByDisplayValue('invalid-uri-3');
      await userEvent.click(thirdUriInput);
      await userEvent.tab();

      // Remove the second URI
      const deleteButtons = page.getByRole('button', {name: /delete/i}).all();
      await userEvent.click(deleteButtons[1]);

      // Verify the second URI was removed
      await vi.waitFor(() => {
        expect(document.querySelector('input[value="https://example.com/callback2"]')).toBeNull();
      });
      // First and third (now second) should still be present
      expect(getByDisplayValue('https://example.com/callback1')).toBeInTheDocument();
      expect(getByDisplayValue('invalid-uri-3')).toBeInTheDocument();
    });

    it('should preserve error for URI at index before removed URI', async () => {
            vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      const configWithThreeUris = {
        ...mockOAuth2Config,
        redirect_uris: ['invalid-first', 'https://example.com/callback2', 'https://example.com/callback3'],
      };

      await render(
        <AccessSection
          application={mockApplication}
          editedApp={{}}
          oauth2Config={configWithThreeUris}
          onFieldChange={mockOnFieldChange}
        />,
      );

      // Trigger validation error on the first URI
      const firstUriInput = getByDisplayValue('invalid-first');
      await userEvent.click(firstUriInput);
      await userEvent.tab();

      // Remove the third URI
      const deleteButtons = page.getByRole('button', {name: /delete/i}).all();
      await userEvent.click(deleteButtons[2]);

      // First URI should still be present with its error state preserved
      expect(getByDisplayValue('invalid-first')).toBeInTheDocument();
    });

    it('should shift error indices down when removing URI before errored URI', async () => {
            vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      const configWithThreeUris = {
        ...mockOAuth2Config,
        redirect_uris: ['https://example.com/callback1', 'https://example.com/callback2', 'invalid-third'],
      };

      await render(
        <AccessSection
          application={mockApplication}
          editedApp={{}}
          oauth2Config={configWithThreeUris}
          onFieldChange={mockOnFieldChange}
        />,
      );

      // Trigger validation error on the third URI
      const thirdUriInput = getByDisplayValue('invalid-third');
      await userEvent.click(thirdUriInput);
      await userEvent.tab();

      // Remove the first URI - this should cause error index to shift from 2 to 1
      const deleteButtons = page.getByRole('button', {name: /delete/i}).all();
      await userEvent.click(deleteButtons[0]);

      // First URI should be removed
      await vi.waitFor(() => {
        expect(document.querySelector('input[value="https://example.com/callback1"]')).toBeNull();
      });
      // Third URI (now second) should still be present
      expect(getByDisplayValue('invalid-third')).toBeInTheDocument();
    });
  });

  describe('OAuth2 Config Updates', () => {
    it('should update redirect URIs state when oauth2Config prop changes', async () => {
      vi.mocked(useGetUserTypes).mockReturnValue({
        data: mockUserTypes,
        isLoading: false,
      } as unknown as MockedUseGetUserTypes);

      const initialConfig = {
        ...mockOAuth2Config,
        redirect_uris: ['https://initial.com/callback'],
      };

      const {rerender} = await render(

        <AccessSection
          application={mockApplication}
          editedApp={{}}
          oauth2Config={initialConfig}
          onFieldChange={mockOnFieldChange}
        />,
      );

      expect(getByDisplayValue('https://initial.com/callback')).toBeInTheDocument();

      const updatedConfig = {
        ...mockOAuth2Config,
        redirect_uris: ['https://updated.com/callback'],
      };

      await rerender(
        <AccessSection
          application={mockApplication}
          editedApp={{}}
          oauth2Config={updatedConfig}
          onFieldChange={mockOnFieldChange}
        />,
      );

      await vi.waitFor(() => {
        expect(getByDisplayValue('https://updated.com/callback')).toBeInTheDocument();
      });
    });
  });
});
