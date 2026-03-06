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
import {screen, fireEvent, renderWithProviders} from '@thunder/test-utils';
import DangerZoneSection from '../DangerZoneSection';

// Mock translations
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'applications:edit.general.sections.dangerZone.title': 'Danger Zone',
        'applications:edit.general.sections.dangerZone.description':
          'Irreversible and destructive actions for this application',
        'applications:edit.general.sections.dangerZone.regenerateSecret.title': 'Regenerate Client Secret',
        'applications:edit.general.sections.dangerZone.regenerateSecret.description':
          'Regenerating the client secret will immediately invalidate the current client secret and generate a new one. All active access tokens will be revoked and the application will stop working until the new client secret is updated in your application configuration.',
        'applications:edit.general.sections.dangerZone.regenerateSecret.button': 'Regenerate Client Secret',
      };
      return translations[key] ?? key;
    },
  }),
}));

describe('DangerZoneSection', () => {
  const mockOnRegenerateClick = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should render the danger zone section', () => {
    renderWithProviders(<DangerZoneSection onRegenerateClick={mockOnRegenerateClick} />);

    expect(screen.getByText('Danger Zone')).toBeInTheDocument();
    expect(screen.getByText('Irreversible and destructive actions for this application')).toBeInTheDocument();
  });

  it('should render revoke application title', () => {
    renderWithProviders(<DangerZoneSection onRegenerateClick={mockOnRegenerateClick} />);

    const heading = screen.getByRole('heading', {name: 'Regenerate Client Secret', level: 6});
    expect(heading).toBeInTheDocument();
  });

  it('should render warning description', () => {
    renderWithProviders(<DangerZoneSection onRegenerateClick={mockOnRegenerateClick} />);

    expect(
      screen.getByText(
        'Regenerating the client secret will immediately invalidate the current client secret and generate a new one. All active access tokens will be revoked and the application will stop working until the new client secret is updated in your application configuration.',
      ),
    ).toBeInTheDocument();
  });

  it('should render revoke button', () => {
    renderWithProviders(<DangerZoneSection onRegenerateClick={mockOnRegenerateClick} />);

    const regenerateButton = screen.getByRole('button', {name: 'Regenerate Client Secret'});
    expect(regenerateButton).toBeInTheDocument();
  });

  it('should call onRegenerateClick when revoke button is clicked', () => {
    renderWithProviders(<DangerZoneSection onRegenerateClick={mockOnRegenerateClick} />);

    const regenerateButton = screen.getByRole('button', {name: 'Regenerate Client Secret'});
    fireEvent.click(regenerateButton);

    expect(mockOnRegenerateClick).toHaveBeenCalledTimes(1);
  });

  it('should call onRegenerateClick multiple times when clicked multiple times', () => {
    renderWithProviders(<DangerZoneSection onRegenerateClick={mockOnRegenerateClick} />);

    const regenerateButton = screen.getByRole('button', {name: 'Regenerate Client Secret'});
    fireEvent.click(regenerateButton);
    fireEvent.click(regenerateButton);
    fireEvent.click(regenerateButton);

    expect(mockOnRegenerateClick).toHaveBeenCalledTimes(3);
  });

  it('should render revoke button with error color', () => {
    renderWithProviders(<DangerZoneSection onRegenerateClick={mockOnRegenerateClick} />);

    const regenerateButton = screen.getByRole('button', {name: 'Regenerate Client Secret'});
    expect(regenerateButton).toHaveClass('MuiButton-colorError');
  });
});
