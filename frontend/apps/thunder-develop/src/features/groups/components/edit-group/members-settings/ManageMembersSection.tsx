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

import {useState, useMemo, type JSX, type ReactNode} from 'react';
import {Box, Avatar, Chip, DataGrid, IconButton, useTheme} from '@wso2/oxygen-ui';
import {User, Users, Trash2} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import SettingsCard from '@/components/SettingsCard';
import useDataGridLocaleText from '../../../../../hooks/useDataGridLocaleText';
import useGetGroupMembers from '../../../api/useGetGroupMembers';
import type {Member} from '../../../models/group';

interface ManageMembersSectionProps {
  groupId: string;
  onRemoveMember: (member: Member) => void;
  headerAction?: ReactNode;
}

/**
 * Section component for displaying and managing group members.
 */
export default function ManageMembersSection({groupId, onRemoveMember, headerAction = undefined}: ManageMembersSectionProps): JSX.Element {
  const {t} = useTranslation();
  const theme = useTheme();
  const dataGridLocaleText = useDataGridLocaleText();
  const [paginationModel, setPaginationModel] = useState<DataGrid.GridPaginationModel>({pageSize: 10, page: 0});

  const membersParams = useMemo(
    () => ({
      limit: paginationModel.pageSize,
      offset: paginationModel.page * paginationModel.pageSize,
    }),
    [paginationModel],
  );
  const {data: membersData, isLoading} = useGetGroupMembers(groupId, membersParams);

  const columns: DataGrid.GridColDef<Member>[] = useMemo(
    () => [
      {
        field: 'avatar',
        headerName: '',
        width: 70,
        sortable: false,
        filterable: false,
        renderCell: (params: DataGrid.GridRenderCellParams<Member>): JSX.Element => (
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
              {params.row.type === 'user' ? <User size={14} /> : <Users size={14} />}
            </Avatar>
          </Box>
        ),
      },
      {
        field: 'id',
        headerName: t('groups:edit.members.sections.manage.listing.columns.id'),
        flex: 1,
        minWidth: 250,
      },
      {
        field: 'type',
        headerName: t('groups:edit.members.sections.manage.listing.columns.type'),
        width: 150,
        renderCell: (params: DataGrid.GridRenderCellParams<Member>): JSX.Element => (
          <Chip label={params.row.type} size="small" variant="outlined" sx={{textTransform: 'capitalize'}} />
        ),
      },
      {
        field: 'actions',
        headerName: '',
        width: 60,
        sortable: false,
        filterable: false,
        renderCell: (params: DataGrid.GridRenderCellParams<Member>): JSX.Element => (
          <IconButton
            size="small"
            color="error"
            aria-label={t('common:actions.remove')}
            onClick={(e) => {
              e.stopPropagation();
              onRemoveMember(params.row);
            }}
          >
            <Trash2 size={14} />
          </IconButton>
        ),
      },
    ],
    [t, theme, onRemoveMember],
  );

  return (
    <SettingsCard
      title={t('groups:edit.members.sections.manage.title')}
      description={t('groups:edit.members.sections.manage.description')}
      headerAction={headerAction}
      slotProps={{
        content: {
          sx: {
            p: 0,
          },
        },
      }}
    >
      <Box sx={{height: 400, width: '100%'}}>
        <DataGrid.DataGrid
          rows={membersData?.members ?? []}
          columns={columns}
          loading={isLoading}
          getRowId={(row): string => row.id}
          paginationMode="server"
          rowCount={membersData?.totalResults ?? 0}
          paginationModel={paginationModel}
          onPaginationModelChange={setPaginationModel}
          pageSizeOptions={[5, 10, 25]}
          disableRowSelectionOnClick
          localeText={dataGridLocaleText}
        />
      </Box>
    </SettingsCard>
  );
}
