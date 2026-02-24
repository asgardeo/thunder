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

import {useEffect, useMemo, useRef, useState, type ChangeEvent, type ReactElement} from 'react';
import {useTranslation} from 'react-i18next';
import {
  Box,
  FormControl,
  FormHelperText,
  FormLabel,
  IconButton,
  InputAdornment,
  TextField,
  Tooltip,
  Typography,
} from '@wso2/oxygen-ui';
import {SquareFunction} from '@wso2/oxygen-ui-icons-react';
import startCase from 'lodash-es/startCase';
import useValidationStatus from '../../hooks/useValidationStatus';
import type {Resource} from '../../models/resources';
import DynamicValuePopover from './DynamicValuePopover';
import {isDynamicValuePattern, isI18nPattern as checkIsI18nPattern, resolveI18nValue} from '../../utils/i18nPatternUtils';

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
  const [isDynamicValuePopoverOpen, setIsDynamicValuePopoverOpen] = useState<boolean>(false);
  const [localValue, setLocalValue] = useState<string>(propertyValue);
  const iconButtonRef = useRef<HTMLButtonElement>(null);
  const {selectedNotification} = useValidationStatus();

  /**
   * Sync local state when propertyValue changes from external sources.
   */
  useEffect(() => {
    setLocalValue(propertyValue);
  }, [propertyValue]);

  /**
   * Check if the property value matches any dynamic value pattern (i18n or meta).
   */
  const isDynamic: boolean = useMemo(() => isDynamicValuePattern(propertyValue), [propertyValue]);

  /**
   * Check specifically for i18n pattern to resolve and display a preview.
   */
  const isI18nPattern: boolean = useMemo(() => checkIsI18nPattern(propertyValue), [propertyValue]);

  /**
   * Resolve the i18n value if the pattern is detected.
   */
  const resolvedI18nValue: string = useMemo(() => {
    if (isI18nPattern) {
      return resolveI18nValue(propertyValue, t);
    }

    return '';
  }, [propertyValue, isI18nPattern, t]);

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
   * Handles the toggle of the dynamic value popover.
   */
  const handleDynamicValueToggle = () => {
    setIsDynamicValuePopoverOpen(!isDynamicValuePopoverOpen);
  };

  /**
   * Handles the closing of the dynamic value popover.
   */
  const handleDynamicValueClose = () => {
    setIsDynamicValuePopoverOpen(false);
  };

  return (
    <Box>
      <FormControl fullWidth>
        <FormLabel htmlFor={`${resource.id}-${propertyKey}`}>{startCase(propertyKey)}</FormLabel>
        <TextField
          fullWidth
          value={localValue}
          error={!!errorMessage}
          onChange={(e: ChangeEvent<HTMLInputElement>) => {
            setLocalValue(e.target.value);
            onChange(propertyKey, e.target.value, resource);
          }}
          placeholder={t('flows:core.elements.textPropertyField.placeholder', {propertyName: startCase(propertyKey)})}
          sx={
            isDynamic
              ? {
                  '& .MuiOutlinedInput-root': {
                    backgroundColor: 'rgba(var(--mui-palette-primary-mainChannel) / 0.1)',
                    '& fieldset': {
                      borderColor: 'primary.main',
                    },
                    '&:hover fieldset': {
                      borderColor: 'primary.dark',
                    },
                    '&.Mui-focused fieldset': {
                      borderColor: 'primary.main',
                    },
                  },
                }
              : undefined
          }
          InputProps={{
            endAdornment: (
              <InputAdornment position="end">
                <Tooltip title={t('flows:core.elements.textPropertyField.tooltip.configureDynamicValue')}>
                  <IconButton
                    ref={iconButtonRef}
                    onClick={handleDynamicValueToggle}
                    size="small"
                    edge="end"
                    color={isDynamic ? 'primary' : 'default'}
                  >
                    <SquareFunction size={16} />
                  </IconButton>
                </Tooltip>
              </InputAdornment>
            ),
          }}
          {...rest}
        />
      </FormControl>
      {errorMessage && <FormHelperText error>{errorMessage}</FormHelperText>}
      {isI18nPattern && resolvedI18nValue && (
        <Box
          sx={{
            mt: 1,
            p: 1.5,
            backgroundColor: 'action.hover',
            borderRadius: 1,
            border: '1px solid',
            borderColor: 'divider',
          }}
        >
          <Typography variant="caption" color="text.secondary" sx={{display: 'block', mb: 0.5}}>
            {t('flows:core.elements.textPropertyField.resolvedValue')}
          </Typography>
          <Typography variant="body2" sx={{wordBreak: 'break-word'}}>
            {resolvedI18nValue}
          </Typography>
        </Box>
      )}
      <DynamicValuePopover
        open={isDynamicValuePopoverOpen}
        anchorEl={iconButtonRef.current}
        propertyKey={propertyKey}
        onClose={handleDynamicValueClose}
        value={propertyValue}
        onChange={(newValue: string) => onChange(propertyKey, newValue, resource)}
      />
    </Box>
  );
}

export default TextPropertyField;
