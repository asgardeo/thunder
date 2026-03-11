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
import EditGroupSettings from '../EditGroupSettings';

// Mock child component
vi.mock('../ManageGroupsSection', () => ({
  default: ({organizationUnitId}: {organizationUnitId: string}) => (
    <div data-testid="manage-groups-section">ManageGroupsSection - {organizationUnitId}</div>
  ),
}));

describe('EditGroupSettings', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render ManageGroupsSection', async () => {
    await renderWithProviders(<EditGroupSettings organizationUnitId="ou-123" />);

    await expect.element(page.getByTestId('manage-groups-section')).toBeInTheDocument();
  });

  it('should pass organizationUnitId to ManageGroupsSection', async () => {
    await renderWithProviders(<EditGroupSettings organizationUnitId="ou-456" />);

    await expect.element(page.getByText('ManageGroupsSection - ou-456')).toBeInTheDocument();
  });

  it('should handle different organization unit IDs', async () => {
    const {rerender} = await renderWithProviders(<EditGroupSettings organizationUnitId="ou-123" />);

    await expect.element(page.getByText('ManageGroupsSection - ou-123')).toBeInTheDocument();

    await rerender(<EditGroupSettings organizationUnitId="ou-789" />);

    await expect.element(page.getByText('ManageGroupsSection - ou-789')).toBeInTheDocument();
  });
});
