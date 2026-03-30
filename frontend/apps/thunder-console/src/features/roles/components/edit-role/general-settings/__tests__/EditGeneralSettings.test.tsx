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

import {render, screen, waitFor, fireEvent} from '@thunder/test-utils';
import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';
import type {Role} from '../../../../models/role';
import EditGeneralSettings from '../EditGeneralSettings';

// Mock SettingsCard
vi.mock('@/components/SettingsCard', () => ({
  default: ({title, description, children}: {title: string; description: string; children: React.ReactNode}) => (
    <div data-testid="settings-card">
      <h3>{title}</h3>
      <p>{description}</p>
      {children}
    </div>
  ),
}));

// Mock translations
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'roles:edit.general.sections.organizationUnit.title': 'Organization Unit',
        'roles:edit.general.sections.organizationUnit.description':
          'The organization unit this role belongs to.',
        'roles:edit.general.sections.organizationUnit.copyId': 'Copy Organization Unit ID',
        'roles:edit.general.sections.dangerZone.title': 'Danger Zone',
        'roles:edit.general.sections.dangerZone.description':
          'Actions in this section are irreversible. Proceed with caution.',
        'roles:edit.general.sections.dangerZone.deleteRole': 'Delete this role',
        'roles:edit.general.sections.dangerZone.deleteRoleDescription':
          'Deleting this role is permanent and cannot be undone.',
        'common:actions.delete': 'Delete',
        'common:actions.copied': 'Copied',
      };
      return translations[key] || key;
    },
  }),
}));

describe('EditGeneralSettings', () => {
  const mockOnDeleteClick = vi.fn();
  const mockWriteText = vi.fn().mockResolvedValue(undefined);

  const mockRole: Role = {
    id: 'role-1',
    name: 'Admin Role',
    description: 'Administrator role',
    ouId: 'ou-test-123',
    permissions: [],
  };

  const defaultProps = {
    role: mockRole,
    onDeleteClick: mockOnDeleteClick,
  };

  beforeEach(() => {
    Object.defineProperty(navigator, 'clipboard', {
      value: {writeText: mockWriteText},
      writable: true,
      configurable: true,
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should render Organization Unit section', () => {
    render(<EditGeneralSettings {...defaultProps} />);

    expect(screen.getByRole('heading', {name: 'Organization Unit'})).toBeInTheDocument();
  });

  it('should render Danger Zone section', () => {
    render(<EditGeneralSettings {...defaultProps} />);

    expect(screen.getByText('Danger Zone')).toBeInTheDocument();
  });

  it('should display role ouId in readonly text field', () => {
    render(<EditGeneralSettings {...defaultProps} />);

    const input = screen.getByDisplayValue('ou-test-123');
    expect(input).toBeInTheDocument();
  });

  it('should call onDeleteClick when delete button is clicked', () => {
    render(<EditGeneralSettings {...defaultProps} />);

    const deleteButton = screen.getByRole('button', {name: 'Delete'});
    fireEvent.click(deleteButton);

    expect(mockOnDeleteClick).toHaveBeenCalledTimes(1);
  });

  it('should copy ouId to clipboard when copy button is clicked', async () => {
    render(<EditGeneralSettings {...defaultProps} />);

    const copyButton = screen.getByRole('button', {name: 'Copy Organization Unit ID'});
    fireEvent.click(copyButton);

    await waitFor(() => {
      expect(mockWriteText).toHaveBeenCalledWith('ou-test-123');
    });
  });

  it('should render delete role description', () => {
    render(<EditGeneralSettings {...defaultProps} />);

    expect(screen.getByText('Deleting this role is permanent and cannot be undone.')).toBeInTheDocument();
  });
});
