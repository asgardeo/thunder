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

import {Box, Typography, TextField, Autocomplete, CircularProgress} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';
import {useGetBrandings} from '@thunder/shared-branding';
import type {Application} from '../../../models/application';
import SettingsCard from '../SettingsCard';

interface AppearanceSectionProps {
  application: Application;
  editedApp: Partial<Application>;
  onFieldChange: (field: keyof Application, value: unknown) => void;
}

export default function AppearanceSection({application, editedApp, onFieldChange}: AppearanceSectionProps) {
  const {t} = useTranslation();
  const {data: brandingsData, isLoading: loadingBrandings} = useGetBrandings();

  const themeOptions = brandingsData?.brandings ?? [];

  return (
    <SettingsCard
      title={t('applications:edit.customization.sections.appearance')}
      description={t('applications:edit.customization.sections.appearance.description')}
    >
      <Box>
        <Typography variant="subtitle2" gutterBottom>
          {t('applications:edit.customization.labels.theme')}
        </Typography>
        <Autocomplete
          fullWidth
          options={themeOptions}
          getOptionLabel={(option) => (typeof option === 'string' ? option : option.displayName)}
          value={
            themeOptions.find(
              (theme) =>
                theme.id ===
                (((editedApp as Record<string, unknown>).branding_id as string) ??
                  ((application as Record<string, unknown>).branding_id as string)),
            ) ?? null
          }
          onChange={(_event, newValue) => onFieldChange('branding_id' as keyof Application, newValue?.id ?? '')}
          loading={loadingBrandings}
          renderInput={(params) => (
            <TextField
              {...params}
              placeholder={t('applications:edit.customization.theme.placeholder')}
              helperText={t('applications:edit.customization.theme.hint')}
              InputProps={{
                ...params.InputProps,
                endAdornment: (
                  <>
                    {loadingBrandings ? <CircularProgress color="inherit" size={20} /> : null}
                    {params.InputProps.endAdornment}
                  </>
                ),
              }}
            />
          )}
        />
      </Box>
    </SettingsCard>
  );
}
