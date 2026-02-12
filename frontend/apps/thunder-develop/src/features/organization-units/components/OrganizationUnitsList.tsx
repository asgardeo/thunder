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
  IconButton,
  Typography,
  Menu,
  MenuItem,
  ListItemIcon,
  ListItemText,
  DataGrid,
  useTheme,
} from '@wso2/oxygen-ui';
import {Building, EllipsisVertical, Eye, Trash2} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import useDataGridLocaleText from '../../../hooks/useDataGridLocaleText';
import useGetOrganizationUnits from '../api/useGetOrganizationUnits';
import type {OrganizationUnit} from '../types/organization-units';
import OrganizationUnitDeleteDialog from './OrganizationUnitDeleteDialog';

export default function OrganizationUnitsList(): JSX.Element {
  const theme = useTheme();
  const navigate = useNavigate();
  const {t} = useTranslation();
  const logger = useLogger('OrganizationUnitsList');
  const dataGridLocaleText = useDataGridLocaleText();
  const {data, isLoading, error} = useGetOrganizationUnits();

  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [selectedOU, setSelectedOU] = useState<OrganizationUnit | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState<boolean>(false);

  const handleMenuOpen = useCallback((event: React.MouseEvent<HTMLElement>, ou: OrganizationUnit) => {
    event.stopPropagation();
    setAnchorEl(event.currentTarget);
    setSelectedOU(ou);
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
    setSelectedOU(null);
  };

  const handleViewClick = (): void => {
    handleMenuClose();
    if (selectedOU) {
      (async (): Promise<void> => {
        await navigate(`/organization-units/${selectedOU.id}`);
      })().catch((_error: unknown) => {
        logger.error('Failed to navigate to organization unit', {error: _error, ouId: selectedOU.id});
      });
    }
  };

  const columns: DataGrid.GridColDef<OrganizationUnit>[] = useMemo(
    () => [
      {
        field: 'avatar',
        headerName: '',
        width: 70,
        sortable: false,
        filterable: false,
        renderCell: (): JSX.Element => (
          <Box
            sx={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              height: '100%',
            }}
          >
            <Avatar
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
              <Building size={14} />
            </Avatar>
          </Box>
        ),
      },
      {
        field: 'name',
        headerName: t('organizationUnits:listing.columns.name'),
        flex: 1,
        minWidth: 200,
      },
      {
        field: 'handle',
        headerName: t('organizationUnits:listing.columns.handle'),
        flex: 1,
        minWidth: 150,
      },
      {
        field: 'description',
        headerName: t('organizationUnits:listing.columns.description'),
        flex: 2,
        minWidth: 250,
        valueGetter: (_value, row): string => row.description ?? '-',
      },
      {
        field: 'actions',
        headerName: t('organizationUnits:listing.columns.actions'),
        width: 80,
        sortable: false,
        filterable: false,
        hideable: false,
        renderCell: (params: DataGrid.GridRenderCellParams<OrganizationUnit>): JSX.Element => (
          <IconButton
            size="small"
            aria-label="Open actions menu"
            onClick={(e) => {
              handleMenuOpen(e, params.row);
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
          {t('organizationUnits:listing.error.title')}
        </Typography>
        <Typography variant="body2" color="text.secondary">
          {error.message ?? t('organizationUnits:listing.error.unknown')}
        </Typography>
      </Box>
    );
  }

  return (
    <>
      <Box sx={{height: 600, width: '100%'}}>
        <DataGrid.DataGrid
          rows={data?.organizationUnits ?? []}
          columns={columns}
          loading={isLoading}
          getRowId={(row): string => row.id}
          onRowClick={(params) => {
            const ou = params.row as OrganizationUnit;
            (async (): Promise<void> => {
              await navigate(`/organization-units/${ou.id}`);
            })().catch((_error: unknown) => {
              logger.error('Failed to navigate to organization unit', {error: _error, ouId: ou.id});
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

      <OrganizationUnitDeleteDialog
        open={deleteDialogOpen}
        organizationUnitId={selectedOU?.id ?? null}
        onClose={handleDeleteDialogClose}
      />
    </>
  );
}
