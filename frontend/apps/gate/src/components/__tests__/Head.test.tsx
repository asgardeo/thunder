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

import {render, cleanup} from '@testing-library/react';
import type {ReactNode} from 'react';
import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';
import Head from '../Head';

const mockUseConfig = vi.hoisted(() => vi.fn());
vi.mock('@thunderid/contexts', () => ({
  useConfig: mockUseConfig,
}));

const mockUseColorScheme = vi.hoisted(() => vi.fn());
vi.mock('@wso2/oxygen-ui', () => ({
  useColorScheme: mockUseColorScheme,
}));

vi.mock('@thunderid/components', () => ({
  Helmet: ({children = undefined}: {children?: ReactNode}) => children,
}));

const defaultFavicon = {
  light: 'assets/images/favicon.ico',
  dark: 'assets/images/favicon-inverted.ico',
};

describe('Head', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseConfig.mockReturnValue({config: {brand: {favicon: defaultFavicon}}});
  });

  afterEach(() => {
    cleanup();
    document.head.querySelectorAll('link[rel="icon"]').forEach((el) => el.remove());
  });

  it('renders a single favicon link tag', () => {
    mockUseColorScheme.mockReturnValue({mode: 'light', systemMode: 'light'});
    render(<Head />);
    expect(document.head.querySelectorAll('link[rel="icon"]')).toHaveLength(1);
  });

  it('renders the light favicon when mode is "light"', () => {
    mockUseColorScheme.mockReturnValue({mode: 'light', systemMode: 'light'});
    render(<Head />);
    expect(document.head.querySelector('link[rel="icon"]')).toHaveAttribute('href', defaultFavicon.light);
  });

  it('renders the dark favicon when mode is "dark"', () => {
    mockUseColorScheme.mockReturnValue({mode: 'dark', systemMode: 'dark'});
    render(<Head />);
    expect(document.head.querySelector('link[rel="icon"]')).toHaveAttribute('href', defaultFavicon.dark);
  });

  it('renders the light favicon when mode is "system" and systemMode is "light"', () => {
    mockUseColorScheme.mockReturnValue({mode: 'system', systemMode: 'light'});
    render(<Head />);
    expect(document.head.querySelector('link[rel="icon"]')).toHaveAttribute('href', defaultFavicon.light);
  });

  it('renders the dark favicon when mode is "system" and systemMode is "dark"', () => {
    mockUseColorScheme.mockReturnValue({mode: 'system', systemMode: 'dark'});
    render(<Head />);
    expect(document.head.querySelector('link[rel="icon"]')).toHaveAttribute('href', defaultFavicon.dark);
  });

  it('falls back to the light favicon when mode is neither "dark" nor "system"', () => {
    mockUseColorScheme.mockReturnValue({mode: undefined, systemMode: undefined});
    render(<Head />);
    expect(document.head.querySelector('link[rel="icon"]')).toHaveAttribute('href', defaultFavicon.light);
  });

  it('reflects custom favicon paths from config', () => {
    mockUseConfig.mockReturnValue({
      config: {
        brand: {
          favicon: {light: 'custom/light.ico', dark: 'custom/dark.ico'},
        },
      },
    });
    mockUseColorScheme.mockReturnValue({mode: 'dark', systemMode: 'dark'});
    render(<Head />);
    expect(document.head.querySelector('link[rel="icon"]')).toHaveAttribute('href', 'custom/dark.ico');
  });
});
