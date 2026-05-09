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

import {useTemplateLiteralResolver} from '@thunderid/hooks';
import {FormControl, FormLabel, MenuItem, Select} from '@wso2/oxygen-ui';
import type {ReactElement} from 'react';
import {useTranslation} from 'react-i18next';
import {Hint} from '../../hint';
import type {Element as FlowElement} from '@/features/flows/models/elements';

/**
 * Select element type with properties at top level.
 */
export type SelectElement = FlowElement & {
  hint?: string;
  label?: string;
  placeholder?: string;
  required?: boolean;
  options?: unknown[];
};

/**
 * Props interface of {@link SelectAdapter}
 */
export interface SelectAdapterPropsInterface {
  /**
   * The select element properties.
   */
  resource: FlowElement;
}

/**
 * Adapter for SELECT input elements in the flow builder.
 * Renders a dropdown preview. Options are populated dynamically
 * at runtime via ForwardedData from upstream executors.
 *
 * @param props - Props injected to the component.
 * @returns The SelectAdapter component.
 */
function SelectAdapter({resource}: SelectAdapterPropsInterface): ReactElement {
  const {t} = useTranslation();
  const {resolve} = useTemplateLiteralResolver();

  const selectElement = resource as SelectElement;
  const label = resolve(selectElement?.label, {t}) ?? selectElement?.label ?? '';
  const placeholder = resolve(selectElement?.placeholder, {t}) ?? selectElement?.placeholder ?? '';

  return (
    <FormControl fullWidth>
      <FormLabel required={selectElement?.required}>{label}</FormLabel>
      <Select displayEmpty size="small" fullWidth value="">
        <MenuItem value="" disabled>
          {placeholder}
        </MenuItem>
      </Select>
      {selectElement?.hint && <Hint hint={selectElement.hint} />}
    </FormControl>
  );
}

export default SelectAdapter;
