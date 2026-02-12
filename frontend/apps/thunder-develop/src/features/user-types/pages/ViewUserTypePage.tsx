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
import {useState, useEffect, useMemo} from 'react';
import {useTranslation} from 'react-i18next';
import {useLogger} from '@thunder/logger/react';
import {
  Box,
  Stack,
  Typography,
  Button,
  Paper,
  Divider,
  CircularProgress,
  Alert,
  Snackbar,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  DialogActions,
  FormControl,
  FormLabel,
  Select,
  MenuItem,
  TextField,
  Checkbox,
  FormControlLabel,
  Chip,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
} from '@wso2/oxygen-ui';
import {ArrowLeft, Edit, Save, X, Trash2, Check} from '@wso2/oxygen-ui-icons-react';
import useGetUserType from '../api/useGetUserType';
import useUpdateUserType from '../api/useUpdateUserType';
import useDeleteUserType from '../api/useDeleteUserType';
import useGetOrganizationUnits from '../../organization-units/api/useGetOrganizationUnits';
import type {PropertyDefinition, UserSchemaDefinition, PropertyType, SchemaPropertyInput} from '../types/user-types';

export default function ViewUserTypePage() {
  const navigate = useNavigate();
  const {t} = useTranslation();
  const logger = useLogger('ViewUserTypePage');
  const {id} = useParams<{id: string}>();
  const [isEditMode, setIsEditMode] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

  const {data: userType, loading: isUserTypeLoading, error: userTypeError, refetch} = useGetUserType(id);
  const {updateUserType, error: updateUserTypeError, reset: resetUpdateError} = useUpdateUserType();
  const {deleteUserType, loading: isDeleting, error: deleteUserTypeError} = useDeleteUserType();
  const {
    data: organizationUnitsResponse,
    isLoading: organizationUnitsLoading,
    error: organizationUnitsError,
  } = useGetOrganizationUnits();

  const [name, setName] = useState('');
  const [ouId, setOuId] = useState('');
  const [allowSelfRegistration, setAllowSelfRegistration] = useState(false);
  const [properties, setProperties] = useState<SchemaPropertyInput[]>([]);
  const [enumInput, setEnumInput] = useState<Record<string, string>>({});
  const [validationError, setValidationError] = useState<string | null>(null);
  const [snackbarOpen, setSnackbarOpen] = useState(false);
  const organizationUnits = useMemo(
    () => organizationUnitsResponse?.organizationUnits ?? [],
    [organizationUnitsResponse],
  );
  const selectedOrganizationUnit = useMemo(
    () => organizationUnits.find((unit) => unit.id === ouId),
    [organizationUnits, ouId],
  );

  const convertSchemaToProperties = (schema: UserSchemaDefinition) => {
    const props: SchemaPropertyInput[] = Object.entries(schema).map(([key, value], index) => ({
      id: `${index}`,
      name: key,
      type: value.type,
      required: value.required ?? false,
      unique: 'unique' in value ? (value.unique ?? false) : false,
      enum: 'enum' in value ? (value.enum ?? []) : [],
      regex: 'regex' in value ? (value.regex ?? '') : '',
    }));
    setProperties(props);
  };

  useEffect(() => {
    if (userType) {
      setName(userType.name);
      setOuId(userType.ouId);
      setAllowSelfRegistration(userType.allowSelfRegistration ?? false);
      convertSchemaToProperties(userType.schema);
    }
  }, [userType]);

  const handleEdit = () => {
    setIsEditMode(true);
  };

  const handleCancel = () => {
    setIsEditMode(false);
    resetUpdateError();
    if (userType) {
      setName(userType.name);
      setOuId(userType.ouId);
      setAllowSelfRegistration(userType.allowSelfRegistration ?? false);
      convertSchemaToProperties(userType.schema);
    }
    setValidationError(null);
    setSnackbarOpen(false);
  };

  const handleBack = async () => {
    await navigate('/user-types');
  };

  const handleCloseSnackbar = () => {
    setSnackbarOpen(false);
  };

  const handlePropertyChange = <K extends keyof SchemaPropertyInput>(
    propertyId: string,
    field: K,
    value: SchemaPropertyInput[K],
  ) => {
    setProperties(
      properties.map((prop) =>
        prop.id === propertyId
          ? {
              ...prop,
              [field]: value,
              ...(field === 'type' && {
                enum: [],
                regex: '',
                unique:
                  (value as PropertyType) === 'string' || (value as PropertyType) === 'number' ? prop.unique : false,
              }),
            }
          : prop,
      ),
    );
  };

  const handleAddEnumValue = (propertyId: string) => {
    const inputValue = enumInput[propertyId]?.trim();
    if (!inputValue) return;

    setProperties(
      properties.map((prop) => (prop.id === propertyId ? {...prop, enum: [...prop.enum, inputValue]} : prop)),
    );

    setEnumInput({...enumInput, [propertyId]: ''});
  };

  const handleRemoveEnumValue = (propertyId: string, enumValue: string) => {
    setProperties(
      properties.map((prop) =>
        prop.id === propertyId ? {...prop, enum: prop.enum.filter((val) => val !== enumValue)} : prop,
      ),
    );
  };

  const handleSave = async () => {
    if (!id) return;

    setValidationError(null);
    setSnackbarOpen(false);
    const trimmedOuId = ouId.trim();
    if (!trimmedOuId) {
      setValidationError(t('userTypes:validationErrors.ouIdRequired'));
      setSnackbarOpen(true);
      return;
    }

    try {
      setIsSubmitting(true);

      const validProperties = properties.filter((prop) => prop.name.trim());

      // Convert properties to schema definition
      const schema: UserSchemaDefinition = {};
      validProperties.forEach((prop) => {
        // Convert UI type to actual PropertyType (enum -> string)
        const actualType: PropertyType = prop.type === 'enum' ? 'string' : prop.type;

        const propDef: Partial<PropertyDefinition> = {
          type: actualType,
          required: prop.required,
        };

        if (prop.type === 'string' || prop.type === 'number' || prop.type === 'enum') {
          if (prop.unique) {
            (propDef as {unique?: boolean}).unique = true;
          }
        }

        if (prop.type === 'string' || prop.type === 'enum') {
          if (prop.enum.length > 0) {
            (propDef as {enum?: string[]}).enum = prop.enum;
          }
          if (prop.regex.trim()) {
            (propDef as {regex?: string}).regex = prop.regex;
          }
        }

        if (prop.type === 'array') {
          (propDef as {items?: {type: string}}).items = {type: 'string'};
        } else if (prop.type === 'object') {
          (propDef as {properties?: Record<string, PropertyDefinition>}).properties = {};
        }

        schema[prop.name.trim()] = propDef as PropertyDefinition;
      });

      await updateUserType(id, {
        name: name.trim(),
        ouId: trimmedOuId,
        allowSelfRegistration,
        schema,
      });

      // Refetch user type data to show updated values
      await refetch(id);

      // Exit edit mode
      setIsEditMode(false);
    } catch (error) {
      // Error is already handled in the hook and displayed in the UI
      // Keep the form in edit mode so the user can correct the error
      // eslint-disable-next-line no-console
      console.error('Failed to update user type:', error);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDeleteClick = () => {
    setDeleteDialogOpen(true);
  };

  const handleDeleteCancel = () => {
    setDeleteDialogOpen(false);
  };

  const handleDeleteConfirm = async () => {
    if (!id) return;

    try {
      await deleteUserType(id);
      setDeleteDialogOpen(false);
      // Navigate back to user types list after successful deletion
      await navigate('/user-types');
    } catch (error) {
      logger.error('Failed to delete user type', {error, userTypeId: id});
      setDeleteDialogOpen(false);
    }
  };

  // Loading state
  if (isUserTypeLoading) {
    return (
      <Box sx={{display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '400px'}}>
        <CircularProgress />
      </Box>
    );
  }

  // Error state
  if (userTypeError) {
    return (
      <Box sx={{maxWidth: 1000, mx: 'auto', pt: 6}}>
        <Alert severity="error" sx={{mb: 2}}>
          {userTypeError.message ?? 'Failed to load user type information'}
        </Alert>
        <Button
          onClick={() => {
            handleBack().catch(() => {
              // Handle navigation error
            });
          }}
          startIcon={<ArrowLeft size={16} />}
        >
          Back to User Types
        </Button>
      </Box>
    );
  }

  // No user type found
  if (!userType) {
    return (
      <Box sx={{maxWidth: 1000, mx: 'auto', pt: 6}}>
        <Alert severity="warning" sx={{mb: 2}}>
          User type not found
        </Alert>
        <Button
          onClick={() => {
            handleBack().catch(() => {
              // Handle navigation error
            });
          }}
          startIcon={<ArrowLeft size={16} />}
        >
          Back to User Types
        </Button>
      </Box>
    );
  }

  return (
    <Box>
      <Button
        onClick={() => {
          handleBack().catch(() => {
            // Handle navigation error
          });
        }}
        variant="text"
        sx={{mb: 3}}
        aria-label="Go back"
        startIcon={<ArrowLeft size={16} />}
      >
        Back
      </Button>

      <Stack direction="row" alignItems="flex-start" justifyContent="space-between" mb={4} gap={2}>
        <Box>
          <Typography variant="h1" gutterBottom>
            User Type Details
          </Typography>
          <Typography variant="subtitle1" color="text.secondary">
            View and manage user type schema
          </Typography>
        </Box>
        {!isEditMode && (
          <Stack direction="row" spacing={2}>
            <Button variant="outlined" color="error" startIcon={<Trash2 size={16} />} onClick={handleDeleteClick}>
              Delete
            </Button>
            <Button variant="contained" startIcon={<Edit size={16} />} onClick={handleEdit}>
              Edit
            </Button>
          </Stack>
        )}
      </Stack>

      <Paper sx={{p: 4}}>
        {/* Basic Information */}
        <Box sx={{mb: 4}}>
          <Typography variant="h6" gutterBottom>
            Basic Information
          </Typography>
          <Divider sx={{mb: 3}} />
          <Stack spacing={2}>
            <Box>
              <Typography variant="caption" color="text.secondary" sx={{display: 'block', mb: 0.5}}>
                ID
              </Typography>
              <Typography variant="body1" sx={{fontFamily: 'monospace', fontSize: '0.875rem'}}>
                {userType.id}
              </Typography>
            </Box>
            <Box>
              <Typography variant="caption" color="text.secondary" sx={{display: 'block', mb: 0.5}}>
                Name
              </Typography>
              {!isEditMode ? (
                <Typography variant="body1" sx={{fontWeight: 500}}>
                  {userType.name}
                </Typography>
              ) : (
                <TextField
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="User type name"
                  size="small"
                  fullWidth
                />
              )}
            </Box>
            <Box>
              <Typography variant="caption" color="text.secondary" sx={{display: 'block', mb: 0.5}}>
                {t('userTypes:organizationUnit')}
              </Typography>
              {!isEditMode ? (
                <Typography variant="body1" sx={{fontWeight: 500}}>
                  {selectedOrganizationUnit ? selectedOrganizationUnit.name : userType.ouId}
                </Typography>
              ) : (
                <Select
                  value={ouId}
                  onChange={(event) => setOuId(event.target.value ?? '')}
                  size="small"
                  fullWidth
                  displayEmpty
                  aria-label={t('userTypes:organizationUnit')}
                  renderValue={(selected) => {
                    const value = typeof selected === 'string' ? selected : '';
                    if (!value) {
                      return t('userTypes:ouSelectPlaceholder');
                    }
                    const currentUnit = organizationUnits.find((unit) => unit.id === value);
                    return currentUnit ? currentUnit.name : value;
                  }}
                >
                  {organizationUnitsLoading && (
                    <MenuItem value="" disabled>
                      {t('common:status.loading')}
                    </MenuItem>
                  )}

                  {!organizationUnitsLoading && organizationUnitsError && (
                    <MenuItem value="" disabled>
                      {organizationUnitsError.message}
                    </MenuItem>
                  )}

                  {!organizationUnitsLoading && !organizationUnitsError && organizationUnits.length === 0 && (
                    <MenuItem value="" disabled>
                      {t('userTypes:noOrganizationUnits')}
                    </MenuItem>
                  )}

                  {organizationUnits.map((unit) => (
                    <MenuItem key={unit.id} value={unit.id}>
                      <Typography variant="body2" sx={{fontWeight: 500}}>
                        {unit.name}
                      </Typography>
                    </MenuItem>
                  ))}
                </Select>
              )}
            </Box>
            <Box>
              <Typography variant="caption" color="text.secondary" sx={{display: 'block', mb: 0.5}}>
                {t('userTypes:allowSelfRegistration')}
              </Typography>
              {!isEditMode ? (
                <Chip
                  label={userType.allowSelfRegistration ? t('common:status.enabled') : t('common:status.disabled')}
                  color={userType.allowSelfRegistration ? 'success' : 'default'}
                  size="small"
                />
              ) : (
                <FormControlLabel
                  control={
                    <Checkbox
                      checked={allowSelfRegistration}
                      onChange={(e) => setAllowSelfRegistration(e.target.checked)}
                      size="small"
                    />
                  }
                  label={t('userTypes:allowSelfRegistration')}
                />
              )}
            </Box>
          </Stack>
        </Box>

        <Divider sx={{my: 4}} />

        {/* Schema Properties */}
        <Box>
          <Typography variant="h6" gutterBottom>
            Schema Properties
          </Typography>
          <Divider sx={{mb: 3}} />

          {!isEditMode ? (
            // View Mode - Display properties in a table
            <TableContainer>
              <Table sx={{'& .MuiTableCell-root': {py: 2}}}>
                <TableHead>
                  <TableRow>
                    <TableCell sx={{fontWeight: 600}}>Property Name</TableCell>
                    <TableCell sx={{fontWeight: 600}}>Type</TableCell>
                    <TableCell sx={{fontWeight: 600}}>Required</TableCell>
                    <TableCell sx={{fontWeight: 600}}>Unique</TableCell>
                    <TableCell sx={{fontWeight: 600}}>Constraints</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {Object.entries(userType.schema).map(([key, value]) => (
                    <TableRow key={key} sx={{'&:hover': {bgcolor: 'action.hover'}}}>
                      <TableCell>
                        <Typography variant="body2" sx={{fontWeight: 500}}>
                          {key}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Typography
                          variant="body2"
                          sx={{
                            fontFamily: 'monospace',
                            color: 'primary.main',
                            fontSize: '0.875rem',
                          }}
                        >
                          {value.type}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        {value.required ? (
                          <Check size={18} color="green" />
                        ) : (
                          <Typography variant="body2" color="text.secondary">
                            -
                          </Typography>
                        )}
                      </TableCell>
                      <TableCell>
                        {'unique' in value && value.unique ? (
                          <Check size={18} color="orange" />
                        ) : (
                          <Typography variant="body2" color="text.secondary">
                            -
                          </Typography>
                        )}
                      </TableCell>
                      <TableCell>
                        <Stack spacing={0.5}>
                          {'enum' in value && value.enum && value.enum.length > 0 && (
                            <Typography variant="body2" sx={{fontSize: '0.875rem'}}>
                              <Box component="span" sx={{color: 'text.secondary', mr: 0.5}}>
                                Allowed:
                              </Box>
                              {value.enum.join(', ')}
                            </Typography>
                          )}
                          {'regex' in value && value.regex && (
                            <Typography variant="body2" sx={{fontSize: '0.875rem'}}>
                              <Box component="span" sx={{color: 'text.secondary', mr: 0.5}}>
                                Pattern:
                              </Box>
                              <Box
                                component="code"
                                sx={{
                                  fontFamily: 'monospace',
                                  fontSize: '0.75rem',
                                  bgcolor: 'grey.100',
                                  px: 0.5,
                                  py: 0.25,
                                  borderRadius: 0.5,
                                }}
                              >
                                {value.regex}
                              </Box>
                            </Typography>
                          )}
                          {!('enum' in value || 'regex' in value) && (
                            <Typography variant="body2" color="text.secondary">
                              -
                            </Typography>
                          )}
                        </Stack>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          ) : (
            // Edit Mode - Display form fields
            <Box>
              {properties.map((property) => (
                <Paper key={property.id} variant="outlined" sx={{p: 3, mb: 2}}>
                  <Stack spacing={2}>
                    <Box sx={{display: 'grid', gridTemplateColumns: {xs: '1fr', md: '1fr 1fr'}, gap: 2}}>
                      <FormControl fullWidth>
                        <FormLabel>Property Name</FormLabel>
                        <TextField
                          value={property.name}
                          onChange={(e) => handlePropertyChange(property.id, 'name', e.target.value)}
                          placeholder="e.g., email, age, address"
                          size="small"
                          disabled
                        />
                      </FormControl>

                      <FormControl fullWidth>
                        <FormLabel>Type</FormLabel>
                        <Select
                          value={property.type}
                          onChange={(e) => handlePropertyChange(property.id, 'type', e.target.value as PropertyType)}
                          size="small"
                        >
                          <MenuItem value="string">String</MenuItem>
                          <MenuItem value="number">Number</MenuItem>
                          <MenuItem value="boolean">Boolean</MenuItem>
                          <MenuItem value="array">Array</MenuItem>
                          <MenuItem value="object">Object</MenuItem>
                        </Select>
                      </FormControl>
                    </Box>

                    <Stack direction="row" spacing={2}>
                      <FormControlLabel
                        control={
                          <Checkbox
                            checked={property.required}
                            onChange={(e) => handlePropertyChange(property.id, 'required', e.target.checked)}
                          />
                        }
                        label="Required"
                      />
                      {(property.type === 'string' || property.type === 'number') && (
                        <FormControlLabel
                          control={
                            <Checkbox
                              checked={property.unique}
                              onChange={(e) => handlePropertyChange(property.id, 'unique', e.target.checked)}
                            />
                          }
                          label="Unique"
                        />
                      )}
                    </Stack>

                    {property.type === 'string' && (
                      <>
                        <FormControl fullWidth>
                          <FormLabel>Regular Expression Pattern (Optional)</FormLabel>
                          <TextField
                            value={property.regex}
                            onChange={(e) => handlePropertyChange(property.id, 'regex', e.target.value)}
                            placeholder="e.g., ^[a-zA-Z0-9]+$"
                            size="small"
                          />
                        </FormControl>

                        <FormControl fullWidth>
                          <FormLabel>Allowed Values (Enum) - Optional</FormLabel>
                          <Box sx={{display: 'flex', gap: 1, mb: 1}}>
                            <TextField
                              value={enumInput[property.id] ?? ''}
                              onChange={(e) => setEnumInput({...enumInput, [property.id]: e.target.value})}
                              onKeyDown={(e) => {
                                if (e.key === 'Enter') {
                                  e.preventDefault();
                                  handleAddEnumValue(property.id);
                                }
                              }}
                              placeholder="Add value and press Enter"
                              size="small"
                              fullWidth
                            />
                            <Button variant="outlined" size="small" onClick={() => handleAddEnumValue(property.id)}>
                              Add
                            </Button>
                          </Box>
                          {property.enum.length > 0 && (
                            <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                              {property.enum.map((val) => (
                                <Chip
                                  key={val}
                                  label={val}
                                  onDelete={() => handleRemoveEnumValue(property.id, val)}
                                  size="small"
                                />
                              ))}
                            </Stack>
                          )}
                        </FormControl>
                      </>
                    )}
                  </Stack>
                </Paper>
              ))}

              {/* Update Error Display */}
              {updateUserTypeError && (
                <Alert severity="error" sx={{mt: 2}}>
                  <Typography variant="body2" sx={{fontWeight: 'bold', mb: 0.5}}>
                    {updateUserTypeError.message}
                  </Typography>
                  {updateUserTypeError.description && (
                    <Typography variant="body2">{updateUserTypeError.description}</Typography>
                  )}
                </Alert>
              )}

              {/* Form Actions */}
              <Stack direction="row" spacing={2} justifyContent="flex-end" sx={{mt: 3}}>
                <Button variant="outlined" onClick={handleCancel} disabled={isSubmitting} startIcon={<X size={16} />}>
                  Cancel
                </Button>
                <Button
                  onClick={() => {
                    handleSave().catch(() => {
                      // Handle error
                    });
                  }}
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
        <DialogTitle>Delete User Type</DialogTitle>
        <DialogContent>
          <DialogContentText>
            Are you sure you want to delete this user type? This action cannot be undone and may affect existing users
            of this type.
          </DialogContentText>
          {deleteUserTypeError && (
            <Alert severity="error" sx={{mt: 2}}>
              <Typography variant="body2" sx={{fontWeight: 'bold'}}>
                {deleteUserTypeError.message}
              </Typography>
              {deleteUserTypeError.description && (
                <Typography variant="caption">{deleteUserTypeError.description}</Typography>
              )}
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

      <Snackbar
        open={snackbarOpen}
        autoHideDuration={6000}
        onClose={handleCloseSnackbar}
        anchorOrigin={{vertical: 'top', horizontal: 'right'}}
      >
        <Alert onClose={handleCloseSnackbar} severity="error" sx={{width: '100%'}}>
          {validationError}
        </Alert>
      </Snackbar>
    </Box>
  );
}
