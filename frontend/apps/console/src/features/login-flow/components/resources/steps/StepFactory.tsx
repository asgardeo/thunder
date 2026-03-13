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

import CommonStepFactory, {
  type CommonStepFactoryPropsInterface,
} from '@/features/flows/components/resources/steps/CommonStepFactory';
import type {Node} from '@xyflow/react';
import {memo, type ReactElement} from 'react';
import type {Resources} from '@/features/flows/models/resources';
import type {Element} from '@/features/flows/models/elements';

/**
 * Props interface of {@link StepFactory}
 */
export interface StepFactoryPropsInterface extends CommonStepFactoryPropsInterface {
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
 * Factory for creating steps.
 * Extends the {@link CommonStepFactory} component.
 *
 * @param props - Props injected to the component.
 * @returns The StepFactory component.
 */
function StepFactory({
  resourceId,
  resources,
  allResources = undefined,
  onAddElement = undefined,
  onAddElementToForm = undefined,
  ...rest
}: StepFactoryPropsInterface & Node): ReactElement {
  return (
    <CommonStepFactory
      resourceId={resourceId}
      resources={resources}
      allResources={allResources}
      onAddElement={onAddElement}
      onAddElementToForm={onAddElementToForm}
      {...rest}
    />
  );
}

// Memoize to prevent re-renders during drag operations
export default memo(StepFactory, (prevProps, nextProps) =>
  prevProps.id === nextProps.id &&
  prevProps.data === nextProps.data &&
  prevProps.resources === nextProps.resources &&
  prevProps.allResources === nextProps.allResources &&
  prevProps.onAddElement === nextProps.onAddElement &&
  prevProps.onAddElementToForm === nextProps.onAddElementToForm
);
