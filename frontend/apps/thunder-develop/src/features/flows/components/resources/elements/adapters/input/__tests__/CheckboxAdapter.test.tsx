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
import {render} from '@thunder/test-utils/browser';
import {page} from 'vitest/browser';
import type {ReactNode} from 'react';
import type {Element as FlowElement} from '@/features/flows/models/elements';
import CheckboxAdapter from '../CheckboxAdapter';

// Mock dependencies
vi.mock('@/features/flows/hooks/useRequiredFields', () => ({
  default: vi.fn(),
}));

vi.mock('@/features/flows/components/resources/elements/hint', () => ({
  Hint: ({hint}: {hint: string}) => <span data-testid="hint">{hint}</span>,
}));

vi.mock('@/features/flows/components/resources/elements/adapters/PlaceholderComponent', () => ({
  default: ({value}: {value: string}) => <span data-testid="placeholder">{value}</span>,
}));

describe('CheckboxAdapter', () => {
  const createMockElement = (overrides: Partial<FlowElement> & Record<string, unknown> = {}): FlowElement =>
    ({
      id: 'checkbox-1',
      type: 'CHECKBOX',
      category: 'FIELD',
      config: {},
      label: 'Accept terms and conditions',
      ...overrides,
    }) as FlowElement;

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render FormControlLabel component', async () => {
      const resource = createMockElement();

      const {container} = await render(<CheckboxAdapter resource={resource} />);

      expect(container.querySelector('.MuiFormControlLabel-root')).toBeInTheDocument();
    });

    it('should render Checkbox component', async () => {
      const resource = createMockElement();

      await render(<CheckboxAdapter resource={resource} />);

      await expect.element(page.getByRole('checkbox')).toBeInTheDocument();
    });

    it('should render label via PlaceholderComponent', async () => {
      const resource = createMockElement({label: 'I agree'});

      await render(<CheckboxAdapter resource={resource} />);

      await expect.element(page.getByTestId('placeholder')).toHaveTextContent('I agree');
    });

    it('should render checkbox as checked by default', async () => {
      const resource = createMockElement();

      await render(<CheckboxAdapter resource={resource} />);

      expect(page.getByRole('checkbox')).toBeChecked();
    });
  });

  describe('Required Field', () => {
    it('should show required indicator when required is true', async () => {
      const resource = createMockElement({required: true});

      const {container} = await render(<CheckboxAdapter resource={resource} />);

      expect(container.querySelector('.MuiFormControlLabel-asterisk')).toBeInTheDocument();
    });

    it('should not show required indicator when required is false', async () => {
      const resource = createMockElement({required: false});

      const {container} = await render(<CheckboxAdapter resource={resource} />);

      expect(container.querySelector('.MuiFormControlLabel-asterisk')).not.toBeInTheDocument();
    });
  });

  describe('Hint Text', () => {
    it('should render hint when provided', async () => {
      const resource = createMockElement({hint: 'Please read the terms'});

      await render(<CheckboxAdapter resource={resource} />);

      await expect.element(page.getByTestId('hint')).toHaveTextContent('Please read the terms');
    });

    it('should not render hint when not provided', async () => {
      const resource = createMockElement({hint: undefined});

      await render(<CheckboxAdapter resource={resource} />);

      await expect.element(page.getByTestId('hint')).not.toBeInTheDocument();
    });

    it('should not render hint when empty', async () => {
      const resource = createMockElement({hint: ''});

      await render(<CheckboxAdapter resource={resource} />);

      await expect.element(page.getByTestId('hint')).not.toBeInTheDocument();
    });
  });

  describe('Custom Styling', () => {
    it('should apply className when provided', async () => {
      const resource = createMockElement({className: 'custom-checkbox'});

      const {container} = await render(<CheckboxAdapter resource={resource} />);

      expect(container.querySelector('.custom-checkbox')).toBeInTheDocument();
    });

    it('should apply styles when provided', async () => {
      const resource = createMockElement({styles: {marginTop: '10px'}});

      const {container} = await render(<CheckboxAdapter resource={resource} />);

      const label = container.querySelector('.MuiFormControlLabel-root');
      expect(label).toHaveStyle({marginTop: '10px'});
    });
  });

  describe('Empty Label', () => {
    it('should handle empty label', async () => {
      const resource = createMockElement({label: ''});

      await render(<CheckboxAdapter resource={resource} />);

      await expect.element(page.getByTestId('placeholder')).toHaveTextContent('');
    });

    it('should handle undefined label', async () => {
      const resource = createMockElement({label: undefined});

      await render(<CheckboxAdapter resource={resource} />);

      await expect.element(page.getByTestId('placeholder')).toHaveTextContent('');
    });
  });

  describe('Validation', () => {
    it('should call useRequiredFields with resource', async () => {
      const useRequiredFields = await import('@/features/flows/hooks/useRequiredFields');
      const mockUseRequiredFields = vi.mocked(useRequiredFields.default);

      const resource = createMockElement();

      await render(<CheckboxAdapter resource={resource} />);

      expect(mockUseRequiredFields).toHaveBeenCalled();
    });
  });
});
