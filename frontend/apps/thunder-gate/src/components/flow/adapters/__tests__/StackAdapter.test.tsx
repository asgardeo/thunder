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

import {describe, it, expect, vi} from 'vitest';
import {render, screen} from '@testing-library/react';
import StackAdapter from '../StackAdapter';
import type {FlowComponent} from '../../../../models/flow';

vi.mock('@wso2/oxygen-ui', () => ({
  Stack: ({children, direction, spacing, alignItems, justifyContent}: any) => (
    <div
      data-testid="stack"
      data-direction={direction}
      data-spacing={String(spacing)}
      data-align={alignItems}
      data-justify={justifyContent}
    >
      {children}
    </div>
  ),
}));

vi.mock('../../FlowComponentRenderer', () => ({
  default: ({component, maxImageSize}: any) => (
    <div data-testid="flow-component-renderer" data-type={component.type} data-max-image-size={maxImageSize} />
  ),
}));

const baseComponent: FlowComponent = {
  id: 'stack-1',
  type: 'STACK',
  direction: 'row',
  gap: 3,
  align: 'center',
  justify: 'space-between',
  components: [
    {id: 'child-1', type: 'TEXT'},
    {id: 'child-2', type: 'IMAGE'},
  ],
};

const baseProps = {
  component: baseComponent,
  resolve: (s: string | undefined) => s,
  values: {},
  isLoading: false,
  onInputChange: vi.fn(),
  onSubmit: vi.fn(),
};

describe('StackAdapter', () => {
  it('renders a Stack container', () => {
    render(<StackAdapter {...baseProps} />);
    expect(screen.getByTestId('stack')).toBeInTheDocument();
  });

  it('passes direction from component to Stack', () => {
    render(<StackAdapter {...baseProps} />);
    expect(screen.getByTestId('stack')).toHaveAttribute('data-direction', 'row');
  });

  it('passes gap from component as spacing to Stack', () => {
    render(<StackAdapter {...baseProps} />);
    expect(screen.getByTestId('stack')).toHaveAttribute('data-spacing', '3');
  });

  it('passes align from component as alignItems to Stack', () => {
    render(<StackAdapter {...baseProps} />);
    expect(screen.getByTestId('stack')).toHaveAttribute('data-align', 'center');
  });

  it('passes justify from component as justifyContent to Stack', () => {
    render(<StackAdapter {...baseProps} />);
    expect(screen.getByTestId('stack')).toHaveAttribute('data-justify', 'space-between');
  });

  it('defaults direction to column when not specified', () => {
    const props = {...baseProps, component: {...baseComponent, direction: undefined}};
    render(<StackAdapter {...props} />);
    expect(screen.getByTestId('stack')).toHaveAttribute('data-direction', 'column');
  });

  it('defaults gap to 2 when not specified', () => {
    const props = {...baseProps, component: {...baseComponent, gap: undefined}};
    render(<StackAdapter {...props} />);
    expect(screen.getByTestId('stack')).toHaveAttribute('data-spacing', '2');
  });

  it('renders a FlowComponentRenderer for each nested component', () => {
    render(<StackAdapter {...baseProps} />);
    expect(screen.getAllByTestId('flow-component-renderer')).toHaveLength(2);
  });

  it('passes the correct type to each FlowComponentRenderer', () => {
    render(<StackAdapter {...baseProps} />);
    const renderers = screen.getAllByTestId('flow-component-renderer');
    expect(renderers[0]).toHaveAttribute('data-type', 'TEXT');
    expect(renderers[1]).toHaveAttribute('data-type', 'IMAGE');
  });

  it('passes STACK_IMAGE_MAX_SIZE (80) as maxImageSize to each child renderer', () => {
    render(<StackAdapter {...baseProps} />);
    screen.getAllByTestId('flow-component-renderer').forEach((el) => {
      expect(el).toHaveAttribute('data-max-image-size', '80');
    });
  });

  it('renders nothing when components list is empty', () => {
    const props = {...baseProps, component: {...baseComponent, components: []}};
    render(<StackAdapter {...props} />);
    expect(screen.queryAllByTestId('flow-component-renderer')).toHaveLength(0);
  });

  it('renders nothing when components is undefined', () => {
    const props = {...baseProps, component: {...baseComponent, components: undefined}};
    render(<StackAdapter {...props} />);
    expect(screen.queryAllByTestId('flow-component-renderer')).toHaveLength(0);
  });
});
