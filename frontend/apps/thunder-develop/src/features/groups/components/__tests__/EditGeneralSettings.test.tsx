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
import {page, userEvent} from 'vitest/browser';
import {renderWithProviders, getByDisplayValue} from '@thunder/test-utils/browser';
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

  it('should render organization unit section', async () => {
    await renderWithProviders(<EditGeneralSettings {...defaultProps} />);

    await expect.element(page.getByRole('heading', {name: 'Organization Unit'})).toBeInTheDocument();
    await expect.element(getByDisplayValue('ou-123')).toBeInTheDocument();
  });

  it('should render danger zone section', async () => {
    await renderWithProviders(<EditGeneralSettings {...defaultProps} />);

    await expect.element(page.getByRole('heading', {name: 'Danger Zone'})).toBeInTheDocument();
    await expect.element(page.getByRole('heading', {name: 'Delete this group'})).toBeInTheDocument();
  });

  it('should call onDeleteClick when delete button is clicked', async () => {
    await renderWithProviders(<EditGeneralSettings {...defaultProps} />);

    await userEvent.click(page.getByRole('button', {name: 'Delete'}));

    expect(defaultProps.onDeleteClick).toHaveBeenCalled();
  });

  it('should have read-only organization unit field', async () => {
    await renderWithProviders(<EditGeneralSettings {...defaultProps} />);

    await expect.element(getByDisplayValue('ou-123')).toHaveAttribute('readonly');
  });

  it('should render copy button for organization unit ID', async () => {
    await renderWithProviders(<EditGeneralSettings {...defaultProps} />);

    await expect.element(
      page.getByLabelText('Copy organization unit ID'),
    ).toBeInTheDocument();
  });

  it('should copy organization unit ID to clipboard when copy button is clicked', async () => {
    await renderWithProviders(<EditGeneralSettings {...defaultProps} />);

    await userEvent.click(page.getByLabelText('Copy organization unit ID'));

    await vi.waitFor(() => {
      expect(mockWriteText).toHaveBeenCalledWith('ou-123');
    });
  });
});
