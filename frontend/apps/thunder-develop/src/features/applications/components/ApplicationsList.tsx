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

import {useMemo, useCallback, useState, type JSX} from 'react';
import {useNavigate} from 'react-router';
import {useLogger} from '@thunder/logger/react';
import {
  Box,
  Avatar,
  Chip,
  IconButton,
  Typography,
  Menu,
  MenuItem,
  ListItemIcon,
  ListItemText,
  DataGrid,
  useTheme,
} from '@wso2/oxygen-ui';
import {AppWindow, EllipsisVertical, Eye, Trash2} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import useDataGridLocaleText from '../../../hooks/useDataGridLocaleText';
import useGetApplications from '../api/useGetApplications';
import type {BasicApplication} from '../models/application';
import ApplicationDeleteDialog from './ApplicationDeleteDialog';
import getTemplateMetadata from '../utils/getTemplateMetadata';

export default function ApplicationsList(): JSX.Element {
  const theme = useTheme();
  const navigate = useNavigate();
  const {t} = useTranslation();
  const logger = useLogger('ApplicationsList');
  const dataGridLocaleText = useDataGridLocaleText();
  const {data, isLoading, error} = useGetApplications();

  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [selectedAppId, setSelectedAppId] = useState<string | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState<boolean>(false);

  const handleMenuOpen = useCallback((event: React.MouseEvent<HTMLElement>, appId: string) => {
    event.stopPropagation();
    setAnchorEl(event.currentTarget);
    setSelectedAppId(appId);
  }, []);

  const handleMenuClose = (): void => {
    setAnchorEl(null);
  };

  const handleDeleteClick = (): void => {
    handleMenuClose();
    setDeleteDialogOpen(true);
  };

  const handleDeleteDialogClose = (): void => {
    setDeleteDialogOpen(false);
    setSelectedAppId(null);
  };

  const handleViewClick = (): void => {
    handleMenuClose();
    if (selectedAppId) {
      (async (): Promise<void> => {
        await navigate(`/applications/${selectedAppId}`);
      })().catch((_error: unknown) => {
        logger.error('Failed to navigate to application details', {error: _error, applicationId: selectedAppId});
      });
    }
  };

  const columns: DataGrid.GridColDef<BasicApplication>[] = useMemo(
    () => [
      {
        field: 'avatar',
        headerName: '',
        width: 70,
        sortable: false,
        filterable: false,
        renderCell: (params: DataGrid.GridRenderCellParams<BasicApplication>): JSX.Element => (
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              height: '100%',
            }}
          >
            <Avatar
              src={params.row.logo_url}
              slotProps={{
                img: {
                  onError: (e: React.SyntheticEvent<HTMLImageElement>) => {
                    e.currentTarget.src = '';
                  },
                },
              }}
              sx={{
                p: 0.5,
                backgroundColor: theme.vars?.palette.grey[500],
                width: 30,
                height: 30,
                fontSize: '0.875rem',
                ...theme.applyStyles('dark', {
                  backgroundColor: theme.vars?.palette.grey[900],
                }),
              }}
            >
              <AppWindow size={14} />
            </Avatar>
          </Box>
        ),
      },
      {
        field: 'name',
        headerName: t('applications:listing.columns.name'),
        flex: 1,
        minWidth: 200,
      },
      {
        field: 'description',
        headerName: t('common:form.description'),
        flex: 1.5,
        minWidth: 250,
        valueGetter: (_value, row): string => row.description ?? '-',
      },
      {
        field: 'template',
        headerName: t('applications:listing.columns.template'),
        flex: 0.8,
        minWidth: 120,
        renderCell: (params: DataGrid.GridRenderCellParams<BasicApplication>): JSX.Element => {
          const templateMetadata = getTemplateMetadata(params.row.template);
          return templateMetadata ? (
            <Chip
              icon={
                <Box sx={{display: 'flex', alignItems: 'center', '& > *': {width: 16, height: 16}}}>
                  {templateMetadata.icon}
                </Box>
              }
              label={templateMetadata.displayName}
              size="small"
              variant="outlined"
              sx={{
                fontSize: '0.75rem',
              }}
            />
          ) : (
            <>-</>
          );
        },
      },
      {
        field: 'client_id',
        headerName: t('applications:listing.columns.clientId'),
        flex: 1,
        minWidth: 200,
        renderCell: (params: DataGrid.GridRenderCellParams<BasicApplication>): JSX.Element =>
          params.row.client_id ? (
            <Chip
              label={params.row.client_id}
              size="small"
              variant="outlined"
              sx={{
                fontFamily: 'monospace',
                fontSize: '0.7rem',
                maxWidth: '100%',
                '& .MuiChip-label': {
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                },
              }}
            />
          ) : (
            <>-</>
          ),
      },
      {
        field: 'actions',
        headerName: t('applications:listing.columns.actions'),
        width: 80,
        sortable: false,
        filterable: false,
        hideable: false,
        renderCell: (params: DataGrid.GridRenderCellParams<BasicApplication>): JSX.Element => (
          <IconButton
            size="small"
            aria-label="Open actions menu"
            onClick={(e) => {
              handleMenuOpen(e, params.row.id);
            }}
          >
            <EllipsisVertical size={16} />
          </IconButton>
        ),
      },
    ],
    [handleMenuOpen, t, theme],
  );

  if (error) {
    return (
      <Box sx={{textAlign: 'center', py: 8}}>
        <Typography variant="h6" color="error" gutterBottom>
          Failed to load applications
        </Typography>
        <Typography variant="body2" color="text.secondary">
          {error.message ?? 'Unknown error'}
        </Typography>
      </Box>
    );
  }

  return (
    <>
      <Box sx={{height: 600, width: '100%'}}>
        <DataGrid.DataGrid
          rows={data?.applications ?? []}
          columns={columns}
          loading={isLoading}
          getRowId={(row): string => row.id}
          onRowClick={(params) => {
            const applicationId = (params.row as BasicApplication).id;
            (async (): Promise<void> => {
              await navigate(`/applications/${applicationId}`);
            })().catch((_error: unknown) => {
              logger.error('Failed to navigate to application', {error: _error, applicationId});
            });
          }}
          initialState={{
            pagination: {
              paginationModel: {pageSize: 10},
            },
          }}
          pageSizeOptions={[5, 10, 25, 50]}
          disableRowSelectionOnClick
          localeText={dataGridLocaleText}
          sx={{
            '& .MuiDataGrid-row': {
              cursor: 'pointer',
            },
          }}
        />
      </Box>

      {/* Actions Menu */}
      <Menu anchorEl={anchorEl} open={Boolean(anchorEl)} onClose={handleMenuClose}>
        <MenuItem onClick={handleViewClick}>
          <ListItemIcon>
            <Eye size={16} />
          </ListItemIcon>
          <ListItemText>{t('common:actions.view')}</ListItemText>
        </MenuItem>
        <MenuItem onClick={handleDeleteClick}>
          <ListItemIcon>
            <Trash2 size={16} color={theme.vars?.palette.error.main} />
          </ListItemIcon>
          <ListItemText sx={{color: 'error.main'}}>{t('common:actions.delete')}</ListItemText>
        </MenuItem>
      </Menu>

      <ApplicationDeleteDialog
        open={deleteDialogOpen}
        applicationId={selectedAppId}
        onClose={handleDeleteDialogClose}
      />
    </>
  );
}
