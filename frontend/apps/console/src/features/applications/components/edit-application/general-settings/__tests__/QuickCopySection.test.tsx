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
import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import QuickCopySection from '../QuickCopySection';
import type {Application} from '../../../../models/application';
import type {OAuth2Config} from '../../../../models/oauth';

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

describe('QuickCopySection', () => {
  const mockOnCopyToClipboard = vi.fn();
  const mockApplication: Application = {
    id: 'app-123',
    name: 'Test App',
  } as Application;

  const mockOAuth2Config: OAuth2Config = {
    client_id: 'client-123',
    client_secret: 'secret-456',
  } as OAuth2Config;

  beforeEach(() => {
    vi.clearAllMocks();
    mockOnCopyToClipboard.mockResolvedValue(undefined);
  });

  describe('Rendering', () => {
    it('should render the settings card with title and description', () => {
      render(
        <QuickCopySection application={mockApplication} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />,
      );

      expect(screen.getByTestId('card-title')).toHaveTextContent('Quick Copy');
      expect(screen.getByTestId('card-description')).toHaveTextContent(
        'Copy application identifiers for use in your code.',
      );
    });

    it('should render application ID field', () => {
      render(
        <QuickCopySection application={mockApplication} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />,
      );

      expect(screen.getByLabelText('Application ID')).toBeInTheDocument();
      expect(screen.getByDisplayValue('app-123')).toBeInTheDocument();
    });

    it('should render client ID field when OAuth2 config is provided', () => {
      render(
        <QuickCopySection
          application={mockApplication}
          oauth2Config={mockOAuth2Config}
          copiedField={null}
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      expect(screen.getByLabelText('Client ID')).toBeInTheDocument();
      expect(screen.getByDisplayValue('client-123')).toBeInTheDocument();
    });

    it('should not render client ID field when OAuth2 config is not provided', () => {
      render(
        <QuickCopySection application={mockApplication} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />,
      );

      expect(screen.queryByLabelText('Client ID')).not.toBeInTheDocument();
    });

    it('should render both copy buttons', () => {
      render(
        <QuickCopySection
          application={mockApplication}
          oauth2Config={mockOAuth2Config}
          copiedField={null}
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      const copyButtons = screen.getAllByRole('button');
      expect(copyButtons).toHaveLength(2);
    });
  });

  describe('Copy Functionality', () => {
    it('should call onCopyToClipboard when application ID copy button is clicked', async () => {
      const user = userEvent.setup();
      render(
        <QuickCopySection application={mockApplication} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />,
      );

      const copyButtons = screen.getAllByRole('button');
      await user.click(copyButtons[0]);

      expect(mockOnCopyToClipboard).toHaveBeenCalledWith('app-123', 'app_id');
    });

    it('should call onCopyToClipboard when client ID copy button is clicked', async () => {
      const user = userEvent.setup();
      render(
        <QuickCopySection
          application={mockApplication}
          oauth2Config={mockOAuth2Config}
          copiedField={null}
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      const copyButtons = screen.getAllByRole('button');
      await user.click(copyButtons[1]);

      expect(mockOnCopyToClipboard).toHaveBeenCalledWith('client-123', 'client_id');
    });

    it('should not render client ID copy button when client ID is not available', () => {
      render(
        <QuickCopySection application={mockApplication} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />,
      );

      const copyButtons = screen.getAllByRole('button');
      expect(copyButtons).toHaveLength(1);
    });

    it('should handle copy errors gracefully', async () => {
      const user = userEvent.setup();
      mockOnCopyToClipboard.mockRejectedValue(new Error('Copy failed'));

      render(
        <QuickCopySection application={mockApplication} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />,
      );

      const copyButtons = screen.getAllByRole('button');
      await user.click(copyButtons[0]);

      expect(mockOnCopyToClipboard).toHaveBeenCalledWith('app-123', 'app_id');
    });
  });

  describe('Visual Feedback', () => {
    it('should show check icon for application ID when it is copied', () => {
      render(
        <QuickCopySection
          application={mockApplication}
          copiedField="app_id"
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      // Should show "Copied!" tooltip for app_id field
      expect(screen.getByLabelText('Copied!')).toBeInTheDocument();
    });

    it('should show check icon for client ID when it is copied', () => {
      render(
        <QuickCopySection
          application={mockApplication}
          oauth2Config={mockOAuth2Config}
          copiedField="client_id"
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      // Should show "Copied!" tooltip for client_id field
      expect(screen.getByLabelText('Copied!')).toBeInTheDocument();
    });

    it('should show copy icon when nothing is copied', () => {
      render(
        <QuickCopySection
          application={mockApplication}
          oauth2Config={mockOAuth2Config}
          copiedField={null}
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      // Both fields should show "Copy" tooltip
      const copyButtons = screen.getAllByLabelText('Copy');
      expect(copyButtons).toHaveLength(2);
    });

    it('should show copy icon for application ID when client ID is copied', () => {
      render(
        <QuickCopySection
          application={mockApplication}
          oauth2Config={mockOAuth2Config}
          copiedField="client_id"
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      // Should have one "Copy" button and one "Copied!" button
      expect(screen.getByLabelText('Copy')).toBeInTheDocument();
      expect(screen.getByLabelText('Copied!')).toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('should have proper labels for form controls', () => {
      render(
        <QuickCopySection
          application={mockApplication}
          oauth2Config={mockOAuth2Config}
          copiedField={null}
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      expect(screen.getByLabelText('Application ID')).toBeInTheDocument();
      expect(screen.getByLabelText('Client ID')).toBeInTheDocument();
    });

    it('should have input IDs for accessibility', () => {
      render(
        <QuickCopySection
          application={mockApplication}
          oauth2Config={mockOAuth2Config}
          copiedField={null}
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      expect(document.getElementById('application-id-input')).toBeInTheDocument();
      expect(document.getElementById('client-id-input')).toBeInTheDocument();
    });

    it('should display helper text for inputs', () => {
      render(
        <QuickCopySection
          application={mockApplication}
          oauth2Config={mockOAuth2Config}
          copiedField={null}
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      expect(screen.getByText('Unique identifier for your application')).toBeInTheDocument();
      expect(screen.getByText('OAuth2 client identifier used for authentication')).toBeInTheDocument();
    });
  });

  describe('Read-only Behavior', () => {
    it('should render application ID field as read-only', () => {
      render(
        <QuickCopySection application={mockApplication} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />,
      );

      const appIdInput = screen.getByDisplayValue('app-123');
      expect(appIdInput).toHaveAttribute('readonly');
    });

    it('should render client ID field as read-only', () => {
      render(
        <QuickCopySection
          application={mockApplication}
          oauth2Config={mockOAuth2Config}
          copiedField={null}
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      const clientIdInput = screen.getByDisplayValue('client-123');
      expect(clientIdInput).toHaveAttribute('readonly');
    });
  });
});
