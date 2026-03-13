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
import SettingsCard from '@/components/SettingsCard';
import type {OrganizationUnit} from '../../../models/organization-unit';

/**
 * Props for the {@link QuickCopySection} component.
 */
interface QuickCopySectionProps {
  /**
   * The organization unit being displayed
   */
  organizationUnit: OrganizationUnit;
  /**
   * The name of the field that was recently copied to clipboard
   */
  copiedField: string | null;
  /**
   * Callback function to copy text to clipboard
   * @param text - The text to copy
   * @param fieldName - The name of the field being copied
   */
  onCopyToClipboard: (text: string, fieldName: string) => Promise<void>;
}

/**
 * Section component for quickly copying organization unit identifiers.
 *
 * Displays read-only text fields with copy buttons for:
 * - Handle (unique identifier)
 * - Organization Unit ID
 *
 * Provides visual feedback when values are copied.
 *
 * @param props - Component props
 * @returns Quick copy UI within a SettingsCard
 */
export default function QuickCopySection({organizationUnit, copiedField, onCopyToClipboard}: QuickCopySectionProps) {
  const {t} = useTranslation();

  return (
    <SettingsCard
      title={t('organizationUnits:edit.general.sections.quickCopy.title')}
      description={t('organizationUnits:edit.general.sections.quickCopy.description')}
    >
      <Stack spacing={3}>
        <FormControl fullWidth>
          <FormLabel htmlFor="handle-input">{t('organizationUnits:edit.general.handle.label')}</FormLabel>
          <TextField
            fullWidth
            id="handle-input"
            value={organizationUnit.handle}
            InputProps={{
              readOnly: true,
              endAdornment: (
                <InputAdornment position="end">
                  <Tooltip title={copiedField === 'handle' ? t('common:actions.copied') : t('common:actions.copy')}>
                    <IconButton
                      aria-label={copiedField === 'handle' ? t('common:actions.copied') : t('common:actions.copy')}
                      onClick={() => {
                        onCopyToClipboard(organizationUnit.handle, 'handle').catch(() => {});
                      }}
                      edge="end"
                    >
                      {copiedField === 'handle' ? <Check size={16} /> : <Copy size={16} />}
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

        <FormControl fullWidth>
          <FormLabel htmlFor="ou-id-input">{t('organizationUnits:edit.general.ou.id.label')}</FormLabel>
          <TextField
            fullWidth
            id="ou-id-input"
            value={organizationUnit.id}
            InputProps={{
              readOnly: true,
              endAdornment: (
                <InputAdornment position="end">
                  <Tooltip title={copiedField === 'ou_id' ? t('common:actions.copied') : t('common:actions.copy')}>
                    <IconButton
                      aria-label={copiedField === 'ou_id' ? t('common:actions.copied') : t('common:actions.copy')}
                      onClick={() => {
                        onCopyToClipboard(organizationUnit.id, 'ou_id').catch(() => {});
                      }}
                      edge="end"
                    >
                      {copiedField === 'ou_id' ? <Check size={16} /> : <Copy size={16} />}
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
