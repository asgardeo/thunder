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
import FlowsList from '../FlowsList';
import type {BasicFlowDefinition} from '../../models/responses';

// Mock @thunder/logger/react
vi.mock('@thunder/logger/react', () => ({
  useLogger: () => ({
    debug: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
    withComponent: vi.fn(() => ({
      debug: vi.fn(),
      info: vi.fn(),
      warn: vi.fn(),
      error: vi.fn(),
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

// Mock DataGrid
vi.mock('@wso2/oxygen-ui', async () => {
  const actual = await vi.importActual('@wso2/oxygen-ui');
  return {
    ...actual,
    DataGrid: {
      DataGrid: ({rows, columns, loading, onRowClick, getRowId}: MockDataGridProps) => (
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
      ),
    },
    useTheme: () => ({
      vars: {
        palette: {
          grey: {500: '#9e9e9e', 900: '#212121'},
          error: {main: '#f44336'},
        },
      },
      applyStyles: () => ({}),
    }),
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
});
