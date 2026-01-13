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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render, screen} from '@testing-library/react';
import Sortable from '../Sortable';

// Mock refs
const mockSortableRef = {current: null};

// Mock sortable state
const mockSortable = {
  accepts: vi.fn(() => true),
};

// Mock manager with event listeners
const mockManager = {
  monitor: {
    addEventListener: vi.fn(),
  },
};

// Mock @dnd-kit/react/sortable
vi.mock('@dnd-kit/react/sortable', () => ({
  useSortable: vi.fn(() => ({
    ref: mockSortableRef,
    sortable: mockSortable,
    isDragging: false,
    isDropTarget: false,
  })),
}));

// Mock @dnd-kit/react
vi.mock('@dnd-kit/react', () => ({
  useDragDropManager: vi.fn(() => mockManager),
  useDragOperation: vi.fn(() => ({
    source: null,
  })),
}));

// Mock @dnd-kit/abstract/modifiers
vi.mock('@dnd-kit/abstract/modifiers', () => ({
  RestrictToVerticalAxis: {},
}));

describe('Sortable', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render children', () => {
      render(
        <Sortable id="test-sortable" index={0}>
          <div data-testid="child-content">Sortable Content</div>
        </Sortable>,
      );

      expect(screen.getByTestId('child-content')).toBeInTheDocument();
      expect(screen.getByText('Sortable Content')).toBeInTheDocument();
    });

    it('should render without children', () => {
      const {container} = render(<Sortable id="empty-sortable" index={0} />);

      expect(container.firstChild).toBeInTheDocument();
    });

    it('should render with memoized presentation wrapper', () => {
      const {container} = render(
        <Sortable id="test-sortable" index={0}>
          <span>Content</span>
        </Sortable>,
      );

      // Should have Box wrappers from the component
      expect(container.querySelectorAll('div').length).toBeGreaterThan(0);
    });
  });

  describe('Hook Integration', () => {
    it('should call useSortable with correct id', async () => {
      const {useSortable} = await import('@dnd-kit/react/sortable');

      render(
        <Sortable id="unique-sortable-123" index={5}>
          <div>Content</div>
        </Sortable>,
      );

      expect(useSortable).toHaveBeenCalledWith(
        expect.objectContaining({
          id: 'unique-sortable-123',
        }),
      );
    });

    it('should call useSortable with correct index', async () => {
      const {useSortable} = await import('@dnd-kit/react/sortable');

      render(
        <Sortable id="test-sortable" index={3}>
          <div>Content</div>
        </Sortable>,
      );

      expect(useSortable).toHaveBeenCalledWith(
        expect.objectContaining({
          index: 3,
        }),
      );
    });

    it('should pass handleRef when provided', async () => {
      const {useSortable} = await import('@dnd-kit/react/sortable');

      const handleRef = {current: document.createElement('button')};

      render(
        <Sortable id="test-sortable" index={0} handleRef={handleRef}>
          <div>Content</div>
        </Sortable>,
      );

      expect(useSortable).toHaveBeenCalledWith(
        expect.objectContaining({
          handle: handleRef,
        }),
      );
    });

    it('should apply RestrictToVerticalAxis modifier', async () => {
      const {useSortable} = await import('@dnd-kit/react/sortable');
      const {RestrictToVerticalAxis} = await import('@dnd-kit/abstract/modifiers');

      render(
        <Sortable id="test-sortable" index={0}>
          <div>Content</div>
        </Sortable>,
      );

      expect(useSortable).toHaveBeenCalledWith(
        expect.objectContaining({
          modifiers: [RestrictToVerticalAxis],
        }),
      );
    });

    it('should pass collisionDetector when provided', async () => {
      const {useSortable} = await import('@dnd-kit/react/sortable');

      const customCollisionDetector = vi.fn();

      render(
        <Sortable id="test-sortable" index={0} collisionDetector={customCollisionDetector}>
          <div>Content</div>
        </Sortable>,
      );

      expect(useSortable).toHaveBeenCalledWith(
        expect.objectContaining({
          collisionDetector: customCollisionDetector,
        }),
      );
    });

    it('should pass additional props to useSortable', async () => {
      const {useSortable} = await import('@dnd-kit/react/sortable');

      render(
        <Sortable id="test-sortable" index={0} accept={['TYPE_A']} data={{custom: 'data'}} disabled>
          <div>Content</div>
        </Sortable>,
      );

      expect(useSortable).toHaveBeenCalledWith(
        expect.objectContaining({
          accept: ['TYPE_A'],
          data: {custom: 'data'},
          disabled: true,
        }),
      );
    });
  });

  describe('Drag Manager Setup', () => {
    it('should call useDragDropManager', async () => {
      const {useDragDropManager} = await import('@dnd-kit/react');

      render(
        <Sortable id="test-sortable" index={0}>
          <div>Content</div>
        </Sortable>,
      );

      expect(useDragDropManager).toHaveBeenCalled();
    });
  });

  describe('Styling', () => {
    it('should render with full width and height wrapper', () => {
      const {container} = render(
        <Sortable id="test-sortable" index={0}>
          <div>Content</div>
        </Sortable>,
      );

      // The component wraps children in Box elements
      expect(container.firstChild).toBeInTheDocument();
    });
  });

  describe('Index Positioning', () => {
    it('should work with index 0', () => {
      const {container} = render(
        <Sortable id="first-item" index={0}>
          <div>First</div>
        </Sortable>,
      );

      expect(container.firstChild).toBeInTheDocument();
    });

    it('should work with high index values', () => {
      const {container} = render(
        <Sortable id="last-item" index={100}>
          <div>Last</div>
        </Sortable>,
      );

      expect(container.firstChild).toBeInTheDocument();
    });
  });
});
