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

import {useState, useMemo, useCallback, type JSX} from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Box,
  Alert,
  DataGrid,
  Avatar,
  useTheme,
} from '@wso2/oxygen-ui';
import {User} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import useGetUsers from '../../../../users/api/useGetUsers';
import type {ApiUser} from '../../../../users/types/users';
import useDataGridLocaleText from '../../../../../hooks/useDataGridLocaleText';
import type {Member} from '../../../models/group';

interface AddMemberDialogProps {
  open: boolean;
  onClose: () => void;
  onAdd: (members: Member[]) => void;
}

/**
 * Dialog for searching and adding user members to a group.
 */
export default function AddMemberDialog({open, onClose, onAdd}: AddMemberDialogProps): JSX.Element {
  const {t} = useTranslation();
  const theme = useTheme();
  const dataGridLocaleText = useDataGridLocaleText();
  const [selectionModel, setSelectionModel] = useState<DataGrid.GridRowSelectionModel>({
    type: 'include',
    ids: new Set(),
  });
  const [paginationModel, setPaginationModel] = useState<DataGrid.GridPaginationModel>({pageSize: 10, page: 0});

  const usersParams = useMemo(
    () => ({
      limit: paginationModel.pageSize,
      offset: paginationModel.page * paginationModel.pageSize,
    }),
    [paginationModel],
  );
  const {data: usersData, loading: usersLoading, error: usersError} = useGetUsers(usersParams);

  const users: ApiUser[] = useMemo(() => usersData?.users ?? [], [usersData]);

  const columns: DataGrid.GridColDef<ApiUser>[] = useMemo(
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
              <User size={14} />
            </Avatar>
          </Box>
        ),
      },
      {
        field: 'id',
        headerName: t('groups:addMember.columns.userId'),
        flex: 1,
        minWidth: 250,
      },
    ],
    [theme, t],
  );

  const handleAdd = useCallback(() => {
    const newMembers: Member[] = [...selectionModel.ids].map((id) => ({id: String(id), type: 'user' as const}));
    onAdd(newMembers);
    setSelectionModel({type: 'include', ids: new Set()});
  }, [selectionModel, onAdd]);

  const handleClose = (): void => {
    setSelectionModel({type: 'include', ids: new Set()});
    onClose();
  };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle>{t('groups:addMember.title')}</DialogTitle>
      <DialogContent>
        {usersError && !usersLoading && (
          <Alert severity="error" sx={{mb: 2}}>
            {usersError.message ?? t('groups:addMember.fetchError')}
          </Alert>
        )}
        {!usersError && users.length === 0 && !usersLoading && (
          <Alert severity="info" sx={{mb: 2}}>
            {t('groups:addMember.noResults')}
          </Alert>
        )}

        <Box sx={{height: 400, width: '100%'}}>
          <DataGrid.DataGrid
            rows={users}
            columns={columns}
            loading={usersLoading}
            getRowId={(row): string => row.id}
            checkboxSelection
            rowSelectionModel={selectionModel}
            onRowSelectionModelChange={(newSelection) => {
              setSelectionModel(newSelection);
            }}
            paginationMode="server"
            rowCount={usersData?.totalResults ?? 0}
            paginationModel={paginationModel}
            onPaginationModelChange={setPaginationModel}
            pageSizeOptions={[5, 10]}
            disableRowSelectionOnClick
            localeText={dataGridLocaleText}
          />
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose}>{t('common:actions.cancel')}</Button>
        <Button variant="contained" onClick={handleAdd} disabled={selectionModel.ids.size === 0}>
          {t('groups:addMember.add')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
