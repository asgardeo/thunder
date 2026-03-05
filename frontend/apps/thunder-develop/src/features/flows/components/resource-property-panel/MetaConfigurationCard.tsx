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

import lowerCase from 'lodash-es/lowerCase';
import startCase from 'lodash-es/startCase';
import React, {type ReactElement, type SyntheticEvent, useCallback, useState} from 'react';
import {useTranslation} from 'react-i18next';
import {
  Autocomplete,
  type AutocompleteRenderInputParams,
  Box,
  Card,
  CardContent,
  CardHeader,
  Divider,
  IconButton,
  Popover,
  TextField,
  Tooltip,
  Typography,
} from '@wso2/oxygen-ui';
import {XIcon} from '@wso2/oxygen-ui-icons-react';
import DynamicValueSyntax from './DynamicValueSyntax';

/**
 * Common meta variable paths available for selection.
 */
const COMMON_META_FIELDS: string[] = [
  'application.name',
  'application.description',
  'application.loginPageUrl',
  'application.logoUrl',
];

/**
 * Props interface for the MetaConfigurationCardContent component.
 */
export interface MetaConfigurationCardContentProps {
  metaKey: string;
  onChange: (metaKey: string) => void;
}

/**
 * Props interface of {@link MetaConfigurationCard}
 */
export interface MetaConfigurationCardPropsInterface {
  open: boolean;
  anchorEl: HTMLElement | null;
  propertyKey: string;
  onClose: () => void;
  metaKey: string;
  onChange: (metaKey: string) => void;
}

/**
 * Inner content for meta variable configuration (no Popover shell).
 * Used both in standalone MetaConfigurationCard and embedded in DynamicValuePopover tabs.
 */
export function MetaConfigurationCardContent({metaKey, onChange}: MetaConfigurationCardContentProps): ReactElement {
  const {t} = useTranslation();
  const [localValue, setLocalValue] = useState<string>(metaKey);

  const handleChange = useCallback(
    (value: string) => {
      setLocalValue(value);
      onChange(value);
    },
    [onChange],
  );

  return (
    <Box sx={{display: 'flex', flexDirection: 'column', gap: 2}}>
      <div>
        <Typography variant="subtitle2" gutterBottom>
          {t('flows:core.elements.textPropertyField.metaCard.variablePath')}
        </Typography>
        <Autocomplete
          freeSolo
          options={COMMON_META_FIELDS}
          value={localValue}
          onChange={(_event: SyntheticEvent, newValue: string | null) => {
            handleChange(newValue ?? '');
          }}
          onInputChange={(_event: React.SyntheticEvent, newValue: string) => {
            handleChange(newValue);
          }}
          renderInput={(params: AutocompleteRenderInputParams) => (
            <TextField
              {...params}
              size="small"
              placeholder={t('flows:core.elements.textPropertyField.metaCard.variablePathPlaceholder')}
              helperText={t('flows:core.elements.textPropertyField.metaCard.variablePathHint')}
            />
          )}
          renderOption={({key, ...props}: React.HTMLAttributes<HTMLLIElement> & {key: string}, option: string) => (
            <li key={key} {...props}>
              <Tooltip title={option} placement="bottom">
                <span>{option}</span>
              </Tooltip>
            </li>
          )}
        />
      </div>

      {localValue && (
        <>
          <Divider />
          <Box
            sx={{
              p: 1.5,
              backgroundColor: 'action.hover',
              borderRadius: 1,
              border: '1px solid',
              borderColor: 'divider',
            }}
          >
            <Typography variant="caption" color="text.secondary" sx={{display: 'block', mb: 0.5}}>
              {t('flows:core.elements.textPropertyField.metaCard.formattedValue')}
            </Typography>
            <DynamicValueSyntax value={`{{meta(${localValue})}}`} />
          </Box>
        </>
      )}
    </Box>
  );
}

/**
 * Meta variable configuration floating card component.
 * Provides an autocomplete to select or type a meta variable path
 * and displays the formatted {{meta(...)}} syntax.
 */
function MetaConfigurationCard({
  open,
  anchorEl,
  propertyKey,
  onClose,
  onChange,
  metaKey,
}: MetaConfigurationCardPropsInterface): ReactElement {
  const {t} = useTranslation();

  return (
    <Popover
      open={open}
      anchorEl={anchorEl}
      onClose={onClose}
      anchorOrigin={{
        vertical: 'top',
        horizontal: 'right',
      }}
      transformOrigin={{
        vertical: 'top',
        horizontal: 'left',
      }}
    >
      <Card sx={{width: 400}}>
        <CardHeader
          title={t('flows:core.elements.textPropertyField.metaCard.title', {
            field: startCase(lowerCase(propertyKey)),
          })}
          action={
            <IconButton aria-label={t('common:close')} onClick={onClose} size="small">
              <XIcon />
            </IconButton>
          }
        />
        <CardContent>
          <MetaConfigurationCardContent metaKey={metaKey} onChange={onChange} />
        </CardContent>
      </Card>
    </Popover>
  );
}

export default MetaConfigurationCard;
