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

import React, {PropsWithChildren, useEffect} from 'react';
import {OxygenUIThemeProvider, AcrylicOrangeTheme} from '@wso2/oxygen-ui';
import {LoggerProvider, LogLevel} from '@thunder/logger/react';
import { useLocation } from '@docusaurus/router';

export default function Root({children = null}: PropsWithChildren<Record<string, unknown>>) {
  const location = useLocation();

  useEffect(() => {
    const html = document.documentElement;
    const pagePath = location.pathname.replace(/\//g, '-').replace(/^-|-$/g, '') || 'home';

    html.setAttribute('data-page', pagePath);

    return () => {
      html.removeAttribute('data-page');
    };
  }, [location.pathname]);

  return (
    <OxygenUIThemeProvider theme={AcrylicOrangeTheme}>
      <LoggerProvider
        logger={{
          level: LogLevel.DEBUG,
        }}
      >
        {children}
      </LoggerProvider>
    </OxygenUIThemeProvider>
  );
}
