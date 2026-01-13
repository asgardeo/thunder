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

import {BlockTypes, ElementCategories, ElementTypes, type Element} from '@/features/flows/models/elements';
import type {ReactElement} from 'react';
import FormAdapter from './adapters/FormAdapter';
import BlockAdapter from './adapters/BlockAdapter';
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
   * @defaultValue undefined
   */
  elementIndex?: number;
  /**
   * List of available elements that can be added.
   * @defaultValue undefined
   */
  availableElements?: Element[];
  /**
   * Callback for adding an element to a form.
   * @param element - The element to add.
   * @param formId - The ID of the form to add to.
   * @defaultValue undefined
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
  elementIndex = undefined,
  availableElements = undefined,
  onAddElementToForm = undefined,
}: CommonElementFactoryPropsInterface): ReactElement | null {
  if (resource.type === BlockTypes.Form) {
    // Use FormAdapter for blocks with category BLOCK (forms with fields)
    // Use BlockAdapter for blocks with other categories (e.g., ACTION for social buttons)
    if (resource.category === ElementCategories.Block) {
      return (
        <FormAdapter
          stepId={stepId}
          resource={resource}
          availableElements={availableElements}
          onAddElementToForm={onAddElementToForm}
        />
      );
    }

    return (
      <BlockAdapter
        resource={resource}
        availableElements={availableElements}
        onAddElementToForm={onAddElementToForm}
      />
    );
  }
  if (resource.type === ElementTypes.Checkbox) {
    return <CheckboxAdapter resource={resource} />;
  }
  if (resource.type === ElementTypes.PhoneInput) {
    return <PhoneNumberInputAdapter resource={resource} />;
  }
  if (resource.type === ElementTypes.OtpInput) {
    return <OTPInputAdapter resource={resource} />;
  }
  if (
    resource.type === ElementTypes.TextInput ||
    resource.type === ElementTypes.PasswordInput ||
    resource.type === ElementTypes.EmailInput ||
    resource.type === ElementTypes.NumberInput ||
    resource.type === ElementTypes.DateInput
  ) {
    return <DefaultInputAdapter resource={resource} />;
  }
  if (resource.type === ElementTypes.Dropdown) {
    return <ChoiceAdapter resource={resource} />;
  }
  if (resource.type === ElementTypes.Action) {
    return <ButtonAdapter resource={resource} elementIndex={elementIndex} />;
  }
  if (resource.type === ElementTypes.Text) {
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
