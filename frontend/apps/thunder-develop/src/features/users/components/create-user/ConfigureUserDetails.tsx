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

import {useEffect} from 'react';
import {useForm} from 'react-hook-form';
import {Box, Stack, Typography} from '@wso2/oxygen-ui';
import type {JSX} from 'react';
import {useTranslation} from 'react-i18next';
import type {ApiUserSchema} from '../../types/users';
import renderSchemaField from '../../utils/renderSchemaField';

/**
 * Props for the {@link ConfigureUserDetails} component.
 *
 * @public
 */
export interface ConfigureUserDetailsProps {
  schema: ApiUserSchema;
  defaultValues: Record<string, unknown>;
  onFormValuesChange: (values: Record<string, unknown>) => void;
  onReadyChange?: (isReady: boolean) => void;
}

type UserDetailsFormData = Record<string, unknown>;

/**
 * Step 2 of the user creation wizard: fill in the dynamic form fields.
 *
 * @public
 */
export default function ConfigureUserDetails({
  schema,
  defaultValues,
  onFormValuesChange,
  onReadyChange = undefined,
}: ConfigureUserDetailsProps): JSX.Element {
  const {t} = useTranslation();

  const {
    control,
    watch,
    formState: {errors, isValid},
  } = useForm<UserDetailsFormData>({
    defaultValues,
    mode: 'onChange',
  });

  useEffect(() => {
    const subscription = watch((values) => {
      onFormValuesChange(values as Record<string, unknown>);
    });

    return () => subscription.unsubscribe();
  }, [watch, onFormValuesChange]);

  useEffect((): void => {
    if (onReadyChange) {
      onReadyChange(isValid);
    }
  }, [isValid, onReadyChange]);

  return (
    <Stack direction="column" spacing={4} data-testid="configure-user-details">
      <Typography variant="h1" gutterBottom>
        {t('users:createWizard.userDetails.title')}
      </Typography>
      <Typography variant="body1" color="text.secondary">
        {t('users:createWizard.userDetails.subtitle')}
      </Typography>

      <Box sx={{display: 'flex', flexDirection: 'column', gap: 2}}>
        {schema.schema &&
          Object.entries(schema.schema).map(([fieldName, fieldDef]) =>
            renderSchemaField(fieldName, fieldDef, control, errors),
          )}
      </Box>
    </Stack>
  );
}
