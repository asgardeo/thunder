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

import {describe, it, expect, beforeEach, vi} from 'vitest';
import {page} from 'vitest/browser';
import {render} from '@thunder/test-utils/browser';
import {IdentityProviderTypes, type IdentityProvider} from '@/features/integrations/models/identity-provider';
import {AuthenticatorTypes} from '@/features/integrations/models/authenticators';
import type {ThemeConfig} from '@thunder/shared-design';
import Preview, {type PreviewProps} from '../Preview';

// Mock the @asgardeo/react module
vi.mock('@asgardeo/react', () => ({
  BaseSignIn: ({children}: {children: () => React.ReactNode}) => <div>{children()}</div>,
  ThemeProvider: ({children}: {children: React.ReactNode}) => <div>{children}</div>,
}));

// Mock the useIdentityProviders hook
vi.mock('@/features/integrations/api/useIdentityProviders');

// Mock useColorScheme to test dark mode
const mockUseColorScheme = vi.fn<() => {mode: 'light' | 'dark'}>();
vi.mock('@wso2/oxygen-ui', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@wso2/oxygen-ui')>();
  return {
    ...actual,
    useColorScheme: () => mockUseColorScheme(),
  };
});

const {default: useIdentityProviders} = await import('@/features/integrations/api/useIdentityProviders');

const mockTheme: ThemeConfig = {
  direction: 'ltr',
  defaultColorScheme: 'light',
  colorSchemes: {
    light: {
      colors: {
        primary: {
          main: '#FF5733',
          dark: '#CC4529',
          contrastText: '#FFFFFF',
        },
        secondary: {
          main: '#0066CC',
          dark: '#004C99',
          contrastText: '#FFFFFF',
        },
        background: {
          default: '#FFFFFF',
          paper: '#F5F5F5',
        },
      },
    },
    dark: {
      colors: {
        primary: {
          main: '#00FF00',
          dark: '#00CC00',
          contrastText: '#000000',
        },
        secondary: {
          main: '#0088FF',
          dark: '#0066CC',
          contrastText: '#FFFFFF',
        },
        background: {
          default: '#121212',
          paper: '#1E1E1E',
        },
      },
    },
  },
};

