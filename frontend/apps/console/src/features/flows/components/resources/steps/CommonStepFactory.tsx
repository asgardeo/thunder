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

import {StepTypes, type Step} from '@/features/flows/models/steps';
import type {ReactElement} from 'react';
import type {NodeProps} from '@xyflow/react';
import View from './view/View';
import End from './end/End';
import Execution from './execution/Execution';
import Rule from './rule/Rule';
import type {Resources} from '../../../models/resources';
import type {Element} from '../../../models/elements';

/**
 * Props interface of {@link CommonStepFactory}
 */
export interface CommonStepFactoryPropsInterface extends NodeProps {
  /**
   * The flow id of the resource.
   */
  resourceId: string;
  /**
   * All the resources corresponding to the type.
   */
  resources: Step[];
  /**
   * All available resources in the flow.
   * @defaultValue undefined
   */
  allResources?: Resources;
  /**
   * Callback for adding an element to the view.
   * @defaultValue undefined
   */
  onAddElement?: (element: Element) => void;
  /**
   * Callback for adding an element to a form.
   * @param element - The element to add.
   * @param formId - The ID of the form to add to.
   * @defaultValue undefined
   */
  onAddElementToForm?: (element: Element, formId: string) => void;
}

/**
 * Factory for creating common steps.
 *
 * @param props - Props injected to the component.
 * @returns The CommonStepFactory component.
 */
function CommonStepFactory({
  resources,
  data,
  allResources = undefined,
  onAddElement = undefined,
  onAddElementToForm = undefined,
  ...rest
}: CommonStepFactoryPropsInterface): ReactElement | null {
  if (resources && resources[0].type === StepTypes.View) {
    return (
      <View
        resources={resources}
        data={data}
        availableElements={allResources?.elements}
        onAddElement={onAddElement}
        onAddElementToForm={onAddElementToForm}
        {...rest}
      />
    );
  }

  if (resources[0].type === StepTypes.Rule) {
    return <Rule resources={resources} data={data} {...rest} />;
  }

  if (resources[0].type === StepTypes.Execution) {
    return <Execution resources={resources} data={data} {...rest} />;
  }

  if (resources && resources[0].type === StepTypes.End) {
    return <End resources={resources} data={data} {...rest} />;
  }

  return null;
}

export default CommonStepFactory;
