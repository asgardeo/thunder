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
import {page, userEvent} from 'vitest/browser';
import {renderWithProviders} from '@thunder/test-utils/browser';
import QuickCopySection from '../QuickCopySection';
import type {OrganizationUnit} from '../../../../models/organization-unit';

describe('QuickCopySection', () => {
  const mockOrganizationUnit: OrganizationUnit = {
    id: 'ou-123',
    handle: 'engineering',
    name: 'Engineering',
    description: 'Engineering department',
    parent: null,
  };

  const mockOnCopyToClipboard = vi.fn().mockResolvedValue(undefined);

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render the quick copy section', async () => {
    await renderWithProviders(
      <QuickCopySection
        organizationUnit={mockOrganizationUnit}
        copiedField={null}
        onCopyToClipboard={mockOnCopyToClipboard}
      />,
    );

    await expect.element(page.getByText('Quick Copy')).toBeInTheDocument();
    await expect.element(page.getByText('Copy organization unit identifiers')).toBeInTheDocument();
  });

  it('should render handle field with correct value', async () => {
    await renderWithProviders(
      <QuickCopySection
        organizationUnit={mockOrganizationUnit}
        copiedField={null}
        onCopyToClipboard={mockOnCopyToClipboard}
      />,
    );

    const handleInput = page.getByRole('textbox', {name: 'Handle'});
    await expect.element(handleInput).toHaveValue('engineering');
    await expect.element(handleInput).toHaveAttribute('readonly');
  });

  it('should render organization unit ID field with correct value', async () => {
    await renderWithProviders(
      <QuickCopySection
        organizationUnit={mockOrganizationUnit}
        copiedField={null}
        onCopyToClipboard={mockOnCopyToClipboard}
      />,
    );

    const idInput = page.getByRole('textbox', {name: 'Organization Unit ID'});
    await expect.element(idInput).toHaveValue('ou-123');
    await expect.element(idInput).toHaveAttribute('readonly');
  });

  it('should call onCopyToClipboard when handle copy button is clicked', async () => {
    await renderWithProviders(
      <QuickCopySection
        organizationUnit={mockOrganizationUnit}
        copiedField={null}
        onCopyToClipboard={mockOnCopyToClipboard}
      />,
    );

    const copyButtons = page.getByRole('button', {name: 'Copy'}).all();
    await userEvent.click(copyButtons[0]); // First copy button is for handle

    await vi.waitFor(() => {
      expect(mockOnCopyToClipboard).toHaveBeenCalledWith('engineering', 'handle');
    });
  });

  it('should call onCopyToClipboard when ID copy button is clicked', async () => {
    await renderWithProviders(
      <QuickCopySection
        organizationUnit={mockOrganizationUnit}
        copiedField={null}
        onCopyToClipboard={mockOnCopyToClipboard}
      />,
    );

    const copyButtons = page.getByRole('button', {name: 'Copy'}).all();
    await userEvent.click(copyButtons[1]); // Second copy button is for ID

    await vi.waitFor(() => {
      expect(mockOnCopyToClipboard).toHaveBeenCalledWith('ou-123', 'ou_id');
    });
  });

  it('should show check icon when handle is copied', async () => {
    await renderWithProviders(
      <QuickCopySection
        organizationUnit={mockOrganizationUnit}
        copiedField="handle"
        onCopyToClipboard={mockOnCopyToClipboard}
      />,
    );

    await expect.element(page.getByLabelText('Copied')).toBeInTheDocument();
  });

  it('should show check icon when ID is copied', async () => {
    await renderWithProviders(
      <QuickCopySection
        organizationUnit={mockOrganizationUnit}
        copiedField="ou_id"
        onCopyToClipboard={mockOnCopyToClipboard}
      />,
    );

    await expect.element(page.getByLabelText('Copied')).toBeInTheDocument();
  });

  it('should handle copy errors gracefully', async () => {
    const mockOnCopyError = vi.fn().mockRejectedValue(new Error('Copy failed'));

    await renderWithProviders(
      <QuickCopySection
        organizationUnit={mockOrganizationUnit}
        copiedField={null}
        onCopyToClipboard={mockOnCopyError}
      />,
    );

    const copyButtons = page.getByLabelText('Copy').all();
    await userEvent.click(copyButtons[0]);

    await vi.waitFor(() => {
      expect(mockOnCopyError).toHaveBeenCalled();
    });

    // Should not throw error
  });
});
