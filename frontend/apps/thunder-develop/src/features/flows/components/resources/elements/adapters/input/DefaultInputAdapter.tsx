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
import {TextField} from '@wso2/oxygen-ui';
import PlaceholderComponent from '../PlaceholderComponent';
import {Hint} from '../../hint';

/**
 * Configuration interface for Input element.
 */
interface InputConfig {
  className?: string;
  defaultValue?: string;
  hint?: string;
  maxLength?: number;
  minLength?: number;
  label?: string;
  multiline?: boolean;
  placeholder?: string;
  required?: boolean;
  type?: string;
  styles?: CSSProperties;
}

/**
 * Input element type.
 */
export type InputElement = FlowElement<InputConfig>;

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

  const fields: RequiredFieldInterface[] = useMemo(
    () => [
      {
        errorMessage: t('flows:core.validation.fields.input.label'),
        name: 'label',
      },
      {
        errorMessage: t('flows:core.validation.fields.input.identifier'),
        name: 'identifier',
      },
    ],
    [t],
  );

  useRequiredFields(resource, generalMessage, fields);

  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment -- Config type is validated at runtime
  const inputConfig = resource.config as InputConfig | undefined;

  return (
    <TextField
      fullWidth
      className={inputConfig?.className}
      defaultValue={inputConfig?.defaultValue}
      helperText={inputConfig?.hint && <Hint hint={inputConfig?.hint} />}
      inputProps={{
        maxLength: inputConfig?.maxLength,
        minLength: inputConfig?.minLength,
      }}
      label={<PlaceholderComponent value={inputConfig?.label ?? ''} />}
      multiline={inputConfig?.multiline}
      placeholder={inputConfig?.placeholder ?? ''}
      required={inputConfig?.required}
      InputLabelProps={{
        required: inputConfig?.required,
      }}
      type={inputConfig?.type}
      style={inputConfig?.styles}
      autoComplete={inputConfig?.type === 'password' ? 'new-password' : 'off'}
    />
  );
}

export default DefaultInputAdapter;
