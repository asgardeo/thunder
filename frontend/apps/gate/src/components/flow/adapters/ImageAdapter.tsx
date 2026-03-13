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

import type {JSX} from 'react';
import {Box} from '@wso2/oxygen-ui';
import type {FlowComponent} from '../../../models/flow';

interface ImageAdapterProps {
  component: FlowComponent;
  resolve: (template: string | undefined) => string | undefined;
  maxWidth?: number | string;
  maxHeight?: number | string;
}

export default function ImageAdapter({component, resolve, maxWidth = '100%', maxHeight = '100%'}: ImageAdapterProps): JSX.Element {
  const resolvedSrc = resolve(component.src ?? '');

  return (
    <Box
      component="img"
      src={resolvedSrc ?? component.src ?? ''}
      alt={component.alt ?? ''}
      sx={{
        width: component.width ? `${component.width}px` : 'auto',
        height: component.height ? `${component.height}px` : 'auto',
        maxWidth,
        maxHeight,
        objectFit: 'contain',
      }}
    />
  );
}
