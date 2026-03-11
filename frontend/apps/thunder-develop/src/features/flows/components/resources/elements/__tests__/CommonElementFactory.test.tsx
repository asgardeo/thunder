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

import {describe, it, expect, vi} from 'vitest';
import {render} from '@thunder/test-utils/browser';
import {page} from 'vitest/browser';
import CommonElementFactory from '../CommonElementFactory';
import {BlockTypes, ElementCategories, ElementTypes, type Element} from '../../../../models/elements';

// Mock all adapter components
vi.mock('../adapters/FormAdapter', () => ({
  default: ({stepId, resource}: {stepId: string; resource: Element}) => (
    <div data-testid="form-adapter" data-step-id={stepId} data-resource-id={resource.id}>
      Form Adapter
    </div>
  ),
}));

vi.mock('../adapters/BlockAdapter', () => ({
  default: ({resource}: {resource: Element}) => (
    <div data-testid="block-adapter" data-resource-id={resource.id}>
      Block Adapter
    </div>
  ),
}));

vi.mock('../adapters/input/CheckboxAdapter', () => ({
  default: ({resource}: {resource: Element}) => (
    <div data-testid="checkbox-adapter" data-resource-id={resource.id}>
      Checkbox Adapter
    </div>
  ),
}));

vi.mock('../adapters/input/PhoneNumberInputAdapter', () => ({
  default: ({resource}: {resource: Element}) => (
    <div data-testid="phone-input-adapter" data-resource-id={resource.id}>
      Phone Input Adapter
    </div>
  ),
}));

vi.mock('../adapters/input/OTPInputAdapter', () => ({
  default: ({resource}: {resource: Element}) => (
    <div data-testid="otp-input-adapter" data-resource-id={resource.id}>
      OTP Input Adapter
    </div>
  ),
}));

vi.mock('../adapters/input/DefaultInputAdapter', () => ({
  default: ({resource}: {resource: Element}) => (
    <div data-testid="default-input-adapter" data-resource-id={resource.id} data-type={resource.type}>
      Default Input Adapter
    </div>
  ),
}));

vi.mock('../adapters/ChoiceAdapter', () => ({
  default: ({resource}: {resource: Element}) => (
    <div data-testid="choice-adapter" data-resource-id={resource.id}>
      Choice Adapter
    </div>
  ),
}));

vi.mock('../adapters/ButtonAdapter', () => ({
  default: ({resource, elementIndex}: {resource: Element; elementIndex?: number}) => (
    <div data-testid="button-adapter" data-resource-id={resource.id} data-element-index={elementIndex}>
      Button Adapter
    </div>
  ),
}));

vi.mock('../adapters/TypographyAdapter', () => ({
  default: ({stepId, resource}: {stepId: string; resource: Element}) => (
    <div data-testid="typography-adapter" data-step-id={stepId} data-resource-id={resource.id}>
      Typography Adapter
    </div>
  ),
}));

vi.mock('../adapters/DividerAdapter', () => ({
  default: ({resource}: {resource: Element}) => (
    <div data-testid="divider-adapter" data-resource-id={resource.id}>
      Divider Adapter
    </div>
  ),
}));

vi.mock('../adapters/RichTextAdapter', () => ({
  default: ({resource}: {resource: Element}) => (
    <div data-testid="rich-text-adapter" data-resource-id={resource.id}>
      Rich Text Adapter
    </div>
  ),
}));

vi.mock('../adapters/ImageAdapter', () => ({
  default: ({resource}: {resource: Element}) => (
    <div data-testid="image-adapter" data-resource-id={resource.id}>
      Image Adapter
    </div>
  ),
}));

vi.mock('../adapters/CaptchaAdapter', () => ({
  default: ({resource}: {resource: Element}) => (
    <div data-testid="captcha-adapter" data-resource-id={resource.id}>
      Captcha Adapter
    </div>
  ),
}));

