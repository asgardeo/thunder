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

import {Stack, Button, TextField, InputAdornment, PageContent, PageTitle} from '@wso2/oxygen-ui';
import {Plus, Search} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useTranslation} from 'react-i18next';
import {useNavigate} from 'react-router';
import {useLogger} from '@thunder/logger/react';
import ApplicationsList from '../components/ApplicationsList';

export default function ApplicationsListPage(): JSX.Element {
  const navigate = useNavigate();
  const {t} = useTranslation();
  const logger = useLogger('ApplicationsListPage');

  return (
    <PageContent>
      {/* Header */}
      <PageTitle>
        <PageTitle.Header>{t('applications:listing.title')}</PageTitle.Header>
        <PageTitle.SubHeader>{t('applications:listing.subtitle')}</PageTitle.SubHeader>
        <PageTitle.Actions>
          <Button
            variant="contained"
            startIcon={<Plus size={18} />}
            onClick={() => {
              (async () => {
                await navigate('/applications/create');
              })().catch((error: unknown) => {
                logger.error('Failed to navigate to create application page', {error});
              });
            }}
          >
            {t('applications:listing.addApplication')}
          </Button>
        </PageTitle.Actions>
      </PageTitle>

      {/* Search and Filters */}
      <Stack direction="row" spacing={2} mb={4} flexWrap="wrap" useFlexGap>
        <TextField
          placeholder={t('applications:listing.search.placeholder')}
          size="small"
          sx={{flexGrow: 1, minWidth: 300}}
          slotProps={{
            input: {
              startAdornment: (
                <InputAdornment position="start">
                  <Search size={16} />
                </InputAdornment>
              ),
            },
          }}
        />
      </Stack>
      <ApplicationsList />
    </PageContent>
  );
}
