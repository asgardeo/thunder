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
import {useDesign} from '@thunder/shared-design';
import {OxygenUIThemeProvider, ColorSchemeToggle, CircularProgress, Box, AcrylicOrangeTheme} from '@wso2/oxygen-ui';
import App from './App';

export default function AppWithTheme(): JSX.Element {
  const {transformedTheme, isLoading} = useDesign(AcrylicOrangeTheme);

  return (
    <OxygenUIThemeProvider theme={transformedTheme ?? AcrylicOrangeTheme}>
      <ColorSchemeToggle
        sx={{
          position: 'fixed',
          top: '2.3rem',
          right: '3rem',
          zIndex: 2,
        }}
      />
      {isLoading ? (
        <Box
          sx={{
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            height: '100vh',
            width: '100vw',
          }}
        >
          <CircularProgress />
        </Box>
      ) : (
        <App />
      )}
    </OxygenUIThemeProvider>
  );
}
