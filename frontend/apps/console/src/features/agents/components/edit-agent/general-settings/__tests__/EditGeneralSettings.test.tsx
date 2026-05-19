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

import userEvent from '@testing-library/user-event';
import {render, screen} from '@thunderid/test-utils';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import type {Agent, OAuthAgentConfig} from '../../../../models/agent';
import EditGeneralSettings from '../EditGeneralSettings';

// Mock the child sections / dialogs so we test wiring only.
vi.mock('../QuickCopySection', () => ({
  default: () => <div data-testid="quick-copy">QuickCopy</div>,
}));

vi.mock('../OrganizationUnitSection', () => ({
  default: () => <div data-testid="organization-unit">OU</div>,
}));

vi.mock('../DangerZoneSection', () => ({
  default: ({
    showRegenerateSecret,
    onRegenerateClick,
    onDeleteClick,
  }: {
    showRegenerateSecret?: boolean;
    onRegenerateClick?: () => void;
    onDeleteClick: () => void;
  }) => (
    <div data-testid="danger-zone">
      <span data-testid="show-regenerate">{String(Boolean(showRegenerateSecret))}</span>
      {showRegenerateSecret && onRegenerateClick && (
        <button type="button" onClick={onRegenerateClick}>
          Regenerate
        </button>
      )}
      <button type="button" onClick={onDeleteClick}>
        Delete
      </button>
    </div>
  ),
}));

vi.mock('../../../AgentDeleteDialog', () => ({
  default: ({open, onClose, onSuccess}: {open: boolean; onClose: () => void; onSuccess?: () => void}) =>
    open ? (
      <div role="dialog" data-testid="delete-dialog">
        <button type="button" onClick={onClose}>
          Close Delete
        </button>
        {onSuccess && (
          <button type="button" onClick={onSuccess}>
            Trigger Success
          </button>
        )}
      </div>
    ) : null,
}));

vi.mock('../../../RegenerateSecretDialog', () => ({
  default: ({open, onClose, onSuccess}: {open: boolean; onClose: () => void; onSuccess?: (secret: string) => void}) =>
    open ? (
      <div role="dialog" data-testid="regenerate-dialog">
        <button type="button" onClick={onClose}>
          Close Regenerate
        </button>
        <button type="button" onClick={() => onSuccess?.('new-secret-value')}>
          Trigger Regenerate Success
        </button>
      </div>
    ) : null,
}));

vi.mock('../../../ClientSecretSuccessDialog', () => ({
  default: ({open, clientSecret, onClose}: {open: boolean; clientSecret: string; onClose: () => void}) =>
    open ? (
      <div role="dialog" data-testid="secret-success-dialog">
        <span data-testid="new-secret">{clientSecret}</span>
        <button type="button" onClick={onClose}>
          Close Success
        </button>
      </div>
    ) : null,
}));

describe('EditGeneralSettings', () => {
  const mockAgent: Agent = {
    id: 'agent-1',
    ouId: 'ou-1',
    type: 'default',
    name: 'Test Agent',
  };

  const confidentialOAuthConfig: OAuthAgentConfig = {
    grantTypes: ['client_credentials'],
    responseTypes: [],
    tokenEndpointAuthMethod: 'client_secret_basic',
  };

  const publicOAuthConfig: OAuthAgentConfig = {
    grantTypes: ['authorization_code'],
    responseTypes: ['code'],
    tokenEndpointAuthMethod: 'none',
  };

  const mockOnCopy = vi.fn().mockResolvedValue(undefined);

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders QuickCopy, OU and DangerZone sections', () => {
    render(<EditGeneralSettings agent={mockAgent} copiedField={null} onCopyToClipboard={mockOnCopy} />);

    expect(screen.getByTestId('quick-copy')).toBeInTheDocument();
    expect(screen.getByTestId('organization-unit')).toBeInTheDocument();
    expect(screen.getByTestId('danger-zone')).toBeInTheDocument();
  });

  it('hides regenerate option for public clients', () => {
    render(
      <EditGeneralSettings
        agent={mockAgent}
        oauth2Config={publicOAuthConfig}
        copiedField={null}
        onCopyToClipboard={mockOnCopy}
      />,
    );

    expect(screen.getByTestId('show-regenerate')).toHaveTextContent('false');
  });

  it('shows regenerate option for confidential clients', () => {
    render(
      <EditGeneralSettings
        agent={mockAgent}
        oauth2Config={confidentialOAuthConfig}
        copiedField={null}
        onCopyToClipboard={mockOnCopy}
      />,
    );

    expect(screen.getByTestId('show-regenerate')).toHaveTextContent('true');
  });

  it('opens the delete dialog when Delete is clicked', async () => {
    const user = userEvent.setup();
    render(<EditGeneralSettings agent={mockAgent} copiedField={null} onCopyToClipboard={mockOnCopy} />);

    expect(screen.queryByTestId('delete-dialog')).not.toBeInTheDocument();
    await user.click(screen.getByText('Delete'));
    expect(screen.getByTestId('delete-dialog')).toBeInTheDocument();
  });

  it('forwards onDeleteSuccess from the delete dialog', async () => {
    const onDeleteSuccess = vi.fn();
    const user = userEvent.setup();
    render(
      <EditGeneralSettings
        agent={mockAgent}
        copiedField={null}
        onCopyToClipboard={mockOnCopy}
        onDeleteSuccess={onDeleteSuccess}
      />,
    );

    await user.click(screen.getByText('Delete'));
    await user.click(screen.getByText('Trigger Success'));

    expect(onDeleteSuccess).toHaveBeenCalledTimes(1);
  });

  it('opens the regenerate dialog and shows the success dialog with the new secret', async () => {
    const user = userEvent.setup();
    render(
      <EditGeneralSettings
        agent={mockAgent}
        oauth2Config={confidentialOAuthConfig}
        copiedField={null}
        onCopyToClipboard={mockOnCopy}
      />,
    );

    await user.click(screen.getByText('Regenerate'));
    expect(screen.getByTestId('regenerate-dialog')).toBeInTheDocument();

    await user.click(screen.getByText('Trigger Regenerate Success'));

    expect(screen.getByTestId('secret-success-dialog')).toBeInTheDocument();
    expect(screen.getByTestId('new-secret')).toHaveTextContent('new-secret-value');
  });

  it('clears the new client secret when the success dialog is closed', async () => {
    const user = userEvent.setup();
    render(
      <EditGeneralSettings
        agent={mockAgent}
        oauth2Config={confidentialOAuthConfig}
        copiedField={null}
        onCopyToClipboard={mockOnCopy}
      />,
    );

    await user.click(screen.getByText('Regenerate'));
    await user.click(screen.getByText('Trigger Regenerate Success'));

    expect(screen.getByTestId('secret-success-dialog')).toBeInTheDocument();

    await user.click(screen.getByText('Close Success'));

    expect(screen.queryByTestId('secret-success-dialog')).not.toBeInTheDocument();
  });
});
