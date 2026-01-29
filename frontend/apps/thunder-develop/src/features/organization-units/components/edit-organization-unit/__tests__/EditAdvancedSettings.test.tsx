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

import {describe, it, expect, vi} from 'vitest';
import {screen, fireEvent} from '@testing-library/react';
import {renderWithProviders} from '../../../../../test/test-utils';
import EditAdvancedSettings from '../advanced-settings/EditAdvancedSettings';

// Mock translations
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'organizationUnits:view.advanced.dangerZone': 'Danger Zone',
        'organizationUnits:view.advanced.dangerZoneDescription':
          'Actions here are irreversible. Please proceed with caution.',
        'organizationUnits:view.advanced.deleteButton': 'Delete Organization Unit',
      };
      return translations[key] ?? key;
    },
  }),
}));

describe('EditAdvancedSettings', () => {
  it('should render danger zone title', () => {
    const onDeleteClick = vi.fn();
    renderWithProviders(<EditAdvancedSettings onDeleteClick={onDeleteClick} />);

    expect(screen.getByText('Danger Zone')).toBeInTheDocument();
  });

  it('should render danger zone description', () => {
    const onDeleteClick = vi.fn();
    renderWithProviders(<EditAdvancedSettings onDeleteClick={onDeleteClick} />);

    expect(screen.getByText('Actions here are irreversible. Please proceed with caution.')).toBeInTheDocument();
  });

  it('should render delete button', () => {
    const onDeleteClick = vi.fn();
    renderWithProviders(<EditAdvancedSettings onDeleteClick={onDeleteClick} />);

    expect(screen.getByText('Delete Organization Unit')).toBeInTheDocument();
  });

  it('should call onDeleteClick when delete button is clicked', () => {
    const onDeleteClick = vi.fn();
    renderWithProviders(<EditAdvancedSettings onDeleteClick={onDeleteClick} />);

    fireEvent.click(screen.getByText('Delete Organization Unit'));

    expect(onDeleteClick).toHaveBeenCalledTimes(1);
  });

  it('should render delete button with error color variant', () => {
    const onDeleteClick = vi.fn();
    renderWithProviders(<EditAdvancedSettings onDeleteClick={onDeleteClick} />);

    const deleteButton = screen.getByText('Delete Organization Unit');
    expect(deleteButton).toBeInTheDocument();
    // The button should have error color styling
    expect(deleteButton.closest('button')).toHaveClass('MuiButton-outlinedError');
  });
});
