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
import Handle from '../Handle';

describe('Handle', () => {
  describe('Rendering', () => {
    it('should render as a button element', async () => {
      await render(<Handle label="Drag handle">Handle Content</Handle>);

      await expect.element(page.getByRole('button')).toBeInTheDocument();
    });

    it('should render children content', async () => {
      await render(<Handle label="Label">Icon Content</Handle>);

      await expect.element(page.getByText('Icon Content')).toBeInTheDocument();
    });

    it('should wrap children in a span', async () => {
      await render(<Handle label="Label">Content</Handle>);

      const span = page.getByText('Content').element().closest('span');
      expect(span).toBeInTheDocument();
    });
  });

  describe('Label/Tooltip', () => {
    it('should have Tooltip with label', async () => {
      await render(<Handle label="Drag to reorder">Handle</Handle>);

      // The button should be present, Tooltip provides aria support
      await expect.element(page.getByRole('button')).toBeInTheDocument();
    });

    it('should accept ReactNode as label', async () => {
      await render(
        <Handle label={<strong>Bold Label</strong>}>
          Handle
        </Handle>,
      );

      await expect.element(page.getByRole('button')).toBeInTheDocument();
    });

    it('should accept string as label', async () => {
      await render(<Handle label="String Label">Handle</Handle>);

      await expect.element(page.getByRole('button')).toBeInTheDocument();
    });
  });

  describe('Ref Forwarding', () => {
    it('should forward ref to Action/button element', async () => {
      const ref = createRef<HTMLButtonElement>();

      await render(
        <Handle label="Handle" ref={ref}>
          With Ref
        </Handle>,
      );

      expect(ref.current).toBe(page.getByRole('button'));
    });

    it('should handle null ref gracefully', async () => {
      await render(
        <Handle label="Handle" ref={null}>
          No Ref
        </Handle>,
      );

      await expect.element(page.getByRole('button')).toBeInTheDocument();
    });
  });

  describe('Cursor Styles', () => {
    it('should have pointer cursor by default', async () => {
      await render(<Handle label="Label">Default</Handle>);

      const button = page.getByRole('button');
      expect(button).toHaveStyle({cursor: 'pointer'});
    });

    it('should accept custom cursor style', async () => {
      await render(
        <Handle label="Label" cursor="grab">
          Grab
        </Handle>,
      );

      const button = page.getByRole('button');
      expect(button).toHaveStyle({cursor: 'grab'});
    });

    it('should accept grabbing cursor', async () => {
      await render(
        <Handle label="Label" cursor="grabbing">
          Grabbing
        </Handle>,
      );

      const button = page.getByRole('button');
      expect(button).toHaveStyle({cursor: 'grabbing'});
    });
  });

  describe('Event Handlers', () => {
    it('should call onClick handler', async () => {
      const onClick = vi.fn();
      await render(
        <Handle label="Label" onClick={onClick}>
          Click
        </Handle>,
      );

      await userEvent.click(page.getByRole('button'));

      expect(onClick).toHaveBeenCalledTimes(1);
    });

    it('should call onMouseDown handler', async () => {
      const onMouseDown = vi.fn();
      await render(
        <Handle label="Label" onMouseDown={onMouseDown}>
          Mouse Down
        </Handle>,
      );

      await userEvent.click(page.getByRole('button'));

      expect(onMouseDown).toHaveBeenCalledTimes(1);
    });

    it('should call onKeyDown handler', async () => {
      const onKeyDown = vi.fn();
      await render(
        <Handle label="Label" onKeyDown={onKeyDown}>
          Key Down
        </Handle>,
      );

      await userEvent.keyboard('{Space}');

      expect(onKeyDown).toHaveBeenCalledTimes(1);
    });
  });

  describe('Integration with Action', () => {
    it('should pass props to underlying Action component', async () => {
      await render(
        <Handle label="Handle" className="custom-handle-class">
          Content
        </Handle>,
      );

      expect(page.getByRole('button')).toHaveClass('custom-handle-class');
    });

    it('should inherit Action hover effects', async () => {
      await render(<Handle label="Handle">Hover</Handle>);

      const button = page.getByRole('button');
      await userEvent.hover(button);

      expect(button).toHaveStyle({backgroundColor: 'rgba(0, 0, 0, 0.15)'});
    });
  });

  describe('Accessibility', () => {
    it('should be focusable', async () => {
      await render(<Handle label="Handle">Focus Me</Handle>);

      const button = page.getByRole('button');
      button.element().focus();

      expect(document.activeElement).toBe(button);
    });

    it('should accept aria attributes', async () => {
      await render(
        <Handle label="Handle" aria-pressed="true">
          Aria
        </Handle>,
      );

      await expect.element(page.getByRole('button')).toHaveAttribute('aria-pressed', 'true');
    });
  });

  describe('Children Variants', () => {
    it('should render icon as child', async () => {
      await render(
        <Handle label="Drag">
          <svg data-testid="icon" />
        </Handle>,
      );

      await expect.element(page.getByTestId('icon')).toBeInTheDocument();
    });

    it('should render text as child', async () => {
      await render(<Handle label="Handle">Text Handle</Handle>);

      await expect.element(page.getByText('Text Handle')).toBeInTheDocument();
    });

    it('should render multiple children', async () => {
      await render(
        <Handle label="Handle">
          <span data-testid="icon">Icon</span>
          <span data-testid="text">Text</span>
        </Handle>,
      );

      await expect.element(page.getByTestId('icon')).toBeInTheDocument();
      await expect.element(page.getByTestId('text')).toBeInTheDocument();
    });
  });
});
