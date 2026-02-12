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

import {useState, useEffect, type ReactNode, type ChangeEvent} from 'react';
import {useTranslation} from 'react-i18next';
import {Divider, FormHelperText, FormLabel, Stack, TextField} from '@wso2/oxygen-ui';
import type {CommonResourcePropertiesPropsInterface} from '@/features/flows/components/resource-property-panel/ResourceProperties';
import type {Element} from '@/features/flows/models/elements';

/**
 * Props interface of {@link ButtonExtendedProperties}
 */
export type ButtonExtendedPropertiesPropsInterface = CommonResourcePropertiesPropsInterface;

/**
 * Extended properties for the button elements.
 * Provides optional start icon and end icon configuration.
 *
 * @param props - Props injected to the component.
 * @returns The ButtonExtendedProperties component.
 */
function ButtonExtendedProperties({resource, onChange}: ButtonExtendedPropertiesPropsInterface): ReactNode {
  const {t} = useTranslation();

  // Use local state for immediate input feedback
  const [startIconValue, setStartIconValue] = useState(() => {
    const element = resource as Element & {startIcon?: string};
    return element?.startIcon ?? '';
  });

  const [endIconValue, setEndIconValue] = useState(() => {
    const element = resource as Element & {endIcon?: string};
    return element?.endIcon ?? '';
  });

  // Sync local state when resource changes (e.g., switching to a different button)
  useEffect(() => {
    const element = resource as Element & {startIcon?: string; endIcon?: string};
    setStartIconValue(element?.startIcon ?? '');
    setEndIconValue(element?.endIcon ?? '');
  }, [resource]);

  // Handle startIcon change - update local state immediately, propagate via onChange
  const handleStartIconChange = (value: string): void => {
    setStartIconValue(value);
    onChange('startIcon', value, resource);
  };

  // Handle endIcon change - update local state immediately, propagate via onChange
  const handleEndIconChange = (value: string): void => {
    setEndIconValue(value);
    onChange('endIcon', value, resource);
  };

  return (
    <Stack gap={2}>
      <Divider sx={{marginY: 2}} />

      <div>
        <FormLabel htmlFor="start-icon-input">
          {t('flows:core.buttonExtendedProperties.startIcon.label')}
        </FormLabel>
        <TextField
          id="start-icon-input"
          value={startIconValue}
          onChange={(e: ChangeEvent<HTMLInputElement>) => handleStartIconChange(e.target.value)}
          placeholder={t('flows:core.buttonExtendedProperties.startIcon.placeholder')}
          fullWidth
          size="small"
        />
        <FormHelperText>
          {t('flows:core.buttonExtendedProperties.startIcon.hint')}
        </FormHelperText>
      </div>

      <div>
        <FormLabel htmlFor="end-icon-input">
          {t('flows:core.buttonExtendedProperties.endIcon.label')}
        </FormLabel>
        <TextField
          id="end-icon-input"
          value={endIconValue}
          onChange={(e: ChangeEvent<HTMLInputElement>) => handleEndIconChange(e.target.value)}
          placeholder={t('flows:core.buttonExtendedProperties.endIcon.placeholder')}
          fullWidth
          size="small"
        />
        <FormHelperText>
          {t('flows:core.buttonExtendedProperties.endIcon.hint')}
        </FormHelperText>
      </div>

      <Divider sx={{marginY: 2}} />
    </Stack>
  );
}

export default ButtonExtendedProperties;
