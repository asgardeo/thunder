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

import {useCallback, useEffect, useMemo, useState} from 'react';
import {useNavigate} from 'react-router';
import {
  Box,
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
  Chip,
} from '@wso2/oxygen-ui';
import {EllipsisVertical, Trash2, Eye} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import useDataGridLocaleText from '../../../hooks/useDataGridLocaleText';
import useGetUserTypes from '../api/useGetUserTypes';
import useDeleteUserType from '../api/useDeleteUserType';
import useGetOrganizationUnits from '../../organization-units/api/useGetOrganizationUnits';
import type {UserSchemaListItem} from '../types/user-types';

type GridColDef<R extends DataGrid.GridValidRowModel = DataGrid.GridValidRowModel> = DataGrid.GridColDef<R>;
type GridRenderCellParams<R extends DataGrid.GridValidRowModel = DataGrid.GridValidRowModel> =
  DataGrid.GridRenderCellParams<R>;

export default function UserTypesList() {
  const navigate = useNavigate();
  const {t} = useTranslation();
  const dataGridLocaleText = useDataGridLocaleText();

  const {
    data: userTypesData,
    loading: isUserTypesRequestLoading,
    error: userTypesRequestError,
    refetch,
  } = useGetUserTypes();
  const {deleteUserType, loading: isDeleting, error: deleteUserTypeError} = useDeleteUserType();
  const {
    data: organizationUnitsResponse,
    isLoading: organizationUnitsLoading,
    error: organizationUnitsError,
  } = useGetOrganizationUnits();

  const error = userTypesRequestError ?? organizationUnitsError;
  const isLoading = isUserTypesRequestLoading || organizationUnitsLoading;
  const organizationUnits = useMemo(() => organizationUnitsResponse?.organizationUnits ?? [], [organizationUnitsResponse]);
  const organizationUnitMap = useMemo(() => {
    const map = new Map<string, string>();
    organizationUnits.forEach((unit) => {
      map.set(unit.id, unit.name);
    });
    return map;
  }, [organizationUnits]);

  const [snackbarOpen, setSnackbarOpen] = useState(false);
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [selectedUserTypeId, setSelectedUserTypeId] = useState<string | null>(null);
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

  const handleMenuOpen = useCallback((event: React.MouseEvent<HTMLElement>, userTypeId: string) => {
    event.stopPropagation();
    setAnchorEl(event.currentTarget);
    setSelectedUserTypeId(userTypeId);
  }, []);

  const handleMenuClose = () => {
    setAnchorEl(null);
  };

  const handleViewUserType = async () => {
    if (selectedUserTypeId) {
      await navigate(`/user-types/${selectedUserTypeId}`);
    }
    handleMenuClose();
  };

  const handleDeleteClick = () => {
    setDeleteDialogOpen(true);
    handleMenuClose();
  };

  const handleDeleteCancel = () => {
    setDeleteDialogOpen(false);
    setSelectedUserTypeId(null);
  };

  const handleDeleteConfirm = async () => {
    if (!selectedUserTypeId) return;

    try {
      await deleteUserType(selectedUserTypeId);
      setDeleteDialogOpen(false);
      setSelectedUserTypeId(null);
      // Refetch user types list after successful deletion
      await refetch();
    } catch {
      // Error is already handled in the hook
      setDeleteDialogOpen(false);
    }
  };

  const handleRowClick = useCallback(
    async (userTypeId: string) => {
      await navigate(`/user-types/${userTypeId}`);
    },
    [navigate],
  );

  const columns: GridColDef<UserSchemaListItem>[] = useMemo(
    () => [
      {
        field: 'name',
        headerName: t('common:form.name'),
        flex: 1,
        minWidth: 200,
        valueGetter: (_value, row) => row.name ?? null,
      },
      {
        field: 'id',
        headerName: 'ID',
        flex: 1,
        minWidth: 250,
        valueGetter: (_value, row) => row.id ?? null,
      },
      {
        field: 'ou',
        headerName: t('userTypes:organizationUnit'),
        flex: 1,
        minWidth: 220,
        renderCell: (params: GridRenderCellParams<UserSchemaListItem>) => {
          const resolvedUnitName = params.row.ouId ? organizationUnitMap.get(params.row.ouId) : undefined;
          const content = (() => {
            if (!params.row.ouId) {
              return t('common:messages.noData');
            }
            if (!resolvedUnitName) {
              return params.row.ouId;
            }
            return resolvedUnitName;
          })();

          return (
            <Box sx={{display: 'flex', alignItems: 'center', width: '100%', height: '100%'}}>
              <Typography
                variant="body2"
                sx={{
                  fontFamily: !resolvedUnitName && params.row.ouId ? 'monospace' : undefined,
                  fontSize: '0.875rem',
                  width: '100%'
                }}
              >
                {content}
              </Typography>
            </Box>
          );
        },
      },
      {
        field: 'allowSelfRegistration',
        headerName: t('userTypes:allowSelfRegistration'),
        width: 200,
        renderCell: (params: GridRenderCellParams<UserSchemaListItem>) => (
          <Chip
            label={params.row.allowSelfRegistration ? t('common:status.enabled') : t('common:status.disabled')}
            color={params.row.allowSelfRegistration ? 'success' : 'default'}
            size="small"
          />
        ),
      },
      {
        field: 'actions',
        headerName: t('users:actions'),
        width: 80,
        sortable: false,
        filterable: false,
        hideable: false,
        renderCell: (params: GridRenderCellParams<UserSchemaListItem>) => (
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
    [organizationUnitMap, t, handleMenuOpen],
  );

  return (
    <>
      <Box sx={{height: 600, width: '100%'}}>
        <DataGrid.DataGrid
          rows={userTypesData?.schemas ?? []}
          columns={columns}
          loading={isLoading}
          getRowId={(row) => row.id}
          onRowClick={(params) => {
            const userTypeId = (params.row as UserSchemaListItem).id;
            handleRowClick(userTypeId).catch(() => {
              // Handle error
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
        <MenuItem
          onClick={() => {
            handleViewUserType().catch(() => {
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
          <ListItemIcon>
            <Trash2 size={16} color="red" />
          </ListItemIcon>
          <ListItemText sx={{color: 'error.main'}}>{t('common:actions.delete')}</ListItemText>
        </MenuItem>
      </Menu>

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onClose={handleDeleteCancel}>
        <DialogTitle>{t('userTypes:deleteUserType')}</DialogTitle>
        <DialogContent>
          <DialogContentText>{t('userTypes:confirmDeleteUserType')}</DialogContentText>
          {deleteUserTypeError && (
            <Alert severity="error" sx={{mt: 2}}>
              <Typography variant="body2" sx={{fontWeight: 'bold'}}>
                {deleteUserTypeError.message}
              </Typography>
              {deleteUserTypeError.description && (
                <Typography variant="caption">{deleteUserTypeError.description}</Typography>
              )}
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
        anchorOrigin={{vertical: 'bottom', horizontal: 'right'}}
      >
        <Alert onClose={handleCloseSnackbar} severity="error" sx={{width: '100%'}}>
          {error?.message ?? t('common:messages.saveError')}
        </Alert>
      </Snackbar>
    </>
  );
}
