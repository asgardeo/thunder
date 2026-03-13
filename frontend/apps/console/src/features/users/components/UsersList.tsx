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

import {useEffect, useMemo, useState, useCallback} from 'react';
import {useNavigate} from 'react-router';
import {
  Avatar,
  IconButton,
  Tooltip,
  Typography,
  Snackbar,
  Alert,
  ListingTable,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  DialogActions,
  Button,
  DataGrid,
} from '@wso2/oxygen-ui';
import {Eye, Trash2} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import {useLogger} from '@thunder/logger/react';
import useDataGridLocaleText from '../../../hooks/useDataGridLocaleText';
import useGetUsers from '../api/useGetUsers';
import useDeleteUser from '../api/useDeleteUser';
import type {UserWithDetails} from '../types/users';

export default function UsersList() {
  const navigate = useNavigate();
  const {t} = useTranslation();
  const logger = useLogger('UsersList');
  const dataGridLocaleText = useDataGridLocaleText();

  const {data: userData, isLoading, error: usersRequestError} = useGetUsers();
  const deleteUserMutation = useDeleteUser();

  const error = usersRequestError;

  const [snackbarOpen, setSnackbarOpen] = useState(false);
  const [selectedUserId, setSelectedUserId] = useState<string | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

  // Show snackbar when error occurs
  useEffect(() => {
    if (error) {
      setSnackbarOpen(true);
    }
  }, [error]);

  const handleCloseSnackbar = () => {
    setSnackbarOpen(false);
  };

  const handleDeleteClick = useCallback((userId: string): void => {
    setSelectedUserId(userId);
    setDeleteDialogOpen(true);
  }, []);

  const handleViewClick = useCallback(
    (userId: string): void => {
      (async (): Promise<void> => {
        await navigate(`/users/${userId}`);
      })().catch((_error: unknown) => {
        logger.error('Failed to navigate to user details', {error: _error, userId});
      });
    },
    [logger, navigate],
  );

  const handleDeleteCancel = () => {
    setDeleteDialogOpen(false);
    setSelectedUserId(null);
  };

  const handleDeleteConfirm = async () => {
    if (!selectedUserId) return;

    try {
      await deleteUserMutation.mutateAsync(selectedUserId);
      setDeleteDialogOpen(false);
      setSelectedUserId(null);
    } catch (err) {
      // Error is already handled in the hook
      setDeleteDialogOpen(false);
      logger.error('Failed to delete user', {error: err, userId: selectedUserId});
    }
  };

  const getInitials = (name?: string) => {
    if (!name) return '?';
    return name
      .split(' ')
      .map((n) => n[0])
      .join('')
      .toUpperCase()
      .slice(0, 2);
  };

  const columns: DataGrid.GridColDef<UserWithDetails>[] = useMemo(
    () => [
      {
        field: 'display',
        headerName: t('users:displayName', 'Display Name'),
        flex: 1,
        minWidth: 200,
        renderCell: (params: DataGrid.GridRenderCellParams<UserWithDetails>) => {
          const displayVal = params.row.display ?? params.row.id;

          return (
            <ListingTable.CellIcon
              sx={{width: '100%'}}
              icon={
                <Avatar
                  sx={{
                    width: 30,
                    height: 30,
                    bgcolor: 'primary.main',
                    fontSize: '0.875rem',
                  }}
                >
                  {getInitials(displayVal)}
                </Avatar>
              }
              primary={displayVal}
            />
          );
        },
      },
      {
        field: 'id',
        headerName: t('users:userId', 'User ID'),
        flex: 1,
        minWidth: 200,
        renderCell: (params: DataGrid.GridRenderCellParams<UserWithDetails>) => (
          <Typography
            variant="body2"
            sx={{fontFamily: 'monospace', fontSize: '0.875rem'}}
          >
            {params.row.id}
          </Typography>
        ),
      },
      {
        field: 'actions',
        headerName: t('users:actions'),
        width: 150,
        align: 'center',
        headerAlign: 'center',
        sortable: false,
        filterable: false,
        hideable: false,
        renderCell: (params: DataGrid.GridRenderCellParams<UserWithDetails>) => (
          <ListingTable.RowActions visibility="hover">
            <Tooltip title={t('common:actions.view')}>
              <IconButton
                size="small"
                onClick={(e) => {
                  e.stopPropagation();
                  handleViewClick(params.row.id);
                }}
              >
                <Eye size={16} />
              </IconButton>
            </Tooltip>
            <Tooltip title={t('common:actions.delete')}>
              <IconButton
                size="small"
                color="error"
                onClick={(e) => {
                  e.stopPropagation();
                  handleDeleteClick(params.row.id);
                }}
              >
                <Trash2 size={16} />
              </IconButton>
            </Tooltip>
          </ListingTable.RowActions>
        ),
      },
    ],
    [handleDeleteClick, handleViewClick, t],
  );

  return (
    <>
      <ListingTable.Provider variant="data-grid-card" loading={isLoading}>
        <ListingTable.Container disablePaper>
          <ListingTable.DataGrid
            rows={userData?.users ?? []}
            columns={columns}
            getRowId={(row) => (row as UserWithDetails).id}
            onRowClick={(params) => {
              handleViewClick((params.row as UserWithDetails).id);
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
              height: 'auto',
              '& .MuiDataGrid-row': {
                cursor: 'pointer',
              },
            }}
          />
        </ListingTable.Container>
      </ListingTable.Provider>

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onClose={handleDeleteCancel}>
        <DialogTitle>{t('users:deleteUser')}</DialogTitle>
        <DialogContent>
          <DialogContentText>{t('users:confirmDeleteUser')}</DialogContentText>
          {deleteUserMutation.error && (
            <Alert severity="error" sx={{mt: 2}}>
              <Typography variant="body2" sx={{fontWeight: 'bold'}}>
                {deleteUserMutation.error.message}
              </Typography>
            </Alert>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={handleDeleteCancel} disabled={deleteUserMutation.isPending}>
            {t('common:actions.cancel')}
          </Button>
          <Button
            onClick={() => {
              handleDeleteConfirm().catch(() => {
                // Handle error
              });
            }}
            color="error"
            variant="contained"
            disabled={deleteUserMutation.isPending}
          >
            {deleteUserMutation.isPending ? t('common:status.loading') : t('common:actions.delete')}
          </Button>
        </DialogActions>
      </Dialog>

      <Snackbar
        open={snackbarOpen}
        autoHideDuration={6000}
        onClose={handleCloseSnackbar}
        anchorOrigin={{vertical: 'top', horizontal: 'right'}}
      >
        <Alert onClose={handleCloseSnackbar} severity="error" sx={{width: '100%'}}>
          {error?.message ?? t('common:messages.saveError')}
        </Alert>
      </Snackbar>
    </>
  );
}