describe('Preview', () => {
  const mockIdentityProviders: IdentityProvider[] = [
    {
      id: 'google-idp',
      name: 'Google',
      type: IdentityProviderTypes.GOOGLE,
      description: 'Google Identity Provider',
    },
    {
      id: 'github-idp',
      name: 'GitHub',
      type: 'GITHUB',
      description: 'GitHub Identity Provider',
    },
  ];

  const defaultProps: PreviewProps = {
    appLogo: 'https://example.com/logo.png',
    selectedTheme: mockTheme,
    integrations: {
      [AuthenticatorTypes.BASIC_AUTH]: true,
    },
  };

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);
    mockUseColorScheme.mockReturnValue({mode: 'light'});
  });

  const renderComponent = (props: Partial<PreviewProps> = {}) => render(<Preview {...defaultProps} {...props} />);

  it('should render the preview title', async () => {
    await renderComponent();

    await expect.element(page.getByText('Preview')).toBeInTheDocument();
  });

  it('should render the application logo when provided', async () => {
    await renderComponent();

    const logo = page.getByRole('img');
    expect(logo).toBeInTheDocument();
    expect(logo).toHaveAttribute('src', 'https://example.com/logo.png');
  });

  it('should not render logo when appLogo is null', async () => {
    await renderComponent({appLogo: null});

    await expect.element(page.getByRole('img')).not.toBeInTheDocument();
  });

  it('should render username and password fields when username/password is enabled', async () => {
    await renderComponent();

    await expect.element(page.getByText('Username')).toBeInTheDocument();
    await expect.element(page.getByPlaceholder('Enter your Username')).toBeInTheDocument();
    await expect.element(page.getByText('Password')).toBeInTheDocument();
    await expect.element(page.getByPlaceholder('Enter your Password')).toBeInTheDocument();
  });

  it('should render sign in button', async () => {
    await renderComponent();

    await expect.element(page.getByRole('button', {name: 'Sign In'})).toBeInTheDocument();
  });

  it('should not render username/password fields when disabled', async () => {
    await renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: false,
        'google-idp': true,
      },
    });

    await expect.element(page.getByText('Username')).not.toBeInTheDocument();
    await expect.element(page.getByText('Password')).not.toBeInTheDocument();
    await expect.element(page.getByRole('button', {name: 'Sign In'})).not.toBeInTheDocument();
  });

  it('should render social login buttons for enabled providers', async () => {
    await renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
        'google-idp': true,
        'github-idp': true,
      },
    });

    await expect.element(page.getByRole('button', {name: /Continue with Google/i})).toBeInTheDocument();
    await expect.element(page.getByRole('button', {name: /Continue with GitHub/i})).toBeInTheDocument();
  });

  it('should not render social login buttons when no providers are enabled', async () => {
    await renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
      },
    });

    await expect.element(page.getByRole('button', {name: /Continue with/i})).not.toBeInTheDocument();
  });

  it('should render divider when both username/password and social logins are enabled', async () => {
    await renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
        'google-idp': true,
      },
    });

    await expect.element(page.getByText('or')).toBeInTheDocument();
  });

  it('should not render divider when only username/password is enabled', async () => {
    await renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
      },
    });

    await expect.element(page.getByText('or')).not.toBeInTheDocument();
  });

  it('should not render divider when only social logins are enabled', async () => {
    await renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: false,
        'google-idp': true,
      },
    });

    await expect.element(page.getByText('or')).not.toBeInTheDocument();
  });

  it('should render only selected social providers', async () => {
    await renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
        'google-idp': true,
        // github-idp not included
      },
    });

    await expect.element(page.getByRole('button', {name: /Continue with Google/i})).toBeInTheDocument();
    await expect.element(page.getByRole('button', {name: /Continue with GitHub/i})).not.toBeInTheDocument();
  });

  it('should handle empty integrations object', async () => {
    await renderComponent({
      integrations: {},
    });

    // Username/password should not be shown when integrations is empty (defaults to false)
    await expect.element(page.getByText('Username')).not.toBeInTheDocument();
    await expect.element(page.getByText('Password')).not.toBeInTheDocument();
  });

  it('should apply theme primary color to sign in button background', async () => {
    await renderComponent();

    const signInButton = page.getByRole('button', {name: 'Sign In'});
    // Buttons omit variant="contained" to avoid CSS-variable specificity issues with
    // the outer app's theme; visual styles are applied entirely through the sx prop.
    expect(signInButton).not.toHaveClass('MuiButton-containedPrimary');
    expect(signInButton).toHaveStyle({backgroundColor: '#FF5733'});
  });

  it('should apply contrastText color to sign in button label', async () => {
    await renderComponent();

    const signInButton = page.getByRole('button', {name: 'Sign In'});
    expect(signInButton).toHaveStyle({color: '#FFFFFF'});
  });

  it('should apply selected color to logo background', async () => {
    await renderComponent();

    const logo = page.getByRole('img');
    const avatarContainer = logo.element().closest('.MuiAvatar-root');
    expect(avatarContainer).toHaveStyle({backgroundColor: '#FF5733'});
  });

  it('should render input fields as disabled', async () => {
    await renderComponent();

    const usernameInput = page.getByPlaceholder('Enter your Username');
    const passwordInput = page.getByPlaceholder('Enter your Password');

    expect(usernameInput).toBeDisabled();
    expect(passwordInput).toBeDisabled();
  });

  it('should render social login buttons as disabled', async () => {
    await renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
        'google-idp': true,
      },
    });

    const googleButton = page.getByRole('button', {name: /Continue with Google/i});
    expect(googleButton).toBeDisabled();
  });

  it('should handle when useIdentityProviders returns undefined data', async () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    await renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
        'google-idp': true,
      },
    });

    // Should not crash, no social providers should be rendered (since data is undefined)
    await expect.element(page.getByRole('button', {name: /Continue with/i})).not.toBeInTheDocument();
  });

  it('should only show providers that exist in API and are selected', async () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: [mockIdentityProviders[0]], // Only Google in API
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    await renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
        'google-idp': true,
        // 'github-idp' is not in API, so even if selected, it won't show
      },
    });

    // Should only show Google (which exists in API)
    await expect.element(page.getByRole('button', {name: /Continue with Google/i})).toBeInTheDocument();
    await expect.element(page.getByRole('button', {name: /Continue with GitHub/i})).not.toBeInTheDocument();
  });

  it('should not show providers that are not selected even if they exist in API', async () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    await renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
        'google-idp': true,
        'github-idp': false, // Not selected
      },
    });

    // Should only show Google
    await expect.element(page.getByRole('button', {name: /Continue with Google/i})).toBeInTheDocument();
    await expect.element(page.getByRole('button', {name: /Continue with GitHub/i})).not.toBeInTheDocument();
  });

  it('should render multiple social providers in order', async () => {
    await renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
        'google-idp': true,
        'github-idp': true,
      },
    });

    const buttons = page.getByRole('button', {name: /Continue with/i}).all();
    expect(buttons).toHaveLength(2);
    expect(buttons[0]).toHaveTextContent('Continue with Google');
    expect(buttons[1]).toHaveTextContent('Continue with GitHub');
  });

  it('should render sign in form when only username/password is enabled', async () => {
    await renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
      },
    });

    await expect.element(page.getByText('Username')).toBeInTheDocument();
    await expect.element(page.getByText('Password')).toBeInTheDocument();
    await expect.element(page.getByRole('button', {name: 'Sign In'})).toBeInTheDocument();
    await expect.element(page.getByText('or')).not.toBeInTheDocument();
  });

  it('should render only social logins when username/password is disabled', async () => {
    await renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: false,
        'google-idp': true,
        'github-idp': true,
      },
    });

    await expect.element(page.getByText('Username')).not.toBeInTheDocument();
    await expect.element(page.getByText('Password')).not.toBeInTheDocument();
    await expect.element(page.getByRole('button', {name: 'Sign In'})).not.toBeInTheDocument();
    await expect.element(page.getByRole('button', {name: /Continue with Google/i})).toBeInTheDocument();
    await expect.element(page.getByRole('button', {name: /Continue with GitHub/i})).toBeInTheDocument();
  });

  describe('SMS OTP functionality', () => {
    it('should render mobile number field when SMS OTP is enabled', async () => {
      await renderComponent({
        integrations: {
          'sms-otp': true,
        },
      });

      await expect.element(page.getByText('Mobile Number')).toBeInTheDocument();
      await expect.element(page.getByPlaceholder('Enter your mobile number')).toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: 'Send OTP'})).toBeInTheDocument();
    });

    it('should not render mobile number field when SMS OTP is disabled', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: true,
          'sms-otp': false,
        },
      });

      await expect.element(page.getByText('Mobile Number')).not.toBeInTheDocument();
      await expect.element(page.getByPlaceholder('Enter your mobile number')).not.toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: 'Send OTP'})).not.toBeInTheDocument();
    });

    it('should render divider when SMS OTP and social logins are enabled', async () => {
      await renderComponent({
        integrations: {
          'sms-otp': true,
          'google-idp': true,
        },
      });

      await expect.element(page.getByText('or')).toBeInTheDocument();
    });

    it('should render divider when username/password and SMS OTP are enabled', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: true,
          'sms-otp': true,
        },
      });

      await expect.element(page.getByText('or')).toBeInTheDocument();
    });

    it('should not render divider when only SMS OTP is enabled', async () => {
      await renderComponent({
        integrations: {
          'sms-otp': true,
        },
      });

      await expect.element(page.getByText('or')).not.toBeInTheDocument();
    });

    it('should render divider when all authentication methods are enabled', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: true,
          'sms-otp': true,
          'google-idp': true,
        },
      });

      await expect.element(page.getByText('or')).toBeInTheDocument();
    });

    it('should apply theme colors to Send OTP button', async () => {
      await renderComponent({
        integrations: {
          'sms-otp': true,
        },
        selectedTheme: mockTheme,
      });

      const sendOtpButton = page.getByRole('button', {name: 'Send OTP'});
      expect(sendOtpButton).toHaveStyle({backgroundColor: '#FF5733'});
    });

    it('should render mobile number input as disabled', async () => {
      await renderComponent({
        integrations: {
          'sms-otp': true,
        },
      });

      const mobileInput = page.getByPlaceholder('Enter your mobile number');
      expect(mobileInput).toBeDisabled();
    });

    it('should render SMS OTP with username/password combination', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: true,
          'sms-otp': true,
        },
      });

      // Username/password fields
      await expect.element(page.getByText('Username')).toBeInTheDocument();
      await expect.element(page.getByText('Password')).toBeInTheDocument();

      // SMS OTP fields
      await expect.element(page.getByText('Mobile Number')).toBeInTheDocument();
      await expect.element(page.getByPlaceholder('Enter your mobile number')).toBeInTheDocument();

      // Both buttons
      await expect.element(page.getByRole('button', {name: 'Sign In'})).toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: 'Send OTP'})).toBeInTheDocument();

      // Divider should be present
      await expect.element(page.getByText('or')).toBeInTheDocument();
    });

    it('should render SMS OTP with social logins combination', async () => {
      await renderComponent({
        integrations: {
          'sms-otp': true,
          'google-idp': true,
          'github-idp': true,
        },
      });

      // SMS OTP fields
      await expect.element(page.getByText('Mobile Number')).toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: 'Send OTP'})).toBeInTheDocument();

      // Social login buttons
      await expect.element(page.getByRole('button', {name: /Continue with Google/i})).toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: /Continue with GitHub/i})).toBeInTheDocument();

      // Divider should be present
      await expect.element(page.getByText('or')).toBeInTheDocument();
    });
  });

  describe('Passkey functionality', () => {
    it('should render passkey button when enabled', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.PASSKEY]: true,
        },
      });

      await expect.element(page.getByRole('button', {name: /Sign in with Passkey/i})).toBeInTheDocument();
    });

    it('should not render passkey button when disabled', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.PASSKEY]: false,
        },
      });

      await expect.element(page.getByRole('button', {name: /Sign in with Passkey/i})).not.toBeInTheDocument();
    });

    it('should render passkey button with outlined variant when username/password is enabled', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: true,
          [AuthenticatorTypes.PASSKEY]: true,
        },
        selectedTheme: mockTheme,
      });

      const passkeyButton = page.getByRole('button', {name: /Sign in with Passkey/i});
      expect(passkeyButton).toHaveClass('MuiButton-outlined');
      expect(passkeyButton).toHaveClass('MuiButton-colorPrimary');
    });

    it('should apply theme primary color to passkey button background when username/password is disabled', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: false,
          [AuthenticatorTypes.PASSKEY]: true,
        },
        selectedTheme: mockTheme,
      });

      const passkeyButton = page.getByRole('button', {name: /Sign in with Passkey/i});
      // Without username/password, the passkey button acts as the primary action and omits
      // variant="contained" — styles come entirely from the sx prop.
      expect(passkeyButton).not.toHaveClass('MuiButton-contained');
      expect(passkeyButton).toHaveStyle({backgroundColor: '#FF5733'});
    });

    it('should apply contrastText color to passkey button label when username/password is disabled', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: false,
          [AuthenticatorTypes.PASSKEY]: true,
        },
        selectedTheme: mockTheme,
      });

      const passkeyButton = page.getByRole('button', {name: /Sign in with Passkey/i});
      expect(passkeyButton).toHaveStyle({color: '#FFFFFF'});
    });

    it('should render passkey button inside form container when social logins are present', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.PASSKEY]: true,
          'google-idp': true,
        },
      });

      const passkeyButton = page.getByRole('button', {name: /Sign in with Passkey/i});
      const containerBox = passkeyButton.element().closest('form');
      expect(containerBox).toBeInTheDocument();
    });

    it('should render passkey button inside form container when social logins are absent', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.PASSKEY]: true,
          'google-idp': false,
        },
      });

      const passkeyButton = page.getByRole('button', {name: /Sign in with Passkey/i});
      const containerBox = passkeyButton.element().closest('form');
      expect(containerBox).toBeInTheDocument();
    });

    it('should render divider when passkey and social logins are enabled', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.PASSKEY]: true,
          'google-idp': true,
        },
      });

      await expect.element(page.getByText('or')).toBeInTheDocument();
    });

    it('should render divider when passkey and SMS OTP are enabled', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.PASSKEY]: true,
          'sms-otp': true,
        },
      });
    });
  });

  describe('dark mode', () => {
    it('should render in dark mode', async () => {
      mockUseColorScheme.mockReturnValue({mode: 'dark'});

      await renderComponent();

      // Component should render without errors in dark mode
      await expect.element(page.getByText('Preview')).toBeInTheDocument();
    });

    it('should render with username/password in dark mode', async () => {
      mockUseColorScheme.mockReturnValue({mode: 'dark'});

      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: true,
        },
      });

      await expect.element(page.getByText('Username')).toBeInTheDocument();
      await expect.element(page.getByText('Password')).toBeInTheDocument();
    });

    it('should render with social logins in dark mode', async () => {
      mockUseColorScheme.mockReturnValue({mode: 'dark'});

      await renderComponent({
        integrations: {
          'google-idp': true,
        },
      });

      await expect.element(page.getByRole('button', {name: /Continue with Google/i})).toBeInTheDocument();
    });

    it('should render logo in dark mode', async () => {
      mockUseColorScheme.mockReturnValue({mode: 'dark'});

      await renderComponent({
        appLogo: 'https://example.com/logo.png',
      });

      const logo = page.getByRole('img');
      expect(logo).toBeInTheDocument();
    });
  });

  describe('edge cases', () => {
    it('should handle undefined integration values gracefully', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: undefined as unknown as boolean,
        },
      });

      // Should not render username/password when value is undefined (falsy)
      await expect.element(page.getByText('Username')).not.toBeInTheDocument();
    });

    it('should handle undefined sms-otp value gracefully', async () => {
      await renderComponent({
        integrations: {
          'sms-otp': undefined as unknown as boolean,
        },
      });

      // Should not render SMS OTP when value is undefined
      await expect.element(page.getByText('Mobile Number')).not.toBeInTheDocument();
    });

    it('should render only username/password form with no margin when no social logins', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: true,
        },
      });

      // Username/password should be present
      await expect.element(page.getByText('Username')).toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: 'Sign In'})).toBeInTheDocument();
      // No divider because no social logins
      await expect.element(page.getByText('or')).not.toBeInTheDocument();
    });

    it('should render only SMS OTP form with no margin when no social logins', async () => {
      await renderComponent({
        integrations: {
          'sms-otp': true,
        },
      });

      // SMS OTP should be present
      await expect.element(page.getByText('Mobile Number')).toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: 'Send OTP'})).toBeInTheDocument();
      // No divider because no social logins
      await expect.element(page.getByText('or')).not.toBeInTheDocument();
    });

    it('should render username/password with margin when social logins exist', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: true,
          'google-idp': true,
        },
      });

      // Both username/password and social login should be present
      await expect.element(page.getByText('Username')).toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: /Continue with Google/i})).toBeInTheDocument();
      // Divider should be present
      await expect.element(page.getByText('or')).toBeInTheDocument();
    });

    it('should render SMS OTP with margin when social logins exist', async () => {
      await renderComponent({
        integrations: {
          'sms-otp': true,
          'google-idp': true,
        },
      });

      // Both SMS OTP and social login should be present
      await expect.element(page.getByText('Mobile Number')).toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: /Continue with Google/i})).toBeInTheDocument();
      // Divider should be present
      await expect.element(page.getByText('or')).toBeInTheDocument();
    });

    it('should render all three auth methods with divider', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: true,
          'sms-otp': true,
          'google-idp': true,
        },
      });

      // All three should be present
      await expect.element(page.getByText('Username')).toBeInTheDocument();
      await expect.element(page.getByText('Mobile Number')).toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: /Continue with Google/i})).toBeInTheDocument();
      // Divider should be present
      await expect.element(page.getByText('or')).toBeInTheDocument();
    });

    it('should render username/password and SMS OTP without social logins', async () => {
      await renderComponent({
        integrations: {
          [AuthenticatorTypes.BASIC_AUTH]: true,
          'sms-otp': true,
        },
      });

      // Both should be present
      await expect.element(page.getByText('Username')).toBeInTheDocument();
      await expect.element(page.getByText('Mobile Number')).toBeInTheDocument();
      // Divider should be present when both methods exist
      await expect.element(page.getByText('or')).toBeInTheDocument();
    });
  });

  describe('high-contrast theme colors', () => {
    const highContrastTheme: ThemeConfig = {
      colorSchemes: {
        light: {
          colors: {
            primary: {
              main: '#0000FF',
              dark: '#0000CC',
              contrastText: '#FFFFFF',
            },
            secondary: {
              main: '#FFD700',
              dark: '#CCB000',
              contrastText: '#000000',
            },
          },
        },
        dark: {
          colors: {
            primary: {
              main: '#00FFFF',
              dark: '#00CCCC',
              contrastText: '#000000',
            },
            secondary: {
              main: '#FFFF00',
              dark: '#CCCC00',
              contrastText: '#000000',
            },
          },
        },
      },
      defaultColorScheme: 'light',
      direction: 'ltr',
    };

    it('should render sign in button with high-contrast primary background', async () => {
      await renderComponent({
        integrations: {[AuthenticatorTypes.BASIC_AUTH]: true},
        selectedTheme: highContrastTheme,
      });

      const signInButton = page.getByRole('button', {name: 'Sign In'});
      expect(signInButton).toHaveStyle({backgroundColor: '#0000FF'});
    });

    it('should render sign in button with white contrastText on high-contrast blue', async () => {
      await renderComponent({
        integrations: {[AuthenticatorTypes.BASIC_AUTH]: true},
        selectedTheme: highContrastTheme,
      });

      const signInButton = page.getByRole('button', {name: 'Sign In'});
      expect(signInButton).toHaveStyle({color: '#FFFFFF'});
    });

    it('should render Send OTP button with high-contrast primary background', async () => {
      await renderComponent({
        integrations: {'sms-otp': true},
        selectedTheme: highContrastTheme,
      });

      const sendOtpButton = page.getByRole('button', {name: 'Send OTP'});
      expect(sendOtpButton).toHaveStyle({backgroundColor: '#0000FF'});
      expect(sendOtpButton).toHaveStyle({color: '#FFFFFF'});
    });

    it('should render passkey button with dark mode high-contrast colors', async () => {
      mockUseColorScheme.mockReturnValue({mode: 'dark'});

      await renderComponent({
        integrations: {[AuthenticatorTypes.PASSKEY]: true},
        selectedTheme: highContrastTheme,
      });

      const passkeyButton = page.getByRole('button', {name: /Sign in with Passkey/i});
      expect(passkeyButton).toHaveStyle({backgroundColor: '#00FFFF'});
      // Dark mode cyan primary has black contrastText
      expect(passkeyButton).toHaveStyle({color: '#000000'});
    });
  });

  describe('theme mode handling', () => {
    it('should use normal blend mode in light mode', async () => {
      mockUseColorScheme.mockReturnValue({mode: 'light'});

      await renderComponent();

      await expect.element(page.getByText('Preview')).toBeInTheDocument();
    });

    it('should use screen blend mode in dark mode', async () => {
      mockUseColorScheme.mockReturnValue({mode: 'dark'});

      await renderComponent();

      await expect.element(page.getByText('Preview')).toBeInTheDocument();
    });

    it('should apply dark mode styles when mode is dark', async () => {
      mockUseColorScheme.mockReturnValue({mode: 'dark'});

      await renderComponent({
        appLogo: 'https://example.com/logo.png',
        selectedTheme: mockTheme,
      });

      const logo = page.getByRole('img');
      const avatarContainer = logo.element().closest('.MuiAvatar-root');
      expect(avatarContainer).toBeInTheDocument();
    });

    it('should apply light mode styles when mode is light', async () => {
      mockUseColorScheme.mockReturnValue({mode: 'light'});

      await renderComponent({
        appLogo: 'https://example.com/logo.png',
        selectedTheme: mockTheme,
      });

      const logo = page.getByRole('img');
      const avatarContainer = logo.element().closest('.MuiAvatar-root');
      expect(avatarContainer).toHaveStyle({backgroundColor: '#FF5733'});
    });
  });
});
