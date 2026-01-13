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
  Typography,
  Stack,
  TextField,
  Radio,
  RadioGroup,
  FormControlLabel,
  Alert,
  FormControl,
  FormLabel,
  Autocomplete,
} from '@wso2/oxygen-ui';
import {Globe} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useEffect} from 'react';
import {useTranslation} from 'react-i18next';
import {useForm, Controller, useWatch} from 'react-hook-form';
import {zodResolver} from '@hookform/resolvers/zod';
import {z} from 'zod';
import {useLogger} from '@thunder/logger/react';
import type {UserSchemaListItem} from '@/features/user-types/types/user-types';
import getConfigurationTypeFromTemplate from '../../utils/getConfigurationTypeFromTemplate';
import type {PlatformApplicationTemplate, TechnologyApplicationTemplate} from '../../models/application-templates';
import useApplicationCreate from '../../contexts/ApplicationCreate/useApplicationCreate';
import {ApplicationCreateFlowConfiguration} from '../../models/application-create-flow';

/**
 * Zod schema for validating URL inputs (hosting URLs and callback URLs).
 * Ensures URLs are properly formatted with http:// or https:// protocol.
 *
 * @internal
 */
const urlSchema: z.ZodString = z
  .string()
  .trim()
  .min(1, 'URL is required')
  .url('Please enter a valid URL')
  .refine((url) => url.startsWith('http://') || url.startsWith('https://'), {
    message: 'URL must start with http:// or https://',
  });

/**
 * Zod schema for validating deep links and universal links for mobile applications.
 * Accepts custom URL schemes (e.g., myapp://) or universal links (https://).
 *
 * @internal
 */
const deeplinkSchema: z.ZodString = z
  .string()
  .trim()
  .min(1, 'Deep link is required')
  .refine(
    (link) =>
      // Allow custom URL schemes (e.g., myapp://) or universal links (https://)
      /^[a-zA-Z][a-zA-Z0-9+.-]*:\/\/.+/.test(link),
    {
      message:
        'Please enter a valid deep link or universal link (e.g., myapp://callback or https://example.com/callback)',
    },
  );

/**
 * Zod schema for the configuration details form.
 * Validates hosting URLs, callback URLs, deep links, and user type selections
 * based on the configuration type required by the selected application template.
 *
 * @internal
 */
const formSchema = z
  .object({
    hostingUrl: z.string().optional(),
    callbackUrl: z.string().optional(),
    callbackMode: z.enum(['same', 'custom']),
    deeplink: z.string().optional(),
    userTypes: z.array(z.string()).optional(),
  })
  .superRefine((data, ctx) => {
    // Validate hostingUrl for URL-based platforms
    if (data.hostingUrl !== undefined && data.hostingUrl !== '') {
      const result = urlSchema.safeParse(data.hostingUrl);
      if (!result.success) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: result.error.issues[0]?.message || 'Invalid URL',
          path: ['hostingUrl'],
        });
      }
    }

    // Validate callbackUrl when custom mode
    if (data.callbackMode === 'custom' && data.callbackUrl) {
      const result = urlSchema.safeParse(data.callbackUrl);
      if (!result.success) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: result.error.issues[0]?.message || 'Invalid callback URL',
          path: ['callbackUrl'],
        });
      }
    }

    // Validate deeplink for mobile platforms
    if (data.deeplink !== undefined && data.deeplink !== '') {
      const result = deeplinkSchema.safeParse(data.deeplink);
      if (!result.success) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: result.error.issues[0]?.message || 'Invalid deep link',
          path: ['deeplink'],
        });
      }
    }
  });

/**
 * Type definition for form data inferred from the Zod schema.
 *
 * @internal
 */
type FormData = z.infer<typeof formSchema>;

/**
 * Props for the {@link ConfigureDetails} component.
 *
 * @public
 */
export interface ConfigureDetailsProps {
  /**
   * The selected technology template (e.g., React, Next.js, Angular)
   */
  technology: TechnologyApplicationTemplate | null;

  /**
   * The selected platform template (e.g., Browser, Mobile, Backend)
   */
  platform: PlatformApplicationTemplate | null;

  /**
   * Callback function invoked when the hosting URL changes
   */
  onHostingUrlChange: (url: string) => void;

  /**
   * Callback function invoked when the callback URL (or deep link) changes
   */
  onCallbackUrlChange: (url: string) => void;

  /**
   * Callback function to notify parent component whether this step is ready to proceed
   */
  onReadyChange: (isReady: boolean) => void;

  /**
   * Optional array of available user types for selection
   */
  userTypes?: UserSchemaListItem[];

  /**
   * Optional array of currently selected user type names
   */
  selectedUserTypes?: string[];

