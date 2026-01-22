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
import {render, screen, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {LoggerProvider, LogLevel} from '@thunder/logger';
import ConfigureDetails from '../ConfigureDetails';
import type {ApplicationTemplate} from '../../../models/application-templates';
import ApplicationCreateContext, {
  type ApplicationCreateContextType,
} from '../../../contexts/ApplicationCreate/ApplicationCreateContext';
import {TechnologyApplicationTemplate, PlatformApplicationTemplate} from '../../../models/application-templates';
import {getDefaultOAuthConfig, TokenEndpointAuthMethods} from '../../../models/oauth';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

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
    selectedColor: '#ffffff',
    setSelectedColor: vi.fn(),
    appLogo: null,
    setAppLogo: vi.fn(),
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

  it('renders the no-configuration message when redirect URIs are already populated', () => {
    const template = createTemplate('Browser App', ['https://example.com/callback']);

    renderWithContext(
      {
        technology: TechnologyApplicationTemplate.REACT,
        platform: PlatformApplicationTemplate.BROWSER,
        onHostingUrlChange: vi.fn(),
        onCallbackUrlChange: vi.fn(),
        onReadyChange: vi.fn(),
      },
      {selectedTemplateConfig: template},
    );

    expect(screen.getByText('applications:onboarding.configure.details.noConfigRequired.title')).toBeInTheDocument();
    expect(
      screen.getByText('applications:onboarding.configure.details.noConfigRequired.description'),
    ).toBeInTheDocument();
  });

  it('shows URL configuration inputs and notifies callbacks when values change', async () => {
    const template = createTemplate('Browser App', []);
    const onHostingUrlChange = vi.fn();
    const onCallbackUrlChange = vi.fn();
    const onReadyChange = vi.fn();

    renderWithContext(
      {
        technology: TechnologyApplicationTemplate.REACT,
        platform: PlatformApplicationTemplate.BROWSER,
        onHostingUrlChange,
        onCallbackUrlChange,
        onReadyChange,
      },
      {selectedTemplateConfig: template},
    );

    const hostingUrlInput = screen.getByPlaceholderText(
      'applications:onboarding.configure.details.hostingUrl.placeholder',
    );
    const user = userEvent.setup();

    await user.type(hostingUrlInput, 'https://example.com');

    expect(onHostingUrlChange).toHaveBeenLastCalledWith('https://example.com');

    const customRadio = screen.getByRole('radio', {
      name: 'applications:onboarding.configure.details.callbackMode.custom',
    });
    await user.click(customRadio);

    const callbackUrlInput = document.getElementById('callback-url-input') as HTMLInputElement;
    await user.clear(callbackUrlInput);
    await user.type(callbackUrlInput, 'https://example.com/callback');

    await waitFor(() => expect(onCallbackUrlChange).toHaveBeenLastCalledWith('https://example.com/callback'));
    expect(onReadyChange).toHaveBeenCalled();
  });

  it('displays deep link configuration and forwards values for mobile templates', async () => {
    const template = createTemplate('Mobile App', []);
    const onCallbackUrlChange = vi.fn();
    const onHostingUrlChange = vi.fn();
    const onReadyChange = vi.fn();

    renderWithContext(
      {
        technology: TechnologyApplicationTemplate.OTHER,
        platform: PlatformApplicationTemplate.MOBILE,
        onHostingUrlChange,
        onCallbackUrlChange,
        onReadyChange,
      },
      {selectedTemplateConfig: template},
    );

    expect(screen.getByText('applications:onboarding.configure.details.mobile.info')).toBeInTheDocument();

    const deeplinkInput = screen.getByPlaceholderText('applications:onboarding.configure.details.deeplink.placeholder');
    const user = userEvent.setup();
    await user.type(deeplinkInput, 'myapp://callback');

    await waitFor(() => expect(onCallbackUrlChange).toHaveBeenLastCalledWith('myapp://callback'));
    expect(onReadyChange).toHaveBeenCalled();
  });

  it('validates hosting URL input and shows validation errors', async () => {
    const template = createTemplate('Browser App', []);
    const onHostingUrlChange = vi.fn();
    const onCallbackUrlChange = vi.fn();
    const onReadyChange = vi.fn();

    renderWithContext(
      {
        technology: TechnologyApplicationTemplate.REACT,
        platform: PlatformApplicationTemplate.BROWSER,
        onHostingUrlChange,
        onCallbackUrlChange,
        onReadyChange,
      },
      {selectedTemplateConfig: template},
    );

    const hostingUrlInput = screen.getByPlaceholderText(
      'applications:onboarding.configure.details.hostingUrl.placeholder',
    );
    const user = userEvent.setup();

    // Type invalid URL
    await user.type(hostingUrlInput, 'not-a-url');
    await user.tab(); // Trigger validation

    await waitFor(() => {
      expect(screen.getByText('Please enter a valid URL')).toBeInTheDocument();
    });

    // Clear and type valid URL
    await user.clear(hostingUrlInput);
    await user.type(hostingUrlInput, 'https://example.com');

    await waitFor(() => {
      expect(screen.queryByText('Please enter a valid URL')).not.toBeInTheDocument();
      expect(onHostingUrlChange).toHaveBeenLastCalledWith('https://example.com');
    });
  });

  it('validates callback URL when in custom mode', async () => {
    const template = createTemplate('Browser App', []);
    const onHostingUrlChange = vi.fn();
    const onCallbackUrlChange = vi.fn();
    const onReadyChange = vi.fn();

    renderWithContext(
      {
        technology: TechnologyApplicationTemplate.REACT,
        platform: PlatformApplicationTemplate.BROWSER,
        onHostingUrlChange,
        onCallbackUrlChange,
        onReadyChange,
      },
      {selectedTemplateConfig: template},
    );

    const user = userEvent.setup();

    // Switch to custom callback mode
    const customRadio = screen.getByRole('radio', {
      name: 'applications:onboarding.configure.details.callbackMode.custom',
    });
    await user.click(customRadio);

    const callbackUrlInput = document.getElementById('callback-url-input') as HTMLInputElement;

    // Type invalid URL
    await user.type(callbackUrlInput, 'invalid-url');
    await user.tab(); // Trigger validation

    await waitFor(() => {
      expect(screen.getByText('Please enter a valid URL')).toBeInTheDocument();
    });
  });

  it('validates deep link input for mobile apps', async () => {
    const template = createTemplate('Mobile App', []);
    const onCallbackUrlChange = vi.fn();
    const onHostingUrlChange = vi.fn();
    const onReadyChange = vi.fn();

    renderWithContext(
      {
        technology: TechnologyApplicationTemplate.OTHER,
        platform: PlatformApplicationTemplate.MOBILE,
        onHostingUrlChange,
        onCallbackUrlChange,
        onReadyChange,
      },
      {selectedTemplateConfig: template},
    );

    const deeplinkInput = screen.getByPlaceholderText('applications:onboarding.configure.details.deeplink.placeholder');
    const user = userEvent.setup();

    // Type invalid deep link
    await user.type(deeplinkInput, 'invalid-deeplink');
    await user.tab(); // Trigger validation

    await waitFor(() => {
      expect(screen.getByText(/Please enter a valid deep link/)).toBeInTheDocument();
    });
  });

  it('handles same as hosting URL callback mode correctly', async () => {
    const template = createTemplate('Browser App', []);
    const onHostingUrlChange = vi.fn();
    const onCallbackUrlChange = vi.fn();
    const onReadyChange = vi.fn();

    renderWithContext(
      {
        technology: TechnologyApplicationTemplate.REACT,
        platform: PlatformApplicationTemplate.BROWSER,
        onHostingUrlChange,
        onCallbackUrlChange,
        onReadyChange,
      },
      {selectedTemplateConfig: template},
    );

    const hostingUrlInput = screen.getByPlaceholderText(
      'applications:onboarding.configure.details.hostingUrl.placeholder',
    );
    const user = userEvent.setup();

    // Type hosting URL
    await user.type(hostingUrlInput, 'https://example.com');

    // By default, "Same as hosting" should be selected, so callback URL should sync
    await waitFor(() => {
      expect(onCallbackUrlChange).toHaveBeenLastCalledWith('https://example.com');
    });
  });

  it('renders user type selection when multiple user types are available', () => {
    // Create template with empty allowed_user_types array to trigger user type selection
    const template = {
      ...createTemplate('Browser App', []),
      allowed_user_types: [], // Empty array means user types selection is required
    };
    const userTypes = [
      {id: 'user-type-1', name: 'Customer', ouId: 'ou-1', allowSelfRegistration: true},
      {id: 'user-type-2', name: 'Employee', ouId: 'ou-2', allowSelfRegistration: false},
    ];

    renderWithContext(
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

    expect(screen.getByText('applications:onboarding.configure.details.userTypes.label')).toBeInTheDocument();
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

    renderWithContext(
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

    const autocomplete = screen.getByRole('combobox');
    const user = userEvent.setup();
    await user.click(autocomplete);

    const customerOption = await screen.findByText('Customer');
    await user.click(customerOption);

    expect(onUserTypesChange).toHaveBeenCalledWith(['Customer']);
  });

  it('does not render user type selection when no user types are provided', () => {
    const template = createTemplate('Browser App', []);

    renderWithContext(
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

    expect(screen.queryByText('applications:onboarding.configure.details.userTypes.label')).not.toBeInTheDocument();
  });

  it('notifies readiness based on form validity', async () => {
    const template = createTemplate('Browser App', []);
    const onReadyChange = vi.fn();

    renderWithContext(
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
    await waitFor(() => {
      expect(onReadyChange).toHaveBeenCalledWith(false);
    });

    const hostingUrlInput = screen.getByPlaceholderText(
      'applications:onboarding.configure.details.hostingUrl.placeholder',
    );
    const user = userEvent.setup();

    // Enter valid URL - should become ready
    await user.type(hostingUrlInput, 'https://example.com');

    await waitFor(() => {
      expect(onReadyChange).toHaveBeenCalledWith(true);
    });
  });

  it('handles server applications configuration correctly', () => {
    const template = createTemplate('Server Application', []);

    renderWithContext(
      {
        technology: TechnologyApplicationTemplate.NEXTJS,
        platform: PlatformApplicationTemplate.SERVER,
        onHostingUrlChange: vi.fn(),
        onCallbackUrlChange: vi.fn(),
        onReadyChange: vi.fn(),
      },
      {selectedTemplateConfig: template},
    );

    expect(screen.getByText('applications:onboarding.configure.details.title')).toBeInTheDocument();
    expect(
      screen.getByPlaceholderText('applications:onboarding.configure.details.hostingUrl.placeholder'),
    ).toBeInTheDocument();
  });
});
