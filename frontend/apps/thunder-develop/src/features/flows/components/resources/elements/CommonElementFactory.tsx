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

import {BlockTypes, ElementTypes, InputVariants, type Element} from '@/features/flows/models/elements';
import FlowEventTypes from '@/features/flows/models/extension';
import PluginRegistry from '@/features/flows/plugins/PluginRegistry';
import type {ReactElement} from 'react';
import FormAdapter from './adapters/FormAdapter';
import CheckboxAdapter from './adapters/input/CheckboxAdapter';
import PhoneNumberInputAdapter from './adapters/input/PhoneNumberInputAdapter';
import OTPInputAdapter from './adapters/input/OTPInputAdapter';
import DefaultInputAdapter from './adapters/input/DefaultInputAdapter';
import ChoiceAdapter from './adapters/ChoiceAdapter';
import ButtonAdapter from './adapters/ButtonAdapter';
import TypographyAdapter from './adapters/TypographyAdapter';
import DividerAdapter from './adapters/DividerAdapter';
import RichTextAdapter from './adapters/RichTextAdapter';
import ImageAdapter from './adapters/ImageAdapter';
import CaptchaAdapter from './adapters/CaptchaAdapter';
import ResendButtonAdapter from './adapters/ResendButtonAdapter';

/**
 * Props interface of {@link CommonElementFactory}
 */
export interface CommonElementFactoryPropsInterface {
  /**
   * The step id the resource resides on.
   */
  stepId: string;
  /**
   * The element properties.
   */
  resource: Element;
  /**
   * The index of the element in its parent container.
   * Used to trigger handle position updates when elements are reordered.
   */
  elementIndex?: number;
  /**
   * List of available elements that can be added.
   */
  availableElements?: Element[];
  /**
   * Callback for adding an element to a form.
   * @param element - The element to add.
   * @param formId - The ID of the form to add to.
   */
  onAddElementToForm?: (element: Element, formId: string) => void;
}

/**
 * Factory for creating common components.
 *
 * @param props - Props injected to the component.
 * @returns The CommonComponentFactory component.
 */
function CommonElementFactory({
  stepId,
  resource,
  elementIndex,
}: CommonElementFactoryPropsInterface): ReactElement | null {
  const overrideElements: ReactElement[] = [];

  if (
    !PluginRegistry.getInstance().executeSync(FlowEventTypes.ON_NODE_ELEMENT_RENDER, stepId, resource, overrideElements)
  ) {
    if (overrideElements.length > 0) {
      return overrideElements.length === 1 ? overrideElements[0] : <div>{overrideElements}</div>;
    }
  }

  if (resource.type === BlockTypes.Form) {
    return <FormAdapter stepId={stepId} resource={resource} />;
  }
  if (resource.type === ElementTypes.Input) {
    if (resource.variant === InputVariants.Checkbox) {
      return <CheckboxAdapter resource={resource} />;
    }

    if (resource.variant === InputVariants.Telephone) {
      return <PhoneNumberInputAdapter resource={resource} />;
    }

    if (resource.variant === InputVariants.OTP) {
      return <OTPInputAdapter resource={resource} />;
    }

    return <DefaultInputAdapter resource={resource} />;
  }
  if (resource.type === ElementTypes.Choice) {
    return <ChoiceAdapter resource={resource} />;
  }
  if (resource.type === ElementTypes.Button) {
    return <ButtonAdapter resource={resource} elementIndex={elementIndex} />;
  }
  if (resource.type === ElementTypes.Typography) {
    return <TypographyAdapter stepId={stepId} resource={resource} />;
  }
  if (resource.type === ElementTypes.RichText) {
    return <RichTextAdapter resource={resource} />;
  }
  if (resource.type === ElementTypes.Divider) {
    return <DividerAdapter resource={resource} />;
  }
  if (resource.type === ElementTypes.Image) {
    return <ImageAdapter resource={resource} />;
  }
  if (resource.type === ElementTypes.Captcha) {
    return <CaptchaAdapter resource={resource} />;
  }
  if (resource.type === ElementTypes.Resend) {
    return <ResendButtonAdapter stepId={stepId} resource={resource} />;
  }

  return null;
}

export default CommonElementFactory;
