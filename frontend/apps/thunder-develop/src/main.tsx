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
import {StyledEngineProvider, ThemeProvider} from '@mui/material/styles';
import {theme} from '@thunder/ui';
import CssBaseline from '@mui/material/CssBaseline';
import {StrictMode} from 'react';
import {AsgardeoProvider} from '@asgardeo/react';
import App from './App';

ReactDOM.createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <AsgardeoProvider
      baseUrl={import.meta.env.VITE_ASGARDEO_BASE_URL}
      clientId={import.meta.env.VITE_ASGARDEO_CLIENT_ID}
      platform="AsgardeoV2"
    >
      <StyledEngineProvider injectFirst>
        <ThemeProvider theme={theme}>
          <CssBaseline enableColorScheme />
          <App />
        </ThemeProvider>
      </StyledEngineProvider>
    </AsgardeoProvider>
  </StrictMode>,
);
