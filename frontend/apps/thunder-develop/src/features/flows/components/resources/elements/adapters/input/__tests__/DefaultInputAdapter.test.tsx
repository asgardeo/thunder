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
import DefaultInputAdapter from '../DefaultInputAdapter';

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

describe('DefaultInputAdapter', () => {
  const createMockElement = (overrides: Partial<FlowElement> & Record<string, unknown> = {}): FlowElement =>
    ({
      id: 'input-1',
      type: 'TEXT_INPUT',
      category: 'FIELD',
      config: {},
      label: 'Username',
      inputType: 'text',
      ...overrides,
    }) as FlowElement;

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render TextField component', async () => {
      const resource = createMockElement();

      const {container} = await render(<DefaultInputAdapter resource={resource} />);

      expect(container.querySelector('.MuiTextField-root')).toBeInTheDocument();
    });

    it('should render input element', async () => {
      const resource = createMockElement();

      await render(<DefaultInputAdapter resource={resource} />);

      await expect.element(page.getByRole('textbox')).toBeInTheDocument();
    });

    it('should render label via PlaceholderComponent', async () => {
      const resource = createMockElement({label: 'Email Address'});

      await render(<DefaultInputAdapter resource={resource} />);

      // MUI TextField renders label in two places (label and legend), so we use getAllByTestId
      const placeholders = page.getByTestId('placeholder').all();
      expect(placeholders[0]).toHaveTextContent('Email Address');
    });

    it('should render with full width', async () => {
      const resource = createMockElement();

      const {container} = await render(<DefaultInputAdapter resource={resource} />);

      expect(container.querySelector('.MuiFormControl-fullWidth')).toBeInTheDocument();
    });
  });

  describe('Input Types', () => {
    it('should render text input type', async () => {
      const resource = createMockElement({inputType: 'text'});

      await render(<DefaultInputAdapter resource={resource} />);

      await expect.element(page.getByRole('textbox')).toHaveAttribute('type', 'text');
    });

    it('should render email input type', async () => {
      const resource = createMockElement({inputType: 'email'});

      await render(<DefaultInputAdapter resource={resource} />);

      await expect.element(page.getByRole('textbox')).toHaveAttribute('type', 'email');
    });

    it('should render password input type with autocomplete off', async () => {
      const resource = createMockElement({inputType: 'password'});

      const {container} = await render(<DefaultInputAdapter resource={resource} />);

      const input = container.querySelector('input');
      expect(input).toHaveAttribute('type', 'password');
      expect(input).toHaveAttribute('autocomplete', 'new-password');
    });
  });

  describe('Placeholder', () => {
    it('should render placeholder when provided', async () => {
      const resource = createMockElement({placeholder: 'Enter your username'});

      await render(<DefaultInputAdapter resource={resource} />);

      await expect.element(page.getByRole('textbox')).toHaveAttribute('placeholder', 'Enter your username');
    });

    it('should render empty placeholder when not provided', async () => {
      const resource = createMockElement({placeholder: undefined});

      await render(<DefaultInputAdapter resource={resource} />);

      await expect.element(page.getByRole('textbox')).toHaveAttribute('placeholder', '');
    });
  });

  describe('Default Value', () => {
    it('should render with default value when provided', async () => {
      const resource = createMockElement({defaultValue: 'default text'});

      await render(<DefaultInputAdapter resource={resource} />);

      await expect.element(page.getByRole('textbox')).toHaveValue('default text');
    });
  });

  describe('Required Field', () => {
    it('should show required indicator when required is true', async () => {
      const resource = createMockElement({required: true});

      const {container} = await render(<DefaultInputAdapter resource={resource} />);

      expect(container.querySelector('.MuiFormLabel-asterisk')).toBeInTheDocument();
    });

    it('should not show required indicator when required is false', async () => {
      const resource = createMockElement({required: false});

      const {container} = await render(<DefaultInputAdapter resource={resource} />);

      expect(container.querySelector('.MuiFormLabel-asterisk')).not.toBeInTheDocument();
    });
  });

  describe('Hint Text', () => {
    it('should render hint when provided', async () => {
      const resource = createMockElement({hint: 'Enter a valid email'});

      await render(<DefaultInputAdapter resource={resource} />);

      await expect.element(page.getByTestId('hint')).toHaveTextContent('Enter a valid email');
    });

    it('should not render hint when not provided', async () => {
      const resource = createMockElement({hint: undefined});

      await render(<DefaultInputAdapter resource={resource} />);

      await expect.element(page.getByTestId('hint')).not.toBeInTheDocument();
    });
  });

  describe('Input Constraints', () => {
    it('should set minLength when provided', async () => {
      const resource = createMockElement({minLength: 5});

      await render(<DefaultInputAdapter resource={resource} />);

      await expect.element(page.getByRole('textbox')).toHaveAttribute('minlength', '5');
    });

    it('should set maxLength when provided', async () => {
      const resource = createMockElement({maxLength: 100});

      await render(<DefaultInputAdapter resource={resource} />);

      await expect.element(page.getByRole('textbox')).toHaveAttribute('maxlength', '100');
    });
  });

  describe('Multiline', () => {
    it('should render as multiline when multiline is true', async () => {
      const resource = createMockElement({multiline: true});

      const {container} = await render(<DefaultInputAdapter resource={resource} />);

      expect(container.querySelector('textarea')).toBeInTheDocument();
    });

    it('should render as single line when multiline is false', async () => {
      const resource = createMockElement({multiline: false});

      const {container} = await render(<DefaultInputAdapter resource={resource} />);

      expect(container.querySelector('textarea')).not.toBeInTheDocument();
    });
  });

  describe('Custom Styling', () => {
    it('should apply className when provided', async () => {
      const resource = createMockElement({className: 'custom-input'});

      const {container} = await render(<DefaultInputAdapter resource={resource} />);

      expect(container.querySelector('.custom-input')).toBeInTheDocument();
    });

    it('should apply styles when provided', async () => {
      const resource = createMockElement({styles: {marginTop: '10px'}});

      const {container} = await render(<DefaultInputAdapter resource={resource} />);

      const textField = container.querySelector('.MuiTextField-root');
      expect(textField).toHaveStyle({marginTop: '10px'});
    });
  });

  describe('Empty Label', () => {
    it('should handle empty label', async () => {
      const resource = createMockElement({label: ''});

      await render(<DefaultInputAdapter resource={resource} />);

      // MUI TextField renders label in two places (label and legend), so we use getAllByTestId
      const placeholders = page.getByTestId('placeholder').all();
      expect(placeholders[0]).toHaveTextContent('');
    });

    it('should handle undefined label', async () => {
      const resource = createMockElement({label: undefined});

      await render(<DefaultInputAdapter resource={resource} />);

      // MUI TextField renders label in two places (label and legend), so we use getAllByTestId
      const placeholders = page.getByTestId('placeholder').all();
      expect(placeholders[0]).toHaveTextContent('');
    });
  });

  describe('Validation', () => {
    it('should call useRequiredFields with resource', async () => {
      const useRequiredFields = await import('@/features/flows/hooks/useRequiredFields');
      const mockUseRequiredFields = vi.mocked(useRequiredFields.default);

      const resource = createMockElement();

      await render(<DefaultInputAdapter resource={resource} />);

      expect(mockUseRequiredFields).toHaveBeenCalled();
    });
  });
});
