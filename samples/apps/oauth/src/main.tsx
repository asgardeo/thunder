/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import ThemeProvider from './theme/ThemeProvider.tsx'
import { ensureConfigReady } from './config.tsx'

// Ensure configuration is loaded before rendering the app
ensureConfigReady().then(() => {
  createRoot(document.getElementById('root')!).render(
    <StrictMode>
      <ThemeProvider>
        <App />
      </ThemeProvider>
    </StrictMode>,
  )
}).catch((error) => {
  console.error('Failed to initialize application configuration:', error);
  // Render error state or fallback UI
  document.getElementById('root')!.innerHTML = '<div style="padding: 20px; text-align: center;"><h1>Configuration Error</h1><p>Failed to load application configuration. Please refresh the page.</p></div>';
});
