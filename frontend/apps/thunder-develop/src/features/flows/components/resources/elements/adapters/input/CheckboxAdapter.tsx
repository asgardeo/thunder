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
import {Checkbox, FormControlLabel, FormHelperText} from '@wso2/oxygen-ui';
import PlaceholderComponent from '../PlaceholderComponent';
import {Hint} from '../../hint';

/**
 * Configuration interface for Checkbox element.
 */
interface CheckboxConfig {
  className?: string;
  defaultValue?: string;
  label?: string;
  required?: boolean;
  styles?: CSSProperties;
  hint?: string;
}

/**
 * Checkbox element type.
 */
export type CheckboxElement = FlowElement<CheckboxConfig>;

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

  const fields: RequiredFieldInterface[] = useMemo(
    () => [
      {
        errorMessage: t('flows:core.validation.fields.checkbox.label'),
        name: 'label',
      },
      {
        errorMessage: t('flows:core.validation.fields.checkbox.identifier'),
        name: 'identifier',
      },
    ],
    [t],
  );

  useRequiredFields(resource, generalMessage, fields);

  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment -- Config type is validated at runtime
  const checkboxConfig = resource.config as CheckboxConfig | undefined;

  return (
    <div>
      <FormControlLabel
        control={<Checkbox defaultChecked />}
        className={checkboxConfig?.className}
        label={<PlaceholderComponent value={checkboxConfig?.label ?? ''} />}
        required={checkboxConfig?.required}
        style={checkboxConfig?.styles}
      />
      {checkboxConfig?.hint && (
        <FormHelperText>
          <Hint hint={checkboxConfig?.hint} />
        </FormHelperText>
      )}
    </div>
  );
}

export default CheckboxAdapter;
