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

import {describe, it, expect, vi} from 'vitest';
import {render} from 'vitest-browser-react';
import {page} from 'vitest/browser';

// Mock the app routes module to return simple routes
// This avoids loading complex page components with many dependencies
// Note: We use render from vitest-browser-react directly because App has its own BrowserRouter
vi.mock('../config/appRoutes', () => ({
  default: [
    {
      path: '/',
      element: <div data-testid="mock-layout">Mock Layout</div>,
      children: [
        {
          path: 'test',
          element: <div data-testid="mock-page">Mock Page</div>,
        },
      ],
    },
  ],
}));

// Import App after the mock is set up
// eslint-disable-next-line import/first
import App from '../App';

describe('App', () => {
  it('renders without crashing', async () => {
    // App uses BrowserRouter internally, so we render it directly
    // The test verifies the component mounts without errors
    await render(<App />);

    // The App should render and show content - verify the body is not empty
    // We're testing that the routing structure works, not specific route content
    await expect.element(page.getByTestId('mock-layout')).toBeInTheDocument();
  });

  it('renders the mocked route layout', async () => {
    await render(<App />);
    await expect.element(page.getByTestId('mock-layout')).toBeInTheDocument();
  });
});
