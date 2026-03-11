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
import OTPInputAdapter from '../OTPInputAdapter';

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

describe('OTPInputAdapter', () => {
  const createMockElement = (overrides: Partial<FlowElement> & Record<string, unknown> = {}): FlowElement =>
    ({
      id: 'otp-1',
      type: 'OTP_INPUT',
      category: 'FIELD',
      config: {},
      label: 'Enter OTP',
      inputType: 'text',
      ...overrides,
    }) as FlowElement;

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render InputLabel component', async () => {
      const resource = createMockElement();

      const {container} = await render(<OTPInputAdapter resource={resource} />);

      expect(container.querySelector('.MuiInputLabel-root')).toBeInTheDocument();
    });

    it('should render label via PlaceholderComponent', async () => {
      const resource = createMockElement({label: 'Verification Code'});

      await render(<OTPInputAdapter resource={resource} />);

      await expect.element(page.getByTestId('placeholder')).toHaveTextContent('Verification Code');
    });

    it('should render 6 input boxes for OTP', async () => {
      const resource = createMockElement();

      const {container} = await render(<OTPInputAdapter resource={resource} />);

      const inputs = container.querySelectorAll('.MuiOutlinedInput-root');
      expect(inputs).toHaveLength(6);
    });
  });

  describe('Required Field', () => {
    it('should show required indicator when required is true', async () => {
      const resource = createMockElement({required: true});

      const {container} = await render(<OTPInputAdapter resource={resource} />);

      expect(container.querySelector('.MuiFormLabel-asterisk')).toBeInTheDocument();
    });

    it('should not show required indicator when required is false', async () => {
      const resource = createMockElement({required: false});

      const {container} = await render(<OTPInputAdapter resource={resource} />);

      expect(container.querySelector('.MuiFormLabel-asterisk')).not.toBeInTheDocument();
    });
  });

  describe('Hint Text', () => {
    it('should render hint when provided', async () => {
      const resource = createMockElement({hint: 'Check your email for the code'});

      await render(<OTPInputAdapter resource={resource} />);

      await expect.element(page.getByTestId('hint')).toHaveTextContent('Check your email for the code');
    });

    it('should not render hint when not provided', async () => {
      const resource = createMockElement({hint: undefined});

      await render(<OTPInputAdapter resource={resource} />);

      await expect.element(page.getByTestId('hint')).not.toBeInTheDocument();
    });

    it('should not render hint when empty', async () => {
      const resource = createMockElement({hint: ''});

      await render(<OTPInputAdapter resource={resource} />);

      await expect.element(page.getByTestId('hint')).not.toBeInTheDocument();
    });
  });

  describe('Placeholder', () => {
    it('should render placeholder on OTP inputs when provided', async () => {
      const resource = createMockElement({placeholder: '0'});

      const {container} = await render(<OTPInputAdapter resource={resource} />);

      const inputs = container.querySelectorAll('input');
      inputs.forEach((input) => {
        expect(input).toHaveAttribute('placeholder', '0');
      });
    });

    it('should render empty placeholder when not provided', async () => {
      const resource = createMockElement({placeholder: undefined});

      const {container} = await render(<OTPInputAdapter resource={resource} />);

      const inputs = container.querySelectorAll('input');
      inputs.forEach((input) => {
        expect(input).toHaveAttribute('placeholder', '');
      });
    });
  });

  describe('Input Type', () => {
    it('should apply input type to OTP fields', async () => {
      const resource = createMockElement({inputType: 'number'});

      const {container} = await render(<OTPInputAdapter resource={resource} />);

      const inputs = container.querySelectorAll('input');
      inputs.forEach((input) => {
        expect(input).toHaveAttribute('type', 'number');
      });
    });
  });

  describe('Custom Styling', () => {
    it('should apply className when provided', async () => {
      const resource = createMockElement({className: 'custom-otp'});

      const {container} = await render(<OTPInputAdapter resource={resource} />);

      expect(container.firstChild).toHaveClass('custom-otp');
    });

    it('should apply styles to inputs when provided', async () => {
      const resource = createMockElement({styles: {width: '40px'}});

      const {container} = await render(<OTPInputAdapter resource={resource} />);

      const outlinedInputs = container.querySelectorAll('.MuiOutlinedInput-root');
      outlinedInputs.forEach((input) => {
        expect(input).toHaveStyle({width: '40px'});
      });
    });
  });

  describe('Empty Label', () => {
    it('should handle empty label', async () => {
      const resource = createMockElement({label: ''});

      await render(<OTPInputAdapter resource={resource} />);

      await expect.element(page.getByTestId('placeholder')).toHaveTextContent('');
    });

    it('should handle undefined label', async () => {
      const resource = createMockElement({label: undefined});

      await render(<OTPInputAdapter resource={resource} />);

      await expect.element(page.getByTestId('placeholder')).toHaveTextContent('');
    });
  });

  describe('Validation', () => {
    it('should call useRequiredFields with resource', async () => {
      const useRequiredFields = await import('@/features/flows/hooks/useRequiredFields');
      const mockUseRequiredFields = vi.mocked(useRequiredFields.default);

      const resource = createMockElement();

      await render(<OTPInputAdapter resource={resource} />);

      expect(mockUseRequiredFields).toHaveBeenCalled();
    });
  });
});
