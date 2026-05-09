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

import {zodResolver} from '@hookform/resolvers/zod';
import {generateRandomHumanReadableIdentifiers} from '@thunderid/utils';
import {Box, Chip, FormControl, FormLabel, Stack, TextField, Typography, useTheme} from '@wso2/oxygen-ui';
import {Lightbulb} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useEffect, useMemo, useRef} from 'react';
import {Controller, useForm} from 'react-hook-form';
import {useTranslation} from 'react-i18next';
import {z} from 'zod';

const formSchema = z.object({
  name: z.string().trim().min(1),
  handle: z
    .string()
    .trim()
    .min(1)
    .regex(/^[a-z0-9-]+$/),
});

type FormData = z.infer<typeof formSchema>;

export interface ConfigureFlowNameValue {
  name: string;
  handle: string;
}

interface ConfigureFlowNameProps {
  value: ConfigureFlowNameValue;
  onChange: (value: ConfigureFlowNameValue) => void;
  onReadyChange: (ready: boolean) => void;
}

export default function ConfigureFlowName({value, onChange, onReadyChange}: ConfigureFlowNameProps): JSX.Element {
  const {t} = useTranslation();
  const theme = useTheme();
  const isHandleManuallyEditedRef = useRef(false);

  const nameSuggestions = useMemo(() => generateRandomHumanReadableIdentifiers(), []);

  const generateHandle = (name: string): string =>
    name
      .toLowerCase()
      .replace(/\s+/g, '-')
      .replace(/[^a-z0-9-]/g, '');

  const {
    control,
    setValue,
    formState: {isValid},
  } = useForm<FormData>({
    resolver: zodResolver(formSchema),
    mode: 'onChange',
    defaultValues: {name: value.name, handle: value.handle},
  });

  useEffect(() => {
    onReadyChange(isValid);
  }, [isValid, onReadyChange]);

  const handleNameChange = (newName: string): void => {
    setValue('name', newName, {shouldValidate: true});
    onChange({
      name: newName,
      handle: isHandleManuallyEditedRef.current ? value.handle : generateHandle(newName),
    });
    if (!isHandleManuallyEditedRef.current) {
      setValue('handle', generateHandle(newName), {shouldValidate: true});
    }
  };

  const handleHandleChange = (newHandle: string): void => {
    isHandleManuallyEditedRef.current = true;
    setValue('handle', newHandle, {shouldValidate: true});
    onChange({name: value.name, handle: newHandle});
  };

  const handleSuggestionClick = (suggestion: string): void => {
    setValue('name', suggestion, {shouldValidate: true});
    onChange({
      name: suggestion,
      handle: isHandleManuallyEditedRef.current ? value.handle : generateHandle(suggestion),
    });
    if (!isHandleManuallyEditedRef.current) {
      setValue('handle', generateHandle(suggestion), {shouldValidate: true});
    }
  };

  return (
    <Stack direction="column" spacing={4} data-testid="configure-flow-name">
      <Typography variant="h1" gutterBottom>
        {t('flows:create.configure.title', 'Name your flow')}
      </Typography>

      <FormControl fullWidth required>
        <FormLabel htmlFor="flow-name-input">{t('flows:create.configure.name.label', 'Flow name')}</FormLabel>
        <Controller
          name="name"
          control={control}
          render={({field, fieldState}) => (
            <TextField
              {...field}
              fullWidth
              id="flow-name-input"
              onChange={(e) => handleNameChange(e.target.value)}
              placeholder={t('flows:create.configure.name.placeholder', 'e.g. Customer Sign-in')}
              error={!!fieldState.error}
            />
          )}
        />
      </FormControl>

      <Stack direction="column" spacing={2}>
        <Stack direction="row" alignItems="center" spacing={1}>
          <Lightbulb size={20} color={theme.vars?.palette.warning.main} />
          <Typography variant="body2" color="text.secondary">
            {t('flows:create.configure.suggestions.label', 'Need inspiration? Try one of these:')}
          </Typography>
        </Stack>
        <Box sx={{display: 'flex', flexWrap: 'wrap', gap: 1}}>
          {nameSuggestions.map((suggestion) => (
            <Chip
              key={suggestion}
              label={suggestion}
              onClick={() => handleSuggestionClick(suggestion)}
              variant="outlined"
              clickable
              sx={{
                '&:hover': {
                  bgcolor: 'primary.main',
                  color: 'primary.contrastText',
                  borderColor: 'primary.main',
                },
              }}
            />
          ))}
        </Box>
      </Stack>

      <FormControl fullWidth required>
        <FormLabel htmlFor="flow-handle-input">{t('flows:create.configure.handle.label', 'Handle')}</FormLabel>
        <Controller
          name="handle"
          control={control}
          render={({field, fieldState}) => (
            <TextField
              {...field}
              fullWidth
              id="flow-handle-input"
              onChange={(e) => handleHandleChange(e.target.value)}
              placeholder={t('flows:create.configure.handle.placeholder', 'e.g. customer-sign-in')}
              error={!!fieldState.error}
              helperText={
                fieldState.error?.message ??
                t('flows:create.configure.handle.hint', 'Lowercase letters, numbers, and hyphens only')
              }
            />
          )}
        />
      </FormControl>
    </Stack>
  );
}
