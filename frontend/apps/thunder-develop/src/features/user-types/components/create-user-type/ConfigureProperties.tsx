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

import {
  Alert,
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
  Tooltip,
  Chip,
} from '@wso2/oxygen-ui';
import {Plus, Trash2, Info} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useEffect, useRef} from 'react';
import {useTranslation} from 'react-i18next';
import type {SchemaPropertyInput, UIPropertyType} from '../../types/user-types';

/**
 * Props for the {@link ConfigureProperties} component.
 *
 * @public
 */
export interface ConfigurePropertiesProps {
  properties: SchemaPropertyInput[];
  onPropertiesChange: (properties: SchemaPropertyInput[]) => void;
  enumInput: Record<string, string>;
  onEnumInputChange: (enumInput: Record<string, string>) => void;
  onReadyChange?: (isReady: boolean) => void;
}

/**
 * Step 3 of the user type creation wizard: configure schema properties.
 *
 * @public
 */
export default function ConfigureProperties({
  properties,
  onPropertiesChange,
  enumInput,
  onEnumInputChange,
  onReadyChange = undefined,
}: ConfigurePropertiesProps): JSX.Element {
  const {t} = useTranslation();
  const nextId = useRef(properties.length + 1);

  // Broadcast readiness - ready when at least one property has a name
  useEffect((): void => {
    if (onReadyChange) {
      const hasValidProperty = properties.some((prop) => prop.name.trim().length > 0);
      onReadyChange(hasValidProperty);
    }
  }, [properties, onReadyChange]);

  const handleAddProperty = () => {
    const id = String(nextId.current);
    nextId.current += 1;
    const newProperty: SchemaPropertyInput = {
      id,
      name: '',
      type: 'string',
      required: false,
      unique: false,
      credential: false,
      enum: [],
      regex: '',
    };
    onPropertiesChange([...properties, newProperty]);
  };

  const handleRemoveProperty = (id: string) => {
    onPropertiesChange(properties.filter((prop) => prop.id !== id));
    const newEnumInput = {...enumInput};
    delete newEnumInput[id];
    onEnumInputChange(newEnumInput);
  };

  const handlePropertyChange = <K extends keyof SchemaPropertyInput>(
    id: string,
    field: K,
    value: SchemaPropertyInput[K],
  ) => {
    onPropertiesChange(
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
                credential:
                  (value as UIPropertyType) === 'string' || (value as UIPropertyType) === 'number'
                    ? prop.credential
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

    const property = properties.find((prop) => prop.id === propertyId);
    if (property?.enum.includes(inputValue)) return;

    onPropertiesChange(
      properties.map((prop) => (prop.id === propertyId ? {...prop, enum: [...prop.enum, inputValue]} : prop)),
    );

    onEnumInputChange({...enumInput, [propertyId]: ''});
  };

  const handleRemoveEnumValue = (propertyId: string, enumValue: string) => {
    onPropertiesChange(
      properties.map((prop) =>
        prop.id === propertyId ? {...prop, enum: prop.enum.filter((val) => val !== enumValue)} : prop,
      ),
    );
  };

  const supportsUnique = (type: UIPropertyType): boolean =>
    type === 'string' || type === 'number' || type === 'enum';

  const supportsCredential = (type: UIPropertyType): boolean =>
    type === 'string' || type === 'number';

  return (
    <Stack direction="column" spacing={4} data-testid="configure-properties">
      <Stack direction="column" spacing={1}>
        <Typography variant="h1" gutterBottom>
          {t('userTypes:createWizard.properties.title')}
        </Typography>
        <Typography variant="subtitle1" color="text.secondary">
          {t('userTypes:createWizard.properties.subtitle')}
        </Typography>
      </Stack>

      {properties.map((property) => (
        <Paper
          key={property.id}
          variant="outlined"
          sx={{
            position: 'relative',
            px: 3,
            py: 3,
            borderRadius: 2,
            transition: 'border-color 0.2s',
            '&:hover': {borderColor: 'primary.main'},
            '&:hover .property-delete-btn': {opacity: 1},
          }}
        >
          {/* Remove button - visible on hover */}
          {properties.length > 1 && (
            <Tooltip title={t('userTypes:removeProperty')}>
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

          {/* Name and Type fields */}
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

          {/* Checkbox options with info tooltips */}
          <Box sx={{mt: 2.5, display: 'flex', gap: 3}}>
            <Tooltip title={t('userTypes:tooltips.required')} placement="top" arrow>
              <FormControlLabel
                control={
                  <Checkbox
                    checked={property.required}
                    onChange={(e) => handlePropertyChange(property.id, 'required', e.target.checked)}
                  />
                }
                label={
                  <Stack direction="row" alignItems="center" spacing={0.5}>
                    <span>{t('common:form.required')}</span>
                    <Info size={14} color="inherit" />
                  </Stack>
                }
              />
            </Tooltip>
            {supportsUnique(property.type) && (
              <Tooltip title={t('userTypes:tooltips.unique')} placement="top" arrow>
                <FormControlLabel
                  control={
                    <Checkbox
                      checked={property.unique}
                      onChange={(e) => handlePropertyChange(property.id, 'unique', e.target.checked)}
                    />
                  }
                  label={
                    <Stack direction="row" alignItems="center" spacing={0.5}>
                      <span>{t('userTypes:unique')}</span>
                      <Info size={14} color="inherit" />
                    </Stack>
                  }
                />
              </Tooltip>
            )}
            {supportsCredential(property.type) && (
              <Tooltip title={t('userTypes:tooltips.credential')} placement="top" arrow>
                <FormControlLabel
                  control={
                    <Checkbox
                      checked={property.credential}
                      onChange={(e) => handlePropertyChange(property.id, 'credential', e.target.checked)}
                    />
                  }
                  label={
                    <Stack direction="row" alignItems="center" spacing={0.5}>
                      <span>{t('userTypes:credential')}</span>
                      <Info size={14} color="inherit" />
                    </Stack>
                  }
                />
              </Tooltip>
            )}
          </Box>

          {/* Credential indicator */}
          {property.credential && (
            <Alert severity="info" variant="outlined" sx={{mt: 2}}>
              {t('userTypes:credentialHint')}
            </Alert>
          )}

          {/* String: regex pattern */}
          {property.type === 'string' && (
            <FormControl fullWidth sx={{mt: 2.5}}>
              <FormLabel>{t('userTypes:regexPattern')}</FormLabel>
              <TextField
                value={property.regex}
                onChange={(e) => handlePropertyChange(property.id, 'regex', e.target.value)}
                placeholder={t('userTypes:regexPlaceholder')}
                size="small"
              />
            </FormControl>
          )}

          {/* Enum: value input + chips */}
          {property.type === 'enum' && (
            <FormControl fullWidth sx={{mt: 2.5}}>
              <FormLabel>{t('userTypes:enumValues')}</FormLabel>
              <Box sx={{display: 'flex', gap: 1}}>
                <TextField
                  value={enumInput[property.id] ?? ''}
                  onChange={(e) => onEnumInputChange({...enumInput, [property.id]: e.target.value})}
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
                <Box sx={{mt: 1.5, display: 'flex', flexWrap: 'wrap', gap: 1}}>
                  {property.enum.map((val) => (
                    <Chip
                      key={val}
                      label={val}
                      size="small"
                      onDelete={() => handleRemoveEnumValue(property.id, val)}
                    />
                  ))}
                </Box>
              )}
            </FormControl>
          )}
        </Paper>
      ))}

      <Button
        variant="outlined"
        startIcon={<Plus size={16} />}
        onClick={handleAddProperty}
        fullWidth
        sx={{
          py: 1.5,
          borderStyle: 'dashed',
          '&:hover': {borderStyle: 'dashed'},
        }}
      >
        {t('userTypes:addProperty')}
      </Button>
    </Stack>
  );
}
