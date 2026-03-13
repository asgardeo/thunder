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

import {describe, it, expect, vi} from 'vitest';
import {render, screen} from '@thunder/test-utils';
import userEvent from '@testing-library/user-event';
import ColorEditRow from '../ColorEditRow';

describe('ColorEditRow', () => {
  describe('Compact mode', () => {
    it('renders the label', () => {
      render(<ColorEditRow label="Primary" value="#ff0000" onChange={vi.fn()} compact />);
      expect(screen.getByText('Primary')).toBeInTheDocument();
    });

    it('renders the color value as text', () => {
      render(<ColorEditRow label="Background" value="#1a2b3c" onChange={vi.fn()} compact />);
      expect(screen.getByText('#1a2b3c')).toBeInTheDocument();
    });

    it('renders a color swatch box', () => {
      const {container} = render(<ColorEditRow label="Color" value="#abcdef" onChange={vi.fn()} compact />);
      // The swatch is a Box with bgcolor — rendered as a div
      expect(container.querySelector('div')).toBeInTheDocument();
    });

    it('renders the color picker input (hidden) when value is a valid hex', () => {
      const {container} = render(<ColorEditRow label="Color" value="#abcdef" onChange={vi.fn()} compact />);
      expect(container.querySelector('input[type="color"]')).toBeInTheDocument();
    });

    it('does not render the color picker input when value is not a valid hex', () => {
      const {container} = render(<ColorEditRow label="Color" value="var(--token)" onChange={vi.fn()} compact />);
      expect(container.querySelector('input[type="color"]')).not.toBeInTheDocument();
    });

    it('does not render a text field in compact mode', () => {
      render(<ColorEditRow label="Color" value="#ff0000" onChange={vi.fn()} compact />);
      expect(screen.queryByRole('textbox')).not.toBeInTheDocument();
    });
  });

  describe('Full mode (default)', () => {
    it('renders the label in full mode', () => {
      render(<ColorEditRow label="Primary Color" value="#123456" onChange={vi.fn()} />);
      expect(screen.getByText('Primary Color')).toBeInTheDocument();
    });

    it('renders a text input in full mode', () => {
      render(<ColorEditRow label="Color" value="#aabbcc" onChange={vi.fn()} />);
      expect(screen.getByRole('textbox')).toBeInTheDocument();
    });

    it('text input shows the current value', () => {
      render(<ColorEditRow label="Color" value="#aabbcc" onChange={vi.fn()} />);
      expect(screen.getByRole('textbox')).toHaveValue('#aabbcc');
    });

    it('renders a color swatch and hidden color picker in full mode', () => {
      const {container} = render(<ColorEditRow label="Color" value="#11aaff" onChange={vi.fn()} />);
      expect(container.querySelector('input[type="color"]')).toBeInTheDocument();
    });

    it('renders hex hint for valid hex values', () => {
      render(<ColorEditRow label="Color" value="#11aaff" onChange={vi.fn()} />);
      // The hex hint shows the lowercase version of the value
      expect(screen.getByText('#11aaff')).toBeInTheDocument();
    });
  });

  describe('onChange behaviour (full mode)', () => {
    it('calls onChange when a valid 6-digit hex is typed', async () => {
      const onChange = vi.fn();
      const user = userEvent.setup();
      render(<ColorEditRow label="Color" value="#000000" onChange={onChange} />);

      const input = screen.getByRole('textbox');
      await user.clear(input);
      await user.type(input, '#ffffff');

      expect(onChange).toHaveBeenCalledWith('#ffffff');
    });

    it('does NOT call onChange for an incomplete hex string', async () => {
      const onChange = vi.fn();
      const user = userEvent.setup();
      render(<ColorEditRow label="Color" value="#000000" onChange={onChange} />);

      const input = screen.getByRole('textbox');
      await user.clear(input);
      await user.type(input, '#abc');

      // Only 3 chars — not a valid full hex
      expect(onChange).not.toHaveBeenCalledWith('#abc');
    });

    it('does NOT call onChange for a non-hex string', async () => {
      const onChange = vi.fn();
      const user = userEvent.setup();
      render(<ColorEditRow label="Color" value="#000000" onChange={onChange} />);

      const input = screen.getByRole('textbox');
      await user.clear(input);
      await user.type(input, 'notacolor');

      expect(onChange).not.toHaveBeenCalled();
    });

    it('resets editValue to current value on blur', async () => {
      const user = userEvent.setup();
      render(<ColorEditRow label="Color" value="#ff0000" onChange={vi.fn()} />);

      const input = screen.getByRole('textbox');
      await user.click(input);
      await user.clear(input);
      await user.type(input, '#bad');
      await user.tab(); // trigger blur

      // After blur, value should revert to the prop value
      expect(input).toHaveValue('#ff0000');
    });
  });

  describe('Value sync', () => {
    it('updates the text input when the external value changes and input is not focused', () => {
      const {rerender} = render(<ColorEditRow label="Color" value="#111111" onChange={vi.fn()} />);
      expect(screen.getByRole('textbox')).toHaveValue('#111111');

      rerender(<ColorEditRow label="Color" value="#222222" onChange={vi.fn()} />);
      expect(screen.getByRole('textbox')).toHaveValue('#222222');
    });
  });
});
