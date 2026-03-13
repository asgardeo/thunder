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
import {Box, FormControl, FormLabel, TextField, Typography} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';
import type {FlowFieldProps} from '../../../models/flow';

const OTP_LENGTH = 6;

export default function OtpInputAdapter({
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

  const hasError = !!(touched?.[ref] && fieldErrors?.[ref]);
  const otpValue = values[ref] ?? '';
  const otpDigits = otpValue.padEnd(OTP_LENGTH, ' ').split('').slice(0, OTP_LENGTH);

  const focusDigit = (idx: number) => {
    const input = document.querySelector<HTMLInputElement>(`input[aria-label="OTP digit ${idx + 1}"]`);
    input?.focus();
  };

  return (
    <FormControl required={component.required}>
      <FormLabel htmlFor={ref}>{t(resolve(component.label)!)}</FormLabel>
      <Box sx={{display: 'flex', gap: 1, justifyContent: 'center', mt: 1}}>
        {otpDigits.map((digit, idx) => (
          <TextField
            // eslint-disable-next-line react/no-array-index-key
            key={`${ref}-otp-${idx}`}
            slotProps={{
              htmlInput: {
                maxLength: 1,
                style: {textAlign: 'center', fontSize: '1.5rem'},
                'aria-label': `OTP digit ${idx + 1}`,
              },
            }}
            value={digit.trim()}
            onChange={(e) => {
              const {value} = e.target;
              if (!/^\d*$/.test(value)) return;
              const newOtp = otpDigits.map((d, i) => (i === idx ? value : d));
              onInputChange(ref, newOtp.join(''));
              if (value && idx < OTP_LENGTH - 1) focusDigit(idx + 1);
            }}
            onKeyDown={(e) => {
              if (e.key === 'Backspace' && !otpDigits[idx].trim() && idx > 0) focusDigit(idx - 1);
            }}
            onPaste={(e) => {
              e.preventDefault();
              const digits = e.clipboardData.getData('text/plain').replace(/\D/g, '').slice(0, OTP_LENGTH);
              onInputChange(ref, digits);
              focusDigit(Math.min(digits.length, OTP_LENGTH - 1));
            }}
            error={hasError}
            disabled={isLoading}
            variant="outlined"
            sx={{width: 48, '& input': {padding: '12px 8px'}}}
          />
        ))}
      </Box>
      {hasError && (
        <Typography variant="caption" color="error" sx={{mt: 0.5, ml: 1.75}}>
          {fieldErrors?.[ref]}
        </Typography>
      )}
    </FormControl>
  );
}
