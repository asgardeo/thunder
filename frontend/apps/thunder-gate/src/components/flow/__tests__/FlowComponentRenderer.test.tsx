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

import {describe, it, expect, vi} from 'vitest';
import {render, screen} from '@testing-library/react';
import FlowComponentRenderer from '../FlowComponentRenderer';
import type {FlowComponentRendererProps} from '../../../models/flow';

vi.mock('@asgardeo/react', () => ({
  EmbeddedFlowComponentType: {
    Text: 'TEXT',
    Block: 'BLOCK',
    TextInput: 'TEXT_INPUT',
    PasswordInput: 'PASSWORD_INPUT',
    Action: 'ACTION',
  },
  EmbeddedFlowEventType: {
    Submit: 'SUBMIT',
    Trigger: 'TRIGGER',
  },
}));

vi.mock('../adapters/TextAdapter', () => ({
  default: () => <div data-testid="text-adapter" />,
}));
vi.mock('../adapters/RichTextAdapter', () => ({
  default: () => <div data-testid="rich-text-adapter" />,
}));
vi.mock('../adapters/ImageAdapter', () => ({
  default: () => <div data-testid="image-adapter" />,
}));
vi.mock('../adapters/IconAdapter', () => ({
  default: () => <div data-testid="icon-adapter" />,
}));
vi.mock('../adapters/StackAdapter', () => ({
  default: () => <div data-testid="stack-adapter" />,
}));
vi.mock('../adapters/BlockAdapter', () => ({
  default: () => <div data-testid="block-adapter" />,
}));
vi.mock('../adapters/StandaloneTriggerAdapter', () => ({
  default: () => <div data-testid="standalone-trigger-adapter" />,
}));

const baseProps: FlowComponentRendererProps = {
  component: {id: 'c1', type: 'TEXT'},
  index: 0,
  values: {},
  isLoading: false,
  resolve: (s) => s,
  onInputChange: vi.fn(),
  onSubmit: vi.fn(),
};

describe('FlowComponentRenderer', () => {
  it('renders TextAdapter for TEXT type', () => {
    render(<FlowComponentRenderer {...baseProps} component={{id: 'c1', type: 'TEXT'}} />);
    expect(screen.getByTestId('text-adapter')).toBeInTheDocument();
  });

  it('renders TextAdapter for EmbeddedFlowComponentType.Text value', () => {
    // EmbeddedFlowComponentType.Text resolves to 'TEXT' in mock
    render(<FlowComponentRenderer {...baseProps} component={{id: 'c1', type: 'TEXT'}} />);
    expect(screen.getByTestId('text-adapter')).toBeInTheDocument();
  });

  it('renders RichTextAdapter for RICH_TEXT type', () => {
    render(<FlowComponentRenderer {...baseProps} component={{id: 'c2', type: 'RICH_TEXT'}} />);
    expect(screen.getByTestId('rich-text-adapter')).toBeInTheDocument();
  });

  it('renders ImageAdapter for IMAGE type', () => {
    render(<FlowComponentRenderer {...baseProps} component={{id: 'c3', type: 'IMAGE'}} />);
    expect(screen.getByTestId('image-adapter')).toBeInTheDocument();
  });

  it('renders IconAdapter for ICON type', () => {
    render(<FlowComponentRenderer {...baseProps} component={{id: 'c4', type: 'ICON'}} />);
    expect(screen.getByTestId('icon-adapter')).toBeInTheDocument();
  });

  it('renders StackAdapter for STACK type', () => {
    render(<FlowComponentRenderer {...baseProps} component={{id: 'c5', type: 'STACK'}} />);
    expect(screen.getByTestId('stack-adapter')).toBeInTheDocument();
  });

  it('renders BlockAdapter for BLOCK type', () => {
    render(<FlowComponentRenderer {...baseProps} component={{id: 'c6', type: 'BLOCK'}} />);
    expect(screen.getByTestId('block-adapter')).toBeInTheDocument();
  });

  it('renders StandaloneTriggerAdapter for ACTION/TRIGGER type', () => {
    render(
      <FlowComponentRenderer
        {...baseProps}
        component={{id: 'c7', type: 'ACTION', eventType: 'TRIGGER'}}
      />,
    );
    expect(screen.getByTestId('standalone-trigger-adapter')).toBeInTheDocument();
  });

  it('returns null for unrecognised component types', () => {
    const {container} = render(
      <FlowComponentRenderer {...baseProps} component={{id: 'c8', type: 'UNKNOWN_TYPE'}} />,
    );
    expect(container.firstChild).toBeNull();
  });

  it('returns null for ACTION type without TRIGGER eventType', () => {
    const {container} = render(
      <FlowComponentRenderer {...baseProps} component={{id: 'c9', type: 'ACTION', eventType: 'SUBMIT'}} />,
    );
    expect(container.firstChild).toBeNull();
  });

  it('passes maxImageSize to ImageAdapter', () => {
    // ImageAdapter mock doesn't expose this, but the render should not throw
    expect(() =>
      render(
        <FlowComponentRenderer {...baseProps} component={{id: 'img', type: 'IMAGE'}} maxImageSize={80} />,
      ),
    ).not.toThrow();
  });
});
