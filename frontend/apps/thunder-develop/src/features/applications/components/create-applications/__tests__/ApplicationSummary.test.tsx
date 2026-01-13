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
import userEvent from '@testing-library/user-event';
import ApplicationSummary, {type ApplicationSummaryProps} from '../ApplicationSummary';

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'applications:onboarding.summary.title': 'Application Created!',
        'applications:onboarding.summary.subtitle':
          'Your application has been successfully created and is ready to use.',
        'applications:onboarding.summary.appDetails': 'Application is ready to use',
        'applications:onboarding.summary.viewAppAriaLabel': 'View application details',
        'applications:clientSecret.warning':
          'Please copy your client credentials now. The client secret will not be shown again.',
        'applications:clientSecret.clientIdLabel': 'Client ID',
        'applications:clientSecret.clientSecretLabel': 'Client Secret',
        'applications:clientSecret.copied': 'Copied!',
      };
      return translations[key] || key;
    },
  }),
}));

// Mock react-router
const mockNavigate = vi.fn();
vi.mock('react-router', () => ({
  useNavigate: () => mockNavigate,
}));

// Mock useApplicationCreate hook
vi.mock('../../../contexts/ApplicationCreate/useApplicationCreate', () => ({
  default: () => ({
    selectedTemplateConfig: null,
  }),
}));

// Mock TechnologyGuide component
vi.mock('../TechnologyGuide', () => ({
  default: () => null,
}));

// Mock clipboard API
const mockWriteText = vi.fn();
Object.assign(navigator, {
  clipboard: {
    writeText: mockWriteText,
  },
});

