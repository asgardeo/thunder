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

import {SettingsCard} from '@thunderid/components';
import {Stack, TextField, InputAdornment, Tooltip, IconButton, FormControl, FormLabel} from '@wso2/oxygen-ui';
import {Copy, Check} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import type {ApiUserType} from '../../../types/user-types';

interface QuickCopySectionProps {
  userType: ApiUserType;
  copiedField: string | null;
  onCopyToClipboard: (text: string, fieldName: string) => Promise<void>;
}

export default function QuickCopySection({userType, copiedField, onCopyToClipboard}: QuickCopySectionProps) {
  const {t} = useTranslation();

  return (
    <SettingsCard
      title={t('userTypes:edit.general.sections.quickCopy.title', 'Quick Copy')}
      description={t(
        'userTypes:edit.general.sections.quickCopy.description',
        'Copy user type identifiers for use in your application.',
      )}
    >
      <Stack spacing={3}>
        <FormControl fullWidth>
          <FormLabel htmlFor="user-type-id-input">
            {t('userTypes:edit.general.labels.userTypeId', 'User Type ID')}
          </FormLabel>
          <TextField
            fullWidth
            id="user-type-id-input"
            value={userType.id}
            InputProps={{
              readOnly: true,
              endAdornment: (
                <InputAdornment position="end">
                  <Tooltip
                    title={
                      copiedField === 'user_type_id'
                        ? t('common:actions.copied', 'Copied')
                        : t('userTypes:edit.copyId', 'Copy user type ID')
                    }
                  >
                    <IconButton
                      aria-label={
                        copiedField === 'user_type_id'
                          ? t('common:actions.copied', 'Copied')
                          : t('userTypes:edit.copyId', 'Copy user type ID')
                      }
                      onClick={() => {
                        onCopyToClipboard(userType.id, 'user_type_id').catch(() => null);
                      }}
                      edge="end"
                    >
                      {copiedField === 'user_type_id' ? <Check size={16} /> : <Copy size={16} />}
                    </IconButton>
                  </Tooltip>
                </InputAdornment>
              ),
            }}
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
