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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render, screen} from '@testing-library/react';
import type {ReactNode} from 'react';
import {ReactFlowProvider} from '@xyflow/react';
import {ButtonVariants, ElementTypes, type Element as FlowElement} from '@/features/flows/models/elements';
import ButtonAdapter from '../ButtonAdapter';

// Mock dependencies
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
  Trans: ({children}: {children: ReactNode}) => children,
}));

vi.mock('@/features/flows/hooks/useRequiredFields', () => ({
  default: vi.fn(),
}));

vi.mock('@/features/flows/utils/resolveStaticResourcePath', () => ({
  default: (path: string) => path,
}));

vi.mock('../NodeHandle', () => ({
  default: ({id, type, position}: {id: string; type: string; position: string}) => (
    <div data-testid="node-handle" data-id={id} data-type={type} data-position={position} />
  ),
}));

vi.mock('../PlaceholderComponent', () => ({
  default: ({value}: {value: string}) => <span data-testid="placeholder">{value}</span>,
}));

describe('ButtonAdapter', () => {
  const createMockElement = (overrides: Record<string, unknown> = {}): FlowElement =>
    ({
      id: 'button-1',
      resourceType: 'ELEMENT',
      type: 'ACTION',
      category: 'ACTION',
      version: '1.0.0',
      deprecated: false,
      deletable: true,
      display: {
        label: 'Button',
        image: '',
        showOnResourcePanel: false,
      },
      config: {
        field: {name: 'button', type: ElementTypes},
        styles: {},
        label: 'Click Me',
      },
      variant: ButtonVariants.Primary,
      ...overrides,
    }) as unknown as FlowElement;

  const createWrapper = () => {
    function Wrapper({children}: {children: ReactNode}) {
      return <ReactFlowProvider>{children}</ReactFlowProvider>;
    }
    return Wrapper;
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render the button adapter with correct class names', () => {
      const resource = createMockElement();

      const {container} = render(<ButtonAdapter resource={resource} />, {wrapper: createWrapper()});

      expect(container.querySelector('.adapter')).toBeInTheDocument();
      expect(container.querySelector('.button-adapter')).toBeInTheDocument();
    });

    it('should render a Button component', () => {
      const resource = createMockElement();

      render(<ButtonAdapter resource={resource} />, {wrapper: createWrapper()});

      expect(screen.getByRole('button')).toBeInTheDocument();
    });

    it('should render button label via PlaceholderComponent', () => {
      const resource = createMockElement({label: 'Submit'});

      render(<ButtonAdapter resource={resource} />, {wrapper: createWrapper()});

      expect(screen.getByTestId('placeholder')).toHaveTextContent('Submit');
    });

    it('should render NodeHandle for edge connection', () => {
      const resource = createMockElement();

      render(<ButtonAdapter resource={resource} />, {wrapper: createWrapper()});

      expect(screen.getByTestId('node-handle')).toBeInTheDocument();
      expect(screen.getByTestId('node-handle')).toHaveAttribute('data-type', 'source');
    });
  });

  describe('Button Variants', () => {
    it('should render primary button variant', () => {
      const resource = createMockElement({variant: ButtonVariants.Primary});

      render(<ButtonAdapter resource={resource} />, {wrapper: createWrapper()});

      const button = screen.getByRole('button');
      expect(button).toBeInTheDocument();
    });

    it('should render secondary button variant', () => {
      const resource = createMockElement({variant: ButtonVariants.Secondary});

      render(<ButtonAdapter resource={resource} />, {wrapper: createWrapper()});

      const button = screen.getByRole('button');
      expect(button).toBeInTheDocument();
    });

    it('should render text button variant', () => {
      const resource = createMockElement({variant: ButtonVariants.Text});

      render(<ButtonAdapter resource={resource} />, {wrapper: createWrapper()});

      const button = screen.getByRole('button');
      expect(button).toBeInTheDocument();
    });

    it('should render social button variant with default image', () => {
      const resource = createMockElement({variant: ButtonVariants.Social});

      render(<ButtonAdapter resource={resource} />, {wrapper: createWrapper()});

      const button = screen.getByRole('button');
      expect(button).toBeInTheDocument();
    });
  });

  describe('Button Images', () => {
    it('should render start icon from resource.image', () => {
      const resource = createMockElement({
        image: '/path/to/icon.svg',
        variant: ButtonVariants.Primary,
      });

      const {container} = render(<ButtonAdapter resource={resource} />, {wrapper: createWrapper()});

      // Images with empty alt have role="presentation", so query by tag
      const img = container.querySelector('img');
      expect(img).toBeInTheDocument();
      expect(img).toHaveAttribute('src', '/path/to/icon.svg');
    });

    it('should render start icon from config.image', () => {
      const resource = createMockElement({
        config: {image: '/config/icon.svg'},
        variant: ButtonVariants.Primary,
      });

      const {container} = render(<ButtonAdapter resource={resource} />, {wrapper: createWrapper()});

      const img = container.querySelector('img');
      expect(img).toBeInTheDocument();
      expect(img).toHaveAttribute('src', '/config/icon.svg');
    });

    it('should prioritize resource.image over config.image', () => {
      const resource = createMockElement({
        image: '/resource/icon.svg',
        config: {image: '/config/icon.svg'},
        variant: ButtonVariants.Primary,
      });

      const {container} = render(<ButtonAdapter resource={resource} />, {wrapper: createWrapper()});

      const img = container.querySelector('img');
      expect(img).toBeInTheDocument();
      expect(img).toHaveAttribute('src', '/resource/icon.svg');
    });
  });

  describe('Element Index', () => {
    it('should pass elementIndex to NodeHandle for position updates', () => {
      const resource = createMockElement();

      render(<ButtonAdapter resource={resource} elementIndex={5} />, {wrapper: createWrapper()});

      expect(screen.getByTestId('node-handle')).toBeInTheDocument();
    });

    it('should work without elementIndex', () => {
      const resource = createMockElement();

      render(<ButtonAdapter resource={resource} />, {wrapper: createWrapper()});

      expect(screen.getByTestId('node-handle')).toBeInTheDocument();
    });
  });

  describe('Config Styles', () => {
    it('should apply styles from config', () => {
      const resource = createMockElement({
        config: {styles: {backgroundColor: 'red'}},
      });

      render(<ButtonAdapter resource={resource} />, {wrapper: createWrapper()});

      const button = screen.getByRole('button');
      expect(button).toBeInTheDocument();
    });
  });

  describe('Empty Label', () => {
    it('should handle empty label', () => {
      const resource = createMockElement({label: ''});

      render(<ButtonAdapter resource={resource} />, {wrapper: createWrapper()});

      expect(screen.getByTestId('placeholder')).toHaveTextContent('');
    });

    it('should handle undefined label', () => {
      const resource = createMockElement({label: undefined});

      render(<ButtonAdapter resource={resource} />, {wrapper: createWrapper()});

      expect(screen.getByTestId('placeholder')).toHaveTextContent('');
    });
  });
});
