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

/* eslint-disable react/require-default-props */

import {describe, it, expect, vi} from 'vitest';
import {renderWithProviders} from '@thunder/test-utils/browser';
import {page} from 'vitest/browser';
import FullScreenLayout from '../FullScreenLayout';

// Mock Outlet
vi.mock('react-router', async () => {
  const actual = await vi.importActual<typeof import('react-router')>('react-router');
  return {
    ...actual,
    Outlet: () => <div data-testid="outlet">Outlet Content</div>,
  };
});

interface MockComponentProps {
  children: React.ReactNode;
  sx?: Record<string, unknown>;
}

// Mock @wso2/oxygen-ui Layout and Box
vi.mock('@wso2/oxygen-ui', async () => {
  const actual = await vi.importActual<typeof import('@wso2/oxygen-ui')>('@wso2/oxygen-ui');

  return {
    ...actual,
    Box: ({children, sx, ...props}: MockComponentProps) => (
      <div data-testid="box" data-sx={JSON.stringify(sx)} {...props}>
        {children}
      </div>
    ),
    Layout: Object.assign(
      ({children, sx, ...props}: MockComponentProps) => (
        <div data-testid="layout-root" data-sx={JSON.stringify(sx)} {...props}>
          {children}
        </div>
      ),
      {
        Content: ({children}: {children: React.ReactNode}) => (
          <div data-testid="layout-content">{children}</div>
        ),
      },
    ),
  };
});

describe('FullScreenLayout', () => {
  it('renders Layout component with minHeight 100vh', async () => {
    await renderWithProviders(<FullScreenLayout />);

    const layout = page.getByTestId('layout-root');
    await expect.element(layout).toBeInTheDocument();
    await expect.element(layout).toHaveAttribute('data-sx', JSON.stringify({minHeight: '100vh'}));
  });

  it('renders Layout.Content', async () => {
    await renderWithProviders(<FullScreenLayout />);

    await expect.element(page.getByTestId('layout-content')).toBeInTheDocument();
  });

  it('renders Box with minHeight 100vh', async () => {
    await renderWithProviders(<FullScreenLayout />);

    const box = page.getByTestId('box');
    await expect.element(box).toBeInTheDocument();
    await expect.element(box).toHaveAttribute('data-sx', JSON.stringify({minHeight: '100vh'}));
  });

  it('renders Outlet for nested routes', async () => {
    await renderWithProviders(<FullScreenLayout />);

    await expect.element(page.getByTestId('outlet')).toBeInTheDocument();
    await expect.element(page.getByTestId('outlet')).toHaveTextContent('Outlet Content');
  });

  it('renders complete layout structure in correct hierarchy', async () => {
    await renderWithProviders(<FullScreenLayout />);

    // Verify all components are rendered
    const layout = page.getByTestId('layout-root');
    const content = page.getByTestId('layout-content');
    const box = page.getByTestId('box');
    const outlet = page.getByTestId('outlet');

    await expect.element(layout).toBeInTheDocument();
    await expect.element(content).toBeInTheDocument();
    await expect.element(box).toBeInTheDocument();
    await expect.element(outlet).toBeInTheDocument();

    // Verify hierarchy: Layout > Layout.Content > Box > Outlet
    // Use scoped locators to verify containment
    await expect.element(layout.getByTestId('layout-content')).toBeInTheDocument();
    await expect.element(content.getByTestId('box')).toBeInTheDocument();
    await expect.element(box.getByTestId('outlet')).toBeInTheDocument();
  });
});
