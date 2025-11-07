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
import type {RequiredFieldInterface} from '@/features/flows/hooks/useRequiredFields';
import useRequiredFields from '@/features/flows/hooks/useRequiredFields';
import {Box} from '@wso2/oxygen-ui';
import type {Element as FlowElement} from '@/features/flows/models/elements';

/**
 * Configuration interface for Image element.
 */
interface ImageConfig {
  src?: string;
  alt?: string;
  styles?: CSSProperties;
}

/**
 * Image element type.
 */
export type ImageElement = FlowElement<ImageConfig>;

/**
 * Props interface of {@link ImageAdapter}
 */
export interface ImageAdapterPropsInterface {
  /**
   * The image element properties.
   */
  resource: FlowElement;
}

/**
 * Adapter for displaying images.
 *
 * @param props - Props injected to the component.
 * @returns The ImageAdapter component.
 */
function ImageAdapter({resource}: ImageAdapterPropsInterface): ReactElement {
  const {t} = useTranslation();

  const generalMessage: ReactElement = useMemo(
    () => (
      <Trans i18nKey="flows:core.validation.fields.image.general" values={{id: resource.id}}>
        Required fields are not properly configured for the image with ID <code>{resource.id}</code>.
      </Trans>
    ),
    [resource.id],
  );

  const fields: RequiredFieldInterface[] = useMemo(
    () => [
      {
        errorMessage: t('flows:core.validation.fields.image.src'),
        name: 'src',
      },
      {
        errorMessage: t('flows:core.validation.fields.image.variant'),
        name: 'variant',
      },
    ],
    [t],
  );

  useRequiredFields(resource, generalMessage, fields);

  // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment -- Config type is validated at runtime
  const imageConfig = resource.config as ImageConfig | undefined;

  return (
    <Box display="flex" alignItems="center" justifyContent="center">
      <img src={imageConfig?.src} alt={imageConfig?.alt} width="100%" style={imageConfig?.styles} />
    </Box>
  );
}

export default ImageAdapter;
