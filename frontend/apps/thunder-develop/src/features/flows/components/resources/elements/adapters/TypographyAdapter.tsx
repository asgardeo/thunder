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

import {memo, useMemo, type CSSProperties, type ReactElement} from 'react';
import {Trans} from 'react-i18next';
import type {RequiredFieldInterface} from '@/features/flows/hooks/useRequiredFields';
import useRequiredFields from '@/features/flows/hooks/useRequiredFields';
import {Typography, type TypographyProps} from '@wso2/oxygen-ui';
import {TypographyVariants, type Element} from '@/features/flows/models/elements';
import PlaceholderComponent from './PlaceholderComponent';

// PERFORMANCE: Define fields outside component to prevent recreation on every render
const TYPOGRAPHY_VALIDATION_FIELDS: RequiredFieldInterface[] = [
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
 * Configuration interface for Typography element.
 */
interface TypographyConfig {
  text?: string;
  styles?: CSSProperties;
}

/**
 * Typography element with specific variant type.
 */
export interface TypographyElement extends Element<TypographyConfig> {
  variant: (typeof TypographyVariants)[keyof typeof TypographyVariants];
}

/**
 * Props interface of {@link TypographyAdapter}
 */
export interface TypographyAdapterPropsInterface {
  /**
   * The step id the resource resides on.
   */
  stepId: string;
  /**
   * The typography element properties.
   */
  resource: Element;
}

/**
 * Adapter for the Typography component.
 *
 * PERFORMANCE: This component has been optimized to:
 * 1. Use static validation fields defined outside the component
 * 2. Remove useTranslation hook to avoid re-renders
 * 3. Memoize the general message and config
 *
 * @param props - Props injected to the component.
 * @returns The TypographyAdapter component.
 */
function TypographyAdapter({resource}: TypographyAdapterPropsInterface): ReactElement {
  // PERFORMANCE: Memoize general message - only depends on resource.id
  const generalMessage: ReactElement = useMemo(
    () => (
      <Trans i18nKey="flows:core.validation.fields.typography.general" values={{id: resource.id}}>
        Required fields are not properly configured for the typography with ID <code>{resource.id}</code>.
      </Trans>
    ),
    [resource?.id],
  );

  // PERFORMANCE: Use static fields array defined outside component
  useRequiredFields(resource, generalMessage, TYPOGRAPHY_VALIDATION_FIELDS);

  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment -- Config type is validated at runtime
  const typographyConfig = resource.config as TypographyConfig | undefined;
  const variantStr = resource?.variant as string | undefined;

  // PERFORMANCE: Memoize config object
  const config: TypographyProps = useMemo(() => {
    if (
      variantStr === TypographyVariants.H1 ||
      variantStr === TypographyVariants.H2 ||
      variantStr === TypographyVariants.H3 ||
      variantStr === TypographyVariants.H4 ||
      variantStr === TypographyVariants.H5 ||
      variantStr === TypographyVariants.H6
    ) {
      return {textAlign: 'center'};
    }
    return {};
  }, [variantStr]);

  return (
    <Typography
      variant={variantStr?.toLowerCase() as TypographyProps['variant']}
      style={typographyConfig?.styles}
      {...config}
    >
      <PlaceholderComponent value={typographyConfig?.text ?? ''} />
    </Typography>
  );
}

// PERFORMANCE: Memoize to prevent re-renders during drag operations
export default memo(TypographyAdapter, (prevProps, nextProps) =>
  prevProps.resource === nextProps.resource &&
  prevProps.stepId === nextProps.stepId
);
