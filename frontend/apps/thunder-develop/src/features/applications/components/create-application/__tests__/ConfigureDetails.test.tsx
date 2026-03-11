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

import {beforeEach, describe, expect, it, vi} from 'vitest';
import {page, userEvent} from 'vitest/browser';
import {render} from '@thunder/test-utils/browser';
import {LoggerProvider, LogLevel} from '@thunder/logger';
import ConfigureDetails from '../ConfigureDetails';
import type {ApplicationTemplate} from '../../../models/application-templates';
import ApplicationCreateContext, {
  type ApplicationCreateContextType,
} from '../../../contexts/ApplicationCreate/ApplicationCreateContext';
import {TechnologyApplicationTemplate, PlatformApplicationTemplate} from '../../../models/application-templates';
import {getDefaultOAuthConfig, TokenEndpointAuthMethods} from '../../../models/oauth';
import {AuthenticatorTypes} from '../../../../integrations/models/authenticators';

const createTemplate = (name: string, redirectUris?: string[]): ApplicationTemplate => ({
  name,
  description: `${name} description`,
  inbound_auth_config: [
    {
      type: 'oauth2',
      config: {
        ...getDefaultOAuthConfig(),
        redirect_uris: redirectUris,
        token_endpoint_auth_method: TokenEndpointAuthMethods.CLIENT_SECRET_BASIC,
      },
    },
  ],
});

const renderWithContext = (
  props: Parameters<typeof ConfigureDetails>[0],
  contextOverrides: Partial<ApplicationCreateContextType> = {},
) => {
  const baseContext: ApplicationCreateContextType = {
    currentStep: null as unknown as ApplicationCreateContextType['currentStep'],
    setCurrentStep: vi.fn(),
    appName: 'Test App',
    setAppName: vi.fn(),
    themeId: null,
    setThemeId: vi.fn(),
    selectedTheme: null,
    setSelectedTheme: vi.fn(),
    appLogo: null,
    setAppLogo: vi.fn(),
    selectedColor: '',
    setSelectedColor: vi.fn(),
    integrations: {},
    setIntegrations: vi.fn(),
    toggleIntegration: vi.fn(),
    selectedAuthFlow: null,
    setSelectedAuthFlow: vi.fn(),
    signInApproach: null as unknown as ApplicationCreateContextType['signInApproach'],
    setSignInApproach: vi.fn(),
    selectedTechnology: null,
    setSelectedTechnology: vi.fn(),
    selectedPlatform: null,
    setSelectedPlatform: vi.fn(),
    selectedTemplateConfig: null,
    setSelectedTemplateConfig: vi.fn(),
    hostingUrl: '',
    setHostingUrl: vi.fn(),
    callbackUrlFromConfig: '',
    setCallbackUrlFromConfig: vi.fn(),
    hasCompletedOnboarding: false,
    setHasCompletedOnboarding: vi.fn(),
    error: null,
    setError: vi.fn(),
    reset: vi.fn(),
    relyingPartyId: '',
    setRelyingPartyId: vi.fn(),
    relyingPartyName: '',
    setRelyingPartyName: vi.fn(),
    ...contextOverrides,
  };

  return render(
    <LoggerProvider
      logger={{
        level: LogLevel.ERROR,
        transports: [],
      }}
    >
      <ApplicationCreateContext.Provider value={baseContext}>
        <ConfigureDetails {...props} />
      </ApplicationCreateContext.Provider>
    </LoggerProvider>,
  );
};

