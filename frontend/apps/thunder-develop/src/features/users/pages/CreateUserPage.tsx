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

import {useNavigate} from 'react-router';
import {useForm, Controller} from 'react-hook-form';
import {useMemo, useState, useEffect} from 'react';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Paper from '@mui/material/Paper';
import FormLabel from '@mui/material/FormLabel';
import FormControl from '@mui/material/FormControl';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import {ArrowLeft, Plus, Save} from 'lucide-react';
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

      // Prepare the request body according to the API spec
      const requestBody = {
        organizationUnit: 'test-ou', // TODO: Add organization unit field or get from context
        type: schema,
        attributes,
      };

      // Call the API to create the user
      const result = await createUser(requestBody);

      // eslint-disable-next-line no-console
      console.log('User created successfully:', result);

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

  const handleBack = async () => {
    await navigate('/users');
  };

  const handleCreateUserType = () => {
    // TODO: Implement navigation to create user schema page
    // eslint-disable-next-line no-console
    console.log('Navigate to create user type page');
  };

  return (
    <Box sx={{maxWidth: 800, mx: 'auto', px: 2, position: 'relative', pt: 6}}>
      <Button
        onClick={() => {
          handleBack().catch(() => {
            // Handle navigation error
          });
        }}
        sx={{
          position: 'absolute',
          top: 0,
          left: 0,
          zIndex: 10,
          bgcolor: 'background.paper',
          ml: 2,
          p: 2,
        }}
        aria-label="Go back"
        startIcon={<ArrowLeft size={16} />}
      >
        Back
      </Button>

      <Stack direction="row" alignItems="flex-start" mb={4} gap={2}>
        <Box>
          <Typography variant="h4" component="h1" gutterBottom>
            Create New User
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Add a new user to your organization
          </Typography>
        </Box>
      </Stack>

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
    </Box>
  );
}
