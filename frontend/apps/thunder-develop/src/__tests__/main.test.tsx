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

import {describe, it, expect, vi} from 'vitest';

// Mock ReactDOM before importing main
const mockRender = vi.fn();
const mockUnmount = vi.fn();
const mockCreateRoot = vi.fn(() => ({
  render: mockRender,
  unmount: mockUnmount,
}));

vi.mock('react-dom/client', () => ({
  default: {
    createRoot: mockCreateRoot,
  },
  createRoot: mockCreateRoot,
}));

// Mock ConfigProvider
vi.mock('@thunder/shared-contexts', () => ({
  ConfigProvider: ({children}: {children: React.ReactNode}) => <div data-testid="config-provider">{children}</div>,
}));

// Mock AppWithConfig
vi.mock('../AppWithConfig', () => ({
  default: () => <div data-testid="app-with-config">AppWithConfig</div>,
}));

describe('main', () => {
  it('imports without errors', async () => {
    // Set up DOM
    document.body.innerHTML = '<div id="root"></div>';

    // Import main module
    await import('../main');

    // Verify createRoot was called
    expect(mockCreateRoot).toHaveBeenCalled();
  });

  it('calls render on the root', async () => {
    // Set up DOM
    document.body.innerHTML = '<div id="root"></div>';

    // Import main module
    await import('../main');

    // Verify render was called
    expect(mockRender).toHaveBeenCalled();
  });
});
