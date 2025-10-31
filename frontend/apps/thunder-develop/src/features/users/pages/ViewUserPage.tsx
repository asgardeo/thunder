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

import {useNavigate, useParams} from 'react-router';
import {useForm} from 'react-hook-form';
import {useState, useEffect, useMemo} from 'react';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Paper from '@mui/material/Paper';
import Divider from '@mui/material/Divider';
import CircularProgress from '@mui/material/CircularProgress';
import Alert from '@mui/material/Alert';
import Dialog from '@mui/material/Dialog';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogActions from '@mui/material/DialogActions';
import {ArrowLeft, Edit, Save, X, Trash2} from 'lucide-react';
import useGetUser from '../api/useGetUser';
import useGetUserSchemas from '../api/useGetUserSchemas';
import useGetUserSchema from '../api/useGetUserSchema';
import useUpdateUser from '../api/useUpdateUser';
import useDeleteUser from '../api/useDeleteUser';
import renderSchemaField from '../utils/renderSchemaField';

type UpdateUserFormData = Record<string, string | number | boolean>;

export default function ViewUserPage() {
  const navigate = useNavigate();
  const {userId} = useParams<{userId: string}>();
  const [isEditMode, setIsEditMode] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

  const {data: user, loading: isUserLoading, error: userError, refetch: refetchUser} = useGetUser(userId);
  const {updateUser, error: updateUserError, reset: resetUpdateError} = useUpdateUser();
  const {deleteUser, loading: isDeleting, error: deleteUserError} = useDeleteUser();

  // Get all schemas to find the schema ID from the schema name
  const {data: userSchemas} = useGetUserSchemas();

  // Find the schema ID based on the user's type (which is the schema name)
  const schemaId = useMemo(() => {
    if (!user?.type || !userSchemas?.schemas) return undefined;
    const schema = userSchemas.schemas.find((s) => s.name === user.type);
    return schema?.id;
  }, [user?.type, userSchemas?.schemas]);

  const {data: userSchema, loading: isSchemaLoading, error: schemaError} = useGetUserSchema(schemaId);

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
    if (!userId || !user?.organizationUnit || !user?.type) return;

    try {
      setIsSubmitting(true);

      const requestBody = {
        organizationUnit: user.organizationUnit,
        type: user.type,
        attributes: data,
      };

      await updateUser(userId, requestBody);

      // Refetch user data to show updated values
      await refetchUser();

      // Exit edit mode
      setIsEditMode(false);
    } catch (error) {
      // Error is already handled in the hook and displayed in the UI
      // Keep the form in edit mode so the user can correct the error
      // eslint-disable-next-line no-console
      console.error('Failed to update user:', error);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCancel = () => {
    setIsEditMode(false);
    resetUpdateError();
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
      await deleteUser(userId);
      setDeleteDialogOpen(false);
      // Navigate back to users list after successful deletion
      await navigate('/users');
    } catch (error) {
      // Error is already handled in the hook
      // eslint-disable-next-line no-console
      console.error('Failed to delete user:', error);
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
      <Box sx={{maxWidth: 800, mx: 'auto', px: 2, pt: 6}}>
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
      </Box>
    );
  }

  // No user found
  if (!user) {
    return (
      <Box sx={{maxWidth: 800, mx: 'auto', px: 2, pt: 6}}>
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
      </Box>
    );
  }

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

      <Stack direction="row" alignItems="flex-start" justifyContent="space-between" mb={4} gap={2}>
        <Box>
          <Typography variant="h4" component="h1" gutterBottom>
            User Profile
          </Typography>
          <Typography variant="body2" color="text.secondary">
            View and manage user information
          </Typography>
        </Box>
        {!isEditMode && (
          <Stack direction="row" spacing={2}>
            <Button variant="outlined" color="error" startIcon={<Trash2 size={16} />} onClick={handleDeleteClick}>
              Delete
            </Button>
            <Button variant="contained" startIcon={<Edit size={16} />} onClick={() => setIsEditMode(true)}>
              Edit
            </Button>
          </Stack>
        )}
      </Stack>

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
                  .filter(([fieldName]) => fieldName !== 'password')
                  .map(([fieldName, fieldDef]) => renderSchemaField(fieldName, fieldDef, control, errors))
              ) : (
                <Typography variant="body2" color="text.secondary">
                  No schema available for editing
                </Typography>
              )}

              {/* Update User Error Display */}
              {updateUserError && (
                <Alert severity="error" sx={{mt: 2}}>
                  <Typography variant="body2" sx={{fontWeight: 'bold', mb: 0.5}}>
                    {updateUserError.message}
                  </Typography>
                  {updateUserError.description && (
                    <Typography variant="body2">{updateUserError.description}</Typography>
                  )}
                  {updateUserError.code === 'USR-1014' && (
                    <Typography variant="body2" sx={{mt: 1, fontStyle: 'italic'}}>
                      Please check the unique fields (e.g., email, username) and ensure they are not already in use by
                      another user.
                    </Typography>
                  )}
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
          {deleteUserError && (
            <Alert severity="error" sx={{mt: 2}}>
              <Typography variant="body2" sx={{fontWeight: 'bold'}}>
                {deleteUserError.message}
              </Typography>
              {deleteUserError.description && <Typography variant="caption">{deleteUserError.description}</Typography>}
            </Alert>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={handleDeleteCancel} disabled={isDeleting}>
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
            disabled={isDeleting}
          >
            {isDeleting ? 'Deleting...' : 'Delete'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
