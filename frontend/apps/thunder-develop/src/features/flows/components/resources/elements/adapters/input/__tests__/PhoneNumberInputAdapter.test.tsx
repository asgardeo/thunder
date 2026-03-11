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
import PhoneNumberInputAdapter from '../PhoneNumberInputAdapter';

// Mock dependencies
vi.mock('@/features/flows/hooks/useRequiredFields', () => ({
  default: vi.fn(),
}));

vi.mock('@/features/flows/components/resources/elements/hint', () => ({
  Hint: ({hint}: {hint: string}) => <span data-testid="hint">{hint}</span>,
}));

describe('PhoneNumberInputAdapter', () => {
  const createMockElement = (overrides: Partial<FlowElement> & Record<string, unknown> = {}): FlowElement =>
    ({
      id: 'phone-1',
      type: 'PHONE_INPUT',
      category: 'FIELD',
      config: {},
      label: 'Phone Number',
      ...overrides,
    }) as FlowElement;

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render TextField component', async () => {
      const resource = createMockElement();

      const {container} = await render(<PhoneNumberInputAdapter resource={resource} />);

      expect(container.querySelector('.MuiTextField-root')).toBeInTheDocument();
    });

    it('should render input with type number', async () => {
      const resource = createMockElement();

      await render(<PhoneNumberInputAdapter resource={resource} />);

      await expect.element(page.getByRole('spinbutton')).toHaveAttribute('type', 'number');
    });

    it('should render with label', async () => {
      const resource = createMockElement({label: 'Mobile Number'});

      await render(<PhoneNumberInputAdapter resource={resource} />);

      await expect.element(page.getByLabelText('Mobile Number')).toBeInTheDocument();
    });
  });

  describe('Placeholder', () => {
    it('should render placeholder when provided', async () => {
      const resource = createMockElement({placeholder: '+1 (555) 123-4567'});

      await render(<PhoneNumberInputAdapter resource={resource} />);

      await expect.element(page.getByRole('spinbutton')).toHaveAttribute('placeholder', '+1 (555) 123-4567');
    });

    it('should render empty placeholder when not provided', async () => {
      const resource = createMockElement({placeholder: undefined});

      await render(<PhoneNumberInputAdapter resource={resource} />);

      await expect.element(page.getByRole('spinbutton')).toHaveAttribute('placeholder', '');
    });
  });

  describe('Required Field', () => {
    it('should show required indicator when required is true', async () => {
      const resource = createMockElement({required: true});

      const {container} = await render(<PhoneNumberInputAdapter resource={resource} />);

      expect(container.querySelector('.MuiFormLabel-asterisk')).toBeInTheDocument();
    });

    it('should not show required indicator when required is false', async () => {
      const resource = createMockElement({required: false});

      const {container} = await render(<PhoneNumberInputAdapter resource={resource} />);

      expect(container.querySelector('.MuiFormLabel-asterisk')).not.toBeInTheDocument();
    });
  });

  describe('Hint Text', () => {
    it('should render hint when provided', async () => {
      const resource = createMockElement({hint: 'Include country code'});

      await render(<PhoneNumberInputAdapter resource={resource} />);

      await expect.element(page.getByTestId('hint')).toHaveTextContent('Include country code');
    });

    it('should not render hint when not provided', async () => {
      const resource = createMockElement({hint: undefined});

      await render(<PhoneNumberInputAdapter resource={resource} />);

      await expect.element(page.getByTestId('hint')).not.toBeInTheDocument();
    });

    it('should not render hint when empty', async () => {
      const resource = createMockElement({hint: ''});

      await render(<PhoneNumberInputAdapter resource={resource} />);

      await expect.element(page.getByTestId('hint')).not.toBeInTheDocument();
    });
  });

  describe('Custom Styling', () => {
    it('should apply className when provided', async () => {
      const resource = createMockElement({className: 'custom-phone'});

      const {container} = await render(<PhoneNumberInputAdapter resource={resource} />);

      expect(container.querySelector('.custom-phone')).toBeInTheDocument();
    });
  });

  describe('Empty Label', () => {
    it('should handle empty label', async () => {
      const resource = createMockElement({label: ''});

      const {container} = await render(<PhoneNumberInputAdapter resource={resource} />);

      expect(container.querySelector('.MuiTextField-root')).toBeInTheDocument();
    });

    it('should handle undefined label', async () => {
      const resource = createMockElement({label: undefined});

      const {container} = await render(<PhoneNumberInputAdapter resource={resource} />);

      expect(container.querySelector('.MuiTextField-root')).toBeInTheDocument();
    });
  });

  describe('Validation', () => {
    it('should call useRequiredFields with resource', async () => {
      const useRequiredFields = await import('@/features/flows/hooks/useRequiredFields');
      const mockUseRequiredFields = vi.mocked(useRequiredFields.default);

      const resource = createMockElement();

      await render(<PhoneNumberInputAdapter resource={resource} />);

      expect(mockUseRequiredFields).toHaveBeenCalled();
    });
  });
});
