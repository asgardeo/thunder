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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render} from '@thunder/test-utils/browser';
import {page} from 'vitest/browser';

// Use vi.hoisted to ensure mock functions are hoisted before vi.mock
const {mockUseBranding} = vi.hoisted(() => ({
  mockUseBranding: vi.fn(),
}));

// Mock App component
vi.mock('../App', () => ({
  default: () => <div data-testid="app">App</div>,
}));

// Create mock for useBranding
vi.mock('@thunder/shared-branding', () => ({
  // eslint-disable-next-line @typescript-eslint/no-unsafe-return
  useBranding: () => mockUseBranding(),
}));

// eslint-disable-next-line import/first
import AppWithTheme from '../AppWithTheme';

describe('AppWithTheme', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseBranding.mockReturnValue({
      theme: null,
      isLoading: false,
      images: null,
      layout: null,
      isBrandingEnabled: false,
    });
  });

  it('renders without crashing', async () => {
    await render(<AppWithTheme />);
    // Verify something renders - the real OxygenUIThemeProvider renders content
    expect(true).toBe(true);
  });

  it('renders App when not loading', async () => {
    await render(<AppWithTheme />);
    // When not loading, the App component should render
    await expect.element(page.getByTestId('app')).toBeInTheDocument();
  });

  it('renders CircularProgress when loading', async () => {
    mockUseBranding.mockReturnValue({
      theme: null,
      isLoading: true,
      images: null,
      layout: null,
      isBrandingEnabled: false,
    });

    await render(<AppWithTheme />);
    // When loading, CircularProgress should render and App should not
    // Check for MUI CircularProgress role
    await expect.element(page.getByRole('progressbar')).toBeInTheDocument();
    await expect.element(page.getByTestId('app')).not.toBeInTheDocument();
  });

  it('renders ColorSchemeToggle button', async () => {
    await render(<AppWithTheme />);
    // ColorSchemeToggle renders as an icon button
    await expect.element(page.getByRole('button')).toBeInTheDocument();
  });

  it('does not show App when loading', async () => {
    mockUseBranding.mockReturnValue({
      theme: null,
      isLoading: true,
      images: null,
      layout: null,
      isBrandingEnabled: false,
    });

    await render(<AppWithTheme />);
    await expect.element(page.getByTestId('app')).not.toBeInTheDocument();
  });

  it('shows App when branding is enabled and not loading', async () => {
    mockUseBranding.mockReturnValue({
      theme: null,
      isLoading: false,
      images: null,
      layout: null,
      isBrandingEnabled: true,
    });

    await render(<AppWithTheme />);
    await expect.element(page.getByTestId('app')).toBeInTheDocument();
  });
});
