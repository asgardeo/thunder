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

import {useMemo, type JSX} from 'react';
import {Box, DataGrid, Avatar} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';
import SettingsCard from '@/components/SettingsCard';
import useDataGridLocaleText from '../../../../../hooks/useDataGridLocaleText';
import useGetOrganizationUnitUsers from '../../../api/useGetOrganizationUnitUsers';
import type {ApiUser} from '../../../../users/types/users';

/**
 * Props for the {@link ManageUsersSection} component.
 */
interface ManageUsersSectionProps {
  /**
   * The ID of the organization unit
   */
  organizationUnitId: string;
}

const getInitials = (name?: string) => {
  if (!name) return '?';
  return name
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);
};

/**
 * Section component for managing users belonging to an organization unit.
 *
 * Displays a DataGrid of users with:
 * - Avatar with initials
 * - Display Name (falls back to User ID)
 * - User ID
 * - User Type
 *
 * @param props - Component props
 * @returns Manage users section within a SettingsCard
 */
export default function ManageUsersSection({organizationUnitId}: ManageUsersSectionProps): JSX.Element {
  const {t} = useTranslation();
  const dataGridLocaleText = useDataGridLocaleText();

  const {data: usersData, isLoading} = useGetOrganizationUnitUsers(organizationUnitId);

  const columns: DataGrid.GridColDef<ApiUser>[] = useMemo(
    () => [
      {
        field: 'avatar',
        headerName: '',
        width: 70,
        sortable: false,
        filterable: false,
        renderCell: (params: DataGrid.GridRenderCellParams<ApiUser>): JSX.Element => {
          const displayVal = params.row.display ?? params.row.id;

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
                {getInitials(displayVal)}
              </Avatar>
            </Box>
          );
        },
      },
      {
        field: 'display',
        headerName: t('organizationUnits:edit.users.sections.manage.listing.columns.name'),
        flex: 1,
        minWidth: 200,
        valueGetter: (_value: unknown, row: ApiUser) => row.display ?? row.id,
      },
      {
        field: 'id',
        headerName: t('organizationUnits:edit.users.sections.manage.listing.columns.id'),
        flex: 1,
        minWidth: 250,
      },
      {
        field: 'type',
        headerName: t('organizationUnits:edit.users.sections.manage.listing.columns.type'),
        flex: 0.6,
        minWidth: 120,
      },
    ],
    [t],
  );

  return (
    <SettingsCard
      title={t('organizationUnits:edit.users.sections.manage.title')}
      description={t('organizationUnits:edit.users.sections.manage.description')}
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
          rows={usersData?.users ?? []}
          columns={columns}
          loading={isLoading}
          getRowId={(row): string => row.id}
          initialState={{
            pagination: {
              paginationModel: {pageSize: 10},
            },
          }}
          pageSizeOptions={[5, 10, 25]}
          disableRowSelectionOnClick
          localeText={dataGridLocaleText}
        />
      </Box>
    </SettingsCard>
  );
}
