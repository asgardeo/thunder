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

import {cn} from '@thunderid/utils';
import {FormControl, FormLabel, IconButton, InputAdornment, TextField} from '@wso2/oxygen-ui';
import {Eye, EyeClosed} from '@wso2/oxygen-ui-icons-react';
import type {JSX} from 'react';
import {useState} from 'react';
import {useTranslation} from 'react-i18next';
import type {FlowFieldProps} from '../../../models/flow';

export interface PasswordInputAdapterProps extends FlowFieldProps {
  /**
   * Browser autocomplete hint.
   * Use `'current-password'` for sign-in and `'new-password'` for sign-up / invite.
   * @default 'current-password'
   */
  passwordAutoComplete?: 'current-password' | 'new-password';
}

export default function PasswordInputAdapter({
  component,
  values,
  touched,
  fieldErrors,
  isLoading,
  resolve,
  onInputChange,
  passwordAutoComplete = 'current-password',
}: PasswordInputAdapterProps): JSX.Element | null {
  const {t} = useTranslation();
  const [showPassword, setShowPassword] = useState(false);
  const {ref} = component;

  if (!ref || typeof ref !== 'string') return null;

  const hasError = !!(touched?.[ref] && fieldErrors?.[ref]);
  const value = values[ref] ?? '';
  const autoComplete = ref === 'password' ? (passwordAutoComplete ?? 'current-password') : 'off';

  return (
    <FormControl required={component.required} className={cn('Flow--passwordInput', 'FormControl--root')}>
      <FormLabel htmlFor={ref} className={cn('Label--root')}>
        {t(resolve(component.label)!)}
      </FormLabel>
      <TextField
        fullWidth
        className={cn('TextField--root')}
        id={ref}
        name={ref}
        type={showPassword ? 'text' : 'password'}
        placeholder={t(resolve(component.placeholder) ?? component.placeholder ?? '')}
        autoComplete={autoComplete}
        required={component.required}
        variant="outlined"
        disabled={isLoading}
        error={hasError}
        helperText={hasError ? fieldErrors?.[ref] : undefined}
        color={hasError ? 'error' : 'primary'}
        value={value}
        onChange={(e) => onInputChange(ref, e.target.value)}
        slotProps={{
          input: {
            endAdornment: (
              <InputAdornment position="end">
                <IconButton
                  className={cn('IconButton--root', 'PasswordInput--toggle')}
                  aria-label="toggle password visibility"
                  onClick={() => setShowPassword((prev) => !prev)}
                  edge="end"
                  disabled={isLoading}
                >
                  {showPassword ? <Eye /> : <EyeClosed />}
                </IconButton>
              </InputAdornment>
            ),
          },
        }}
      />
    </FormControl>
  );
}
