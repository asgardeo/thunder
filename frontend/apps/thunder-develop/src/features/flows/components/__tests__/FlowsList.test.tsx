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

import React from 'react';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render} from '@thunder/test-utils/browser';
import {page, userEvent} from 'vitest/browser';
import {MemoryRouter} from 'react-router';
import {DataGrid} from '@wso2/oxygen-ui';
import FlowsList from '../FlowsList';
import type {BasicFlowDefinition} from '../../models/responses';

// Use vi.hoisted so mockLoggerError is available inside vi.mock factories
const {mockLoggerError} = vi.hoisted(() => ({
  mockLoggerError: vi.fn(),
}));

// Mock @thunder/logger/react with accessible mock functions
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

// Use vi.hoisted for captured variables so they're available in the mock
const {capturedColumns, capturedOnRowClick} = vi.hoisted(() => ({
  capturedColumns: {value: [] as DataGrid.GridColDef<BasicFlowDefinition>[]},
  capturedOnRowClick: {value: undefined as ((params: {row: {id: string; flowType: string}}) => void) | undefined},
}));

// Mock DataGrid - captures columns for testing
vi.mock('@wso2/oxygen-ui', async () => {
  const actual = await vi.importActual('@wso2/oxygen-ui');
  return {
    ...actual,
    DataGrid: {
      ...((actual as {DataGrid?: Record<string, unknown>}).DataGrid ?? {}),
      GridColDef: {},
      GridRenderCellParams: {},
    },
    ListingTable: {
      Provider: ({children, loading}: {children: React.ReactNode; loading?: boolean}) => (
        <div data-testid="listing-table-provider" data-loading={loading ? 'true' : 'false'}>
          {children}
        </div>
      ),
      Container: ({children}: {children: React.ReactNode}): React.ReactElement => children as React.ReactElement,
      DataGrid: ({rows, columns, loading, onRowClick, getRowId}: MockDataGridProps) => {
        // Capture columns and onRowClick for testing
        capturedColumns.value = columns as unknown as DataGrid.GridColDef<BasicFlowDefinition>[];
        capturedOnRowClick.value = onRowClick as typeof capturedOnRowClick.value;
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
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        );
      },
      CellIcon: ({primary, icon}: {primary: string; icon?: React.ReactNode}) => (
        <>
          {icon}
          <span>{primary}</span>
        </>
      ),
      RowActions: ({children}: {children: React.ReactNode}): React.ReactElement => children as React.ReactElement,
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
    it('should render DataGrid component', async () => {
      await render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      await expect.element(page.getByTestId('data-grid')).toBeInTheDocument();
    });

    it('should render column headers', async () => {
      await render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      await expect.element(page.getByText('Name')).toBeInTheDocument();
      await expect.element(page.getByText('Type')).toBeInTheDocument();
      await expect.element(page.getByText('Version')).toBeInTheDocument();
      await expect.element(page.getByText('Last Updated')).toBeInTheDocument();
      // Actions appears multiple times (header + rows), use getAllByText
      expect(page.getByText('Actions').length).toBeGreaterThanOrEqual(1);
    });

    it('should render flow data in rows', async () => {
      await render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      // Only AUTHENTICATION flows are shown (HIDE_NON_EDITABLE_FLOWS = true)
      await expect.element(page.getByText('Login Flow')).toBeInTheDocument();
      await expect.element(page.getByText('AUTHENTICATION')).toBeInTheDocument();
      // REGISTRATION flows are filtered out
      await expect.element(page.getByText('Registration Flow')).not.toBeInTheDocument();
    });
  });

  describe('Loading State', () => {
    it('should pass loading state to DataGrid', async () => {
      mockUseGetFlowsReturn = {
        data: null as unknown as {flows: BasicFlowDefinition[]},
        isLoading: true,
        error: null,
      };

      await render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      await expect.element(page.getByTestId('listing-table-provider')).toHaveAttribute('data-loading', 'true');
    });
  });

  describe('Error State', () => {
    it('should display error message when error occurs', async () => {
      mockUseGetFlowsReturn = {
        data: null as unknown as {flows: BasicFlowDefinition[]},
        isLoading: false,
        error: new Error('Failed to fetch flows'),
      };

      await render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      await expect.element(page.getByText('Failed to load flows')).toBeInTheDocument();
      await expect.element(page.getByText('Failed to fetch flows')).toBeInTheDocument();
    });

    it('should display unknown error message when error has no message', async () => {
      mockUseGetFlowsReturn = {
        data: null as unknown as {flows: BasicFlowDefinition[]},
        isLoading: false,
        error: {} as Error,
      };

      await render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      await expect.element(page.getByText('Failed to load flows')).toBeInTheDocument();
      await expect.element(page.getByText('An unknown error occurred')).toBeInTheDocument();
    });
  });

  describe('Row Click Navigation', () => {
    it('should navigate to flow page when authentication flow row is clicked', async () => {
      await render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const authFlowRow = page.getByTestId('row-flow-1');
      await userEvent.click(authFlowRow);

      await vi.waitFor(() => {
        expect(mockNavigate).toHaveBeenCalledWith('/flows/signin/flow-1');
      });
    });

    it('should not navigate when non-authentication flow row is clicked', async () => {
      await render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      // REGISTRATION flows are filtered from the table (HIDE_NON_EDITABLE_FLOWS = true)
      // Only AUTHENTICATION flow row is rendered
      await expect.element(page.getByTestId('row-flow-2')).not.toBeInTheDocument();
      // Verify only authentication row is present and no navigation has occurred
      await expect.element(page.getByTestId('row-flow-1')).toBeInTheDocument();
      expect(mockNavigate).not.toHaveBeenCalled();
    });
  });

  describe('Empty State', () => {
    it('should render empty table when no flows exist', async () => {
      mockUseGetFlowsReturn = {
        data: {flows: []},
        isLoading: false,
        error: null,
      };

      await render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      await expect.element(page.getByTestId('data-grid')).toBeInTheDocument();
      await expect.element(page.getByTestId('row-flow-1')).not.toBeInTheDocument();
    });

    it('should handle undefined data gracefully', async () => {
      mockUseGetFlowsReturn = {
        data: undefined as unknown as {flows: BasicFlowDefinition[]},
        isLoading: false,
        error: null,
      };

      await render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      await expect.element(page.getByTestId('data-grid')).toBeInTheDocument();
    });
  });

  describe('Actions Menu', () => {
    it('should have actions renderCell for authentication rows', async () => {
      await render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const actionsColumn = capturedColumns.value.find((col) => col.field === 'actions');
      expect(actionsColumn?.renderCell).toBeDefined();

      if (actionsColumn?.renderCell) {
        // AUTHENTICATION flow should render buttons
        const {container: authContainer} = await render(
          actionsColumn.renderCell({row: mockFlowsData.flows[0]} as DataGrid.GridRenderCellParams<BasicFlowDefinition>) as React.ReactElement,
        );
        expect(authContainer.querySelectorAll('button').length).toBeGreaterThan(0);

        // REGISTRATION flow should render nothing
        const result = actionsColumn.renderCell({
          row: mockFlowsData.flows[1],
        } as DataGrid.GridRenderCellParams<BasicFlowDefinition>);
        expect(result).toBeNull();
      }
    });
  });

  describe('Navigation Error Handling', () => {
    it('should not navigate when a non-AUTHENTICATION flow row is clicked', async () => {
      await render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      // Directly invoke onRowClick with a REGISTRATION flow, bypassing the
      // HIDE_NON_EDITABLE_FLOWS filter that prevents such rows being rendered.
      // This covers the early-return guard in handleEditClick (FlowsList.tsx line 60).
      capturedOnRowClick.value?.({row: {id: 'flow-2', flowType: 'REGISTRATION'}});

      expect(mockNavigate).not.toHaveBeenCalled();
    });

    it('should handle navigation errors gracefully', async () => {
      mockNavigate.mockRejectedValue(new Error('Navigation failed'));

      await render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const authFlowRow = page.getByTestId('row-flow-1');
      await userEvent.click(authFlowRow);

      // Verify navigation was attempted and error was logged
      await vi.waitFor(() => {
        expect(mockNavigate).toHaveBeenCalled();
        expect(mockLoggerError).toHaveBeenCalledWith(
          'Failed to navigate to flow builder',
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
    it('should apply different cursor style based on flow type', async () => {
      await render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const authFlowRow = page.getByTestId('row-flow-1');
      // Authentication flows should have pointer cursor
      expect(authFlowRow).toHaveStyle({cursor: 'pointer'});
      // REGISTRATION flow row is not rendered (HIDE_NON_EDITABLE_FLOWS = true)
      await expect.element(page.getByTestId('row-flow-2')).not.toBeInTheDocument();
    });
  });

  describe('Version Display', () => {
    it('should display version numbers with v prefix', async () => {
      await render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      // Only AUTHENTICATION flow is shown (HIDE_NON_EDITABLE_FLOWS = true)
      await expect.element(page.getByText('v1')).toBeInTheDocument();
      // flow-2 (REGISTRATION) is filtered out so v2 is not in the DOM
      await expect.element(page.getByText('v2')).not.toBeInTheDocument();
    });
  });

  describe('Column RenderCell Functions', () => {
    it('should capture column definitions', async () => {
      await render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      // Verify columns are captured
      expect(capturedColumns.value.length).toBeGreaterThan(0);

      // Verify expected columns exist
      const nameColumn = capturedColumns.value.find((col) => col.field === 'name');
      const flowTypeColumn = capturedColumns.value.find((col) => col.field === 'flowType');
      const versionColumn = capturedColumns.value.find((col) => col.field === 'activeVersion');
      const updatedAtColumn = capturedColumns.value.find((col) => col.field === 'updatedAt');
      const actionsColumn = capturedColumns.value.find((col) => col.field === 'actions');

      expect(nameColumn).toBeDefined();
      expect(flowTypeColumn).toBeDefined();
      expect(versionColumn).toBeDefined();
      expect(updatedAtColumn).toBeDefined();
      expect(actionsColumn).toBeDefined();
    });

    it('should have renderCell functions defined for columns', async () => {
      await render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const nameColumn = capturedColumns.value.find((col) => col.field === 'name');
      const flowTypeColumn = capturedColumns.value.find((col) => col.field === 'flowType');
      const versionColumn = capturedColumns.value.find((col) => col.field === 'activeVersion');
      const actionsColumn = capturedColumns.value.find((col) => col.field === 'actions');

      expect(nameColumn?.renderCell).toBeDefined();
      expect(flowTypeColumn?.renderCell).toBeDefined();
      expect(versionColumn?.renderCell).toBeDefined();
      expect(actionsColumn?.renderCell).toBeDefined();
    });

    it('should have valueGetter for updatedAt column', async () => {
      await render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const updatedAtColumn = capturedColumns.value.find((col) => col.field === 'updatedAt');
      expect(updatedAtColumn?.valueGetter).toBeDefined();

      if (updatedAtColumn?.valueGetter) {
        const formattedDate = (updatedAtColumn.valueGetter as (value: unknown, row: unknown) => string)(
          undefined,
          mockFlowsData.flows[0],
        );
        // Check that the formatted date contains expected parts
        expect(formattedDate).toContain('2025');
        expect(formattedDate).toContain('Jan');
      }
    });
  });

  describe('Menu Interactions', () => {
    it('should have actions column with renderCell defined', async () => {
      await render(
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
    it('should render delete dialog component', async () => {
      await render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      // Delete dialog is rendered but not visible initially
      await expect.element(page.getByTestId('flow-delete-dialog')).not.toBeInTheDocument();
    });
  });

  describe('Column RenderCell Execution', () => {
    it('should render name cell with GitBranch icon', async () => {
      await render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const nameColumn = capturedColumns.value.find((col) => col.field === 'name');
      expect(nameColumn?.renderCell).toBeDefined();

      if (nameColumn?.renderCell) {
        const {container} = await render(
          nameColumn.renderCell({row: mockFlowsData.flows[0]} as DataGrid.GridRenderCellParams<BasicFlowDefinition>) as React.ReactElement,
        );
        expect(container.querySelector('[class*="MuiAvatar"]')).toBeInTheDocument();
      }
    });

    it('should render actions cell with IconButton', async () => {
      await render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const actionsColumn = capturedColumns.value.find((col) => col.field === 'actions');
      expect(actionsColumn?.renderCell).toBeDefined();

      if (actionsColumn?.renderCell) {
        const {container} = await render(
          actionsColumn.renderCell({row: mockFlowsData.flows[0]} as DataGrid.GridRenderCellParams<BasicFlowDefinition>) as React.ReactElement,
        );
        // Authentication flows should render action buttons (Pencil + Trash2)
        const buttons = container.querySelectorAll('button');
        expect(buttons.length).toBeGreaterThan(0);
      }
    });

    it('should not throw when action button is clicked in renderCell', async () => {
      await render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const actionsColumn = capturedColumns.value.find((col) => col.field === 'actions');

      if (actionsColumn?.renderCell) {
        const {container} = await render(
          actionsColumn.renderCell({row: mockFlowsData.flows[0]} as DataGrid.GridRenderCellParams<BasicFlowDefinition>) as React.ReactElement,
        );
        const buttons = container.querySelectorAll('button');
        expect(buttons.length).toBeGreaterThan(0);

        // Click should not throw
        expect(async () => userEvent.click(buttons[0])).not.toThrow();
      }
    });
  });

  describe('Menu Handler Functions', () => {
    it('should open delete dialog when Trash2 button is clicked', async () => {
      await render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const actionsColumn = capturedColumns.value.find((col) => col.field === 'actions');
      if (actionsColumn?.renderCell) {
        const {container} = await render(
          actionsColumn.renderCell({row: mockFlowsData.flows[0]} as DataGrid.GridRenderCellParams<BasicFlowDefinition>) as React.ReactElement,
        );

        // Find the error (delete) button - Trash2
        const deleteButton = container.querySelector('button.MuiIconButton-colorError');
        expect(deleteButton).toBeInTheDocument();
        await userEvent.click(deleteButton!);

        await vi.waitFor(async () => {
          await expect.element(page.getByTestId('flow-delete-dialog')).toBeInTheDocument();
          await expect.element(page.getByTestId('flow-delete-dialog')).toHaveAttribute('data-flow-id', 'flow-1');
        });
      }
    });

    it('should close delete dialog and reset selected flow', async () => {
      await render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const actionsColumn = capturedColumns.value.find((col) => col.field === 'actions');
      if (actionsColumn?.renderCell) {
        const {container} = await render(
          actionsColumn.renderCell({row: mockFlowsData.flows[0]} as DataGrid.GridRenderCellParams<BasicFlowDefinition>) as React.ReactElement,
        );

        const deleteButton = container.querySelector('button.MuiIconButton-colorError');
        expect(deleteButton).toBeInTheDocument();
        await userEvent.click(deleteButton!);

        await vi.waitFor(async () => {
          await expect.element(page.getByTestId('flow-delete-dialog')).toBeInTheDocument();
        });

        await userEvent.click(page.getByText('Close'));

        await vi.waitFor(async () => {
          await expect.element(page.getByTestId('flow-delete-dialog')).not.toBeInTheDocument();
        });
      }
    });

    it('should navigate to flow builder when Pencil button is clicked for authentication flow', async () => {
      mockNavigate.mockResolvedValue(undefined);

      await render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const actionsColumn = capturedColumns.value.find((col) => col.field === 'actions');
      if (actionsColumn?.renderCell) {
        const {container} = await render(
          actionsColumn.renderCell({row: mockFlowsData.flows[0]} as DataGrid.GridRenderCellParams<BasicFlowDefinition>) as React.ReactElement,
        );

        // First button is Pencil (edit)
        const buttons = container.querySelectorAll('button');
        expect(buttons.length).toBeGreaterThanOrEqual(1);
        await userEvent.click(buttons[0]);

        await vi.waitFor(() => {
          expect(mockNavigate).toHaveBeenCalledWith('/flows/signin/flow-1');
        });
      }
    });

    it('should log error when navigation fails', async () => {
      mockNavigate.mockRejectedValue(new Error('Navigation failed'));

      await render(
        <MemoryRouter>
          <FlowsList />
        </MemoryRouter>,
      );

      const actionsColumn = capturedColumns.value.find((col) => col.field === 'actions');
      if (actionsColumn?.renderCell) {
        const {container} = await render(
          actionsColumn.renderCell({row: mockFlowsData.flows[0]} as DataGrid.GridRenderCellParams<BasicFlowDefinition>) as React.ReactElement,
        );

        const buttons = container.querySelectorAll('button');
        await userEvent.click(buttons[0]);

        await vi.waitFor(() => {
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
