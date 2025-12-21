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
import {Trans, useTranslation} from 'react-i18next';
import {TextField} from '@wso2/oxygen-ui';
import type {RequiredFieldInterface} from '@/features/flows/hooks/useRequiredFields';
import useRequiredFields from '@/features/flows/hooks/useRequiredFields';
import type {Element as FlowElement} from '@/features/flows/models/elements';
import {Hint} from '../../hint';
import PlaceholderComponent from '../PlaceholderComponent';

const INPUT_VALIDATION_FIELD_NAMES = {
  label: 'label',
  identifier: 'identifier',
} as const;

/**
 * Input element type with properties at top level.
 */
export type InputElement = FlowElement & {
  className?: string;
  defaultValue?: string;
  hint?: string;
  maxLength?: number;
  minLength?: number;
  label?: string;
  multiline?: boolean;
  placeholder?: string;
  required?: boolean;
  inputType?: string;
  styles?: CSSProperties;
};

/**
 * Props interface of {@link DefaultInputAdapter}
 */
export interface DefaultInputAdapterPropsInterface {
  /**
   * The input element properties.
   */
  resource: FlowElement;
}

/**
 * Fallback adapter for the inputs.
 *
 * @param props - Props injected to the component.
 * @returns The DefaultInputAdapter component.
 */
function DefaultInputAdapter({resource}: DefaultInputAdapterPropsInterface): ReactElement {
  const {t} = useTranslation();

  const generalMessage: ReactElement = useMemo(
    () => (
      <Trans i18nKey="flows:core.validation.fields.input.general" values={{id: resource.id}}>
        Required fields are not properly configured for the input field with ID <code>{resource.id}</code>.
      </Trans>
    ),
    [resource?.id],
  );

  const validationFields: RequiredFieldInterface[] = useMemo(
    () => [
      {
        errorMessage: t('flows:core.validation.fields.input.label'),
        name: INPUT_VALIDATION_FIELD_NAMES.label,
      },
      {
        errorMessage: t('flows:core.validation.fields.input.identifier'),
        name: INPUT_VALIDATION_FIELD_NAMES.identifier,
      },
    ],
    [t],
  );

  useRequiredFields(resource, generalMessage, validationFields);

  const inputElement = resource as InputElement;

  return (
    <TextField
      fullWidth
      className={inputElement?.className}
      defaultValue={inputElement?.defaultValue}
      helperText={inputElement?.hint && <Hint hint={inputElement?.hint} />}
      inputProps={{
        maxLength: inputElement?.maxLength,
        minLength: inputElement?.minLength,
      }}
      label={<PlaceholderComponent value={inputElement?.label ?? ''} />}
      multiline={inputElement?.multiline}
      placeholder={inputElement?.placeholder ?? ''}
      required={inputElement?.required}
      InputLabelProps={{
        required: inputElement?.required,
      }}
      type={inputElement?.inputType}
      style={inputElement?.styles}
      autoComplete={inputElement?.inputType === 'password' ? 'new-password' : 'off'}
    />
  );
}

export default DefaultInputAdapter;
