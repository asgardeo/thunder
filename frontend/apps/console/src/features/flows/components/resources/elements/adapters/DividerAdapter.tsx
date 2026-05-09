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

import {useTemplateLiteralResolver} from '@thunderid/hooks';
import {Divider, type DividerProps} from '@wso2/oxygen-ui';
import {useMemo, type ReactElement} from 'react';
import {useTranslation} from 'react-i18next';
import {DividerVariants, type Element as FlowElement} from '@/features/flows/models/elements';

/**
 * Divider element type.
 */
export type DividerElement = FlowElement & {
  variant?: string;
  label?: string;
};

/**
 * Props interface of {@link DividerAdapter}
 */
export interface DividerAdapterPropsInterface {
  /**
   * The divider element properties.
   */
  resource: FlowElement;
}

/**
 * Adapter for the Divider component.
 *
 * @param props - Props injected to the component.
 * @returns The DividerAdapter component.
 */
function DividerAdapter({resource}: DividerAdapterPropsInterface): ReactElement {
  const {t} = useTranslation();
  const {resolve} = useTemplateLiteralResolver();

  const dividerElement = resource as DividerElement;
  const variantStr = resource?.variant as string | undefined;

  const config: DividerProps = useMemo(() => {
    if (variantStr === DividerVariants.Horizontal || variantStr === DividerVariants.Vertical) {
      return {
        orientation: variantStr.toLowerCase() as 'horizontal' | 'vertical',
      };
    }
    if (variantStr) {
      return {
        variant: variantStr.toLowerCase() as DividerProps['variant'],
      };
    }
    return {};
  }, [variantStr]);

  return (
    <Divider {...config}>
      {dividerElement?.label && (resolve(dividerElement.label, {t}) ?? dividerElement.label)}
    </Divider>
  );
}

export default DividerAdapter;
