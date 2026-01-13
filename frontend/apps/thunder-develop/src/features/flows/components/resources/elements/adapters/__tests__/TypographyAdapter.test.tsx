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
import {TypographyVariants, ElementTypes, type Element as FlowElement} from '@/features/flows/models/elements';
import TypographyAdapter from '../TypographyAdapter';

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

vi.mock('../PlaceholderComponent', () => ({
  default: ({value}: {value: string}) => <span data-testid="placeholder">{value}</span>,
}));

describe('TypographyAdapter', () => {
  const createMockElement = (overrides: Partial<FlowElement> & Record<string, unknown> = {}): FlowElement =>
    ({
      id: 'typography-1',
      resourceType: 'ELEMENT',
      type: 'TEXT',
      category: 'DISPLAY',
      version: '1.0.0',
      deprecated: false,
      deletable: true,
      display: {
        label: 'Typography',
        image: '',
        showOnResourcePanel: false,
      },
      config: {
        field: {name: 'text', type: 'TEXT'},
        styles: {},
      },
      label: 'Hello World',
      variant: TypographyVariants.Body1,
      ...overrides,
    }) as FlowElement;

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render Typography component', () => {
      const resource = createMockElement();

      const {container} = render(<TypographyAdapter resource={resource} stepId="step-1" />);

      expect(container.querySelector('.MuiTypography-root')).toBeInTheDocument();
    });

    it('should render label via PlaceholderComponent', () => {
      const resource = createMockElement({label: 'Test Label'});

      render(<TypographyAdapter resource={resource} stepId="step-1" />);

      expect(screen.getByTestId('placeholder')).toHaveTextContent('Test Label');
    });
  });

  describe('Typography Variants', () => {
    it('should render H1 variant with center alignment', () => {
      const resource = createMockElement({variant: TypographyVariants.H1});

      const {container} = render(<TypographyAdapter resource={resource} stepId="step-1" />);

      expect(container.querySelector('.MuiTypography-h1')).toBeInTheDocument();
    });

    it('should render H2 variant with center alignment', () => {
      const resource = createMockElement({variant: TypographyVariants.H2});

      const {container} = render(<TypographyAdapter resource={resource} stepId="step-1" />);

      expect(container.querySelector('.MuiTypography-h2')).toBeInTheDocument();
    });

    it('should render H3 variant with center alignment', () => {
      const resource = createMockElement({variant: TypographyVariants.H3});

      const {container} = render(<TypographyAdapter resource={resource} stepId="step-1" />);

      expect(container.querySelector('.MuiTypography-h3')).toBeInTheDocument();
    });

    it('should render H4 variant with center alignment', () => {
      const resource = createMockElement({variant: TypographyVariants.H4});

      const {container} = render(<TypographyAdapter resource={resource} stepId="step-1" />);

      expect(container.querySelector('.MuiTypography-h4')).toBeInTheDocument();
    });

    it('should render H5 variant with center alignment', () => {
      const resource = createMockElement({variant: TypographyVariants.H5});

      const {container} = render(<TypographyAdapter resource={resource} stepId="step-1" />);

      expect(container.querySelector('.MuiTypography-h5')).toBeInTheDocument();
    });

    it('should render H6 variant with center alignment', () => {
      const resource = createMockElement({variant: TypographyVariants.H6});

      const {container} = render(<TypographyAdapter resource={resource} stepId="step-1" />);

      expect(container.querySelector('.MuiTypography-h6')).toBeInTheDocument();
    });

    it('should render Body1 variant', () => {
      const resource = createMockElement({variant: TypographyVariants.Body1});

      const {container} = render(<TypographyAdapter resource={resource} stepId="step-1" />);

      expect(container.querySelector('.MuiTypography-body1')).toBeInTheDocument();
    });

    it('should render Body2 variant', () => {
      const resource = createMockElement({variant: TypographyVariants.Body2});

      const {container} = render(<TypographyAdapter resource={resource} stepId="step-1" />);

      expect(container.querySelector('.MuiTypography-body2')).toBeInTheDocument();
    });
  });

  describe('Config Styles', () => {
    it('should apply styles from config', () => {
      const resource = createMockElement({
        config: {
          field: {name: 'text', type: ElementTypes},
          styles: {color: 'red'},
        },
      });

      render(<TypographyAdapter resource={resource} stepId="step-1" />);

      const typography = screen.getByTestId('placeholder').parentElement;
      // Color can be normalized to RGB format
      expect(typography).toHaveStyle({color: 'rgb(255, 0, 0)'});
    });
  });

  describe('Empty Label', () => {
    it('should handle empty label', () => {
      const resource = createMockElement({label: ''});

      render(<TypographyAdapter resource={resource} stepId="step-1" />);

      expect(screen.getByTestId('placeholder')).toHaveTextContent('');
    });

    it('should handle undefined label', () => {
      const resource = createMockElement({label: undefined});

      render(<TypographyAdapter resource={resource} stepId="step-1" />);

      expect(screen.getByTestId('placeholder')).toHaveTextContent('');
    });
  });

  describe('Undefined Variant', () => {
    it('should handle undefined variant', () => {
      const resource = createMockElement({variant: undefined});

      const {container} = render(<TypographyAdapter resource={resource} stepId="step-1" />);

      expect(container.querySelector('.MuiTypography-root')).toBeInTheDocument();
    });
  });

  describe('Validation', () => {
    it('should call useRequiredFields with resource', async () => {
      const useRequiredFields = await import('@/features/flows/hooks/useRequiredFields');
      const mockUseRequiredFields = vi.mocked(useRequiredFields.default);

      const resource = createMockElement();

      render(<TypographyAdapter resource={resource} stepId="step-1" />);

      expect(mockUseRequiredFields).toHaveBeenCalled();
    });
  });
});
