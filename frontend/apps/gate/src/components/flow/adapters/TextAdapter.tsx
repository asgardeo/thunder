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
import {Typography} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';
import {mapEmbeddedFlowTextVariant, useDesign} from '@thunder/shared-design';
import type {FlowComponent} from '../../../models/flow';

interface TextAdapterProps {
  component: FlowComponent;
  resolve: (template: string | undefined) => string | undefined;
}

export default function TextAdapter({component, resolve}: TextAdapterProps): JSX.Element {
  const {t} = useTranslation();
  const {isDesignEnabled} = useDesign();

  return (
    <Typography
      variant={mapEmbeddedFlowTextVariant(component.variant)}
      sx={{mb: 1, textAlign: isDesignEnabled ? 'center' : 'left'}}
    >
      {t(resolve(component.label)!)}
    </Typography>
  );
}
