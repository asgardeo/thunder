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

import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';
import {page, userEvent} from 'vitest/browser';
import {renderWithProviders} from '@thunder/test-utils/browser';
import DangerZoneSection from '../DangerZoneSection';

describe('DangerZoneSection', () => {
  const mockOnDeleteClick = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should render the danger zone section', async () => {
    await renderWithProviders(<DangerZoneSection onDeleteClick={mockOnDeleteClick} />);

    await expect.element(page.getByText('Danger Zone')).toBeInTheDocument();
    await expect.element(page.getByText('Actions in this section are irreversible. Proceed with caution.')).toBeInTheDocument();
  });

  it('should render delete organization unit title', async () => {
    await renderWithProviders(<DangerZoneSection onDeleteClick={mockOnDeleteClick} />);

    await expect.element(page.getByRole('heading', {name: 'Delete Organization Unit', level: 6})).toBeInTheDocument();
  });

  it('should render warning description', async () => {
    await renderWithProviders(<DangerZoneSection onDeleteClick={mockOnDeleteClick} />);

    await expect.element(
      page.getByText('Deleting this organization unit is permanent and cannot be undone.'),
    ).toBeInTheDocument();
  });

  it('should render delete button', async () => {
    await renderWithProviders(<DangerZoneSection onDeleteClick={mockOnDeleteClick} />);

    await expect.element(page.getByRole('button', {name: 'Delete Organization Unit'})).toBeInTheDocument();
  });

  it('should call onDeleteClick when delete button is clicked', async () => {
    await renderWithProviders(<DangerZoneSection onDeleteClick={mockOnDeleteClick} />);

    await userEvent.click(page.getByRole('button', {name: 'Delete Organization Unit'}));

    expect(mockOnDeleteClick).toHaveBeenCalledTimes(1);
  });

  it('should call onDeleteClick multiple times when clicked multiple times', async () => {
    await renderWithProviders(<DangerZoneSection onDeleteClick={mockOnDeleteClick} />);

    const deleteButton = page.getByRole('button', {name: 'Delete Organization Unit'});
    await userEvent.click(deleteButton);
    await userEvent.click(deleteButton);
    await userEvent.click(deleteButton);

    expect(mockOnDeleteClick).toHaveBeenCalledTimes(3);
  });

  it('should render delete button with error color', async () => {
    await renderWithProviders(<DangerZoneSection onDeleteClick={mockOnDeleteClick} />);

    const deleteButton = page.getByRole('button', {name: 'Delete Organization Unit'});
    await expect.element(deleteButton).toHaveClass('MuiButton-colorError');
  });
});
