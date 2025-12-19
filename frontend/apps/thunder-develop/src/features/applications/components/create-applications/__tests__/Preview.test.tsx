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
import {render, screen} from '@testing-library/react';
import {IdentityProviderTypes, type IdentityProvider} from '@/features/integrations/models/identity-provider';
import {AuthenticatorTypes} from '@/features/integrations/models/authenticators';
import Preview, {type PreviewProps} from '../Preview';

// Mock the @asgardeo/react module
vi.mock('@asgardeo/react', () => ({
  BaseSignIn: ({children}: {children: () => React.ReactNode}) => <div>{children()}</div>,
  ThemeProvider: ({children}: {children: React.ReactNode}) => <div>{children}</div>,
}));

// Mock the useIdentityProviders hook
vi.mock('@/features/integrations/api/useIdentityProviders');

const {default: useIdentityProviders} = await import('@/features/integrations/api/useIdentityProviders');

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
    appName: 'My Application',
    appLogo: 'https://example.com/logo.png',
    selectedColor: '#FF5733',
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
  });

  const renderComponent = (props: Partial<PreviewProps> = {}) => render(<Preview {...defaultProps} {...props} />);

  it('should render the preview title', () => {
    renderComponent();

    expect(screen.getByText('Preview')).toBeInTheDocument();
  });

  it('should render the application logo when provided', () => {
    renderComponent();

    const logo = screen.getByRole('img');
    expect(logo).toBeInTheDocument();
    expect(logo).toHaveAttribute('src', 'https://example.com/logo.png');
  });

  it('should not render logo when appLogo is null', () => {
    renderComponent({appLogo: null});

    expect(screen.queryByRole('img')).not.toBeInTheDocument();
  });

  it('should render sign in heading with app name', () => {
    renderComponent();

    expect(screen.getByText('Sign in to My Application')).toBeInTheDocument();
  });

  it('should render welcome message', () => {
    renderComponent();

    expect(screen.getByText('Welcome back! Please sign in to continue.')).toBeInTheDocument();
  });

  it('should render username and password fields when username/password is enabled', () => {
    renderComponent();

    expect(screen.getByText('Username')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Enter your Username')).toBeInTheDocument();
    expect(screen.getByText('Password')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Enter your Password')).toBeInTheDocument();
  });

  it('should render sign in button', () => {
    renderComponent();

    expect(screen.getByRole('button', {name: 'Sign In'})).toBeInTheDocument();
  });

  it('should not render username/password fields when disabled', () => {
    renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: false,
        'google-idp': true,
      },
    });

    expect(screen.queryByText('Username')).not.toBeInTheDocument();
    expect(screen.queryByText('Password')).not.toBeInTheDocument();
    expect(screen.queryByRole('button', {name: 'Sign In'})).not.toBeInTheDocument();
  });

  it('should render social login buttons for enabled providers', () => {
    renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
        'google-idp': true,
        'github-idp': true,
      },
    });

    expect(screen.getByRole('button', {name: /Continue with Google/i})).toBeInTheDocument();
    expect(screen.getByRole('button', {name: /Continue with GitHub/i})).toBeInTheDocument();
  });

  it('should not render social login buttons when no providers are enabled', () => {
    renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
      },
    });

    expect(screen.queryByRole('button', {name: /Continue with/i})).not.toBeInTheDocument();
  });

  it('should render divider when both username/password and social logins are enabled', () => {
    renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
        'google-idp': true,
      },
    });

    expect(screen.getByText('or')).toBeInTheDocument();
  });

  it('should not render divider when only username/password is enabled', () => {
    renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
      },
    });

    expect(screen.queryByText('or')).not.toBeInTheDocument();
  });

  it('should not render divider when only social logins are enabled', () => {
    renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: false,
        'google-idp': true,
      },
    });

    expect(screen.queryByText('or')).not.toBeInTheDocument();
  });

  it('should render only selected social providers', () => {
    renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
        'google-idp': true,
        // github-idp not included
      },
    });

    expect(screen.getByRole('button', {name: /Continue with Google/i})).toBeInTheDocument();
    expect(screen.queryByRole('button', {name: /Continue with GitHub/i})).not.toBeInTheDocument();
  });

  it('should handle empty integrations object', () => {
    renderComponent({
      integrations: {},
    });

    // Username/password should not be shown when integrations is empty (defaults to false)
    expect(screen.queryByText('Username')).not.toBeInTheDocument();
    expect(screen.queryByText('Password')).not.toBeInTheDocument();
  });

  it('should handle null app name', () => {
    renderComponent({appName: null});

    // Should still render but with null in the heading
    expect(screen.getByText('Sign in to')).toBeInTheDocument();
  });

  it('should apply selected color to sign in button', () => {
    renderComponent();

    const signInButton = screen.getByRole('button', {name: 'Sign In'});
    expect(signInButton).toHaveStyle({backgroundColor: '#FF5733'});
  });

  it('should apply selected color to logo background', () => {
    renderComponent();

    const logo = screen.getByRole('img');
    const avatarContainer = logo.closest('.MuiAvatar-root');
    expect(avatarContainer).toHaveStyle({backgroundColor: '#FF5733'});
  });

  it('should render input fields as disabled', () => {
    renderComponent();

    const usernameInput = screen.getByPlaceholderText('Enter your Username');
    const passwordInput = screen.getByPlaceholderText('Enter your Password');

    expect(usernameInput).toBeDisabled();
    expect(passwordInput).toBeDisabled();
  });

  it('should render social login buttons as disabled', () => {
    renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
        'google-idp': true,
      },
    });

    const googleButton = screen.getByRole('button', {name: /Continue with Google/i});
    expect(googleButton).toBeDisabled();
  });

  it('should handle when useIdentityProviders returns undefined data', () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
        'google-idp': true,
      },
    });

    // Should not crash, no social providers should be rendered (since data is undefined)
    expect(screen.queryByRole('button', {name: /Continue with/i})).not.toBeInTheDocument();
  });

  it('should only show providers that exist in API and are selected', () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: [mockIdentityProviders[0]], // Only Google in API
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
        'google-idp': true,
        // 'github-idp' is not in API, so even if selected, it won't show
      },
    });

    // Should only show Google (which exists in API)
    expect(screen.getByRole('button', {name: /Continue with Google/i})).toBeInTheDocument();
    expect(screen.queryByRole('button', {name: /Continue with GitHub/i})).not.toBeInTheDocument();
  });

  it('should not show providers that are not selected even if they exist in API', () => {
    vi.mocked(useIdentityProviders).mockReturnValue({
      data: mockIdentityProviders,
      isLoading: false,
      error: null,
    } as ReturnType<typeof useIdentityProviders>);

    renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
        'google-idp': true,
        'github-idp': false, // Not selected
      },
    });

    // Should only show Google
    expect(screen.getByRole('button', {name: /Continue with Google/i})).toBeInTheDocument();
    expect(screen.queryByRole('button', {name: /Continue with GitHub/i})).not.toBeInTheDocument();
  });

  it('should render multiple social providers in order', () => {
    renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
        'google-idp': true,
        'github-idp': true,
      },
    });

    const buttons = screen.getAllByRole('button', {name: /Continue with/i});
    expect(buttons).toHaveLength(2);
    expect(buttons[0]).toHaveTextContent('Continue with Google');
    expect(buttons[1]).toHaveTextContent('Continue with GitHub');
  });

  it('should render sign in form when only username/password is enabled', () => {
    renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: true,
      },
    });

    expect(screen.getByText('Username')).toBeInTheDocument();
    expect(screen.getByText('Password')).toBeInTheDocument();
    expect(screen.getByRole('button', {name: 'Sign In'})).toBeInTheDocument();
    expect(screen.queryByText('or')).not.toBeInTheDocument();
  });

  it('should render only social logins when username/password is disabled', () => {
    renderComponent({
      integrations: {
        [AuthenticatorTypes.BASIC_AUTH]: false,
        'google-idp': true,
        'github-idp': true,
      },
    });

    expect(screen.queryByText('Username')).not.toBeInTheDocument();
    expect(screen.queryByText('Password')).not.toBeInTheDocument();
    expect(screen.queryByRole('button', {name: 'Sign In'})).not.toBeInTheDocument();
    expect(screen.getByRole('button', {name: /Continue with Google/i})).toBeInTheDocument();
    expect(screen.getByRole('button', {name: /Continue with GitHub/i})).toBeInTheDocument();
  });

  describe('Custom flow selected', () => {
    it('should render custom flow overlay when isCustomFlowSelected is true', () => {
      renderComponent({
        isCustomFlowSelected: true,
      });

      expect(screen.getByText('Custom Flow Selected')).toBeInTheDocument();
    });

    it('should render custom flow name when provided', () => {
      renderComponent({
        isCustomFlowSelected: true,
        customFlowName: 'My Custom Flow',
      });

      expect(screen.getByText('My Custom Flow')).toBeInTheDocument();
    });

    it('should render hint message about Flow Builder', () => {
      renderComponent({
        isCustomFlowSelected: true,
      });

      expect(screen.getByText(/Preview is not available for custom flows/i)).toBeInTheDocument();
    });

    it('should not render overlay when isCustomFlowSelected is false', () => {
      renderComponent({
        isCustomFlowSelected: false,
      });

      expect(screen.queryByText('Custom Flow Selected')).not.toBeInTheDocument();
    });

    it('should still render preview container when custom flow is selected', () => {
      renderComponent({
        isCustomFlowSelected: true,
      });

      expect(screen.getByText('Preview')).toBeInTheDocument();
    });
  });
});
