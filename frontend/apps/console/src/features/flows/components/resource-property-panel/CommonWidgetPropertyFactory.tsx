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
import type {Resource} from '../../models/resources';

/**
 * Props interface of {@link CommonWidgetPropertyFactory}
 */
export interface CommonWidgetPropertyFactoryPropsInterface {
  /**
   * The resource associated with the property.
   */
  resource: Resource;
  /**
   * The key of the property.
   */
  // propertyKey: string;
  /**
   * The value of the property.
   */
  // propertyValue: unknown;
  /**
   * The event handler for the property change.
   * @param propertyKey - The key of the property.
   * @param newValue - The new value of the property.
   * @param resource - The resource associated with the property.
   */
  // onChange: (propertyKey: string, newValue: unknown, resource: Resource) => void;
  /**
   * Additional props.
   */
  [key: string]: unknown;
}

/**
 * Factory to generate the common property configurator for the given widget.
 * TODO: Implement the common widgets like RE-CAPTCHA, etc.
 *
 * @param props - Props injected to the component.
 * @returns The CommonWidgetPropertyFactory component.
 */
function CommonWidgetPropertyFactory({resource}: CommonWidgetPropertyFactoryPropsInterface): ReactElement | null {
  switch (resource.type) {
    default:
      return null;
  }
}

export default CommonWidgetPropertyFactory;
