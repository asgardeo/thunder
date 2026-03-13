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
import withTheme from '../withTheme';

let capturedTheme: unknown;

function MockChild() {
  return <div data-testid="mock-child">Child</div>;
}
const WithThemeComponent = withTheme(MockChild);

vi.mock('@wso2/oxygen-ui', () => ({
  AcrylicOrangeTheme: {palette: {primary: {main: '#ff5700'}}},
  OxygenUIThemeProvider: ({
    children,
    theme = {palette: {primary: {main: '#ff5700'}}},
  }: {
    children: React.ReactNode;
    theme?: unknown;
  }) => {
    capturedTheme = theme;
    return <div data-testid="theme-provider">{children}</div>;
  },
}));

describe('withTheme (console)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    capturedTheme = undefined;
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

  it('uses AcrylicOrangeTheme', () => {
    render(<WithThemeComponent />);
    expect(capturedTheme).toEqual({palette: {primary: {main: '#ff5700'}}});
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
});
