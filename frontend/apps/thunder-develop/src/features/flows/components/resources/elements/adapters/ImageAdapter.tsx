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

import {useCallback, useMemo, useState, type CSSProperties, type ReactElement} from 'react';
import {Trans, useTranslation} from 'react-i18next';
import {Box, Typography} from '@wso2/oxygen-ui';
import {ImageIcon} from '@wso2/oxygen-ui-icons-react';
import type {RequiredFieldInterface} from '@/features/flows/hooks/useRequiredFields';
import useRequiredFields from '@/features/flows/hooks/useRequiredFields';
import type {Element as FlowElement} from '@/features/flows/models/elements';

/**
 * Image element type with properties at top level.
 */
export type ImageElement = FlowElement & {
  src?: string;
  alt?: string;
  styles?: CSSProperties;
};

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
  const imageElement = resource as ImageElement;
  const [hasError, setHasError] = useState(false);

  const generalMessage: ReactElement = useMemo(
    () => (
      <Trans i18nKey="flows:core.validation.fields.image.general" values={{id: resource.id}}>
        Required fields are not properly configured for the image with ID <code>{resource.id}</code>.
      </Trans>
    ),
    [resource.id],
  );

  const validationFields: RequiredFieldInterface[] = useMemo(
    () => [
      {
        errorMessage: t('flows:core.validation.fields.image.src'),
        name: 'src',
      },
    ],
    [t],
  );

  useRequiredFields(resource, generalMessage, validationFields);

  // Check if src is empty or has errored
  const src = imageElement?.src?.trim() ?? '';
  const shouldShowPlaceholder = !src || hasError;

  const handleImageError = useCallback(() => {
    setHasError(true);
  }, []);

  // Placeholder UI
  const placeholderUI = (
    <Box
      display="flex"
      flexDirection="column"
      alignItems="center"
      justifyContent="center"
      sx={{
        width: '100%',
        minHeight: 120,
        backgroundColor: 'rgba(0, 0, 0, 0.04)',
        borderRadius: 1,
        border: '1px dashed rgba(0, 0, 0, 0.2)',
      }}
    >
      <ImageIcon style={{fontSize: 48, color: 'rgba(0, 0, 0, 0.3)'}} />
      <Typography variant="body2" color="textSecondary" sx={{mt: 1}}>
        {t('flows:core.placeholders.image')}
      </Typography>
    </Box>
  );

  if (shouldShowPlaceholder) {
    return placeholderUI;
  }

  return (
    <Box display="flex" alignItems="center" justifyContent="center">
      <img
        src={src}
        alt={imageElement?.alt}
        width="100%"
        style={imageElement?.styles}
        onError={handleImageError}
      />
    </Box>
  );
}

export default ImageAdapter;
