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

import type {JSX} from 'react';
import {OxygenUIThemeProvider} from '@wso2/oxygen-ui';
import {AsgardeoProvider} from '@asgardeo/react';
import {useConfig} from '@thunder/shared-contexts';
import {QueryClient, QueryClientProvider} from '@tanstack/react-query';
import {ReactQueryDevtools} from '@tanstack/react-query-devtools';
import I18nProvider from './i18n/I18nProvider';
import App from './App';

const queryClient: QueryClient = new QueryClient();

export default function AppWithConfig(): JSX.Element {
  const {getClientId, getServerUrl, getClientUrl, getScopes} = useConfig();

  return (
    <AsgardeoProvider
      baseUrl={getServerUrl() ?? (import.meta.env.VITE_ASGARDEO_BASE_URL as string)}
      clientId={getClientId() ?? (import.meta.env.VITE_ASGARDEO_CLIENT_ID as string)}
      afterSignInUrl={getClientUrl() ?? (import.meta.env.VITE_ASGARDEO_AFTER_SIGN_IN_URL as string)}
      scopes={getScopes().length > 0 ? getScopes() : undefined}
      platform="AsgardeoV2"
    >
      <OxygenUIThemeProvider radialBackground>
        <QueryClientProvider client={queryClient}>
          <I18nProvider>
            <App />
          </I18nProvider>
          <ReactQueryDevtools initialIsOpen={false} />
        </QueryClientProvider>
      </OxygenUIThemeProvider>
    </AsgardeoProvider>
  );
}
