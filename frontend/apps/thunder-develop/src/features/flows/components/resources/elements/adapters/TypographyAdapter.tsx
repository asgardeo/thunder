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

import {useMemo, type CSSProperties, type ReactElement} from 'react';
import {Trans, useTranslation} from 'react-i18next';
import {Typography, type TypographyProps} from '@wso2/oxygen-ui';
import type {RequiredFieldInterface} from '@/features/flows/hooks/useRequiredFields';
import useRequiredFields from '@/features/flows/hooks/useRequiredFields';
import {TypographyVariants, type Element} from '@/features/flows/models/elements';
import PlaceholderComponent from './PlaceholderComponent';

const TYPOGRAPHY_VALIDATION_FIELD_NAMES = {
  label: 'label',
  variant: 'variant',
} as const;

/**
 * Maps our typography variant names to Material UI typography variant names.
 */
const VARIANT_TO_MUI_MAP: Record<string, TypographyProps['variant']> = {
  [TypographyVariants.H1]: 'h1',
  [TypographyVariants.H2]: 'h2',
  [TypographyVariants.H3]: 'h3',
  [TypographyVariants.H4]: 'h4',
  [TypographyVariants.H5]: 'h5',
  [TypographyVariants.H6]: 'h6',
  [TypographyVariants.Body1]: 'body1',
  [TypographyVariants.Body2]: 'body2',
};

/**
 * Configuration interface for Typography element.
 */
interface TypographyConfig {
  styles?: CSSProperties;
}

/**
 * Typography element with specific variant type.
 */
export interface TypographyElement extends Element<TypographyConfig> {
  variant: (typeof TypographyVariants)[keyof typeof TypographyVariants];
  label?: string;
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
 * @param props - Props injected to the component.
 * @returns The TypographyAdapter component.
 */
function TypographyAdapter({resource}: TypographyAdapterPropsInterface): ReactElement {
  const {t} = useTranslation();

  const generalMessage: ReactElement = useMemo(
    () => (
      <Trans i18nKey="flows:core.validation.fields.typography.general" values={{id: resource.id}}>
        Required fields are not properly configured for the typography with ID <code>{resource.id}</code>.
      </Trans>
    ),
    [resource?.id],
  );

  const validationFields: RequiredFieldInterface[] = useMemo(
    () => [
      {
        errorMessage: t('flows:core.validation.fields.typography.label'),
        name: TYPOGRAPHY_VALIDATION_FIELD_NAMES.label,
      },
      {
        errorMessage: t('flows:core.validation.fields.typography.variant'),
        name: TYPOGRAPHY_VALIDATION_FIELD_NAMES.variant,
      },
    ],
    [t],
  );

  useRequiredFields(resource, generalMessage, validationFields);

  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment -- Config type is validated at runtime
  const typographyConfig = resource.config as TypographyConfig | undefined;
  const typographyElement = resource as TypographyElement;
  const variantStr = resource?.variant as string | undefined;

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

  const muiVariant = variantStr ? VARIANT_TO_MUI_MAP[variantStr] : undefined;

  return (
    <Typography
      variant={muiVariant}
      style={typographyConfig?.styles}
      {...config}
    >
      <PlaceholderComponent value={typographyElement?.label ?? ''} />
    </Typography>
  );
}

export default TypographyAdapter;
