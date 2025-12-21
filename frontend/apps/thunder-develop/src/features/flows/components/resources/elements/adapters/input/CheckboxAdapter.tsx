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
import {Checkbox, FormControlLabel, FormHelperText} from '@wso2/oxygen-ui';
import type {RequiredFieldInterface} from '@/features/flows/hooks/useRequiredFields';
import useRequiredFields from '@/features/flows/hooks/useRequiredFields';
import type {Element as FlowElement} from '@/features/flows/models/elements';
import {Hint} from '../../hint';
import PlaceholderComponent from '../PlaceholderComponent';

const CHECKBOX_VALIDATION_FIELD_NAMES = {
  label: 'label',
  identifier: 'identifier',
} as const;

/**
 * Checkbox element type with properties at top level.
 */
export type CheckboxElement = FlowElement & {
  className?: string;
  defaultValue?: string;
  label?: string;
  required?: boolean;
  styles?: CSSProperties;
  hint?: string;
};

/**
 * Props interface of {@link CheckboxAdapter}
 */
export interface CheckboxAdapterPropsInterface {
  /**
   * The checkbox element properties.
   */
  resource: FlowElement;
}

/**
 * Adapter for the Checkbox component.
 *
 * @param props - Props injected to the component.
 * @returns The CheckboxAdapter component.
 */
function CheckboxAdapter({resource}: CheckboxAdapterPropsInterface): ReactElement {
  const {t} = useTranslation();

  const generalMessage: ReactElement = useMemo(
    () => (
      <Trans i18nKey="flows:core.validation.fields.checkbox.general" values={{id: resource.id}}>
        Required fields are not properly configured for the checkbox field with ID <code>{resource.id}</code>.
      </Trans>
    ),
    [resource?.id],
  );

  const validationFields: RequiredFieldInterface[] = useMemo(
    () => [
      {
        errorMessage: t('flows:core.validation.fields.checkbox.label'),
        name: CHECKBOX_VALIDATION_FIELD_NAMES.label,
      },
      {
        errorMessage: t('flows:core.validation.fields.checkbox.identifier'),
        name: CHECKBOX_VALIDATION_FIELD_NAMES.identifier,
      },
    ],
    [t],
  );

  useRequiredFields(resource, generalMessage, validationFields);

  const checkboxElement = resource as CheckboxElement;

  return (
    <div>
      <FormControlLabel
        control={<Checkbox defaultChecked />}
        className={checkboxElement?.className}
        label={<PlaceholderComponent value={checkboxElement?.label ?? ''} />}
        required={checkboxElement?.required}
        style={checkboxElement?.styles}
      />
      {checkboxElement?.hint && (
        <FormHelperText>
          <Hint hint={checkboxElement?.hint} />
        </FormHelperText>
      )}
    </div>
  );
}

export default CheckboxAdapter;
