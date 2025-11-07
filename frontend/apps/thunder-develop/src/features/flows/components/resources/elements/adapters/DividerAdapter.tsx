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

import {useMemo, type ReactElement} from 'react';
import {DividerVariants, type Element as FlowElement} from '@/features/flows/models/elements';
import {Trans, useTranslation} from 'react-i18next';
import type {RequiredFieldInterface} from '@/features/flows/hooks/useRequiredFields';
import {Divider, type DividerProps} from '@wso2/oxygen-ui';
import useRequiredFields from '@/features/flows/hooks/useRequiredFields';

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
 * @param props - Props injected to the component.
 * @returns The DividerAdapter component.
 */
function DividerAdapter({resource}: DividerAdapterPropsInterface): ReactElement {
  const {t} = useTranslation();

  const generalMessage: ReactElement = useMemo(
    () => (
      <Trans i18nKey="flows:core.validation.fields.divider.general" values={{id: resource.id}}>
        Required fields are not properly configured for the divider with ID <code>{resource.id}</code>.
      </Trans>
    ),
    [resource?.id],
  );

  const fields: RequiredFieldInterface[] = useMemo(
    () => [
      {
        errorMessage: t('flows:core.validation.fields.divider.variant'),
        name: 'variant',
      },
    ],
    [t],
  );

  useRequiredFields(resource, generalMessage, fields);

  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment -- Config type is validated at runtime
  const dividerConfig = resource.config as DividerConfig | undefined;
  const variantStr = resource?.variant as string | undefined;

  let config: DividerProps = {};

  if (variantStr === DividerVariants.Horizontal || variantStr === DividerVariants.Vertical) {
    config = {
      ...config,
      orientation: variantStr.toLowerCase() as 'horizontal' | 'vertical',
    };
  } else if (variantStr) {
    config = {
      ...config,
      variant: variantStr.toLowerCase() as DividerProps['variant'],
    };
  }

  return <Divider {...config}>{dividerConfig?.text}</Divider>;
}

export default DividerAdapter;
