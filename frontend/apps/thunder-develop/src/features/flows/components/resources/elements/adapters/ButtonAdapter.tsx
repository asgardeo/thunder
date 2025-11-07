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
import {ButtonVariants, type Element as FlowElement} from '@/features/flows/models/elements';
import {Trans, useTranslation} from 'react-i18next';
import type {RequiredFieldInterface} from '@/features/flows/hooks/useRequiredFields';
import useRequiredFields from '@/features/flows/hooks/useRequiredFields';
import {Button, type ButtonProps, type SxProps, type Theme} from '@wso2/oxygen-ui';
import {Position} from '@xyflow/react';
import VisualFlowConstants from '@/features/flows/constants/VisualFlowConstants';
import resolveStaticResourcePath from '@/features/flows/utils/resolveStaticResourcePath';
import PlaceholderComponent from './PlaceholderComponent';
import NodeHandle from './NodeHandle';

/**
 * Configuration interface for Button element.
 */
interface ButtonConfig {
  styles?: SxProps<Theme>;
  image?: string;
  text?: string;
}

/**
 * Button element type.
 */
export type ButtonElement = FlowElement<ButtonConfig> & {
  variant?: string;
};

/**
 * Props interface of {@link ButtonAdapter}
 */
export interface ButtonAdapterPropsInterface {
  /**
   * The button element properties.
   */
  resource: FlowElement;
  /**
   * The index of the element in its parent container.
   * Used to trigger handle position updates when elements are reordered.
   */
  elementIndex?: number;
}

/**
 * Adapter for the Button component.
 *
 * @param props - Props injected to the component.
 * @returns The ButtonAdapter component.
 */
function ButtonAdapter({resource, elementIndex}: ButtonAdapterPropsInterface): ReactElement {
  const {t} = useTranslation();

  const generalMessage: ReactElement = useMemo(
    () => (
      <Trans i18nKey="flows:core.validation.fields.button.general" values={{id: resource.id}}>
        Required fields are not properly configured for the button with ID <code>{resource.id}</code>.
      </Trans>
    ),
    [resource?.id],
  );

  const fields: RequiredFieldInterface[] = useMemo(
    () => [
      {
        errorMessage: t('flows:core.validation.fields.button.action'),
        name: 'action',
      },
      {
        errorMessage: t('flows:core.validation.fields.button.text'),
        name: 'text',
      },
      {
        errorMessage: t('flows:core.validation.fields.button.variant'),
        name: 'variant',
      },
    ],
    [t],
  );

  useRequiredFields(resource, generalMessage, fields);

  // usePasswordExecutorValidation(resource as unknown as Element);

  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment -- Config type is validated at runtime
  const buttonConfig = resource.config as ButtonConfig | undefined;

  let config: ButtonProps = {};
  let image = '';

  if (resource.variant === ButtonVariants.Primary) {
    config = {
      ...config,
      color: 'primary',
      fullWidth: true,
      variant: 'contained',
    };
  } else if (resource.variant === ButtonVariants.Secondary) {
    config = {
      ...config,
      color: 'secondary',
      fullWidth: true,
      variant: 'contained',
    };
  } else if (resource.variant === ButtonVariants.Text) {
    config = {
      ...config,
      fullWidth: true,
      variant: 'text',
    };
  } else if (resource.variant === ButtonVariants.Social) {
    // TODO: Figure out a way to identify the social connection from the next step.
    image = 'https://www.svgrepo.com/show/475656/google.svg';

    config = {
      ...config,
      fullWidth: true,
      variant: 'outlined',
    };
  }

  return (
    <div className="adapter button-adapter">
      <Button
        sx={buttonConfig?.styles}
        startIcon={
          buttonConfig?.image ? (
            <img src={resolveStaticResourcePath(buttonConfig?.image)} height={20} alt="" />
          ) : (
            image && <img src={resolveStaticResourcePath(image)} height={20} alt="" />
          )
        }
        {...config}
      >
        <PlaceholderComponent value={buttonConfig?.text ?? ''} />
      </Button>
      <NodeHandle
        id={`${resource?.id}${VisualFlowConstants.FLOW_BUILDER_NEXT_HANDLE_SUFFIX}`}
        type="source"
        position={Position.Right}
        positionKey={elementIndex}
      />
    </div>
  );
}

export default ButtonAdapter;
