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

import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';
import {page, userEvent} from 'vitest/browser';
import {render, getByDisplayValue} from '@thunder/test-utils/browser';
import IntegrationGuide from '../IntegrationGuide';

// Use vi.hoisted() so mockNavigate is available inside the vi.mock factory
const {mockNavigate} = vi.hoisted(() => ({mockNavigate: vi.fn()}));

// Mock navigate function
vi.mock('react-router', async () => {
  const actual = await vi.importActual('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

// Mock TechnologyGuide component
vi.mock('../../edit-application/integration-guides/TechnologyGuide', () => ({
  default: () => <div data-testid="technology-guide">Technology Guide</div>,
}));

describe('IntegrationGuide', () => {
  const defaultProps = {
    appName: 'Test Application',
    appLogo: 'https://example.com/logo.png',
    selectedColor: '#FF5733',
    hasOAuthConfig: false,
    applicationId: 'app-123',
  };

  // Store mock at module level so tests can access it
  let clipboardWriteTextMock: ReturnType<typeof vi.fn>;
  // Store original document.execCommand to restore after tests
  let originalExecCommand: typeof document.execCommand;

  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers({shouldAdvanceTime: true});
    // Mock clipboard API using defineProperty
    clipboardWriteTextMock = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, 'clipboard', {
      value: {
        writeText: clipboardWriteTextMock,
      },
      writable: true,
      configurable: true,
    });
    // Save original execCommand before any test modifies it
    // eslint-disable-next-line @typescript-eslint/unbound-method
    originalExecCommand = document.execCommand;
  });

  afterEach(() => {
    vi.useRealTimers();
    // Restore original execCommand to prevent test pollution
    document.execCommand = originalExecCommand;
  });

  describe('Rendering', () => {
    it('should render the component', async () => {
      await render(<IntegrationGuide {...defaultProps} />);

      await expect.element(page.getByText('Application Created Successfully!')).toBeInTheDocument();
    });

    it('should display app name', async () => {
      await render(<IntegrationGuide {...defaultProps} />);

      await expect.element(page.getByText(defaultProps.appName)).toBeInTheDocument();
    });

    it('should display success message when OAuth not configured', async () => {
      await render(<IntegrationGuide {...defaultProps} hasOAuthConfig={false} />);

      await expect.element(page.getByText('Your application is ready to use')).toBeInTheDocument();
    });

    it('should display guides subtitle when integrationGuides are provided', async () => {
      const props = {
        ...defaultProps,
        integrationGuides: {
          react: {
            llm_prompt: {
              id: 'test-guide',
              title: 'Test Guide',
              description: 'Test description',
              type: 'llm' as const,
              icon: 'test-icon',
              overview: 'Test overview',
              prerequisites: [],
              steps: [],
            },
            manual_steps: [],
          },
        },
      };

      await render(<IntegrationGuide {...props} />);

      await expect.element(page.getByText('Choose how you want to integrate sign-in to your application')).toBeInTheDocument();
    });

    it('should render success icon', async () => {
      await render(<IntegrationGuide {...defaultProps} />);

      await expect.element(page.getByRole('img', {name: 'Success'})).toBeInTheDocument();
    });
  });

  describe('OAuth Configuration', () => {
    it('should display client ID when hasOAuthConfig is true', async () => {
      const props = {
        ...defaultProps,
        hasOAuthConfig: true,
        clientId: 'test_client_id',
      };

      await render(<IntegrationGuide {...props} />);

      await expect.element(getByDisplayValue('test_client_id')).toBeInTheDocument();
    });

    it('should display client secret when provided', async () => {
      const props = {
        ...defaultProps,
        hasOAuthConfig: true,
        clientId: 'test_client_id',
        clientSecret: 'test_secret',
      };

      await render(<IntegrationGuide {...props} />);

      // Secret should be hidden by default
      const secretInput = getByDisplayValue('test_secret');
      expect(secretInput).toHaveAttribute('type', 'password');
    });

    it('should not display OAuth credentials when hasOAuthConfig is false', async () => {
      await render(<IntegrationGuide {...defaultProps} hasOAuthConfig={false} />);

      await expect.element(page.getByText('applications:create.integrationGuide.clientId')).not.toBeInTheDocument();
    });

    it('should display warning alert when client secret is present', async () => {
      const props = {
        ...defaultProps,
        hasOAuthConfig: true,
        clientId: 'test_client_id',
        clientSecret: 'test_secret',
      };

      await render(<IntegrationGuide {...props} />);

      await expect.element(page.getByText("Make sure to copy your client secret now. You won't be able to see it again for security reasons.")).toBeInTheDocument();
    });

    it('should not display warning alert for public clients (no secret)', async () => {
      const props = {
        ...defaultProps,
        hasOAuthConfig: true,
        clientId: 'test_client_id',
        clientSecret: '',
      };

      await render(<IntegrationGuide {...props} />);

      await expect.element(page.getByText("Make sure to copy your client secret now. You won't be able to see it again for security reasons.")).not.toBeInTheDocument();
    });

    it('should toggle client secret visibility', async () => {
      const props = {
        ...defaultProps,
        hasOAuthConfig: true,
        clientId: 'test_client_id',
        clientSecret: 'test_secret',
      };

      await render(<IntegrationGuide {...props} />);

      const secretInput = getByDisplayValue('test_secret');
      expect(secretInput).toHaveAttribute('type', 'password');

      // Find the visibility toggle button - it's in the client secret field
      // The buttons are: app card, client ID copy, visibility toggle, client secret copy
      const buttons = page.getByRole('button').all();
      // Visibility toggle is the second-to-last button (before client secret copy)
      const visibilityButton = buttons[buttons.length - 2];

      await userEvent.click(visibilityButton);
      await vi.waitFor(() => {
        expect(secretInput).toHaveAttribute('type', 'text');
      });

      await userEvent.click(visibilityButton);
      await vi.waitFor(() => {
        expect(secretInput).toHaveAttribute('type', 'password');
      });
    });

    it('should copy client ID to clipboard and show copied message', async () => {
      const props = {
        ...defaultProps,
        hasOAuthConfig: true,
        clientId: 'test_client_id',
      };

      await render(<IntegrationGuide {...props} />);

      // Find the copy button for client ID
      // Buttons are: app card (button role), client ID copy button
      const copyButtons = page.getByRole('button').all();
      // The client ID copy button is the second button (index 1) - after app card
      const clientIdCopyButton = copyButtons[1];

      await userEvent.click(clientIdCopyButton);

      // Check copied message appears (indicates copy was successful)
      await vi.waitFor(() => {
        expect(
page.getByText('Copied to clipboard')).toBeInTheDocument();
      });

      // Advance timers to clear the copied state
      vi.advanceTimersByTime(2500);

      // After timeout, copied message should disappear
      await vi.waitFor(() => {
        expect(
page.getByText('Copied to clipboard')).not.toBeInTheDocument();
      });
    });

    it('should copy client secret to clipboard and show copied message', async () => {
      const props = {
        ...defaultProps,
        hasOAuthConfig: true,
        clientId: 'test_client_id',
        clientSecret: 'test_secret',
      };

      await render(<IntegrationGuide {...props} />);

      // Find all copy buttons - the last one should be for client secret
      const copyButtons = page.getByRole('button').all();
      // Get the last copy button (for client secret)
      const secretCopyButton = copyButtons[copyButtons.length - 1];

      await userEvent.click(secretCopyButton);

      // Check copied message appears (indicates copy was successful)
      await vi.waitFor(() => {
        // There may be two copied messages (one for client ID area, one for secret)
        const copiedMessages = page.getByText('Copied to clipboard').all();
        expect(copiedMessages.length).toBeGreaterThan(0);
      });
    });

    it('should use fallback copy method when clipboard API fails', async () => {

      // Mock clipboard to fail using defineProperty
      Object.defineProperty(navigator, 'clipboard', {
        value: {
          writeText: vi.fn().mockRejectedValue(new Error('Clipboard not available')),
        },
        writable: true,
        configurable: true,
      });

      // Mock document.execCommand
      const execCommandMock = vi.fn().mockReturnValue(true);
      document.execCommand = execCommandMock;

      const props = {
        ...defaultProps,
        hasOAuthConfig: true,
        clientId: 'test_client_id',
      };

      await render(<IntegrationGuide {...props} />);

      // Buttons are: app card, client ID copy button
      const copyButtons = page.getByRole('button').all();
      await userEvent.click(copyButtons[1]);

      await vi.waitFor(() => {
        expect(execCommandMock).toHaveBeenCalledWith('copy');
      });
    });

    it('should handle fallback copy failure gracefully when execCommand throws', async () => {

      // Mock clipboard to fail
      Object.defineProperty(navigator, 'clipboard', {
        value: {
          writeText: vi.fn().mockRejectedValue(new Error('Clipboard not available')),
        },
        writable: true,
        configurable: true,
      });

      // Mock document.execCommand to throw an error
      const execCommandMock = vi.fn().mockImplementation(() => {
        throw new Error('execCommand not supported');
      });
      document.execCommand = execCommandMock;

      const props = {
        ...defaultProps,
        hasOAuthConfig: true,
        clientId: 'test_client_id',
      };

      await render(<IntegrationGuide {...props} />);

      // Buttons are: app card, client ID copy button
      const copyButtons = page.getByRole('button').all();

      // Should not throw even when both copy methods fail
      await expect(userEvent.click(copyButtons[1])).resolves.not.toThrow();

      // execCommand should have been called
      expect(execCommandMock).toHaveBeenCalledWith('copy');
    });

    it('should handle fallback copy failure gracefully when execCommand returns false', async () => {

      // Mock clipboard to fail
      Object.defineProperty(navigator, 'clipboard', {
        value: {
          writeText: vi.fn().mockRejectedValue(new Error('Clipboard not available')),
        },
        writable: true,
        configurable: true,
      });

      // Mock document.execCommand to return false (copy failed)
      const execCommandMock = vi.fn().mockReturnValue(false);
      document.execCommand = execCommandMock;

      const props = {
        ...defaultProps,
        hasOAuthConfig: true,
        clientId: 'test_client_id',
      };

      await render(<IntegrationGuide {...props} />);

      const copyButtons = page.getByRole('button').all();

      // Should not throw
      await expect(userEvent.click(copyButtons[1])).resolves.not.toThrow();
    });

    it('should show copied state on successful fallback copy', async () => {

      // Mock clipboard to fail
      Object.defineProperty(navigator, 'clipboard', {
        value: {
          writeText: vi.fn().mockRejectedValue(new Error('Clipboard not available')),
        },
        writable: true,
        configurable: true,
      });

      // Mock document.execCommand to succeed
      document.execCommand = vi.fn().mockReturnValue(true);

      const props = {
        ...defaultProps,
        hasOAuthConfig: true,
        clientId: 'test_client_id',
      };

      await render(<IntegrationGuide {...props} />);

      const copyButtons = page.getByRole('button').all();
      await userEvent.click(copyButtons[1]);

      // Should show copied message after fallback copy succeeds
      await vi.waitFor(() => {
        expect(
page.getByText('Copied to clipboard')).toBeInTheDocument();
      });

      // Advance timers to clear copied state
      vi.advanceTimersByTime(2500);

      await vi.waitFor(() => {
        expect(
page.getByText('Copied to clipboard')).not.toBeInTheDocument();
      });
    });

    it('should clear existing timeout before setting new one in fallback copy', async () => {

      // Mock clipboard to fail
      Object.defineProperty(navigator, 'clipboard', {
        value: {
          writeText: vi.fn().mockRejectedValue(new Error('Clipboard not available')),
        },
        writable: true,
        configurable: true,
      });

      // Mock document.execCommand to succeed
      document.execCommand = vi.fn().mockReturnValue(true);

      const props = {
        ...defaultProps,
        hasOAuthConfig: true,
        clientId: 'test_client_id',
      };

      await render(<IntegrationGuide {...props} />);

      const copyButtons = page.getByRole('button').all();

      // Click twice rapidly to trigger timeout clearing logic
      await userEvent.click(copyButtons[1]);
      await userEvent.click(copyButtons[1]);

      // Should still show copied message
      await vi.waitFor(() => {
        expect(
page.getByText('Copied to clipboard')).toBeInTheDocument();
      });
    });
  });

  describe('Integration Guides', () => {
    it('should render TechnologyGuide when integrationGuides are provided', async () => {
      const props = {
        ...defaultProps,
        integrationGuides: {
          react: {
            llm_prompt: {
              id: 'test-guide',
              title: 'Test Guide',
              description: 'Test description',
              type: 'llm' as const,
              icon: 'test-icon',
              overview: 'Test overview',
              prerequisites: [],
              steps: [],
            },
            manual_steps: [],
          },
        },
      };

      await render(<IntegrationGuide {...props} />);

      await expect.element(page.getByTestId('technology-guide')).toBeInTheDocument();
    });

    it('should not render TechnologyGuide when integrationGuides are null', async () => {
      const props = {
        ...defaultProps,
        integrationGuides: null,
      };

      await render(<IntegrationGuide {...props} />);

      await expect.element(page.getByTestId('technology-guide')).not.toBeInTheDocument();
    });

    it('should not render TechnologyGuide when integrationGuides are undefined', async () => {
      await render(<IntegrationGuide {...defaultProps} />);

      await expect.element(page.getByTestId('technology-guide')).not.toBeInTheDocument();
    });

    it('should not render app details section when integrationGuides are provided', async () => {
      const props = {
        ...defaultProps,
        integrationGuides: {
          react: {
            llm_prompt: {
              id: 'test-guide',
              title: 'Test Guide',
              description: 'Test description',
              type: 'llm' as const,
              icon: 'test-icon',
              overview: 'Test overview',
              prerequisites: [],
              steps: [],
            },
            manual_steps: [],
          },
        },
      };

      await render(<IntegrationGuide {...props} />);

      await expect.element(page.getByText('Application successfully created')).not.toBeInTheDocument();
    });
  });

  describe('App Logo', () => {
    it('should display app logo when provided', async () => {
      await render(<IntegrationGuide {...defaultProps} />);

      const logo = page.getByAltText('Test Application logo');
      expect(logo).toHaveAttribute('src', defaultProps.appLogo);
    });

    it('should handle null logo gracefully', async () => {
      const props = {
        ...defaultProps,
        appLogo: null,
      };

      await render(<IntegrationGuide {...props} />);

      // Should still render without crashing
      await expect.element(page.getByText(defaultProps.appName)).toBeInTheDocument();
    });

    it('should display first letter avatar when logo is null', async () => {
      const props = {
        ...defaultProps,
        appLogo: null,
      };

      await render(<IntegrationGuide {...props} />);

      // Should display 'T' for 'Test Application'
      await expect.element(page.getByText('T')).toBeInTheDocument();
    });
  });

  describe('Navigation', () => {
    it('should navigate to application details on click when applicationId is provided', async () => {
      mockNavigate.mockResolvedValue(undefined);

      await render(<IntegrationGuide {...defaultProps} />);

      const appCard = page.getByRole('button', {name: 'Click to view application details'});
      await userEvent.click(appCard);

      await vi.waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/applications/app-123');
      });
    });

    it('should navigate on Enter key press', async () => {
      mockNavigate.mockResolvedValue(undefined);

      await render(<IntegrationGuide {...defaultProps} />);

      const appCard = page.getByRole('button', {name: 'Click to view application details'});
      await appCard.focus();
      await userEvent.keyboard('{Enter}');

      await vi.waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/applications/app-123');
      });
    });

    it('should navigate on Space key press', async () => {
      mockNavigate.mockResolvedValue(undefined);

      await render(<IntegrationGuide {...defaultProps} />);

      const appCard = page.getByRole('button', {name: 'Click to view application details'});
      await appCard.focus();
      await userEvent.keyboard(' ');

      await vi.waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/applications/app-123');
      });
    });

    it('should not navigate on other key press', async () => {
      await render(<IntegrationGuide {...defaultProps} />);

      const appCard = page.getByRole('button', {name: 'Click to view application details'});
      await appCard.focus();
      await userEvent.keyboard('{Escape}');

      expect(mockNavigate).not.toHaveBeenCalled();
    });

    it('should not be clickable when applicationId is null', async () => {
      const props = {
        ...defaultProps,
        applicationId: null,
      };

      await render(<IntegrationGuide {...props} />);

      // Should not have button role when applicationId is null
      await expect.element(page.getByRole('button', {name: 'Click to view application details'})).not.toBeInTheDocument();
    });

    it('should handle navigation errors gracefully', async () => {
      mockNavigate.mockRejectedValue(new Error('Navigation failed'));

      await render(<IntegrationGuide {...defaultProps} />);

      const appCard = page.getByRole('button', {name: 'Click to view application details'});

      // Should not throw
      await userEvent.click(appCard);

      expect(mockNavigate).toHaveBeenCalled();
    });
  });

  describe('Cleanup', () => {
    it('should clean up copy timeouts on unmount', async () => {
      const props = {
        ...defaultProps,
        hasOAuthConfig: true,
        clientId: 'test_client_id',
      };

      const {unmount} = await render(<IntegrationGuide {...props} />);

      const copyButtons = page.getByRole('button').all();
      await userEvent.click(copyButtons[0]);

      // Unmount before timeout completes
      await unmount();

      // Advance timers - should not cause errors
      vi.advanceTimersByTime(3000);
    });
  });
});
