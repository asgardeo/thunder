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

import {Stack, TextField, InputAdornment, Tooltip, IconButton, FormControl, FormLabel} from '@wso2/oxygen-ui';
import {Copy, Check} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import type {Application} from '../../../models/application';
import type {OAuth2Config} from '../../../models/oauth';
import SettingsCard from '../SettingsCard';

interface QuickCopySectionProps {
  application: Application;
  oauth2Config?: OAuth2Config;
  copiedField: string | null;
  onCopyToClipboard: (text: string, fieldName: string) => Promise<void>;
}

export default function QuickCopySection({
  application,
  oauth2Config = undefined,
  copiedField,
  onCopyToClipboard,
}: QuickCopySectionProps) {
  const {t} = useTranslation();

  return (
    <SettingsCard
      title={t('applications:edit.general.sections.quickCopy')}
      description={t('applications:edit.general.sections.quickCopy.description')}
    >
      <Stack spacing={3}>
        <FormControl fullWidth>
          <FormLabel htmlFor="application-id-input">{t('applications:edit.general.labels.applicationId')}</FormLabel>
          <TextField
            fullWidth
            id="application-id-input"
            value={application.id}
            InputProps={{
              readOnly: true,
              endAdornment: (
                <InputAdornment position="end">
                  <Tooltip title={copiedField === 'app_id' ? t('common:actions.copied') : t('common:actions.copy')}>
                    <IconButton
                      onClick={() => {
                        onCopyToClipboard(application.id, 'app_id').catch(() => {});
                      }}
                      edge="end"
                    >
                      {copiedField === 'app_id' ? <Check size={16} /> : <Copy size={16} />}
                    </IconButton>
                  </Tooltip>
                </InputAdornment>
              ),
            }}
            helperText={t('applications:edit.general.applicationId.hint')}
            sx={{
              '& input': {
                fontFamily: 'monospace',
                fontSize: '0.875rem',
              },
            }}
          />
        </FormControl>

        <FormControl fullWidth>
          <FormLabel htmlFor="client-id-input">{t('applications:edit.general.labels.clientId')}</FormLabel>
          <TextField
            fullWidth
            id="client-id-input"
            value={oauth2Config?.client_id ?? ''}
            InputProps={{
              readOnly: true,
              endAdornment: (
                <InputAdornment position="end">
                  <Tooltip title={copiedField === 'client_id' ? t('common:actions.copied') : t('common:actions.copy')}>
                    <IconButton
                      onClick={() => {
                        if (oauth2Config?.client_id) {
                          onCopyToClipboard(oauth2Config.client_id, 'client_id').catch(() => {});
                        }
                      }}
                      edge="end"
                    >
                      {copiedField === 'client_id' ? <Check size={16} /> : <Copy size={16} />}
                    </IconButton>
                  </Tooltip>
                </InputAdornment>
              ),
            }}
            helperText={t('applications:edit.general.clientId.hint')}
            sx={{
              '& input': {
                fontFamily: 'monospace',
                fontSize: '0.875rem',
              },
            }}
          />
        </FormControl>
      </Stack>
    </SettingsCard>
  );
}
