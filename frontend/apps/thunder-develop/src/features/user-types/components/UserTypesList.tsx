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
import {useLogger} from '@thunder/logger/react';
import {
  Avatar,
  Box,
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
  useTheme,
} from '@wso2/oxygen-ui';
import {Eye, Trash2, UserRoundCog} from '@wso2/oxygen-ui-icons-react';
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
  const theme = useTheme();
  const navigate = useNavigate();
  const {t} = useTranslation();
  const logger = useLogger('UserTypesList');
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
  const organizationUnits = useMemo(
    () => organizationUnitsResponse?.organizationUnits ?? [],
    [organizationUnitsResponse],
  );
  const organizationUnitMap = useMemo(() => {
    const map = new Map<string, string>();
    organizationUnits.forEach((unit) => {
      map.set(unit.id, unit.name);
    });
    return map;
  }, [organizationUnits]);

  const [snackbarOpen, setSnackbarOpen] = useState(false);
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

  const handleDeleteClick = useCallback((userTypeId: string): void => {
    setSelectedUserTypeId(userTypeId);
    setDeleteDialogOpen(true);
  }, []);

  const handleViewClick = useCallback(
    (userTypeId: string): void => {
      (async (): Promise<void> => {
        await navigate(`/user-types/${userTypeId}`);
      })().catch((_error: unknown) => {
        logger.error('Failed to navigate to user type', {error: _error, userTypeId});
      });
    },
    [logger, navigate],
  );

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

  const columns: GridColDef<UserSchemaListItem>[] = useMemo(
    () => [
      {
        field: 'name',
        headerName: t('common:edit.general.name.label'),
        flex: 1.5,
        minWidth: 220,
        renderCell: (params: GridRenderCellParams<UserSchemaListItem>) => (
          <ListingTable.CellIcon
            sx={{width: '100%'}}
            icon={
              <Avatar
                sx={{
                  backgroundColor: theme.vars?.palette.grey[500],
                  width: 30,
                  height: 30,
                  fontSize: '0.875rem',
                  ...theme.applyStyles('dark', {
                    backgroundColor: theme.vars?.palette.grey[900],
                  }),
                }}
              >
                <UserRoundCog size={14} />
              </Avatar>
            }
            primary={params.row.name ?? '-'}
          />
        ),
      },
      {
        field: 'id',
        headerName: 'ID',
        width: 350,
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
                  width: '100%',
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
        width: 150,
        align: 'center',
        headerAlign: 'center',
        sortable: false,
        filterable: false,
        hideable: false,
        renderCell: (params: GridRenderCellParams<UserSchemaListItem>) => (
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
    [organizationUnitMap, t, handleDeleteClick, handleViewClick, theme],
  );

  return (
    <>
      <ListingTable.Provider variant="data-grid-card" loading={isLoading}>
        <ListingTable.Container disablePaper>
          <ListingTable.DataGrid
            rows={userTypesData?.schemas ?? []}
            columns={columns}
            getRowId={(row) => (row as UserSchemaListItem).id}
            onRowClick={(params) => {
              handleViewClick((params.row as UserSchemaListItem).id);
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
