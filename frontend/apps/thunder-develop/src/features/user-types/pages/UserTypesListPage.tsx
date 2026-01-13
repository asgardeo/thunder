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

import {useNavigate} from 'react-router';
import {Box, Stack, Typography, Button} from '@wso2/oxygen-ui';
import {Plus} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import {useLogger} from '@thunder/logger/react';
import UserTypesList from '../components/UserTypesList';

export default function UserTypesListPage() {
  const navigate = useNavigate();
  const {t} = useTranslation();
  const logger = useLogger('UserTypesListPage');

  return (
    <Box>
      <Stack direction="row" justifyContent="space-between" alignItems="center" mb={4} flexWrap="wrap" gap={2}>
        <Box>
          <Typography variant="h1" gutterBottom>
            {t('userTypes:title')}
          </Typography>
          <Typography variant="subtitle1" color="text.secondary">
            {t('userTypes:createDescription')}
          </Typography>
        </Box>
        <Stack direction="row" spacing={2}>
          <Button
            variant="contained"
            startIcon={<Plus size={18} />}
            onClick={() => {
              const handler = async () => {
                await navigate('/user-types/create');
              };

              handler().catch((error: unknown) => {
                logger.error('Failed to navigate to create user type page', {error});
              });
            }}
          >
            {t('userTypes:createUserType')}
          </Button>
        </Stack>
      </Stack>
      <UserTypesList />
    </Box>
  );
}
