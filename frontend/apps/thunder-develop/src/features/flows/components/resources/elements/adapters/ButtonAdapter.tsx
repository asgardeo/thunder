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
import {Button, type ButtonProps, type SxProps, type Theme} from '@wso2/oxygen-ui';
import {Position} from '@xyflow/react';
import type {RequiredFieldInterface} from '@/features/flows/hooks/useRequiredFields';
import useRequiredFields from '@/features/flows/hooks/useRequiredFields';
import {ButtonVariants, type Element as FlowElement} from '@/features/flows/models/elements';
import VisualFlowConstants from '@/features/flows/constants/VisualFlowConstants';
import resolveStaticResourcePath from '@/features/flows/utils/resolveStaticResourcePath';
import PlaceholderComponent from './PlaceholderComponent';
import NodeHandle from './NodeHandle';

const BUTTON_VALIDATION_FIELD_NAMES = {
  label: 'label',
  variant: 'variant',
} as const;

/**
 * Configuration interface for Button element.
 */
interface ButtonConfig {
  styles?: SxProps<Theme>;
  image?: string;
}

/**
 * Button element type.
 */
export type ButtonElement = FlowElement<ButtonConfig> & {
  variant?: string;
  label?: string;
  image?: string;
  startIcon?: string;
  endIcon?: string;
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
   * @defaultValue undefined
   */
  elementIndex?: number;
}

/**
 * Adapter for the Button component.
 *
 * @param props - Props injected to the component.
 * @returns The ButtonAdapter component.
 */
function ButtonAdapter({resource, elementIndex = undefined}: ButtonAdapterPropsInterface): ReactElement {
  const {t} = useTranslation();

  const generalMessage: ReactElement = useMemo(
    () => (
      <Trans i18nKey="flows:core.validation.fields.button.general" values={{id: resource.id}}>
        Required fields are not properly configured for the button with ID <code>{resource.id}</code>.
      </Trans>
    ),
    [resource?.id],
  );

  const validationFields: RequiredFieldInterface[] = useMemo(
    () => [
      {
        errorMessage: t('flows:core.validation.fields.button.label'),
        name: BUTTON_VALIDATION_FIELD_NAMES.label,
      },
      {
        errorMessage: t('flows:core.validation.fields.button.variant'),
        name: BUTTON_VALIDATION_FIELD_NAMES.variant,
      },
    ],
    [t],
  );

  useRequiredFields(resource, generalMessage, validationFields);

  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment -- Config type is validated at runtime
  const buttonConfig = resource.config as ButtonConfig | undefined;

  const {config, image} = useMemo(() => {
    let buttonProps: ButtonProps = {};
    let buttonImage = '';

    if (resource.variant === ButtonVariants.Primary) {
      buttonProps = {
        color: 'primary',
        fullWidth: true,
        variant: 'contained',
      };
    } else if (resource.variant === ButtonVariants.Secondary) {
      buttonProps = {
        color: 'secondary',
        fullWidth: true,
        variant: 'contained',
      };
    } else if (resource.variant === ButtonVariants.Text) {
      buttonProps = {
        fullWidth: true,
        variant: 'text',
      };
    } else if (resource.variant === ButtonVariants.Social) {
      buttonImage = 'https://www.svgrepo.com/show/475656/google.svg';
      buttonProps = {
        fullWidth: true,
        variant: 'outlined',
      };
    }

    return {config: buttonProps, image: buttonImage};
  }, [resource.variant]);

  // Cast resource to ButtonElement to access label and image properties
  const buttonElement = resource as ButtonElement;

  const startIcon = useMemo(() => {
    // Check resource.startIcon first (new format), then resource.image for backwards compatibility,
    // then config.image, then variant default
    if (buttonElement?.startIcon) {
      return <img src={resolveStaticResourcePath(buttonElement.startIcon)} height={20} alt="" />;
    }
    if (buttonElement?.image) {
      return <img src={resolveStaticResourcePath(buttonElement.image)} height={20} alt="" />;
    }
    if (buttonConfig?.image) {
      return <img src={resolveStaticResourcePath(buttonConfig.image)} height={20} alt="" />;
    }
    if (image) {
      return <img src={resolveStaticResourcePath(image)} height={20} alt="" />;
    }
    return undefined;
  }, [buttonElement?.startIcon, buttonElement?.image, buttonConfig?.image, image]);

  const endIcon = useMemo(() => {
    if (buttonElement?.endIcon) {
      return <img src={resolveStaticResourcePath(buttonElement.endIcon)} height={20} alt="" />;
    }
    return undefined;
  }, [buttonElement?.endIcon]);

  return (
    <div className="adapter button-adapter">
      <Button
        sx={buttonConfig?.styles}
        startIcon={startIcon}
        endIcon={endIcon}
        {...config}
      >
        <PlaceholderComponent value={buttonElement?.label ?? ''} />
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
