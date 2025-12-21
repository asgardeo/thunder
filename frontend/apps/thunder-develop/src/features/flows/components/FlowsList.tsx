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
import {GitBranch, EllipsisVertical, Eye, Trash2} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import useDataGridLocaleText from '../../../hooks/useDataGridLocaleText';
import useGetFlows from '../api/useGetFlows';
import type {BasicFlowDefinition} from '../models/responses';
import FlowDeleteDialog from './FlowDeleteDialog';

export default function FlowsList(): JSX.Element {
  const theme = useTheme();
  const navigate = useNavigate();
  const {t} = useTranslation();
  const dataGridLocaleText = useDataGridLocaleText();
  const {data, isLoading, error} = useGetFlows();

  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [selectedFlow, setSelectedFlow] = useState<BasicFlowDefinition | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState<boolean>(false);

  const handleMenuOpen = useCallback((event: React.MouseEvent<HTMLElement>, flow: BasicFlowDefinition) => {
    event.stopPropagation();
    setAnchorEl(event.currentTarget);
    setSelectedFlow(flow);
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
    setSelectedFlow(null);
  };

  const handleViewClick = (): void => {
    handleMenuClose();
    // Only authentication flows are editable for now
    if (selectedFlow?.flowType === 'AUTHENTICATION') {
      (async (): Promise<void> => {
        await navigate(`/flows/login/${selectedFlow.id}`);
      })().catch(() => {
        // TODO: Log the errors
      });
    }
  };

  const columns: DataGrid.GridColDef<BasicFlowDefinition>[] = useMemo(
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
              <GitBranch size={14} />
            </Avatar>
          </Box>
        ),
      },
      {
        field: 'name',
        headerName: t('flows:listing.columns.name'),
        flex: 1,
        minWidth: 200,
      },
      {
        field: 'flowType',
        headerName: t('flows:listing.columns.flowType'),
        flex: 1,
        minWidth: 150,
        renderCell: (params: DataGrid.GridRenderCellParams<BasicFlowDefinition>): JSX.Element => (
          <Chip
            label={params.row.flowType}
            size="small"
            color="primary"
            variant="outlined"
            sx={{
              fontSize: '0.7rem',
            }}
          />
        ),
      },
      {
        field: 'activeVersion',
        headerName: t('flows:listing.columns.version'),
        width: 100,
        renderCell: (params: DataGrid.GridRenderCellParams<BasicFlowDefinition>): JSX.Element => (
          <Chip
            label={`v${params.row.activeVersion}`}
            size="small"
            variant="outlined"
            sx={{
              fontFamily: 'monospace',
              fontSize: '0.7rem',
            }}
          />
        ),
      },
      {
        field: 'updatedAt',
        headerName: t('flows:listing.columns.updatedAt'),
        flex: 1,
        minWidth: 180,
        valueGetter: (_value, row): string => {
          const date = new Date(row.updatedAt);
          return date.toLocaleDateString(undefined, {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
          });
        },
      },
      {
        field: 'actions',
        headerName: t('flows:listing.columns.actions'),
        width: 80,
        sortable: false,
        filterable: false,
        hideable: false,
        renderCell: (params: DataGrid.GridRenderCellParams<BasicFlowDefinition>): JSX.Element => (
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
          {t('flows:listing.error.title')}
        </Typography>
        <Typography variant="body2" color="text.secondary">
          {error.message ?? t('flows:listing.error.unknown')}
        </Typography>
      </Box>
    );
  }

  return (
    <>
      <Box sx={{height: 600, width: '100%'}}>
        <DataGrid.DataGrid
          rows={data?.flows ?? []}
          columns={columns}
          loading={isLoading}
          getRowId={(row): string => row.id}
          onRowClick={(params) => {
            const flow = params.row as BasicFlowDefinition;
            // Only authentication flows are editable for now
            if (flow.flowType !== 'AUTHENTICATION') {
              return;
            }
            (async (): Promise<void> => {
              await navigate(`/flows/login/${flow.id}`);
            })().catch(() => {
              // TODO: Log the errors
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
          getRowClassName={(params) =>
            params.row.flowType === 'AUTHENTICATION' ? 'row-clickable' : 'row-not-clickable'
          }
          sx={{
            '& .MuiDataGrid-row.row-clickable': {
              cursor: 'pointer',
            },
            '& .MuiDataGrid-row.row-not-clickable': {
              cursor: 'default',
            },
          }}
        />
      </Box>

      {/* Actions Menu */}
      <Menu anchorEl={anchorEl} open={Boolean(anchorEl)} onClose={handleMenuClose}>
        {selectedFlow?.flowType === 'AUTHENTICATION' && (
          <MenuItem onClick={handleViewClick}>
            <ListItemIcon>
              <Eye size={16} />
            </ListItemIcon>
            <ListItemText>{t('common:actions.view')}</ListItemText>
          </MenuItem>
        )}
        <MenuItem onClick={handleDeleteClick}>
          <ListItemIcon>
            <Trash2 size={16} color={theme.vars?.palette.error.main} />
          </ListItemIcon>
          <ListItemText sx={{color: 'error.main'}}>{t('common:actions.delete')}</ListItemText>
        </MenuItem>
      </Menu>

      <FlowDeleteDialog open={deleteDialogOpen} flowId={selectedFlow?.id ?? null} onClose={handleDeleteDialogClose} />
    </>
  );
}
