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

import {
  Box,
  Button,
  Chip,
  DataGrid,
  IconButton,
  ListItemIcon,
  ListItemText,
  Menu,
  MenuItem,
  PageContent,
  PageTitle,
  Typography,
} from '@wso2/oxygen-ui';
import {EllipsisVertical, Pencil, Plus} from '@wso2/oxygen-ui-icons-react';
import {useGetLanguages} from '@thunder/i18n';
import {useConfig} from '@thunder/shared-contexts';
import {useCallback, useMemo, useState, type JSX} from 'react';
import {useTranslation} from 'react-i18next';
import {useNavigate} from 'react-router';
import {useLogger} from '@thunder/logger/react';
import useDataGridLocaleText from '../../../hooks/useDataGridLocaleText';
import normalizeLanguageCode from '../utils/normalizeLanguageCode';
import getLanguageFlag from '../utils/getLanguageFlag';
import getLanguageDisplayName from '../utils/getLanguageDisplayName';

export default function TranslationsListPage(): JSX.Element {
  const {t} = useTranslation('translations');
  const {getServerUrl} = useConfig();
  const serverUrl = getServerUrl();
  const navigate = useNavigate();
  const logger = useLogger('TranslationsListPage');
  const dataGridLocaleText = useDataGridLocaleText();

  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [selectedLanguage, setSelectedLanguage] = useState<string | null>(null);

  const {data, isLoading} = useGetLanguages({serverUrl});

  const handleNavigate = useCallback(
    (language: string) => {
      (async (): Promise<void> => {
        await navigate(`/translations/${language}`);
      })().catch((_error: unknown) => {
        logger.error('Failed to navigate to translation editor', {error: _error, language});
      });
    },
    [navigate, logger],
  );

  const handleAddLanguage = useCallback(() => {
    (async (): Promise<void> => {
      await navigate('/translations/create');
    })().catch((_error: unknown) => {
      logger.error('Failed to navigate to translation create page', {error: _error});
    });
  }, [navigate, logger]);

  const handleMenuOpen = useCallback((event: React.MouseEvent<HTMLElement>, language: string) => {
    event.stopPropagation();
    setAnchorEl(event.currentTarget);
    setSelectedLanguage(language);
  }, []);

  const handleMenuClose = (): void => {
    setAnchorEl(null);
  };

  const handleViewClick = (): void => {
    handleMenuClose();
    if (selectedLanguage) {
      handleNavigate(selectedLanguage);
    }
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
              {getLanguageFlag(params.row.code)}
            </Typography>
            <Box sx={{display: 'flex', flexDirection: 'column', gap: 0.5, alignItems: 'flex-start'}}>
              <Typography variant="body2">{getLanguageDisplayName(params.row.code)}</Typography>
              <Chip
                label={normalizeLanguageCode(params.row.code)}
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
        width: 80,
        sortable: false,
        filterable: false,
        hideable: false,
        renderCell: (params: DataGrid.GridRenderCellParams<{id: string; code: string}>): JSX.Element => (
          <IconButton
            size="small"
            aria-label={t('common:actions.openActionsMenu', {ns: 'common'})}
            onClick={(e) => {
              handleMenuOpen(e, params.row.code);
            }}
          >
            <EllipsisVertical size={16} />
          </IconButton>
        ),
      },
    ],
    [handleMenuOpen, t],
  );

  return (
    <PageContent>
      <PageTitle>
        <PageTitle.Header>{t('page.title')}</PageTitle.Header>
        <PageTitle.SubHeader>{t('page.subtitle')}</PageTitle.SubHeader>
        <PageTitle.Actions>
          <Button variant="contained" startIcon={<Plus size={18} />} onClick={handleAddLanguage}>
            {t('listing.addLanguage')}
          </Button>
        </PageTitle.Actions>
      </PageTitle>

      <Box sx={{height: 600, width: '100%'}}>
        <DataGrid.DataGrid
          rows={rows}
          columns={columns}
          loading={isLoading}
          getRowId={(row): string => row.id}
          onRowClick={(params) => {
            handleNavigate((params.row as {id: string; code: string}).code);
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
            '& .MuiDataGrid-row': {
              cursor: 'pointer',
            },
          }}
        />
      </Box>

      <Menu anchorEl={anchorEl} open={Boolean(anchorEl)} onClose={handleMenuClose}>
        <MenuItem onClick={handleViewClick}>
          <ListItemIcon>
            <Pencil size={16} />
          </ListItemIcon>
          <ListItemText>{t('common:actions.edit', {ns: 'common'})}</ListItemText>
        </MenuItem>
      </Menu>
    </PageContent>
  );
}
