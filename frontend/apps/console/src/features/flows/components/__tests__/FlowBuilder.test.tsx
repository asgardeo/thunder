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
import {render, screen} from '@testing-library/react';
import type {Element} from '@/features/flows/models/elements';
import type {Step} from '@/features/flows/models/steps';
import FlowBuilder from '../FlowBuilder';

// Mock @xyflow/react
vi.mock('@xyflow/react', () => ({
  ReactFlowProvider: ({children}: {children: React.ReactNode}) => (
    <div data-testid="react-flow-provider">{children}</div>
  ),
}));

// Mock DecoratedVisualFlow
vi.mock('../visual-flow/DecoratedVisualFlow', () => ({
  default: (props: Record<string, unknown>) => (
    <div data-testid="decorated-visual-flow" data-props={JSON.stringify(Object.keys(props))}>
      Decorated Visual Flow
    </div>
  ),
}));

describe('FlowBuilder', () => {
  const mockResources = {
    steps: [],
    executors: [],
    templates: [],
    widgets: [],
    elements: [],
  };

  const mockProps = {
    resources: mockResources,
    flowTitle: 'Test Flow',
    flowHandle: 'test-flow',
    onFlowTitleChange: vi.fn(),
    mutateComponents: vi.fn((components: Element[]) => components),
    onTemplateLoad: vi.fn(() => [[], [], undefined, undefined] as [never[], never[], undefined, undefined]),
    onWidgetLoad: vi.fn(() => [[], [], null, null] as [never[], never[], null, null]),
    onStepLoad: vi.fn((step: Step) => step),
    nodes: [],
    edges: [],
    onResourceAdd: vi.fn(),
    setNodes: vi.fn(),
    setEdges: vi.fn(),
    onNodesChange: vi.fn(),
    onEdgesChange: vi.fn(),
  };

  it('should render ReactFlowProvider wrapper', () => {
    render(<FlowBuilder {...mockProps} />);

    expect(screen.getByTestId('react-flow-provider')).toBeInTheDocument();
  });

  it('should render DecoratedVisualFlow inside provider', () => {
    render(<FlowBuilder {...mockProps} />);

    expect(screen.getByTestId('decorated-visual-flow')).toBeInTheDocument();
  });

  it('should pass props to DecoratedVisualFlow', () => {
    render(<FlowBuilder {...mockProps} />);

    const decoratedFlow = screen.getByTestId('decorated-visual-flow');
    const propsKeys = JSON.parse(decoratedFlow.getAttribute('data-props') ?? '[]') as string[];

    expect(propsKeys).toContain('resources');
    expect(propsKeys).toContain('flowTitle');
    expect(propsKeys).toContain('flowHandle');
    expect(propsKeys).toContain('onFlowTitleChange');
  });

  it('should render with minimal required props', () => {
    render(<FlowBuilder {...mockProps} />);

    expect(screen.getByText('Decorated Visual Flow')).toBeInTheDocument();
  });

  it('should render with additional optional props', () => {
    const extendedProps = {
      ...mockProps,
      initialNodes: [{id: 'node-1', position: {x: 0, y: 0}, data: {}}],
      initialEdges: [{id: 'edge-1', source: 'node-1', target: 'node-2'}],
      triggerAutoLayoutOnLoad: true,
      onSave: vi.fn(),
    };

    render(<FlowBuilder {...extendedProps} />);

    expect(screen.getByTestId('react-flow-provider')).toBeInTheDocument();
    expect(screen.getByTestId('decorated-visual-flow')).toBeInTheDocument();
  });
});
