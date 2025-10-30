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
import {useState} from 'react';
import Box from '@mui/material/Box';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Button from '@mui/material/Button';
import Paper from '@mui/material/Paper';
import FormLabel from '@mui/material/FormLabel';
import FormControl from '@mui/material/FormControl';
import Select from '@mui/material/Select';
import MenuItem from '@mui/material/MenuItem';
import TextField from '@mui/material/TextField';
import Checkbox from '@mui/material/Checkbox';
import FormControlLabel from '@mui/material/FormControlLabel';
import IconButton from '@mui/material/IconButton';
import Alert from '@mui/material/Alert';
import Snackbar from '@mui/material/Snackbar';
import {ArrowLeft, Plus, Save, X} from 'lucide-react';
import useCreateUserType from '../api/useCreateUserType';
import type {PropertyDefinition, UserSchemaDefinition, PropertyType, SchemaPropertyInput} from '../types/user-types';

export default function CreateUserTypePage() {
  const navigate = useNavigate();
  const {createUserType, loading, error: createError} = useCreateUserType();

  const [name, setName] = useState('');
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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setValidationError(null);

    // Validate
    if (!name.trim()) {
      setValidationError('Please enter a user type name');
      setSnackbarOpen(true);
      return;
    }

    const validProperties = properties.filter((prop) => prop.name.trim());
    if (validProperties.length === 0) {
      setValidationError('Please add at least one property');
      setSnackbarOpen(true);
      return;
    }

    // Check for duplicate property names
    const propertyNames = validProperties.map((prop) => prop.name.trim());
    const duplicates = propertyNames.filter((propName, index) => propertyNames.indexOf(propName) !== index);
    if (duplicates.length > 0) {
      setValidationError(`Duplicate property names found: ${duplicates.join(', ')}`);
      setSnackbarOpen(true);
      return;
    }

    // Convert properties to schema definition
    const schema: UserSchemaDefinition = {};
    validProperties.forEach((prop) => {
      const propDef: Partial<PropertyDefinition> = {
        type: prop.type,
        required: prop.required,
      };

      // Add type-specific fields
      if (prop.type === 'string' || prop.type === 'number') {
        if (prop.unique) {
          (propDef as {unique?: boolean}).unique = true;
        }
      }

      if (prop.type === 'string') {
        if (prop.enum.length > 0) {
          (propDef as {enum?: string[]}).enum = prop.enum;
        }
        if (prop.regex.trim()) {
          (propDef as {regex?: string}).regex = prop.regex;
        }
      }

      // For array and object types, we'll use basic definitions for now
      if (prop.type === 'array') {
        (propDef as {items?: {type: string}}).items = {type: 'string'};
      } else if (prop.type === 'object') {
        (propDef as {properties?: Record<string, PropertyDefinition>}).properties = {};
      }

      schema[prop.name.trim()] = propDef as PropertyDefinition;
    });

    try {
      await createUserType({
        name: name.trim(),
        schema,
      });

      // Navigate back to list on success
      await navigate('/user-types');
    } catch {
      // TODO: Log the errors
      // Tracker: https://github.com/asgardeo/thunder/issues/618
    }
  };

  const handleCloseSnackbar = () => {
    setSnackbarOpen(false);
  };

  return (
    <Box sx={{maxWidth: 1200, mx: 'auto', px: 2, position: 'relative', pt: 6}}>
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
            Create User Type
          </Typography>
          <Typography variant="body2" color="text.secondary">
            Define a new user type schema for your organization
          </Typography>
        </Box>
      </Stack>

      <Paper sx={{p: 4}}>
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
              User Type Name <span style={{color: 'red'}}>*</span>
            </FormLabel>
            <TextField
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g., Employee, Customer, Partner"
              required
              variant="outlined"
            />
          </FormControl>

          <Typography variant="h6" gutterBottom sx={{mt: 4, mb: 2}}>
            Schema Properties
          </Typography>

          {properties.map((property) => (
            <Paper key={property.id} variant="outlined" sx={{p: 2, mb: 2}}>
              <Stack direction="row" justifyContent="space-between" alignItems="center" mb={2}>
                {properties.length > 1 && (
                  <IconButton size="small" color="error" onClick={() => handleRemoveProperty(property.id)}>
                    <X size={16} />
                  </IconButton>
                )}
              </Stack>

              <Box sx={{display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 2}}>
                <FormControl>
                  <FormLabel>Property Name</FormLabel>
                  <TextField
                    value={property.name}
                    onChange={(e) => handlePropertyChange(property.id, 'name', e.target.value)}
                    placeholder="e.g., email, age, address"
                    size="small"
                  />
                </FormControl>

                <FormControl>
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

              <Box sx={{mt: 2, display: 'flex', gap: 2}}>
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
              </Box>

              {property.type === 'string' && (
                <>
                  <FormControl fullWidth sx={{mt: 2}}>
                    <FormLabel>Regular Expression Pattern (Optional)</FormLabel>
                    <TextField
                      value={property.regex}
                      onChange={(e) => handlePropertyChange(property.id, 'regex', e.target.value)}
                      placeholder="e.g., ^[a-zA-Z0-9]+$"
                      size="small"
                    />
                  </FormControl>

                  <FormControl fullWidth sx={{mt: 2}}>
                    <FormLabel>Allowed Values (Enum) - Optional</FormLabel>
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
                        placeholder="Add value and press Enter"
                        size="small"
                        fullWidth
                      />
                      <Button variant="outlined" size="small" onClick={() => handleAddEnumValue(property.id)}>
                        Add
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
                              border: '1px solid #ccc',
                              borderRadius: 1,
                              px: 1,
                              py: 0.5,
                            }}
                          >
                            <Typography variant="body2">{val}</Typography>
                            <IconButton
                              size="small"
                              onClick={() => handleRemoveEnumValue(property.id, val)}
                              sx={{ml: 0.5}}
                            >
                              <X size={14} />
                            </IconButton>
                          </Box>
                        ))}
                      </Box>
                    )}
                  </FormControl>
                </>
              )}
            </Paper>
          ))}

          <Button variant="outlined" startIcon={<Plus size={16} />} onClick={handleAddProperty} sx={{mb: 3}}>
            Add Property
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
              Cancel
            </Button>
            <Button type="submit" variant="contained" disabled={loading} startIcon={<Save size={16} />}>
              {loading ? 'Creating...' : 'Create User Type'}
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
