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

import cloneDeep from 'lodash-es/cloneDeep';
import {
  BlockTypes,
  ButtonTypes,
  ButtonVariants,
  ElementCategories,
  ElementTypes,
  type Element,
} from '@/features/flows/models/elements';
import LoginFlowConstants from '../constants/LoginFlowConstants';

const {ExecutorNames} = LoginFlowConstants;

/**
 * Set of input element types for quick lookup.
 * Used to determine if an element should be added to a form.
 */
export const INPUT_ELEMENT_TYPES = new Set<string>([
  ElementTypes.TextInput,
  ElementTypes.PasswordInput,
  ElementTypes.EmailInput,
  ElementTypes.PhoneInput,
  ElementTypes.NumberInput,
  ElementTypes.DateInput,
  ElementTypes.OtpInput,
  ElementTypes.Checkbox,
  ElementTypes.Dropdown,
]);

/**
 * Process form components to set button types and auto-assign executors.
 * Optimized to use a single pass through the components.
 *
 * @param formComponents - The form components to process.
 * @returns The processed form components with proper button types and executors.
 */
export const processFormComponents = (formComponents: Element[] | undefined): Element[] | undefined => {
  if (!formComponents || formComponents.length === 0) {
    return formComponents;
  }

  // Single pass to collect information and transform
  let hasPasswordField = false;
  let hasOtpField = false;
  let submitButtonCount = 0;

  // First pass: collect info and set PRIMARY buttons to submit
  const updatedComponents = formComponents.map((formComponent: Element) => {
    // Check for field types
    if (formComponent.type === ElementTypes.PasswordInput) {
      hasPasswordField = true;
    } else if (formComponent.type === ElementTypes.OtpInput) {
      hasOtpField = true;
    }

    // Set PRIMARY buttons to submit type
    if (formComponent.type === ElementTypes.Action && formComponent.variant === ButtonVariants.Primary) {
      const updatedButton = {
        ...formComponent,
        config: {
          ...formComponent.config,
          type: ButtonTypes.Submit,
        },
      };
      submitButtonCount += 1;
      return updatedButton;
    }

    // Count existing submit buttons
    if (
      formComponent.type === ElementTypes.Action &&
      (formComponent.config as {type?: string})?.type === ButtonTypes.Submit
    ) {
      submitButtonCount += 1;
    }

    return formComponent;
  });

  // If exactly one submit button and has password/otp field, assign executor
  if (submitButtonCount === 1 && (hasPasswordField || hasOtpField)) {
    const executorName = hasPasswordField ? ExecutorNames.PASSWORD_PROVISIONING : ExecutorNames.EMAIL_OTP;

    return updatedComponents.map((formComponent: Element) => {
      if (formComponent.type === ElementTypes.Action) {
        return {
          ...formComponent,
          action: {
            ...(formComponent?.action ?? {}),
            executor: {name: executorName},
            type: LoginFlowConstants.ActionTypes.EXECUTOR,
          },
        };
      }
      return formComponent;
    });
  }

  return updatedComponents;
};

/**
 * Mutate components to ensure proper form structure and button actions.
 *
 * Optimizations:
 * - Single pass for filtering and counting forms
 * - Separated form processing logic for clarity
 * - Uses typed constants for executor names
 *
 * @param components - The components to mutate.
 * @returns The mutated components with proper form structure.
 */
export const mutateComponents = (components: Element[]): Element[] => {
  // Clone and filter in single pass, tracking form count
  let firstFormFound = false;

  const modifiedComponents = cloneDeep(components).filter((component) => {
    // Filter out non-element resources
    if (component.resourceType && component.resourceType !== 'ELEMENT') {
      return false;
    }

    // Keep only the first form (category: BLOCK, type: BLOCK)
    // Note: Social login blocks have category: ACTION, type: BLOCK, so they should not be filtered
    if (component.type === BlockTypes.Form && component.category === ElementCategories.Block) {
      if (firstFormFound) {
        return false;
      }
      firstFormFound = true;
    }

    return true;
  });

  // Process forms and their components (only actual forms with category: BLOCK)
  return modifiedComponents.map((component: Element) => {
    if (component.type === BlockTypes.Form && component.category === ElementCategories.Block) {
      return {
        ...component,
        components: processFormComponents(component.components),
      };
    }
    return component;
  });
};
