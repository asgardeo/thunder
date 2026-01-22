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

import {
  Box,
  Stack,
  TextField,
  Chip,
  Autocomplete,
  CircularProgress,
  FormControl,
  FormLabel,
  Typography,
  Button,
  IconButton,
  Tooltip,
} from '@wso2/oxygen-ui';
import {Trash, Plus} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import React from 'react';
import {useForm, Controller} from 'react-hook-form';
import {zodResolver} from '@hookform/resolvers/zod';
import {z} from 'zod';
import useGetUserTypes from '../../../../user-types/api/useGetUserTypes';
import type {Application} from '../../../models/application';
import type {OAuth2Config} from '../../../models/oauth';
import SettingsCard from '../SettingsCard';

interface AccessSectionProps {
  application: Application;
  editedApp: Partial<Application>;
  oauth2Config?: OAuth2Config;
  onFieldChange: (field: keyof Application, value: unknown) => void;
}

export default function AccessSection({
  application,
  editedApp,
  oauth2Config = undefined,
  onFieldChange,
}: AccessSectionProps) {
  const {t} = useTranslation();
  const {data: userTypesData, loading: loadingUserTypes} = useGetUserTypes();

  // State for managing redirect URIs
  const [redirectUris, setRedirectUris] = React.useState<string[]>(oauth2Config?.redirect_uris ?? []);
  const [uriErrors, setUriErrors] = React.useState<Record<number, string>>({});

  const userTypeOptions = userTypesData?.schemas.map((schema) => schema.name) ?? [];

  const generalSettingsSchema = z.object({
    url: z.string().url('Please enter a valid URL').or(z.literal('')).optional(),
  });

  type GeneralSettingsFormData = z.infer<typeof generalSettingsSchema>;

  const {
    control,
    formState: {errors},
    watch,
  } = useForm<GeneralSettingsFormData>({
    resolver: zodResolver(generalSettingsSchema),
    mode: 'onChange',
    defaultValues: {
      url: editedApp.url ?? application.url ?? '',
    },
  });

  const url = watch('url');

  // Sync form changes back to parent
  React.useEffect(() => {
    if (url !== (editedApp.url ?? application.url ?? '')) {
      onFieldChange('url', url);
    }
  }, [url, editedApp.url, application.url, onFieldChange]);

  // Handlers for redirect URIs
  const validateUri = (uri: string, index: number): boolean => {
    if (!uri || uri.trim() === '') {
      setUriErrors((prev) => ({...prev, [index]: t('applications:edit.general.redirectUris.error.empty')}));
      return false;
    }
    try {
      new URL(uri);
      setUriErrors((prev) => {
        const newErrors = {...prev};
        delete newErrors[index];
        return newErrors;
      });
      return true;
    } catch {
      setUriErrors((prev) => ({...prev, [index]: t('applications:edit.general.redirectUris.error.invalid')}));
      return false;
    }
  };

  const handleAddUri = () => {
    setRedirectUris((prev) => [...prev, '']);
  };

  const handleRemoveUri = (index: number) => {
    setRedirectUris((prev) => prev.filter((_, i) => i !== index));
    setUriErrors((prev) => {
      const newErrors = {...prev};
      delete newErrors[index];
      // Reindex remaining errors
      const reindexed: Record<number, string> = {};
      Object.entries(newErrors).forEach(([key, value]) => {
        const oldIndex = parseInt(key, 10);
        if (oldIndex > index) {
          reindexed[oldIndex - 1] = value;
        } else if (oldIndex < index) {
          reindexed[oldIndex] = value;
        }
      });
      return reindexed;
    });
  };

  const handleUriChange = (index: number, value: string) => {
    setRedirectUris((prev) => {
      const newUris = [...prev];
      newUris[index] = value;
      return newUris;
    });
    // Clear error when user starts typing
    if (value.trim() !== '') {
      setUriErrors((prev) => {
        const newErrors = {...prev};
        delete newErrors[index];
        return newErrors;
      });
    }
  };

  const handleUriBlur = (index: number) => {
    const uri = redirectUris[index];
    if (validateUri(uri, index) && uri.trim() !== '') {
      // Update the parent component with valid URIs
      updateRedirectUris();
    }
  };

  const updateRedirectUris = () => {
    const validUris = redirectUris.filter((uri) => uri.trim() !== '');
    if (!oauth2Config) return;

    const updatedConfig = {
      ...oauth2Config,
      redirect_uris: validUris,
    };
    const updatedInboundAuth = application.inbound_auth_config?.map((config) => {
      if (config.type === 'oauth2') {
        return {...config, config: updatedConfig};
      }
      return config;
    });
    onFieldChange('inbound_auth_config', updatedInboundAuth);
  };

  // Initialize redirect URIs from oauth2Config
  React.useEffect(() => {
    if (oauth2Config?.redirect_uris) {
      setRedirectUris(oauth2Config.redirect_uris);
    }
  }, [oauth2Config?.redirect_uris]);

  return (
    <SettingsCard
      title={t('applications:edit.general.sections.access')}
      description={t('applications:edit.general.sections.access.description')}
    >
      <Stack spacing={3}>
        <FormControl fullWidth>
          <FormLabel htmlFor="allowed-user-types-autocomplete">
            {t('applications:edit.general.labels.allowedUserTypes')}
          </FormLabel>
          <Autocomplete
            multiple
            fullWidth
            id="allowed-user-types-autocomplete"
            options={userTypeOptions}
            value={editedApp.allowed_user_types ?? application.allowed_user_types ?? []}
            onChange={(_event, newValue) => onFieldChange('allowed_user_types', newValue)}
            loading={loadingUserTypes}
            renderInput={(params) => (
              <TextField
                {...params}
                placeholder={t('applications:edit.general.allowedUserTypes.placeholder')}
                helperText={t('applications:edit.general.allowedUserTypes.hint')}
                InputProps={{
                  ...params.InputProps,
                  endAdornment: (
                    <>
                      {loadingUserTypes ? <CircularProgress color="inherit" size={20} /> : null}
                      {params.InputProps.endAdornment}
                    </>
                  ),
                }}
              />
            )}
            renderTags={(value, getTagProps) =>
              value.map((option, index) => <Chip label={option} {...getTagProps({index})} key={option} />)
            }
            freeSolo={false}
            disableClearable={false}
          />
        </FormControl>

        <FormControl fullWidth>
          <FormLabel htmlFor="application-url-input">{t('applications:edit.general.labels.applicationUrl')}</FormLabel>
          <Controller
            name="url"
            control={control}
            render={({field}) => (
              <TextField
                {...field}
                fullWidth
                id="application-url-input"
                placeholder="https://example.com"
                error={!!errors.url}
                helperText={errors.url?.message ?? t('applications:edit.general.applicationUrl.hint')}
              />
            )}
          />
        </FormControl>

        {oauth2Config && (
          <FormControl fullWidth>
            <FormLabel htmlFor="redirect-uris-section">{t('applications:edit.general.redirectUris.title')}</FormLabel>
            <Typography variant="caption" color="text.secondary" sx={{display: 'block', mb: 2}}>
              {t('applications:edit.general.redirectUris.description')}
            </Typography>

            <Stack spacing={2} id="redirect-uris-section">
              {redirectUris.map((uri, index) => (
                <Stack key={uri} direction="row" spacing={1} alignItems="flex-start">
                  <FormControl fullWidth required sx={{flex: 1}}>
                    <TextField
                      fullWidth
                      id={`redirect-uri-${index}-input`}
                      value={uri}
                      onChange={(e) => handleUriChange(index, e.target.value)}
                      onBlur={() => handleUriBlur(index)}
                      error={!!uriErrors[index]}
                      helperText={uriErrors[index]}
                      placeholder="https://example.com/callback"
                    />
                  </FormControl>
                  <Tooltip title={t('common:actions.delete')}>
                    <IconButton onClick={() => handleRemoveUri(index)} color="error" sx={{mt: 1}}>
                      <Trash size={20} />
                    </IconButton>
                  </Tooltip>
                </Stack>
              ))}

              <Box>
                <Button variant="outlined" startIcon={<Plus />} onClick={handleAddUri} size="small">
                  {t('applications:edit.general.redirectUris.addUri')}
                </Button>
              </Box>
            </Stack>
          </FormControl>
        )}
      </Stack>
    </SettingsCard>
  );
}
