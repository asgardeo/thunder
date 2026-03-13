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
import {screen, fireEvent, waitFor, renderWithProviders} from '@thunder/test-utils';
import QuickCopySection from '../QuickCopySection';
import type {OrganizationUnit} from '../../../../models/organization-unit';

// Mock translations
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'organizationUnits:edit.general.sections.quickCopy.title': 'Quick Copy',
        'organizationUnits:edit.general.sections.quickCopy.description': 'Copy organization unit identifiers',
        'organizationUnits:edit.general.handle.label': 'Handle',
        'organizationUnits:edit.general.ou.id.label': 'Organization Unit ID',
        'common:actions.copy': 'Copy',
        'common:actions.copied': 'Copied',
      };
      return translations[key] ?? key;
    },
  }),
}));

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

  it('should render the quick copy section', () => {
    renderWithProviders(
      <QuickCopySection
        organizationUnit={mockOrganizationUnit}
        copiedField={null}
        onCopyToClipboard={mockOnCopyToClipboard}
      />,
    );

    expect(screen.getByText('Quick Copy')).toBeInTheDocument();
    expect(screen.getByText('Copy organization unit identifiers')).toBeInTheDocument();
  });

  it('should render handle field with correct value', () => {
    renderWithProviders(
      <QuickCopySection
        organizationUnit={mockOrganizationUnit}
        copiedField={null}
        onCopyToClipboard={mockOnCopyToClipboard}
      />,
    );

    const handleInput = screen.getByDisplayValue('engineering');
    expect(handleInput).toBeInTheDocument();
    expect(handleInput).toHaveAttribute('readonly');
  });

  it('should render organization unit ID field with correct value', () => {
    renderWithProviders(
      <QuickCopySection
        organizationUnit={mockOrganizationUnit}
        copiedField={null}
        onCopyToClipboard={mockOnCopyToClipboard}
      />,
    );

    const idInput = screen.getByDisplayValue('ou-123');
    expect(idInput).toBeInTheDocument();
    expect(idInput).toHaveAttribute('readonly');
  });

  it('should call onCopyToClipboard when handle copy button is clicked', async () => {
    renderWithProviders(
      <QuickCopySection
        organizationUnit={mockOrganizationUnit}
        copiedField={null}
        onCopyToClipboard={mockOnCopyToClipboard}
      />,
    );

    const copyButtons = screen.getAllByRole('button', {name: 'Copy'});
    fireEvent.click(copyButtons[0]); // First copy button is for handle

    await waitFor(() => {
      expect(mockOnCopyToClipboard).toHaveBeenCalledWith('engineering', 'handle');
    });
  });

  it('should call onCopyToClipboard when ID copy button is clicked', async () => {
    renderWithProviders(
      <QuickCopySection
        organizationUnit={mockOrganizationUnit}
        copiedField={null}
        onCopyToClipboard={mockOnCopyToClipboard}
      />,
    );

    const copyButtons = screen.getAllByRole('button', {name: 'Copy'});
    fireEvent.click(copyButtons[1]); // Second copy button is for ID

    await waitFor(() => {
      expect(mockOnCopyToClipboard).toHaveBeenCalledWith('ou-123', 'ou_id');
    });
  });

  it('should show check icon when handle is copied', () => {
    renderWithProviders(
      <QuickCopySection
        organizationUnit={mockOrganizationUnit}
        copiedField="handle"
        onCopyToClipboard={mockOnCopyToClipboard}
      />,
    );

    const copiedButton = screen.getByLabelText('Copied');
    expect(copiedButton).toBeInTheDocument();
  });

  it('should show check icon when ID is copied', () => {
    renderWithProviders(
      <QuickCopySection
        organizationUnit={mockOrganizationUnit}
        copiedField="ou_id"
        onCopyToClipboard={mockOnCopyToClipboard}
      />,
    );

    const copiedButton = screen.getByLabelText('Copied');
    expect(copiedButton).toBeInTheDocument();
  });

  it('should handle copy errors gracefully', async () => {
    const mockOnCopyError = vi.fn().mockRejectedValue(new Error('Copy failed'));

    renderWithProviders(
      <QuickCopySection
        organizationUnit={mockOrganizationUnit}
        copiedField={null}
        onCopyToClipboard={mockOnCopyError}
      />,
    );

    const copyButtons = screen.getAllByLabelText('Copy');
    fireEvent.click(copyButtons[0]);

    await waitFor(() => {
      expect(mockOnCopyError).toHaveBeenCalled();
    });

    // Should not throw error
  });
});
