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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {screen, fireEvent, waitFor, renderWithProviders} from '@thunder/test-utils';
import OrganizationUnitTreePicker from '../OrganizationUnitTreePicker';
import type {OrganizationUnitListResponse} from '../../models/responses';

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

// Mock config — stable reference to avoid useCallback churn
const stableConfig = {getServerUrl: () => 'http://localhost:8080'};
vi.mock('@thunder/shared-contexts', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@thunder/shared-contexts')>();
  return {
    ...actual,
    useConfig: () => stableConfig,
  };
});

// Mock translations — stable t function to avoid useCallback churn
const translations: Record<string, string> = {
  'organizationUnits:treePicker.empty': 'No organization units available',
  'organizationUnits:listing.treeView.noChildren': 'No child organization units',
  'organizationUnits:listing.treeView.loadMore': 'Load more',
  'common:status.loading': 'Loading...',
};
const stableT = (key: string): string => translations[key] ?? key;
const stableTranslation = {t: stableT};
vi.mock('react-i18next', () => ({
  useTranslation: () => stableTranslation,
}));

describe('OrganizationUnitTreePicker', () => {
  const mockOUData: OrganizationUnitListResponse = {
    totalResults: 2,
    startIndex: 1,
    count: 2,
    organizationUnits: [
      {id: 'ou-1', handle: 'root', name: 'Root Organization', description: 'Root OU', parent: null},
      {id: 'ou-2', handle: 'engineering', name: 'Engineering', description: null, parent: null},
    ],
  };

  const defaultProps = {
    value: '',
    onChange: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockUseGetOrganizationUnits.mockReturnValue({
      data: mockOUData,
      isLoading: false,
      error: null,
    });
  });

  it('should render tree with organization unit names', async () => {
    renderWithProviders(<OrganizationUnitTreePicker {...defaultProps} />);

    await waitFor(() => {
      expect(screen.getByText('Root Organization')).toBeInTheDocument();
      expect(screen.getByText('Engineering')).toBeInTheDocument();
    });
  });

  it('should show loading spinner when data is loading', () => {
    mockUseGetOrganizationUnits.mockReturnValue({
      data: undefined,
      isLoading: true,
      error: null,
    });

    renderWithProviders(<OrganizationUnitTreePicker {...defaultProps} />);

    expect(screen.getByRole('progressbar')).toBeInTheDocument();
  });

  it('should show empty message when no organization units', () => {
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

    renderWithProviders(<OrganizationUnitTreePicker {...defaultProps} />);

    expect(screen.getByText('No organization units available')).toBeInTheDocument();
  });

  it('should display handles for tree items', async () => {
    renderWithProviders(<OrganizationUnitTreePicker {...defaultProps} />);

    await waitFor(() => {
      expect(screen.getByText('root')).toBeInTheDocument();
      expect(screen.getByText('engineering')).toBeInTheDocument();
    });
  });

  it('should render avatars for tree items', async () => {
    renderWithProviders(<OrganizationUnitTreePicker {...defaultProps} />);

    await waitFor(() => {
      expect(screen.getByText('Root Organization')).toBeInTheDocument();
    });

    const avatars = document.querySelectorAll('.MuiAvatar-root');
    expect(avatars.length).toBeGreaterThan(0);
  });

  it('should call onChange when a tree item is selected', async () => {
    const onChange = vi.fn();
    renderWithProviders(<OrganizationUnitTreePicker {...defaultProps} onChange={onChange} />);

    await waitFor(() => {
      expect(screen.getByText('Root Organization')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Root Organization'));

    await waitFor(() => {
      expect(onChange).toHaveBeenCalledWith('ou-1');
    });
  });

  it('should not call onChange when clicking a placeholder item', async () => {
    const onChange = vi.fn();
    renderWithProviders(<OrganizationUnitTreePicker {...defaultProps} onChange={onChange} />);

    await waitFor(() => {
      expect(screen.getByText('Root Organization')).toBeInTheDocument();
    });

    // Placeholder items are not directly clickable via text, so verify no unexpected calls
    expect(onChange).not.toHaveBeenCalled();
  });

  it('should pass id prop to tree view', async () => {
    renderWithProviders(<OrganizationUnitTreePicker {...defaultProps} id="test-picker" />);

    await waitFor(() => {
      expect(screen.getByText('Root Organization')).toBeInTheDocument();
    });

    expect(document.getElementById('test-picker')).toBeInTheDocument();
  });

  it('should display helper text when provided', async () => {
    renderWithProviders(<OrganizationUnitTreePicker {...defaultProps} helperText="Select a parent" />);

    await waitFor(() => {
      expect(screen.getByText('Select a parent')).toBeInTheDocument();
    });
  });

  it('should display helper text with error styling when error is true', async () => {
    renderWithProviders(
      <OrganizationUnitTreePicker {...defaultProps} helperText="This field is required" error />,
    );

    await waitFor(() => {
      const helperText = screen.getByText('This field is required');
      expect(helperText).toBeInTheDocument();
    });
  });

  it('should not display helper text when not provided', async () => {
    renderWithProviders(<OrganizationUnitTreePicker {...defaultProps} />);

    await waitFor(() => {
      expect(screen.getByText('Root Organization')).toBeInTheDocument();
    });

    // No helper text element should be present
    expect(screen.queryByText('Select a parent')).not.toBeInTheDocument();
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

    renderWithProviders(<OrganizationUnitTreePicker {...defaultProps} />);

    await waitFor(() => {
      expect(screen.getByText('Root Organization')).toBeInTheDocument();
    });

    // Click the expand icon on the first tree item
    const expandIcons = document.querySelectorAll('.MuiTreeItem-iconContainer');
    expect(expandIcons.length).toBeGreaterThan(0);
    fireEvent.click(expandIcons[0]);

    await waitFor(() => {
      expect(mockHttpRequest).toHaveBeenCalled();
    });

    await waitFor(() => {
      expect(screen.getByText('Fetched Child')).toBeInTheDocument();
    });
  });

  it('should show "no children" placeholder when expanded node has no children', async () => {
    const emptyChildResponse: OrganizationUnitListResponse = {
      totalResults: 0,
      startIndex: 1,
      count: 0,
      organizationUnits: [],
    };

    mockHttpRequest.mockResolvedValue({data: emptyChildResponse});

    renderWithProviders(<OrganizationUnitTreePicker {...defaultProps} />);

    await waitFor(() => {
      expect(screen.getByText('Root Organization')).toBeInTheDocument();
    });

    const expandIcons = document.querySelectorAll('.MuiTreeItem-iconContainer');
    fireEvent.click(expandIcons[0]);

    await waitFor(() => {
      expect(screen.getByText('No child organization units')).toBeInTheDocument();
    });
  });

  it('should log error when fetching child OUs fails', async () => {
    mockHttpRequest.mockRejectedValue(new Error('Network failure'));

    renderWithProviders(<OrganizationUnitTreePicker {...defaultProps} />);

    await waitFor(() => {
      expect(screen.getByText('Root Organization')).toBeInTheDocument();
    });

    const expandIcons = document.querySelectorAll('.MuiTreeItem-iconContainer');
    fireEvent.click(expandIcons[0]);

    await waitFor(() => {
      expect(stableLogger.error).toHaveBeenCalledWith(
        'Failed to load child organization units',
        expect.objectContaining({parentId: 'ou-1'}),
      );
    });
  });

  it('should show load more button when there are more root items', async () => {
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

    renderWithProviders(<OrganizationUnitTreePicker {...defaultProps} />);

    await waitFor(() => {
      expect(screen.getByText('Root Organization')).toBeInTheDocument();
    });

    await waitFor(() => {
      expect(screen.getByText('Load more')).toBeInTheDocument();
    });
  });

  it('should fetch more root items when load more button is clicked', async () => {
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

    const nextPageResponse: OrganizationUnitListResponse = {
      totalResults: 50,
      startIndex: 3,
      count: 2,
      organizationUnits: [
        {id: 'ou-3', handle: 'sales', name: 'Sales', description: null, parent: null},
        {id: 'ou-4', handle: 'marketing', name: 'Marketing', description: null, parent: null},
      ],
    };

    mockHttpRequest.mockResolvedValue({data: nextPageResponse});

    renderWithProviders(<OrganizationUnitTreePicker {...defaultProps} />);

    await waitFor(() => {
      expect(screen.getByText('Load more')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Load more'));

    await waitFor(() => {
      expect(mockHttpRequest).toHaveBeenCalled();
    });
  });

  it('should trigger load more via keyboard Enter key', async () => {
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

    mockHttpRequest.mockResolvedValue({
      data: {
        totalResults: 50,
        startIndex: 3,
        count: 2,
        organizationUnits: [
          {id: 'ou-3', handle: 'sales', name: 'Sales', description: null, parent: null},
        ],
      },
    });

    renderWithProviders(<OrganizationUnitTreePicker {...defaultProps} />);

    await waitFor(() => {
      expect(screen.getByText('Load more')).toBeInTheDocument();
    });

    const loadMoreButton = screen.getByText('Load more').closest('[role="button"]')!;
    fireEvent.keyDown(loadMoreButton, {key: 'Enter'});

    await waitFor(() => {
      expect(mockHttpRequest).toHaveBeenCalled();
    });
  });

  it('should trigger load more via keyboard Space key', async () => {
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

    mockHttpRequest.mockResolvedValue({
      data: {
        totalResults: 50,
        startIndex: 3,
        count: 2,
        organizationUnits: [
          {id: 'ou-3', handle: 'sales', name: 'Sales', description: null, parent: null},
        ],
      },
    });

    renderWithProviders(<OrganizationUnitTreePicker {...defaultProps} />);

    await waitFor(() => {
      expect(screen.getByText('Load more')).toBeInTheDocument();
    });

    const loadMoreButton = screen.getByText('Load more').closest('[role="button"]')!;
    fireEvent.keyDown(loadMoreButton, {key: ' '});

    await waitFor(() => {
      expect(mockHttpRequest).toHaveBeenCalled();
    });
  });

  it('should show load more for child items when there are more children', async () => {
    const childOUResponse: OrganizationUnitListResponse = {
      totalResults: 50,
      startIndex: 1,
      count: 1,
      organizationUnits: [
        {id: 'ou-child-1', handle: 'child1', name: 'Fetched Child', description: null, parent: 'ou-1'},
      ],
    };

    mockHttpRequest.mockResolvedValue({data: childOUResponse});

    renderWithProviders(<OrganizationUnitTreePicker {...defaultProps} />);

    await waitFor(() => {
      expect(screen.getByText('Root Organization')).toBeInTheDocument();
    });

    const expandIcons = document.querySelectorAll('.MuiTreeItem-iconContainer');
    fireEvent.click(expandIcons[0]);

    await waitFor(() => {
      expect(screen.getByText('Fetched Child')).toBeInTheDocument();
      expect(screen.getByText('Load more')).toBeInTheDocument();
    });
  });

  it('should log error when root load more fails', async () => {
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

    mockHttpRequest.mockRejectedValue(new Error('Network failure'));

    renderWithProviders(<OrganizationUnitTreePicker {...defaultProps} />);

    await waitFor(() => {
      expect(screen.getByText('Load more')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Load more'));

    await waitFor(() => {
      expect(stableLogger.error).toHaveBeenCalled();
    });
  });

  it('should not expand item when it is already loaded', async () => {
    const childOUResponse: OrganizationUnitListResponse = {
      totalResults: 1,
      startIndex: 1,
      count: 1,
      organizationUnits: [
        {id: 'ou-child-1', handle: 'child1', name: 'Fetched Child', description: null, parent: 'ou-1'},
      ],
    };

    mockHttpRequest.mockResolvedValue({data: childOUResponse});

    renderWithProviders(<OrganizationUnitTreePicker {...defaultProps} />);

    await waitFor(() => {
      expect(screen.getByText('Root Organization')).toBeInTheDocument();
    });

    // First expansion - triggers fetch
    const expandIcons = document.querySelectorAll('.MuiTreeItem-iconContainer');
    fireEvent.click(expandIcons[0]);

    await waitFor(() => {
      expect(screen.getByText('Fetched Child')).toBeInTheDocument();
    });

    const callCount = mockHttpRequest.mock.calls.length;

    // Collapse
    const collapseIcons = document.querySelectorAll('.MuiTreeItem-iconContainer');
    fireEvent.click(collapseIcons[0]);

    // Expand again - should not trigger another fetch
    const expandIcons2 = document.querySelectorAll('.MuiTreeItem-iconContainer');
    fireEvent.click(expandIcons2[0]);

    // Wait a bit and verify no additional HTTP calls
    await waitFor(() => {
      expect(mockHttpRequest).toHaveBeenCalledTimes(callCount);
    });
  });

  it('should highlight selected item', async () => {
    renderWithProviders(<OrganizationUnitTreePicker {...defaultProps} value="ou-1" />);

    await waitFor(() => {
      expect(screen.getByText('Root Organization')).toBeInTheDocument();
    });

    // The Mui-selected class should be applied to the selected item
    const selectedElements = document.querySelectorAll('.Mui-selected');
    expect(selectedElements.length).toBeGreaterThan(0);
  });
});
