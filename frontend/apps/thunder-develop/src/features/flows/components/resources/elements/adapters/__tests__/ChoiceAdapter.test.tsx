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
import type {Element as FlowElement} from '@/features/flows/models/elements';
import type {FieldOption} from '@/features/flows/models/base';
import ChoiceAdapter from '../ChoiceAdapter';

// Mock the Hint component
vi.mock('../../hint', () => ({
  Hint: ({hint}: {hint: string}) => <span data-testid="hint">{hint}</span>,
}));

describe('ChoiceAdapter', () => {
  const createMockElement = (overrides: Partial<FlowElement> & Record<string, unknown> = {}): FlowElement =>
    ({
      id: 'choice-1',
      type: 'DROPDOWN',
      category: 'FIELD',
      config: {},
      label: 'Select an option',
      ...overrides,
    }) as FlowElement;

  const createMockOptions = (): FieldOption[] => [
    {key: 'opt1', value: 'option1', label: 'Option 1'},
    {key: 'opt2', value: 'option2', label: 'Option 2'},
    {key: 'opt3', value: 'option3', label: 'Option 3'},
  ];

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render FormControl container', async () => {
      const resource = createMockElement();

      const {container} = await render(<ChoiceAdapter resource={resource} />);

      expect(container.querySelector('.MuiFormControl-root')).toBeInTheDocument();
    });

    it('should render form label with resource label', async () => {
      const resource = createMockElement({label: 'Choose your preference'});

      await render(<ChoiceAdapter resource={resource} />);

      await expect.element(page.getByText('Choose your preference')).toBeInTheDocument();
    });

    it('should render RadioGroup component', async () => {
      const resource = createMockElement();

      const {container} = await render(<ChoiceAdapter resource={resource} />);

      expect(container.querySelector('.MuiRadioGroup-root')).toBeInTheDocument();
    });
  });

  describe('Options Rendering', () => {
    it('should render radio options for each field option', async () => {
      const options = createMockOptions();
      const resource = createMockElement({options});

      await render(<ChoiceAdapter resource={resource} />);

      await expect.element(page.getByText('Option 1')).toBeInTheDocument();
      await expect.element(page.getByText('Option 2')).toBeInTheDocument();
      await expect.element(page.getByText('Option 3')).toBeInTheDocument();
    });

    it('should render radio buttons for each option', async () => {
      const options = createMockOptions();
      const resource = createMockElement({options});

      await render(<ChoiceAdapter resource={resource} />);

      const radioButtons = page.getByRole('radio');
      expect(radioButtons).toHaveLength(3);
    });

    it('should handle empty options array', async () => {
      const resource = createMockElement({options: []});

      const {container} = await render(<ChoiceAdapter resource={resource} />);

      const radioButtons = container.querySelectorAll('input[type="radio"]');
      expect(radioButtons).toHaveLength(0);
    });

    it('should handle undefined options', async () => {
      const resource = createMockElement({options: undefined});

      const {container} = await render(<ChoiceAdapter resource={resource} />);

      const radioButtons = container.querySelectorAll('input[type="radio"]');
      expect(radioButtons).toHaveLength(0);
    });
  });

  describe('Default Value', () => {
    it('should set default value on RadioGroup', async () => {
      const options = createMockOptions();
      const resource = createMockElement({
        options,
        defaultValue: 'option2',
      });

      await render(<ChoiceAdapter resource={resource} />);

      const radioButtons = page.getByRole('radio').all();
      expect(radioButtons[1]).toBeChecked();
    });

    it('should handle no default value', async () => {
      const options = createMockOptions();
      const resource = createMockElement({
        options,
        defaultValue: undefined,
      });

      await render(<ChoiceAdapter resource={resource} />);

      const radioButtons = page.getByRole('radio').all();
      radioButtons.forEach((radio) => {
        expect(radio).not.toBeChecked();
      });
    });
  });

  describe('Hint Text', () => {
    it('should render hint when provided', async () => {
      const resource = createMockElement({hint: 'Select one of the options above'});

      await render(<ChoiceAdapter resource={resource} />);

      await expect.element(page.getByTestId('hint')).toBeInTheDocument();
      await expect.element(page.getByTestId('hint')).toHaveTextContent('Select one of the options above');
    });

    it('should not render hint when not provided', async () => {
      const resource = createMockElement({hint: undefined});

      await render(<ChoiceAdapter resource={resource} />);

      await expect.element(page.getByTestId('hint')).not.toBeInTheDocument();
    });

    it('should not render hint when empty string', async () => {
      const resource = createMockElement({hint: ''});

      await render(<ChoiceAdapter resource={resource} />);

      await expect.element(page.getByTestId('hint')).not.toBeInTheDocument();
    });
  });

  describe('Form Label', () => {
    it('should render with resource id as label id', async () => {
      const resource = createMockElement({id: 'my-choice-field'});

      const {container} = await render(<ChoiceAdapter resource={resource} />);

      const label = container.querySelector('.MuiFormLabel-root');
      expect(label).toHaveAttribute('id', 'my-choice-field');
    });

    it('should handle empty label', async () => {
      const resource = createMockElement({label: ''});

      const {container} = await render(<ChoiceAdapter resource={resource} />);

      const label = container.querySelector('.MuiFormLabel-root');
      expect(label).toBeInTheDocument();
    });

    it('should handle undefined label', async () => {
      const resource = createMockElement({label: undefined});

      const {container} = await render(<ChoiceAdapter resource={resource} />);

      const label = container.querySelector('.MuiFormLabel-root');
      expect(label).toBeInTheDocument();
    });
  });
});
