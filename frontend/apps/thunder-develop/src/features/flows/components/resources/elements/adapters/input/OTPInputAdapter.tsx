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

import {useMemo, type CSSProperties, type ReactElement} from 'react';
import type {Element as FlowElement} from '@/features/flows/models/elements';
import {Trans, useTranslation} from 'react-i18next';
import type {RequiredFieldInterface} from '@/features/flows/hooks/useRequiredFields';
import useRequiredFields from '@/features/flows/hooks/useRequiredFields';
import {Box, FormHelperText, InputLabel, OutlinedInput} from '@wso2/oxygen-ui';
import PlaceholderComponent from '../PlaceholderComponent';
import {Hint} from '../../hint';

/**
 * OTP Input element type with properties at top level.
 */
export type OTPInputElement = FlowElement & {
  className?: string;
  label?: string;
  required?: boolean;
  inputType?: string;
  styles?: CSSProperties;
  placeholder?: string;
  hint?: string;
};

/**
 * Props interface of {@link OTPInputAdapter}
 */
export interface OTPInputAdapterPropsInterface {
  /**
   * The OTP input element properties.
   */
  resource: FlowElement;
}

/**
 * Adapter for the OTP inputs.
 *
 * @param props - Props injected to the component.
 * @returns The OTPInputAdapter component.
 */
function OTPInputAdapter({resource}: OTPInputAdapterPropsInterface): ReactElement {
  const {t} = useTranslation();

  const generalMessage: ReactElement = useMemo(
    () => (
      <Trans i18nKey="flows:core.validation.fields.otpInput.general" values={{id: resource.id}}>
        Required fields are not properly configured for the OTP input field with ID <code>{resource.id}</code>.
      </Trans>
    ),
    [resource?.id],
  );

  const fields: RequiredFieldInterface[] = useMemo(
    () => [
      {
        errorMessage: t('flows:core.validation.fields.otpInput.label'),
        name: 'label',
      },
    ],
    [t],
  );

  useRequiredFields(resource, generalMessage, fields);

  const otpElement = resource as OTPInputElement;

  return (
    <div className={otpElement?.className}>
      <InputLabel htmlFor="otp-input-adapter" required={otpElement?.required} disableAnimation>
        <PlaceholderComponent value={otpElement?.label ?? ''} />
      </InputLabel>
      <Box display="flex" flexDirection="row" gap={1}>
        {Array.from({length: 6}, (_, index) => (
          <OutlinedInput
            key={index}
            size="small"
            id="otp-input-adapter"
            type={otpElement?.inputType}
            style={otpElement?.styles}
            placeholder={otpElement?.placeholder ?? ''}
          />
        ))}
      </Box>
      {otpElement?.hint && (
        <FormHelperText>
          <Hint hint={otpElement?.hint} />
        </FormHelperText>
      )}
    </div>
  );
}

export default OTPInputAdapter;
