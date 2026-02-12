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
import {render, screen} from '@testing-library/react';
import AppWithTheme from '../AppWithTheme';

// Track the theme passed to OxygenUIThemeProvider
let capturedThemeProviderProps: Record<string, unknown> | undefined;

// Mock App component
vi.mock('../App', () => ({
  default: () => <div data-testid="app">App</div>,
}));

// Create mock for useBranding
const mockUseBranding = vi.fn();
vi.mock('@thunder/shared-branding', () => ({
  // eslint-disable-next-line @typescript-eslint/no-unsafe-return
  useBranding: () => mockUseBranding(),
}));

// Mock OxygenUI components - capture props passed to theme provider
vi.mock('@wso2/oxygen-ui', () => ({
  OxygenUIThemeProvider: ({children, ...rest}: {children: React.ReactNode; theme?: unknown; radialBackground?: boolean}) => {
    capturedThemeProviderProps = {...rest};
    return <div data-testid="theme-provider">{children}</div>;
  },
  ColorSchemeToggle: () => <div data-testid="color-scheme-toggle">Toggle</div>,
  CircularProgress: () => <div data-testid="circular-progress">Loading...</div>,
  Box: ({children}: {children: React.ReactNode}) => <div data-testid="box">{children}</div>,
}));

describe('AppWithTheme', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    capturedThemeProviderProps = undefined;
    mockUseBranding.mockReturnValue({
      theme: null,
      isLoading: false,
      images: null,
      layout: null,
      isBrandingEnabled: false,
    });
  });

  it('renders without crashing', () => {
    const {container} = render(<AppWithTheme />);
    expect(container).toBeInTheDocument();
  });

  it('renders OxygenUIThemeProvider', () => {
    render(<AppWithTheme />);
    expect(screen.getByTestId('theme-provider')).toBeInTheDocument();
  });

  it('renders ColorSchemeToggle', () => {
    render(<AppWithTheme />);
    expect(screen.getByTestId('color-scheme-toggle')).toBeInTheDocument();
  });

  it('renders App when not loading', () => {
    render(<AppWithTheme />);
    expect(screen.getByTestId('app')).toBeInTheDocument();
  });

  it('renders CircularProgress when loading', () => {
    mockUseBranding.mockReturnValue({
      theme: null,
      isLoading: true,
      images: null,
      layout: null,
      isBrandingEnabled: false,
    });

    render(<AppWithTheme />);
    expect(screen.getByTestId('circular-progress')).toBeInTheDocument();
    expect(screen.queryByTestId('app')).not.toBeInTheDocument();
  });

  it('does not pass theme to OxygenUIThemeProvider when theme is null', () => {
    mockUseBranding.mockReturnValue({
      theme: null,
      isLoading: false,
      images: null,
      layout: null,
      isBrandingEnabled: false,
    });

    render(<AppWithTheme />);
    expect(capturedThemeProviderProps).toBeDefined();
    expect(capturedThemeProviderProps?.theme).toBeUndefined();
    expect(capturedThemeProviderProps?.radialBackground).toBe(true);
  });

  it('does not pass theme to OxygenUIThemeProvider when theme is undefined', () => {
    mockUseBranding.mockReturnValue({
      theme: undefined,
      isLoading: false,
      images: null,
      layout: null,
      isBrandingEnabled: false,
    });

    render(<AppWithTheme />);
    expect(capturedThemeProviderProps).toBeDefined();
    expect(capturedThemeProviderProps?.theme).toBeUndefined();
  });

  it('passes theme to OxygenUIThemeProvider when theme is available', () => {
    const mockTheme = {palette: {primary: {main: '#ff0000'}}};
    mockUseBranding.mockReturnValue({
      theme: mockTheme,
      isLoading: false,
      images: null,
      layout: null,
      isBrandingEnabled: true,
    });

    render(<AppWithTheme />);
    expect(screen.getByTestId('theme-provider')).toBeInTheDocument();
    expect(screen.getByTestId('app')).toBeInTheDocument();
    expect(capturedThemeProviderProps?.theme).toEqual(mockTheme);
  });

  it('shows loading spinner when isLoading is true and theme is present', () => {
    const mockTheme = {palette: {primary: {main: '#ff0000'}}};
    mockUseBranding.mockReturnValue({
      theme: mockTheme,
      isLoading: true,
      images: null,
      layout: null,
      isBrandingEnabled: true,
    });

    render(<AppWithTheme />);
    expect(screen.getByTestId('circular-progress')).toBeInTheDocument();
    expect(screen.queryByTestId('app')).not.toBeInTheDocument();
  });
});
