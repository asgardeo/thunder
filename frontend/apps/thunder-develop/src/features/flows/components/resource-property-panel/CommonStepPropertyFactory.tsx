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

import type {ChangeEvent, ReactElement, SyntheticEvent} from 'react';
import {Checkbox, FormControl, FormControlLabel, FormLabel, TextField} from '@wso2/oxygen-ui';
import startCase from 'lodash-es/startCase';
import RichTextWithTranslation from './rich-text/RichTextWithTranslation';
import {ElementTypes} from '../../models/elements';
import type {Resource} from '../../models/resources';

/**
 * Props interface of {@link CommonStepPropertyFactory}
 */
export interface CommonStepPropertyFactoryPropsInterface {
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
   * @param propertyKey - The key of the property.=
   * @param newValue - The new value of the property.
   * @param resource - The resource associated with the property.
   */
  onChange: (propertyKey: string, newValue: string | boolean, resource: Resource) => void;
  /**
   * Additional props.
   */
  [key: string]: unknown;
}

/**
 * Factory to generate the common property configurator for the given step.
 *
 * @param props - Props injected to the component.
 * @returns The CommonStepPropertyFactory component.
 */
function CommonStepPropertyFactory({
  resource,
  propertyKey,
  propertyValue,
  onChange,
  ...rest
}: CommonStepPropertyFactoryPropsInterface): ReactElement | null {
  if (propertyKey === 'text') {
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

  if (typeof propertyValue === 'boolean') {
    return (
      <FormControlLabel
        control={<Checkbox checked={propertyValue} />}
        label={startCase(propertyKey)}
        onChange={(_event: SyntheticEvent, checked: boolean) => onChange(propertyKey, checked, resource)}
        {...rest}
      />
    );
  }

  if (typeof propertyValue === 'string') {
    return (
      <FormControl fullWidth sx={{mb: 3}}>
        <FormLabel htmlFor="name">{startCase(propertyKey)} </FormLabel>
        <TextField
          fullWidth
          defaultValue={propertyValue}
          onChange={(e: ChangeEvent<HTMLInputElement>) => onChange(propertyKey, e.target.value, resource)}
          placeholder={`Enter ${startCase(propertyKey)}`}
          {...rest}
        />
      </FormControl>
    );
  }

  return null;
}

export default CommonStepPropertyFactory;
