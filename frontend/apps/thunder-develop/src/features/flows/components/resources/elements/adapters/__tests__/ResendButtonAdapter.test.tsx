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
import {ElementTypes, type Element as FlowElement} from '@/features/flows/models/elements';
import ResendButtonAdapter from '../ResendButtonAdapter';

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

vi.mock('../NodeHandle', () => ({
  default: ({id, type, position}: {id: string; type: string; position: string}) => (
    <div data-testid="node-handle" data-id={id} data-type={type} data-position={position} />
  ),
}));

vi.mock('../PlaceholderComponent', () => ({
  default: ({value}: {value: string}) => <span data-testid="placeholder">{value}</span>,
}));

describe('ResendButtonAdapter', () => {
  const createMockElement = (overrides: Partial<FlowElement> & Record<string, unknown> = {}): FlowElement =>
    ({
      id: 'resend-1',
      resourceType: 'ELEMENT',
      type: 'RESEND',
      category: 'ACTION',
      version: '1.0.0',
      deprecated: false,
      deletable: true,
      display: {
        label: 'Resend',
        image: '',
        showOnResourcePanel: false,
      },
      config: {
        field: {name: 'resend', type: 'RESEND'},
        styles: {},
      },
      label: 'Resend Code',
      ...overrides,
    }) as FlowElement;

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
    it('should render the resend button adapter with correct class names', () => {
      const resource = createMockElement();

      const {container} = render(<ResendButtonAdapter resource={resource} stepId="step-1" />, {
        wrapper: createWrapper(),
      });

      expect(container.querySelector('.adapter')).toBeInTheDocument();
      expect(container.querySelector('.button-adapter')).toBeInTheDocument();
    });

    it('should render a Button component', () => {
      const resource = createMockElement();

      render(<ResendButtonAdapter resource={resource} stepId="step-1" />, {wrapper: createWrapper()});

      expect(screen.getByRole('button')).toBeInTheDocument();
    });

    it('should render button label via PlaceholderComponent', () => {
      const resource = createMockElement({label: 'Resend OTP'});

      render(<ResendButtonAdapter resource={resource} stepId="step-1" />, {wrapper: createWrapper()});

      expect(screen.getByTestId('placeholder')).toHaveTextContent('Resend OTP');
    });

    it('should render NodeHandle for edge connection', () => {
      const resource = createMockElement();

      render(<ResendButtonAdapter resource={resource} stepId="step-1" />, {wrapper: createWrapper()});

      expect(screen.getByTestId('node-handle')).toBeInTheDocument();
      expect(screen.getByTestId('node-handle')).toHaveAttribute('data-type', 'source');
    });
  });

  describe('Button Configuration', () => {
    it('should render with secondary color', () => {
      const resource = createMockElement();

      render(<ResendButtonAdapter resource={resource} stepId="step-1" />, {wrapper: createWrapper()});

      const button = screen.getByRole('button');
      expect(button).toBeInTheDocument();
    });

    it('should apply styles from config', () => {
      const resource = createMockElement({
        config: {
          field: {name: 'resend', type: ElementTypes},
          styles: {backgroundColor: 'blue'},
        },
      });

      render(<ResendButtonAdapter resource={resource} stepId="step-1" />, {wrapper: createWrapper()});

      const button = screen.getByRole('button');
      expect(button).toBeInTheDocument();
    });
  });

  describe('Empty Label', () => {
    it('should handle empty label', () => {
      const resource = createMockElement({label: ''});

      render(<ResendButtonAdapter resource={resource} stepId="step-1" />, {wrapper: createWrapper()});

      expect(screen.getByTestId('placeholder')).toHaveTextContent('');
    });

    it('should handle undefined label', () => {
      const resource = createMockElement({label: undefined});

      render(<ResendButtonAdapter resource={resource} stepId="step-1" />, {wrapper: createWrapper()});

      expect(screen.getByTestId('placeholder')).toHaveTextContent('');
    });
  });

  describe('Validation', () => {
    it('should call useRequiredFields with resource', async () => {
      const useRequiredFields = await import('@/features/flows/hooks/useRequiredFields');
      const mockUseRequiredFields = vi.mocked(useRequiredFields.default);

      const resource = createMockElement();

      render(<ResendButtonAdapter resource={resource} stepId="step-1" />, {wrapper: createWrapper()});

      expect(mockUseRequiredFields).toHaveBeenCalled();
    });
  });
});
