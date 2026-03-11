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
    it('should render the settings card with title and description', async () => {
      await render(
        <QuickCopySection application={mockApplication} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />,
      );

      await expect.element(page.getByTestId('card-title')).toHaveTextContent('Quick Copy');
      await expect.element(page.getByTestId('card-description')).toHaveTextContent(
        'Copy application identifiers for use in your code.',
      );
    });

    it('should render application ID field', async () => {
      await render(
        <QuickCopySection application={mockApplication} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />,
      );

      await expect.element(page.getByLabelText('Application ID')).toBeInTheDocument();
      expect(getByDisplayValue('app-123')).toBeInTheDocument();
    });

    it('should render client ID field when OAuth2 config is provided', async () => {
      await render(
        <QuickCopySection
          application={mockApplication}
          oauth2Config={mockOAuth2Config}
          copiedField={null}
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      await expect.element(page.getByLabelText('Client ID')).toBeInTheDocument();
      expect(getByDisplayValue('client-123')).toBeInTheDocument();
    });

    it('should render empty client ID field when OAuth2 config is not provided', async () => {
      await render(
        <QuickCopySection application={mockApplication} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />,
      );

      const clientIdInput = page.getByLabelText('Client ID');
      expect(clientIdInput).toHaveAttribute('value', '');
    });

    it('should render both copy buttons', async () => {
      await render(
        <QuickCopySection
          application={mockApplication}
          oauth2Config={mockOAuth2Config}
          copiedField={null}
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      const copyButtons = page.getByRole('button');
      expect(copyButtons).toHaveLength(2);
    });
  });

  describe('Copy Functionality', () => {
    it('should call onCopyToClipboard when application ID copy button is clicked', async () => {
            await render(
        <QuickCopySection application={mockApplication} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />,
      );

      const copyButtons = page.getByRole('button').all();
      await userEvent.click(copyButtons[0]);

      expect(mockOnCopyToClipboard).toHaveBeenCalledWith('app-123', 'app_id');
    });

    it('should call onCopyToClipboard when client ID copy button is clicked', async () => {
            await render(
        <QuickCopySection
          application={mockApplication}
          oauth2Config={mockOAuth2Config}
          copiedField={null}
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      const copyButtons = page.getByRole('button').all();
      await userEvent.click(copyButtons[1]);

      expect(mockOnCopyToClipboard).toHaveBeenCalledWith('client-123', 'client_id');
    });

    it('should not call onCopyToClipboard when client ID is not available', async () => {
            await render(
        <QuickCopySection application={mockApplication} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />,
      );

      const copyButtons = page.getByRole('button').all();
      await userEvent.click(copyButtons[1]);

      expect(mockOnCopyToClipboard).not.toHaveBeenCalled();
    });

    it('should handle copy errors gracefully', async () => {
            mockOnCopyToClipboard.mockRejectedValue(new Error('Copy failed'));

      await render(
        <QuickCopySection application={mockApplication} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />,
      );

      const copyButtons = page.getByRole('button').all();
      await userEvent.click(copyButtons[0]);

      expect(mockOnCopyToClipboard).toHaveBeenCalledWith('app-123', 'app_id');
    });
  });

  describe('Visual Feedback', () => {
    it('should show check icon for application ID when it is copied', async () => {
      await render(
        <QuickCopySection
          application={mockApplication}
          copiedField="app_id"
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      // Should show "Copied!" tooltip for app_id field
      await expect.element(page.getByLabelText('Copied!')).toBeInTheDocument();
    });

    it('should show check icon for client ID when it is copied', async () => {
      await render(
        <QuickCopySection
          application={mockApplication}
          oauth2Config={mockOAuth2Config}
          copiedField="client_id"
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      // Should show "Copied!" tooltip for client_id field
      await expect.element(page.getByLabelText('Copied!')).toBeInTheDocument();
    });

    it('should show copy icon when nothing is copied', async () => {
      await render(
        <QuickCopySection
          application={mockApplication}
          oauth2Config={mockOAuth2Config}
          copiedField={null}
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      // Both fields should show "Copy" tooltip
      const copyButtons = page.getByLabelText('Copy').all();
      expect(copyButtons).toHaveLength(2);
    });

    it('should show copy icon for application ID when client ID is copied', async () => {
      await render(
        <QuickCopySection
          application={mockApplication}
          oauth2Config={mockOAuth2Config}
          copiedField="client_id"
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      // Should have one "Copy" button and one "Copied!" button
      await expect.element(page.getByLabelText('Copy')).toBeInTheDocument();
      await expect.element(page.getByLabelText('Copied!')).toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('should have proper labels for form controls', async () => {
      await render(
        <QuickCopySection
          application={mockApplication}
          oauth2Config={mockOAuth2Config}
          copiedField={null}
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      await expect.element(page.getByLabelText('Application ID')).toBeInTheDocument();
      await expect.element(page.getByLabelText('Client ID')).toBeInTheDocument();
    });

    it('should have input IDs for accessibility', async () => {
      await render(
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

    it('should display helper text for inputs', async () => {
      await render(
        <QuickCopySection
          application={mockApplication}
          oauth2Config={mockOAuth2Config}
          copiedField={null}
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      await expect.element(page.getByText('Unique identifier for your application')).toBeInTheDocument();
      await expect.element(page.getByText('OAuth2 client identifier used for authentication')).toBeInTheDocument();
    });
  });

  describe('Read-only Behavior', () => {
    it('should render application ID field as read-only', async () => {
      await render(
        <QuickCopySection application={mockApplication} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />,
      );

      const appIdInput = getByDisplayValue('app-123');
      expect(appIdInput).toHaveAttribute('readonly');
    });

    it('should render client ID field as read-only', async () => {
      await render(
        <QuickCopySection
          application={mockApplication}
          oauth2Config={mockOAuth2Config}
          copiedField={null}
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      const clientIdInput = getByDisplayValue('client-123');
      expect(clientIdInput).toHaveAttribute('readonly');
    });
  });
});
