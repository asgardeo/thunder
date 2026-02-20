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
import {Users, EllipsisVertical, Eye, Trash2} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import useDataGridLocaleText from '../../../hooks/useDataGridLocaleText';
import useGetGroups from '../api/useGetGroups';
import type {GroupBasic} from '../models/group';
import GroupDeleteDialog from './GroupDeleteDialog';

/**
 * DataGrid component for displaying the list of groups.
 */
export default function GroupsList(): JSX.Element {
  const theme = useTheme();
  const navigate = useNavigate();
  const {t} = useTranslation();
  const logger = useLogger('GroupsList');
  const dataGridLocaleText = useDataGridLocaleText();
  const [paginationModel, setPaginationModel] = useState<DataGrid.GridPaginationModel>({pageSize: 10, page: 0});

  const groupsParams = useMemo(
    () => ({
      limit: paginationModel.pageSize,
      offset: paginationModel.page * paginationModel.pageSize,
    }),
    [paginationModel],
  );
  const {data, isLoading, error} = useGetGroups(groupsParams);

  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [selectedGroupId, setSelectedGroupId] = useState<string | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState<boolean>(false);

  const handleMenuOpen = useCallback((event: React.MouseEvent<HTMLElement>, groupId: string) => {
    event.stopPropagation();
    setAnchorEl(event.currentTarget);
    setSelectedGroupId(groupId);
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
    setSelectedGroupId(null);
  };

  const handleViewClick = (): void => {
    handleMenuClose();
    if (selectedGroupId) {
      (async (): Promise<void> => {
        await navigate(`/groups/${selectedGroupId}`);
      })().catch((_error: unknown) => {
        logger.error('Failed to navigate to group details', {error: _error, groupId: selectedGroupId});
      });
    }
  };

  const columns: DataGrid.GridColDef<GroupBasic>[] = useMemo(
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
              <Users size={14} />
            </Avatar>
          </Box>
        ),
      },
      {
        field: 'name',
        headerName: t('groups:listing.columns.name'),
        flex: 1,
        minWidth: 200,
      },
      {
        field: 'description',
        headerName: t('groups:listing.columns.description'),
        flex: 1.5,
        minWidth: 250,
        valueGetter: (_value, row): string => row.description ?? '-',
      },
      {
        field: 'organizationUnitId',
        headerName: t('groups:listing.columns.organizationUnit'),
        flex: 1,
        minWidth: 200,
      },
      {
        field: 'actions',
        headerName: t('groups:listing.columns.actions'),
        width: 80,
        sortable: false,
        filterable: false,
        hideable: false,
        renderCell: (params: DataGrid.GridRenderCellParams<GroupBasic>): JSX.Element => (
          <IconButton
            size="small"
            aria-label={t('common:actions.openActionsMenu')}
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
          {t('groups:listing.error')}
        </Typography>
        <Typography variant="body2" color="text.secondary">
          {error.message ?? t('common:messages.somethingWentWrong')}
        </Typography>
      </Box>
    );
  }

  return (
    <>
      <Box sx={{height: 600, width: '100%'}}>
        <DataGrid.DataGrid
          rows={data?.groups ?? []}
          columns={columns}
          loading={isLoading}
          getRowId={(row): string => row.id}
          onRowClick={(params) => {
            const groupId = (params.row as GroupBasic).id;
            (async (): Promise<void> => {
              await navigate(`/groups/${groupId}`);
            })().catch((_error: unknown) => {
              logger.error('Failed to navigate to group', {error: _error, groupId});
            });
          }}
          paginationMode="server"
          rowCount={data?.totalResults ?? 0}
          paginationModel={paginationModel}
          onPaginationModelChange={setPaginationModel}
          pageSizeOptions={[5, 10, 25]}
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

      <GroupDeleteDialog open={deleteDialogOpen} groupId={selectedGroupId} onClose={handleDeleteDialogClose} />
    </>
  );
}
