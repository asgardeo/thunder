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

import {Box, Typography, Stack, Button, TextField, InputAdornment} from '@wso2/oxygen-ui';
import {Plus, Search} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useTranslation} from 'react-i18next';
import {useNavigate} from 'react-router';
import {useLogger} from '@thunder/logger/react';
import GroupsList from '../components/GroupsList';

export default function GroupsListPage(): JSX.Element {
  const navigate = useNavigate();
  const {t} = useTranslation();
  const logger = useLogger('GroupsListPage');

  return (
    <Box>
      <Stack direction="row" justifyContent="space-between" alignItems="center" mb={4} flexWrap="wrap" gap={2}>
        <Box>
          <Typography variant="h1" gutterBottom>
            {t('groups:listing.title')}
          </Typography>
          <Typography variant="subtitle1" color="text.secondary">
            {t('groups:listing.subtitle')}
          </Typography>
        </Box>
        <Stack direction="row" spacing={2}>
          <Button
            variant="contained"
            startIcon={<Plus size={18} />}
            onClick={() => {
              (async () => {
                await navigate('/groups/create');
              })().catch((error: unknown) => {
                logger.error('Failed to navigate to create group page', {error});
              });
            }}
          >
            {t('groups:listing.addGroup')}
          </Button>
        </Stack>
      </Stack>

      {/* TODO: Connect search field to state and implement server-side filtering. */}
      <Stack direction="row" spacing={2} mb={4} flexWrap="wrap" useFlexGap>
        <TextField
          placeholder={t('groups:listing.search.placeholder')}
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
      <GroupsList />
    </Box>
  );
}
