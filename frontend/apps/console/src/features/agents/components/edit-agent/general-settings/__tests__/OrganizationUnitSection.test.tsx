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
import type {Agent} from '../../../../models/agent';
import OrganizationUnitSection from '../OrganizationUnitSection';

vi.mock('@thunderid/components', () => ({
  SettingsCard: ({title, description, children}: {title: string; description: string; children: React.ReactNode}) => (
    <div>
      <div data-testid="card-title">{title}</div>
      <div data-testid="card-description">{description}</div>
      {children}
    </div>
  ),
}));

describe('OrganizationUnitSection', () => {
  const mockOnCopyToClipboard = vi.fn();

  const baseAgent: Agent = {
    id: 'agent-1',
    ouId: 'ou-uuid-123',
    ouHandle: 'engineering',
    type: 'default',
    name: 'Test',
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockOnCopyToClipboard.mockResolvedValue(undefined);
  });

  it('renders the OU handle and ID values', () => {
    render(<OrganizationUnitSection agent={baseAgent} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />);

    expect(screen.getByDisplayValue('engineering')).toBeInTheDocument();
    expect(screen.getByDisplayValue('ou-uuid-123')).toBeInTheDocument();
  });

  it('shows "-" placeholder when ouHandle is missing', () => {
    render(
      <OrganizationUnitSection
        agent={{...baseAgent, ouHandle: undefined}}
        copiedField={null}
        onCopyToClipboard={mockOnCopyToClipboard}
      />,
    );

    expect(screen.getByDisplayValue('-')).toBeInTheDocument();
  });

  it('does not render a copy button next to the handle when ouHandle is missing', () => {
    render(
      <OrganizationUnitSection
        agent={{...baseAgent, ouHandle: undefined}}
        copiedField={null}
        onCopyToClipboard={mockOnCopyToClipboard}
      />,
    );

    // Only one copy button should render — for the OU ID
    const buttons = screen.getAllByRole('button');
    expect(buttons).toHaveLength(1);
  });

  it('copies the OU handle when its copy button is clicked', async () => {
    const user = userEvent.setup();
    render(<OrganizationUnitSection agent={baseAgent} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />);

    const handleInput = screen.getByDisplayValue('engineering');
    const button = handleInput.closest('.MuiInputBase-root')?.querySelector('button');
    await user.click(button as HTMLElement);

    expect(mockOnCopyToClipboard).toHaveBeenCalledWith('engineering', 'ouHandle');
  });

  it('copies the OU ID when its copy button is clicked', async () => {
    const user = userEvent.setup();
    render(<OrganizationUnitSection agent={baseAgent} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />);

    const idInput = screen.getByDisplayValue('ou-uuid-123');
    const button = idInput.closest('.MuiInputBase-root')?.querySelector('button');
    await user.click(button as HTMLElement);

    expect(mockOnCopyToClipboard).toHaveBeenCalledWith('ou-uuid-123', 'ouId');
  });

  it('renders the input fields as readonly', () => {
    render(<OrganizationUnitSection agent={baseAgent} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />);

    expect(screen.getByDisplayValue('engineering')).toHaveAttribute('readonly');
    expect(screen.getByDisplayValue('ou-uuid-123')).toHaveAttribute('readonly');
  });

  it('shows copied state for ouHandle when copiedField === "ouHandle"', () => {
    render(
      <OrganizationUnitSection agent={baseAgent} copiedField="ouHandle" onCopyToClipboard={mockOnCopyToClipboard} />,
    );

    // Tooltip / aria-label text changes to "Copied"
    expect(screen.getByLabelText(/copied/i)).toBeInTheDocument();
  });

  it('shows copied state for ouId when copiedField === "ouId"', () => {
    render(<OrganizationUnitSection agent={baseAgent} copiedField="ouId" onCopyToClipboard={mockOnCopyToClipboard} />);

    expect(screen.getByLabelText(/copied/i)).toBeInTheDocument();
  });

  it('handles copy errors gracefully', async () => {
    const user = userEvent.setup();
    mockOnCopyToClipboard.mockRejectedValueOnce(new Error('Copy failed'));

    render(<OrganizationUnitSection agent={baseAgent} copiedField={null} onCopyToClipboard={mockOnCopyToClipboard} />);

    const idInput = screen.getByDisplayValue('ou-uuid-123');
    const button = idInput.closest('.MuiInputBase-root')?.querySelector('button');
    await user.click(button as HTMLElement);

    expect(mockOnCopyToClipboard).toHaveBeenCalled();
  });
});
