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

import {Box, Stack, Typography, Chip} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';
import type {OAuth2Config} from '../../../models/oauth';
import SettingsCard from '../../../../../components/SettingsCard';

/**
 * Props for the {@link OAuth2ConfigSection} component.
 */
interface OAuth2ConfigSectionProps {
  /**
   * OAuth2 configuration to display (optional)
   */
  oauth2Config?: OAuth2Config;
}

/**
 * Section component for displaying OAuth2 configuration (read-only).
 *
 * Shows:
 * - Allowed grant types (authorization_code, refresh_token, etc.)
 * - Allowed response types (code)
 * - Public client status
 * - PKCE requirement status
 *
 * Returns null if no OAuth2 configuration is provided.
 *
 * @param props - Component props
 * @returns OAuth2 configuration display UI within a SettingsCard, or null
 */
export default function OAuth2ConfigSection({oauth2Config = undefined}: OAuth2ConfigSectionProps) {
  const {t} = useTranslation();

  if (!oauth2Config) return null;

  return (
    <SettingsCard
      title={t('applications:edit.advanced.labels.oauth2Config')}
      description={t('applications:edit.advanced.oauth2Config.intro')}
    >
      <Stack spacing={2}>
        <Box>
          <Typography variant="subtitle2" color="text.secondary" gutterBottom>
            {t('applications:edit.advanced.labels.grantTypes')}
          </Typography>
          <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
            {oauth2Config.grant_types?.map((grant) => (
              <Chip key={grant} label={grant} size="small" variant="outlined" />
            ))}
          </Stack>
          <Typography variant="caption" color="text.secondary" sx={{mt: 1, display: 'block'}}>
            {t('applications:edit.advanced.grantTypes.hint')}
          </Typography>
        </Box>

        <Box>
          <Typography variant="subtitle2" color="text.secondary" gutterBottom>
            {t('applications:edit.advanced.labels.responseTypes')}
          </Typography>
          <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
            {oauth2Config.response_types?.map((response) => (
              <Chip key={response} label={response} size="small" variant="outlined" />
            ))}
          </Stack>
        </Box>

        <Box>
          <Typography variant="subtitle2" color="text.secondary">
            {t('applications:edit.advanced.labels.publicClient')}
          </Typography>
          <Typography variant="body1">
            {oauth2Config.public_client
              ? t('applications:edit.advanced.publicClient.yes')
              : t('applications:edit.advanced.publicClient.no')}
          </Typography>
          <Typography variant="caption" color="text.secondary" sx={{mt: 0.5, display: 'block'}}>
            {oauth2Config.public_client
              ? t('applications:edit.advanced.publicClient.public')
              : t('applications:edit.advanced.publicClient.confidential')}
          </Typography>
        </Box>

        <Box>
          <Typography variant="subtitle2" color="text.secondary">
            {t('applications:edit.advanced.labels.pkceRequired')}
          </Typography>
          <Typography variant="body1">
            {oauth2Config.pkce_required
              ? t('applications:edit.advanced.pkce.yes')
              : t('applications:edit.advanced.pkce.no')}
          </Typography>
          <Typography variant="caption" color="text.secondary" sx={{mt: 0.5, display: 'block'}}>
            {oauth2Config.pkce_required
              ? t('applications:edit.advanced.pkce.enabled')
              : t('applications:edit.advanced.pkce.disabled')}
          </Typography>
        </Box>
      </Stack>
    </SettingsCard>
  );
}
