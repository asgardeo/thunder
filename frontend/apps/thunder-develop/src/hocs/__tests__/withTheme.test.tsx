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
import {within} from '@testing-library/react';
import withTheme from '../withTheme';

// Use vi.hoisted so capturedTheme is available inside vi.mock factory
const {capturedTheme} = vi.hoisted(() => ({
  capturedTheme: {current: undefined as unknown},
}));

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
    capturedTheme.current = theme;
    return <div data-testid="theme-provider">{children}</div>;
  },
}));

describe('withTheme (thunder-develop)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    capturedTheme.current = undefined;
  });

  it('renders without crashing', async () => {
    const {container} = await render(<WithThemeComponent />);
    expect(container).toBeInTheDocument();
  });

  it('renders the wrapped component', async () => {
    const {container} = await render(<WithThemeComponent />);
    expect(within(container).getByTestId('mock-child')).toBeInTheDocument();
  });

  it('wraps with OxygenUIThemeProvider', async () => {
    const {container} = await render(<WithThemeComponent />);
    // The mock-child is wrapped inside the theme-provider rendered by WithThemeComponent
    const mockChild = within(container).getByTestId('mock-child');
    expect(mockChild.closest('[data-testid="theme-provider"]')).toBeInTheDocument();
  });

  it('uses AcrylicOrangeTheme', async () => {
    await render(<WithThemeComponent />);
    expect(capturedTheme.current).toEqual({palette: {primary: {main: '#ff5700'}}});
  });

  it('wraps different components correctly', async () => {
    function AnotherChild() {
      return <div data-testid="another-child">Another</div>;
    }
    const AnotherWrapped = withTheme(AnotherChild);

    const {container} = await render(<AnotherWrapped />);
    expect(within(container).getByTestId('another-child')).toBeInTheDocument();
    // The another-child is wrapped inside the theme-provider rendered by withTheme
    const anotherChild = within(container).getByTestId('another-child');
    expect(anotherChild.closest('[data-testid="theme-provider"]')).toBeInTheDocument();
  });

  it('passes props through to the wrapped component', async () => {
    function PropsChild({label}: {label: string}) {
      return <div data-testid="props-child">{label}</div>;
    }
    const WrappedWithProps = withTheme(PropsChild);

    const {container} = await render(<WrappedWithProps label="test-label" />);
    expect(within(container).getByTestId('props-child')).toHaveTextContent('test-label');
  });
});
