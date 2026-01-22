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
import React from 'react';
import {useForm, Controller} from 'react-hook-form';
import {zodResolver} from '@hookform/resolvers/zod';
import {z} from 'zod';
import type {Application} from '../../../models/application';
import SettingsCard from '../SettingsCard';

interface UrlsSectionProps {
  application: Application;
  editedApp: Partial<Application>;
  onFieldChange: (field: keyof Application, value: unknown) => void;
}

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
    watch,
  } = useForm<UrlsFormData>({
    resolver: zodResolver(urlsSchema),
    mode: 'onChange',
    defaultValues: {
      tosUri: editedApp.tos_uri ?? application.tos_uri ?? '',
      policyUri: editedApp.policy_uri ?? application.policy_uri ?? '',
    },
  });

  const tosUri = watch('tosUri');
  const policyUri = watch('policyUri');

  // Sync form changes back to parent
  React.useEffect(() => {
    if (tosUri !== (editedApp.tos_uri ?? application.tos_uri ?? '')) {
      onFieldChange('tos_uri', tosUri);
    }
  }, [tosUri, editedApp.tos_uri, application.tos_uri, onFieldChange]);

  React.useEffect(() => {
    if (policyUri !== (editedApp.policy_uri ?? application.policy_uri ?? '')) {
      onFieldChange('policy_uri', policyUri);
    }
  }, [policyUri, editedApp.policy_uri, application.policy_uri, onFieldChange]);

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
