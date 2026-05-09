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

import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import type {Agent, OAuthAgentConfig} from '../../../../models/agent';
import QuickCopySection from '../QuickCopySection';

// Mock the SettingsCard wrapper so DOM is easy to query.
vi.mock('@thunderid/components', () => ({
  SettingsCard: ({title, description, children}: {title: string; description: string; children: React.ReactNode}) => (
    <div data-testid="settings-card">
      <div data-testid="card-title">{title}</div>
      <div data-testid="card-description">{description}</div>
      {children}
    </div>
  ),
}));

describe('QuickCopySection (agent)', () => {
  const mockOnCopyToClipboard = vi.fn();

  const mockAgent: Agent = {
    id: 'agent-123',
    ouId: 'ou-1',
    type: 'default',
    name: 'Test Agent',
    owner: 'user-1',
  };

  const mockOAuth2Config: OAuthAgentConfig = {
    clientId: 'client-123',
  } as OAuthAgentConfig;

  beforeEach(() => {
    vi.clearAllMocks();
    mockOnCopyToClipboard.mockResolvedValue(undefined);
  });

  describe('Rendering', () => {
    it('renders the settings card', () => {
      render(<QuickCopySection agent={mockAgent} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />);

      expect(screen.getByTestId('card-title')).toBeInTheDocument();
      expect(screen.getByTestId('card-description')).toBeInTheDocument();
    });

    it('renders the agent ID field', () => {
      render(<QuickCopySection agent={mockAgent} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />);

      expect(screen.getByDisplayValue('agent-123')).toBeInTheDocument();
    });

    it('renders the client ID field when oauth2Config is provided', () => {
      render(
        <QuickCopySection
          agent={mockAgent}
          oauth2Config={mockOAuth2Config}
          copiedField={null}
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      expect(screen.getByDisplayValue('client-123')).toBeInTheDocument();
    });

    it('does not render the client ID field when oauth2Config is missing', () => {
      render(<QuickCopySection agent={mockAgent} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />);

      expect(screen.queryByDisplayValue('client-123')).not.toBeInTheDocument();
    });

    it('renders the owner ID field when agent has an owner', () => {
      render(<QuickCopySection agent={mockAgent} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />);

      expect(screen.getByDisplayValue('user-1')).toBeInTheDocument();
    });

    it('does not render the owner ID field when agent has no owner', () => {
      render(
        <QuickCopySection
          agent={{...mockAgent, owner: undefined}}
          copiedField={null}
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      expect(screen.queryByDisplayValue('user-1')).not.toBeInTheDocument();
    });
  });

  describe('Copy Functionality', () => {
    it('copies the agent ID', async () => {
      const user = userEvent.setup();
      render(<QuickCopySection agent={mockAgent} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />);

      const buttons = screen.getAllByRole('button');
      await user.click(buttons[0]);

      expect(mockOnCopyToClipboard).toHaveBeenCalledWith('agent-123', 'agent_id');
    });

    it('copies the client ID', async () => {
      const user = userEvent.setup();
      render(
        <QuickCopySection
          agent={mockAgent}
          oauth2Config={mockOAuth2Config}
          copiedField={null}
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      // Find the button next to the client ID input
      const clientIdInput = screen.getByDisplayValue('client-123');
      const button = clientIdInput.closest('.MuiInputBase-root')?.querySelector('button');
      expect(button).toBeTruthy();
      await user.click(button as HTMLElement);

      expect(mockOnCopyToClipboard).toHaveBeenCalledWith('client-123', 'clientId');
    });

    it('copies the owner ID', async () => {
      const user = userEvent.setup();
      render(<QuickCopySection agent={mockAgent} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />);

      const ownerInput = screen.getByDisplayValue('user-1');
      const button = ownerInput.closest('.MuiInputBase-root')?.querySelector('button');
      await user.click(button as HTMLElement);

      expect(mockOnCopyToClipboard).toHaveBeenCalledWith('user-1', 'owner');
    });

    it('handles copy errors gracefully', async () => {
      const user = userEvent.setup();
      mockOnCopyToClipboard.mockRejectedValueOnce(new Error('Copy failed'));

      render(<QuickCopySection agent={mockAgent} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />);

      const buttons = screen.getAllByRole('button');
      await user.click(buttons[0]);

      expect(mockOnCopyToClipboard).toHaveBeenCalled();
    });
  });

  describe('Visual Feedback', () => {
    it('shows the copied state for agent ID when copiedField === "agent_id"', () => {
      render(<QuickCopySection agent={mockAgent} copiedField="agent_id" onCopyToClipboard={mockOnCopyToClipboard} />);

      // Copy buttons render different icons; check via tooltip text
      expect(screen.getByLabelText(/copied/i)).toBeInTheDocument();
    });

    it('shows the copied state for client ID when copiedField === "clientId"', () => {
      render(
        <QuickCopySection
          agent={mockAgent}
          oauth2Config={mockOAuth2Config}
          copiedField="clientId"
          onCopyToClipboard={mockOnCopyToClipboard}
        />,
      );

      expect(screen.getByLabelText(/copied/i)).toBeInTheDocument();
    });
  });

  describe('Read-only Behavior', () => {
    it('renders agent ID as read-only', () => {
      render(<QuickCopySection agent={mockAgent} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />);

      expect(screen.getByDisplayValue('agent-123')).toHaveAttribute('readonly');
    });
  });
});
