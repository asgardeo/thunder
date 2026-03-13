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
import {useState, useEffect, useMemo, useRef} from 'react';
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
  IconButton,
  Tooltip,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  PageContent,
  PageTitle,
} from '@wso2/oxygen-ui';
import {ArrowLeft, Edit, Save, X, Trash2, Check, Plus, Info} from '@wso2/oxygen-ui-icons-react';
import {useResolveDisplayName} from '@thunder/shared-hooks';
import useGetUserType from '../api/useGetUserType';
import useUpdateUserType from '../api/useUpdateUserType';
import useDeleteUserType from '../api/useDeleteUserType';
import useGetOrganizationUnits from '../../organization-units/api/useGetOrganizationUnits';
import type {PropertyDefinition, UserSchemaDefinition, PropertyType, SchemaPropertyInput} from '../types/user-types';
import I18nTextInput from '../components/create-user-type/I18nTextInput';

export default function ViewUserTypePage() {
  const navigate = useNavigate();
  const {t} = useTranslation();
  const logger = useLogger('ViewUserTypePage');
  const {id} = useParams<{id: string}>();
  const [isEditMode, setIsEditMode] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const {resolveDisplayName} = useResolveDisplayName({handlers: {t}});

  const {data: userType, isLoading: isUserTypeLoading, error: userTypeError} = useGetUserType(id);
  const updateUserTypeMutation = useUpdateUserType();
  const deleteUserTypeMutation = useDeleteUserType();
  const {
    data: organizationUnitsResponse,
    isLoading: organizationUnitsLoading,
    error: organizationUnitsError,
  } = useGetOrganizationUnits();

  const [name, setName] = useState('');
  const [ouId, setOuId] = useState('');
  const [allowSelfRegistration, setAllowSelfRegistration] = useState(false);
  const [properties, setProperties] = useState<SchemaPropertyInput[]>([]);
  const [displayAttribute, setDisplayAttribute] = useState('');
  const [enumInput, setEnumInput] = useState<Record<string, string>>({});
  const [validationError, setValidationError] = useState<string | null>(null);
  const [snackbarOpen, setSnackbarOpen] = useState(false);
  const nextPropertyId = useRef(0);
  const organizationUnits = useMemo(
    () => organizationUnitsResponse?.organizationUnits ?? [],
    [organizationUnitsResponse],
  );
  const selectedOrganizationUnit = useMemo(
    () => organizationUnits.find((unit) => unit.id === ouId),
    [organizationUnits, ouId],
  );
  const eligibleDisplayProperties = useMemo(
    () =>
      properties.filter(
        (p) =>
          (p.type === 'string' || p.type === 'number' || p.type === 'enum') &&
          !p.credential &&
          p.name.trim().length > 0,
      ),
    [properties],
  );

  // Clear display attribute if the selected property becomes ineligible (e.g. marked as credential)
  useEffect(() => {
    if (displayAttribute) {
      const eligibleNames = eligibleDisplayProperties.map((p) => p.name.trim());
      if (!eligibleNames.includes(displayAttribute)) {
        setDisplayAttribute('');
      }
    }
  }, [eligibleDisplayProperties, displayAttribute]);

  const convertSchemaToProperties = (schema: UserSchemaDefinition) => {
    const schemaProperties: SchemaPropertyInput[] = Object.entries(schema).map(([key, value], index) => ({
      id: `${index}`,
      name: key,
      displayName: 'displayName' in value ? (value.displayName ?? '') : '',
      type: value.type === 'string' && 'enum' in value && Array.isArray(value.enum) && value.enum.length > 0 ? 'enum' : value.type,
      required: value.required ?? false,
      unique: 'unique' in value ? (value.unique ?? false) : false,
      credential: 'credential' in value ? (value.credential ?? false) : false,
      enum: 'enum' in value ? (value.enum ?? []) : [],
      regex: 'regex' in value ? (value.regex ?? '') : '',
    }));
    nextPropertyId.current = schemaProperties.length + 1;
    setProperties(schemaProperties);
  };

  useEffect(() => {
    if (userType) {
      setName(userType.name);
      setOuId(userType.ouId);
      setAllowSelfRegistration(userType.allowSelfRegistration ?? false);
      setDisplayAttribute(userType.systemAttributes?.display ?? '');
      convertSchemaToProperties(userType.schema);
    }
  }, [userType]);

  const handleEdit = () => {
    setIsEditMode(true);
  };

  const handleCancel = () => {
    setIsEditMode(false);
    updateUserTypeMutation.reset();
    if (userType) {
      setName(userType.name);
      setOuId(userType.ouId);
      setAllowSelfRegistration(userType.allowSelfRegistration ?? false);
      setDisplayAttribute(userType.systemAttributes?.display ?? '');
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

  const handleAddProperty = () => {
    const newId = String(nextPropertyId.current);
    nextPropertyId.current += 1;
    const newProperty: SchemaPropertyInput = {
      id: newId,
      name: '',
      displayName: '',
      type: 'string',
      required: false,
      unique: false,
      credential: false,
      enum: [],
      regex: '',
    };
    setProperties([...properties, newProperty]);
  };

  const handleRemoveProperty = (propertyId: string) => {
    setProperties(properties.filter((prop) => prop.id !== propertyId));
    const newEnumInput = {...enumInput};
    delete newEnumInput[propertyId];
    setEnumInput(newEnumInput);
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
                enum: (value as string) === 'enum' ? prop.enum : [],
                regex: '',
                unique:
                  (value as string) === 'string' || (value as string) === 'number' || (value as string) === 'enum'
                    ? prop.unique
                    : false,
                credential:
                  (value as string) === 'string' || (value as string) === 'number' ? prop.credential : false,
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
          ...(prop.displayName.trim() ? {displayName: prop.displayName.trim()} : {}),
        };

        if (prop.type === 'string' || prop.type === 'number' || prop.type === 'enum') {
          if (prop.unique) {
            (propDef as {unique?: boolean}).unique = true;
          }
          if (prop.credential) {
            (propDef as {credential?: boolean}).credential = true;
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

      await updateUserTypeMutation.mutateAsync({
        userTypeId: id,
        data: {
          name: name.trim(),
          ouId: trimmedOuId,
          allowSelfRegistration,
          ...(displayAttribute ? {systemAttributes: {display: displayAttribute}} : {}),
          schema,
        },
      });

      // Exit edit mode
      setIsEditMode(false);
    } catch (error) {
      // Error is already handled in the hook and displayed in the UI
      // Keep the form in edit mode so the user can correct the error
      logger.error('Failed to update user type', {error: error as Error, userTypeId: id});
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDeleteClick = () => {
    setDeleteDialogOpen(true);
  };

  const handleDeleteCancel = () => {
    setDeleteDialogOpen(false);
    deleteUserTypeMutation.reset();
  };

  const handleDeleteConfirm = async () => {
    if (!id) return;

    try {
      await deleteUserTypeMutation.mutateAsync(id);
      setDeleteDialogOpen(false);
      // Navigate back to user types list after successful deletion
      await navigate('/user-types');
    } catch (err) {
      // Keep dialog open so inline error is visible and user can retry
      logger.error('Failed to delete user type', {error: err as Error, userTypeId: id});
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
      <PageContent>
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
      </PageContent>
    );
  }

  // No user type found
  if (!userType) {
    return (
      <PageContent>
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
      </PageContent>
    );
  }

  return (
    <PageContent>
      {/* Header */}
      <PageTitle>
        <PageTitle.BackButton component={<Link to="/user-types" />} />
        <PageTitle.Header>{t('userTypes:manageUserType.title')}</PageTitle.Header>
        <PageTitle.SubHeader>{t('userTypes:manageUserType.subtitle')}</PageTitle.SubHeader>
        <PageTitle.Actions>
          {!isEditMode && (
            <>
              <Button variant="outlined" color="error" startIcon={<Trash2 size={16} />} onClick={handleDeleteClick}>
                Delete
              </Button>
              <Button variant="contained" startIcon={<Edit size={16} />} onClick={handleEdit}>
                Edit
              </Button>
            </>
          )}
        </PageTitle.Actions>
      </PageTitle>

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
            <Box>
              <Typography variant="caption" color="text.secondary" sx={{display: 'block', mb: 0.5}}>
                {t('userTypes:displayAttribute', 'Display Attribute')}
              </Typography>
              {!isEditMode ? (
                <Typography variant="body1" sx={{fontWeight: 500}}>
                  {userType.systemAttributes?.display ?? t('userTypes:noDisplayAttribute', 'Not configured')}
                </Typography>
              ) : (
                <Select
                  value={displayAttribute}
                  onChange={(event) => setDisplayAttribute(event.target.value)}
                  size="small"
                  fullWidth
                  displayEmpty
                  renderValue={(selected) => {
                    const value = typeof selected === 'string' ? selected : '';
                    if (!value) {
                      return (
                        <Typography variant="body2" color="text.secondary">
                          {t('userTypes:selectDisplayAttribute', 'Select a display attribute')}
                        </Typography>
                      );
                    }
                    const matchedProp = eligibleDisplayProperties.find((p) => p.name.trim() === value);
                    const resolved = matchedProp?.displayName ? resolveDisplayName(matchedProp.displayName) : '';
                    return resolved && resolved !== value ? `${resolved} (${value})` : value;
                  }}
                >
                  <MenuItem value="">
                    <Typography variant="body2" color="text.secondary">
                      {t('common:none', 'None')}
                    </Typography>
                  </MenuItem>
                  {eligibleDisplayProperties.map((prop) => {
                    const propName = prop.name.trim();
                    const resolved = prop.displayName ? resolveDisplayName(prop.displayName) : '';
                    return (
                      <MenuItem key={prop.id} value={propName}>
                        {resolved && resolved !== propName ? `${resolved} (${propName})` : propName}
                      </MenuItem>
                    );
                  })}
                </Select>
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
            <TableContainer component={Paper}>
              <Table sx={{'& .MuiTableCell-root': {py: 2}}}>
                <TableHead>
                  <TableRow>
                    <TableCell sx={{fontWeight: 600}}>{t('userTypes:propertyName')}</TableCell>
                    <TableCell sx={{fontWeight: 600}}>{t('userTypes:displayName', 'Display Name')}</TableCell>
                    <TableCell sx={{fontWeight: 600}}>Type</TableCell>
                    <TableCell sx={{fontWeight: 600}}>Required</TableCell>
                    <TableCell sx={{fontWeight: 600}}>Unique</TableCell>
                    <TableCell sx={{fontWeight: 600}}>{t('userTypes:credential', 'Credential')}</TableCell>
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
                        {(() => {
                          const resolved = value.displayName ? resolveDisplayName(value.displayName) : '';
                          return (
                            <Typography variant="body2" color={resolved ? 'text.primary' : 'text.secondary'}>
                              {resolved || '-'}
                            </Typography>
                          );
                        })()}
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
                        {'credential' in value && value.credential ? (
                          <Check size={18} color="red" />
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
                                  bgcolor: 'action.hover',
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
                  <Paper
                    key={property.id}
                    variant="outlined"
                    sx={{
                      position: 'relative',
                      p: 3,
                      mb: 2,
                      borderRadius: 2,
                      transition: 'border-color 0.2s',
                      '&:hover': {borderColor: 'primary.main'},
                      '&:hover .property-delete-btn': {opacity: 1},
                    }}
                  >
                    {/* Remove button - visible on hover */}
                    {properties.length > 1 && (
                      <Tooltip title={t('userTypes:removeProperty', 'Remove property')}>
                        <IconButton
                          className="property-delete-btn"
                          size="small"
                          color="error"
                          onClick={() => handleRemoveProperty(property.id)}
                          sx={{position: 'absolute', top: 8, right: 8, opacity: 0, transition: 'opacity 0.2s'}}
                        >
                          <Trash2 size={16} />
                        </IconButton>
                      </Tooltip>
                    )}

                    <Stack spacing={2}>
                      <Box sx={{display: 'grid', gridTemplateColumns: {xs: '1fr', md: '1fr 1fr'}, gap: 2}}>
                        <FormControl fullWidth>
                          <FormLabel>{t('userTypes:propertyName')}</FormLabel>
                          <TextField
                            value={property.name}
                            onChange={(e) => handlePropertyChange(property.id, 'name', e.target.value)}
                            placeholder={t('userTypes:propertyNamePlaceholder', 'e.g., email, age, address')}
                            size="small"
                          />
                        </FormControl>

                        <FormControl fullWidth>
                          <FormLabel>{t('userTypes:propertyType', 'Type')}</FormLabel>
                          <Select
                            value={property.type}
                            onChange={(e) => handlePropertyChange(property.id, 'type', e.target.value as PropertyType)}
                            size="small"
                          >
                            <MenuItem value="string">{t('userTypes:types.string', 'String')}</MenuItem>
                            <MenuItem value="number">{t('userTypes:types.number', 'Number')}</MenuItem>
                            <MenuItem value="boolean">{t('userTypes:types.boolean', 'Boolean')}</MenuItem>
                            <MenuItem value="enum">{t('userTypes:types.enum', 'Enum')}</MenuItem>
                          </Select>
                        </FormControl>
                      </Box>

                      <I18nTextInput
                        label={t('userTypes:displayName', 'Display Name')}
                        value={property.displayName}
                        onChange={(newValue: string) => handlePropertyChange(property.id, 'displayName', newValue)}
                        placeholder={t('userTypes:displayNamePlaceholder', 'e.g., First Name')}
                        defaultNewKey={name.trim() && property.name.trim() ? `${name.trim()}.${property.name.trim()}` : undefined}
                      />

                      {/* Checkbox options with info tooltips */}
                      <Box sx={{display: 'flex', gap: 3}}>
                        <Tooltip title={t('userTypes:tooltips.required', 'This field must be provided when creating a user')} placement="top" arrow>
                          <FormControlLabel
                            control={
                              <Checkbox
                                checked={property.required}
                                onChange={(e) => handlePropertyChange(property.id, 'required', e.target.checked)}
                              />
                            }
                            label={
                              <Stack direction="row" alignItems="center" spacing={0.5}>
                                <span>{t('common:form.required', 'Required')}</span>
                                <Info size={14} color="inherit" />
                              </Stack>
                            }
                          />
                        </Tooltip>
                        {(property.type === 'string' || property.type === 'number' || property.type === 'enum') && !property.credential && (
                          <Tooltip title={t('userTypes:tooltips.unique', 'Each user must have a distinct value for this field')} placement="top" arrow>
                            <FormControlLabel
                              control={
                                <Checkbox
                                  checked={property.unique}
                                  onChange={(e) => handlePropertyChange(property.id, 'unique', e.target.checked)}
                                />
                              }
                              label={
                                <Stack direction="row" alignItems="center" spacing={0.5}>
                                  <span>{t('userTypes:unique', 'Unique')}</span>
                                  <Info size={14} color="inherit" />
                                </Stack>
                              }
                            />
                          </Tooltip>
                        )}
                        {(property.type === 'string' || property.type === 'number') && (
                          <Tooltip title={t('userTypes:tooltips.credential', 'Values will be hashed and not returned in API responses')} placement="top" arrow>
                            <FormControlLabel
                              control={
                                <Checkbox
                                  checked={property.credential}
                                  onChange={({target: {checked}}) => {
                                    setProperties(
                                      properties.map((prop) =>
                                        prop.id === property.id
                                          ? {...prop, credential: checked, ...(checked && {unique: false})}
                                          : prop,
                                      ),
                                    );
                                  }}
                                />
                              }
                              label={
                                <Stack direction="row" alignItems="center" spacing={0.5}>
                                  <span>{t('userTypes:credential', 'Credential')}</span>
                                  <Info size={14} color="inherit" />
                                </Stack>
                              }
                            />
                          </Tooltip>
                        )}
                      </Box>

                      {/* Credential indicator */}
                      {property.credential && (
                        <Alert severity="info" variant="outlined">
                          {t('userTypes:credentialHint', 'This field will be treated as a secret. Values will be hashed and cannot be retrieved.')}
                        </Alert>
                      )}

                      {/* String: regex pattern */}
                      {property.type === 'string' && (
                        <FormControl fullWidth>
                          <FormLabel>{t('userTypes:regexPattern', 'Regular Expression Pattern (Optional)')}</FormLabel>
                          <TextField
                            value={property.regex}
                            onChange={(e) => handlePropertyChange(property.id, 'regex', e.target.value)}
                            placeholder={t('userTypes:regexPlaceholder', 'e.g., ^[a-zA-Z0-9]+$')}
                            size="small"
                          />
                        </FormControl>
                      )}

                      {/* Enum: value input + chips */}
                      {property.type === 'enum' && (
                        <FormControl fullWidth>
                          <FormLabel>{t('userTypes:enumValues', 'Allowed Values (Enum)')}</FormLabel>
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
                              placeholder={t('userTypes:enumPlaceholder', 'Add value and press Enter')}
                              size="small"
                              fullWidth
                            />
                            <Button variant="outlined" size="small" onClick={() => handleAddEnumValue(property.id)}>
                              {t('common:actions.add', 'Add')}
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
                      )}
                    </Stack>
                  </Paper>
              ))}

              {/* Add Property Button */}
              <Button
                variant="outlined"
                startIcon={<Plus size={16} />}
                onClick={handleAddProperty}
                fullWidth
                sx={{
                  py: 1.5,
                  mb: 2,
                  borderStyle: 'dashed',
                  '&:hover': {borderStyle: 'dashed'},
                }}
              >
                {t('userTypes:addProperty', 'Add Property')}
              </Button>

              {/* Update Error Display */}
              {updateUserTypeMutation.error && (
                <Alert severity="error" sx={{mt: 2}}>
                  <Typography variant="body2" sx={{fontWeight: 'bold', mb: 0.5}}>
                    {updateUserTypeMutation.error.message}
                  </Typography>
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
          {deleteUserTypeMutation.error && (
            <Alert severity="error" sx={{mt: 2}}>
              <Typography variant="body2" sx={{fontWeight: 'bold'}}>
                {deleteUserTypeMutation.error.message}
              </Typography>
            </Alert>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={handleDeleteCancel} disabled={deleteUserTypeMutation.isPending}>
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
            disabled={deleteUserTypeMutation.isPending}
          >
            {deleteUserTypeMutation.isPending ? 'Deleting...' : 'Delete'}
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
    </PageContent>
  );
}
