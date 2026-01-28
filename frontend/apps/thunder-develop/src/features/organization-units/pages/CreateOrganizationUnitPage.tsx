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

import {useState, useMemo, useRef, type JSX} from 'react';
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
  Chip,
  useTheme,
  Autocomplete,
} from '@wso2/oxygen-ui';
import {X, Lightbulb} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import {useForm, Controller} from 'react-hook-form';
import {zodResolver} from '@hookform/resolvers/zod';
import {z} from 'zod';
import {useLogger} from '@thunder/logger/react';
import useCreateOrganizationUnit from '../api/useCreateOrganizationUnit';
import useGetOrganizationUnits from '../api/useGetOrganizationUnits';
import type {CreateOrganizationUnitRequest, OrganizationUnit} from '../types/organization-units';
import generateOUNameSuggestions from '../utils/generateOUNameSuggestions';

/**
 * Creates a Zod schema for the create organization unit form with i18n support.
 * Validates name, handle, description, and parent fields.
 */
const createFormSchema = (t: (key: string) => string) =>
  z.object({
    name: z.string().trim().min(1, t('organizationUnits:form.validation.nameRequired')),
    handle: z
      .string()
      .trim()
      .min(1, t('organizationUnits:form.validation.handleRequired'))
      .regex(/^[a-z0-9-]+$/, t('organizationUnits:form.validation.handleFormat')),
    description: z.string().optional(),
    parentId: z.string().nullable(),
  });

/**
 * Type definition for form data inferred from the Zod schema.
 */
type FormData = z.infer<ReturnType<typeof createFormSchema>>;

