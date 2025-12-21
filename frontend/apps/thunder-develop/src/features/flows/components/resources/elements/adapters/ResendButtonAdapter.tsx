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
import {Trans, useTranslation} from 'react-i18next';
import type {RequiredFieldInterface} from '@/features/flows/hooks/useRequiredFields';
import useRequiredFields from '@/features/flows/hooks/useRequiredFields';
import {Button, type ButtonProps, type SxProps, type Theme} from '@wso2/oxygen-ui';
import {Position} from '@xyflow/react';
import type {Element as FlowElement} from '@/features/flows/models/elements';
import VisualFlowConstants from '@/features/flows/constants/VisualFlowConstants';
import PlaceholderComponent from './PlaceholderComponent';
import NodeHandle from './NodeHandle';

/**
 * Configuration interface for ResendButton element.
 */
interface ResendButtonConfig {
  styles?: SxProps<Theme>;
}

/**
 * ResendButton element type.
 */
export type ResendButtonElement = FlowElement<ResendButtonConfig> & {
  label?: string;
};

/**
 * Props interface of {@link ResendButtonAdapter}
 */
export interface ResendButtonAdapterPropsInterface {
  /**
   * The step id the resource resides on.
   */
  stepId: string;
  /**
   * The resend button element properties.
   */
  resource: FlowElement;
}

/**
 * Adapter for the ResendButton component.
 *
 * @param props - Props injected to the component.
 * @returns The ResendButtonAdapter component.
 */
function ResendButtonAdapter({resource}: ResendButtonAdapterPropsInterface): ReactElement {
  const {t} = useTranslation();

  const generalMessage: ReactElement = useMemo(
    () => (
      <Trans i18nKey="flows:core.validation.fields.button.general" values={{id: resource.id}}>
        Required fields are not properly configured for the resend button with ID <code>{resource.id}</code>.
      </Trans>
    ),
    [resource.id],
  );

  const fields: RequiredFieldInterface[] = useMemo(
    () => [
      {
        errorMessage: t('flows:core.validation.fields.button.label'),
        name: 'label',
      },
    ],
    [t],
  );

  useRequiredFields(resource, generalMessage, fields);

  const config: ButtonProps = {
    color: 'secondary',
    fullWidth: true,
    variant: 'contained',
  };

  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment -- Config type is validated at runtime
  const resendConfig = resource.config as ResendButtonConfig | undefined;
  const resendElement = resource as ResendButtonElement;

  return (
    <div className="adapter button-adapter">
      <Button sx={resendConfig?.styles} {...config}>
        <PlaceholderComponent value={resendElement?.label ?? ''} />
      </Button>
      <NodeHandle
        id={`${resource?.id}${VisualFlowConstants.FLOW_BUILDER_NEXT_HANDLE_SUFFIX}`}
        type="source"
        position={Position.Right}
      />
    </div>
  );
}

export default ResendButtonAdapter;
