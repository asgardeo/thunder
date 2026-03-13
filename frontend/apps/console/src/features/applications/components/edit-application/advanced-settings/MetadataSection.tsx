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
import type {Application} from '../../../models/application';
import SettingsCard from '../../../../../components/SettingsCard';

/**
 * Props for the {@link MetadataSection} component.
 */
interface MetadataSectionProps {
  /**
   * The application to display metadata for
   */
  application: Application;
}

/**
 * Section component for displaying application metadata (read-only).
 *
 * Shows:
 * - Created at timestamp (formatted as locale string)
 * - Updated at timestamp (formatted as locale string)
 *
 * Returns null if no metadata timestamps are available.
 *
 * @param props - Component props
 * @returns Metadata display UI within a SettingsCard, or null
 */
export default function MetadataSection({application}: MetadataSectionProps) {
  const {t} = useTranslation();

  if (!application.created_at && !application.updated_at) {
    return null;
  }

  return (
    <SettingsCard title={t('applications:edit.advanced.labels.metadata')}>
      <Stack spacing={2}>
        {application.created_at && (
          <Box>
            <Typography variant="subtitle2" color="text.secondary">
              {t('applications:edit.advanced.labels.createdAt')}
            </Typography>
            <Typography variant="body1">{new Date(application.created_at).toLocaleString()}</Typography>
          </Box>
        )}
        {application.updated_at && (
          <Box>
            <Typography variant="subtitle2" color="text.secondary">
              {t('applications:edit.advanced.labels.updatedAt')}
            </Typography>
            <Typography variant="body1">{new Date(application.updated_at).toLocaleString()}</Typography>
          </Box>
        )}
      </Stack>
    </SettingsCard>
  );
}
