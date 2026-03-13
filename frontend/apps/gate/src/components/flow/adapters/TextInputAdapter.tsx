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
import {FormControl, FormLabel, TextField} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';
import type {FlowFieldProps} from '../../../models/flow';

type TextInputVariant = 'TEXT_INPUT' | 'EMAIL_INPUT' | 'PHONE_INPUT';

const HTML_INPUT_TYPE: Record<TextInputVariant, string> = {
  TEXT_INPUT: 'text',
  EMAIL_INPUT: 'email',
  PHONE_INPUT: 'tel',
};

const AUTO_COMPLETE_MAP: Record<TextInputVariant, (ref: string) => string> = {
  TEXT_INPUT: (ref) => {
    if (ref === 'username') return 'username';
    if (ref === 'email') return 'email';
    return 'off';
  },
  EMAIL_INPUT: () => 'email',
  PHONE_INPUT: () => 'tel',
};

function resolveTextVariant(type: string): TextInputVariant {
  if (type === 'EMAIL_INPUT') return 'EMAIL_INPUT';
  if (type === 'PHONE_INPUT') return 'PHONE_INPUT';
  return 'TEXT_INPUT';
}

export default function TextInputAdapter({
  component,
  values,
  touched,
  fieldErrors,
  isLoading,
  resolve,
  onInputChange,
}: FlowFieldProps): JSX.Element | null {
  const {t} = useTranslation();
  const {ref} = component;

  if (!ref || typeof ref !== 'string') return null;

  const variant = resolveTextVariant(String(component.type));
  const htmlType = HTML_INPUT_TYPE[variant];
  const autoComplete = AUTO_COMPLETE_MAP[variant](ref);
  const autoFocus = ref === 'username';
  const hasError = !!(touched?.[ref] && fieldErrors?.[ref]);
  const value = values[ref] ?? '';

  return (
    <FormControl required={component.required}>
      <FormLabel htmlFor={ref}>{t(resolve(component.label)!)}</FormLabel>
      <TextField
        fullWidth
        id={ref}
        name={ref}
        type={htmlType}
        placeholder={t(resolve(component.placeholder) ?? component.placeholder ?? '')}
        autoComplete={autoComplete}
        autoFocus={autoFocus}
        required={component.required}
        variant="outlined"
        disabled={isLoading}
        error={hasError}
        helperText={hasError ? fieldErrors?.[ref] : undefined}
        color={hasError ? 'error' : 'primary'}
        value={value}
        onChange={(e) => onInputChange(ref, e.target.value)}
      />
    </FormControl>
  );
}
