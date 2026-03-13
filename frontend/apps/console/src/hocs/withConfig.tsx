/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

import type {JSX, ComponentType} from 'react';
import {AsgardeoProvider} from '@asgardeo/react';
import {useConfig} from '@thunder/shared-contexts';

export default function withConfig<P extends object>(WrappedComponent: ComponentType<P>) {
  return function WithConfig(props: P): JSX.Element {
    const {getClientId, getServerUrl, getClientUrl, getScopes} = useConfig();

    return (
      <AsgardeoProvider
        baseUrl={getServerUrl() ?? (import.meta.env.VITE_ASGARDEO_BASE_URL as string)}
        clientId={getClientId() ?? (import.meta.env.VITE_ASGARDEO_CLIENT_ID as string)}
        afterSignInUrl={getClientUrl() ?? (import.meta.env.VITE_ASGARDEO_AFTER_SIGN_IN_URL as string)}
        scopes={getScopes().length > 0 ? getScopes() : undefined}
        platform="AsgardeoV2"
      >
        <WrappedComponent {...props} />
      </AsgardeoProvider>
    );
  };
}
