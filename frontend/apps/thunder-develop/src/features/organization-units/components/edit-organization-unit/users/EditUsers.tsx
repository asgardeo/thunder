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
import {Box, Typography, Paper, DataGrid, Avatar, useTheme} from '@wso2/oxygen-ui';
import {User} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import useDataGridLocaleText from '../../../../../hooks/useDataGridLocaleText';
import useGetOrganizationUnitUsers from '../../../api/useGetOrganizationUnitUsers';
import type {ApiUser} from '../../../../users/types/users';

/**
 * Props for the {@link EditUsers} component.
 */
interface EditUsersProps {
  /**
   * The ID of the organization unit
   */
  organizationUnitId: string;
}

/**
 * Users tab content for the Organization Unit edit page.
 *
 * Displays a DataGrid of users belonging to the organization unit with:
 * - Avatar icon
 * - User ID
 * - User Type
 *
 * @param props - Component props
 * @returns Users tab content
 */
export default function EditUsers({organizationUnitId}: EditUsersProps): JSX.Element {
  const {t} = useTranslation();
  const theme = useTheme();
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
        headerName: t('organizationUnits:view.users.columns.id'),
        flex: 1,
        minWidth: 200,
      },
      {
        field: 'type',
        headerName: t('organizationUnits:view.users.columns.type'),
        flex: 1,
        minWidth: 150,
      },
    ],
    [t, theme],
  );

  return (
    <Paper sx={{p: 3, mb: 3}}>
      <Typography variant="h6" gutterBottom>
        {t('organizationUnits:view.users.title')}
      </Typography>
      <Typography variant="body2" color="text.secondary" sx={{mb: 3}}>
        {t('organizationUnits:view.users.subtitle')}
      </Typography>

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
    </Paper>
  );
}