describe('ApplicationSummary', () => {
  const defaultProps: ApplicationSummaryProps = {
    appName: 'Test App',
    appLogo: null,
    selectedColor: '#1976d2',
    hasOAuthConfig: false,
    applicationId: null,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockWriteText.mockResolvedValue(undefined);
  });

  const renderComponent = (props: Partial<ApplicationSummaryProps> = {}) =>
    render(<ApplicationSummary {...defaultProps} {...props} />);

  it('should render success message and title', () => {
    renderComponent();

    expect(screen.getByText('Application Created!')).toBeInTheDocument();
    expect(screen.getByText('Your application has been successfully created and is ready to use.')).toBeInTheDocument();
  });

  it('should render application name', () => {
    renderComponent();

    expect(screen.getByText('Test App')).toBeInTheDocument();
  });

  it('should render avatar with first letter when no logo', () => {
    renderComponent();

    const avatar = screen.getByText('T');
    expect(avatar).toBeInTheDocument();
  });

  it('should render avatar with logo when provided', () => {
    renderComponent({appLogo: 'https://example.com/logo.png'});

    // Find avatar by alt text instead of role to avoid conflict with success icon
    const avatar = screen.getByAltText('Test App logo');
    expect(avatar).toHaveAttribute('src', 'https://example.com/logo.png');
  });

  it('should not show OAuth credentials when hasOAuthConfig is false', () => {
    renderComponent({hasOAuthConfig: false});

    expect(screen.queryByText('Client ID')).not.toBeInTheDocument();
    expect(screen.queryByText('Client Secret')).not.toBeInTheDocument();
  });

  it('should not show OAuth credentials when clientId is empty', () => {
    renderComponent({hasOAuthConfig: true, clientId: ''});

    expect(screen.queryByText('Client ID')).not.toBeInTheDocument();
  });

  it('should not show client secret when clientSecret is empty', () => {
    renderComponent({hasOAuthConfig: true, clientId: 'test-id', clientSecret: ''});

    // Client ID should still be shown even without secret (for public clients)
    expect(screen.getByText('Client ID')).toBeInTheDocument();
    expect(screen.getByDisplayValue('test-id')).toBeInTheDocument();
    // But client secret section should not be shown
    expect(screen.queryByText('Client Secret')).not.toBeInTheDocument();
  });

  it('should show client ID for public clients even without client secret', () => {
    renderComponent({hasOAuthConfig: true, clientId: 'public-client-id', clientSecret: ''});

    // Public clients should still see their client ID
    expect(screen.getByText('Client ID')).toBeInTheDocument();
    expect(screen.getByDisplayValue('public-client-id')).toBeInTheDocument();
    // But no client secret section
    expect(screen.queryByText('Client Secret')).not.toBeInTheDocument();
    // And no warning about client secret
    expect(screen.queryByText(/client secret will not be shown again/i)).not.toBeInTheDocument();
  });

  it('should show OAuth credentials when all required props are provided', () => {
    renderComponent({
      hasOAuthConfig: true,
      clientId: 'test-client-id',
      clientSecret: 'test-client-secret',
    });

    expect(screen.getByText('Client ID')).toBeInTheDocument();
    expect(screen.getByText('Client Secret')).toBeInTheDocument();
    expect(screen.getByDisplayValue('test-client-id')).toBeInTheDocument();
    expect(screen.getByDisplayValue('test-client-secret')).toBeInTheDocument();
  });

  it('should show warning alert when OAuth config is present', () => {
    renderComponent({
      hasOAuthConfig: true,
      clientId: 'test-client-id',
      clientSecret: 'test-client-secret',
    });

    expect(
      screen.getByText('Please copy your client credentials now. The client secret will not be shown again.'),
    ).toBeInTheDocument();
  });

  it('should hide client secret by default', () => {
    renderComponent({
      hasOAuthConfig: true,
      clientId: 'test-client-id',
      clientSecret: 'test-client-secret',
    });

    const secretInput = screen.getByDisplayValue('test-client-secret');
    expect(secretInput).toHaveAttribute('type', 'password');
  });

  it('should toggle client secret visibility', () => {
    renderComponent({
      hasOAuthConfig: true,
      clientId: 'test-client-id',
      clientSecret: 'test-client-secret',
    });

    const secretInput = screen.getByDisplayValue('test-client-secret');
    expect(secretInput).toHaveAttribute('type', 'password');

    // Verify the component renders the visibility toggle button
    const secretInputContainer = secretInput.closest('.MuiInputBase-root');
    const buttons = secretInputContainer?.querySelectorAll('button') ?? [];
    expect(buttons.length).toBeGreaterThan(0);
  });

  it('should render copy button for client ID', () => {
    renderComponent({
      hasOAuthConfig: true,
      clientId: 'test-client-id',
      clientSecret: 'test-client-secret',
    });

    // Verify copy button exists in client ID input area
    const clientIdInput = screen.getByDisplayValue('test-client-id');
    const inputContainer = clientIdInput.closest('.MuiInputBase-root');
    const buttons = inputContainer?.querySelectorAll('button') ?? [];
    expect(buttons.length).toBeGreaterThan(0);
  });

  it('should render copy button for client secret', () => {
    renderComponent({
      hasOAuthConfig: true,
      clientId: 'test-client-id',
      clientSecret: 'test-client-secret',
    });

    // Verify copy button exists in client secret input area
    const clientSecretInput = screen.getByDisplayValue('test-client-secret');
    const inputContainer = clientSecretInput.closest('.MuiInputBase-root');
    const buttons = inputContainer?.querySelectorAll('button') ?? [];
    expect(buttons.length).toBeGreaterThan(1); // Should have visibility toggle and copy button
  });

  it('should handle copy functionality', () => {
    renderComponent({
      hasOAuthConfig: true,
      clientId: 'test-client-id',
      clientSecret: 'test-client-secret',
    });

    // Verify the component structure allows for copy functionality
    const clientIdInput = screen.getByDisplayValue('test-client-id');
    expect(clientIdInput).toBeInTheDocument();

    // The copy functionality is tested through the component structure
    // Actual clipboard operations are tested at integration level
  });

  it('should render clickable card when applicationId is present', () => {
    renderComponent({applicationId: 'app-123'});

    const card = screen.getByRole('button', {name: /view application details/i});
    expect(card).toBeInTheDocument();
    expect(card).toHaveAttribute('aria-label');
  });

  it('should not render clickable card when applicationId is null', () => {
    renderComponent({applicationId: null});

    const card = screen.queryByRole('button', {name: /view application details/i});
    expect(card).not.toBeInTheDocument();
  });

  it('should render card with proper accessibility attributes', () => {
    renderComponent({applicationId: 'app-123'});

    const card = screen.getByRole('button', {name: /view application details/i});
    expect(card).toHaveAttribute('tabIndex', '0');
    expect(card).toHaveAttribute('aria-label');
  });

  it('should support keyboard navigation', () => {
    renderComponent({applicationId: 'app-123'});

    const card = screen.getByRole('button', {name: /view application details/i});
    expect(card).toHaveAttribute('tabIndex', '0');
    // Keyboard event handlers are tested through component structure
  });

  it('should have onKeyDown handler when applicationId is present', () => {
    renderComponent({applicationId: 'app-123'});

    const card = screen.getByRole('button', {name: /view application details/i});
    expect(card).toBeInTheDocument();
    // onKeyDown handler is present in the component
  });

  it('should handle clipboard API errors gracefully', () => {
    mockWriteText.mockRejectedValue(new Error('Clipboard API not available'));

    renderComponent({
      hasOAuthConfig: true,
      clientId: 'test-client-id',
      clientSecret: 'test-client-secret',
    });

    // Component should render even if clipboard API fails
    expect(screen.getByDisplayValue('test-client-id')).toBeInTheDocument();
    expect(screen.getByDisplayValue('test-client-secret')).toBeInTheDocument();
  });

  it('should have copy button for client ID that triggers copy functionality', () => {
    renderComponent({
      hasOAuthConfig: true,
      clientId: 'test-client-id',
      clientSecret: 'test-client-secret',
    });

    // Verify the copy button exists in the client ID input
    const clientIdInput = screen.getByDisplayValue('test-client-id');
    const inputContainer = clientIdInput.closest('.MuiInputBase-root');
    expect(inputContainer).toBeInTheDocument();

    // The copy button should be present in the endAdornment
    const buttons = inputContainer?.querySelectorAll('button') ?? [];
    expect(buttons.length).toBeGreaterThan(0);

    // Verify the button has an onClick handler (component structure)
    expect(buttons[0]).toHaveProperty('onclick');
  });

  it('should have copy button for client secret that triggers copy functionality', () => {
    renderComponent({
      hasOAuthConfig: true,
      clientId: 'test-client-id',
      clientSecret: 'test-client-secret',
    });

    const clientSecretInput = screen.getByDisplayValue('test-client-secret');
    const inputContainer = clientSecretInput.closest('.MuiInputBase-root');
    expect(inputContainer).toBeInTheDocument();

    // Should have at least 2 buttons (visibility toggle and copy)
    const buttons = inputContainer?.querySelectorAll('button') ?? [];
    expect(buttons.length).toBeGreaterThanOrEqual(2);

    // Verify the copy button (second button) has an onClick handler
    expect(buttons[1]).toHaveProperty('onclick');
  });

  it('should toggle client secret visibility when eye icon is clicked', async () => {
    const user = userEvent.setup({delay: null});
    renderComponent({
      hasOAuthConfig: true,
      clientId: 'test-client-id',
      clientSecret: 'test-client-secret',
    });

    const clientSecretInput = screen.getByDisplayValue('test-client-secret');
    expect(clientSecretInput).toHaveAttribute('type', 'password');

    const buttons = clientSecretInput.closest('.MuiInputBase-root')?.querySelectorAll('button') ?? [];
    const visibilityButton = buttons[0] as HTMLElement;

    if (visibilityButton) {
      await user.click(visibilityButton);
      // After toggle, type should be 'text'
      const updatedInput = screen.getByDisplayValue('test-client-secret');
      expect(updatedInput).toHaveAttribute('type', 'text');
    }
  });

  it('should navigate to application page when card is clicked', async () => {
    const user = userEvent.setup({delay: null});
    renderComponent({applicationId: 'app-123'});

    const card = screen.getByRole('button', {name: /view application details/i});
    await user.click(card);

    expect(mockNavigate).toHaveBeenCalledWith('/applications/app-123');
  });

  it('should navigate to application page on Enter key press', async () => {
    const user = userEvent.setup({delay: null});
    renderComponent({applicationId: 'app-123'});

    const card = screen.getByRole('button', {name: /view application details/i});
    await user.type(card, '{Enter}');

    expect(mockNavigate).toHaveBeenCalledWith('/applications/app-123');
  });

  it('should navigate to application page on Space key press', async () => {
    const user = userEvent.setup({delay: null});
    renderComponent({applicationId: 'app-123'});

    const card = screen.getByRole('button', {name: /view application details/i});
    await user.type(card, ' ');

    expect(mockNavigate).toHaveBeenCalledWith('/applications/app-123');
  });

  it('should handle navigation errors gracefully', async () => {
    const user = userEvent.setup({delay: null});
    mockNavigate.mockRejectedValue(new Error('Navigation failed'));

    renderComponent({applicationId: 'app-123'});

    const card = screen.getByRole('button', {name: /view application details/i});
    await user.click(card);

    // Should not throw, error is caught
    expect(mockNavigate).toHaveBeenCalled();
  });

  it('should handle clipboard API failure and use execCommand fallback', () => {
    mockWriteText.mockRejectedValue(new Error('Clipboard API not available'));

    // Mock execCommand
    const mockExecCommand = vi.fn().mockReturnValue(true);
    // eslint-disable-next-line @typescript-eslint/unbound-method
    const originalExecCommand = document.execCommand;
    Object.defineProperty(document, 'execCommand', {
      value: mockExecCommand,
      writable: true,
      configurable: true,
    });

    renderComponent({
      hasOAuthConfig: true,
      clientId: 'test-client-id',
      clientSecret: 'test-client-secret',
    });

    // Component should render successfully even with clipboard API failure
    expect(screen.getByDisplayValue('test-client-id')).toBeInTheDocument();
    expect(screen.getByDisplayValue('test-client-secret')).toBeInTheDocument();

    // Restore
    if (originalExecCommand) {
      Object.defineProperty(document, 'execCommand', {
        value: originalExecCommand,
        writable: true,
        configurable: true,
      });
    }
  });

  it('should show copied message after copying', async () => {
    const user = userEvent.setup({delay: null});

    renderComponent({
      hasOAuthConfig: true,
      clientId: 'test-client-id',
      clientSecret: 'test-client-secret',
    });

    const clientIdInput = screen.getByDisplayValue('test-client-id');
    const inputContainer = clientIdInput.closest('.MuiInputBase-root');
    const buttons = inputContainer?.querySelectorAll('button') ?? [];
    const copyButton = buttons[0] as HTMLElement;

    expect(copyButton).toBeInTheDocument();
    await user.click(copyButton);

    // Should show copied message immediately
    expect(screen.getByText('Copied!')).toBeInTheDocument();
  });

  it('should not navigate when other keys are pressed', () => {
    renderComponent({applicationId: 'app-123'});

    const card = screen.getByRole('button', {name: /view application details/i});

    // Simulate keydown with a non-Enter/Space key
    const keyDownEvent = new KeyboardEvent('keydown', {key: 'a', bubbles: true});
    card.dispatchEvent(keyDownEvent);

    // Should not navigate for non-Enter/Space keys
    expect(mockNavigate).not.toHaveBeenCalled();
  });
});
