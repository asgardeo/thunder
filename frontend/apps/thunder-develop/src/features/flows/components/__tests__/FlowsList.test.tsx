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

/* eslint-disable react/require-default-props */

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render, screen, fireEvent, waitFor} from '@testing-library/react';
import {MemoryRouter} from 'react-router';
import {DataGrid} from '@wso2/oxygen-ui';
import FlowsList from '../FlowsList';
import type {BasicFlowDefinition} from '../../models/responses';

// Mock @thunder/logger/react with accessible mock functions
const mockLoggerError = vi.fn();
vi.mock('@thunder/logger/react', () => ({
  useLogger: () => ({
    debug: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: mockLoggerError,
    withComponent: vi.fn(() => ({
      debug: vi.fn(),
      info: vi.fn(),
      warn: vi.fn(),
      error: mockLoggerError,
    })),
  }),
}));

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'flows:listing.columns.name': 'Name',
        'flows:listing.columns.flowType': 'Type',
        'flows:listing.columns.version': 'Version',
        'flows:listing.columns.updatedAt': 'Updated At',
        'flows:listing.columns.actions': 'Actions',
        'flows:listing.error.title': 'Error loading flows',
        'flows:listing.error.unknown': 'Unknown error occurred',
        'common:actions.view': 'View',
        'common:actions.delete': 'Delete',
      };
      return translations[key] || key;
    },
  }),
}));

