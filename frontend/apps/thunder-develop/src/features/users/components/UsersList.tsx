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
  Box,
  Avatar,
  Chip,
  IconButton,
  Typography,
  Snackbar,
  Alert,
  Menu,
  MenuItem,
  ListItemIcon,
  ListItemText,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  DialogActions,
  Button,
  DataGrid,
} from '@wso2/oxygen-ui';
import {EllipsisVertical, Trash2, Eye} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import {useLogger} from '@thunder/logger/react';
import useDataGridLocaleText from '../../../hooks/useDataGridLocaleText';
import useGetUsers from '../api/useGetUsers';
import useGetUserType from '../api/useGetUserType';
import useDeleteUser from '../api/useDeleteUser';
import type {UserWithDetails} from '../types/users';

interface UsersListProps {
  selectedUserTypeId?: string;
  selectedSchema?: string;
}

export default function UsersList(props: UsersListProps) {
  const selectedUserTypeId = props.selectedUserTypeId ?? props.selectedSchema ?? '';
  const navigate = useNavigate();
  const {t} = useTranslation();
  const logger = useLogger('UsersList');
  const dataGridLocaleText = useDataGridLocaleText();

  const {data: userData, loading: isUsersRequestLoading, error: usersRequestError, refetch} = useGetUsers();
  const {deleteUser, loading: isDeleting, error: deleteUserError} = useDeleteUser();

  const {
    data: userTypeDefinition,
    loading: isUserTypeRequestLoading,
    error: userTypeRequestError,
  } = useGetUserType(selectedUserTypeId);

  const error = usersRequestError ?? userTypeRequestError;
  const isLoading = isUsersRequestLoading || isUserTypeRequestLoading;

  const [snackbarOpen, setSnackbarOpen] = useState(false);
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
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

  const handleMenuOpen = useCallback((event: React.MouseEvent<HTMLElement>, userId: string) => {
    event.stopPropagation();
    setAnchorEl(event.currentTarget);
    setSelectedUserId(userId);
  }, []);

  const handleMenuClose = () => {
    setAnchorEl(null);
  };

  const handleViewUser = async () => {
    if (selectedUserId) {
      await navigate(`/users/${selectedUserId}`);
    }
    handleMenuClose();
  };

  const handleDeleteClick = () => {
    setDeleteDialogOpen(true);
    handleMenuClose();
  };

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
    if (!userTypeDefinition) {
      // Return basic columns if schema is not loaded yet
      return [];
    }

    const schemaColumns: DataGrid.GridColDef<UserWithDetails>[] = [];
    const schemaEntries = Object.entries(userTypeDefinition.schema);

    // Helper function to format field names
    const formatHeaderName = (fieldName: string): string =>
      fieldName
        .replace(/([A-Z])/g, ' $1')
        .replace(/^./, (str) => str.toUpperCase())
        .trim();

    // Add avatar column if firstname/lastname exist in schema
    const hasNameFields = schemaEntries.some(([key]) => key === 'firstname' || key === 'lastname');

    if (hasNameFields) {
      schemaColumns.push({
        field: 'avatar',
        headerName: '',
        width: 70,
        sortable: false,
        filterable: false,
        renderCell: (params: DataGrid.GridRenderCellParams<UserWithDetails>) => {
          const firstname = params.row.attributes?.firstname as string | undefined;
          const lastname = params.row.attributes?.lastname as string | undefined;
          const username = params.row.attributes?.username as string | undefined;
          const displayName = [firstname, lastname, username].filter(Boolean).join(' ');
          return (
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
                  width: 30,
                  height: 30,
                  bgcolor: 'primary.main',
                  fontSize: '0.875rem',
                }}
              >
                {getInitials(displayName)}
              </Avatar>
            </Box>
          );
        },
      });
    }

    // Dynamically generate columns from schema
    schemaEntries.forEach(([fieldName, fieldDef]) => {
      // Special handling for username to show with full name
      if (fieldName === 'username') {
        schemaColumns.push({
          field: fieldName,
          headerName: formatHeaderName(fieldName),
          flex: 1,
          minWidth: 150,
          renderCell: (params: DataGrid.GridRenderCellParams<UserWithDetails>) => {
            const username = (params.row.attributes?.username as string | undefined) ?? '-';
            const firstname = params.row.attributes?.firstname as string | undefined;
            const lastname = params.row.attributes?.lastname as string | undefined;
            const fullName = [firstname, lastname].filter(Boolean).join(' ');

            return (
              <Box
                sx={{
                  lineHeight: 1.2,
                  display: 'flex',
                  flexDirection: 'column',
                  justifyContent: 'center',
                  height: '100%',
                }}
              >
                <Typography variant="body2" sx={{lineHeight: 1.3}}>
                  {username}
                </Typography>
                {fullName && (
                  <Typography variant="caption" color="text.secondary" sx={{lineHeight: 1.2}}>
                    {fullName}
                  </Typography>
                )}
              </Box>
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
      width: 80,
      sortable: false,
      filterable: false,
      hideable: false,
      renderCell: (params: DataGrid.GridRenderCellParams<UserWithDetails>) => (
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
    });

    return schemaColumns;
  }, [handleMenuOpen, t, userTypeDefinition]);

  // Calculate initial column visibility: show first 4 columns, hide the rest
  const initialColumnVisibility = useMemo(() => {
    if (!columns || columns.length === 0) return {};

    const visibility: Record<string, boolean> = {};
    const VISIBLE_COLUMN_COUNT = 4;

    columns.forEach((column, index) => {
      // Always show avatar and actions columns
      if (column.field === 'avatar' || column.field === 'actions') {
        visibility[column.field] = true;
      } else {
        // Show first VISIBLE_COLUMN_COUNT data columns, hide the rest
        const dataColumnIndex = columns
          .slice(0, index)
          .filter((col) => col.field !== 'avatar' && col.field !== 'actions').length;

        visibility[column.field] = dataColumnIndex < VISIBLE_COLUMN_COUNT;
      }
    });

    return visibility;
  }, [columns]);

  return (
    <>
      <Box sx={{height: 600, width: '100%'}}>
        <DataGrid.DataGrid
          rows={userData?.users}
          columns={columns}
          loading={isLoading}
          getRowId={(row) => row.id}
          onRowClick={(params) => {
            const userId = (params.row as UserWithDetails).id;

            (async () => {
              await navigate(`/users/${userId}`);
            })().catch((_error: unknown) => {
              logger.error('Failed to navigate to user details', {error: _error, userId});
            });
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
            '& .MuiDataGrid-row': {
              cursor: 'pointer',
            },
          }}
        />
      </Box>

      {/* Actions Menu */}
      <Menu anchorEl={anchorEl} open={Boolean(anchorEl)} onClose={handleMenuClose}>
        <MenuItem
          onClick={() => {
            handleViewUser().catch(() => {
              // Handle error
            });
          }}
        >
          <ListItemIcon>
            <Eye size={16} />
          </ListItemIcon>
          <ListItemText>{t('common:actions.view')}</ListItemText>
        </MenuItem>
        <MenuItem onClick={handleDeleteClick}>
          <ListItemIcon sx={{color: 'error.main'}}>
            <Trash2 size={16} />
          </ListItemIcon>
          <ListItemText sx={{color: 'error.main'}}>{t('common:actions.delete')}</ListItemText>
        </MenuItem>
      </Menu>

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
