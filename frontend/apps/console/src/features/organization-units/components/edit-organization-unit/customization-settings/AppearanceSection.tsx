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

import type {JSX} from 'react';
import {Box, Typography, TextField, Autocomplete, CircularProgress} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';
import {useGetThemes} from '@thunder/shared-design';
import SettingsCard from '@/components/SettingsCard';
import type {OrganizationUnit} from '../../../models/organization-unit';

/**
 * Props for the {@link AppearanceSection} component.
 */
interface AppearanceSectionProps {
  /**
   * The organization unit being edited
   */
  organizationUnit: OrganizationUnit;
  /**
   * Partial organization unit object containing edited fields
   */
  editedOU: Partial<OrganizationUnit>;
  /**
   * Callback function to handle field value changes
   * @param field - The organization unit field being updated
   * @param value - The new value for the field
   */
  onFieldChange: (field: keyof OrganizationUnit, value: unknown) => void;
}

/**
 * Section component for configuring organization unit appearance.
 *
 * Provides an autocomplete dropdown to select a theme from available options.
 * The selected theme affects the look and feel of the organization unit's pages.
 *
 * @param props - Component props
 * @returns Appearance configuration UI within a SettingsCard
 */
export default function AppearanceSection({
  organizationUnit,
  editedOU,
  onFieldChange,
}: AppearanceSectionProps): JSX.Element {
  const {t} = useTranslation();
  const {data: themesData, isLoading: loadingThemes} = useGetThemes();

  const themeOptions = themesData?.themes ?? [];

  return (
    <SettingsCard
      title={t('organizationUnits:edit.customization.sections.appearance')}
      description={t('organizationUnits:edit.customization.sections.appearance.description')}
    >
      <Box>
        <Typography variant="subtitle2" gutterBottom>
          {t('organizationUnits:edit.customization.labels.theme')}
        </Typography>
        <Autocomplete
          fullWidth
          options={themeOptions}
          getOptionLabel={(option) => (typeof option === 'string' ? option : option.displayName)}
          value={themeOptions.find((theme) => theme.id === (editedOU.theme_id ?? organizationUnit.theme_id)) ?? null}
          onChange={(_event, newValue) => onFieldChange('theme_id', newValue?.id ?? '')}
          loading={loadingThemes}
          renderInput={(params) => (
            <TextField
              {...params}
              placeholder={t('organizationUnits:edit.customization.theme.placeholder')}
              helperText={t('organizationUnits:edit.customization.theme.hint')}
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
