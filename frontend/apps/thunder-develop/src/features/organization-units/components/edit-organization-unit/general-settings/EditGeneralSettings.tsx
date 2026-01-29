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

import type {JSX} from 'react';
import {Link} from 'react-router';
import {Box, Stack, Typography, Paper, CircularProgress} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';
import type {OrganizationUnit, OUNavigationState} from '../../../types/organization-units';
import useGetOrganizationUnit from '../../../api/useGetOrganizationUnit';

/**
 * Props for the {@link EditGeneralSettings} component.
 */
interface EditGeneralSettingsProps {
  /**
   * The organization unit being displayed
   */
  organizationUnit: OrganizationUnit;
}

/**
 * General settings tab content for the Organization Unit edit page.
 *
 * Displays basic information about the organization unit including:
 * - Handle (unique identifier)
 * - Organization Unit ID
 * - Parent Organization Unit (if any)
 *
 * @param props - Component props
 * @returns General settings tab content
 */
export default function EditGeneralSettings({organizationUnit}: EditGeneralSettingsProps): JSX.Element {
  const {t} = useTranslation();

  const {data: parentOU, isLoading: isLoadingParent} = useGetOrganizationUnit(
    organizationUnit.parent ?? undefined,
    Boolean(organizationUnit.parent),
  );

  const renderParentInfo = (): JSX.Element => {
    if (!organizationUnit.parent) {
      return (
        <Typography variant="body2" color="text.secondary">
          {t('organizationUnits:view.general.noParent')}
        </Typography>
      );
    }

    if (isLoadingParent) {
      return <CircularProgress size={16} />;
    }

    if (parentOU) {
      const navigationState: OUNavigationState = {
        fromOU: {
          id: organizationUnit.id,
          name: organizationUnit.name,
        },
      };

      return (
        <Stack direction="row" spacing={1} alignItems="center">
          <Typography
            component={Link}
            to={`/organization-units/${parentOU.id}`}
            state={navigationState}
            variant="body2"
            sx={{
              color: 'primary.main',
              textDecoration: 'none',
              '&:hover': {
                textDecoration: 'underline',
              },
            }}
          >
            {parentOU.name}
          </Typography>
          <Typography variant="body2" color="text.secondary">
            ({parentOU.id})
          </Typography>
        </Stack>
      );
    }

    return (
      <Typography variant="body2" color="text.secondary">
        {organizationUnit.parent}
      </Typography>
    );
  };

  return (
    <Paper sx={{p: 3, mb: 3}}>
      <Typography variant="h6" gutterBottom>
        {t('organizationUnits:view.general.title')}
      </Typography>
      <Typography variant="body2" color="text.secondary" sx={{mb: 3}}>
        {t('organizationUnits:view.general.subtitle')}
      </Typography>

      <Stack spacing={3}>
        <Box>
          <Typography variant="subtitle2" gutterBottom>
            {t('organizationUnits:form.handle')}
          </Typography>
          <Typography variant="body2" color="text.secondary">
            {organizationUnit.handle}
          </Typography>
        </Box>

        <Box>
          <Typography variant="subtitle2" gutterBottom>
            {t('organizationUnits:view.general.id')}
          </Typography>
          <Typography variant="body2" color="text.secondary">
            {organizationUnit.id}
          </Typography>
        </Box>

        <Box>
          <Typography variant="subtitle2" gutterBottom>
            {t('organizationUnits:view.general.parent')}
          </Typography>
          {renderParentInfo()}
        </Box>
      </Stack>
    </Paper>
  );
}