  /**
   * Optional callback function invoked when user type selection changes
   */
  onUserTypesChange?: (userTypes: string[]) => void;
}

/**
 * React component that renders the configuration details step in the
 * application creation onboarding flow.
 *
 * This component dynamically displays configuration options based on the selected
 * application template's requirements. It handles three configuration types:
 *
 * 1. **URL Configuration** (Browser/Server applications):
 *    - Hosting URL input for where the application is hosted
 *    - Callback URL configuration with options to use the same URL or a custom one
 *    - Real-time validation and synchronization of callback URL when "same as hosting" is selected
 *
 * 2. **Deep Link Configuration** (Mobile applications):
 *    - Deep link or universal link input for mobile app authentication redirects
 *    - Validation for custom URL schemes (e.g., myapp://) and universal links
 *
 * 3. **No Configuration** (Backend services):
 *    - Displays a message indicating no additional configuration is needed
 *
 * Additionally, if the selected template requires user type selection (indicated by an empty
 * allowed_user_types array) and multiple user types are available, the component displays
 * a multi-select autocomplete for choosing applicable user types.
 *
 * The component uses React Hook Form with Zod validation to provide real-time form
 * validation and error messages. It notifies the parent component of readiness status
 * based on form validity and configuration requirements.
 *
 * @param props - The component props
 * @param props.technology - Selected technology template
 * @param props.platform - Selected platform template
 * @param props.onHostingUrlChange - Callback for hosting URL changes
 * @param props.onCallbackUrlChange - Callback for callback URL changes
 * @param props.onReadyChange - Callback for step readiness changes
 * @param props.userTypes - Available user types for selection
 * @param props.selectedUserTypes - Currently selected user type names
 * @param props.onUserTypesChange - Callback for user type selection changes
 *
 * @returns JSX element displaying the appropriate configuration interface
 *
 * @example
 * ```tsx
 * import ConfigureDetails from './ConfigureDetails';
 *
 * function OnboardingFlow() {
 *   const [hostingUrl, setHostingUrl] = useState('');
 *   const [callbackUrl, setCallbackUrl] = useState('');
 *   const [isReady, setIsReady] = useState(false);
 *
 *   return (
 *     <ConfigureDetails
 *       technology="react"
 *       platform="browser"
 *       onHostingUrlChange={setHostingUrl}
 *       onCallbackUrlChange={setCallbackUrl}
 *       onReadyChange={setIsReady}
 *       userTypes={[{id: '1', name: 'Customer'}, {id: '2', name: 'Employee'}]}
 *       selectedUserTypes={['Customer']}
 *       onUserTypesChange={(types) => console.log('Selected types:', types)}
 *     />
 *   );
 * }
 * ```
 *
 * @public
 */
