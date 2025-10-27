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

/* eslint-disable react-refresh/only-export-components */
import type {ReactElement, ReactNode} from 'react';
import {render, type RenderOptions} from '@testing-library/react';
import {MemoryRouter} from 'react-router';
import {createTheme, ThemeProvider} from '@mui/material/styles';

// Create a basic theme for testing
const theme = createTheme();

interface ProvidersProps {
  children: ReactNode;
}

// Wrapper component with common providers
function Providers({children}: ProvidersProps) {
  return (
    <MemoryRouter>
      <ThemeProvider theme={theme}>{children}</ThemeProvider>
    </MemoryRouter>
  );
}

// Custom render function that includes providers
function customRender(ui: ReactElement, options?: Omit<RenderOptions, 'wrapper'>) {
  return render(ui, {wrapper: Providers, ...options});
}

export default customRender;
