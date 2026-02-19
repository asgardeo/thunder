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
import {useEffect, useMemo, useState} from 'react';
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
  TextField,
  Checkbox,
  FormControlLabel,
  IconButton,
  Alert,
  Snackbar,
} from '@wso2/oxygen-ui';
import {ArrowLeft, Plus, Save, X} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import {useLogger} from '@thunder/logger/react';
import useCreateUserType from '../api/useCreateUserType';
import useGetOrganizationUnits from '../../organization-units/api/useGetOrganizationUnits';
import type {
  PropertyDefinition,
  UserTypeDefinition,
  UIPropertyType,
  SchemaPropertyInput,
  CreateUserTypeRequest,
} from '../types/user-types';

export default function CreateUserTypePage() {
  const navigate = useNavigate();
  const {t} = useTranslation();
  const logger = useLogger('CreateUserTypePage');
  const {createUserType, loading, error: createError} = useCreateUserType();
  const {
    data: organizationUnitsResponse,
    isLoading: organizationUnitsLoading,
    error: organizationUnitsError,
  } = useGetOrganizationUnits();

  const [name, setName] = useState('');
  const [ouId, setOuId] = useState('');
  const [allowSelfRegistration, setAllowSelfRegistration] = useState(false);
  const [properties, setProperties] = useState<SchemaPropertyInput[]>([
    {
      id: '1',
      name: '',
      type: 'string',
      required: false,
      unique: false,
      enum: [],
      regex: '',
    },
  ]);
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

  useEffect(() => {
    if (!ouId && organizationUnits.length > 0) {
      setOuId(organizationUnits[0].id);
    }
  }, [organizationUnits, ouId]);

  const handleBack = async () => {
    await navigate('/user-types');
  };

  const handleAddProperty = () => {
    const newProperty: SchemaPropertyInput = {
      id: Date.now().toString(),
      name: '',
      type: 'string',
      required: false,
      unique: false,
      enum: [],
      regex: '',
    };
    setProperties([...properties, newProperty]);
  };

  const handleRemoveProperty = (id: string) => {
    setProperties(properties.filter((prop) => prop.id !== id));
    const newEnumInput = {...enumInput};
    delete newEnumInput[id];
    setEnumInput(newEnumInput);
  };

  const handlePropertyChange = <K extends keyof SchemaPropertyInput>(
    id: string,
    field: K,
    value: SchemaPropertyInput[K],
  ) => {
    setProperties(
      properties.map((prop) =>
        prop.id === id
          ? {
              ...prop,
              [field]: value,
              // Reset type-specific fields when type changes
              ...(field === 'type' && {
                enum: (value as UIPropertyType) === 'enum' ? prop.enum : [],
                regex: '',
                unique:
                  (value as UIPropertyType) === 'string' ||
                  (value as UIPropertyType) === 'number' ||
                  (value as UIPropertyType) === 'enum'
                    ? prop.unique
                    : false,
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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setValidationError(null);

    // Validate
    if (!name.trim()) {
      setValidationError(t('userTypes:validationErrors.nameRequired'));
      setSnackbarOpen(true);
      return;
    }

    const trimmedOuId = ouId.trim();
    if (!trimmedOuId) {
      setValidationError(t('userTypes:validationErrors.ouIdRequired'));
      setSnackbarOpen(true);
      return;
    }

    const validProperties = properties.filter((prop) => prop.name.trim());
    if (validProperties.length === 0) {
      setValidationError(t('userTypes:validationErrors.propertiesRequired'));
      setSnackbarOpen(true);
      return;
    }

    // Check for duplicate property names
    const propertyNames = validProperties.map((prop) => prop.name.trim());
    const duplicates = propertyNames.filter((propName, index) => propertyNames.indexOf(propName) !== index);
    if (duplicates.length > 0) {
      setValidationError(t('userTypes:validationErrors.duplicateProperties', {duplicates: duplicates.join(', ')}));
      setSnackbarOpen(true);
      return;
    }

    // Convert properties to schema definition
    const schema: UserTypeDefinition = {};
    validProperties.forEach((prop) => {
      // Convert enum type to string type with enum values
      const actualType = prop.type === 'enum' ? 'string' : prop.type;

      const propDef: Partial<PropertyDefinition> = {
        type: actualType,
        required: prop.required,
      };

      // Add type-specific fields
      if (actualType === 'string' || actualType === 'number') {
        if (prop.unique) {
          (propDef as {unique?: boolean}).unique = true;
        }
      }

      if (actualType === 'string') {
        // For enum type or string with enum values, add enum array
        if (prop.type === 'enum' || prop.enum.length > 0) {
          (propDef as {enum?: string[]}).enum = prop.enum;
        }
        if (prop.regex.trim()) {
          (propDef as {regex?: string}).regex = prop.regex;
        }
      }

      // For array and object types, we'll use basic definitions for now
      if (actualType === 'array') {
        (propDef as {items?: {type: string}}).items = {type: 'string'};
      } else if (actualType === 'object') {
        (propDef as {properties?: Record<string, PropertyDefinition>}).properties = {};
      }

      schema[prop.name.trim()] = propDef as PropertyDefinition;
    });

    const requestBody: CreateUserTypeRequest = {
      name: name.trim(),
      ouId: trimmedOuId,
      schema,
    };

    if (allowSelfRegistration) {
      requestBody.allowSelfRegistration = true;
    }

    try {
      await createUserType(requestBody);

      // Navigate back to list on success
      await navigate('/user-types');
    } catch (error) {
      logger.error('Failed to create user type or navigate', {error, userTypeName: name});
    }
  };

  const handleCloseSnackbar = () => {
    setSnackbarOpen(false);
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
        {t('common:actions.back')}
      </Button>

      <Stack direction="row" alignItems="flex-start" justifyContent="space-between" mb={4} gap={2}>
        <Box>
          <Typography variant="h1" gutterBottom>
            {t('userTypes:addUserType')}
          </Typography>
          <Typography variant="subtitle1" color="text.secondary">
            {t('userTypes:createDescription')}
          </Typography>
        </Box>
      </Stack>

      <Paper sx={{p: 4}}>
        {!organizationUnitsLoading && organizationUnitsError && (
          <Alert severity="error" sx={{mb: 3}}>
            <Typography variant="body2" sx={{fontWeight: 600, mb: 0.5}}>
              {t('userTypes:errors.organizationUnitsFailedTitle')}
            </Typography>
            <Typography variant="body2">{organizationUnitsError.message}</Typography>
          </Alert>
        )}
        <Box
          component="form"
          onSubmit={(e) => {
            handleSubmit(e).catch(() => {
              // Handle error
            });
          }}
          noValidate
        >
          <FormControl fullWidth sx={{mb: 3}}>
            <FormLabel htmlFor="name">
              {t('userTypes:typeName')} <span style={{color: 'red'}}>*</span>
            </FormLabel>
            <TextField
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder={t('userTypes:typeNamePlaceholder')}
              required
              variant="outlined"
            />
          </FormControl>

          <FormControl fullWidth sx={{mb: 3}}>
            <FormLabel htmlFor="ouId">
              {t('userTypes:organizationUnit')} <span style={{color: 'red'}}>*</span>
            </FormLabel>
            <Select
              id="ouId"
              aria-label={t('userTypes:organizationUnit')}
              value={ouId}
              onChange={(event) => setOuId(event.target.value ?? '')}
              displayEmpty
              required
              renderValue={(selected) => {
                const value = typeof selected === 'string' ? selected : '';
                if (!value) {
                  return t('userTypes:ouSelectPlaceholder');
                }

                return selectedOrganizationUnit ? selectedOrganizationUnit.name : value;
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
          </FormControl>

          <FormControlLabel
            control={
              <Checkbox checked={allowSelfRegistration} onChange={(e) => setAllowSelfRegistration(e.target.checked)} />
            }
            label={t('userTypes:allowSelfRegistration')}
            sx={{mb: 2}}
          />

          <Typography variant="h6" gutterBottom sx={{mt: 4, mb: 2}}>
            {t('userTypes:schemaProperties')}
          </Typography>

          {properties.map((property) => (
            <Paper key={property.id} variant="outlined" sx={{px: 2, py: 4, mb: 2}}>
              {properties.length > 1 && (
                <Stack direction="row" justifyContent="flex-end" alignItems="center">
                  <IconButton size="small" color="error" onClick={() => handleRemoveProperty(property.id)}>
                    <X size={16} />
                  </IconButton>
                </Stack>
              )}
              <Box sx={{display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 2}}>
                <FormControl>
                  <FormLabel>{t('userTypes:propertyName')}</FormLabel>
                  <TextField
                    value={property.name}
                    onChange={(e) => handlePropertyChange(property.id, 'name', e.target.value)}
                    placeholder={t('userTypes:propertyNamePlaceholder')}
                    size="small"
                  />
                </FormControl>

                <FormControl>
                  <FormLabel>{t('userTypes:propertyType')}</FormLabel>
                  <Select
                    value={property.type}
                    onChange={(e) => handlePropertyChange(property.id, 'type', e.target.value as UIPropertyType)}
                    size="small"
                  >
                    <MenuItem value="string">{t('userTypes:types.string')}</MenuItem>
                    <MenuItem value="number">{t('userTypes:types.number')}</MenuItem>
                    <MenuItem value="boolean">{t('userTypes:types.boolean')}</MenuItem>
                    <MenuItem value="enum">{t('userTypes:types.enum')}</MenuItem>
                  </Select>
                </FormControl>
              </Box>

              <Box sx={{mt: 2, display: 'flex', gap: 2}}>
                <FormControlLabel
                  control={
                    <Checkbox
                      checked={property.required}
                      onChange={(e) => handlePropertyChange(property.id, 'required', e.target.checked)}
                    />
                  }
                  label={t('common:form.required')}
                />
                {(property.type === 'string' || property.type === 'number' || property.type === 'enum') && (
                  <FormControlLabel
                    control={
                      <Checkbox
                        checked={property.unique}
                        onChange={(e) => handlePropertyChange(property.id, 'unique', e.target.checked)}
                      />
                    }
                    label={t('userTypes:unique')}
                  />
                )}
              </Box>

              {property.type === 'string' && (
                <FormControl fullWidth sx={{mt: 2}}>
                  <FormLabel>{t('userTypes:regexPattern')}</FormLabel>
                  <TextField
                    value={property.regex}
                    onChange={(e) => handlePropertyChange(property.id, 'regex', e.target.value)}
                    placeholder={t('userTypes:regexPlaceholder')}
                    size="small"
                  />
                </FormControl>
              )}

              {property.type === 'enum' && (
                <FormControl fullWidth sx={{mt: 2}}>
                  <FormLabel>{t('userTypes:enumValues')}</FormLabel>
                  <Box sx={{display: 'flex', gap: 1}}>
                    <TextField
                      value={enumInput[property.id] ?? ''}
                      onChange={(e) => setEnumInput({...enumInput, [property.id]: e.target.value})}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter') {
                          e.preventDefault();
                          handleAddEnumValue(property.id);
                        }
                      }}
                      placeholder={t('userTypes:enumPlaceholder')}
                      size="small"
                      fullWidth
                    />
                    <Button variant="outlined" onClick={() => handleAddEnumValue(property.id)}>
                      {t('common:actions.add')}
                    </Button>
                  </Box>
                  {property.enum.length > 0 && (
                    <Box sx={{mt: 1, display: 'flex', flexWrap: 'wrap', gap: 1}}>
                      {property.enum.map((val) => (
                        <Box
                          key={val}
                          sx={{
                            display: 'inline-flex',
                            alignItems: 'center',
                            border: '1px solid',
                            borderColor: 'divider',
                            borderRadius: 1,
                            px: 1.5,
                            py: 0.5,
                          }}
                        >
                          <Typography variant="body2">{val}</Typography>
                          <IconButton
                            size="small"
                            onClick={() => handleRemoveEnumValue(property.id, val)}
                            sx={{
                              ml: 0.5,
                              border: 'none',
                            }}
                          >
                            <X size={14} />
                          </IconButton>
                        </Box>
                      ))}
                    </Box>
                  )}
                </FormControl>
              )}
            </Paper>
          ))}

          <Button variant="outlined" startIcon={<Plus size={16} />} onClick={handleAddProperty} sx={{mb: 3}}>
            {t('userTypes:addProperty')}
          </Button>

          {createError && (
            <Alert severity="error" sx={{mb: 3}}>
              <Typography variant="body2" sx={{fontWeight: 'bold', mb: 0.5}}>
                {createError.message}
              </Typography>
              {createError.description && <Typography variant="body2">{createError.description}</Typography>}
            </Alert>
          )}

          <Stack direction="row" spacing={2} justifyContent="flex-end" sx={{mt: 4}}>
            <Button
              variant="outlined"
              onClick={() => {
                handleBack().catch(() => {
                  // Handle navigation error
                });
              }}
              disabled={loading}
              startIcon={<X size={16} />}
            >
              {t('common:actions.cancel')}
            </Button>
            <Button type="submit" variant="contained" disabled={loading} startIcon={<Save size={16} />}>
              {loading ? t('common:status.saving') : t('userTypes:createUserType')}
            </Button>
          </Stack>
        </Box>
      </Paper>

      {/* Validation Error Snackbar */}
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
