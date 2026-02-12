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

import * as ReactDOM from 'react-dom/client';
import {StrictMode} from 'react';
import {ConfigProvider} from '@thunder/shared-contexts';
import {LoggerProvider, LogLevel} from '@thunder/logger/react';
import i18n from 'i18next';
import {initReactI18next} from 'react-i18next';
import enUS from '@thunder/i18n/locales/en-US';
import {QueryClient, QueryClientProvider} from '@tanstack/react-query';
import {ReactQueryDevtools} from '@tanstack/react-query-devtools';
import AppWithConfig from './AppWithConfig';
import './index.css';

// Initialize i18n before rendering the app
await i18n.use(initReactI18next).init({
  resources: {
    'en-US': enUS,
  },
  lng: 'en-US',
  fallbackLng: 'en-US',
  defaultNS: 'common',
  interpolation: {
    escapeValue: false,
  },
  debug: import.meta.env.DEV,
});

const queryClient: QueryClient = new QueryClient();

ReactDOM.createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ConfigProvider>
      <LoggerProvider
        logger={{
          level: import.meta.env.DEV ? LogLevel.DEBUG : LogLevel.INFO,
        }}
      >
        <QueryClientProvider client={queryClient}>
          <AppWithConfig />
          <ReactQueryDevtools initialIsOpen={false} />
        </QueryClientProvider>
      </LoggerProvider>
    </ConfigProvider>
  </StrictMode>,
);
