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

import type {JSX} from 'react';
import {FormControl, FormLabel, MenuItem, Select, Typography} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';
import type {FlowFieldProps} from '../../../models/flow';

function getOptionValue(option: unknown): string {
  if (typeof option === 'string') return option;
  if (typeof option === 'object' && option !== null && 'value' in option) {
    const {value} = option as {value: unknown};
    if (typeof value === 'string') return value;
    return JSON.stringify(value ?? option);
  }
  return JSON.stringify(option);
}

function getOptionLabel(option: unknown): string {
  if (typeof option === 'string') return option;
  if (typeof option === 'object' && option !== null && 'label' in option) {
    const {label} = option as {label: unknown};
    if (typeof label === 'string') return label;
    return JSON.stringify(label ?? option);
  }
  return JSON.stringify(option);
}

export default function SelectAdapter({
  component,
  values,
  touched,
  fieldErrors,
  isLoading,
  resolve,
  onInputChange,
}: FlowFieldProps): JSX.Element | null {
  const {t} = useTranslation();
  const {ref, options, hint} = component;

  if (!ref || typeof ref !== 'string' || !options) return null;

  const hasError = !!(touched?.[ref] && fieldErrors?.[ref]);
  const value = values[ref] ?? '';

  return (
    <FormControl fullWidth>
      <FormLabel htmlFor={ref}>{t(resolve(component.label)!)}</FormLabel>
      <Select
        displayEmpty
        size="small"
        id={ref}
        name={ref}
        required={component.required}
        fullWidth
        disabled={isLoading}
        error={hasError}
        value={value}
        onChange={(e) => onInputChange(ref, e.target.value)}
      >
        <MenuItem value="" disabled>
          {t(resolve(component.placeholder) ?? 'Select an option')}
        </MenuItem>
        {options.map((option: unknown) => (
          <MenuItem key={getOptionValue(option)} value={getOptionValue(option)}>
            {getOptionLabel(option)}
          </MenuItem>
        ))}
      </Select>
      {hasError && (
        <Typography variant="caption" color="error.main" sx={{mt: 0.5}}>
          {fieldErrors?.[ref]}
        </Typography>
      )}
      {hint && (
        <Typography variant="caption" color="text.secondary">
          {hint}
        </Typography>
      )}
    </FormControl>
  );
}
