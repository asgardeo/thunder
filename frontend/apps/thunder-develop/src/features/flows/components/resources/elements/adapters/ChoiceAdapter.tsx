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

import type {ReactElement} from 'react';
import type {Element as FlowElement} from '@/features/flows/models/elements';
import type {FieldOption} from '@/features/flows/models/base';
import {FormControl, FormControlLabel, FormHelperText, FormLabel, Radio, RadioGroup} from '@wso2/oxygen-ui';
import {Hint} from '../hint';

/**
 * Configuration interface for Choice element.
 */
interface ChoiceConfig {
  id?: string;
  defaultValue?: string;
  options?: FieldOption[];
}

/**
 * Choice element type.
 */
export type ChoiceElement = FlowElement<ChoiceConfig>;

/**
 * Props interface of {@link ChoiceAdapter}
 */
export interface ChoiceAdapterPropsInterface {
  /**
   * The choice element properties.
   */
  resource: FlowElement;
}

/**
 * Adapter for the Choice component that renders a radio group.
 *
 * @param props - Props injected to the component.
 * @returns The ChoiceAdapter component.
 */
function ChoiceAdapter({resource}: ChoiceAdapterPropsInterface): ReactElement {
  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment -- Config type is validated at runtime
  const choiceConfig = resource.config as ChoiceConfig | undefined;

  return (
    <FormControl sx={{my: 2}}>
      <FormLabel id={choiceConfig?.id}>{resource.config?.label}</FormLabel>
      <RadioGroup defaultValue={choiceConfig?.defaultValue}>
        {choiceConfig?.options?.map((option: FieldOption) => (
          <FormControlLabel key={option?.key} value={option?.value} control={<Radio />} label={option?.label} />
        ))}
      </RadioGroup>
      {resource.config?.hint && (
        <FormHelperText>
          <Hint hint={resource.config?.hint} />
        </FormHelperText>
      )}
    </FormControl>
  );
}

export default ChoiceAdapter;
