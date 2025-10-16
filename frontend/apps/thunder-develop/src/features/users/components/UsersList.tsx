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

import {useEffect, useMemo, useState} from 'react';
import {useNavigate} from 'react-router';
import Box from '@mui/material/Box';
import Avatar from '@mui/material/Avatar';
import Chip from '@mui/material/Chip';
import IconButton from '@mui/material/IconButton';
import Typography from '@mui/material/Typography';
import Snackbar from '@mui/material/Snackbar';
import Alert from '@mui/material/Alert';
import MoreVertIcon from '@mui/icons-material/MoreVert';
import {DataGrid, type GridColDef, type GridRenderCellParams} from '@mui/x-data-grid';
import useGetUsers from '../api/useGetUsers';
import useGetUserSchema from '../api/useGetUserSchema';
import type {UserWithDetails} from '../types/users';

interface UsersListProps {
  selectedSchema: string;
}

export default function UsersList(props: UsersListProps) {
  const {selectedSchema} = props;

  const {data: userData, loading: isUsersRequestLoading, error: usersRequestError} = useGetUsers();

  const {
    data: defaultUserSchema,
    loading: isDefaultUserSchemaRequestLoading,
    error: defaultUserSchemaRequestError,
  } = useGetUserSchema(selectedSchema);

  const error = usersRequestError ?? defaultUserSchemaRequestError;
  const isLoading = isUsersRequestLoading || isDefaultUserSchemaRequestLoading;

  const [snackbarOpen, setSnackbarOpen] = useState(false);

  const navigate = useNavigate();

  // Show snackbar when error occurs
  useEffect(() => {
    if (error) {
      setSnackbarOpen(true);
    }
  }, [error]);

  const handleCloseSnackbar = () => {
    setSnackbarOpen(false);
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

  const columns: GridColDef<UserWithDetails>[] = useMemo(() => {
    if (!defaultUserSchema) {
      // Return basic columns if schema is not loaded yet
      return [];
    }

    const schemaColumns: GridColDef<UserWithDetails>[] = [];
    const schemaEntries = Object.entries(defaultUserSchema.schema);

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
        renderCell: (params: GridRenderCellParams<UserWithDetails>) => {
          const firstname = params.row.attributes?.firstname as string | undefined;
          const lastname = params.row.attributes?.lastname as string | undefined;
          const username = params.row.attributes?.username as string | undefined;
          const displayName = [firstname, lastname, username].filter(Boolean).join(' ') || '?';
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
          renderCell: (params: GridRenderCellParams<UserWithDetails>) => {
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
          renderCell: (params: GridRenderCellParams<UserWithDetails>) => {
            const value = params.row.attributes?.[fieldName] as boolean | string | undefined;
            if (value === undefined || value === null) return null;

            const isActive = typeof value === 'boolean' ? value : value === 'active';
            return (
              <Chip label={isActive ? 'Active' : 'Inactive'} size="small" color={isActive ? 'success' : 'default'} />
            );
          },
        });
        return;
      }

      // Handle different field types
      const columnDef: GridColDef<UserWithDetails> = {
        field: fieldName,
        headerName: formatHeaderName(fieldName),
        flex: 1,
        minWidth: 150,
      };

      // Type-specific configuration
      switch (fieldDef.type) {
        case 'boolean':
          columnDef.type = 'boolean';
          columnDef.renderCell = (params: GridRenderCellParams<UserWithDetails>) => {
            const value = params.row.attributes?.[fieldName] as boolean | undefined;
            if (value === undefined || value === null) return '-';
            return value ? 'Yes' : 'No';
          };
          break;

        case 'number':
          columnDef.type = 'number';
          columnDef.valueGetter = (_value, row) => {
            const value = row.attributes?.[fieldName] as number | undefined;
            return value ?? null;
          };
          break;

        case 'array':
          columnDef.sortable = false;
          columnDef.renderCell = (params: GridRenderCellParams<UserWithDetails>) => {
            const value = params.row.attributes?.[fieldName] as unknown[] | undefined;
            if (!value || !Array.isArray(value) || value.length === 0) return '-';
            return value.join(', ');
          };
          break;

        case 'object':
          columnDef.sortable = false;
          columnDef.renderCell = (params: GridRenderCellParams<UserWithDetails>) => {
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
      headerName: 'Actions',
      width: 80,
      sortable: false,
      filterable: false,
      hideable: false,
      renderCell: () => (
        <IconButton
          size="small"
          onClick={(e) => {
            e.stopPropagation();
            // Add action handler here
          }}
        >
          <MoreVertIcon />
        </IconButton>
      ),
    });

    return schemaColumns;
  }, [defaultUserSchema]);

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
        <DataGrid
          rows={userData?.users}
          columns={columns}
          loading={isLoading}
          getRowId={(row) => row.id}
          onRowClick={(params) => {
            const userId = (params.row as UserWithDetails).id;
            navigate(`/users/${userId}`);
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
          sx={{
            '& .MuiDataGrid-row': {
              cursor: 'pointer',
            },
          }}
        />
      </Box>
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
