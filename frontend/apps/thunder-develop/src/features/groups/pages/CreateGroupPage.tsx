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

import {useState, useMemo, type JSX} from 'react';
import {useNavigate} from 'react-router';
import {
  Box,
  Stack,
  Typography,
  Button,
  TextField,
  Alert,
  IconButton,
  LinearProgress,
  FormControl,
  FormLabel,
} from '@wso2/oxygen-ui';
import {X} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import {useForm, Controller} from 'react-hook-form';
import {zodResolver} from '@hookform/resolvers/zod';
import {z} from 'zod';
import {useLogger} from '@thunder/logger/react';
import useCreateGroup from '../api/useCreateGroup';
import OrganizationUnitTreePicker from '../../organization-units/components/OrganizationUnitTreePicker';
import type {CreateGroupRequest} from '../models/requests';

/**
 * Creates a Zod schema for the create group form with i18n support.
 */
const createFormSchema = (t: (key: string) => string) =>
  z.object({
    name: z.string().trim().min(1, t('groups:create.form.name.required')),
    description: z.string().optional(),
    organizationUnitId: z.string().min(1, t('groups:create.form.organizationUnit.required')),
  });

type FormData = z.infer<ReturnType<typeof createFormSchema>>;

export default function CreateGroupPage(): JSX.Element {
  const navigate = useNavigate();
  const {t} = useTranslation();
  const logger = useLogger('CreateGroupPage');
  const createGroup = useCreateGroup();

  const [error, setError] = useState<string | null>(null);

  const formSchema = useMemo(() => createFormSchema(t), [t]);

  const {
    control,
    handleSubmit,
    formState: {errors, isValid},
  } = useForm<FormData>({
    resolver: zodResolver(formSchema),
    mode: 'onChange',
    defaultValues: {
      name: '',
      description: '',
      organizationUnitId: '',
    },
  });

  const listUrl = '/groups';

  const handleClose = (): void => {
    (async (): Promise<void> => {
      await navigate(listUrl);
    })().catch((_error: unknown) => {
      logger.error('Failed to navigate back to groups list', {error: _error});
    });
  };

  const onSubmit = (data: FormData): void => {
    setError(null);

    const requestData: CreateGroupRequest = {
      name: data.name,
      description: data.description?.trim() ? data.description.trim() : undefined,
      organizationUnitId: data.organizationUnitId,
    };

    createGroup.mutate(requestData, {
      onSuccess: () => {
        (async (): Promise<void> => {
          await navigate(listUrl);
        })().catch((_error: unknown) => {
          logger.error('Failed to navigate after creating group', {error: _error});
        });
      },
      onError: (err: Error) => {
        setError(err.message ?? t('groups:create.error'));
      },
    });
  };

  return (
    <Box sx={{minHeight: '100vh', display: 'flex', flexDirection: 'column'}}>
      <LinearProgress variant="determinate" value={100} sx={{height: 6}} />

      <Box sx={{flex: 1, display: 'flex', flexDirection: 'row'}}>
        <Box
          sx={{
            flex: 1,
            display: 'flex',
            flexDirection: 'column',
          }}
        >
          {/* Header with close button */}
          <Box sx={{p: 4, display: 'flex', justifyContent: 'space-between', alignItems: 'center'}}>
            <Stack direction="row" alignItems="center" spacing={2}>
              <IconButton
                onClick={handleClose}
                aria-label={t('common:actions.close')}
                sx={{
                  bgcolor: 'background.paper',
                  '&:hover': {bgcolor: 'action.hover'},
                  boxShadow: 1,
                }}
              >
                <X size={24} />
              </IconButton>
              <Typography variant="h5">{t('groups:create.title')}</Typography>
            </Stack>
          </Box>

          {/* Main content */}
          <Box sx={{flex: 1, display: 'flex', minHeight: 0}}>
            <Box
              sx={{
                flex: 1,
                display: 'flex',
                flexDirection: 'column',
                py: 8,
                px: 20,
              }}
            >
              <Box
                sx={{
                  width: '100%',
                  maxWidth: 800,
                  display: 'flex',
                  flexDirection: 'column',
                }}
              >
                {/* Error Alert */}
                {error && (
                  <Alert severity="error" sx={{my: 3}} onClose={() => setError(null)}>
                    {error}
                  </Alert>
                )}

                <form
                  onSubmit={(e) => {
                    e.preventDefault();
                    handleSubmit(onSubmit)(e).catch((err: unknown) => {
                      logger.error('Form submission error', {error: err});
                    });
                  }}
                >
                  <Stack direction="column" spacing={4}>
                    <Typography variant="h1" gutterBottom>
                      {t('groups:create.heading')}
                    </Typography>

                    {/* Name field */}
                    <FormControl fullWidth required>
                      <FormLabel htmlFor="group-name-input">{t('groups:create.form.name.label')}</FormLabel>
                      <Controller
                        name="name"
                        control={control}
                        render={({field}) => (
                          <TextField
                            {...field}
                            fullWidth
                            id="group-name-input"
                            placeholder={t('groups:create.form.name.placeholder')}
                            error={!!errors.name}
                            helperText={errors.name?.message}
                          />
                        )}
                      />
                    </FormControl>

                    {/* Description field */}
                    <FormControl fullWidth>
                      <FormLabel htmlFor="group-description-input">
                        {t('groups:create.form.description.label')}
                      </FormLabel>
                      <Controller
                        name="description"
                        control={control}
                        render={({field}) => (
                          <TextField
                            {...field}
                            fullWidth
                            id="group-description-input"
                            placeholder={t('groups:create.form.description.placeholder')}
                            multiline
                            rows={3}
                          />
                        )}
                      />
                    </FormControl>

                    {/* Organization Unit tree picker */}
                    <FormControl fullWidth required>
                      <FormLabel htmlFor="group-ou-picker">
                        {t('groups:create.form.organizationUnit.label')}
                      </FormLabel>
                      <Controller
                        name="organizationUnitId"
                        control={control}
                        render={({field}) => (
                          <OrganizationUnitTreePicker
                            id="group-ou-picker"
                            value={field.value}
                            onChange={field.onChange}
                            error={!!errors.organizationUnitId}
                            helperText={errors.organizationUnitId?.message}
                          />
                        )}
                      />
                    </FormControl>

                    {/* Submit button */}
                    <Box
                      sx={{
                        mt: 4,
                        display: 'flex',
                        justifyContent: 'flex-start',
                        gap: 2,
                      }}
                    >
                      <Button
                        type="submit"
                        variant="contained"
                        disabled={createGroup.isPending || !isValid}
                        sx={{minWidth: 100}}
                      >
                        {createGroup.isPending ? t('common:status.saving') : t('common:actions.create')}
                      </Button>
                    </Box>
                  </Stack>
                </form>
              </Box>
            </Box>
          </Box>
        </Box>
      </Box>
    </Box>
  );
}