export default function ConfigureDetails({
  onHostingUrlChange,
  onCallbackUrlChange,
  onReadyChange,
  userTypes = [],
  selectedUserTypes = [],
  onUserTypesChange = undefined,
}: ConfigureDetailsProps): JSX.Element {
  const {t} = useTranslation();
  const logger = useLogger('ConfigureDetails');
  const {selectedTemplateConfig} = useApplicationCreate();
  const {
    control,
    formState: {errors, isValid},
    setValue,
    trigger,
  } = useForm<FormData>({
    resolver: zodResolver(formSchema),
    mode: 'onChange',
    defaultValues: {
      hostingUrl: '',
      callbackUrl: '',
      callbackMode: 'same',
      deeplink: '',
      userTypes: selectedUserTypes,
    },
  });

  const configurationType: ApplicationCreateFlowConfiguration =
    getConfigurationTypeFromTemplate(selectedTemplateConfig);

  const hostingUrl: string = useWatch({control, name: 'hostingUrl'}) ?? '';
  const callbackUrl: string = useWatch({control, name: 'callbackUrl'}) ?? '';
  const callbackMode: 'same' | 'custom' = useWatch({control, name: 'callbackMode'}) ?? 'same';
  const deeplink: string = useWatch({control, name: 'deeplink'}) ?? '';
  const defaultHostDisplay: string = hostingUrl;

  /**
   * Sync callback URL with hosting URL when checkbox is checked.
   */
  useEffect((): void => {
    const syncCallbackUrl = async (): Promise<void> => {
      if (callbackMode === 'same') {
        setValue('callbackUrl', hostingUrl);
        onCallbackUrlChange(hostingUrl);

        try {
          await trigger('callbackUrl');
        } catch (error) {
          logger.error('Failed to trigger callback URL validation', {error});
        }
      }
    };

    syncCallbackUrl().catch((): void => {
      // optional: swallow/handle error
    });
  }, [callbackMode, hostingUrl, setValue, onCallbackUrlChange, trigger, logger]);

  /**
   * Notify parent of hosting URL changes.
   */
  useEffect((): void => {
    onHostingUrlChange(hostingUrl);
  }, [hostingUrl, onHostingUrlChange]);

  /**
   * Notify parent of callback URL changes (when not using same as hosting).
   */
  useEffect((): void => {
    if (callbackMode === 'custom') {
      onCallbackUrlChange(callbackUrl);
    }
  }, [callbackUrl, callbackMode, onCallbackUrlChange]);

  /**
   * Notify parent of deep link changes for mobile platforms.
   */
  useEffect((): void => {
    if (configurationType === ApplicationCreateFlowConfiguration.DEEPLINK) {
      onCallbackUrlChange(deeplink);
    }
  }, [deeplink, configurationType, onCallbackUrlChange]);

  /**
   * Determine if step is ready based on validity and configuration type.
   */
  useEffect((): void => {
    if (configurationType === ApplicationCreateFlowConfiguration.NONE) {
      onReadyChange(true);
      return;
    }

    // Check if user type selection is required and valid
    const requiresUserTypes: boolean =
      selectedTemplateConfig?.allowed_user_types !== undefined &&
      Array.isArray(selectedTemplateConfig.allowed_user_types) &&
      selectedTemplateConfig.allowed_user_types.length === 0;
    const hasMultipleUserTypes: boolean = userTypes.length > 1;
    const isUserTypeSelectionValid: boolean =
      !requiresUserTypes || !hasMultipleUserTypes || selectedUserTypes.length > 0;

    // For URL-based config, need valid hosting URL
    if (configurationType === ApplicationCreateFlowConfiguration.URL) {
      const hasValidHostingUrl: boolean = !!hostingUrl && !errors.hostingUrl;
      const hasValidCallbackUrl: boolean = callbackMode === 'same' || (!!callbackUrl && !errors.callbackUrl);
      onReadyChange(!!hasValidHostingUrl && !!hasValidCallbackUrl && isUserTypeSelectionValid);
      return;
    }

    // For deeplink config, need valid deeplink
    if (configurationType === ApplicationCreateFlowConfiguration.DEEPLINK) {
      onReadyChange(!!deeplink && !errors.deeplink && isUserTypeSelectionValid);
      return;
    }

    onReadyChange(isValid && isUserTypeSelectionValid);
  }, [
    isValid,
    configurationType,
    hostingUrl,
    callbackUrl,
    callbackMode,
    deeplink,
    errors,
    onReadyChange,
    userTypes.length,
    selectedUserTypes.length,
    selectedTemplateConfig,
  ]);

  // For platforms that don't require configuration
  if (configurationType === ApplicationCreateFlowConfiguration.NONE) {
    return (
      <Stack spacing={3}>
        <Box sx={{textAlign: 'center', py: 4}}>
          <Globe size={48} style={{color: 'var(--oxygen-palette-text-secondary)', marginBottom: '16px'}} />
          <Typography variant="h6" gutterBottom>
            {t('applications:onboarding.configure.details.noConfigRequired.title')}
          </Typography>
          <Typography variant="body2" color="text.secondary">
            {t('applications:onboarding.configure.details.noConfigRequired.description')}
          </Typography>
        </Box>
      </Stack>
    );
  }

  return (
    <Stack spacing={3}>
      <Stack direction="column" spacing={1}>
        <Typography variant="h1" gutterBottom>
          {t('applications:onboarding.configure.details.title')}
        </Typography>
        <Typography variant="subtitle1" gutterBottom>
          {configurationType === ApplicationCreateFlowConfiguration.DEEPLINK
            ? t('applications:onboarding.configure.details.mobile.description')
            : t('applications:onboarding.configure.details.description')}
        </Typography>
      </Stack>

      {/* Mobile platform - Deep link / Universal link configuration */}
      {configurationType === ApplicationCreateFlowConfiguration.DEEPLINK && (
        <>
          <FormControl fullWidth required>
            <FormLabel htmlFor="deeplink-input">
              {t('applications:onboarding.configure.details.deeplink.label')}
            </FormLabel>
            <Controller
              name="deeplink"
              control={control}
              render={({field}) => (
                <TextField
                  {...field}
                  fullWidth
                  id="deeplink-input"
                  placeholder={t('applications:onboarding.configure.details.deeplink.placeholder')}
                  error={!!errors.deeplink}
                  helperText={
                    errors.deeplink?.message ?? t('applications:onboarding.configure.details.deeplink.helperText')
                  }
                />
              )}
            />
          </FormControl>

          <Alert severity="info">{t('applications:onboarding.configure.details.mobile.info')}</Alert>
        </>
      )}

      {/* Browser/Server platform - URL configuration */}
      {configurationType === ApplicationCreateFlowConfiguration.URL && (
        <>
          {/* Hosting URL */}
          <FormControl fullWidth required>
            <FormLabel htmlFor="hosting-url-input">
              {t('applications:onboarding.configure.details.hostingUrl.label')}
            </FormLabel>
            <Controller
              name="hostingUrl"
              control={control}
              render={({field}) => (
                <TextField
                  {...field}
                  fullWidth
                  id="hosting-url-input"
                  placeholder={t('applications:onboarding.configure.details.hostingUrl.placeholder')}
                  error={!!errors.hostingUrl}
                  helperText={
                    errors.hostingUrl?.message ?? t('applications:onboarding.configure.details.hostingUrl.helperText')
                  }
                />
              )}
            />
          </FormControl>

          {/* After Sign-in URL (Callback URL) */}
          <Stack spacing={2}>
            <FormControl component="fieldset">
              <FormLabel id="callback-url-label">
                {t('applications:onboarding.configure.details.callbackUrl.label')}
              </FormLabel>
              <Controller
                name="callbackMode"
                control={control}
                render={({field}) => (
                  <RadioGroup {...field} aria-labelledby="callback-url-label">
                    <FormControlLabel
                      value="same"
                      control={<Radio />}
                      label={
                        <Stack direction="row" alignItems="center" spacing={1}>
                          <Typography variant="body1">
                            {t('applications:onboarding.configure.details.callbackMode.same')}
                          </Typography>
                          {defaultHostDisplay && (
                            <Typography variant="body2" color="text.secondary">
                              ({defaultHostDisplay})
                            </Typography>
                          )}
                        </Stack>
                      }
                    />
                    <FormControlLabel
                      value="custom"
                      control={<Radio />}
                      label={t('applications:onboarding.configure.details.callbackMode.custom')}
                    />
                  </RadioGroup>
                )}
              />
            </FormControl>

            {callbackMode === 'custom' && (
              <FormControl fullWidth>
                <FormLabel htmlFor="callback-url-input" id="custom-callback-url-label">
                  {t('applications:onboarding.configure.details.callbackUrl.label')}
                </FormLabel>
                <Controller
                  name="callbackUrl"
                  control={control}
                  render={({field}) => (
                    <TextField
                      {...field}
                      fullWidth
                      id="callback-url-input"
                      placeholder={t('applications:onboarding.configure.details.callbackUrl.placeholder')}
                      error={!!errors.callbackUrl}
                      helperText={
                        errors.callbackUrl?.message ??
                        t('applications:onboarding.configure.details.callbackUrl.helperText')
                      }
                    />
                  )}
                />
              </FormControl>
            )}

            <Alert severity="info">{t('applications:onboarding.configure.details.callbackUrl.info')}</Alert>
          </Stack>
        </>
      )}

      {/* User Type Selection - Only show if template requires it (has empty allowed_user_types array) and there are multiple user types */}
      {selectedTemplateConfig?.allowed_user_types !== undefined &&
        Array.isArray(selectedTemplateConfig.allowed_user_types) &&
        selectedTemplateConfig.allowed_user_types.length === 0 &&
        userTypes.length > 1 &&
        onUserTypesChange && (
          <FormControl fullWidth required>
            <FormLabel htmlFor="user-types-autocomplete">
              {t('applications:onboarding.configure.details.userTypes.label')}
            </FormLabel>
            <Controller
              name="userTypes"
              control={control}
              rules={{
                validate: (value: string[] | undefined): string | boolean => {
                  if (userTypes.length > 1 && (!value || value.length === 0)) {
                    return t('applications:onboarding.configure.details.userTypes.error');
                  }
                  return true;
                },
              }}
              render={({field, fieldState}) => (
                <Autocomplete
                  multiple
                  id="user-types-autocomplete"
                  size="small"
                  options={userTypes}
                  getOptionLabel={(option) => option.name}
                  value={userTypes.filter((ut: UserSchemaListItem) => field.value?.includes(ut.name)) || []}
                  onChange={(_event, newValue: UserSchemaListItem[]): void => {
                    const userTypeNames: string[] = newValue.map((item: UserSchemaListItem): string => item.name);
                    field.onChange(userTypeNames);
                    onUserTypesChange(userTypeNames);
                  }}
                  onBlur={field.onBlur}
                  renderInput={(params) => (
                    <TextField
                      {...params}
                      placeholder={t('applications:onboarding.configure.details.userTypes.description')}
                      error={fieldState.isTouched && !!fieldState.error}
                      helperText={fieldState.isTouched && fieldState.error?.message}
                    />
                  )}
                  isOptionEqualToValue={(option: UserSchemaListItem, value: UserSchemaListItem): boolean =>
                    option.name === value.name
                  }
                />
              )}
            />
          </FormControl>
        )}
    </Stack>
  );
}