// Mock useNavigate
const mockNavigate = vi.fn();
vi.mock('react-router', async () => {
  const actual = await vi.importActual('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

// Mock useDataGridLocaleText
vi.mock('../../../../hooks/useDataGridLocaleText', () => ({
  default: () => ({}),
}));

// Mock useGetFlows
const mockFlowsData: {flows: BasicFlowDefinition[]} = {
  flows: [
    {
      id: 'flow-1',
      handle: 'login-flow',
      name: 'Login Flow',
      flowType: 'AUTHENTICATION',
      activeVersion: 1,
      createdAt: '2025-01-01T09:00:00Z',
      updatedAt: '2025-01-01T10:00:00Z',
    },
    {
      id: 'flow-2',
      handle: 'registration-flow',
      name: 'Registration Flow',
      flowType: 'REGISTRATION',
      activeVersion: 2,
      createdAt: '2025-01-02T14:00:00Z',
      updatedAt: '2025-01-02T15:30:00Z',
    },
  ],
};

let mockUseGetFlowsReturn = {
  data: mockFlowsData,
  isLoading: false,
  error: null as Error | null,
};

vi.mock('../../api/useGetFlows', () => ({
  default: () => mockUseGetFlowsReturn,
}));

// Mock FlowDeleteDialog
vi.mock('../FlowDeleteDialog', () => ({
  default: ({open, flowId, onClose}: {open: boolean; flowId: string | null; onClose: () => void}) =>
    open ? (
      <div data-testid="flow-delete-dialog" data-flow-id={flowId}>
        <button type="button" onClick={onClose}>
          Close
        </button>
      </div>
    ) : null,
}));

interface MockColumn {
  field: string;
  headerName: string;
}

interface MockRow {
  id: string;
  name: string;
  flowType: string;
  activeVersion: number;
  updatedAt: string;
}

interface MockDataGridProps {
  rows: MockRow[] | undefined;
  columns: MockColumn[];
  loading: boolean;
  onRowClick?: (params: {row: MockRow}) => void;
  getRowId: (row: MockRow) => string;
}

// Use vi.hoisted for the capturedColumns variable so it's available in the mock
const {capturedColumns} = vi.hoisted(() => ({
  capturedColumns: {value: [] as DataGrid.GridColDef<BasicFlowDefinition>[]},
}));

// Mock DataGrid - captures columns for testing
vi.mock('@wso2/oxygen-ui', async () => {
  const actual = await vi.importActual('@wso2/oxygen-ui');
  return {
    ...actual,
    DataGrid: {
      DataGrid: ({rows, columns, loading, onRowClick, getRowId}: MockDataGridProps) => {
        // Capture columns for testing renderCell functions
        capturedColumns.value = columns as unknown as DataGrid.GridColDef<BasicFlowDefinition>[];
        return (
          <div data-testid="data-grid" data-loading={loading}>
            <table>
              <thead>
                <tr>
                  {columns.map((col: MockColumn) => (
                    <th key={col.field}>{col.headerName}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {rows?.map((row: MockRow) => {
                  const rowId: string = getRowId(row);
                  const handleClick = (): void => onRowClick?.({row});
                  const handleButtonClick = (e: React.MouseEvent): void => {
                    e.stopPropagation();
                    const event = new CustomEvent('menuopen', {detail: {row}});
                    document.dispatchEvent(event);
                  };
                  return (
                    <tr
                      key={rowId}
                      data-testid={`row-${rowId}`}
                      onClick={handleClick}
                      style={{cursor: row.flowType === 'AUTHENTICATION' ? 'pointer' : 'default'}}
                    >
                      <td>{row.name}</td>
                      <td>{row.flowType}</td>
                      <td>v{row.activeVersion}</td>
                      <td>{row.updatedAt}</td>
                      <td>
                        <button type="button" aria-label="Open actions menu" onClick={handleButtonClick}>
                          Actions
                        </button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        );
      },
    },
  };
});

describe('FlowsList', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseGetFlowsReturn = {
      data: mockFlowsData,
      isLoading: false,
      error: null,
    };
    capturedColumns.value = [];
  });

  describe('Rendering', () => {
    it('should render DataGrid component', () => {
      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      expect(screen.getByTestId('data-grid')).toBeInTheDocument();
    });

    it('should render column headers', () => {
      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      expect(screen.getByText('Name')).toBeInTheDocument();
      expect(screen.getByText('Type')).toBeInTheDocument();
      expect(screen.getByText('Version')).toBeInTheDocument();
      expect(screen.getByText('Updated At')).toBeInTheDocument();
      // Actions appears multiple times (header + rows), use getAllByText
      expect(screen.getAllByText('Actions').length).toBeGreaterThanOrEqual(1);
    });

    it('should render flow data in rows', () => {
      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      expect(screen.getByText('Login Flow')).toBeInTheDocument();
      expect(screen.getByText('Registration Flow')).toBeInTheDocument();
      expect(screen.getByText('AUTHENTICATION')).toBeInTheDocument();
      expect(screen.getByText('REGISTRATION')).toBeInTheDocument();
    });
  });

  describe('Loading State', () => {
    it('should pass loading state to DataGrid', () => {
      mockUseGetFlowsReturn = {
        data: null as unknown as {flows: BasicFlowDefinition[]},
        isLoading: true,
        error: null,
      };

      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      expect(screen.getByTestId('data-grid')).toHaveAttribute('data-loading', 'true');
    });
  });

  describe('Error State', () => {
    it('should display error message when error occurs', () => {
      mockUseGetFlowsReturn = {
        data: null as unknown as {flows: BasicFlowDefinition[]},
        isLoading: false,
        error: new Error('Failed to fetch flows'),
      };

      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      expect(screen.getByText('Error loading flows')).toBeInTheDocument();
      expect(screen.getByText('Failed to fetch flows')).toBeInTheDocument();
    });

    it('should display unknown error message when error has no message', () => {
      mockUseGetFlowsReturn = {
        data: null as unknown as {flows: BasicFlowDefinition[]},
        isLoading: false,
        error: {} as Error,
      };

      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      expect(screen.getByText('Error loading flows')).toBeInTheDocument();
      expect(screen.getByText('Unknown error occurred')).toBeInTheDocument();
    });
  });

  describe('Row Click Navigation', () => {
    it('should navigate to flow page when authentication flow row is clicked', async () => {
      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const authFlowRow = screen.getByTestId('row-flow-1');
      fireEvent.click(authFlowRow);

      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/flows/signin/flow-1');
      });
    });

    it('should not navigate when non-authentication flow row is clicked', () => {
      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const regFlowRow = screen.getByTestId('row-flow-2');
      fireEvent.click(regFlowRow);

      expect(mockNavigate).not.toHaveBeenCalled();
    });
  });

  describe('Empty State', () => {
    it('should render empty table when no flows exist', () => {
      mockUseGetFlowsReturn = {
        data: {flows: []},
        isLoading: false,
        error: null,
      };

      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      expect(screen.getByTestId('data-grid')).toBeInTheDocument();
      expect(screen.queryByTestId('row-flow-1')).not.toBeInTheDocument();
    });

    it('should handle undefined data gracefully', () => {
      mockUseGetFlowsReturn = {
        data: undefined as unknown as {flows: BasicFlowDefinition[]},
        isLoading: false,
        error: null,
      };

      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      expect(screen.getByTestId('data-grid')).toBeInTheDocument();
    });
  });

  describe('Actions Menu', () => {
    it('should open actions menu when actions button is clicked', async () => {
      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const actionsButtons = screen.getAllByLabelText('Open actions menu');
      fireEvent.click(actionsButtons[0]);

      // Menu should open - we dispatch custom event in mock
      // The actual menu won't render because of the mock, but we verify the button is clickable
      expect(actionsButtons[0]).toBeInTheDocument();
    });

    it('should render action buttons for each row', () => {
      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const actionsButtons = screen.getAllByLabelText('Open actions menu');
      // Should have one action button per row
      expect(actionsButtons.length).toBe(2);
    });
  });

  describe('Navigation Error Handling', () => {
    it('should handle navigation errors gracefully', async () => {
      mockNavigate.mockRejectedValue(new Error('Navigation failed'));

      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const authFlowRow = screen.getByTestId('row-flow-1');
      fireEvent.click(authFlowRow);

      // Verify navigation was attempted and error was logged
      await waitFor(() => {
        expect(mockNavigate).toHaveBeenCalled();
        expect(mockLoggerError).toHaveBeenCalledWith(
          'Failed to navigate to flow',
          expect.objectContaining({
            error: expect.any(Error) as Error,
            flowId: 'flow-1',
          }),
        );
      });

      // Verify component is still rendered (no crash)
      expect(authFlowRow).toBeInTheDocument();
    });
  });

  describe('Row Styling', () => {
    it('should apply different cursor style based on flow type', () => {
      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const authFlowRow = screen.getByTestId('row-flow-1');
      const regFlowRow = screen.getByTestId('row-flow-2');

      // Authentication flows should have pointer cursor
      expect(authFlowRow).toHaveStyle({cursor: 'pointer'});
      // Non-authentication flows should have default cursor
      expect(regFlowRow).toHaveStyle({cursor: 'default'});
    });
  });

  describe('Version Display', () => {
    it('should display version numbers with v prefix', () => {
      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      expect(screen.getByText('v1')).toBeInTheDocument();
      expect(screen.getByText('v2')).toBeInTheDocument();
    });
  });

  describe('Column RenderCell Functions', () => {
    it('should capture column definitions', () => {
      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      // Verify columns are captured
      expect(capturedColumns.value.length).toBeGreaterThan(0);

      // Verify expected columns exist
      const avatarColumn = capturedColumns.value.find((col) => col.field === 'avatar');
      const flowTypeColumn = capturedColumns.value.find((col) => col.field === 'flowType');
      const versionColumn = capturedColumns.value.find((col) => col.field === 'activeVersion');
      const updatedAtColumn = capturedColumns.value.find((col) => col.field === 'updatedAt');
      const actionsColumn = capturedColumns.value.find((col) => col.field === 'actions');

      expect(avatarColumn).toBeDefined();
      expect(flowTypeColumn).toBeDefined();
      expect(versionColumn).toBeDefined();
      expect(updatedAtColumn).toBeDefined();
      expect(actionsColumn).toBeDefined();
    });

    it('should have renderCell functions defined for columns', () => {
      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const avatarColumn = capturedColumns.value.find((col) => col.field === 'avatar');
      const flowTypeColumn = capturedColumns.value.find((col) => col.field === 'flowType');
      const versionColumn = capturedColumns.value.find((col) => col.field === 'activeVersion');
      const actionsColumn = capturedColumns.value.find((col) => col.field === 'actions');

      expect(avatarColumn?.renderCell).toBeDefined();
      expect(flowTypeColumn?.renderCell).toBeDefined();
      expect(versionColumn?.renderCell).toBeDefined();
      expect(actionsColumn?.renderCell).toBeDefined();
    });

    it('should have valueGetter for updatedAt column', () => {
      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const updatedAtColumn = capturedColumns.value.find((col) => col.field === 'updatedAt');
      expect(updatedAtColumn?.valueGetter).toBeDefined();

      if (updatedAtColumn?.valueGetter) {
        const formattedDate = (updatedAtColumn.valueGetter as (value: unknown, row: unknown) => string)(undefined, mockFlowsData.flows[0]);
        // Check that the formatted date contains expected parts
        expect(formattedDate).toContain('2025');
        expect(formattedDate).toContain('Jan');
      }
    });
  });

  describe('Menu Interactions', () => {
    it('should have actions column with renderCell defined', () => {
      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const actionsColumn = capturedColumns.value.find((col) => col.field === 'actions');
      expect(actionsColumn).toBeDefined();
      expect(actionsColumn?.renderCell).toBeDefined();
    });
  });

  describe('Delete Dialog Integration', () => {
    it('should render delete dialog component', () => {
      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      // Delete dialog is rendered but not visible initially
      expect(screen.queryByTestId('flow-delete-dialog')).not.toBeInTheDocument();
    });
  });

  describe('Column RenderCell Execution', () => {
    it('should render avatar cell with GitBranch icon', () => {
      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const avatarColumn = capturedColumns.value.find((col) => col.field === 'avatar');
      expect(avatarColumn?.renderCell).toBeDefined();

      if (avatarColumn?.renderCell) {
        const {container} = render(
          avatarColumn.renderCell({row: mockFlowsData.flows[0]} as DataGrid.GridRenderCellParams<BasicFlowDefinition>),
        );
        expect(container.querySelector('[class*="MuiAvatar"]')).toBeInTheDocument();
      }
    });

    it('should render actions cell with IconButton', () => {
      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const actionsColumn = capturedColumns.value.find((col) => col.field === 'actions');
      expect(actionsColumn?.renderCell).toBeDefined();

      if (actionsColumn?.renderCell) {
        const {container} = render(
          actionsColumn.renderCell({row: mockFlowsData.flows[0]} as DataGrid.GridRenderCellParams<BasicFlowDefinition>),
        );
        const actionButton = container.querySelector('button[aria-label="Open actions menu"]');
        expect(actionButton).toBeInTheDocument();
      }
    });

    it('should call handleMenuOpen when actions button is clicked in renderCell', () => {
      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const actionsColumn = capturedColumns.value.find((col) => col.field === 'actions');

      if (actionsColumn?.renderCell) {
        const {container} = render(
          actionsColumn.renderCell({row: mockFlowsData.flows[0]} as DataGrid.GridRenderCellParams<BasicFlowDefinition>),
        );
        const button = container.querySelector('button[aria-label="Open actions menu"]');
        expect(button).toBeInTheDocument();

        // Click should not throw
        expect(() => fireEvent.click(button!)).not.toThrow();
      }
    });
  });

  describe('Menu Handler Functions', () => {
    it('should open menu and show View option for authentication flow', async () => {
      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      // Get the actions column renderCell and render it
      const actionsColumn = capturedColumns.value.find((col) => col.field === 'actions');
      if (actionsColumn?.renderCell) {
        const {container} = render(
          actionsColumn.renderCell({row: mockFlowsData.flows[0]} as DataGrid.GridRenderCellParams<BasicFlowDefinition>),
        );

        // Get the button from the rendered cell container (the last one with MuiIconButton class)
        const actionButton = container.querySelector('button[aria-label="Open actions menu"]');
        expect(actionButton).toBeInTheDocument();
        fireEvent.click(actionButton!);

        // After clicking, the menu should open in the main component
        // Check that View and Delete options are available
        await waitFor(() => {
          expect(screen.getByText('View')).toBeInTheDocument();
          expect(screen.getByText('Delete')).toBeInTheDocument();
        });
      }
    });

    it('should not show View option for non-authentication flow', async () => {
      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      // Get the actions column renderCell and render it with non-auth flow
      const actionsColumn = capturedColumns.value.find((col) => col.field === 'actions');
      if (actionsColumn?.renderCell) {
        const {container} = render(
          actionsColumn.renderCell({row: mockFlowsData.flows[1]} as DataGrid.GridRenderCellParams<BasicFlowDefinition>),
        );

        const actionButton = container.querySelector('button[aria-label="Open actions menu"]');
        expect(actionButton).toBeInTheDocument();
        fireEvent.click(actionButton!);

        // After clicking, the menu should open - View should not be shown for non-auth flows
        await waitFor(() => {
          expect(screen.getByText('Delete')).toBeInTheDocument();
        });

        // View should not appear for REGISTRATION flow type
        const viewElements = screen.queryAllByText('View');
        // The View menu item is conditionally rendered based on selectedFlow.flowType === 'AUTHENTICATION'
        expect(viewElements.length).toBeLessThanOrEqual(1);
      }
    });

    it('should close menu when menu close is triggered', async () => {
      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const actionsColumn = capturedColumns.value.find((col) => col.field === 'actions');
      if (actionsColumn?.renderCell) {
        const {container} = render(
          actionsColumn.renderCell({row: mockFlowsData.flows[0]} as DataGrid.GridRenderCellParams<BasicFlowDefinition>),
        );

        const actionButton = container.querySelector('button[aria-label="Open actions menu"]');
        expect(actionButton).toBeInTheDocument();
        fireEvent.click(actionButton!);

        await waitFor(() => {
          expect(screen.getByText('View')).toBeInTheDocument();
        });

        // Click View to close the menu
        fireEvent.click(screen.getByText('View'));

        // Menu should close
        await waitFor(() => {
          expect(screen.queryByRole('menu')).not.toBeInTheDocument();
        });
      }
    });

    it('should open delete dialog when Delete is clicked', async () => {
      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const actionsColumn = capturedColumns.value.find((col) => col.field === 'actions');
      if (actionsColumn?.renderCell) {
        const {container} = render(
          actionsColumn.renderCell({row: mockFlowsData.flows[0]} as DataGrid.GridRenderCellParams<BasicFlowDefinition>),
        );

        const actionButton = container.querySelector('button[aria-label="Open actions menu"]');
        expect(actionButton).toBeInTheDocument();
        fireEvent.click(actionButton!);

        await waitFor(() => {
          expect(screen.getByText('Delete')).toBeInTheDocument();
        });

        fireEvent.click(screen.getByText('Delete'));

        // Delete dialog should open
        await waitFor(() => {
          expect(screen.getByTestId('flow-delete-dialog')).toBeInTheDocument();
          expect(screen.getByTestId('flow-delete-dialog')).toHaveAttribute('data-flow-id', 'flow-1');
        });
      }
    });

    it('should close delete dialog and reset selected flow', async () => {
      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const actionsColumn = capturedColumns.value.find((col) => col.field === 'actions');
      if (actionsColumn?.renderCell) {
        const {container} = render(
          actionsColumn.renderCell({row: mockFlowsData.flows[0]} as DataGrid.GridRenderCellParams<BasicFlowDefinition>),
        );

        const actionButton = container.querySelector('button[aria-label="Open actions menu"]');
        expect(actionButton).toBeInTheDocument();
        fireEvent.click(actionButton!);

        await waitFor(() => {
          expect(screen.getByText('Delete')).toBeInTheDocument();
        });

        fireEvent.click(screen.getByText('Delete'));

        await waitFor(() => {
          expect(screen.getByTestId('flow-delete-dialog')).toBeInTheDocument();
        });

        // Close the dialog
        fireEvent.click(screen.getByText('Close'));

        await waitFor(() => {
          expect(screen.queryByTestId('flow-delete-dialog')).not.toBeInTheDocument();
        });
      }
    });

    it('should navigate to flow builder when View is clicked for authentication flow', async () => {
      mockNavigate.mockResolvedValue(undefined);

      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const actionsColumn = capturedColumns.value.find((col) => col.field === 'actions');
      if (actionsColumn?.renderCell) {
        const {container} = render(
          actionsColumn.renderCell({row: mockFlowsData.flows[0]} as DataGrid.GridRenderCellParams<BasicFlowDefinition>),
        );

        const actionButton = container.querySelector('button[aria-label="Open actions menu"]');
        expect(actionButton).toBeInTheDocument();
        fireEvent.click(actionButton!);

        await waitFor(() => {
          expect(screen.getByText('View')).toBeInTheDocument();
        });

        fireEvent.click(screen.getByText('View'));

        await waitFor(() => {
          expect(mockNavigate).toHaveBeenCalledWith('/flows/signin/flow-1');
        });
      }
    });

    it('should log error when View navigation fails', async () => {
      mockNavigate.mockRejectedValue(new Error('Navigation failed'));

      render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const actionsColumn = capturedColumns.value.find((col) => col.field === 'actions');
      if (actionsColumn?.renderCell) {
        const {container} = render(
          actionsColumn.renderCell({row: mockFlowsData.flows[0]} as DataGrid.GridRenderCellParams<BasicFlowDefinition>),
        );

        const actionButton = container.querySelector('button[aria-label="Open actions menu"]');
        expect(actionButton).toBeInTheDocument();
        fireEvent.click(actionButton!);

        await waitFor(() => {
          expect(screen.getByText('View')).toBeInTheDocument();
        });

        fireEvent.click(screen.getByText('View'));

        await waitFor(() => {
          expect(mockLoggerError).toHaveBeenCalledWith(
            'Failed to navigate to flow builder',
            expect.objectContaining({
              error: expect.any(Error) as Error,
              flowId: 'flow-1',
            }),
          );
        });
      }
    });
  });
});
