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
import {Box, Button} from '@wso2/oxygen-ui';
import {type EmbeddedFlowComponent} from '@asgardeo/react';
import {useTranslation} from 'react-i18next';
import getIntegrationIcon from '../../../utils/getIntegrationIcon';
import type {FlowComponent} from '../../../models/flow';

interface StandaloneTriggerAdapterProps {
  component: FlowComponent;
  index: number;
  isLoading: boolean;
  resolve: (template: string | undefined) => string | undefined;
  onSubmit: (action: EmbeddedFlowComponent, inputs: Record<string, string>) => void;
  values: Record<string, string>;
}

export default function StandaloneTriggerAdapter({
  component,
  index,
  isLoading,
  resolve,
  onSubmit,
  values,
}: StandaloneTriggerAdapterProps): JSX.Element {
  const {t} = useTranslation();
  const resolvedStartIcon = resolve(component.startIcon ?? component.image ?? '');

  const iconElement =
    resolvedStartIcon && /^https?:\/\//i.test(resolvedStartIcon) ? (
      <Box component="img" src={resolvedStartIcon} sx={{width: 20, height: 20, objectFit: 'contain'}} />
    ) : (
      getIntegrationIcon(String(component.label ?? ''), resolvedStartIcon ?? '')
    );

  return (
    <Button
      key={component.id ?? index}
      fullWidth
      variant={component.variant === 'OUTLINED' ? 'outlined' : 'contained'}
      disabled={isLoading}
      startIcon={iconElement}
      onClick={() => onSubmit(component, values)}
      sx={{mt: 1}}
    >
      {t(resolve(component.label)!)}
    </Button>
  );
}
