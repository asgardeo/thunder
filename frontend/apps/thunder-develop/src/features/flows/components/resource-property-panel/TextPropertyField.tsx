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

import {useMemo, useRef, useState, type ChangeEvent, type ReactElement} from 'react';
import {useTranslation} from 'react-i18next';
import {Box, FormControl, FormHelperText, FormLabel, TextField} from '@wso2/oxygen-ui';
import startCase from 'lodash-es/startCase';
import useValidationStatus from '../../hooks/useValidationStatus';
import type {Resource} from '../../models/resources';
import I18nConfigurationCard from './I18nConfigurationCard';
import {isI18nPattern as checkIsI18nPattern, extractI18nKey} from '../../utils/i18nPatternUtils';

/**
 * Props interface of {@link TextPropertyField}
 */
export interface TextPropertyFieldPropsInterface {
  /**
   * The resource associated with the property.
   */
  resource: Resource;
  /**
   * The key of the property.
   */
  propertyKey: string;
  /**
   * The value of the property.
   */
  propertyValue: string;
  /**
   * The event handler for the property change.
   * @param propertyKey - The key of the property.
   * @param newValue - The new value of the property.
   * @param resource - The resource associated with the property.
   */
  onChange: (propertyKey: string, newValue: string, resource: Resource) => void;
  /**
   * Additional props.
   */
  [key: string]: unknown;
}

/**
 * Text property field component for rendering text input fields.
 *
 * @param props - Props injected to the component.
 * @returns The TextPropertyField component.
 */
function TextPropertyField({
  resource,
  propertyKey,
  propertyValue,
  onChange,
  ...rest
}: TextPropertyFieldPropsInterface): ReactElement {
  const {t} = useTranslation();
  const [isI18nCardOpen, setIsI18nCardOpen] = useState<boolean>(false);
  const iconButtonRef = useRef<HTMLButtonElement>(null);
  const {selectedNotification} = useValidationStatus();

  /**
   * Check if the property value matches the i18n pattern.
   */
  const isI18nPattern: boolean = useMemo(() => checkIsI18nPattern(propertyValue), [propertyValue]);

  /**
   * Get the error message for the text property field.
   */
  const errorMessage: string = useMemo(() => {
    const key = `${resource?.id}_${propertyKey}`;

    if (selectedNotification?.hasResourceFieldNotification(key)) {
      return selectedNotification?.getResourceFieldNotification(key);
    }

    return '';
  }, [propertyKey, resource?.id, selectedNotification]);

  /**
   * Handles the toggle of the i18n configuration card.
   */
  // const handleI18nToggle = () => {
  //   setIsI18nCardOpen(!isI18nCardOpen);
  // };

  /**
   * Handles the closing of the i18n configuration card.
   */
  const handleI18nClose = () => {
    setIsI18nCardOpen(false);
  };

  return (
    <Box>
      <FormControl fullWidth>
        <FormLabel htmlFor={`${resource.id}-${propertyKey}`}>{startCase(propertyKey)}</FormLabel>
        <TextField
          fullWidth
          defaultValue={propertyValue}
          value={isI18nPattern ? '' : undefined}
          error={!!errorMessage}
          onChange={(e: ChangeEvent<HTMLInputElement>) => onChange(propertyKey, e.target.value, resource)}
          placeholder={
            isI18nPattern
              ? ''
              : t('flows:core.elements.textPropertyField.placeholder', {propertyName: startCase(propertyKey)})
          }
          {...rest}
        />
      </FormControl>
      {errorMessage && <FormHelperText error>{errorMessage}</FormHelperText>}
      {isI18nPattern && (
        <div className="text-property-field-i18n-placeholder">
          <div className="text-property-field-i18n-placeholder-key">{propertyValue}</div>
        </div>
      )}
      {isI18nCardOpen && (
        <I18nConfigurationCard
          open={isI18nCardOpen}
          anchorEl={iconButtonRef.current}
          propertyKey={propertyKey}
          onClose={handleI18nClose}
          i18nKey={extractI18nKey(propertyValue) ?? ''}
          onChange={(i18nKey: string) => onChange(propertyKey, i18nKey ? `{{t(${i18nKey})}}` : '', resource)}
        />
      )}
    </Box>
  );
}

export default TextPropertyField;
