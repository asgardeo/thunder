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

import {type ReactElement, type ComponentType} from 'react';
import {Box} from '@wso2/oxygen-ui';
import * as Icons from '@wso2/oxygen-ui-icons-react';
import type {Element as FlowElement} from '@/features/flows/models/elements';

/**
 * Icon element type with properties at top level.
 */
export type IconElement = FlowElement & {
  name?: string;
  size?: number;
  color?: string;
};

/**
 * Props interface of {@link IconAdapter}
 */
export interface IconAdapterPropsInterface {
  /**
   * The icon element properties.
   */
  resource: FlowElement;
}

/**
 * Adapter for rendering icons from @wso2/oxygen-ui-icons-react (Lucide).
 * The icon is selected by name and rendered at the configured size and color.
 *
 * @param props - Props injected to the component.
 * @returns The IconAdapter component.
 */
function IconAdapter({resource}: IconAdapterPropsInterface): ReactElement {
  const iconElement = resource as IconElement;
  const name = iconElement?.name ?? 'User';
  const size = iconElement?.size ?? 24;
  const color = iconElement?.color ?? 'currentColor';

  const IconComponent = Icons[name as keyof typeof Icons] as ComponentType<{size?: number; color?: string}>;

  if (!IconComponent) {
    return (
      <Box
        display="inline-flex"
        alignItems="center"
        justifyContent="center"
        sx={{
          border: '1px dashed rgba(0, 0, 0, 0.2)',
          borderRadius: 1,
          color: 'text.secondary',
          fontSize: 10,
          height: size,
          padding: '2px 4px',
          width: size,
        }}
      >
        ?
      </Box>
    );
  }

  return <IconComponent size={size} color={color} />;
}

export default IconAdapter;
