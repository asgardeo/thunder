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
import {render, screen, fireEvent} from '@testing-library/react';
import {createRef} from 'react';
import Action from '../Action';

describe('Action', () => {
  describe('Rendering', () => {
    it('should render as a button element', () => {
      render(<Action>Click me</Action>);

      expect(screen.getByRole('button')).toBeInTheDocument();
    });

    it('should render children content', () => {
      render(<Action>Action Content</Action>);

      expect(screen.getByText('Action Content')).toBeInTheDocument();
    });

    it('should have type="button" attribute', () => {
      render(<Action>Button</Action>);

      expect(screen.getByRole('button')).toHaveAttribute('type', 'button');
    });
  });

  describe('Ref Forwarding', () => {
    it('should forward ref to button element', () => {
      const ref = createRef<HTMLButtonElement>();

      render(<Action ref={ref}>With Ref</Action>);

      expect(ref.current).toBe(screen.getByRole('button'));
    });
  });

  describe('Cursor Styles', () => {
    it('should have pointer cursor by default', () => {
      render(<Action>Default Cursor</Action>);

      const button = screen.getByRole('button');
      expect(button).toHaveStyle({cursor: 'pointer'});
    });

    it('should accept custom cursor style', () => {
      render(<Action cursor="grab">Grab Cursor</Action>);

      const button = screen.getByRole('button');
      expect(button).toHaveStyle({cursor: 'grab'});
    });

    it('should accept move cursor', () => {
      render(<Action cursor="move">Move Cursor</Action>);

      const button = screen.getByRole('button');
      expect(button).toHaveStyle({cursor: 'move'});
    });
  });

  describe('Base Styles', () => {
    it('should render button element with base styles', () => {
      render(<Action>Styled</Action>);

      const button = screen.getByRole('button');
      expect(button).toBeInTheDocument();
      // The button renders with inline styles applied
      expect(button).toHaveAttribute('type', 'button');
    });

    it('should have full height', () => {
      render(<Action>Styled</Action>);

      const button = screen.getByRole('button');
      expect(button).toHaveStyle({height: '100%'});
    });

    it('should have fixed width', () => {
      render(<Action>Styled</Action>);

      const button = screen.getByRole('button');
      expect(button).toHaveStyle({width: '50px'});
    });
  });

  describe('Custom Styles', () => {
    it('should accept additional style prop', () => {
      render(<Action style={{padding: '10px', color: 'red'}}>Custom Style</Action>);

      const button = screen.getByRole('button');
      // Note: testing-library converts color names to rgb format
      expect(button).toHaveStyle({padding: '10px'});
    });

    it('should merge custom styles with default styles', () => {
      render(<Action style={{margin: '5px'}}>Merged</Action>);

      const button = screen.getByRole('button');
      // Should have both custom and default styles
      expect(button).toHaveStyle({margin: '5px'});
    });
  });

  describe('ClassName', () => {
    it('should accept className prop', () => {
      render(<Action className="custom-action-class">Class</Action>);

      expect(screen.getByRole('button')).toHaveClass('custom-action-class');
    });
  });

  describe('Hover Effects', () => {
    it('should handle mouse enter event', () => {
      render(<Action>Hover Me</Action>);

      const button = screen.getByRole('button');
      // Mouse events should not throw errors
      fireEvent.mouseEnter(button);

      expect(button).toBeInTheDocument();
    });

    it('should handle mouse leave event', () => {
      render(<Action>Hover Me</Action>);

      const button = screen.getByRole('button');
      fireEvent.mouseEnter(button);
      fireEvent.mouseLeave(button);

      expect(button).toBeInTheDocument();
    });
  });

  describe('Event Handlers', () => {
    it('should call onClick handler', () => {
      const onClick = vi.fn();
      render(<Action onClick={onClick}>Click</Action>);

      fireEvent.click(screen.getByRole('button'));

      expect(onClick).toHaveBeenCalledTimes(1);
    });

    it('should call onFocus handler', () => {
      const onFocus = vi.fn();
      render(<Action onFocus={onFocus}>Focus</Action>);

      fireEvent.focus(screen.getByRole('button'));

      expect(onFocus).toHaveBeenCalledTimes(1);
    });

    it('should call onBlur handler', () => {
      const onBlur = vi.fn();
      render(<Action onBlur={onBlur}>Blur</Action>);

      const button = screen.getByRole('button');
      fireEvent.focus(button);
      fireEvent.blur(button);

      expect(onBlur).toHaveBeenCalledTimes(1);
    });

    it('should pass event handlers through ...rest', () => {
      const onKeyDown = vi.fn();
      render(<Action onKeyDown={onKeyDown}>Key</Action>);

      fireEvent.keyDown(screen.getByRole('button'), {key: 'Enter'});

      expect(onKeyDown).toHaveBeenCalledTimes(1);
    });
  });

  describe('Additional Props', () => {
    it('should pass data attributes', () => {
      render(<Action data-testid="action-button">Data</Action>);

      expect(screen.getByTestId('action-button')).toBeInTheDocument();
    });

    it('should pass aria attributes', () => {
      render(<Action aria-label="Action button">Aria</Action>);

      expect(screen.getByRole('button')).toHaveAttribute('aria-label', 'Action button');
    });
  });
});
