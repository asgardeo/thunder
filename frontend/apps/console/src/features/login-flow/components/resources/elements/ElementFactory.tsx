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

import type {ElementFactoryProps} from '@/features/flows/context/FlowBuilderCoreProvider';
import CommonElementFactory from '@/features/flows/components/resources/elements/CommonElementFactory';
import {ResourceTypes} from '@/features/flows/models/resources';
import {type ReactElement} from 'react';
import type {Element} from '@/features/flows/models/elements';

/**
 * Factory for creating components.
 *
 * @param props - Props injected to the component.
 * @returns The ElementFactory component.
 */
function ElementFactory({resource, stepId, ...rest}: ElementFactoryProps): ReactElement | null {
  // Allow resources without resourceType (template/widget components) or with resourceType === 'ELEMENT'
  if (!resource || (resource.resourceType && resource.resourceType !== ResourceTypes.Element)) {
    return null;
  }

  return <CommonElementFactory resource={resource as Element} stepId={stepId} {...rest} />;
}

export default ElementFactory;
