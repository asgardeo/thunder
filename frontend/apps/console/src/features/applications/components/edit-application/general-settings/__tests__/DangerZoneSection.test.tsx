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

import {screen, fireEvent, renderWithProviders} from '@thunderid/test-utils';
import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';
import DangerZoneSection from '../DangerZoneSection';

// Mock translations
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'applications:edit.general.sections.dangerZone.title': 'Danger Zone',
        'applications:edit.general.sections.dangerZone.description':
          'Actions in this section are irreversible. Proceed with caution.',
        'applications:edit.general.sections.dangerZone.regenerateSecret.title': 'Regenerate Client Secret',
        'applications:edit.general.sections.dangerZone.regenerateSecret.description':
          'Regenerating the client secret will immediately invalidate the current client secret and cannot be undone.',
        'applications:edit.general.sections.dangerZone.regenerateSecret.button': 'Regenerate Client Secret',
        'applications:edit.general.sections.dangerZone.deleteApplication.title': 'Delete Application',
        'applications:edit.general.sections.dangerZone.deleteApplication.description':
          'Permanently delete this application and all associated data. This action cannot be undone.',
        'applications:edit.general.sections.dangerZone.deleteApplication.button': 'Delete Application',
      };
      return translations[key] ?? key;
    },
  }),
}));

describe('DangerZoneSection', () => {
  const mockOnRegenerateClick = vi.fn();
  const mockOnDeleteClick = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should render the danger zone section', () => {
    renderWithProviders(<DangerZoneSection onDeleteClick={mockOnDeleteClick} />);

    expect(screen.getByText('Danger Zone')).toBeInTheDocument();
    expect(screen.getByText('Actions in this section are irreversible. Proceed with caution.')).toBeInTheDocument();
  });

  it('should always render delete application section', () => {
    renderWithProviders(<DangerZoneSection onDeleteClick={mockOnDeleteClick} />);

    expect(screen.getByRole('heading', {name: 'Delete Application', level: 6})).toBeInTheDocument();
    expect(
      screen.getByText('Permanently delete this application and all associated data. This action cannot be undone.'),
    ).toBeInTheDocument();
  });

  it('should render delete button', () => {
    renderWithProviders(<DangerZoneSection onDeleteClick={mockOnDeleteClick} />);

    const deleteButton = screen.getByRole('button', {name: 'Delete Application'});
    expect(deleteButton).toBeInTheDocument();
  });

  it('should call onDeleteClick when delete button is clicked', () => {
    renderWithProviders(<DangerZoneSection onDeleteClick={mockOnDeleteClick} />);

    const deleteButton = screen.getByRole('button', {name: 'Delete Application'});
    fireEvent.click(deleteButton);

    expect(mockOnDeleteClick).toHaveBeenCalledTimes(1);
  });

  it('should render delete button with error color', () => {
    renderWithProviders(<DangerZoneSection onDeleteClick={mockOnDeleteClick} />);

    const deleteButton = screen.getByRole('button', {name: 'Delete Application'});
    expect(deleteButton).toHaveClass('MuiButton-colorError');
  });

  it('should not render regenerate secret section by default', () => {
    renderWithProviders(<DangerZoneSection onDeleteClick={mockOnDeleteClick} />);

    expect(screen.queryByRole('button', {name: 'Regenerate Client Secret'})).not.toBeInTheDocument();
  });

  it('should render regenerate secret section when showRegenerateSecret is true', () => {
    renderWithProviders(
      <DangerZoneSection
        showRegenerateSecret
        onRegenerateClick={mockOnRegenerateClick}
        onDeleteClick={mockOnDeleteClick}
      />,
    );

    expect(screen.getByRole('heading', {name: 'Regenerate Client Secret', level: 6})).toBeInTheDocument();
    expect(
      screen.getByText(
        'Regenerating the client secret will immediately invalidate the current client secret and cannot be undone.',
      ),
    ).toBeInTheDocument();
    expect(screen.getByRole('button', {name: 'Regenerate Client Secret'})).toBeInTheDocument();
  });

  it('should call onRegenerateClick when regenerate button is clicked', () => {
    renderWithProviders(
      <DangerZoneSection
        showRegenerateSecret
        onRegenerateClick={mockOnRegenerateClick}
        onDeleteClick={mockOnDeleteClick}
      />,
    );

    const regenerateButton = screen.getByRole('button', {name: 'Regenerate Client Secret'});
    fireEvent.click(regenerateButton);

    expect(mockOnRegenerateClick).toHaveBeenCalledTimes(1);
  });

  it('should render both sections with a divider when showRegenerateSecret is true', () => {
    renderWithProviders(
      <DangerZoneSection
        showRegenerateSecret
        onRegenerateClick={mockOnRegenerateClick}
        onDeleteClick={mockOnDeleteClick}
      />,
    );

    expect(screen.getByRole('button', {name: 'Regenerate Client Secret'})).toBeInTheDocument();
    expect(screen.getByRole('button', {name: 'Delete Application'})).toBeInTheDocument();
  });
});
