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

import {useState, type JSX} from 'react';
import {useNavigate} from 'react-router';
import {Box, Button, Stack, Typography} from '@wso2/oxygen-ui';
import {Plus} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import {useLogger} from '@thunder/logger';
import useDebounce from '@/features/design-system/hooks/useDebounce';
import SearchField from '@/features/design-system/components/SearchField';
import LayoutsList from '../components/LayoutsList';

/**
 * Page component for layouts listing with search functionality
 */
export default function LayoutsListPage(): JSX.Element {
  const {t} = useTranslation();
  const navigate = useNavigate();
  const logger = useLogger('LayoutsListPage');
  const [searchQuery, setSearchQuery] = useState('');

  // Debounce search query for better performance
  const debouncedSearchQuery = useDebounce(searchQuery, 300);

  return (
    <Box>
      <Stack direction="row" justifyContent="space-between" alignItems="center" mb={4}>
        <Box>
          <Typography variant="h1">{t('layouts:listing.title', {defaultValue: 'Layouts'})}</Typography>
          <Typography variant="subtitle1" color="text.secondary">
            {t('layouts:listing.subtitle', {
              defaultValue: 'Create and manage custom layouts for your authentication flows',
            })}
          </Typography>
        </Box>
        <Stack direction="row" spacing={2}>
          <Button
            variant="contained"
            startIcon={<Plus size={18} />}
            onClick={() => {
              (async () => {
                await navigate('/layouts/builder');
              })().catch((error: unknown) => {
                logger.error('Navigation to layout builder failed', {error});
              });
            }}
          >
            {t('layouts:listing.create', {defaultValue: 'Create Layout'})}
          </Button>
        </Stack>
      </Stack>

      <Stack direction="row" spacing={2} mb={4}>
        <SearchField
          value={searchQuery}
          onChange={setSearchQuery}
          placeholder={t('layouts:listing.search.placeholder', {defaultValue: 'Search layouts...'})}
        />
      </Stack>

      <LayoutsList searchQuery={debouncedSearchQuery} />
    </Box>
  );
}
