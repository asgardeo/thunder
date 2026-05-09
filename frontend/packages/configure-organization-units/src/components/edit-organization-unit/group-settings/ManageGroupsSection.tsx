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

import {SettingsCard, getInitials} from '@thunderid/components';
import {useDataGridLocaleText} from '@thunderid/hooks';
import {Box, DataGrid, Avatar} from '@wso2/oxygen-ui';
import {useMemo, type JSX} from 'react';
import {useTranslation} from 'react-i18next';
import useGetOrganizationUnitGroups from '../../../api/useGetOrganizationUnitGroups';
import type {Group} from '../../../models/group';

/**
 * Props for the {@link ManageGroupsSection} component.
 */
interface ManageGroupsSectionProps {
  /**
   * The ID of the organization unit
   */
  organizationUnitId: string;
}

/**
 * Section component for managing groups belonging to an organization unit.
 *
 * Displays a DataGrid of groups with:
 * - Avatar icon
 * - Group Name
 * - Group ID
 *
 * @param props - Component props
 * @returns Manage groups section within a SettingsCard
 */
export default function ManageGroupsSection({organizationUnitId}: ManageGroupsSectionProps): JSX.Element {
  const {t} = useTranslation();
  const dataGridLocaleText = useDataGridLocaleText();

  const {data: groupsData, isLoading} = useGetOrganizationUnitGroups(organizationUnitId);

  const columns: DataGrid.GridColDef<Group>[] = useMemo(
    () => [
      {
        field: 'avatar',
        headerName: '',
        width: 70,
        sortable: false,
        filterable: false,
        renderCell: (params: DataGrid.GridRenderCellParams<Group>): JSX.Element => {
          const displayVal = params.row.name ?? params.row.id;

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
        field: 'name',
        headerName: t('organizationUnits:edit.groups.sections.manage.listing.columns.name'),
        flex: 1,
        minWidth: 200,
      },
      {
        field: 'id',
        headerName: t('organizationUnits:edit.groups.sections.manage.listing.columns.id'),
        flex: 1,
        minWidth: 250,
      },
    ],
    [t],
  );

  return (
    <SettingsCard
      title={t('organizationUnits:edit.groups.sections.manage.title')}
      description={t('organizationUnits:edit.groups.sections.manage.description')}
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
          rows={groupsData?.groups ?? []}
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
