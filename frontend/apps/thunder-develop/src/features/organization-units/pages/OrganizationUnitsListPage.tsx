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
import {useNavigate} from 'react-router';
import {Box, Stack, Typography, Button} from '@wso2/oxygen-ui';
import {Plus} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import {useLogger} from '@thunder/logger/react';
import OrganizationUnitsList from '../components/OrganizationUnitsList';

export default function OrganizationUnitsListPage(): JSX.Element {
  const navigate = useNavigate();
  const {t} = useTranslation();
  const logger = useLogger('OrganizationUnitsListPage');

  return (
    <Box>
      <Stack direction="row" justifyContent="space-between" alignItems="center" mb={4} flexWrap="wrap" gap={2}>
        <Box>
          <Typography variant="h1" gutterBottom>
            {t('organizationUnits:listing.title')}
          </Typography>
          <Typography variant="subtitle1" color="text.secondary">
            {t('organizationUnits:listing.subtitle')}
          </Typography>
        </Box>
        <Stack direction="row" spacing={2}>
          <Button
            variant="contained"
            startIcon={<Plus size={18} />}
            onClick={() => {
              const handler = async () => {
                await navigate('/organization-units/create');
              };

              handler().catch((error: unknown) => {
                logger.error('Failed to navigate to create organization unit page', {error});
              });
            }}
          >
            {t('organizationUnits:listing.addOrganizationUnit')}
          </Button>
        </Stack>
      </Stack>

      <OrganizationUnitsList />
    </Box>
  );
}
