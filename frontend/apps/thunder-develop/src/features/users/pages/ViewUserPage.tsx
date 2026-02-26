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

import {Link, useNavigate, useParams} from 'react-router';
import {useForm} from 'react-hook-form';
import {useState, useEffect, useMemo} from 'react';
import {
  Box,
  Stack,
  Typography,
  Button,
  Paper,
  Divider,
  CircularProgress,
  Alert,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  DialogActions,
  PageContent,
  PageTitle,
} from '@wso2/oxygen-ui';
import {ArrowLeft, Edit, Save, X, Trash2} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import {useLogger} from '@thunder/logger/react';
import useGetUser from '../api/useGetUser';
import useGetUserSchemas from '../api/useGetUserSchemas';
import useGetUserSchema from '../api/useGetUserSchema';
import useUpdateUser from '../api/useUpdateUser';
import useDeleteUser from '../api/useDeleteUser';
import renderSchemaField from '../utils/renderSchemaField';

type UpdateUserFormData = Record<string, string | number | boolean>;

export default function ViewUserPage() {
  const navigate = useNavigate();
  const {t} = useTranslation();
  const logger = useLogger('ViewUserPage');
  const {userId} = useParams<{userId: string}>();
  const [isEditMode, setIsEditMode] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

  const {data: user, isLoading: isUserLoading, error: userError} = useGetUser(userId);
  const updateUserMutation = useUpdateUser();
  const deleteUserMutation = useDeleteUser();

  // Get all schemas to find the schema ID from the schema name
  const {data: userSchemas} = useGetUserSchemas();

  // Find the schema ID based on the user's type (which is the schema name)
  const matchedSchema = useMemo(() => {
    if (!user?.type || !userSchemas?.schemas) {
      return undefined;
    }

    return userSchemas.schemas.find((s) => s.name === user.type);
  }, [user?.type, userSchemas?.schemas]);

  const schemaId = matchedSchema?.id;
  const trimmedOuId = matchedSchema?.ouId?.trim();
  const schemaOuId = trimmedOuId === '' ? undefined : trimmedOuId;

  const {data: userSchema, isLoading: isSchemaLoading, error: schemaError} = useGetUserSchema(schemaId);

  const {
    control,
    handleSubmit,
    setValue,
    formState: {errors},
  } = useForm<UpdateUserFormData>({
    defaultValues: {},
  });

  // Populate form with user data when user data is loaded
  useEffect(() => {
    if (user?.attributes && userSchema?.schema) {
      Object.entries(user.attributes).forEach(([key, value]) => {
        setValue(key, value as string | number | boolean);
      });
    }
  }, [user, userSchema, setValue]);

  const onSubmit = async (data: UpdateUserFormData) => {
    const organizationUnitId = schemaOuId ?? user?.organizationUnit;

    if (!userId || !organizationUnitId || !user?.type) return;

    try {
      setIsSubmitting(true);

      const requestBody = {
        organizationUnit: organizationUnitId,
        type: user.type,
        attributes: data,
      };

      await updateUserMutation.mutateAsync({userId, data: requestBody});

      // Exit edit mode
      setIsEditMode(false);
    } catch (err) {
      // Error is already handled in the hook and displayed in the UI
      // Keep the form in edit mode so the user can correct the error
      logger.error('Failed to update user', {error: err});
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCancel = () => {
    setIsEditMode(false);
    updateUserMutation.reset();
    // Reset form to original values
    if (user?.attributes && userSchema?.schema) {
      Object.entries(user.attributes).forEach(([key, value]) => {
        setValue(key, value as string | number | boolean);
      });
    }
  };

  const handleBack = async () => {
    await navigate('/users');
  };

  const handleDeleteClick = () => {
    setDeleteDialogOpen(true);
  };

  const handleDeleteCancel = () => {
    setDeleteDialogOpen(false);
  };

  const handleDeleteConfirm = async () => {
    if (!userId) return;

    try {
      await deleteUserMutation.mutateAsync(userId);
      setDeleteDialogOpen(false);
      // Navigate back to users list after successful deletion
      await navigate('/users');
    } catch (err) {
      // Error is already handled in the hook
      logger.error('Failed to delete user', {error: err});
      setDeleteDialogOpen(false);
    }
  };

  // Loading state
  if (isUserLoading || isSchemaLoading) {
    return (
      <Box sx={{display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '400px'}}>
        <CircularProgress />
      </Box>
    );
  }

  // Error state
  if (userError ?? schemaError) {
    return (
      <PageContent>
        <Alert severity="error" sx={{mb: 2}}>
          {userError?.message ?? schemaError?.message ?? 'Failed to load user information'}
        </Alert>
        <Button
          onClick={() => {
            handleBack().catch(() => {
              // Handle navigation error
            });
          }}
          startIcon={<ArrowLeft size={16} />}
        >
          Back to Users
        </Button>
      </PageContent>
    );
  }

  // No user found
  if (!user) {
    return (
      <PageContent>
        <Alert severity="warning" sx={{mb: 2}}>
          User not found
        </Alert>
        <Button
          onClick={() => {
            handleBack().catch(() => {
              // Handle navigation error
            });
          }}
          startIcon={<ArrowLeft size={16} />}
        >
          Back to Users
        </Button>
      </PageContent>
    );
  }

  return (
    <PageContent>
      {/* Header */}
      <PageTitle>
        <PageTitle.BackButton component={<Link to="/users" />} />
        <PageTitle.Header>{t('users:manageUser.title')}</PageTitle.Header>
        <PageTitle.SubHeader>{t('users:manageUser.subtitle')}</PageTitle.SubHeader>
        <PageTitle.Actions>
          {!isEditMode && (
            <>
              <Button variant="outlined" color="error" startIcon={<Trash2 size={16} />} onClick={handleDeleteClick}>
                Delete
              </Button>
              <Button variant="contained" startIcon={<Edit size={16} />} onClick={() => setIsEditMode(true)}>
                Edit
              </Button>
            </>
          )}
        </PageTitle.Actions>
      </PageTitle>

      <Paper sx={{p: 4}}>
        {/* User Basic Information */}
        <Box sx={{mb: 3}}>
          <Typography variant="h6" gutterBottom>
            Basic Information
          </Typography>
          <Divider sx={{mb: 2}} />
          <Stack spacing={2}>
            <Box>
              <Typography variant="caption" color="text.secondary">
                User ID
              </Typography>
              <Typography variant="body1">{user.id}</Typography>
            </Box>
            <Box>
              <Typography variant="caption" color="text.secondary">
                Organization Unit
              </Typography>
              <Typography variant="body1">{user.organizationUnit}</Typography>
            </Box>
            <Box>
              <Typography variant="caption" color="text.secondary">
                User Type
              </Typography>
              <Typography variant="body1">{user.type}</Typography>
            </Box>
          </Stack>
        </Box>

        <Divider sx={{my: 3}} />

        {/* User Attributes - View or Edit Mode */}
        <Box>
          <Typography variant="h6" gutterBottom>
            User Attributes
          </Typography>
          <Divider sx={{mb: 2}} />

          {!isEditMode ? (
            // View Mode - Display attributes as read-only
            <Stack spacing={2}>
              {user.attributes && Object.keys(user.attributes).length > 0 ? (
                Object.entries(user.attributes).map(([key, value]) => {
                  let displayValue: string;
                  if (value === null || value === undefined) {
                    displayValue = '-';
                  } else if (typeof value === 'boolean') {
                    displayValue = value ? 'Yes' : 'No';
                  } else if (Array.isArray(value)) {
                    displayValue = value.join(', ');
                  } else if (typeof value === 'object') {
                    displayValue = JSON.stringify(value);
                  } else if (typeof value === 'string' || typeof value === 'number') {
                    displayValue = String(value);
                  } else {
                    displayValue = '-';
                  }

                  return (
                    <Box key={key}>
                      <Typography variant="caption" color="text.secondary">
                        {key}
                      </Typography>
                      <Typography variant="body1">{displayValue}</Typography>
                    </Box>
                  );
                })
              ) : (
                <Typography variant="body2" color="text.secondary">
                  No attributes available
                </Typography>
              )}
            </Stack>
          ) : (
            // Edit Mode - Display form fields
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
              {/* Dynamic Schema Fields */}
              {userSchema?.schema ? (
                Object.entries(userSchema.schema)
                  .filter(([, fieldDef]) => !((fieldDef.type === 'string' || fieldDef.type === 'number') && fieldDef.credential))
                  .map(([fieldName, fieldDef]) => renderSchemaField(fieldName, fieldDef, control, errors))
              ) : (
                <Typography variant="body2" color="text.secondary">
                  No schema available for editing
                </Typography>
              )}

              {/* Update User Error Display */}
              {updateUserMutation.error && (
                <Alert severity="error" sx={{mt: 2}}>
                  <Typography variant="body2" sx={{fontWeight: 'bold', mb: 0.5}}>
                    {updateUserMutation.error.message}
                  </Typography>
                </Alert>
              )}

              {/* Form Actions */}
              <Stack direction="row" spacing={2} justifyContent="flex-end" sx={{mt: 2}}>
                <Button variant="outlined" onClick={handleCancel} disabled={isSubmitting} startIcon={<X size={16} />}>
                  Cancel
                </Button>
                <Button
                  type="submit"
                  variant="contained"
                  startIcon={isSubmitting ? null : <Save size={16} />}
                  disabled={isSubmitting}
                >
                  {isSubmitting ? 'Saving...' : 'Save Changes'}
                </Button>
              </Stack>
            </Box>
          )}
        </Box>
      </Paper>

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onClose={handleDeleteCancel}>
        <DialogTitle>Delete User</DialogTitle>
        <DialogContent>
          <DialogContentText>
            Are you sure you want to delete this user? This action cannot be undone.
          </DialogContentText>
          {deleteUserMutation.error && (
            <Alert severity="error" sx={{mt: 2}}>
              <Typography variant="body2" sx={{fontWeight: 'bold'}}>
                {deleteUserMutation.error.message}
              </Typography>
            </Alert>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={handleDeleteCancel} disabled={deleteUserMutation.isPending}>
            Cancel
          </Button>
          <Button
            onClick={() => {
              handleDeleteConfirm().catch(() => {
                // Handle error
              });
            }}
            color="error"
            variant="contained"
            disabled={deleteUserMutation.isPending}
          >
            {deleteUserMutation.isPending ? 'Deleting...' : 'Delete'}
          </Button>
        </DialogActions>
      </Dialog>
    </PageContent>
  );
}
