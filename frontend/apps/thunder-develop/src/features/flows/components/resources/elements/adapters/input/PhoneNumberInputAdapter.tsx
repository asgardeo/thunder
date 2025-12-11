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

import {useMemo, type ReactElement} from 'react';
import type {Element as FlowElement} from '@/features/flows/models/elements';
import {Trans, useTranslation} from 'react-i18next';
import type {RequiredFieldInterface} from '@/features/flows/hooks/useRequiredFields';
import useRequiredFields from '@/features/flows/hooks/useRequiredFields';
import {FormHelperText, TextField} from '@wso2/oxygen-ui';
import {Hint} from '../../hint';

/**
 * Configuration interface for Phone Number Input element.
 */
interface PhoneNumberInputConfig {
  className?: string;
  label?: string;
  placeholder?: string;
  required?: boolean;
  hint?: string;
}

/**
 * Phone Number Input element type.
 */
export type PhoneNumberInputElement = FlowElement<PhoneNumberInputConfig>;

/**
 * Props interface of {@link PhoneNumberInputAdapter}
 */
export interface PhoneNumberInputAdapterPropsInterface {
  /**
   * The phone number input element properties.
   */
  resource: FlowElement;
}

/**
 * Adapter for the Phone Number input component.
 *
 * @param props - Props injected to the component.
 * @returns The PhoneNumberInputAdapter component.
 */
function PhoneNumberInputAdapter({resource}: PhoneNumberInputAdapterPropsInterface): ReactElement {
  const {t} = useTranslation();

  const generalMessage: ReactElement = useMemo(
    () => (
      <Trans i18nKey="flows:core.validation.fields.phoneNumberInput.general" values={{id: resource.id}}>
        Required fields are not properly configured for the phone number field with ID <code>{resource.id}</code>.
      </Trans>
    ),
    [resource?.id],
  );

  const fields: RequiredFieldInterface[] = useMemo(
    () => [
      {
        errorMessage: t('flows:core.validation.fields.phoneNumberInput.label'),
        name: 'label',
      },
      {
        errorMessage: t('flows:core.validation.fields.phoneNumberInput.identifier'),
        name: 'identifier',
      },
    ],
    [t],
  );

  useRequiredFields(resource, generalMessage, fields);

  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment -- Config type is validated at runtime
  const phoneConfig = resource.config as PhoneNumberInputConfig | undefined;

  return (
    <>
      <TextField
        className={phoneConfig?.className}
        label={phoneConfig?.label}
        placeholder={phoneConfig?.placeholder ?? ''}
        InputLabelProps={{
          required: phoneConfig?.required,
        }}
        type="number"
      />
      {phoneConfig?.hint && (
        <FormHelperText>
          <Hint hint={phoneConfig?.hint} />
        </FormHelperText>
      )}
    </>
  );
}

export default PhoneNumberInputAdapter;
