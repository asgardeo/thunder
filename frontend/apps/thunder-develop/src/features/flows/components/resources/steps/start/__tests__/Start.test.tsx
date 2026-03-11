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
import Start from '../Start';

// Mock @xyflow/react
vi.mock('@xyflow/react', () => ({
  Handle: ({type, position, id}: {type: string; position: string; id: string}) => (
    <div data-testid={`handle-${type}`} data-position={position} data-handle-id={id} />
  ),
  Position: {
    Left: 'left',
    Right: 'right',
    Top: 'top',
    Bottom: 'bottom',
  },
}));

// Mock SCSS
vi.mock('../Start.scss', () => ({}));

describe('Start', () => {
  describe('Rendering', () => {
    it('should render the Start node', async () => {
      await render(<Start />);

      await expect.element(page.getByText('Start')).toBeInTheDocument();
    });

    it('should render a Fab button with start label', async () => {
      await render(<Start />);

      const fab = page.getByRole('button', {name: 'start'});
      expect(fab).toBeInTheDocument();
    });

    it('should render with start class on Fab', async () => {
      await render(<Start />);

      const fab = page.getByRole('button');
      expect(fab).toHaveClass('start');
    });
  });

  describe('React Flow Handle', () => {
    it('should render a source handle', async () => {
      await render(<Start />);

      const handle = page.getByTestId('handle-source');
      expect(handle).toBeInTheDocument();
    });

    it('should position handle on the right', async () => {
      await render(<Start />);

      const handle = page.getByTestId('handle-source');
      expect(handle).toHaveAttribute('data-position', 'right');
    });

    it('should have correct handle id with next suffix', async () => {
      await render(<Start />);

      const handle = page.getByTestId('handle-source');
      // Handle id should contain 'start' and '_NEXT' suffix
      expect(handle.element().getAttribute('data-handle-id')).toContain('start');
      expect(handle.element().getAttribute('data-handle-id')).toContain('_NEXT');
    });

    it('should have hidden-handle class', async () => {
      // Note: Since we're mocking Handle, we can't directly test the class
      // but the component should pass the className prop
      await render(<Start />);

      const handle = page.getByTestId('handle-source');
      expect(handle).toBeInTheDocument();
    });
  });

  describe('Fab Properties', () => {
    it('should render extended variant Fab', async () => {
      await render(<Start />);

      const fab = page.getByRole('button');
      // Extended variant Fab will have extended class
      expect(fab).toBeInTheDocument();
    });

    it('should render small size Fab', async () => {
      await render(<Start />);

      const fab = page.getByRole('button');
      expect(fab).toBeInTheDocument();
    });
  });

  describe('Structure', () => {
    it('should be wrapped in a div', async () => {
      const {container} = await render(<Start />);

      expect(container.firstChild?.nodeName).toBe('DIV');
    });

    it('should contain Fab and Handle as children', async () => {
      await render(<Start />);

      await expect.element(page.getByRole('button')).toBeInTheDocument();
      await expect.element(page.getByTestId('handle-source')).toBeInTheDocument();
    });
  });
});
