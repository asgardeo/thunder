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

/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable @typescript-eslint/no-unsafe-assignment */
/* eslint-disable @typescript-eslint/no-unsafe-member-access */
/* eslint-disable @typescript-eslint/no-unsafe-return */

import {describe, it, expect, vi} from 'vitest';
import {render, screen} from '@testing-library/react';
import IconAdapter from '../IconAdapter';
import type {FlowComponent} from '../../../../models/flow';

vi.mock('@wso2/oxygen-ui', () => ({
  Box: ({children}: any) => <div data-testid="icon-box">{children}</div>,
}));

vi.mock('@wso2/oxygen-ui-icons-react', () => {
  const icons = {
    Star: ({fontSize, sx}: any) => <span data-testid="star-icon" data-font-size={fontSize} data-color={sx?.color} />,
    ArrowLeftRight: ({fontSize}: any) => <span data-testid="arrow-left-right-icon" data-font-size={fontSize} />,
  };
  return new Proxy(icons, {
    get(target, prop: string) {
      return prop in target ? (target as any)[prop] : undefined;
    },
  });
});

const baseComponent: FlowComponent = {
  id: 'icon-1',
  type: 'ICON',
  name: 'Star',
};

describe('IconAdapter', () => {
  it('renders the named icon', () => {
    render(<IconAdapter component={baseComponent} />);
    expect(screen.getByTestId('star-icon')).toBeInTheDocument();
  });

  it('wraps the icon in a Box container', () => {
    render(<IconAdapter component={baseComponent} />);
    expect(screen.getByTestId('icon-box')).toBeInTheDocument();
  });

  it('passes the size prop to the icon', () => {
    const component = {...baseComponent, size: 32};
    render(<IconAdapter component={component} />);
    expect(screen.getByTestId('star-icon')).toHaveAttribute('data-font-size', '32');
  });

  it('defaults size to 24 when not provided', () => {
    render(<IconAdapter component={baseComponent} />);
    expect(screen.getByTestId('star-icon')).toHaveAttribute('data-font-size', '24');
  });

  it('passes the color prop to the icon', () => {
    const component = {...baseComponent, color: '#ff0000'};
    render(<IconAdapter component={component} />);
    expect(screen.getByTestId('star-icon')).toHaveAttribute('data-color', '#ff0000');
  });

  it('defaults color to currentColor when not provided', () => {
    render(<IconAdapter component={baseComponent} />);
    expect(screen.getByTestId('star-icon')).toHaveAttribute('data-color', 'currentColor');
  });

  it('uses ArrowLeftRight as default icon when name is not provided', () => {
    const component = {...baseComponent, name: undefined};
    render(<IconAdapter component={component} />);
    expect(screen.getByTestId('arrow-left-right-icon')).toBeInTheDocument();
  });

  it('returns null for an unknown icon name', () => {
    const component = {...baseComponent, name: 'NonExistentIcon'};
    const {container} = render(<IconAdapter component={component} />);
    expect(container.firstChild).toBeNull();
  });
});
