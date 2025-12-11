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

import {memo, useMemo, type ReactElement} from 'react';
import {DividerVariants, type Element as FlowElement} from '@/features/flows/models/elements';
import {Trans} from 'react-i18next';
import type {RequiredFieldInterface} from '@/features/flows/hooks/useRequiredFields';
import {Divider, type DividerProps} from '@wso2/oxygen-ui';
import useRequiredFields from '@/features/flows/hooks/useRequiredFields';

// PERFORMANCE: Define fields outside component to prevent recreation on every render
const DIVIDER_VALIDATION_FIELDS: RequiredFieldInterface[] = [
  {
    errorMessage: 'Variant is required',
    name: 'variant',
  },
];

/**
 * Configuration interface for Divider element.
 */
interface DividerConfig {
  text?: string;
}

/**
 * Divider element type.
 */
export type DividerElement = FlowElement<DividerConfig> & {
  variant?: string;
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
 * PERFORMANCE: This component has been optimized to:
 * 1. Use static validation fields defined outside the component
 * 2. Remove useTranslation hook to avoid re-renders
 * 3. Memoize the general message and config
 *
 * @param props - Props injected to the component.
 * @returns The DividerAdapter component.
 */
function DividerAdapter({resource}: DividerAdapterPropsInterface): ReactElement {
  // PERFORMANCE: Memoize general message - only depends on resource.id
  const generalMessage: ReactElement = useMemo(
    () => (
      <Trans i18nKey="flows:core.validation.fields.divider.general" values={{id: resource.id}}>
        Required fields are not properly configured for the divider with ID <code>{resource.id}</code>.
      </Trans>
    ),
    [resource?.id],
  );

  // PERFORMANCE: Use static fields array defined outside component
  useRequiredFields(resource, generalMessage, DIVIDER_VALIDATION_FIELDS);

  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment -- Config type is validated at runtime
  const dividerConfig = resource.config as DividerConfig | undefined;
  const variantStr = resource?.variant as string | undefined;

  // PERFORMANCE: Memoize config object
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

  return <Divider {...config}>{dividerConfig?.text}</Divider>;
}

// PERFORMANCE: Memoize to prevent re-renders during drag operations
export default memo(DividerAdapter, (prevProps, nextProps) =>
  prevProps.resource === nextProps.resource
);
