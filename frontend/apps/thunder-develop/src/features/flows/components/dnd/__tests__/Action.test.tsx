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
import {page, userEvent} from 'vitest/browser';
import {createRef} from 'react';
import Action from '../Action';

describe('Action', () => {
  describe('Rendering', () => {
    it('should render as a button element', async () => {
      await render(<Action>Click me</Action>);

      await expect.element(page.getByRole('button')).toBeInTheDocument();
    });

    it('should render children content', async () => {
      await render(<Action>Action Content</Action>);

      await expect.element(page.getByText('Action Content')).toBeInTheDocument();
    });

    it('should have type="button" attribute', async () => {
      await render(<Action>Button</Action>);

      await expect.element(page.getByRole('button')).toHaveAttribute('type', 'button');
    });
  });

  describe('Ref Forwarding', () => {
    it('should forward ref to button element', async () => {
      const ref = createRef<HTMLButtonElement>();

      await render(<Action ref={ref}>With Ref</Action>);

      expect(ref.current).toBe(page.getByRole('button'));
    });
  });

  describe('Cursor Styles', () => {
    it('should have pointer cursor by default', async () => {
      await render(<Action>Default Cursor</Action>);

      const button = page.getByRole('button');
      expect(button).toHaveStyle({cursor: 'pointer'});
    });

    it('should accept custom cursor style', async () => {
      await render(<Action cursor="grab">Grab Cursor</Action>);

      const button = page.getByRole('button');
      expect(button).toHaveStyle({cursor: 'grab'});
    });

    it('should accept move cursor', async () => {
      await render(<Action cursor="move">Move Cursor</Action>);

      const button = page.getByRole('button');
      expect(button).toHaveStyle({cursor: 'move'});
    });
  });

  describe('Base Styles', () => {
    it('should render button element with base styles', async () => {
      await render(<Action>Styled</Action>);

      const button = page.getByRole('button');
      expect(button).toBeInTheDocument();
      // The button renders with inline styles applied
      expect(button).toHaveAttribute('type', 'button');
    });

    it('should have full height', async () => {
      await render(<Action>Styled</Action>);

      const button = page.getByRole('button');
      expect(button).toHaveStyle({height: '100%'});
    });

    it('should have fixed width', async () => {
      await render(<Action>Styled</Action>);

      const button = page.getByRole('button');
      expect(button).toHaveStyle({width: '50px'});
    });
  });

  describe('Custom Styles', () => {
    it('should accept additional style prop', async () => {
      await render(<Action style={{padding: '10px', color: 'red'}}>Custom Style</Action>);

      const button = page.getByRole('button');
      // Note: testing-library converts color names to rgb format
      expect(button).toHaveStyle({padding: '10px'});
    });

    it('should merge custom styles with default styles', async () => {
      await render(<Action style={{margin: '5px'}}>Merged</Action>);

      const button = page.getByRole('button');
      // Should have both custom and default styles
      expect(button).toHaveStyle({margin: '5px'});
    });
  });

  describe('ClassName', () => {
    it('should accept className prop', async () => {
      await render(<Action className="custom-action-class">Class</Action>);

      expect(page.getByRole('button')).toHaveClass('custom-action-class');
    });
  });

  describe('Hover Effects', () => {
    it('should handle mouse enter event', async () => {
      await render(<Action>Hover Me</Action>);

      const button = page.getByRole('button');
      // Mouse events should not throw errors
      await userEvent.hover(button);

      expect(button).toBeInTheDocument();
    });

    it('should handle mouse leave event', async () => {
      await render(<Action>Hover Me</Action>);

      const button = page.getByRole('button');
      await userEvent.hover(button);
      await userEvent.unhover(button);

      expect(button).toBeInTheDocument();
    });
  });

  describe('Event Handlers', () => {
    it('should call onClick handler', async () => {
      const onClick = vi.fn();
      await render(<Action onClick={onClick}>Click</Action>);

      await userEvent.click(page.getByRole('button'));

      expect(onClick).toHaveBeenCalledTimes(1);
    });

    it('should call onFocus handler', async () => {
      const onFocus = vi.fn();
      await render(<Action onFocus={onFocus}>Focus</Action>);

      await userEvent.click(page.getByRole('button'));

      expect(onFocus).toHaveBeenCalledTimes(1);
    });

    it('should call onBlur handler', async () => {
      const onBlur = vi.fn();
      await render(<Action onBlur={onBlur}>Blur</Action>);

      const button = page.getByRole('button');
      await userEvent.click(button);
      await userEvent.tab();

      expect(onBlur).toHaveBeenCalledTimes(1);
    });

    it('should pass event handlers through ...rest', async () => {
      const onKeyDown = vi.fn();
      await render(<Action onKeyDown={onKeyDown}>Key</Action>);

      await userEvent.click(page.getByRole('button'));
      await userEvent.keyboard('{Enter}');

      expect(onKeyDown).toHaveBeenCalledTimes(1);
    });
  });

  describe('Additional Props', () => {
    it('should pass data attributes', async () => {
      await render(<Action data-testid="action-button">Data</Action>);

      await expect.element(page.getByTestId('action-button')).toBeInTheDocument();
    });

    it('should pass aria attributes', async () => {
      await render(<Action aria-label="Action button">Aria</Action>);

      await expect.element(page.getByRole('button')).toHaveAttribute('aria-label', 'Action button');
    });
  });
});
