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

import {Link, useNavigate} from 'react-router';
import {useForm, Controller} from 'react-hook-form';
import {useMemo, useState, useEffect} from 'react';
import {
  Box,
  Stack,
  Typography,
  Button,
  Paper,
  FormLabel,
  FormControl,
  Select,
  MenuItem,
  PageContent,
  PageTitle,
} from '@wso2/oxygen-ui';
import {Plus, Save} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import useGetUserSchemas from '../api/useGetUserSchemas';
import type {SchemaInterface} from '../types/users';
import useGetUserSchema from '../api/useGetUserSchema';
import useCreateUser from '../api/useCreateUser';
import renderSchemaField from '../utils/renderSchemaField';

interface CreateUserFormData {
  schema: string;
  [key: string]: string | number | boolean;
}

export default function CreateUserPage() {
  const navigate = useNavigate();
  const {t} = useTranslation();
  const [selectedSchema, setSelectedSchema] = useState<SchemaInterface>();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const {data: originalUserSchemas} = useGetUserSchemas();
  const {createUser, error: createUserError} = useCreateUser();

  const {
    data: defaultUserSchema,
    loading: isDefaultUserSchemaRequestLoading,
    error: defaultUserSchemaRequestError,
  } = useGetUserSchema(selectedSchema?.id);

  const userSchemas: SchemaInterface[] = useMemo(() => {
    if (!originalUserSchemas?.schemas) {
      return [];
    }

    if (originalUserSchemas.schemas.length > 0 && !selectedSchema) {
      setSelectedSchema(originalUserSchemas.schemas[0]);
    }

    return originalUserSchemas?.schemas;
  }, [originalUserSchemas, selectedSchema]);

  const {
    control,
    handleSubmit,
    setValue,
    formState: {errors},
  } = useForm<CreateUserFormData>({
    defaultValues: {
      schema: '',
    },
  });

  // Set default schema when schemas are loaded
  useEffect(() => {
    if (selectedSchema) {
      setValue('schema', selectedSchema.name);
    }
  }, [selectedSchema, setValue]);

  const onSubmit = async (data: CreateUserFormData) => {
    try {
      setIsSubmitting(true);

      // Extract schema from form data (schema already contains the schema name)
      const {schema, ...attributes} = data;

      const trimmedOuId = selectedSchema?.ouId?.trim();
      const organizationUnitId = trimmedOuId === '' ? undefined : trimmedOuId;

      if (!organizationUnitId) {
        throw new Error('Organization unit ID is missing for the selected user type');
      }

      // Prepare the request body according to the API spec
      const requestBody = {
        organizationUnit: organizationUnitId,
        type: schema,
        attributes,
      };

      // Call the API to create the user
      await createUser(requestBody);

      // Navigate to users list on success
      await navigate('/users');
    } catch (error) {
      // Error is already handled in the hook
      // eslint-disable-next-line no-console
      console.error('Failed to create user:', error);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCancel = async () => {
    await navigate('/users');
  };

  const handleCreateUserType = () => {
    // TODO: Implement navigation to create user schema page
    // eslint-disable-next-line no-console
    console.log('Navigate to create user type page');
  };

  return (
    <PageContent>
      {/* Header */}
      <PageTitle>
        <PageTitle.BackButton component={<Link to="/users" />} />
        <PageTitle.Header>{t('users:createUser.title')}</PageTitle.Header>
        <PageTitle.SubHeader>{t('users:createUser.subtitle')}</PageTitle.SubHeader>
      </PageTitle>

      <Paper sx={{p: 4}}>
        <Box
          component="form"
          onSubmit={(event) => {
            handleSubmit(onSubmit)(event).catch(() => {
              // Handle form submission error
            });
          }}
          noValidate
          sx={{display: 'flex', flexDirection: 'column', gap: 2}}
        >
          {/* Schema Select Field with Create Button */}
          <Box>
            <FormLabel htmlFor="schema" sx={{mb: 1, display: 'block'}}>
              User Type
            </FormLabel>
            <Stack direction="row" spacing={2} alignItems="flex-start">
              <FormControl sx={{flexGrow: 1}} error={!!errors.schema}>
                <Controller
                  name="schema"
                  control={control}
                  rules={{
                    required: 'User type is required',
                  }}
                  render={({field}) => (
                    <Select
                      {...field}
                      id="schema"
                      value={field.value ?? selectedSchema?.name}
                      onChange={(e) => {
                        field.onChange(e);
                        const schema = userSchemas.find((s) => s.name === e.target.value);
                        setSelectedSchema(schema);
                      }}
                      required
                      error={!!errors.schema}
                      displayEmpty
                    >
                      {userSchemas.length === 0 ? (
                        <MenuItem value="" disabled>
                          Loading schemas...
                        </MenuItem>
                      ) : (
                        userSchemas.map((schema) => (
                          <MenuItem key={schema.name} value={schema.name}>
                            {schema.name}
                          </MenuItem>
                        ))
                      )}
                    </Select>
                  )}
                />
                {errors.schema && (
                  <Typography variant="caption" color="error" sx={{mt: 0.5, ml: 1.75}}>
                    {errors.schema.message}
                  </Typography>
                )}
              </FormControl>
              <Button variant="outlined" startIcon={<Plus size={16} />} onClick={handleCreateUserType}>
                Create
              </Button>
            </Stack>
          </Box>

          {/* Dynamic Schema Fields */}
          {isDefaultUserSchemaRequestLoading && (
            <Box sx={{textAlign: 'center', py: 4}}>
              <Typography variant="body2" color="text.secondary">
                Loading schema fields...
              </Typography>
            </Box>
          )}

          {defaultUserSchemaRequestError && (
            <Box sx={{textAlign: 'center', py: 4}}>
              <Typography variant="body2" color="error">
                Error loading schema: {defaultUserSchemaRequestError.message}
              </Typography>
            </Box>
          )}

          {defaultUserSchema?.schema &&
            Object.entries(defaultUserSchema.schema).map(([fieldName, fieldDef]) =>
              renderSchemaField(fieldName, fieldDef, control, errors),
            )}

          {/* Create User Error Display */}
          {createUserError && (
            <Box sx={{p: 2, bgcolor: 'error.light', borderRadius: 1}}>
              <Typography variant="body2" color="error.dark" sx={{fontWeight: 'bold'}}>
                {createUserError.message}
              </Typography>
              {createUserError.description && (
                <Typography variant="caption" color="error.dark">
                  {createUserError.description}
                </Typography>
              )}
            </Box>
          )}

          {/* Form Actions */}
          <Stack direction="row" spacing={2} justifyContent="flex-end" sx={{mt: 2}}>
            <Button
              variant="outlined"
              onClick={() => {
                handleCancel().catch(() => {
                  // Handle navigation error
                });
              }}
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              variant="contained"
              startIcon={isSubmitting ? null : <Save size={16} />}
              disabled={isSubmitting}
            >
              {isSubmitting ? 'Creating...' : 'Create User'}
            </Button>
          </Stack>
        </Box>
      </Paper>
    </PageContent>
  );
}
