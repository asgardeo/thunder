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
  Chip,
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
import useGetUserSchema from '../api/useGetUserSchema';
import useDeleteUser from '../api/useDeleteUser';
import type {UserWithDetails} from '../types/users';

interface UsersListProps {
  selectedSchema: string;
}

export default function UsersList(props: UsersListProps) {
  const {selectedSchema} = props;
  const navigate = useNavigate();
  const {t} = useTranslation();
  const logger = useLogger('UsersList');
  const dataGridLocaleText = useDataGridLocaleText();

  const {data: userData, loading: isUsersRequestLoading, error: usersRequestError, refetch} = useGetUsers();
  const {deleteUser, loading: isDeleting, error: deleteUserError} = useDeleteUser();

  const {
    data: defaultUserSchema,
    loading: isDefaultUserSchemaRequestLoading,
    error: defaultUserSchemaRequestError,
  } = useGetUserSchema(selectedSchema);

  const error = usersRequestError ?? defaultUserSchemaRequestError;
  const isLoading = isUsersRequestLoading || isDefaultUserSchemaRequestLoading;

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
      await deleteUser(selectedUserId);
      setDeleteDialogOpen(false);
      setSelectedUserId(null);
      // Refetch users list after successful deletion
      await refetch();
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

  const columns: DataGrid.GridColDef<UserWithDetails>[] = useMemo(() => {
    if (!defaultUserSchema) {
      // Return basic columns if schema is not loaded yet
      return [];
    }

    const schemaColumns: DataGrid.GridColDef<UserWithDetails>[] = [];
    const schemaEntries = Object.entries(defaultUserSchema.schema);

    // Helper function to format field names
    const formatHeaderName = (fieldName: string): string =>
      fieldName
        .replace(/([A-Z])/g, ' $1')
        .replace(/^./, (str) => str.toUpperCase())
        .trim();

    // Dynamically generate columns from schema
    schemaEntries.forEach(([fieldName, fieldDef]) => {
      // Special handling for username to show with avatar
      if (fieldName === 'username') {
        schemaColumns.push({
          field: fieldName,
          headerName: formatHeaderName(fieldName),
          flex: 1,
          minWidth: 200,
          renderCell: (params: DataGrid.GridRenderCellParams<UserWithDetails>) => {
            const usernameVal = (params.row.attributes?.username as string | undefined) ?? '-';
            const firstnameVal = params.row.attributes?.firstname as string | undefined;
            const lastnameVal = params.row.attributes?.lastname as string | undefined;
            const displayName = [firstnameVal, lastnameVal, usernameVal].filter(Boolean).join(' ');

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
                    {getInitials(displayName)}
                  </Avatar>
                }
                primary={usernameVal}
              />
            );
          },
        });
        return;
      }

      // Skip firstname/lastname as they're shown with username
      if (fieldName === 'firstname' || fieldName === 'lastname') {
        return;
      }

      // Special handling for isActive/status fields with Chip
      if (fieldName === 'isActive' || fieldName === 'active' || fieldName === 'status') {
        schemaColumns.push({
          field: fieldName,
          headerName: formatHeaderName(fieldName),
          width: 120,
          renderCell: (params: DataGrid.GridRenderCellParams<UserWithDetails>) => {
            const value = params.row.attributes?.[fieldName] as boolean | string | undefined;
            if (value === undefined || value === null) return null;

            const isActive = typeof value === 'boolean' ? value : value === 'active';
            return (
              <Chip
                label={isActive ? t('common:status.active') : t('common:status.inactive')}
                size="small"
                color={isActive ? 'success' : 'default'}
              />
            );
          },
        });
        return;
      }

      // Handle different field types
      const columnDef: DataGrid.GridColDef<UserWithDetails> = {
        field: fieldName,
        headerName: formatHeaderName(fieldName),
        flex: 1,
        minWidth: 150,
      };

      // Type-specific configuration
      switch (fieldDef.type) {
        case 'boolean':
          columnDef.renderCell = (params: DataGrid.GridRenderCellParams<UserWithDetails>) => {
            const value = params.row.attributes?.[fieldName] as boolean | undefined;
            if (value === undefined || value === null) return '-';
            return value ? t('common:actions.yes') : t('common:actions.no');
          };
          break;

        case 'number':
          columnDef.type = 'number';
          columnDef.valueGetter = (_value: unknown, row: UserWithDetails) => {
            const value = row.attributes?.[fieldName] as number | undefined;
            return value ?? null;
          };
          break;

        case 'array':
          columnDef.sortable = false;
          columnDef.renderCell = (params: DataGrid.GridRenderCellParams<UserWithDetails>) => {
            const value = params.row.attributes?.[fieldName] as unknown[] | undefined;
            if (!value || !Array.isArray(value) || value.length === 0) return '-';
            return value.join(', ');
          };
          break;

        case 'object':
          columnDef.sortable = false;
          columnDef.renderCell = (params: DataGrid.GridRenderCellParams<UserWithDetails>) => {
            const value = params.row.attributes?.[fieldName] as Record<string, unknown> | undefined;
            if (!value || typeof value !== 'object') return '-';
            return JSON.stringify(value);
          };
          break;

        default:
          // String and other types
          columnDef.valueGetter = (_value, row) => {
            const value = row.attributes?.[fieldName] as string | number | undefined;
            return value ?? null;
          };
      }

      schemaColumns.push(columnDef);
    });

    // Add actions column at the end (pinned to the right)
    schemaColumns.push({
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
    });

    return schemaColumns;
  }, [defaultUserSchema, handleDeleteClick, handleViewClick, t]);

  // Calculate initial column visibility: show first 4 columns, hide the rest
  const initialColumnVisibility = useMemo(() => {
    if (!columns || columns.length === 0) return {};

    const visibility: Record<string, boolean> = {};
    const VISIBLE_COLUMN_COUNT = 4;

    columns.forEach((column, index) => {
      // Always show actions column
      if (column.field === 'actions') {
        visibility[column.field] = true;
      } else {
        // Show first VISIBLE_COLUMN_COUNT data columns, hide the rest
        const dataColumnIndex = columns.slice(0, index).filter((col) => col.field !== 'actions').length;

        visibility[column.field] = dataColumnIndex < VISIBLE_COLUMN_COUNT;
      }
    });

    return visibility;
  }, [columns]);

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
              columns: {
                columnVisibilityModel: initialColumnVisibility,
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
          {deleteUserError && (
            <Alert severity="error" sx={{mt: 2}}>
              <Typography variant="body2" sx={{fontWeight: 'bold'}}>
                {deleteUserError.message}
              </Typography>
              {deleteUserError.description && <Typography variant="caption">{deleteUserError.description}</Typography>}
            </Alert>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={handleDeleteCancel} disabled={isDeleting}>
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
            disabled={isDeleting}
          >
            {isDeleting ? t('common:status.loading') : t('common:actions.delete')}
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
