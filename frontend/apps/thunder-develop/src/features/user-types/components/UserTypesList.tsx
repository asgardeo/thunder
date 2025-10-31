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

import {useCallback, useEffect, useState} from 'react';
import {useNavigate} from 'react-router';
import Box from '@mui/material/Box';
import IconButton from '@mui/material/IconButton';
import Typography from '@mui/material/Typography';
import Snackbar from '@mui/material/Snackbar';
import Alert from '@mui/material/Alert';
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import {DataGrid, type GridColDef, type GridRenderCellParams} from '@mui/x-data-grid';
import {EllipsisVertical, Trash2, Eye} from 'lucide-react';
import useGetUserTypes from '../api/useGetUserTypes';
import useDeleteUserType from '../api/useDeleteUserType';
import type {UserSchemaListItem} from '../types/user-types';

export default function UserTypesList() {
  const navigate = useNavigate();

  const {data: userTypesData, loading: isUserTypesRequestLoading, error: userTypesRequestError, refetch} = useGetUserTypes();
  const {deleteUserType, loading: isDeleting, error: deleteUserTypeError} = useDeleteUserType();

  const error = userTypesRequestError;
  const isLoading = isUserTypesRequestLoading;

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

  const handleRowClick = useCallback(async (userTypeId: string) => {
    await navigate(`/user-types/${userTypeId}`);
  }, [navigate]);

  const columns: GridColDef<UserSchemaListItem>[] = [
    {
      field: 'name',
      headerName: 'Name',
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
      field: 'actions',
      headerName: 'Actions',
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
  ];

  return (
    <>
      <Box sx={{height: 600, width: '100%'}}>
        <DataGrid
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
          <ListItemText>View</ListItemText>
        </MenuItem>
        <MenuItem onClick={handleDeleteClick}>
          <ListItemIcon>
            <Trash2 size={16} color="red" />
          </ListItemIcon>
          <ListItemText sx={{color: 'error.main'}}>Delete</ListItemText>
        </MenuItem>
      </Menu>

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onClose={handleDeleteCancel}>
        <DialogTitle>Delete User Type</DialogTitle>
        <DialogContent>
          <DialogContentText>
            Are you sure you want to delete this user type? This action cannot be undone.
          </DialogContentText>
          {deleteUserTypeError && (
            <Alert severity="error" sx={{mt: 2}}>
              <Typography variant="body2" sx={{fontWeight: 'bold'}}>
                {deleteUserTypeError.message}
              </Typography>
              {deleteUserTypeError.description && <Typography variant="caption">{deleteUserTypeError.description}</Typography>}
            </Alert>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={handleDeleteCancel} disabled={isDeleting}>
            Cancel
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
            {isDeleting ? 'Deleting...' : 'Delete'}
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
          {error?.message ?? 'An error occurred while loading data'}
        </Alert>
      </Snackbar>
    </>
  );
}
