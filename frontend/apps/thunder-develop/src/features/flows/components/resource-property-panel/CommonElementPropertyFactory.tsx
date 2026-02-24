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

import {type ComponentType, type ReactElement} from 'react';
import {Autocomplete, type AutocompleteRenderInputParams, Box, TextField} from '@wso2/oxygen-ui';
import * as Icons from '@wso2/oxygen-ui-icons-react';
import startCase from 'lodash-es/startCase';
import type {Resource} from '../../models/resources';
import {ElementTypes} from '../../models/elements';
import RichTextWithTranslation from './rich-text/RichTextWithTranslation';
import CheckboxPropertyField from './CheckboxPropertyField';
import TextPropertyField from './TextPropertyField';
import FlowBuilderElementConstants from '../../constants/FlowBuilderElementConstants';

/**
 * All PascalCase named icon exports from the icons package — computed once at module load.
 * Lucide icons are forwardRef exotic components (typeof === 'object'), not plain functions.
 * We identify real icon components by the presence of `displayName` (set by createLucideIcon).
 * The base `Icon` component has no displayName and must be excluded to avoid render errors.
 */
const ICON_NAMES: string[] = Object.keys(Icons).filter((k) => {
  if (!/^[A-Z]/.test(k)) return false;
  const v = Icons[k as keyof typeof Icons];
  return (
    typeof v === 'object' &&
    v !== null &&
    '$$typeof' in (v as object) &&
    typeof (v as Record<string, unknown>).displayName === 'string'
  );
});

/**
 * Props interface of {@link CommonElementPropertyFactory}
 */
export interface CommonElementPropertyFactoryPropsInterface {
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
  propertyValue: unknown;
  /**
   * The event handler for the property change.
   * @param propertyKey - The key of the property.
   * @param newValue - The new value of the property.
   * @param resource - The resource associated with the property.
   */
  onChange: (propertyKey: string, newValue: unknown, resource: Resource) => void;
  /**
   * Additional props.
   */
  [key: string]: unknown;
}

/**
 * Factory to generate the common property configurator for the given element.
 *
 * @param props - Props injected to the component.
 * @returns The CommonElementPropertyFactory component.
 */
function CommonElementPropertyFactory({
  resource,
  propertyKey,
  propertyValue,
  onChange,
  ...rest
}: CommonElementPropertyFactoryPropsInterface): ReactElement | null {
  if (propertyKey === 'label') {
    if (resource.type === ElementTypes.RichText) {
      return (
        <RichTextWithTranslation
          onChange={(html: string) => onChange(propertyKey, html, resource)}
          resource={resource}
          {...rest}
        />
      );
    }
  }

  if (resource.type === ElementTypes.Icon && propertyKey === 'name') {
    const IconPreview = typeof propertyValue === 'string' && propertyValue
      ? (Icons[propertyValue as keyof typeof Icons] as ComponentType<{size?: number}> | undefined)
      : undefined;

    return (
      <Box>
        <Autocomplete
          options={ICON_NAMES}
          value={typeof propertyValue === 'string' ? propertyValue : null}
          onChange={(_event: React.SyntheticEvent, newValue: string | null) => {
            onChange(propertyKey, newValue ?? '', resource);
          }}
          renderInput={(params: AutocompleteRenderInputParams) => {
            const {InputProps: acInputProps, inputProps: acHtmlInputProps, InputLabelProps, ...restParams} = params;
            return (
              <TextField
                {...restParams}
                label={startCase(propertyKey)}
                size="small"
                slotProps={{
                  input: {
                    ...acInputProps,
                    startAdornment: IconPreview ? (
                      <Box sx={{display: 'flex', alignItems: 'center', pl: 0.5, pr: 0.5}}>
                        <IconPreview size={16} />
                      </Box>
                    ) : acInputProps?.startAdornment,
                  },
                  htmlInput: acHtmlInputProps,
                  inputLabel: InputLabelProps,
                }}
              />
            );
          }}
          renderOption={({key, ...props}: React.HTMLAttributes<HTMLLIElement> & {key: string}, option: string) => {
            const Icon = Icons[option as keyof typeof Icons] as ComponentType<{size?: number}> | undefined;
            return (
              <li key={key} {...props}>
                <Box display="flex" alignItems="center" gap={1}>
                  {Icon && <Icon size={16} />}
                  {option}
                </Box>
              </li>
            );
          }}
        />
      </Box>
    );
  }

  if (typeof propertyValue === 'boolean') {
    return (
      <CheckboxPropertyField
        resource={resource}
        propertyKey={propertyKey}
        propertyValue={propertyValue}
        onChange={onChange}
        {...rest}
      />
    );
  }

  if (typeof propertyValue === 'number') {
    return (
      <TextPropertyField
        resource={resource}
        propertyKey={propertyKey}
        propertyValue={String(propertyValue)}
        onChange={(key, value, res) => onChange(key, value !== '' ? Number(value) : 0, res)}
        {...rest}
      />
    );
  }

  if (typeof propertyValue === 'string') {
    return (
      <TextPropertyField
        resource={resource}
        propertyKey={propertyKey}
        propertyValue={propertyValue}
        onChange={onChange}
        {...rest}
      />
    );
  }

  if (resource.type === ElementTypes.Captcha) {
    return (
      <TextField
        fullWidth
        label="Provider"
        defaultValue={FlowBuilderElementConstants.DEFAULT_CAPTCHA_PROVIDER}
        slotProps={{
          htmlInput: {
            disabled: true,
            readOnly: true,
          },
        }}
        {...rest}
      />
    );
  }

  return null;
}

export default CommonElementPropertyFactory;
