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
  TextField,
  InputAdornment,
  Box,
  Alert,
  DataGrid,
  Avatar,
  useTheme,
} from '@wso2/oxygen-ui';
import {Search, User} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import useGetUsers from '../../../../users/api/useGetUsers';
import type {ApiUser} from '../../../../users/types/users';
import useDataGridLocaleText from '../../../../../hooks/useDataGridLocaleText';
import type {Member} from '../../../models/group';

interface AddMemberDialogProps {
  open: boolean;
  groupId: string;
  onClose: () => void;
  onAdd: (members: Member[]) => void;
}

/**
 * Dialog for searching and adding user members to a group.
 */
export default function AddMemberDialog({open, groupId, onClose, onAdd}: AddMemberDialogProps): JSX.Element {
  const {t} = useTranslation();
  const theme = useTheme();
  const dataGridLocaleText = useDataGridLocaleText();
  const [searchQuery, setSearchQuery] = useState('');
  const [selectionModel, setSelectionModel] = useState<DataGrid.GridRowSelectionModel>({
    type: 'include',
    ids: new Set(),
  });

  const usersParams = useMemo(() => ({limit: 50, excludeGroupId: groupId}), [groupId]);
  const {data: usersData, loading: usersLoading} = useGetUsers(usersParams);

  const users: ApiUser[] = useMemo(() => usersData?.users ?? [], [usersData]);

  const filteredUsers = useMemo(() => {
    if (!searchQuery.trim()) return users;
    const query = searchQuery.toLowerCase();
    return users.filter((user) => user.id.toLowerCase().includes(query));
  }, [users, searchQuery]);

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
        headerName: 'User ID',
        flex: 1,
        minWidth: 250,
      },
    ],
    [theme],
  );

  const handleAdd = useCallback(() => {
    const newMembers: Member[] = [...selectionModel.ids].map((id) => ({id: String(id), type: 'user' as const}));
    onAdd(newMembers);
    setSelectionModel({type: 'include', ids: new Set()});
    setSearchQuery('');
  }, [selectionModel, onAdd]);

  const handleClose = (): void => {
    setSelectionModel({type: 'include', ids: new Set()});
    setSearchQuery('');
    onClose();
  };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle>{t('groups:addMember.title')}</DialogTitle>
      <DialogContent>
        <TextField
          placeholder={t('groups:addMember.search.placeholder')}
          size="small"
          fullWidth
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          sx={{my: 2}}
          slotProps={{
            input: {
              startAdornment: (
                <InputAdornment position="start">
                  <Search size={16} />
                </InputAdornment>
              ),
            },
          }}
        />

        {filteredUsers.length === 0 && !usersLoading && (
          <Alert severity="info" sx={{mb: 2}}>
            {t('groups:addMember.noResults')}
          </Alert>
        )}

        <Box sx={{height: 350, width: '100%'}}>
          <DataGrid.DataGrid
            rows={filteredUsers}
            columns={columns}
            loading={usersLoading}
            getRowId={(row): string => row.id}
            checkboxSelection
            rowSelectionModel={selectionModel}
            onRowSelectionModelChange={(newSelection) => {
              setSelectionModel(newSelection);
            }}
            initialState={{
              pagination: {
                paginationModel: {pageSize: 10},
              },
            }}
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
