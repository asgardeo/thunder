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

import {Box, Stack, Typography} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';
import type {Application} from '../../models/application';
import type {OAuth2Config} from '../../models/oauth';
import IntegrationGuide from '../../integrate-application/IntegrationGuide';
import getIntegrationGuidesForTemplate from '../../../utils/getIntegrationGuidesForTemplate';

interface IntegrationGuidesProps {
  application: Application;
  oauth2Config?: OAuth2Config;
}

export default function IntegrationGuides({application, oauth2Config = undefined}: IntegrationGuidesProps) {
  const {t} = useTranslation();

  const integrationGuides = getIntegrationGuidesForTemplate(application.template);

  return (
    <Stack spacing={4}>
      {integrationGuides ? (
        <IntegrationGuide
          appName={application.name}
          appLogo={application.logo_url ?? null}
          selectedColor="#1976d2"
          clientId={oauth2Config?.client_id ?? ''}
          clientSecret={oauth2Config?.client_secret ?? ''}
          hasOAuthConfig={!!oauth2Config}
          applicationId={application.id}
          integrationGuides={integrationGuides}
          templateId={application.template ?? null}
        />
      ) : (
        <Box>
          <Typography variant="body1" color="text.secondary">
            {t('applications:edit.overview.noGuides')}
          </Typography>
        </Box>
      )}
    </Stack>
  );
}
