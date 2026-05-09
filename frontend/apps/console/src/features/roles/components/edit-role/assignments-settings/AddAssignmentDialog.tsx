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

import {useGetUsers} from '@thunderid/configure-users';
import {useDataGridLocaleText} from '@thunderid/hooks';
import type {User} from '@thunderid/types';
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
import {AppWindow, Bot, User as UserIcon, Users} from '@wso2/oxygen-ui-icons-react';
import {useState, useMemo, useCallback, type JSX, type SyntheticEvent} from 'react';
import {useTranslation} from 'react-i18next';
import useGetAgents from '../../../../agents/api/useGetAgents';
import type {BasicAgent} from '../../../../agents/models/agent';
import useGetApplications from '../../../../applications/api/useGetApplications';
import type {BasicApplication} from '../../../../applications/models/application';
import useGetGroups from '../../../../groups/api/useGetGroups';
import type {GroupBasic} from '../../../../groups/models/group';
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
 * Dialog for searching and adding user, group, or app assignments to a role.
 */
export default function AddAssignmentDialog({
  open,
  roleId,
  onClose,
  onAdd,
  initialTab = 0,
}: AddAssignmentDialogProps): JSX.Element {
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
  const [appSelectionModel, setAppSelectionModel] = useState<DataGrid.GridRowSelectionModel>({
    type: 'include',
    ids: new Set(),
  });
  const [agentSelectionModel, setAgentSelectionModel] = useState<DataGrid.GridRowSelectionModel>({
    type: 'include',
    ids: new Set(),
  });
  const [userPaginationModel, setUserPaginationModel] = useState<DataGrid.GridPaginationModel>({pageSize: 10, page: 0});
  const [groupPaginationModel, setGroupPaginationModel] = useState<DataGrid.GridPaginationModel>({
    pageSize: 10,
    page: 0,
  });
  const [appPaginationModel, setAppPaginationModel] = useState<DataGrid.GridPaginationModel>({pageSize: 10, page: 0});
  const [agentPaginationModel, setAgentPaginationModel] = useState<DataGrid.GridPaginationModel>({
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
  const applicationsParams = useMemo(
    () => ({limit: appPaginationModel.pageSize, offset: appPaginationModel.page * appPaginationModel.pageSize}),
    [appPaginationModel],
  );
  const agentsParams = useMemo(
    () => ({limit: agentPaginationModel.pageSize, offset: agentPaginationModel.page * agentPaginationModel.pageSize}),
    [agentPaginationModel],
  );

  const {data: usersData, isLoading: usersLoading, error: usersError} = useGetUsers(usersParams);
  const {data: groupsData, isLoading: groupsLoading, error: groupsError} = useGetGroups(groupsParams);
  const {data: applicationsData, isLoading: appsLoading, error: appsError} = useGetApplications(applicationsParams);
  const {data: agentsData, isLoading: agentsLoading, error: agentsError} = useGetAgents(agentsParams);

  const {data: initialUserAssignments} = useGetRoleAssignments({roleId, type: 'user', limit: 1, offset: 0});
  const {data: existingUserAssignments} = useGetRoleAssignments({
    roleId,
    type: 'user',
    limit: initialUserAssignments?.totalResults ?? 0,
    offset: 0,
    enabled: (initialUserAssignments?.totalResults ?? 0) > 0,
  });
  const {data: initialGroupAssignments} = useGetRoleAssignments({roleId, type: 'group', limit: 1, offset: 0});
  const {data: existingGroupAssignments} = useGetRoleAssignments({
    roleId,
    type: 'group',
    limit: initialGroupAssignments?.totalResults ?? 0,
    offset: 0,
    enabled: (initialGroupAssignments?.totalResults ?? 0) > 0,
  });
  const {data: initialAppAssignments} = useGetRoleAssignments({roleId, type: 'app', limit: 1, offset: 0});
  const {data: existingAppAssignments} = useGetRoleAssignments({
    roleId,
    type: 'app',
    limit: initialAppAssignments?.totalResults ?? 0,
    offset: 0,
    enabled: (initialAppAssignments?.totalResults ?? 0) > 0,
  });
  const {data: initialAgentAssignments} = useGetRoleAssignments({roleId, type: 'agent', limit: 1, offset: 0});
  const {data: existingAgentAssignments} = useGetRoleAssignments({
    roleId,
    type: 'agent',
    limit: initialAgentAssignments?.totalResults ?? 0,
    offset: 0,
    enabled: (initialAgentAssignments?.totalResults ?? 0) > 0,
  });

  const assignedUserIds = useMemo(
    () => new Set(existingUserAssignments?.assignments.map((a) => a.id) ?? []),
    [existingUserAssignments],
  );
  const assignedGroupIds = useMemo(
    () => new Set(existingGroupAssignments?.assignments.map((a) => a.id) ?? []),
    [existingGroupAssignments],
  );
  const assignedAppIds = useMemo(
    () => new Set(existingAppAssignments?.assignments.map((a) => a.id) ?? []),
    [existingAppAssignments],
  );
  const assignedAgentIds = useMemo(
    () => new Set(existingAgentAssignments?.assignments.map((a) => a.id) ?? []),
    [existingAgentAssignments],
  );

  const filteredUsers = useMemo(
    () => (usersData?.users ?? []).filter((u) => !assignedUserIds.has(u.id)),
    [usersData, assignedUserIds],
  );
  const filteredGroups = useMemo(
    () => (groupsData?.groups ?? []).filter((g) => !assignedGroupIds.has(g.id)),
    [groupsData, assignedGroupIds],
  );
  const filteredApplications = useMemo(
    () => (applicationsData?.applications ?? []).filter((app) => !assignedAppIds.has(app.id)),
    [applicationsData, assignedAppIds],
  );
  const filteredAgents = useMemo(
    () => (agentsData?.agents ?? []).filter((a) => !assignedAgentIds.has(a.id)),
    [agentsData, assignedAgentIds],
  );

  const userColumns: DataGrid.GridColDef<User>[] = useMemo(
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
              <UserIcon size={14} />
            </Avatar>
          </Box>
        ),
      },
      {
        field: 'display',
        headerName: t('roles:assignments.dialog.columns.displayName'),
        flex: 1,
        minWidth: 200,
        renderCell: (params): JSX.Element => {
          const row = params.row;
          return (
            <Box
              sx={{
                display: 'flex',
                flexDirection: 'column',
                justifyContent: 'center',
                height: '100%',
                overflow: 'hidden',
              }}
            >
              <Typography variant="body2" noWrap>
                {row.display ?? row.id}
              </Typography>
              <Typography
                variant="caption"
                color="text.secondary"
                noWrap
                sx={{fontFamily: 'monospace', fontSize: '0.7rem'}}
              >
                {row.id}
              </Typography>
            </Box>
          );
        },
      },
      {
        field: 'type',
        headerName: t('roles:assignments.dialog.columns.userType'),
        width: 150,
        renderCell: (params): JSX.Element => (
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
  const agentColumns: DataGrid.GridColDef<BasicAgent>[] = useMemo(
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
              <Bot size={14} />
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

  const appColumns: DataGrid.GridColDef<BasicApplication>[] = useMemo(
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
              <AppWindow size={14} />
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
    const appAssignments: RoleAssignment[] = [...appSelectionModel.ids].map((id) => ({
      id: String(id),
      type: 'app' as const,
    }));
    const agentAssignments: RoleAssignment[] = [...agentSelectionModel.ids].map((id) => ({
      id: String(id),
      type: 'agent' as const,
    }));
    onAdd([...userAssignments, ...groupAssignments, ...appAssignments, ...agentAssignments]);
    setUserSelectionModel({type: 'include', ids: new Set()});
    setGroupSelectionModel({type: 'include', ids: new Set()});
    setAppSelectionModel({type: 'include', ids: new Set()});
    setAgentSelectionModel({type: 'include', ids: new Set()});
  }, [userSelectionModel, groupSelectionModel, appSelectionModel, agentSelectionModel, onAdd]);

  const handleClose = (): void => {
    setUserSelectionModel({type: 'include', ids: new Set()});
    setGroupSelectionModel({type: 'include', ids: new Set()});
    setAppSelectionModel({type: 'include', ids: new Set()});
    setAgentSelectionModel({type: 'include', ids: new Set()});
    onClose();
  };

  const totalSelected =
    userSelectionModel.ids.size +
    groupSelectionModel.ids.size +
    appSelectionModel.ids.size +
    agentSelectionModel.ids.size;

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="md" fullWidth>
      <DialogTitle>{t('roles:assignments.dialog.title')}</DialogTitle>
      <DialogContent>
        <Tabs value={activeTab} onChange={(_e: SyntheticEvent, v: number) => setActiveTab(v)} sx={{mb: 2}}>
          <Tab label={t('roles:assignments.dialog.tabs.users')} />
          <Tab label={t('roles:assignments.dialog.tabs.groups')} />
          <Tab label={t('roles:assignments.dialog.tabs.apps')} />
          <Tab label={t('roles:assignments.dialog.tabs.agents', 'Agents')} />
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

        {activeTab === 2 && (
          <>
            {appsError && !appsLoading && (
              <Alert severity="error" sx={{mb: 2}}>
                {appsError.message ?? t('roles:assignments.dialog.fetchError')}
              </Alert>
            )}
            <Box sx={{height: 400, width: '100%'}}>
              <DataGrid.DataGrid
                rows={filteredApplications}
                columns={appColumns}
                loading={appsLoading}
                getRowId={(row): string => row.id}
                checkboxSelection
                rowSelectionModel={appSelectionModel}
                onRowSelectionModelChange={setAppSelectionModel}
                paginationMode="server"
                rowCount={applicationsData?.totalResults ?? 0}
                paginationModel={appPaginationModel}
                onPaginationModelChange={setAppPaginationModel}
                pageSizeOptions={[5, 10]}
                disableRowSelectionOnClick
                localeText={dataGridLocaleText}
              />
            </Box>
          </>
        )}

        {activeTab === 3 && (
          <>
            {agentsError && !agentsLoading && (
              <Alert severity="error" sx={{mb: 2}}>
                {agentsError.message ?? t('roles:assignments.dialog.fetchError')}
              </Alert>
            )}
            <Box sx={{height: 400, width: '100%'}}>
              <DataGrid.DataGrid
                rows={filteredAgents}
                columns={agentColumns}
                loading={agentsLoading}
                getRowId={(row): string => row.id}
                checkboxSelection
                rowSelectionModel={agentSelectionModel}
                onRowSelectionModelChange={setAgentSelectionModel}
                paginationMode="server"
                rowCount={agentsData?.totalResults ?? 0}
                paginationModel={agentPaginationModel}
                onPaginationModelChange={setAgentPaginationModel}
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
