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

import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest';
import {screen, waitFor, fireEvent} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {renderWithProviders} from '@thunder/test-utils';
import EditGeneralSettings from '../edit-group/general-settings/EditGeneralSettings';
import type {Group} from '../../models/group';

describe('EditGeneralSettings', () => {
  const mockGroup: Group = {
    id: 'g1',
    name: 'Test Group',
    description: 'Test desc',
    organizationUnitId: 'ou-123',
  };

  let mockWriteText: ReturnType<typeof vi.fn>;
  const originalClipboard = navigator.clipboard;

  const defaultProps = {
    group: mockGroup,
    onDeleteClick: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockWriteText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, 'clipboard', {
      value: {writeText: mockWriteText},
      writable: true,
      configurable: true,
    });
  });

  afterEach(() => {
    vi.clearAllMocks();
    Object.defineProperty(navigator, 'clipboard', {
      value: originalClipboard,
      writable: true,
      configurable: true,
    });
  });

  it('should render organization unit section', () => {
    renderWithProviders(<EditGeneralSettings {...defaultProps} />);

    expect(screen.getAllByText('edit.general.sections.organizationUnit.title').length).toBeGreaterThan(0);
    expect(screen.getByDisplayValue('ou-123')).toBeInTheDocument();
  });

  it('should render danger zone section', () => {
    renderWithProviders(<EditGeneralSettings {...defaultProps} />);

    expect(screen.getByText('edit.general.sections.dangerZone.title')).toBeInTheDocument();
    expect(screen.getByText('edit.general.sections.dangerZone.deleteGroup')).toBeInTheDocument();
  });

  it('should call onDeleteClick when delete button is clicked', async () => {
    const user = userEvent.setup();
    renderWithProviders(<EditGeneralSettings {...defaultProps} />);

    await user.click(screen.getByText('Delete'));

    expect(defaultProps.onDeleteClick).toHaveBeenCalled();
  });

  it('should have read-only organization unit field', () => {
    renderWithProviders(<EditGeneralSettings {...defaultProps} />);

    const ouInput = screen.getByDisplayValue('ou-123');
    expect(ouInput).toHaveAttribute('readonly');
  });

  it('should render copy button for organization unit ID', () => {
    renderWithProviders(<EditGeneralSettings {...defaultProps} />);

    expect(
      screen.getByLabelText('edit.general.sections.quickCopy.copyOrganizationUnitId'),
    ).toBeInTheDocument();
  });

  it('should copy organization unit ID to clipboard when copy button is clicked', async () => {
    renderWithProviders(<EditGeneralSettings {...defaultProps} />);

    const copyButton = screen.getByLabelText('edit.general.sections.quickCopy.copyOrganizationUnitId');
    fireEvent.click(copyButton);

    await waitFor(() => {
      expect(mockWriteText).toHaveBeenCalledWith('ou-123');
    });
  });
});
