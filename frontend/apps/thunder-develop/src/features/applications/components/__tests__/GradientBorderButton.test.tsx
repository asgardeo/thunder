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

import {describe, it, expect} from 'vitest';
import {page, userEvent} from 'vitest/browser';
import {render} from '@thunder/test-utils/browser';
import GradientBorderButton from '../GradientBorderButton';

describe('GradientBorderButton', () => {
  describe('Rendering', () => {
    it('should render button with children text', async () => {
      await render(<GradientBorderButton>Click Me</GradientBorderButton>);

      await expect.element(page.getByRole('button', {name: 'Click Me'})).toBeInTheDocument();
    });

    it('should render as text variant by default', async () => {
      await render(<GradientBorderButton>Button</GradientBorderButton>);

      const button = page.getByRole('button');

      expect(button.element().className).toContain('MuiButton-text');
    });

    it('should have ripple disabled by default', async () => {
      await render(<GradientBorderButton>Button</GradientBorderButton>);

      const button = page.getByRole('button');

      // Check that the button has the TouchRipple component disabled via the disableRipple prop
      // MUI conditionally renders the ripple span only when ripple is enabled
      const rippleElement = button.element().querySelector('.MuiTouchRipple-root');
      expect(rippleElement).toBeNull();
    });
  });

  describe('Props', () => {
    it('should accept and apply custom className', async () => {
      await render(<GradientBorderButton className="custom-class">Button</GradientBorderButton>);

      const button = page.getByRole('button');

      expect(button.element().className).toContain('custom-class');
    });

    it('should handle onClick event', async () => {
      let clicked = false;
      const handleClick = () => {
        clicked = true;
      };

      await render(<GradientBorderButton onClick={handleClick}>Click Me</GradientBorderButton>);

      await userEvent.click(page.getByRole('button'));

      expect(clicked).toBe(true);
    });

    it('should be disabled when disabled prop is true', async () => {
      await render(<GradientBorderButton disabled>Button</GradientBorderButton>);

      const button = page.getByRole('button');

      expect(button).toBeDisabled();
    });

    it('should accept startIcon prop', async () => {
      await render(<GradientBorderButton startIcon={<span data-testid="start-icon">→</span>}>Button</GradientBorderButton>);

      await expect.element(page.getByTestId('start-icon')).toBeInTheDocument();
    });

    it('should accept endIcon prop', async () => {
      await render(<GradientBorderButton endIcon={<span data-testid="end-icon">←</span>}>Button</GradientBorderButton>);

      await expect.element(page.getByTestId('end-icon')).toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('should support aria-label', async () => {
      await render(<GradientBorderButton aria-label="Custom Label">Button</GradientBorderButton>);

      await expect.element(page.getByLabelText('Custom Label')).toBeInTheDocument();
    });

    it('should support aria-describedby', async () => {
      await render(
        <>
          <GradientBorderButton aria-describedby="description">Button</GradientBorderButton>
          <div id="description">Button description</div>
        </>,
      );

      const button = page.getByRole('button');

      expect(button).toHaveAttribute('aria-describedby', 'description');
    });

    it('should be keyboard accessible', async () => {
      let clicked = false;
      const handleClick = () => {
        clicked = true;
      };

      await render(<GradientBorderButton onClick={handleClick}>Button</GradientBorderButton>);

      const button = page.getByRole('button');
      button.element().focus();
      await userEvent.keyboard('{Enter}');

      expect(clicked).toBe(true);
    });
  });

  describe('Ref Forwarding', () => {
    it('should forward ref to button element', async () => {
      let buttonRef: HTMLButtonElement | null = null;
      const refCallback = (ref: HTMLButtonElement | null): void => {
        buttonRef = ref;
      };

      await render(<GradientBorderButton ref={refCallback}>Button</GradientBorderButton>);

      expect(buttonRef).toBeInstanceOf(HTMLButtonElement);
      expect(buttonRef).not.toBeNull();
      expect(buttonRef!.tagName).toBe('BUTTON');
    });
  });
});
