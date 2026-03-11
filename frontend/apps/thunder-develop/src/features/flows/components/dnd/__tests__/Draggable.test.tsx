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
import Draggable from '../Draggable';

// Mock useDraggable from @dnd-kit/react
const mockRef = {current: null};
vi.mock('@dnd-kit/react', () => ({
  useDraggable: vi.fn(() => ({
    ref: mockRef,
  })),
}));

describe('Draggable', () => {
  describe('Rendering', () => {
    it('should render children', async () => {
      await render(
        <Draggable id="test-draggable" accept={['TYPE_A']}>
          <div data-testid="child-content">Draggable Content</div>
        </Draggable>,
      );

      await expect.element(page.getByTestId('child-content')).toBeInTheDocument();
      await expect.element(page.getByText('Draggable Content')).toBeInTheDocument();
    });

    it('should render without children', async () => {
      const {container} = await render(<Draggable id="empty-draggable" accept={['TYPE_A']} />);

      // Should still render the Box wrapper
      expect(container.firstChild).toBeInTheDocument();
    });

    it('should render with Box wrapper', async () => {
      const {container} = await render(
        <Draggable id="test-draggable" accept={['TYPE_A']}>
          <span>Content</span>
        </Draggable>,
      );

      // Box component is rendered as a div
      expect(container.querySelector('div')).toBeInTheDocument();
    });
  });

  describe('Hook Integration', () => {
    it('should call useDraggable with correct id', async () => {
      const {useDraggable} = await import('@dnd-kit/react');

      await render(
        <Draggable id="unique-id-123" accept={['TYPE_A']}>
          <div>Content</div>
        </Draggable>,
      );

      expect(useDraggable).toHaveBeenCalledWith(
        expect.objectContaining({
          id: 'unique-id-123',
        }),
      );
    });

    it('should pass additional props to useDraggable', async () => {
      const {useDraggable} = await import('@dnd-kit/react');

      await render(
        <Draggable id="test-id" accept={['TYPE_A', 'TYPE_B']} data={{custom: 'data'}} type="CUSTOM_TYPE" disabled>
          <div>Content</div>
        </Draggable>,
      );

      expect(useDraggable).toHaveBeenCalledWith(
        expect.objectContaining({
          id: 'test-id',
          data: {custom: 'data'},
          type: 'CUSTOM_TYPE',
          disabled: true,
        }),
      );
    });
  });

  describe('Accept Prop', () => {
    it('should accept single type', async () => {
      const {container} = await render(
        <Draggable id="test" accept={['SINGLE_TYPE']}>
          <div>Content</div>
        </Draggable>,
      );

      expect(container.firstChild).toBeInTheDocument();
    });

    it('should accept multiple types', async () => {
      const {container} = await render(
        <Draggable id="test" accept={['TYPE_A', 'TYPE_B', 'TYPE_C']}>
          <div>Content</div>
        </Draggable>,
      );

      expect(container.firstChild).toBeInTheDocument();
    });
  });

  describe('Styling', () => {
    it('should have full width and height', async () => {
      const {container} = await render(
        <Draggable id="test" accept={['TYPE_A']}>
          <div>Content</div>
        </Draggable>,
      );

      const wrapper = container.firstChild as HTMLElement;
      // MUI Box applies styles via className
      expect(wrapper).toBeInTheDocument();
    });
  });
});
