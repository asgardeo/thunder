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

import isEmpty from 'lodash-es/isEmpty';
import type {Element} from '../models/elements';
import generateResourceId from '../utils/generateResourceId';
/**
 * Props interface of {@link useGenerateStepElement}
 */
export interface UseGenerateStepElement {
  generateStepElement: (element: Element) => Element;
}

/**
 * Hook that provides a function to generate a step element with a unique ID.
 *
 * This hook allows components to generate a step element with a unique ID and default variant if applicable.
 *
 * @returns An object containing the `generateStepElement` function.
 *
 * @example
 * ```tsx
 * const { generateStepElement } = useGenerateStepElement();
 * const element = generateStepElement({ category: "ACTION", variants: [...] });
 * ```
 */
const useGenerateStepElement = (): UseGenerateStepElement => {
  const generateStepElement = (element: Element): Element => {
    let updatedElement: Element = {
      ...element,
      id: generateResourceId(element.category.toLowerCase()),
    };

    // If the component has variants, add the default variant to the root.
    if (!isEmpty(updatedElement?.variants)) {
      const defaultVariantType: string =
        updatedElement?.display?.defaultVariant ?? (updatedElement.variants?.[0]?.variant as string);
      const defaultVariant: Element | undefined = updatedElement.variants?.find(
        (variant: Element) => variant.variant === defaultVariantType,
      );

      if (defaultVariant) {
        updatedElement = {
          ...updatedElement,
          ...defaultVariant,
        };
      }
    }

    return updatedElement;
  };

  return {
    generateStepElement,
  };
};

export default useGenerateStepElement;
