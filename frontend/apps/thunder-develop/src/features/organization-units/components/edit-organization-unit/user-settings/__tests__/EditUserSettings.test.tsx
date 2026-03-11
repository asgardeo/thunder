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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {page} from 'vitest/browser';
import {renderWithProviders} from '@thunder/test-utils/browser';
import EditUserSettings from '../EditUserSettings';

// Mock child component
vi.mock('../ManageUsersSection', () => ({
  default: ({organizationUnitId}: {organizationUnitId: string}) => (
    <div data-testid="manage-users-section">ManageUsersSection - {organizationUnitId}</div>
  ),
}));

describe('EditUserSettings', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render ManageUsersSection', async () => {
    await renderWithProviders(<EditUserSettings organizationUnitId="ou-123" />);

    await expect.element(page.getByTestId('manage-users-section')).toBeInTheDocument();
  });

  it('should pass organizationUnitId to ManageUsersSection', async () => {
    await renderWithProviders(<EditUserSettings organizationUnitId="ou-456" />);

    await expect.element(page.getByText('ManageUsersSection - ou-456')).toBeInTheDocument();
  });

  it('should handle different organization unit IDs', async () => {
    const {rerender} = await renderWithProviders(<EditUserSettings organizationUnitId="ou-123" />);

    await expect.element(page.getByText('ManageUsersSection - ou-123')).toBeInTheDocument();

    await rerender(<EditUserSettings organizationUnitId="ou-789" />);

    await expect.element(page.getByText('ManageUsersSection - ou-789')).toBeInTheDocument();
  });
});