vi.mock('../adapters/ResendButtonAdapter', () => ({
  default: ({stepId, resource}: {stepId: string; resource: Element}) => (
    <div data-testid="resend-button-adapter" data-step-id={stepId} data-resource-id={resource.id}>
      Resend Button Adapter
    </div>
  ),
}));

vi.mock('../adapters/IconAdapter', () => ({
  default: ({resource}: {resource: Element}) => (
    <div data-testid="icon-adapter" data-resource-id={resource.id}>
      Icon Adapter
    </div>
  ),
}));

vi.mock('../adapters/StackAdapter', () => ({
  default: ({stepId, resource}: {stepId: string; resource: Element}) => (
    <div data-testid="stack-adapter" data-step-id={stepId} data-resource-id={resource.id}>
      Stack Adapter
    </div>
  ),
}));

describe('CommonElementFactory', () => {
  const createMockElement = (overrides: Partial<Element> = {}): Element =>
    ({
      id: 'element-1',
      type: ElementTypes.TextInput,
      category: ElementCategories.Field,
      config: {},
      ...overrides,
    }) as Element;

  describe('Form Block', () => {
    it('should render FormAdapter for Form block with BLOCK category', async () => {
      const formElement = createMockElement({
        type: BlockTypes.Form,
        category: ElementCategories.Block,
      });

      await render(<CommonElementFactory stepId="step-1" resource={formElement} />);

      await expect.element(page.getByTestId('form-adapter')).toBeInTheDocument();
      await expect.element(page.getByTestId('form-adapter')).toHaveAttribute('data-step-id', 'step-1');
      await expect.element(page.getByTestId('form-adapter')).toHaveAttribute('data-resource-id', 'element-1');
    });

    it('should render BlockAdapter for Form block with non-BLOCK category', async () => {
      const actionBlock = createMockElement({
        type: BlockTypes.Form,
        category: ElementCategories.Action,
      });

      await render(<CommonElementFactory stepId="step-1" resource={actionBlock} />);

      await expect.element(page.getByTestId('block-adapter')).toBeInTheDocument();
    });

    it('should pass availableElements and onAddElementToForm to FormAdapter', async () => {
      const formElement = createMockElement({
        type: BlockTypes.Form,
        category: ElementCategories.Block,
      });

      const availableElements = [createMockElement({id: 'available-1'})];
      const onAddElementToForm = vi.fn();

      await render(
        <CommonElementFactory
          stepId="step-1"
          resource={formElement}
          availableElements={availableElements}
          onAddElementToForm={onAddElementToForm}
        />,
      );

      await expect.element(page.getByTestId('form-adapter')).toBeInTheDocument();
    });
  });

  describe('Checkbox Element', () => {
    it('should render CheckboxAdapter for Checkbox type', async () => {
      const checkboxElement = createMockElement({
        type: ElementTypes.Checkbox,
      });

      await render(<CommonElementFactory stepId="step-1" resource={checkboxElement} />);

      await expect.element(page.getByTestId('checkbox-adapter')).toBeInTheDocument();
    });
  });

  describe('Phone Input Element', () => {
    it('should render PhoneNumberInputAdapter for PhoneInput type', async () => {
      const phoneElement = createMockElement({
        type: ElementTypes.PhoneInput,
      });

      await render(<CommonElementFactory stepId="step-1" resource={phoneElement} />);

      await expect.element(page.getByTestId('phone-input-adapter')).toBeInTheDocument();
    });
  });

  describe('OTP Input Element', () => {
    it('should render OTPInputAdapter for OtpInput type', async () => {
      const otpElement = createMockElement({
        type: ElementTypes.OtpInput,
      });

      await render(<CommonElementFactory stepId="step-1" resource={otpElement} />);

      await expect.element(page.getByTestId('otp-input-adapter')).toBeInTheDocument();
    });
  });

  describe('Default Input Elements', () => {
    it('should render DefaultInputAdapter for TextInput type', async () => {
      const textInputElement = createMockElement({
        type: ElementTypes.TextInput,
      });

      await render(<CommonElementFactory stepId="step-1" resource={textInputElement} />);

      await expect.element(page.getByTestId('default-input-adapter')).toBeInTheDocument();
      await expect.element(page.getByTestId('default-input-adapter')).toHaveAttribute('data-type', ElementTypes.TextInput);
    });

    it('should render DefaultInputAdapter for PasswordInput type', async () => {
      const passwordElement = createMockElement({
        type: ElementTypes.PasswordInput,
      });

      await render(<CommonElementFactory stepId="step-1" resource={passwordElement} />);

      await expect.element(page.getByTestId('default-input-adapter')).toBeInTheDocument();
      await expect.element(page.getByTestId('default-input-adapter')).toHaveAttribute('data-type', ElementTypes.PasswordInput);
    });

    it('should render DefaultInputAdapter for EmailInput type', async () => {
      const emailElement = createMockElement({
        type: ElementTypes.EmailInput,
      });

      await render(<CommonElementFactory stepId="step-1" resource={emailElement} />);

      await expect.element(page.getByTestId('default-input-adapter')).toBeInTheDocument();
      await expect.element(page.getByTestId('default-input-adapter')).toHaveAttribute('data-type', ElementTypes.EmailInput);
    });

    it('should render DefaultInputAdapter for NumberInput type', async () => {
      const numberElement = createMockElement({
        type: ElementTypes.NumberInput,
      });

      await render(<CommonElementFactory stepId="step-1" resource={numberElement} />);

      await expect.element(page.getByTestId('default-input-adapter')).toBeInTheDocument();
      await expect.element(page.getByTestId('default-input-adapter')).toHaveAttribute('data-type', ElementTypes.NumberInput);
    });

    it('should render DefaultInputAdapter for DateInput type', async () => {
      const dateElement = createMockElement({
        type: ElementTypes.DateInput,
      });

      await render(<CommonElementFactory stepId="step-1" resource={dateElement} />);

      await expect.element(page.getByTestId('default-input-adapter')).toBeInTheDocument();
      await expect.element(page.getByTestId('default-input-adapter')).toHaveAttribute('data-type', ElementTypes.DateInput);
    });
  });

  describe('Dropdown Element', () => {
    it('should render ChoiceAdapter for Dropdown type', async () => {
      const dropdownElement = createMockElement({
        type: ElementTypes.Dropdown,
      });

      await render(<CommonElementFactory stepId="step-1" resource={dropdownElement} />);

      await expect.element(page.getByTestId('choice-adapter')).toBeInTheDocument();
    });
  });

  describe('Action Element', () => {
    it('should render ButtonAdapter for Action type', async () => {
      const actionElement = createMockElement({
        type: ElementTypes.Action,
      });

      await render(<CommonElementFactory stepId="step-1" resource={actionElement} />);

      await expect.element(page.getByTestId('button-adapter')).toBeInTheDocument();
    });

    it('should pass elementIndex to ButtonAdapter', async () => {
      const actionElement = createMockElement({
        type: ElementTypes.Action,
      });

      await render(<CommonElementFactory stepId="step-1" resource={actionElement} elementIndex={5} />);

      await expect.element(page.getByTestId('button-adapter')).toHaveAttribute('data-element-index', '5');
    });
  });

  describe('Text Element', () => {
    it('should render TypographyAdapter for Text type', async () => {
      const textElement = createMockElement({
        type: ElementTypes.Text,
      });

      await render(<CommonElementFactory stepId="step-1" resource={textElement} />);

      await expect.element(page.getByTestId('typography-adapter')).toBeInTheDocument();
      await expect.element(page.getByTestId('typography-adapter')).toHaveAttribute('data-step-id', 'step-1');
    });
  });

  describe('RichText Element', () => {
    it('should render RichTextAdapter for RichText type', async () => {
      const richTextElement = createMockElement({
        type: ElementTypes.RichText,
      });

      await render(<CommonElementFactory stepId="step-1" resource={richTextElement} />);

      await expect.element(page.getByTestId('rich-text-adapter')).toBeInTheDocument();
    });
  });

  describe('Divider Element', () => {
    it('should render DividerAdapter for Divider type', async () => {
      const dividerElement = createMockElement({
        type: ElementTypes.Divider,
      });

      await render(<CommonElementFactory stepId="step-1" resource={dividerElement} />);

      await expect.element(page.getByTestId('divider-adapter')).toBeInTheDocument();
    });
  });

  describe('Image Element', () => {
    it('should render ImageAdapter for Image type', async () => {
      const imageElement = createMockElement({
        type: ElementTypes.Image,
      });

      await render(<CommonElementFactory stepId="step-1" resource={imageElement} />);

      await expect.element(page.getByTestId('image-adapter')).toBeInTheDocument();
    });
  });

  describe('Captcha Element', () => {
    it('should render CaptchaAdapter for Captcha type', async () => {
      const captchaElement = createMockElement({
        type: ElementTypes.Captcha,
      });

      await render(<CommonElementFactory stepId="step-1" resource={captchaElement} />);

      await expect.element(page.getByTestId('captcha-adapter')).toBeInTheDocument();
    });
  });

  describe('Resend Element', () => {
    it('should render ResendButtonAdapter for Resend type', async () => {
      const resendElement = createMockElement({
        type: ElementTypes.Resend,
      });

      await render(<CommonElementFactory stepId="step-1" resource={resendElement} />);

      await expect.element(page.getByTestId('resend-button-adapter')).toBeInTheDocument();
      await expect.element(page.getByTestId('resend-button-adapter')).toHaveAttribute('data-step-id', 'step-1');
    });
  });

  describe('Icon Element', () => {
    it('should render IconAdapter for Icon type', () => {
      const iconElement = createMockElement({
        type: ElementTypes.Icon,
      });

      render(<CommonElementFactory stepId="step-1" resource={iconElement} />);

      expect(screen.getByTestId('icon-adapter')).toBeInTheDocument();
      expect(screen.getByTestId('icon-adapter')).toHaveAttribute('data-resource-id', 'element-1');
    });
  });

  describe('Stack Element', () => {
    it('should render StackAdapter for Stack type', () => {
      const stackElement = createMockElement({
        type: ElementTypes.Stack,
      });

      render(<CommonElementFactory stepId="step-1" resource={stackElement} />);

      expect(screen.getByTestId('stack-adapter')).toBeInTheDocument();
      expect(screen.getByTestId('stack-adapter')).toHaveAttribute('data-step-id', 'step-1');
      expect(screen.getByTestId('stack-adapter')).toHaveAttribute('data-resource-id', 'element-1');
    });
  });

  describe('Unknown Element Type', () => {
    it('should return null for unknown element type', async () => {
      const unknownElement = createMockElement({
        type: 'UNKNOWN_TYPE' as typeof ElementTypes[keyof typeof ElementTypes],
      });

      const {container} = await render(<CommonElementFactory stepId="step-1" resource={unknownElement} />);

      expect(container.firstChild).toBeNull();
    });
  });

  describe('Default Props', () => {
    it('should work with undefined elementIndex', async () => {
      const actionElement = createMockElement({
        type: ElementTypes.Action,
      });

      await render(<CommonElementFactory stepId="step-1" resource={actionElement} />);

      await expect.element(page.getByTestId('button-adapter')).toBeInTheDocument();
    });

    it('should work with undefined availableElements', async () => {
      const formElement = createMockElement({
        type: BlockTypes.Form,
        category: ElementCategories.Block,
      });

      await render(<CommonElementFactory stepId="step-1" resource={formElement} />);

      await expect.element(page.getByTestId('form-adapter')).toBeInTheDocument();
    });

    it('should work with undefined onAddElementToForm', async () => {
      const formElement = createMockElement({
        type: BlockTypes.Form,
        category: ElementCategories.Block,
      });

      await render(<CommonElementFactory stepId="step-1" resource={formElement} />);

      await expect.element(page.getByTestId('form-adapter')).toBeInTheDocument();
    });
  });
});
