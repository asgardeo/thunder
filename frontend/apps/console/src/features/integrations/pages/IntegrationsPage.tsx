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

import {Box, Typography, Paper, PageContent, PageTitle} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';

export default function IntegrationsPage() {
  const {t} = useTranslation();

  return (
    <PageContent>
      {/* Header */}
      <PageTitle>
        <PageTitle.Header>{t('integrations:title')}</PageTitle.Header>
        <PageTitle.SubHeader>{t('integrations:subtitle')}</PageTitle.SubHeader>
      </PageTitle>

      <Paper
        sx={{
          p: 8,
          textAlign: 'center',
          minHeight: 400,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <Box>
          <Typography variant="h4" component="h4" gutterBottom color="text.secondary" fontWeight={500}>
            {t('integrations:comingSoon')}
          </Typography>
          <Typography variant="body1" color="text.secondary">
            {t('integrations:comingSoonDescription')}
          </Typography>
        </Box>
      </Paper>
    </PageContent>
  );
}
