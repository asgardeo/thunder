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

import {describe, expect, it, vi, beforeEach} from 'vitest';
import {page, userEvent, renderWithProviders} from '@thunder/test-utils/browser';
import TranslationsList from '../TranslationsList';

const mockNavigate = vi.fn();
vi.mock('react-router', async () => {
  const actual = await vi.importActual<typeof import('react-router')>('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

vi.mock('../../../../hooks/useDataGridLocaleText', () => ({
  default: () => ({}),
}));

const mockMutate = vi.fn();
vi.mock('@thunder/i18n', () => ({
  useGetLanguages: vi.fn().mockReturnValue({
    data: {languages: ['fr-FR', 'de-DE']},
    isLoading: false,
  }),
  useDeleteTranslations: () => ({mutate: mockMutate, isPending: false}),
  getDisplayNameForCode: (code: string) => `Language(${code})`,
  toFlagEmoji: (code: string) => `Flag(${code})`,
}));

vi.mock('@thunder/logger/react', () => ({
  useLogger: () => ({error: vi.fn(), info: vi.fn(), warn: vi.fn(), debug: vi.fn()}),
}));

// Provide lightweight MUI mocks for DataGrid + ListingTable
vi.mock('@wso2/oxygen-ui', async () => {
  const actual = await vi.importActual<typeof import('@wso2/oxygen-ui')>('@wso2/oxygen-ui');
  return {
    ...actual,
    ListingTable: {
      Provider: ({children, loading}: {children: React.ReactNode; loading: boolean}) => (
        <div data-testid="data-grid" data-loading={String(loading)}>
          {children}
        </div>
      ),
      Container: ({children}: {children: React.ReactNode}) => children,
      DataGrid: ({
        rows,
        columns,
        onRowClick = undefined,
      }: {
        rows: {id: string; code: string}[];
        columns: {renderCell?: (params: {row: {id: string; code: string}}) => React.ReactNode}[];
        onRowClick?: (params: {row: {id: string; code: string}}) => void;
      }) => (
        <>
          {rows.map((row) => (
            <div
              key={row.id}
              data-testid={`row-${row.id}`}
              role="row"
              onClick={() => onRowClick?.({row})}
              onKeyDown={(e) => e.key === 'Enter' && onRowClick?.({row})}
              tabIndex={0}
            >
              {row.code}
              {columns.map((col, i) => (
                // eslint-disable-next-line react/no-array-index-key
                <span key={i}>{col.renderCell?.({row})}</span>
              ))}
            </div>
          ))}
        </>
      ),
      RowActions: ({children}: {children: React.ReactNode}) => children,
    },
  };
});

// Mock TranslationDeleteDialog to control it directly
const mockDeleteDialog = vi.fn();
vi.mock('../TranslationDeleteDialog', () => ({
  default: (props: {open: boolean; language: string | null; onClose: () => void}) => {
    mockDeleteDialog(props);
    return props.open ? (
      <div data-testid="delete-dialog">
        <span data-testid="delete-language">{props.language}</span>
        <button type="button" onClick={props.onClose}>
          close-dialog
        </button>
      </div>
    ) : null;
  },
}));

describe('TranslationsList', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders rows for each language', async () => {
    await renderWithProviders(<TranslationsList />);

    await expect.element(page.getByTestId('row-fr-FR')).toBeInTheDocument();
    await expect.element(page.getByTestId('row-de-DE')).toBeInTheDocument();
  });

  it('navigates to the edit page when a row is clicked', async () => {
    await renderWithProviders(<TranslationsList />);

    await userEvent.click(page.getByTestId('row-fr-FR'));

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith('/translations/fr-FR');
    });
  });

  it('navigates to the edit page when the edit button is clicked', async () => {
    await renderWithProviders(<TranslationsList />);

    const editButtons = page.getByRole('button', {name: 'Edit'}).all();
    await editButtons[0].click();

    await vi.waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith(expect.stringMatching(/\/translations\//));
    });
  });

  it('opens the delete dialog when the delete button is clicked', async () => {
    await renderWithProviders(<TranslationsList />);

    const deleteButtons = page.getByRole('button', {name: 'Delete'}).all();
    await deleteButtons[0].click();

    await expect.element(page.getByTestId('delete-dialog')).toBeInTheDocument();
    await expect.element(page.getByTestId('delete-language')).toHaveTextContent('fr-FR');
  });

  it('closes the delete dialog and clears the language when dialog onClose is called', async () => {
    await renderWithProviders(<TranslationsList />);

    // Open dialog
    const deleteButtons = page.getByRole('button', {name: 'Delete'}).all();
    await deleteButtons[0].click();

    await expect.element(page.getByTestId('delete-dialog')).toBeInTheDocument();

    // Close dialog
    await userEvent.click(page.getByText('close-dialog'));

    await expect.element(page.getByTestId('delete-dialog')).not.toBeInTheDocument();
  });
});
