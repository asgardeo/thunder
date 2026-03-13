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

import {
  CommonStaticStepFactory,
  type CommonStaticStepFactoryPropsInterface,
} from '@/features/flows/components/resources/steps/CommonStaticStepFactory';
import type {Node} from '@xyflow/react';
import type {ReactElement} from 'react';

/**
 * Props interface of {@link StaticStepFactory}
 */
export type StaticStepFactoryPropsInterface = CommonStaticStepFactoryPropsInterface;

/**
 * Factory for creating static steps in the visual editor.
 * Extends the {@link CommonStaticStepFactory} component.
 *
 * @param props - Props injected to the component.
 * @returns The StaticStepFactory component.
 */
function StaticStepFactory({type, ...rest}: StaticStepFactoryPropsInterface & Node): ReactElement {
  return <CommonStaticStepFactory type={type} {...rest} />;
}

export default StaticStepFactory;
