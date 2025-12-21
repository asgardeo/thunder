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

import type {ReactElement} from 'react';
import {TextField} from '@wso2/oxygen-ui';
import type {Resource} from '../../models/resources';
import {ElementTypes} from '../../models/elements';
import RichTextWithTranslation from './rich-text/RichTextWithTranslation';
import CheckboxPropertyField from './CheckboxPropertyField';
import TextPropertyField from './TextPropertyField';
import FlowBuilderElementConstants from '../../constants/FlowBuilderElementConstants';

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
        inputProps={{
          disabled: true,
          readOnly: true,
        }}
        {...rest}
      />
    );
  }

  return null;
}

export default CommonElementPropertyFactory;
