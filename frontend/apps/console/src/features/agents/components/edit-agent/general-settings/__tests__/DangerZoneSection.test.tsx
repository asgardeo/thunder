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

import {fireEvent, render, screen} from '@thunderid/test-utils';
import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';
import DangerZoneSection from '../DangerZoneSection';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, fallback?: string) => {
      const translations: Record<string, string> = {
        'applications:edit.general.sections.dangerZone.title': 'Danger Zone',
        'applications:edit.general.sections.dangerZone.description':
          'Actions in this section are irreversible. Proceed with caution.',
        'applications:edit.general.sections.dangerZone.regenerateSecret.title': 'Regenerate Client Secret',
        'applications:edit.general.sections.dangerZone.regenerateSecret.description':
          'Regenerating the client secret will immediately invalidate the current client secret and cannot be undone.',
        'applications:edit.general.sections.dangerZone.regenerateSecret.button': 'Regenerate Client Secret',
        'agents:edit.general.dangerZone.deleteAgent.title': 'Delete Agent',
        'agents:edit.general.dangerZone.deleteAgent.description':
          'Permanently delete this agent and all associated data. This action cannot be undone.',
        'agents:edit.general.dangerZone.deleteAgent.button': 'Delete Agent',
      };
      return translations[key] ?? fallback ?? key;
    },
  }),
}));

describe('DangerZoneSection (agent)', () => {
  const mockOnRegenerateClick = vi.fn();
  const mockOnDeleteClick = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('renders the danger zone section', () => {
    render(<DangerZoneSection onDeleteClick={mockOnDeleteClick} />);

    expect(screen.getByText('Danger Zone')).toBeInTheDocument();
    expect(screen.getByText('Actions in this section are irreversible. Proceed with caution.')).toBeInTheDocument();
  });

  it('always renders the delete agent section', () => {
    render(<DangerZoneSection onDeleteClick={mockOnDeleteClick} />);

    expect(screen.getByRole('heading', {name: 'Delete Agent', level: 6})).toBeInTheDocument();
    expect(
      screen.getByText('Permanently delete this agent and all associated data. This action cannot be undone.'),
    ).toBeInTheDocument();
  });

  it('renders delete button via the data-testid', () => {
    render(<DangerZoneSection onDeleteClick={mockOnDeleteClick} />);

    expect(screen.getByTestId('delete-agent-button')).toBeInTheDocument();
  });

  it('calls onDeleteClick when delete button is clicked', () => {
    render(<DangerZoneSection onDeleteClick={mockOnDeleteClick} />);

    fireEvent.click(screen.getByTestId('delete-agent-button'));

    expect(mockOnDeleteClick).toHaveBeenCalledTimes(1);
  });

  it('renders delete button with error color', () => {
    render(<DangerZoneSection onDeleteClick={mockOnDeleteClick} />);

    expect(screen.getByTestId('delete-agent-button')).toHaveClass('MuiButton-colorError');
  });

  it('does not render the regenerate-secret section by default', () => {
    render(<DangerZoneSection onDeleteClick={mockOnDeleteClick} />);

    expect(screen.queryByRole('button', {name: 'Regenerate Client Secret'})).not.toBeInTheDocument();
  });

  it('renders the regenerate-secret section when showRegenerateSecret is true', () => {
    render(
      <DangerZoneSection
        showRegenerateSecret
        onRegenerateClick={mockOnRegenerateClick}
        onDeleteClick={mockOnDeleteClick}
      />,
    );

    expect(screen.getByRole('heading', {name: 'Regenerate Client Secret', level: 6})).toBeInTheDocument();
    expect(screen.getByRole('button', {name: 'Regenerate Client Secret'})).toBeInTheDocument();
  });

  it('calls onRegenerateClick when the regenerate button is clicked', () => {
    render(
      <DangerZoneSection
        showRegenerateSecret
        onRegenerateClick={mockOnRegenerateClick}
        onDeleteClick={mockOnDeleteClick}
      />,
    );

    fireEvent.click(screen.getByRole('button', {name: 'Regenerate Client Secret'}));

    expect(mockOnRegenerateClick).toHaveBeenCalledTimes(1);
  });
});
