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
import {useGetThemes} from '@thunder/shared-design';
import type {Application} from '../../../models/application';
import SettingsCard from '../../../../../components/SettingsCard';

/**
 * Props for the {@link AppearanceSection} component.
 */
interface AppearanceSectionProps {
  /**
   * The application being edited
   */
  application: Application;
  /**
   * Partial application object containing edited fields
   */
  editedApp: Partial<Application>;
  /**
   * Callback function to handle field value changes
   * @param field - The application field being updated
   * @param value - The new value for the field
   */
  onFieldChange: (field: keyof Application, value: unknown) => void;
}

/**
 * Section component for configuring application appearance.
 *
 * Provides an autocomplete dropdown to select a theme or layout from available options.
 * The selected theme or layout affects the look and feel of the application's login pages.
 *
 * @param props - Component props
 * @returns Appearance configuration UI within a SettingsCard
 */
export default function AppearanceSection({application, editedApp, onFieldChange}: AppearanceSectionProps) {
  const {t} = useTranslation();
  const {data: themesData, isLoading: loadingThemes} = useGetThemes();

  const themeOptions = themesData?.themes ?? [];

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
          value={themeOptions.find((theme) => theme.id === (editedApp.theme_id! ?? application.theme_id!)) ?? null}
          onChange={(_event, newValue) => onFieldChange('theme_id' as keyof Application, newValue?.id ?? '')}
          loading={loadingThemes}
          renderInput={(params) => (
            <TextField
              {...params}
              placeholder={t('applications:edit.customization.theme.placeholder')}
              helperText={t('applications:edit.customization.theme.hint')}
              InputProps={{
                ...params.InputProps,
                endAdornment: (
                  <>
                    {loadingThemes ? <CircularProgress color="inherit" size={20} /> : null}
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
