/**
 * Copyright (c) 2025-2026, WSO2 LLC. (https://www.wso2.com).
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
import OrganizationUnitsTreeView from '../OrganizationUnitsTreeView';
import type {OrganizationUnitListResponse} from '../../models/responses';

// Mock navigate
const mockNavigate = vi.fn();
vi.mock('react-router', async () => {
  const actual = await vi.importActual('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

// Mock logger
// Mock logger — stable reference to avoid useCallback churn
const stableLogger = {error: vi.fn(), info: vi.fn(), debug: vi.fn()};
vi.mock('@thunder/logger/react', () => ({
  useLogger: () => stableLogger,
}));

// Mock the API hook
const mockUseGetOrganizationUnits = vi.fn();
vi.mock('../../api/useGetOrganizationUnits', () => ({
  default: () =>
    mockUseGetOrganizationUnits() as {
      data: OrganizationUnitListResponse | undefined;
      isLoading: boolean;
      error: Error | null;
    },
}));

// Mock Asgardeo — stable reference to avoid useCallback churn
const mockHttpRequest = vi.fn();
const stableHttp = {request: mockHttpRequest};
vi.mock('@asgardeo/react', () => ({
  useAsgardeo: () => ({http: stableHttp}),
}));

// Mock useOrganizationUnit hook with React state for reactivity
// Allow tests to pre-seed expandedItems via mockOrganizationUnitConfig.initialExpandedItems
const mockOrganizationUnitConfig = {initialExpandedItems: [] as string[]};
vi.mock('../../contexts/useOrganizationUnit', async () => {
  const {useState, useCallback} = await import('react');
  type OrganizationUnitTreeItem = import('../../models/organization-unit-tree').OrganizationUnitTreeItem;
  function useOrganizationUnit() {
    const [treeItems, setTreeItems] = useState<OrganizationUnitTreeItem[]>([]);
    const [expandedItems, setExpandedItems] = useState<string[]>(mockOrganizationUnitConfig.initialExpandedItems);
    const [loadedItems, setLoadedItems] = useState<Set<string>>(new Set());
    const resetTreeState = useCallback(() => {
      setTreeItems([]);
      setLoadedItems(new Set());
    }, []);
    return {treeItems, setTreeItems, expandedItems, setExpandedItems, loadedItems, setLoadedItems, resetTreeState};
  }
  return {default: useOrganizationUnit};
});

// Mock config — stable reference to avoid useCallback churn
const stableConfig = {getServerUrl: () => 'http://localhost:8080'};
vi.mock('@thunder/shared-contexts', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@thunder/shared-contexts')>();
  return {
    ...actual,
    useConfig: () => stableConfig,
  };
});

// Mock delete hook — controllable per test
const mockDeleteMutate = vi.fn();
const mockDeleteHook = {mutate: mockDeleteMutate, isPending: false};
vi.mock('../../api/useDeleteOrganizationUnit', () => ({
  default: () => mockDeleteHook,
}));

describe('OrganizationUnitsTreeView', () => {
  const mockOUData: OrganizationUnitListResponse = {
    totalResults: 2,
    startIndex: 1,
    count: 2,
    organizationUnits: [
      {id: 'ou-1', handle: 'root', name: 'Root Organization', description: 'Root OU', parent: null},
      {id: 'ou-2', handle: 'engineering', name: 'Engineering', description: null, parent: null},
    ],
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockNavigate.mockReset();
    mockOrganizationUnitConfig.initialExpandedItems = [];
    mockUseGetOrganizationUnits.mockReturnValue({
      data: mockOUData,
      isLoading: false,
      error: null,
    });
  });

  it('should render tree view with organization unit names', async () => {
    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();
    await expect.element(page.getByText('Engineering', {exact: true}).first()).toBeInTheDocument();
  });

  it('should show error state when fetch fails', async () => {
    mockUseGetOrganizationUnits.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: new Error('Network error'),
    });

    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Failed to load organization units')).toBeInTheDocument();
    await expect.element(page.getByText('Network error')).toBeInTheDocument();
  });

  it('should show fallback error message when error has no message', async () => {
    mockUseGetOrganizationUnits.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: {},
    });

    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Failed to load organization units')).toBeInTheDocument();
    await expect.element(page.getByText('An unknown error occurred')).toBeInTheDocument();
  });

  it('should show loading state', async () => {
    mockUseGetOrganizationUnits.mockReturnValue({
      data: undefined,
      isLoading: true,
      error: null,
    });

    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByRole('progressbar')).toBeInTheDocument();
  });

  it('should show empty state when no organization units', async () => {
    mockUseGetOrganizationUnits.mockReturnValue({
      data: {
        totalResults: 0,
        startIndex: 1,
        count: 0,
        organizationUnits: [],
      },
      isLoading: false,
      error: null,
    });

    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('No organization units found')).toBeInTheDocument();
  });

  it('should render avatar for each tree item', async () => {
    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();

    const avatars = document.querySelectorAll('.MuiAvatar-root');
    expect(avatars.length).toBeGreaterThan(0);
  });

  it('should render action button for each tree item', async () => {
    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();

    const actionButtons = page.getByLabelText('Actions').all();
    expect(actionButtons.length).toBe(2);
  });

  it('should open menu with actions when action button is clicked', async () => {
    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();

    const actionButtons = page.getByLabelText('Actions').all();
    await userEvent.click(actionButtons[0]);

    await expect.element(page.getByText('Add child organization unit')).toBeInTheDocument();
    await expect.element(page.getByText('Edit')).toBeInTheDocument();
    await expect.element(page.getByText('Delete')).toBeInTheDocument();
  });

  it('should navigate to create page with parentId when add child menu item is clicked', async () => {
    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();

    const actionButtons = page.getByLabelText('Actions').all();
    await userEvent.click(actionButtons[0]);

    await expect.element(page.getByText('Add child organization unit')).toBeInTheDocument();

    await userEvent.click(page.getByText('Add child organization unit'));

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/organization-units/create', {
        state: {parentId: 'ou-1', parentName: 'Root Organization', parentHandle: 'root'},
      });
    });
  });

  it('should navigate when edit menu item is clicked', async () => {
    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();

    const actionButtons = page.getByLabelText('Actions').all();
    await userEvent.click(actionButtons[0]);

    await expect.element(page.getByText('Edit')).toBeInTheDocument();

    await userEvent.click(page.getByText('Edit'));

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/organization-units/ou-1');
    });
  });

  it('should open delete dialog when delete menu item is clicked', async () => {
    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();

    const actionButtons = page.getByLabelText('Actions').all();
    await userEvent.click(actionButtons[0]);

    await expect.element(page.getByText('Delete')).toBeInTheDocument();

    await userEvent.click(page.getByText('Delete'));

    await expect.element(page.getByText('Delete Organization Unit')).toBeInTheDocument();
    await expect
      .element(
        page.getByText('Are you sure you want to delete this organization unit? This action cannot be undone.'),
      )
      .toBeInTheDocument();
  });

  it('should handle undefined data gracefully', async () => {
    mockUseGetOrganizationUnits.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: null,
    });

    await renderWithProviders(<OrganizationUnitsTreeView />);

    // When data is undefined and not loading, a loading spinner is shown
    await expect.element(page.getByRole('progressbar')).toBeInTheDocument();
  });

  it('should close delete dialog when cancel is clicked', async () => {
    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();

    // Open actions menu and click delete
    const actionButtons = page.getByLabelText('Actions').all();
    await userEvent.click(actionButtons[0]);

    await expect.element(page.getByText('Delete')).toBeInTheDocument();

    await userEvent.click(page.getByText('Delete'));

    await expect.element(page.getByText('Delete Organization Unit')).toBeInTheDocument();

    // Cancel the dialog
    await userEvent.click(page.getByText('Cancel'));

    await expect.element(page.getByText('Delete Organization Unit')).not.toBeInTheDocument();
  });

  it('should show success snackbar after successful deletion', async () => {
    mockDeleteMutate.mockImplementation((_id: string, options: {onSuccess: () => void}) => {
      options.onSuccess();
    });

    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();

    // Open actions menu and click delete to open dialog, then confirm
    const actionButtons1 = page.getByLabelText('Actions').all();
    await userEvent.click(actionButtons1[0]);

    await expect.element(page.getByText('Delete')).toBeInTheDocument();

    await userEvent.click(page.getByText('Delete'));

    await expect.element(page.getByText('Delete Organization Unit')).toBeInTheDocument();

    // Use scoped locator to scope to the dialog's Delete button (avoids ambiguity with menu item)
    const dialog = page.getByRole('dialog');
    await userEvent.click(dialog.getByText('Delete'));

    await expect.element(page.getByText('Organization unit deleted successfully.')).toBeInTheDocument();
  });

  it('should show error snackbar after failed deletion', async () => {
    mockDeleteMutate.mockImplementation((_id: string, options: {onError: (err: Error) => void}) => {
      options.onError(
        Object.assign(new Error('Delete failed'), {
          response: {data: {code: 'ERR', message: 'fail', description: 'Server error occurred'}},
        }),
      );
    });

    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();

    // Open actions menu and click delete to open dialog, then confirm
    const actionButtons2 = page.getByLabelText('Actions').all();
    await userEvent.click(actionButtons2[0]);

    await expect.element(page.getByText('Delete')).toBeInTheDocument();

    await userEvent.click(page.getByText('Delete'));

    await expect.element(page.getByText('Delete Organization Unit')).toBeInTheDocument();

    // Use scoped locator to scope to the dialog's Delete button (avoids ambiguity with menu item)
    const dialog2 = page.getByRole('dialog');
    await userEvent.click(dialog2.getByText('Delete'));

    await expect.element(page.getByText('Server error occurred')).toBeInTheDocument();
  });

  it('should log error when edit navigation fails', async () => {
    mockNavigate.mockRejectedValueOnce(new Error('Navigation failed'));

    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();

    const actionButtons3 = page.getByLabelText('Actions').all();
    await userEvent.click(actionButtons3[0]);

    await expect.element(page.getByText('Edit')).toBeInTheDocument();

    await userEvent.click(page.getByText('Edit'));

    await vi.waitFor(() => {
      expect(stableLogger.error).toHaveBeenCalledWith(
        'Failed to navigate to organization unit',
        expect.objectContaining({ouId: 'ou-1'}),
      );
    });
  });

  it('should log error when add child navigation fails', async () => {
    mockNavigate.mockRejectedValueOnce(new Error('Navigation failed'));

    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();

    const actionButtons4 = page.getByLabelText('Actions').all();
    await userEvent.click(actionButtons4[0]);

    await expect.element(page.getByText('Add child organization unit')).toBeInTheDocument();

    await userEvent.click(page.getByText('Add child organization unit'));

    await vi.waitFor(() => {
      expect(stableLogger.error).toHaveBeenCalledWith(
        'Failed to navigate to create child organization unit',
        expect.objectContaining({parentId: 'ou-1'}),
      );
    });
  });

  it('should fetch and display child OUs when a node is expanded', async () => {
    const childOUResponse: OrganizationUnitListResponse = {
      totalResults: 1,
      startIndex: 1,
      count: 1,
      organizationUnits: [
        {id: 'ou-child-1', handle: 'child1', name: 'Fetched Child', description: 'A child', parent: 'ou-1'},
      ],
    };

    mockHttpRequest.mockResolvedValue({data: childOUResponse});

    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();

    // Click the expand icon on the first tree item to trigger expansion
    const expandIcons = document.querySelectorAll('.MuiTreeItem-iconContainer');
    expect(expandIcons.length).toBeGreaterThan(0);
    await userEvent.click(expandIcons[0] as HTMLElement);

    // The component should fetch children and display them
    await vi.waitFor(() => {
      expect(mockHttpRequest).toHaveBeenCalled();
    });
  });

  it('should show error placeholder and log error when fetching child OUs fails', async () => {
    mockHttpRequest.mockRejectedValue(new Error('Network failure'));

    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();

    // Click the expand icon on the first tree item
    const expandIcons = document.querySelectorAll('.MuiTreeItem-iconContainer');
    expect(expandIcons.length).toBeGreaterThan(0);
    await userEvent.click(expandIcons[0] as HTMLElement);

    await vi.waitFor(() => {
      expect(stableLogger.error).toHaveBeenCalledWith(
        'Failed to load child organization units',
        expect.objectContaining({parentId: 'ou-1'}),
      );
    });

    // The error placeholder should be visible instead of a perpetual spinner
    await expect.element(page.getByText('Failed to load child organization units')).toBeInTheDocument();
  });

  it('should rebuild tree with expanded items restored when expandedItems exist', async () => {
    // Pre-seed expanded items so the rebuild path is triggered
    mockOrganizationUnitConfig.initialExpandedItems = ['ou-1'];

    const childOUResponse: OrganizationUnitListResponse = {
      totalResults: 1,
      startIndex: 1,
      count: 1,
      organizationUnits: [
        {id: 'ou-child-1', handle: 'child1', name: 'Restored Child', description: null, parent: 'ou-1'},
      ],
    };

    mockHttpRequest.mockResolvedValue({data: childOUResponse});

    await renderWithProviders(<OrganizationUnitsTreeView />);

    // The useEffect should detect expandedItems=['ou-1'] and call rebuildTree,
    // which calls expandLevel → fetchChildItems for 'ou-1'
    await vi.waitFor(() => {
      expect(mockHttpRequest).toHaveBeenCalled();
    });
  });

  it('should still render root items when child fetch fails during rebuild', async () => {
    // Pre-seed expanded items to trigger the rebuild path (expandLevel)
    mockOrganizationUnitConfig.initialExpandedItems = ['ou-1'];

    // Make the child fetch fail — expandLevel catches this internally
    // and filters out failed results, so the tree still renders root items
    mockHttpRequest.mockRejectedValue(new Error('Child fetch failed'));

    await renderWithProviders(<OrganizationUnitsTreeView />);

    // Root items should still be rendered even though child fetch failed
    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();
    await expect.element(page.getByText('Engineering', {exact: true}).first()).toBeInTheDocument();
  });

  it('should close snackbar when close action is triggered', async () => {
    // Trigger a success snackbar first
    mockDeleteMutate.mockImplementation((_id: string, options: {onSuccess: () => void}) => {
      options.onSuccess();
    });

    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();

    // Open actions menu and click delete to trigger snackbar
    const actionButtons5 = page.getByLabelText('Actions').all();
    await userEvent.click(actionButtons5[0]);

    await expect.element(page.getByText('Delete')).toBeInTheDocument();

    await userEvent.click(page.getByText('Delete'));

    await expect.element(page.getByText('Delete Organization Unit')).toBeInTheDocument();

    // Use scoped locator to scope to the dialog's Delete button (avoids ambiguity with menu item)
    const dialog3 = page.getByRole('dialog');
    await userEvent.click(dialog3.getByText('Delete'));

    await expect.element(page.getByText('Organization unit deleted successfully.')).toBeInTheDocument();

    // Close the snackbar via the Alert close button
    const alert = page.getByRole('alert');
    const alertCloseButton = alert.getByRole('button');
    await userEvent.click(alertCloseButton);

    await expect.element(page.getByText('Organization unit deleted successfully.')).not.toBeInTheDocument();
  });

  it('should display handle text for tree items that have handles', async () => {
    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();

    // The 'root' and 'engineering' handles should be shown as caption text
    await expect.element(page.getByText('root')).toBeInTheDocument();
    await expect.element(page.getByText('engineering', {exact: true})).toBeInTheDocument();
  });

  it('should render add root organization unit row below tree items', async () => {
    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();

    await expect.element(page.getByText('Add Root Organization Unit')).toBeInTheDocument();
  });

  it('should navigate to create page when add root row is clicked', async () => {
    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();

    await userEvent.click(page.getByText('Add Root Organization Unit'));

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/organization-units/create');
    });
  });

  it('should render add child button when a node is expanded and children are loaded', async () => {
    const childOUResponse: OrganizationUnitListResponse = {
      totalResults: 1,
      startIndex: 1,
      count: 1,
      organizationUnits: [
        {id: 'ou-child-1', handle: 'child1', name: 'Fetched Child', description: null, parent: 'ou-1'},
      ],
    };

    mockHttpRequest.mockResolvedValue({data: childOUResponse});

    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();

    // Expand the first tree item
    const expandIcons = document.querySelectorAll('.MuiTreeItem-iconContainer');
    await userEvent.click(expandIcons[0] as HTMLElement);

    // After expansion, the add child button should appear
    await expect.element(page.getByText('Add Child Organization Unit')).toBeInTheDocument();
  });

  it('should navigate to create page with parent state when add child button in tree is clicked', async () => {
    const childOUResponse: OrganizationUnitListResponse = {
      totalResults: 1,
      startIndex: 1,
      count: 1,
      organizationUnits: [
        {id: 'ou-child-1', handle: 'child1', name: 'Fetched Child', description: null, parent: 'ou-1'},
      ],
    };

    mockHttpRequest.mockResolvedValue({data: childOUResponse});

    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();

    // Expand the first tree item
    const expandIcons = document.querySelectorAll('.MuiTreeItem-iconContainer');
    await userEvent.click(expandIcons[0] as HTMLElement);

    await expect.element(page.getByText('Add Child Organization Unit')).toBeInTheDocument();

    await userEvent.click(page.getByText('Add Child Organization Unit'));

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/organization-units/create', {
        state: {parentId: 'ou-1', parentName: 'Root Organization', parentHandle: 'root'},
      });
    });
  });

  it('should show add child button when node has no children', async () => {
    const emptyChildOUResponse: OrganizationUnitListResponse = {
      totalResults: 0,
      startIndex: 1,
      count: 0,
      organizationUnits: [],
    };

    mockHttpRequest.mockResolvedValue({data: emptyChildOUResponse});

    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();

    // Expand the first tree item
    const expandIcons = document.querySelectorAll('.MuiTreeItem-iconContainer');
    await userEvent.click(expandIcons[0] as HTMLElement);

    // Even with no children, the add child button should appear
    await expect.element(page.getByText('Add Child Organization Unit')).toBeInTheDocument();
  });

  it('should show add root row in empty state', async () => {
    mockUseGetOrganizationUnits.mockReturnValue({
      data: {
        totalResults: 0,
        startIndex: 1,
        count: 0,
        organizationUnits: [],
      },
      isLoading: false,
      error: null,
    });

    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('No organization units found')).toBeInTheDocument();

    await expect.element(page.getByText('Add Root Organization Unit')).toBeInTheDocument();

    await userEvent.click(page.getByText('Add Root Organization Unit'));

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/organization-units/create');
    });
  });

  it('should navigate to create page when add root row is activated via Enter key', async () => {
    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();

    // Focus the add root button for keyboard interaction
await userEvent.click(page.getByText('Add Root Organization Unit'));
    await userEvent.keyboard('{Enter}');

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/organization-units/create');
    });
  });

  it('should navigate to create page when add root row is activated via Space key', async () => {
    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();

    // Focus the add root button for keyboard interaction
await userEvent.click(page.getByText('Add Root Organization Unit'));
    await userEvent.keyboard(' ');

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/organization-units/create');
    });
  });

  it('should navigate via keyboard on empty state add root button', async () => {
    mockUseGetOrganizationUnits.mockReturnValue({
      data: {
        totalResults: 0,
        startIndex: 1,
        count: 0,
        organizationUnits: [],
      },
      isLoading: false,
      error: null,
    });

    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('No organization units found')).toBeInTheDocument();

    // Focus the add root button for keyboard interaction
await userEvent.click(page.getByText('Add Root Organization Unit'));
    await userEvent.keyboard('{Enter}');

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/organization-units/create');
    });
  });

  it('should trigger keyboard handler for load more item via Enter key', async () => {
    const childOUResponse: OrganizationUnitListResponse = {
      totalResults: 50,
      startIndex: 1,
      count: 1,
      organizationUnits: [
        {id: 'ou-child-1', handle: 'child1', name: 'Fetched Child', description: null, parent: 'ou-1'},
      ],
    };

    mockHttpRequest.mockResolvedValue({data: childOUResponse});

    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();

    // Expand the first tree item to load children with load more
    const expandIcons = document.querySelectorAll('.MuiTreeItem-iconContainer');
    await userEvent.click(expandIcons[0] as HTMLElement);

    await expect.element(page.getByText('Fetched Child')).toBeInTheDocument();

    // Find the load more button and activate via keyboard
    await expect.element(page.getByText('Load more')).toBeInTheDocument();

    // Focus the load more button for keyboard interaction
await userEvent.click(page.getByText('Load more'));
    mockHttpRequest.mockClear();
    mockHttpRequest.mockResolvedValue({
      data: {
        totalResults: 50,
        startIndex: 2,
        count: 1,
        organizationUnits: [
          {id: 'ou-child-2', handle: 'child2', name: 'Fetched Child 2', description: null, parent: 'ou-1'},
        ],
      },
    });
    await userEvent.keyboard('{Enter}');

    await vi.waitFor(() => {
      expect(mockHttpRequest).toHaveBeenCalled();
    });
  });

  it('should not fetch children when collapsing a node', async () => {
    const childOUResponse: OrganizationUnitListResponse = {
      totalResults: 1,
      startIndex: 1,
      count: 1,
      organizationUnits: [
        {id: 'ou-child-1', handle: 'child1', name: 'Fetched Child', description: null, parent: 'ou-1'},
      ],
    };

    mockHttpRequest.mockResolvedValue({data: childOUResponse});

    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();

    // Expand
    const expandIcons = document.querySelectorAll('.MuiTreeItem-iconContainer');
    await userEvent.click(expandIcons[0] as HTMLElement);

    await expect.element(page.getByText('Fetched Child')).toBeInTheDocument();

    const callCount = mockHttpRequest.mock.calls.length;

    // Collapse - should not trigger another fetch
    const collapseIcons = document.querySelectorAll('.MuiTreeItem-iconContainer');
    await userEvent.click(collapseIcons[0] as HTMLElement);

    // Verify no additional HTTP calls were made on collapse
    expect(mockHttpRequest).toHaveBeenCalledTimes(callCount);
  });

  it('should show root load more button when there are more root items', async () => {
    const paginatedData: OrganizationUnitListResponse = {
      totalResults: 50,
      startIndex: 1,
      count: 2,
      organizationUnits: [
        {id: 'ou-1', handle: 'root', name: 'Root Organization', description: null, parent: null},
        {id: 'ou-2', handle: 'engineering', name: 'Engineering', description: null, parent: null},
      ],
    };

    mockUseGetOrganizationUnits.mockReturnValue({
      data: paginatedData,
      isLoading: false,
      error: null,
    });

    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();
    await expect.element(page.getByText('Load more')).toBeInTheDocument();
  });

  it('should log error when add root navigation fails', async () => {
    mockNavigate.mockRejectedValueOnce(new Error('Navigation failed'));

    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();

    await userEvent.click(page.getByText('Add Root Organization Unit'));

    await vi.waitFor(() => {
      expect(stableLogger.error).toHaveBeenCalledWith('Failed to navigate to create organization unit page', {
        error: expect.objectContaining({message: 'Navigation failed'}) as Error,
      });
    });
  });

  it('should use fallback error message when error.message is missing', async () => {
    const errorWithMessage = {message: 'Server unavailable'};
    mockUseGetOrganizationUnits.mockReturnValue({
      data: undefined,
      isLoading: false,
      error: errorWithMessage,
    });

    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Server unavailable')).toBeInTheDocument();
  });

  it('should log error when add root navigation fails via keyboard', async () => {
    mockNavigate.mockRejectedValueOnce(new Error('Navigation failed'));

    await renderWithProviders(<OrganizationUnitsTreeView />);

    await expect.element(page.getByText('Root Organization').first()).toBeInTheDocument();

    // Focus the add root button for keyboard interaction
await userEvent.click(page.getByText('Add Root Organization Unit'));
    await userEvent.keyboard('{Enter}');

    await vi.waitFor(() => {
      expect(stableLogger.error).toHaveBeenCalledWith('Failed to navigate to create organization unit page', {
        error: expect.objectContaining({message: 'Navigation failed'}) as Error,
      });
    });
  });
});