describe('ConfigureDetails', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders the no-configuration message when redirect URIs are already populated', async () => {
    const template = createTemplate('Browser App', ['https://example.com/callback']);

    await renderWithContext(
      {
        technology: TechnologyApplicationTemplate.REACT,
        platform: PlatformApplicationTemplate.BROWSER,
        onHostingUrlChange: vi.fn(),
        onCallbackUrlChange: vi.fn(),
        onReadyChange: vi.fn(),
      },
      {selectedTemplateConfig: template},
    );

    expect(
      page.getByText('Your application is ready to go! You can proceed to the next step.'),
    ).toBeInTheDocument();
  });

  it('renders passkey configuration even when no other configuration is required', async () => {
    const template = createTemplate('Browser App', ['https://example.com/callback']);

    await renderWithContext(
      {
        technology: TechnologyApplicationTemplate.REACT,
        platform: PlatformApplicationTemplate.BROWSER,
        onHostingUrlChange: vi.fn(),
        onCallbackUrlChange: vi.fn(),
        onReadyChange: vi.fn(),
      },
      {
        selectedTemplateConfig: template,
        integrations: {[AuthenticatorTypes.PASSKEY]: true},
        selectedAuthFlow: null,
      },
    );

    await expect.element(page.getByText('No Additional Configuration Needed')).not.toBeInTheDocument();
    await expect.element(page.getByText('Passkey Settings')).toBeInTheDocument();
    expect(
      page.getByPlaceholder('e.g., example.com'),
    ).toBeInTheDocument();
  });

  it('shows URL configuration inputs and notifies callbacks when values change', async () => {
    const template = createTemplate('Browser App', []);
    const onHostingUrlChange = vi.fn();
    const onCallbackUrlChange = vi.fn();
    const onReadyChange = vi.fn();

    await renderWithContext(
      {
        technology: TechnologyApplicationTemplate.REACT,
        platform: PlatformApplicationTemplate.BROWSER,
        onHostingUrlChange,
        onCallbackUrlChange,
        onReadyChange,
      },
      {selectedTemplateConfig: template},
    );

    const hostingUrlInput = page.getByPlaceholder(
      'https://myapp.example.com',
    );

    await userEvent.fill(hostingUrlInput, 'https://example.com');

    await vi.waitFor(() => expect(onHostingUrlChange).toHaveBeenLastCalledWith('https://example.com'));

    const customRadio = page.getByRole('radio', {
      name: 'Custom URL',
    });
    await userEvent.click(customRadio);

    const callbackUrlInput = document.getElementById('callback-url-input') as HTMLInputElement;
    await userEvent.clear(callbackUrlInput);
    await userEvent.fill(callbackUrlInput, 'https://example.com/callback');

    await vi.waitFor(() => expect(onCallbackUrlChange).toHaveBeenLastCalledWith('https://example.com/callback'), {
      timeout: 10000,
    });
    expect(onReadyChange).toHaveBeenCalled();
  }, 15000);

  it('displays deep link configuration and forwards values for mobile templates', async () => {
    const template = createTemplate('Mobile App', []);
    const onCallbackUrlChange = vi.fn();
    const onHostingUrlChange = vi.fn();
    const onReadyChange = vi.fn();

    await renderWithContext(
      {
        technology: TechnologyApplicationTemplate.OTHER,
        platform: PlatformApplicationTemplate.MOBILE,
        onHostingUrlChange,
        onCallbackUrlChange,
        onReadyChange,
      },
      {selectedTemplateConfig: template},
    );

    await expect.element(page.getByText('Deep links (e.g., myapp://callback) or universal links (e.g., https://example.com/callback) are used to redirect users back to your mobile app after authentication.')).toBeInTheDocument();

    const deeplinkInput = page.getByPlaceholder('myapp://callback or https://example.com/callback');
    await userEvent.fill(deeplinkInput, 'myapp://callback');

    await vi.waitFor(() => expect(onCallbackUrlChange).toHaveBeenLastCalledWith('myapp://callback'));
    expect(onReadyChange).toHaveBeenCalled();
  });

  it('validates hosting URL input and shows validation errors', async () => {
    const template = createTemplate('Browser App', []);
    const onHostingUrlChange = vi.fn();
    const onCallbackUrlChange = vi.fn();
    const onReadyChange = vi.fn();

    await renderWithContext(
      {
        technology: TechnologyApplicationTemplate.REACT,
        platform: PlatformApplicationTemplate.BROWSER,
        onHostingUrlChange,
        onCallbackUrlChange,
        onReadyChange,
      },
      {selectedTemplateConfig: template},
    );

    const hostingUrlInput = page.getByPlaceholder(
      'https://myapp.example.com',
    );

    // Type invalid URL
    await userEvent.type(hostingUrlInput, 'not-a-url');
    await userEvent.tab(); // Trigger validation

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Please enter a valid URL')).toBeInTheDocument();
    });

    // Clear and type valid URL
    await userEvent.clear(hostingUrlInput);
    await userEvent.type(hostingUrlInput, 'https://example.com');

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Please enter a valid URL')).not.toBeInTheDocument();
      expect(onHostingUrlChange).toHaveBeenLastCalledWith('https://example.com');
    });
  });

  it('validates callback URL when in custom mode', async () => {
    const template = createTemplate('Browser App', []);
    const onHostingUrlChange = vi.fn();
    const onCallbackUrlChange = vi.fn();
    const onReadyChange = vi.fn();

    await renderWithContext(
      {
        technology: TechnologyApplicationTemplate.REACT,
        platform: PlatformApplicationTemplate.BROWSER,
        onHostingUrlChange,
        onCallbackUrlChange,
        onReadyChange,
      },
      {selectedTemplateConfig: template},
    );

    // Switch to custom callback mode
    const customRadio = page.getByRole('radio', {
      name: 'Custom URL',
    });
    await userEvent.click(customRadio);

    const callbackUrlInput = document.getElementById('callback-url-input') as HTMLInputElement;

    // Type invalid URL
    await userEvent.type(callbackUrlInput, 'invalid-url');
    await userEvent.tab(); // Trigger validation

    await vi.waitFor(async () => {
      await expect.element(page.getByText('Please enter a valid URL')).toBeInTheDocument();
    });
  });

  it('validates deep link input for mobile apps', async () => {
    const template = createTemplate('Mobile App', []);
    const onCallbackUrlChange = vi.fn();
    const onHostingUrlChange = vi.fn();
    const onReadyChange = vi.fn();

    await renderWithContext(
      {
        technology: TechnologyApplicationTemplate.OTHER,
        platform: PlatformApplicationTemplate.MOBILE,
        onHostingUrlChange,
        onCallbackUrlChange,
        onReadyChange,
      },
      {selectedTemplateConfig: template},
    );

    const deeplinkInput = page.getByPlaceholder('myapp://callback or https://example.com/callback');

    // Type invalid deep link
    await userEvent.type(deeplinkInput, 'invalid-deeplink');
    await userEvent.tab(); // Trigger validation

    await vi.waitFor(async () => {
      await expect.element(page.getByText(/Please enter a valid deep link/)).toBeInTheDocument();
    });
  });

  it('handles same as hosting URL callback mode correctly', async () => {
    const template = createTemplate('Browser App', []);
    const onHostingUrlChange = vi.fn();
    const onCallbackUrlChange = vi.fn();
    const onReadyChange = vi.fn();

    await renderWithContext(
      {
        technology: TechnologyApplicationTemplate.REACT,
        platform: PlatformApplicationTemplate.BROWSER,
        onHostingUrlChange,
        onCallbackUrlChange,
        onReadyChange,
      },
      {selectedTemplateConfig: template},
    );

    const hostingUrlInput = page.getByPlaceholder(
      'https://myapp.example.com',
    );

    // Type hosting URL
    await userEvent.fill(hostingUrlInput, 'https://example.com');

    // By default, "Same as hosting" should be selected, so callback URL should sync
    await vi.waitFor(() => {
      expect(onCallbackUrlChange).toHaveBeenLastCalledWith('https://example.com');
    });
  });

  it('renders user type selection when multiple user types are available', async () => {
    // Create template with empty allowed_user_types array to trigger user type selection
    const template = {
      ...createTemplate('Browser App', []),
      allowed_user_types: [], // Empty array means user types selection is required
    };
    const userTypes = [
      {id: 'user-type-1', name: 'Customer', ouId: 'ou-1', allowSelfRegistration: true},
      {id: 'user-type-2', name: 'Employee', ouId: 'ou-2', allowSelfRegistration: false},
    ];

    await renderWithContext(
      {
        technology: TechnologyApplicationTemplate.REACT,
        platform: PlatformApplicationTemplate.BROWSER,
        onHostingUrlChange: vi.fn(),
        onCallbackUrlChange: vi.fn(),
        onReadyChange: vi.fn(),
        userTypes,
        selectedUserTypes: [],
        onUserTypesChange: vi.fn(),
      },
      {selectedTemplateConfig: template},
    );

    await expect.element(page.getByText('onboarding.configure.details.userTypes.label')).toBeInTheDocument();
  });

  it('calls onUserTypesChange when user type selection changes', async () => {
    // Create template with empty allowed_user_types array to trigger user type selection
    const template = {
      ...createTemplate('Browser App', []),
      allowed_user_types: [], // Empty array means user types selection is required
    };
    const userTypes = [
      {id: 'user-type-1', name: 'Customer', ouId: 'ou-1', allowSelfRegistration: true},
      {id: 'user-type-2', name: 'Employee', ouId: 'ou-2', allowSelfRegistration: false},
    ];
    const onUserTypesChange = vi.fn();

    await renderWithContext(
      {
        technology: TechnologyApplicationTemplate.REACT,
        platform: PlatformApplicationTemplate.BROWSER,
        onHostingUrlChange: vi.fn(),
        onCallbackUrlChange: vi.fn(),
        onReadyChange: vi.fn(),
        userTypes,
        selectedUserTypes: [],
        onUserTypesChange,
      },
      {selectedTemplateConfig: template},
    );

    const autocomplete = page.getByRole('combobox');
    await userEvent.click(autocomplete);

    const customerOption = page.getByText('Customer');
    await userEvent.click(customerOption);

    expect(onUserTypesChange).toHaveBeenCalledWith(['Customer']);
  });

  it('does not render user type selection when no user types are provided', async () => {
    const template = createTemplate('Browser App', []);

    await renderWithContext(
      {
        technology: TechnologyApplicationTemplate.REACT,
        platform: PlatformApplicationTemplate.BROWSER,
        onHostingUrlChange: vi.fn(),
        onCallbackUrlChange: vi.fn(),
        onReadyChange: vi.fn(),
        userTypes: [],
        selectedUserTypes: [],
      },
      {selectedTemplateConfig: template},
    );

    await expect.element(page.getByText('onboarding.configure.details.userTypes.label')).not.toBeInTheDocument();
  });

  it('notifies readiness based on form validity', async () => {
    const template = createTemplate('Browser App', []);
    const onReadyChange = vi.fn();

    await renderWithContext(
      {
        technology: TechnologyApplicationTemplate.REACT,
        platform: PlatformApplicationTemplate.BROWSER,
        onHostingUrlChange: vi.fn(),
        onCallbackUrlChange: vi.fn(),
        onReadyChange,
      },
      {selectedTemplateConfig: template},
    );

    // Initially should not be ready (no URLs entered)
    await vi.waitFor(() => {
      expect(onReadyChange).toHaveBeenCalledWith(false);
    });

    const hostingUrlInput = page.getByPlaceholder(
      'https://myapp.example.com',
    );

    // Enter valid URL - should become ready
    await userEvent.fill(hostingUrlInput, 'https://example.com');

    await vi.waitFor(() => {
      expect(onReadyChange).toHaveBeenCalledWith(true);
    });
  });

  it('handles server applications configuration correctly', async () => {
    const template = createTemplate('Server Application', []);

    await renderWithContext(
      {
        technology: TechnologyApplicationTemplate.NEXTJS,
        platform: PlatformApplicationTemplate.SERVER,
        onHostingUrlChange: vi.fn(),
        onCallbackUrlChange: vi.fn(),
        onReadyChange: vi.fn(),
      },
      {selectedTemplateConfig: template},
    );

    await expect.element(page.getByText('Configuration')).toBeInTheDocument();
    expect(
      page.getByPlaceholder('https://myapp.example.com'),
    ).toBeInTheDocument();
  });

  it('allows updating relying party ID for passkey configuration', async () => {
    const template = createTemplate('Browser App', ['https://example.com/callback']);
    const setRelyingPartyId = vi.fn();

    await renderWithContext(
      {
        technology: TechnologyApplicationTemplate.REACT,
        platform: PlatformApplicationTemplate.BROWSER,
        onHostingUrlChange: vi.fn(),
        onCallbackUrlChange: vi.fn(),
        onReadyChange: vi.fn(),
      },
      {
        selectedTemplateConfig: template,
        integrations: {[AuthenticatorTypes.PASSKEY]: true},
        relyingPartyId: 'localhost',
        setRelyingPartyId,
      },
    );

    const relyingPartyIdInput = page.getByPlaceholder(
      'e.g., example.com',
    );

    await userEvent.clear(relyingPartyIdInput);
    await userEvent.fill(relyingPartyIdInput, 'example.com');

    expect(setRelyingPartyId).toHaveBeenCalled();
  });

  it('allows updating relying party name for passkey configuration', async () => {
    const template = createTemplate('Browser App', ['https://example.com/callback']);
    const setRelyingPartyName = vi.fn();

    await renderWithContext(
      {
        technology: TechnologyApplicationTemplate.REACT,
        platform: PlatformApplicationTemplate.BROWSER,
        onHostingUrlChange: vi.fn(),
        onCallbackUrlChange: vi.fn(),
        onReadyChange: vi.fn(),
      },
      {
        selectedTemplateConfig: template,
        integrations: {[AuthenticatorTypes.PASSKEY]: true},
        relyingPartyName: 'Test App',
        setRelyingPartyName,
      },
    );

    const relyingPartyNameInput = page.getByPlaceholder(
      'e.g., My Application',
    );

    await userEvent.clear(relyingPartyNameInput);
    await userEvent.fill(relyingPartyNameInput, 'My Application');

    expect(setRelyingPartyName).toHaveBeenCalled();
  });

  it('renders both passkey and URL configuration when passkey is enabled', async () => {
    const template = createTemplate('Browser App', []);

    await renderWithContext(
      {
        technology: TechnologyApplicationTemplate.REACT,
        platform: PlatformApplicationTemplate.BROWSER,
        onHostingUrlChange: vi.fn(),
        onCallbackUrlChange: vi.fn(),
        onReadyChange: vi.fn(),
      },
      {
        selectedTemplateConfig: template,
        integrations: {[AuthenticatorTypes.PASSKEY]: true},
      },
    );

    // Should show passkey configuration
    await expect.element(page.getByText('Passkey Settings')).toBeInTheDocument();
    expect(
      page.getByPlaceholder('e.g., example.com'),
    ).toBeInTheDocument();

    // Should also show URL configuration
    expect(
      page.getByPlaceholder('https://myapp.example.com'),
    ).toBeInTheDocument();
  });

  it('does not render passkey configuration when BASIC_AUTH is the only authenticator', async () => {
    const template = createTemplate('Browser App', []);

    await renderWithContext(
      {
        technology: TechnologyApplicationTemplate.REACT,
        platform: PlatformApplicationTemplate.BROWSER,
        onHostingUrlChange: vi.fn(),
        onCallbackUrlChange: vi.fn(),
        onReadyChange: vi.fn(),
      },
      {
        selectedTemplateConfig: template,
        integrations: {[AuthenticatorTypes.BASIC_AUTH]: true},
      },
    );

    // Should not show passkey section
    await expect.element(page.getByText('Passkey Settings')).not.toBeInTheDocument();
    expect(
      page.getByPlaceholder('e.g., example.com'),
    ).not.toBeInTheDocument();
  });

  it('initializes passkey relying party defaults from hostname and app name', async () => {
    const template = createTemplate('Browser App', ['https://example.com/callback']);

    await renderWithContext(
      {
        technology: TechnologyApplicationTemplate.REACT,
        platform: PlatformApplicationTemplate.BROWSER,
        onHostingUrlChange: vi.fn(),
        onCallbackUrlChange: vi.fn(),
        onReadyChange: vi.fn(),
      },
      {
        selectedTemplateConfig: template,
        integrations: {[AuthenticatorTypes.PASSKEY]: true},
        selectedAuthFlow: null,
        relyingPartyId: '',
        relyingPartyName: '',
      },
    );

    const relyingPartyIdInput = page.getByPlaceholder(
      'e.g., example.com',
    );
    const relyingPartyNameInput = page.getByPlaceholder('e.g., My Application');

    await vi.waitFor(async () => {
      await expect.element(relyingPartyIdInput).toHaveValue(window.location.hostname);
      await expect.element(relyingPartyNameInput).toHaveValue('Test App');
    });
  });

  it('falls back to default passkey labels and placeholders when translations are empty', async () => {
    const template = createTemplate('Browser App', ['https://example.com/callback']);

    await renderWithContext(
      {
        technology: TechnologyApplicationTemplate.REACT,
        platform: PlatformApplicationTemplate.BROWSER,
        onHostingUrlChange: vi.fn(),
        onCallbackUrlChange: vi.fn(),
        onReadyChange: vi.fn(),
      },
      {
        selectedTemplateConfig: template,
        integrations: {[AuthenticatorTypes.PASSKEY]: true},
        selectedAuthFlow: null,
      },
    );

    await expect.element(page.getByText('Passkey Settings')).toBeInTheDocument();
    await expect.element(page.getByText('Relying Party ID')).toBeInTheDocument();
    await expect.element(page.getByPlaceholder('e.g., example.com')).toBeInTheDocument();
    await expect.element(page.getByText('Relying Party Name')).toBeInTheDocument();
    await expect.element(page.getByPlaceholder('e.g., My App')).toBeInTheDocument();
  });
});
