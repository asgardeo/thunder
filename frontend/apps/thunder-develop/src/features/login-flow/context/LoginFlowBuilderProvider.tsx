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

import FlowBuilderCoreProvider from '@/features/flows/context/FlowBuilderCoreProvider';
import {PreviewScreenType} from '@/features/flows/models/custom-text-preference';
import {useMemo, type PropsWithChildren, type ReactElement} from 'react';
import FlowContextWrapper from './FlowContextWrapper';
import ResourceProperties from '../components/resource-property-panel/ResourceProperties';
import ElementFactory from '../components/resources/elements/ElementFactory';
/**
 * This component provides login flow builder related context to its children.
 *
 * @param props - Props injected to the component.
 * @returns The LoginFlowBuilderProvider component.
 */
function LoginFlowBuilderProvider({children}: PropsWithChildren): ReactElement {
  const screensList: PreviewScreenType[] = useMemo(
    () => [
      PreviewScreenType.SIGN_UP,
      PreviewScreenType.COMMON,
      PreviewScreenType.EMAIL_LINK_EXPIRY,
      PreviewScreenType.SMS_OTP,
      PreviewScreenType.EMAIL_OTP,
    ],
    [],
  );

  return (
    <FlowBuilderCoreProvider
      ElementFactory={ElementFactory}
      ResourceProperties={ResourceProperties}
      screenTypes={screensList}
    >
      <FlowContextWrapper>{children}</FlowContextWrapper>
    </FlowBuilderCoreProvider>
  );
}

export default LoginFlowBuilderProvider;
