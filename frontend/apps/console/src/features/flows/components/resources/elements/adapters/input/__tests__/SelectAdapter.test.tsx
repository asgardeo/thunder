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

import {render, screen} from '@testing-library/react';
import type {ReactNode} from 'react';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import SelectAdapter from '../SelectAdapter';
import type {Element as FlowElement} from '@/features/flows/models/elements';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
  Trans: ({children}: {children: ReactNode}) => children,
}));

vi.mock('@thunderid/hooks', () => ({
  useTemplateLiteralResolver: () => ({
    resolve: (value: string | undefined) => value,
  }),
}));

vi.mock('@/features/flows/components/resources/elements/hint', () => ({
  Hint: ({hint}: {hint: string}) => <span data-testid="hint">{hint}</span>,
}));

describe('SelectAdapter', () => {
  const createMockElement = (overrides: Partial<FlowElement> & Record<string, unknown> = {}): FlowElement =>
    ({
      id: 'select-1',
      type: 'SELECT',
      category: 'FIELD',
      config: {},
      label: 'Country',
      placeholder: 'Select a country',
      ...overrides,
    }) as FlowElement;

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render the select component', () => {
      const resource = createMockElement();

      const {container} = render(<SelectAdapter resource={resource} />);

      expect(container.querySelector('.MuiFormControl-root')).toBeInTheDocument();
    });

    it('should render the label', () => {
      const resource = createMockElement({label: 'Country'});

      render(<SelectAdapter resource={resource} />);

      expect(screen.getByText('Country')).toBeInTheDocument();
    });

    it('should render the placeholder as a disabled menu item', () => {
      const resource = createMockElement({placeholder: 'Choose one'});

      render(<SelectAdapter resource={resource} />);

      expect(screen.getByText('Choose one')).toBeInTheDocument();
    });

    it('should render required indicator when required is true', () => {
      const resource = createMockElement({required: true});

      const {container} = render(<SelectAdapter resource={resource} />);

      // MUI adds an asterisk span with class MuiFormLabel-asterisk for required fields
      const asterisk = container.querySelector('.MuiFormLabel-asterisk');
      expect(asterisk).toBeInTheDocument();
    });
  });

  describe('Hint', () => {
    it('should render hint when provided', () => {
      const resource = createMockElement({hint: 'Select your country of residence'});

      render(<SelectAdapter resource={resource} />);

      expect(screen.getByTestId('hint')).toHaveTextContent('Select your country of residence');
    });

    it('should not render hint when not provided', () => {
      const resource = createMockElement({hint: undefined});

      render(<SelectAdapter resource={resource} />);

      expect(screen.queryByTestId('hint')).not.toBeInTheDocument();
    });
  });

  describe('Default Values', () => {
    it('should handle empty label', () => {
      const resource = createMockElement({label: undefined});

      const {container} = render(<SelectAdapter resource={resource} />);

      expect(container.querySelector('.MuiFormControl-root')).toBeInTheDocument();
    });

    it('should handle empty placeholder', () => {
      const resource = createMockElement({placeholder: undefined});

      const {container} = render(<SelectAdapter resource={resource} />);

      expect(container.querySelector('.MuiFormControl-root')).toBeInTheDocument();
    });
  });
});
