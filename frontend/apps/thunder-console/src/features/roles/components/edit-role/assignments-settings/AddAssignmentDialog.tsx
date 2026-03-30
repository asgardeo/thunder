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
  Chip,
  Typography,
  Tabs,
  Tab,
  useTheme,
} from '@wso2/oxygen-ui';
import {User, Users} from '@wso2/oxygen-ui-icons-react';
import {useState, useMemo, useCallback, type JSX, type SyntheticEvent} from 'react';
import {useTranslation} from 'react-i18next';
import useDataGridLocaleText from '../../../../../hooks/useDataGridLocaleText';
import useGetGroups from '../../../../groups/api/useGetGroups';
import type {GroupBasic} from '../../../../groups/models/group';
import useGetUsers from '../../../../users/api/useGetUsers';
import type {ApiUser} from '../../../../users/types/users';
import useGetRoleAssignments from '../../../api/useGetRoleAssignments';
import type {RoleAssignment} from '../../../models/role';

interface AddAssignmentDialogProps {
  open: boolean;
  roleId: string;
  onClose: () => void;
  onAdd: (assignments: RoleAssignment[]) => void;
  initialTab?: number;
}

/**
 * Dialog for searching and adding user or group assignments to a role.
 */
export default function AddAssignmentDialog({open, roleId, onClose, onAdd, initialTab = 0}: AddAssignmentDialogProps): JSX.Element {
  const {t} = useTranslation();
  const theme = useTheme();
  const dataGridLocaleText = useDataGridLocaleText();

  const [activeTab, setActiveTab] = useState(initialTab);
  const [userSelectionModel, setUserSelectionModel] = useState<DataGrid.GridRowSelectionModel>({
    type: 'include',
    ids: new Set(),
  });
  const [groupSelectionModel, setGroupSelectionModel] = useState<DataGrid.GridRowSelectionModel>({
    type: 'include',
    ids: new Set(),
  });
  const [userPaginationModel, setUserPaginationModel] = useState<DataGrid.GridPaginationModel>({pageSize: 10, page: 0});
  const [groupPaginationModel, setGroupPaginationModel] = useState<DataGrid.GridPaginationModel>({
    pageSize: 10,
    page: 0,
  });

  const usersParams = useMemo(
    () => ({limit: userPaginationModel.pageSize, offset: userPaginationModel.page * userPaginationModel.pageSize}),
    [userPaginationModel],
  );
  const groupsParams = useMemo(
    () => ({limit: groupPaginationModel.pageSize, offset: groupPaginationModel.page * groupPaginationModel.pageSize}),
    [groupPaginationModel],
  );

  const {data: usersData, isLoading: usersLoading, error: usersError} = useGetUsers(usersParams);
  const {data: groupsData, isLoading: groupsLoading, error: groupsError} = useGetGroups(groupsParams);

  const {data: existingUserAssignments} = useGetRoleAssignments({roleId, type: 'user', limit: 100, offset: 0});
  const {data: existingGroupAssignments} = useGetRoleAssignments({roleId, type: 'group', limit: 100, offset: 0});

  const assignedUserIds = useMemo(
    () => new Set(existingUserAssignments?.assignments.map((a) => a.id) ?? []),
    [existingUserAssignments],
  );
  const assignedGroupIds = useMemo(
    () => new Set(existingGroupAssignments?.assignments.map((a) => a.id) ?? []),
    [existingGroupAssignments],
  );

  const filteredUsers = useMemo(
    () => (usersData?.users ?? []).filter((u) => !assignedUserIds.has(u.id)),
    [usersData, assignedUserIds],
  );
  const filteredGroups = useMemo(
    () => (groupsData?.groups ?? []).filter((g) => !assignedGroupIds.has(g.id)),
    [groupsData, assignedGroupIds],
  );

  const userColumns: DataGrid.GridColDef<ApiUser>[] = useMemo(
    () => [
      {
        field: 'avatar',
        headerName: '',
        width: 70,
        sortable: false,
        filterable: false,
        renderCell: (): JSX.Element => (
          <Box sx={{display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%'}}>
            <Avatar
              sx={{
                p: 0.5,
                backgroundColor: theme.vars?.palette.grey[500],
                width: 30,
                height: 30,
                fontSize: '0.875rem',
                ...theme.applyStyles('dark', {backgroundColor: theme.vars?.palette.grey[900]}),
              }}
            >
              <User size={14} />
            </Avatar>
          </Box>
        ),
      },
      {
        field: 'display',
        headerName: t('roles:assignments.dialog.columns.displayName'),
        flex: 1,
        minWidth: 200,
        renderCell: (params: DataGrid.GridRenderCellParams<ApiUser>): JSX.Element => (
          <Box sx={{display: 'flex', flexDirection: 'column', justifyContent: 'center', height: '100%', overflow: 'hidden'}}>
            <Typography variant="body2" noWrap>
              {params.row.display ?? params.row.id}
            </Typography>
            <Typography variant="caption" color="text.secondary" noWrap sx={{fontFamily: 'monospace', fontSize: '0.7rem'}}>
              {params.row.id}
            </Typography>
          </Box>
        ),
      },
      {
        field: 'type',
        headerName: t('roles:assignments.dialog.columns.userType'),
        width: 150,
        renderCell: (params: DataGrid.GridRenderCellParams<ApiUser>): JSX.Element => (
          <Chip label={params.row.type} size="small" variant="outlined" sx={{textTransform: 'capitalize'}} />
        ),
      },
    ],
    [theme, t],
  );

  const groupColumns: DataGrid.GridColDef<GroupBasic>[] = useMemo(
    () => [
      {
        field: 'avatar',
        headerName: '',
        width: 70,
        sortable: false,
        filterable: false,
        renderCell: (): JSX.Element => (
          <Box sx={{display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%'}}>
            <Avatar
              sx={{
                p: 0.5,
                backgroundColor: theme.vars?.palette.grey[500],
                width: 30,
                height: 30,
                fontSize: '0.875rem',
                ...theme.applyStyles('dark', {backgroundColor: theme.vars?.palette.grey[900]}),
              }}
            >
              <Users size={14} />
            </Avatar>
          </Box>
        ),
      },
      {
        field: 'name',
        headerName: t('roles:assignments.dialog.columns.name'),
        flex: 1,
        minWidth: 200,
      },
      {
        field: 'description',
        headerName: t('roles:assignments.dialog.columns.description'),
        flex: 1,
        minWidth: 200,
        valueGetter: (_value, row): string => row.description ?? '-',
      },
    ],
    [theme, t],
  );

  const handleAdd = useCallback(() => {
    const userAssignments: RoleAssignment[] = [...userSelectionModel.ids].map((id) => ({
      id: String(id),
      type: 'user' as const,
    }));
    const groupAssignments: RoleAssignment[] = [...groupSelectionModel.ids].map((id) => ({
      id: String(id),
      type: 'group' as const,
    }));
    onAdd([...userAssignments, ...groupAssignments]);
    setUserSelectionModel({type: 'include', ids: new Set()});
    setGroupSelectionModel({type: 'include', ids: new Set()});
  }, [userSelectionModel, groupSelectionModel, onAdd]);

  const handleClose = (): void => {
    setUserSelectionModel({type: 'include', ids: new Set()});
    setGroupSelectionModel({type: 'include', ids: new Set()});
    onClose();
  };

  const totalSelected = userSelectionModel.ids.size + groupSelectionModel.ids.size;

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="md" fullWidth>
      <DialogTitle>{t('roles:assignments.dialog.title')}</DialogTitle>
      <DialogContent>
        <Tabs value={activeTab} onChange={(_e: SyntheticEvent, v: number) => setActiveTab(v)} sx={{mb: 2}}>
          <Tab label={t('roles:assignments.dialog.tabs.users')} />
          <Tab label={t('roles:assignments.dialog.tabs.groups')} />
        </Tabs>

        {activeTab === 0 && (
          <>
            {usersError && !usersLoading && (
              <Alert severity="error" sx={{mb: 2}}>
                {usersError.message ?? t('roles:assignments.dialog.fetchError')}
              </Alert>
            )}
            <Box sx={{height: 400, width: '100%'}}>
              <DataGrid.DataGrid
                rows={filteredUsers}
                columns={userColumns}
                loading={usersLoading}
                getRowId={(row): string => row.id}
                checkboxSelection
                rowSelectionModel={userSelectionModel}
                onRowSelectionModelChange={setUserSelectionModel}
                paginationMode="server"
                rowCount={usersData?.totalResults ?? 0}
                paginationModel={userPaginationModel}
                onPaginationModelChange={setUserPaginationModel}
                pageSizeOptions={[5, 10]}
                disableRowSelectionOnClick
                localeText={dataGridLocaleText}
              />
            </Box>
          </>
        )}

        {activeTab === 1 && (
          <>
            {groupsError && !groupsLoading && (
              <Alert severity="error" sx={{mb: 2}}>
                {groupsError.message ?? t('roles:assignments.dialog.fetchError')}
              </Alert>
            )}
            <Box sx={{height: 400, width: '100%'}}>
              <DataGrid.DataGrid
                rows={filteredGroups}
                columns={groupColumns}
                loading={groupsLoading}
                getRowId={(row): string => row.id}
                checkboxSelection
                rowSelectionModel={groupSelectionModel}
                onRowSelectionModelChange={setGroupSelectionModel}
                paginationMode="server"
                rowCount={groupsData?.totalResults ?? 0}
                paginationModel={groupPaginationModel}
                onPaginationModelChange={setGroupPaginationModel}
                pageSizeOptions={[5, 10]}
                disableRowSelectionOnClick
                localeText={dataGridLocaleText}
              />
            </Box>
          </>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose}>{t('common:actions.cancel')}</Button>
        <Button variant="contained" onClick={handleAdd} disabled={totalSelected === 0}>
          {t('roles:assignments.dialog.add')} {totalSelected > 0 ? `(${totalSelected})` : ''}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
