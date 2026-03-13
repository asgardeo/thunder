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

import {Box, Stack, Typography, TextField} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';
import {useForm, Controller} from 'react-hook-form';
import {zodResolver} from '@hookform/resolvers/zod';
import {z} from 'zod';
import type {Application} from '../../../models/application';
import SettingsCard from '../../../../../components/SettingsCard';

/**
 * Props for the {@link UrlsSection} component.
 */
interface UrlsSectionProps {
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
 * Section component for configuring application URLs.
 *
 * Provides text fields for:
 * - Terms of Service URL
 * - Privacy Policy URL
 *
 * Includes URL validation using Zod schema and react-hook-form.
 * Changes are synced back to the parent component via onFieldChange.
 *
 * @param props - Component props
 * @returns URLs configuration UI within a SettingsCard
 */
export default function UrlsSection({application, editedApp, onFieldChange}: UrlsSectionProps) {
  const {t} = useTranslation();

  const urlsSchema = z.object({
    tosUri: z.string().url('Please enter a valid URL').or(z.literal('')).optional(),
    policyUri: z.string().url('Please enter a valid URL').or(z.literal('')).optional(),
  });

  type UrlsFormData = z.infer<typeof urlsSchema>;

  const {
    control,
    formState: {errors},
  } = useForm<UrlsFormData>({
    resolver: zodResolver(urlsSchema),
    mode: 'onChange',
    defaultValues: {
      tosUri: editedApp.tos_uri ?? application.tos_uri ?? '',
      policyUri: editedApp.policy_uri ?? application.policy_uri ?? '',
    },
  });

  return (
    <SettingsCard
      title={t('applications:edit.customization.sections.urls')}
      description={t('applications:edit.customization.sections.urls.description')}
    >
      <Stack spacing={3}>
        <Box>
          <Typography variant="subtitle2" gutterBottom>
            {t('applications:edit.customization.labels.tosUri')}
          </Typography>
          <Controller
            name="tosUri"
            control={control}
            render={({field}) => (
              <TextField
                {...field}
                onChange={(e) => {
                  field.onChange(e);
                  onFieldChange('tos_uri', e.target.value);
                }}
                fullWidth
                placeholder={t('applications:edit.customization.tosUri.placeholder')}
                error={!!errors.tosUri}
                helperText={errors.tosUri?.message ?? t('applications:edit.customization.tosUri.hint')}
              />
            )}
          />
        </Box>

        <Box>
          <Typography variant="subtitle2" gutterBottom>
            {t('applications:edit.customization.labels.policyUri')}
          </Typography>
          <Controller
            name="policyUri"
            control={control}
            render={({field}) => (
              <TextField
                {...field}
                onChange={(e) => {
                  field.onChange(e);
                  onFieldChange('policy_uri', e.target.value);
                }}
                fullWidth
                placeholder={t('applications:edit.customization.policyUri.placeholder')}
                error={!!errors.policyUri}
                helperText={errors.policyUri?.message ?? t('applications:edit.customization.policyUri.hint')}
              />
            )}
          />
        </Box>
      </Stack>
    </SettingsCard>
  );
}
