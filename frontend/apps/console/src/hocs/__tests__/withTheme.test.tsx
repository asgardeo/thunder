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

import {render, screen} from '@testing-library/react';
import {describe, it, expect, vi, beforeEach} from 'vitest';

const mockUseConfig = vi.hoisted(() => vi.fn());
vi.mock('@thunderid/contexts', () => ({
  useConfig: mockUseConfig,
}));

vi.mock('../../components/Head', () => ({
  default: () => <div data-testid="head" />,
}));

let capturedThemes: unknown;
let capturedInitialTheme: unknown;

function MockChild() {
  return <div data-testid="mock-child">Child</div>;
}

vi.mock('@wso2/oxygen-ui', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@wso2/oxygen-ui')>();
  return {
    ...actual,
    createOxygenTheme: actual.createOxygenTheme ?? ((theme: unknown) => theme),
    HighContrastTheme: actual.HighContrastTheme ?? {},
    OxygenUIThemeProvider: ({
      children,
      themes = undefined,
      initialTheme = undefined,
    }: {
      children: React.ReactNode;
      themes?: unknown;
      initialTheme?: unknown;
    }) => {
      capturedThemes = themes;
      capturedInitialTheme = initialTheme;
      return <div data-testid="theme-provider">{children}</div>;
    },
  };
});

const {default: withTheme} = await import('../withTheme');
const WithThemeComponent = withTheme(MockChild);

describe('withTheme (console)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    capturedThemes = undefined;
    capturedInitialTheme = undefined;
    mockUseConfig.mockReturnValue({
      config: {
        brand: {
          product_name: 'ThunderID',
          favicon: {light: 'assets/images/favicon.ico', dark: 'assets/images/favicon-inverted.ico'},
        },
      },
    });
  });

  it('renders without crashing', () => {
    const {container} = render(<WithThemeComponent />);
    expect(container).toBeInTheDocument();
  });

  it('renders the wrapped component', () => {
    render(<WithThemeComponent />);
    expect(screen.getByTestId('mock-child')).toBeInTheDocument();
  });

  it('wraps with OxygenUIThemeProvider', () => {
    render(<WithThemeComponent />);
    expect(screen.getByTestId('theme-provider')).toBeInTheDocument();
  });

  it('provides the expected theme collection', () => {
    render(<WithThemeComponent />);
    expect(Array.isArray(capturedThemes)).toBe(true);
    expect(capturedThemes).toEqual(
      expect.arrayContaining([
        expect.objectContaining({key: 'highContrast'}),
        expect.objectContaining({key: 'default'}),
      ]),
    );
    expect(capturedInitialTheme).toBe('default');
  });

  it('wraps different components correctly', () => {
    function AnotherChild() {
      return <div data-testid="another-child">Another</div>;
    }
    const AnotherWrapped = withTheme(AnotherChild);

    render(<AnotherWrapped />);
    expect(screen.getByTestId('another-child')).toBeInTheDocument();
    expect(screen.getByTestId('theme-provider')).toBeInTheDocument();
  });

  it('passes props through to the wrapped component', () => {
    function PropsChild({label}: {label: string}) {
      return <div data-testid="props-child">{label}</div>;
    }
    const WrappedWithProps = withTheme(PropsChild);

    render(<WrappedWithProps label="test-label" />);
    expect(screen.getByTestId('props-child')).toHaveTextContent('test-label');
  });

  it('renders Head', () => {
    render(<WithThemeComponent />);
    expect(screen.getByTestId('head')).toBeInTheDocument();
  });

  it('includes custom object themes from config in the theme list', () => {
    const objectTheme = {palette: {primary: {main: '#aabbcc'}}};
    mockUseConfig.mockReturnValue({
      config: {
        brand: {
          design: {
            themes: [{key: 'custom', label: 'Custom Theme', theme: objectTheme}],
          },
        },
      },
    });

    render(<WithThemeComponent />);
    expect(capturedThemes).toEqual(
      expect.arrayContaining([expect.objectContaining({key: 'custom', label: 'Custom Theme'})]),
    );
  });

  it('includes custom string themes from config in the theme list', () => {
    mockUseConfig.mockReturnValue({
      config: {
        brand: {
          design: {
            themes: [{key: 'external', label: 'External Theme', theme: 'https://example.com/theme.json'}],
          },
        },
      },
    });

    render(<WithThemeComponent />);
    expect(capturedThemes).toEqual(
      expect.arrayContaining([
        expect.objectContaining({key: 'external', label: 'External Theme', theme: 'https://example.com/theme.json'}),
      ]),
    );
  });
});
