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
import AddCard from '../AddCard';

describe('AddCard', () => {
  describe('Rendering', () => {
    it('renders the label text', () => {
      render(<AddCard label="Add Theme" onClick={vi.fn()} />);
      expect(screen.getByText('Add Theme')).toBeInTheDocument();
    });

    it('renders the plus icon', () => {
      const {container} = render(<AddCard label="Add Layout" onClick={vi.fn()} />);
      // The Plus icon is an SVG element
      expect(container.querySelector('svg')).toBeInTheDocument();
    });

    it('is rendered as a clickable box', () => {
      const {container} = render(<AddCard label="Click me" onClick={vi.fn()} />);
      // The top-level Box has onClick — just verify the component renders without errors
      expect(container.firstChild).toBeInTheDocument();
    });
  });

  describe('Interaction', () => {
    it('calls onClick when the card is clicked', async () => {
      const onClick = vi.fn();
      const user = userEvent.setup();
      render(<AddCard label="Add" onClick={onClick} />);

      await user.click(screen.getByText('Add'));

      expect(onClick).toHaveBeenCalledOnce();
    });

    it('calls onClick each time the card is clicked', async () => {
      const onClick = vi.fn();
      const user = userEvent.setup();
      render(<AddCard label="Add" onClick={onClick} />);

      await user.click(screen.getByText('Add'));
      await user.click(screen.getByText('Add'));

      expect(onClick).toHaveBeenCalledTimes(2);
    });

    it('does not call onClick when a different element is clicked', async () => {
      const onClick = vi.fn();
      const user = userEvent.setup();
      render(
        <div>
          <AddCard label="Add" onClick={onClick} />
          <button type="button">Other</button>
        </div>,
      );

      await user.click(screen.getByText('Other'));

      expect(onClick).not.toHaveBeenCalled();
    });
  });
});
