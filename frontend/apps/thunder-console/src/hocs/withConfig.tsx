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

import {AsgardeoProvider} from '@asgardeo/react';
import {useConfig} from '@thunder/contexts';
import type {JSX, ComponentType} from 'react';

export default function withConfig<P extends object>(WrappedComponent: ComponentType<P>) {
  return function WithConfig(props: P): JSX.Element {
    const {getTrustedIssuerUrl, getTrustedIssuerClientId, getClientUrl, getTrustedIssuerScopes, getServerUrl, config} =
      useConfig();

    const signInOptions: Record<string, string> | undefined = config.trusted_issuer
      ? {resource: getServerUrl()}
      : undefined;

    return (
      <AsgardeoProvider
        baseUrl={getTrustedIssuerUrl() ?? (import.meta.env.VITE_ASGARDEO_BASE_URL as string)}
        clientId={getTrustedIssuerClientId() ?? (import.meta.env.VITE_ASGARDEO_CLIENT_ID as string)}
        afterSignInUrl={getClientUrl() ?? (import.meta.env.VITE_ASGARDEO_AFTER_SIGN_IN_URL as string)}
        scopes={getTrustedIssuerScopes().length > 0 ? getTrustedIssuerScopes() : undefined}
        signInOptions={signInOptions}
        platform="AsgardeoV2"
        discovery={{
          wellKnown: {
            enabled: true,
          },
        }}
      >
        <WrappedComponent {...props} />
      </AsgardeoProvider>
    );
  };
}
