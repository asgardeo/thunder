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
import {InfoIcon} from '@wso2/oxygen-ui-icons-react';
import {Stack} from '@wso2/oxygen-ui';
import PlaceholderComponent from './adapters/PlaceholderComponent';

/**
 * Props interface of {@link Hint}
 */
export interface HintPropsInterface {
  /**
   * Hint text to be displayed.
   */
  hint: string;
}

/**
 * Hint component to display additional information for input fields.
 *
 * @param props - Props injected to the component.
 * @returns The Hint component.
 */
export function Hint({hint}: HintPropsInterface): ReactElement {
  return (
    <Stack direction="row" gap={0.5} alignItems="center" justifyContent="flex-start">
      <InfoIcon size={12} />
      <PlaceholderComponent value={hint}>
        <span>{hint}</span>
      </PlaceholderComponent>
    </Stack>
  );
}

export default Hint;
