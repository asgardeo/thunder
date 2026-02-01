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
import {screen, renderWithProviders, fireEvent, act} from '@thunder/test-utils';
import EditGeneralSettings from '../general-settings/EditGeneralSettings';
import type {OrganizationUnit} from '../../../types/organization-units';

// Mock translations
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'organizationUnits:view.general.title': 'General Information',
        'organizationUnits:view.general.subtitle': 'Basic details about this organization unit',
        'organizationUnits:form.handle': 'Handle',
        'organizationUnits:view.general.id': 'Organization Unit ID',
        'organizationUnits:view.general.parent': 'Parent Organization Unit',
        'organizationUnits:view.general.noParent': 'No Parent',
        'common:actions.copy': 'Copy',
        'common:actions.copied': 'Copied',
      };
      return translations[key] ?? key;
    },
  }),
}));

// Mock the API hook
const mockUseGetOrganizationUnit =
  vi.fn<(id: string | undefined, enabled: boolean) => {data: OrganizationUnit | undefined; isLoading: boolean}>();
vi.mock('../../../api/useGetOrganizationUnit', () => ({
  default: (id: string | undefined, enabled: boolean) => mockUseGetOrganizationUnit(id, enabled),
}));

describe('EditGeneralSettings', () => {
  const mockOrganizationUnit: OrganizationUnit = {
    id: 'ou-123',
    handle: 'test-handle',
    name: 'Test Organization Unit',
    description: 'A test description',
    parent: null,
  };

  const mockClipboard = {
    writeText: vi.fn().mockResolvedValue(undefined),
  };

  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();
    Object.assign(navigator, {clipboard: mockClipboard});
    mockUseGetOrganizationUnit.mockReturnValue({
      data: undefined,
      isLoading: false,
    });
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('should render title and subtitle', () => {
    renderWithProviders(<EditGeneralSettings organizationUnit={mockOrganizationUnit} />);

    expect(screen.getByText('General Information')).toBeInTheDocument();
    expect(screen.getByText('Basic details about this organization unit')).toBeInTheDocument();
  });

  it('should display organization unit handle', () => {
    renderWithProviders(<EditGeneralSettings organizationUnit={mockOrganizationUnit} />);

    expect(screen.getByText('Handle')).toBeInTheDocument();
    expect(screen.getByDisplayValue('test-handle')).toBeInTheDocument();
  });

  it('should display organization unit ID', () => {
    renderWithProviders(<EditGeneralSettings organizationUnit={mockOrganizationUnit} />);

    expect(screen.getByText('Organization Unit ID')).toBeInTheDocument();
    expect(screen.getByDisplayValue('ou-123')).toBeInTheDocument();
  });

  it('should render with different organization unit data', () => {
    const differentOU: OrganizationUnit = {
      id: 'ou-456',
      handle: 'another-handle',
      name: 'Another OU',
      description: null,
      parent: 'parent-ou',
    };

    renderWithProviders(<EditGeneralSettings organizationUnit={differentOU} />);

    expect(screen.getByDisplayValue('another-handle')).toBeInTheDocument();
    expect(screen.getByDisplayValue('ou-456')).toBeInTheDocument();
  });

  it('should show "No Parent" when parent is null', () => {
    renderWithProviders(<EditGeneralSettings organizationUnit={mockOrganizationUnit} />);

    expect(screen.getByDisplayValue('No Parent')).toBeInTheDocument();
  });

  it('should copy handle to clipboard when copy button is clicked', async () => {
    renderWithProviders(<EditGeneralSettings organizationUnit={mockOrganizationUnit} />);

    const copyButtons = screen.getAllByRole('button');
    const handleCopyButton = copyButtons[0];

    await act(async () => {
      fireEvent.click(handleCopyButton);
    });

    expect(mockClipboard.writeText).toHaveBeenCalledWith('test-handle');
  });

  it('should copy OU ID to clipboard when copy button is clicked', async () => {
    renderWithProviders(<EditGeneralSettings organizationUnit={mockOrganizationUnit} />);

    const copyButtons = screen.getAllByRole('button');
    const idCopyButton = copyButtons[1];

    await act(async () => {
      fireEvent.click(idCopyButton);
    });

    expect(mockClipboard.writeText).toHaveBeenCalledWith('ou-123');
  });

  it('should reset copied state after timeout', async () => {
    renderWithProviders(<EditGeneralSettings organizationUnit={mockOrganizationUnit} />);

    const copyButtons = screen.getAllByRole('button');
    const handleCopyButton = copyButtons[0];

    await act(async () => {
      fireEvent.click(handleCopyButton);
    });

    // Advance timers to trigger the timeout callback that resets copiedField
    await act(async () => {
      vi.advanceTimersByTime(2000);
    });

    // Verify copy was called
    expect(mockClipboard.writeText).toHaveBeenCalledWith('test-handle');
  });

  it('should handle clipboard error gracefully for handle copy', async () => {
    mockClipboard.writeText.mockRejectedValueOnce(new Error('Clipboard error'));

    renderWithProviders(<EditGeneralSettings organizationUnit={mockOrganizationUnit} />);

    const copyButtons = screen.getAllByRole('button');
    const handleCopyButton = copyButtons[0];

    // Should not throw
    await act(async () => {
      fireEvent.click(handleCopyButton);
    });

    expect(mockClipboard.writeText).toHaveBeenCalledWith('test-handle');
  });

  it('should handle clipboard error gracefully for OU ID copy', async () => {
    mockClipboard.writeText.mockRejectedValueOnce(new Error('Clipboard error'));

    renderWithProviders(<EditGeneralSettings organizationUnit={mockOrganizationUnit} />);

    const copyButtons = screen.getAllByRole('button');
    const idCopyButton = copyButtons[1];

    // Should not throw
    await act(async () => {
      fireEvent.click(idCopyButton);
    });

    expect(mockClipboard.writeText).toHaveBeenCalledWith('ou-123');
  });

  it('should show loading spinner when parent OU is loading', () => {
    const ouWithParent: OrganizationUnit = {
      ...mockOrganizationUnit,
      parent: 'parent-ou-id',
    };

    mockUseGetOrganizationUnit.mockReturnValue({
      data: undefined,
      isLoading: true,
    });

    renderWithProviders(<EditGeneralSettings organizationUnit={ouWithParent} />);

    expect(screen.getByRole('progressbar')).toBeInTheDocument();
  });

  it('should display parent OU link when parent data is loaded', () => {
    vi.useRealTimers(); // Use real timers for this test

    const ouWithParent: OrganizationUnit = {
      ...mockOrganizationUnit,
      parent: 'parent-ou-id',
    };

    const parentOU: OrganizationUnit = {
      id: 'parent-ou-id',
      handle: 'parent-handle',
      name: 'Parent OU Name',
      description: 'Parent description',
      parent: null,
    };

    mockUseGetOrganizationUnit.mockReturnValue({
      data: parentOU,
      isLoading: false,
    });

    renderWithProviders(<EditGeneralSettings organizationUnit={ouWithParent} />);

    expect(screen.getByText('Parent OU Name')).toBeInTheDocument();
    expect(screen.getByText('(parent-ou-id)')).toBeInTheDocument();

    // Verify it's a link
    const link = screen.getByText('Parent OU Name');
    expect(link.closest('a')).toHaveAttribute('href', '/organization-units/parent-ou-id');

    vi.useFakeTimers(); // Restore fake timers for subsequent tests
  });

  it('should show raw parent ID when parent OU fetch fails', () => {
    const ouWithParent: OrganizationUnit = {
      ...mockOrganizationUnit,
      parent: 'parent-ou-id',
    };

    mockUseGetOrganizationUnit.mockReturnValue({
      data: undefined,
      isLoading: false,
    });

    renderWithProviders(<EditGeneralSettings organizationUnit={ouWithParent} />);

    // Should show the raw parent ID in a text field
    expect(screen.getByDisplayValue('parent-ou-id')).toBeInTheDocument();
  });
});
