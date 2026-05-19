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
import {AppWindow, Bot, User as UserIcon} from '@wso2/oxygen-ui-icons-react';
import {useState, useMemo, useCallback, type JSX, type SyntheticEvent} from 'react';
import {useTranslation} from 'react-i18next';
import useGetAgents from '../../../../agents/api/useGetAgents';
import type {BasicAgent} from '../../../../agents/models/agent';
import useGetApplications from '../../../../applications/api/useGetApplications';
import type {BasicApplication} from '../../../../applications/models/application';
import type {Member} from '../../../models/group';

interface AddMemberDialogProps {
  open: boolean;
  onClose: () => void;
  onAdd: (members: Member[]) => void;
}

/**
 * Dialog for searching and adding user or app members to a group.
 */
export default function AddMemberDialog({open, onClose, onAdd}: AddMemberDialogProps): JSX.Element {
  const {t} = useTranslation();
  const theme = useTheme();
  const dataGridLocaleText = useDataGridLocaleText();

  const [activeTab, setActiveTab] = useState(0);
  const [userSelectionModel, setUserSelectionModel] = useState<DataGrid.GridRowSelectionModel>({
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
  const [appPaginationModel, setAppPaginationModel] = useState<DataGrid.GridPaginationModel>({pageSize: 10, page: 0});
  const [agentPaginationModel, setAgentPaginationModel] = useState<DataGrid.GridPaginationModel>({
    pageSize: 10,
    page: 0,
  });

  const usersParams = useMemo(
    () => ({
      limit: userPaginationModel.pageSize,
      offset: userPaginationModel.page * userPaginationModel.pageSize,
    }),
    [userPaginationModel],
  );
  const appsParams = useMemo(
    () => ({
      limit: appPaginationModel.pageSize,
      offset: appPaginationModel.page * appPaginationModel.pageSize,
    }),
    [appPaginationModel],
  );
  const agentsParams = useMemo(
    () => ({
      limit: agentPaginationModel.pageSize,
      offset: agentPaginationModel.page * agentPaginationModel.pageSize,
    }),
    [agentPaginationModel],
  );
  const {data: usersData, isLoading: usersLoading, error: usersError} = useGetUsers(usersParams);
  const {data: appsData, isLoading: appsLoading, error: appsError} = useGetApplications(appsParams);
  const {data: agentsData, isLoading: agentsLoading, error: agentsError} = useGetAgents(agentsParams);

  const users: User[] = useMemo(() => usersData?.users ?? [], [usersData]);
  const applications: BasicApplication[] = useMemo(() => appsData?.applications ?? [], [appsData]);
  const agents: BasicAgent[] = useMemo(() => agentsData?.agents ?? [], [agentsData]);

  const userColumns: DataGrid.GridColDef<User>[] = useMemo(
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
              <UserIcon size={14} />
            </Avatar>
          </Box>
        ),
      },
      {
        field: 'display',
        headerName: t('groups:addMember.columns.displayName'),
        flex: 1,
        minWidth: 200,
        renderCell: (params: DataGrid.GridRenderCellParams<User>): JSX.Element => (
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
              {params.row.display ?? params.row.id}
            </Typography>
            <Typography
              variant="caption"
              color="text.secondary"
              noWrap
              sx={{fontFamily: 'monospace', fontSize: '0.7rem'}}
            >
              {params.row.id}
            </Typography>
          </Box>
        ),
      },
      {
        field: 'type',
        headerName: t('groups:addMember.columns.userType'),
        width: 150,
        renderCell: (params: DataGrid.GridRenderCellParams<User>): JSX.Element => (
          <Chip label={params.row.type} size="small" variant="outlined" sx={{textTransform: 'capitalize'}} />
        ),
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
              <Bot size={14} />
            </Avatar>
          </Box>
        ),
      },
      {
        field: 'name',
        headerName: t('groups:addMember.columns.displayName'),
        flex: 1,
        minWidth: 200,
      },
      {
        field: 'id',
        headerName: t('groups:edit.members.sections.manage.listing.columns.id'),
        flex: 1,
        minWidth: 250,
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
              <AppWindow size={14} />
            </Avatar>
          </Box>
        ),
      },
      {
        field: 'name',
        headerName: t('groups:addMember.columns.displayName'),
        flex: 1,
        minWidth: 200,
      },
      {
        field: 'id',
        headerName: t('groups:edit.members.sections.manage.listing.columns.id'),
        flex: 1,
        minWidth: 250,
      },
    ],
    [theme, t],
  );

  const handleAdd = useCallback(() => {
    const newMembers: Member[] = [
      ...[...userSelectionModel.ids].map((id) => ({id: String(id), type: 'user' as const})),
      ...[...appSelectionModel.ids].map((id) => ({id: String(id), type: 'app' as const})),
      ...[...agentSelectionModel.ids].map((id) => ({id: String(id), type: 'agent' as const})),
    ];
    onAdd(newMembers);
    setUserSelectionModel({type: 'include', ids: new Set()});
    setAppSelectionModel({type: 'include', ids: new Set()});
    setAgentSelectionModel({type: 'include', ids: new Set()});
  }, [userSelectionModel, appSelectionModel, agentSelectionModel, onAdd]);

  const handleClose = (): void => {
    setUserSelectionModel({type: 'include', ids: new Set()});
    setAppSelectionModel({type: 'include', ids: new Set()});
    setAgentSelectionModel({type: 'include', ids: new Set()});
    onClose();
  };

  const totalSelected = userSelectionModel.ids.size + appSelectionModel.ids.size + agentSelectionModel.ids.size;

  const handleTabChange = (_event: SyntheticEvent, tab: number): void => {
    setActiveTab(tab);
  };

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="md" fullWidth>
      <DialogTitle>{t('groups:addMember.title')}</DialogTitle>
      <DialogContent>
        <Tabs value={activeTab} onChange={handleTabChange} sx={{mb: 2}}>
          <Tab icon={<UserIcon size={16} />} iconPosition="start" label={t('groups:addMember.tabs.users')} />
          <Tab icon={<AppWindow size={16} />} iconPosition="start" label={t('groups:addMember.tabs.apps')} />
          <Tab icon={<Bot size={16} />} iconPosition="start" label={t('groups:addMember.tabs.agents', 'Agents')} />
        </Tabs>

        {activeTab === 0 && (
          <>
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
                columns={userColumns}
                loading={usersLoading}
                getRowId={(row): string => row.id}
                checkboxSelection
                rowSelectionModel={userSelectionModel}
                onRowSelectionModelChange={(newSelection) => {
                  setUserSelectionModel(newSelection);
                }}
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
            {appsError && !appsLoading && (
              <Alert severity="error" sx={{mb: 2}}>
                {appsError.message ?? t('groups:addMember.fetchAppsError')}
              </Alert>
            )}
            {!appsError && applications.length === 0 && !appsLoading && (
              <Alert severity="info" sx={{mb: 2}}>
                {t('groups:addMember.noResultsApps')}
              </Alert>
            )}

            <Box sx={{height: 400, width: '100%'}}>
              <DataGrid.DataGrid
                rows={applications}
                columns={appColumns}
                loading={appsLoading}
                getRowId={(row): string => row.id}
                checkboxSelection
                rowSelectionModel={appSelectionModel}
                onRowSelectionModelChange={(newSelection) => {
                  setAppSelectionModel(newSelection);
                }}
                paginationMode="server"
                rowCount={appsData?.totalResults ?? 0}
                paginationModel={appPaginationModel}
                onPaginationModelChange={setAppPaginationModel}
                pageSizeOptions={[5, 10]}
                disableRowSelectionOnClick
                localeText={dataGridLocaleText}
              />
            </Box>
          </>
        )}

        {activeTab === 2 && (
          <>
            {agentsError && !agentsLoading && (
              <Alert severity="error" sx={{mb: 2}}>
                {agentsError.message ?? t('groups:addMember.fetchAgentsError', 'Failed to load agents')}
              </Alert>
            )}
            {!agentsError && agents.length === 0 && !agentsLoading && (
              <Alert severity="info" sx={{mb: 2}}>
                {t('groups:addMember.noResultsAgents', 'No agents found')}
              </Alert>
            )}

            <Box sx={{height: 400, width: '100%'}}>
              <DataGrid.DataGrid
                rows={agents}
                columns={agentColumns}
                loading={agentsLoading}
                getRowId={(row): string => row.id}
                checkboxSelection
                rowSelectionModel={agentSelectionModel}
                onRowSelectionModelChange={(newSelection) => {
                  setAgentSelectionModel(newSelection);
                }}
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
          {t('groups:addMember.add')}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
