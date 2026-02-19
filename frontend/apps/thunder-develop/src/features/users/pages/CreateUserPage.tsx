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
import {Box, Stack, Typography, Button, Paper, FormLabel, FormControl, Select, MenuItem} from '@wso2/oxygen-ui';
import {ArrowLeft, Plus, Save} from '@wso2/oxygen-ui-icons-react';
import useGetUserTypes from '../api/useGetUserTypes';
import type {UserTypeListItem} from '../types/users';
import useGetUserType from '../api/useGetUserType';
import useCreateUser from '../api/useCreateUser';
import renderSchemaField from '../utils/renderSchemaField';

interface CreateUserFormData {
  userType: string;
  [key: string]: string | number | boolean;
}

export default function CreateUserPage() {
  const navigate = useNavigate();
  const [selectedUserType, setSelectedUserType] = useState<UserTypeListItem>();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const {data: originalUserTypes} = useGetUserTypes();
  const {createUser, error: createUserError} = useCreateUser();

  const {
    data: defaultUserType,
    loading: isUserTypeRequestLoading,
    error: userTypeRequestError,
  } = useGetUserType(selectedUserType?.id);

  const userTypes: UserTypeListItem[] = useMemo(() => {
    if (!originalUserTypes?.schemas) {
      return [];
    }

    if (originalUserTypes.schemas.length > 0 && !selectedUserType) {
      setSelectedUserType(originalUserTypes.schemas[0]);
    }

    return originalUserTypes?.schemas;
  }, [originalUserTypes, selectedUserType]);

  const {
    control,
    handleSubmit,
    setValue,
    formState: {errors},
  } = useForm<CreateUserFormData>({
    defaultValues: {
      userType: '',
    },
  });

  // Set default schema when schemas are loaded
  useEffect(() => {
    if (selectedUserType) {
      setValue('userType', selectedUserType.name);
    }
  }, [selectedUserType, setValue]);

  const onSubmit = async (data: CreateUserFormData) => {
    try {
      setIsSubmitting(true);

      // Extract user type from form data (userType already contains the name)
      const {userType, ...attributes} = data;

      const trimmedOuId = selectedUserType?.ouId?.trim();
      const organizationUnitId = trimmedOuId === '' ? undefined : trimmedOuId;

      if (!organizationUnitId) {
        throw new Error('Organization unit ID is missing for the selected user type');
      }

      // Prepare the request body according to the API spec
      const requestBody = {
        organizationUnit: organizationUnitId,
        type: userType,
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

  const handleBack = async () => {
    await navigate('/users');
  };

  const handleCreateUserType = () => {
    // TODO: Implement navigation to create user type page
    // eslint-disable-next-line no-console
    console.log('Navigate to create user type page');
  };

  return (
    <Box>
      <Button
        onClick={() => {
          handleBack().catch(() => {
            // Handle navigation error
          });
        }}
        variant="text"
        sx={{mb: 2}}
        aria-label="Go back"
        startIcon={<ArrowLeft size={16} />}
      >
        Back
      </Button>

      <Stack direction="row" alignItems="flex-start" mb={4} gap={2}>
        <Box>
          <Typography variant="h1" gutterBottom>
            Create User
          </Typography>
          <Typography variant="subtitle1" color="text.secondary">
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
            <FormLabel htmlFor="userType" sx={{mb: 1, display: 'block'}}>
              User Type
            </FormLabel>
            <Stack direction="row" spacing={2} alignItems="flex-start">
              <FormControl sx={{flexGrow: 1}} error={!!errors.userType}>
                <Controller
                  name="userType"
                  control={control}
                  rules={{
                    required: 'User type is required',
                  }}
                  render={({field}) => (
                    <Select
                      {...field}
                      id="userType"
                      value={field.value ?? selectedUserType?.name}
                      onChange={(e) => {
                        field.onChange(e);
                        const userType = userTypes.find((s) => s.name === e.target.value);
                        setSelectedUserType(userType);
                      }}
                      required
                      error={!!errors.userType}
                      displayEmpty
                    >
                      {userTypes.length === 0 ? (
                        <MenuItem value="" disabled>
                          Loading user types...
                        </MenuItem>
                      ) : (
                        userTypes.map((userType) => (
                          <MenuItem key={userType.name} value={userType.name}>
                            {userType.name}
                          </MenuItem>
                        ))
                      )}
                    </Select>
                  )}
                />
                {errors.userType && (
                  <Typography variant="caption" color="error" sx={{mt: 0.5, ml: 1.75}}>
                    {errors.userType.message}
                  </Typography>
                )}
              </FormControl>
              <Button variant="outlined" startIcon={<Plus size={16} />} onClick={handleCreateUserType}>
                Create
              </Button>
            </Stack>
          </Box>

          {/* Dynamic Schema Fields */}
          {isUserTypeRequestLoading && (
            <Box sx={{textAlign: 'center', py: 4}}>
              <Typography variant="body2" color="text.secondary">
                Loading user type fields...
              </Typography>
            </Box>
          )}

          {userTypeRequestError && (
            <Box sx={{textAlign: 'center', py: 4}}>
              <Typography variant="body2" color="error">
                Error loading user type: {userTypeRequestError.message}
              </Typography>
            </Box>
          )}

          {defaultUserType?.schema &&
            Object.entries(defaultUserType.schema).map(([fieldName, fieldDef]) =>
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
