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

import {afterEach, beforeEach, describe, it, expect, vi} from 'vitest';
import {render, screen} from '@testing-library/react';
import FlowComponentRenderer from '../FlowComponentRenderer';
import type {FlowComponentRendererProps} from '../../../models/flow';

const {blockPropsSpy, consentPropsSpy, timerPropsSpy} = vi.hoisted(() => ({
  blockPropsSpy: vi.fn(),
  consentPropsSpy: vi.fn(),
  timerPropsSpy: vi.fn(),
}));

vi.mock('@asgardeo/react', () => ({
  EmbeddedFlowComponentType: {
    Action: 'ACTION',
    Block: 'BLOCK',
    Consent: 'CONSENT',
    Icon: 'ICON',
    Image: 'IMAGE',
    Stack: 'STACK',
    Text: 'TEXT',
    Timer: 'TIMER',
  },
  EmbeddedFlowEventType: {
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
  default: (props: any) => {
    blockPropsSpy(props);
    return <div data-testid="block-adapter" data-loading={String(props.isLoading)} />;
  },
}));
vi.mock('../adapters/StandaloneTriggerAdapter', () => ({
  default: () => <div data-testid="standalone-trigger-adapter" />,
}));
vi.mock('../adapters/ConsentAdapter', () => ({
  default: (props: any) => {
    consentPropsSpy(props);
    return <div data-testid="consent-adapter" />;
  },
}));
vi.mock('../adapters/TimerAdapter', () => ({
  default: (props: any) => {
    timerPropsSpy(props);
    return <div data-testid="timer-adapter" data-expires-in={String(props.expiresIn)} />;
  },
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
  beforeEach(() => {
    blockPropsSpy.mockClear();
    consentPropsSpy.mockClear();
    timerPropsSpy.mockClear();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

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
    expect(screen.queryByTestId('consent-adapter')).toBeNull();
    expect(screen.queryByTestId('timer-adapter')).toBeNull();
    expect(blockPropsSpy).toHaveBeenCalledWith(expect.objectContaining({isLoading: false}));
  });

  it('injects ConsentAdapter for BLOCK when consentPrompt is present', () => {
    render(
      <FlowComponentRenderer
        {...baseProps}
        component={{id: 'block-consent', type: 'BLOCK'}}
        additionalData={{
          consentPrompt: {
            purposes: [
              {
                essential: ['email'],
                optional: ['givenName'],
                purpose_id: 'p1',
                purpose_name: 'Profile',
              },
            ],
          },
        }}
      />,
    );

    expect(screen.getByTestId('consent-adapter')).toBeInTheDocument();
    expect(screen.getByTestId('block-adapter')).toBeInTheDocument();
    expect(screen.queryByTestId('timer-adapter')).toBeNull();
    expect(consentPropsSpy).toHaveBeenCalledWith(
      expect.objectContaining({
        consentData: expect.any(Object),
      }),
    );
  });

  it('renders TimerAdapter for TIMER component type', () => {
    vi.spyOn(Date, 'now').mockReturnValue(1000);

    render(
      <FlowComponentRenderer
        {...baseProps}
        component={{id: 'timer-1', type: 'TIMER', label: 'Expires in {time}'}}
        additionalData={{stepTimeout: String(7000)}}
      />,
    );

    expect(screen.getByTestId('timer-adapter')).toBeInTheDocument();
    expect(screen.getByTestId('timer-adapter')).toHaveAttribute('data-expires-in', '6');
    expect(timerPropsSpy).toHaveBeenCalledWith(expect.objectContaining({textTemplate: 'Expires in {time}'}));
  });

  it('does not render TimerAdapter inside BLOCK when stepTimeout is present', () => {
    vi.spyOn(Date, 'now').mockReturnValue(1000);

    render(
      <FlowComponentRenderer
        {...baseProps}
        component={{id: 'block-timer', type: 'BLOCK'}}
        additionalData={{stepTimeout: String(7000)}}
      />,
    );

    expect(screen.queryByTestId('timer-adapter')).toBeNull();
    expect(screen.getByTestId('block-adapter')).toHaveAttribute('data-loading', 'false');
  });

  it('marks BlockAdapter as loading when stepTimeout is already expired on mount', () => {
    vi.spyOn(Date, 'now').mockReturnValue(5000);

    render(
      <FlowComponentRenderer
        {...baseProps}
        component={{id: 'block-expired', type: 'BLOCK'}}
        additionalData={{stepTimeout: String(3000)}}
      />,
    );

    expect(screen.queryByTestId('timer-adapter')).toBeNull();
    expect(screen.getByTestId('block-adapter')).toHaveAttribute('data-loading', 'true');
  });

  it('keeps BlockAdapter loading true when parent isLoading is true', () => {
    vi.spyOn(Date, 'now').mockReturnValue(1000);

    render(
      <FlowComponentRenderer
        {...baseProps}
        isLoading
        component={{id: 'block-loading', type: 'BLOCK'}}
        additionalData={{stepTimeout: String(9000)}}
      />,
    );

    expect(screen.getByTestId('block-adapter')).toHaveAttribute('data-loading', 'true');
  });

  it('renders StandaloneTriggerAdapter for ACTION/TRIGGER type', () => {
    render(<FlowComponentRenderer {...baseProps} component={{id: 'c7', type: 'ACTION', eventType: 'TRIGGER'}} />);
    expect(screen.getByTestId('standalone-trigger-adapter')).toBeInTheDocument();
  });

  it('returns null for unrecognised component types', () => {
    const {container} = render(<FlowComponentRenderer {...baseProps} component={{id: 'c8', type: 'UNKNOWN_TYPE'}} />);
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
      render(<FlowComponentRenderer {...baseProps} component={{id: 'img', type: 'IMAGE'}} maxImageSize={80} />),
    ).not.toThrow();
  });
});
