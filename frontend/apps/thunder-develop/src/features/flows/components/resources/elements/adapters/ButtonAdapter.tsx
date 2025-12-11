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
import {ButtonVariants, type Element as FlowElement} from '@/features/flows/models/elements';
import {Trans} from 'react-i18next';
import type {RequiredFieldInterface} from '@/features/flows/hooks/useRequiredFields';
import useRequiredFields from '@/features/flows/hooks/useRequiredFields';
import {Button, type ButtonProps, type SxProps, type Theme} from '@wso2/oxygen-ui';
import {Position} from '@xyflow/react';
import VisualFlowConstants from '@/features/flows/constants/VisualFlowConstants';
import resolveStaticResourcePath from '@/features/flows/utils/resolveStaticResourcePath';
import PlaceholderComponent from './PlaceholderComponent';
import NodeHandle from './NodeHandle';

// PERFORMANCE: Define fields outside component to prevent recreation on every render
// These are static strings - using plain text instead of dynamic i18n to avoid
// useTranslation hook overhead on every button component
const BUTTON_VALIDATION_FIELDS: RequiredFieldInterface[] = [
  {
    errorMessage: 'Action is required',
    name: 'action',
  },
  {
    errorMessage: 'Text is required',
    name: 'text',
  },
  {
    errorMessage: 'Variant is required',
    name: 'variant',
  },
];

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
 * PERFORMANCE: This component has been optimized to:
 * 1. Use static validation fields defined outside the component
 * 2. Remove useTranslation hook to avoid re-renders
 * 3. Memoize the general message
 *
 * @param props - Props injected to the component.
 * @returns The ButtonAdapter component.
 */
function ButtonAdapter({resource, elementIndex}: ButtonAdapterPropsInterface): ReactElement {
  // PERFORMANCE: Memoize general message - only depends on resource.id
  const generalMessage: ReactElement = useMemo(
    () => (
      <Trans i18nKey="flows:core.validation.fields.button.general" values={{id: resource.id}}>
        Required fields are not properly configured for the button with ID <code>{resource.id}</code>.
      </Trans>
    ),
    [resource?.id],
  );

  // PERFORMANCE: Use static fields array defined outside component
  useRequiredFields(resource, generalMessage, BUTTON_VALIDATION_FIELDS);

  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment -- Config type is validated at runtime
  const buttonConfig = resource.config as ButtonConfig | undefined;

  // PERFORMANCE: Memoize button config to prevent object recreation
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

  // PERFORMANCE: Memoize start icon to prevent recreation
  const startIcon = useMemo(() => {
    if (buttonConfig?.image) {
      return <img src={resolveStaticResourcePath(buttonConfig.image)} height={20} alt="" />;
    }
    if (image) {
      return <img src={resolveStaticResourcePath(image)} height={20} alt="" />;
    }
    return undefined;
  }, [buttonConfig?.image, image]);

  return (
    <div className="adapter button-adapter">
      <Button
        sx={buttonConfig?.styles}
        startIcon={startIcon}
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

// PERFORMANCE: Memoize to prevent re-renders during drag operations
export default memo(ButtonAdapter, (prevProps, nextProps) =>
  prevProps.resource === nextProps.resource &&
  prevProps.elementIndex === nextProps.elementIndex
);
