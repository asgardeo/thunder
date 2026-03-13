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

import {useCallback, useMemo, useState, type JSX} from 'react';
import {useNavigate} from 'react-router';
import {useLogger} from '@thunder/logger/react';
import {Box, Chip, DataGrid, IconButton, ListingTable, Tooltip, Typography} from '@wso2/oxygen-ui';
import {Pencil, Trash2} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import {getDisplayNameForCode, toFlagEmoji, useGetLanguages} from '@thunder/i18n';
import useDataGridLocaleText from '../../../hooks/useDataGridLocaleText';
import TranslationDeleteDialog from './TranslationDeleteDialog';

export default function TranslationsList(): JSX.Element {
  const {t} = useTranslation('translations');
  const navigate = useNavigate();
  const logger = useLogger('TranslationsList');
  const dataGridLocaleText = useDataGridLocaleText();

  const [selectedLanguage, setSelectedLanguage] = useState<string | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState<boolean>(false);

  const {data, isLoading} = useGetLanguages();

  const handleEditClick = useCallback(
    (language: string): void => {
      (async (): Promise<void> => {
        await navigate(`/translations/${language}`);
      })().catch((_error: unknown) => {
        logger.error('Failed to navigate to translation editor', {error: _error, language});
      });
    },
    [logger, navigate],
  );

  const handleDeleteClick = useCallback((language: string): void => {
    setSelectedLanguage(language);
    setDeleteDialogOpen(true);
  }, []);

  const handleDeleteDialogClose = (): void => {
    setDeleteDialogOpen(false);
    setSelectedLanguage(null);
  };

  const rows = useMemo(() => (data?.languages ?? []).map((code) => ({id: code, code})), [data?.languages]);

  const columns: DataGrid.GridColDef<{id: string; code: string}>[] = useMemo(
    () => [
      {
        field: 'code',
        headerName: t('listing.columns.language'),
        flex: 1,
        minWidth: 240,
        renderCell: (params: DataGrid.GridRenderCellParams<{id: string; code: string}>): JSX.Element => (
          <Box sx={{display: 'flex', alignItems: 'center', gap: 1.5, height: '100%'}}>
            <Typography sx={{fontSize: '1.4rem', lineHeight: 1, userSelect: 'none'}}>
              {toFlagEmoji(params.row.code)}
            </Typography>
            <Box sx={{display: 'flex', flexDirection: 'column', gap: 0.5, alignItems: 'flex-start'}}>
              <Typography variant="body2">{getDisplayNameForCode(params.row.code)}</Typography>
              <Chip
                label={params.row.code}
                size="small"
                variant="outlined"
                sx={{fontSize: '0.7rem', fontFamily: 'monospace', height: 18}}
              />
            </Box>
          </Box>
        ),
      },
      {
        field: 'actions',
        headerName: t('listing.columns.actions'),
        width: 150,
        align: 'center',
        headerAlign: 'center',
        sortable: false,
        filterable: false,
        hideable: false,
        renderCell: (params: DataGrid.GridRenderCellParams<{id: string; code: string}>): JSX.Element => (
          <ListingTable.RowActions visibility="hover">
            <Tooltip title={t('common:actions.edit')}>
              <IconButton
                size="small"
                onClick={(e) => {
                  e.stopPropagation();
                  handleEditClick(params.row.code);
                }}
              >
                <Pencil size={16} />
              </IconButton>
            </Tooltip>
            <Tooltip title={t('common:actions.delete')}>
              <IconButton
                size="small"
                color="error"
                onClick={(e) => {
                  e.stopPropagation();
                  handleDeleteClick(params.row.code);
                }}
              >
                <Trash2 size={16} />
              </IconButton>
            </Tooltip>
          </ListingTable.RowActions>
        ),
      },
    ],
    [handleDeleteClick, handleEditClick, t],
  );

  return (
    <>
      <ListingTable.Provider variant="data-grid-card" loading={isLoading}>
        <ListingTable.Container disablePaper>
          <ListingTable.DataGrid
            rows={rows}
            columns={columns}
            getRowId={(row): string => (row as {id: string; code: string}).id}
            onRowClick={(params) => {
              handleEditClick((params.row as {id: string; code: string}).code);
            }}
            initialState={{
              pagination: {
                paginationModel: {pageSize: 10},
              },
            }}
            pageSizeOptions={[5, 10, 25]}
            rowHeight={56}
            disableRowSelectionOnClick
            localeText={dataGridLocaleText}
            sx={{
              height: 'auto',
              '& .MuiDataGrid-row': {
                cursor: 'pointer',
              },
            }}
          />
        </ListingTable.Container>
      </ListingTable.Provider>

      <TranslationDeleteDialog
        open={deleteDialogOpen}
        language={selectedLanguage}
        onClose={handleDeleteDialogClose}
      />
    </>
  );
}