export default function CreateOrganizationUnitPage(): JSX.Element {
  const navigate = useNavigate();
  const {t} = useTranslation();
  const theme = useTheme();
  const logger = useLogger('CreateOrganizationUnitPage');
  const createOrganizationUnit = useCreateOrganizationUnit();
  const {data: organizationUnitsData} = useGetOrganizationUnits();

  const [error, setError] = useState<string | null>(null);
  const isHandleManuallyEditedRef = useRef<boolean>(false);

  const formSchema = useMemo(() => createFormSchema(t), [t]);

  const {
    control,
    handleSubmit,
    setValue,
    formState: {errors, isValid},
  } = useForm<FormData>({
    resolver: zodResolver(formSchema),
    mode: 'onChange',
    defaultValues: {
      name: '',
      handle: '',
      description: '',
      parentId: null,
    },
  });

  const nameSuggestions: string[] = useMemo((): string[] => generateOUNameSuggestions(), []);
  const availableParentOUs: OrganizationUnit[] = useMemo(
    () => organizationUnitsData?.organizationUnits ?? [],
    [organizationUnitsData],
  );

  /**
   * Generates a handle from the name by lowercasing and replacing spaces with hyphens.
   */
  const generateHandleFromName = (nameValue: string): string => nameValue.toLowerCase().replace(/\s+/g, '-');

  const handleClose = (): void => {
    (async (): Promise<void> => {
      await navigate('/organization-units');
    })().catch((_error: unknown) => {
      logger.error('Failed to navigate back to organization units list', {error: _error});
    });
  };

  const handleNameChange = (newName: string): void => {
    setValue('name', newName, {shouldValidate: true});
    // Auto-generate handle if user hasn't manually edited it
    if (!isHandleManuallyEditedRef.current) {
      setValue('handle', generateHandleFromName(newName), {shouldValidate: true});
    }
  };

  const handleHandleChange = (newHandle: string): void => {
    setValue('handle', newHandle, {shouldValidate: true});
    isHandleManuallyEditedRef.current = true;
  };

  const handleNameSuggestionClick = (suggestion: string): void => {
    setValue('name', suggestion, {shouldValidate: true});
    // Auto-generate handle from suggestion if user hasn't manually edited it
    if (!isHandleManuallyEditedRef.current) {
      setValue('handle', generateHandleFromName(suggestion), {shouldValidate: true});
    }
  };

  const onSubmit = (data: FormData): void => {
    setError(null);

    const requestData: CreateOrganizationUnitRequest = {
      handle: data.handle,
      name: data.name,
      description: data.description?.trim() ? data.description.trim() : null,
      parent: data.parentId,
    };

    createOrganizationUnit.mutate(requestData, {
      onSuccess: () => {
        (async (): Promise<void> => {
          await navigate('/organization-units');
        })().catch((_error: unknown) => {
          logger.error('Failed to navigate after creating organization unit', {error: _error});
        });
      },
      onError: (err: Error) => {
        setError(err.message ?? t('organizationUnits:create.error'));
      },
    });
  };

  return (
    <Box sx={{minHeight: '100vh', display: 'flex', flexDirection: 'column'}}>
      {/* Progress bar at the very top - single step so 100% */}
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
                sx={{
                  bgcolor: 'background.paper',
                  '&:hover': {bgcolor: 'action.hover'},
                  boxShadow: 1,
                }}
              >
                <X size={24} />
              </IconButton>
              <Typography variant="h5">{t('organizationUnits:create.title')}</Typography>
            </Stack>
          </Box>

          {/* Main content */}
          <Box sx={{flex: 1, display: 'flex', minHeight: 0}}>
            {/* Left side - Form content */}
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
                    {/* Large heading - matching application create style */}
                    <Typography variant="h1" gutterBottom>
                      {t('organizationUnits:create.heading')}
                    </Typography>

                    {/* Name field first */}
                    <FormControl fullWidth required>
                      <FormLabel htmlFor="ou-name-input">{t('organizationUnits:form.name')}</FormLabel>
                      <Controller
                        name="name"
                        control={control}
                        render={({field}) => (
                          <TextField
                            {...field}
                            fullWidth
                            id="ou-name-input"
                            onChange={(e) => handleNameChange(e.target.value)}
                            placeholder={t('organizationUnits:form.namePlaceholder')}
                            error={!!errors.name}
                            helperText={errors.name?.message}
                          />
                        )}
                      />
                    </FormControl>

                    {/* Name suggestions */}
                    <Stack direction="column" spacing={2}>
                      <Stack direction="row" alignItems="center" spacing={1}>
                        <Lightbulb size={20} color={theme.vars?.palette.warning.main} />
                        <Typography variant="body2" color="text.secondary">
                          {t('organizationUnits:create.suggestions.label')}
                        </Typography>
                      </Stack>
                      <Box sx={{display: 'flex', flexWrap: 'wrap', gap: 1}}>
                        {nameSuggestions.map(
                          (suggestion: string): JSX.Element => (
                            <Chip
                              key={suggestion}
                              label={suggestion}
                              onClick={(): void => handleNameSuggestionClick(suggestion)}
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
                          ),
                        )}
                      </Box>
                    </Stack>

                    {/* Handle field */}
                    <FormControl fullWidth required>
                      <FormLabel htmlFor="ou-handle-input">{t('organizationUnits:form.handle')}</FormLabel>
                      <Controller
                        name="handle"
                        control={control}
                        render={({field}) => (
                          <TextField
                            {...field}
                            fullWidth
                            id="ou-handle-input"
                            onChange={(e) => handleHandleChange(e.target.value)}
                            placeholder={t('organizationUnits:form.handlePlaceholder')}
                            error={!!errors.handle}
                            helperText={errors.handle?.message ?? t('organizationUnits:form.handleHelperText')}
                          />
                        )}
                      />
                    </FormControl>

                    {/* Description field */}
                    <FormControl fullWidth>
                      <FormLabel htmlFor="ou-description-input">{t('organizationUnits:form.description')}</FormLabel>
                      <Controller
                        name="description"
                        control={control}
                        render={({field}) => (
                          <TextField
                            {...field}
                            fullWidth
                            id="ou-description-input"
                            placeholder={t('organizationUnits:form.descriptionPlaceholder')}
                            multiline
                            rows={3}
                          />
                        )}
                      />
                    </FormControl>

                    {/* Parent OU field */}
                    <FormControl fullWidth>
                      <FormLabel htmlFor="ou-parent-input">{t('organizationUnits:form.parent')}</FormLabel>
                      <Controller
                        name="parentId"
                        control={control}
                        render={({field}) => (
                          <Autocomplete
                            id="ou-parent-input"
                            options={availableParentOUs}
                            getOptionLabel={(option: OrganizationUnit) => option.name}
                            value={availableParentOUs.find((ou) => ou.id === field.value) ?? null}
                            onChange={(_event, newValue: OrganizationUnit | null) =>
                              field.onChange(newValue?.id ?? null)
                            }
                            renderInput={(params) => (
                              <TextField
                                {...params}
                                placeholder={t('organizationUnits:form.parentPlaceholder')}
                                helperText={t('organizationUnits:form.parentHelperText')}
                              />
                            )}
                            isOptionEqualToValue={(option, value) => option.id === value.id}
                          />
                        )}
                      />
                    </FormControl>

                    {/* Navigation buttons */}
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
                        disabled={createOrganizationUnit.isPending || !isValid}
                        sx={{minWidth: 100}}
                      >
                        {createOrganizationUnit.isPending ? t('common:status.saving') : t('common:actions.create')}
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
