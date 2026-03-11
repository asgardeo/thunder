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
import {render, page, userEvent} from '@thunder/test-utils/browser';
import AddCard from '../AddCard';

describe('AddCard', () => {
  describe('Rendering', () => {
    it('renders the label text', async () => {
      await render(<AddCard label="Add Theme" onClick={vi.fn()} />);
      await expect.element(page.getByText('Add Theme')).toBeInTheDocument();
    });

    it('renders the plus icon', async () => {
      const {container} = await render(<AddCard label="Add Layout" onClick={vi.fn()} />);
      // The Plus icon is an SVG element
      expect(container.querySelector('svg')).toBeInTheDocument();
    });

    it('is rendered as a clickable box', async () => {
      const {container} = await render(<AddCard label="Click me" onClick={vi.fn()} />);
      // The top-level Box has onClick — just verify the component renders without errors
      expect(container.firstChild).toBeInTheDocument();
    });
  });

  describe('Interaction', () => {
    it('calls onClick when the card is clicked', async () => {
      const onClick = vi.fn();
      await render(<AddCard label="Add" onClick={onClick} />);

      await userEvent.click(page.getByText('Add'));

      expect(onClick).toHaveBeenCalledOnce();
    });

    it('calls onClick each time the card is clicked', async () => {
      const onClick = vi.fn();
      await render(<AddCard label="Add" onClick={onClick} />);

      await userEvent.click(page.getByText('Add'));
      await userEvent.click(page.getByText('Add'));

      expect(onClick).toHaveBeenCalledTimes(2);
    });

    it('does not call onClick when a different element is clicked', async () => {
      const onClick = vi.fn();
      await render(
        <div>
          <AddCard label="Add" onClick={onClick} />
          <button type="button">Other</button>
        </div>,
      );

      await userEvent.click(page.getByText('Other'));

      expect(onClick).not.toHaveBeenCalled();
    });
  });
});
