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
import type {RangeSelection, TextNode, ElementNode} from 'lexical';
import {$isAtNodeEnd} from '@lexical/selection';
import getSelectedNode from '../getSelectedNode';

// Import the mocked function to control its behavior in tests

// Mock @lexical/selection
vi.mock('@lexical/selection', () => ({
  $isAtNodeEnd: vi.fn(),
}));

describe('getSelectedNode', () => {
  // Helper to create mock nodes
  const createMockNode = (id: string): TextNode | ElementNode =>
    ({
      __key: id,
      getKey: () => id,
    }) as unknown as TextNode | ElementNode;

  // Helper to create mock selection
  const createMockSelection = (
    anchorNode: TextNode | ElementNode,
    focusNode: TextNode | ElementNode,
    isBackward = false,
  ): RangeSelection =>
    ({
      anchor: {
        getNode: () => anchorNode,
      },
      focus: {
        getNode: () => focusNode,
      },
      isBackward: () => isBackward,
    }) as unknown as RangeSelection;

  describe('when anchor and focus nodes are the same', () => {
    it('should return the anchor node', () => {
      const node = createMockNode('node-1');
      const selection = createMockSelection(node, node);

      const result = getSelectedNode(selection);

      expect(result).toBe(node);
    });

    it('should return the same node regardless of isBackward value', () => {
      const node = createMockNode('node-1');
      const selectionForward = createMockSelection(node, node, false);
      const selectionBackward = createMockSelection(node, node, true);

      expect(getSelectedNode(selectionForward)).toBe(node);
      expect(getSelectedNode(selectionBackward)).toBe(node);
    });
  });

  describe('when anchor and focus nodes are different', () => {
    describe('forward selection (isBackward = false)', () => {
      it('should return focusNode when at node end', () => {
        const anchorNode = createMockNode('anchor');
        const focusNode = createMockNode('focus');
        const selection = createMockSelection(anchorNode, focusNode, false);

        vi.mocked($isAtNodeEnd).mockReturnValue(true);

        const result = getSelectedNode(selection);

        expect(result).toBe(focusNode);
      });

      it('should return anchorNode when not at node end', () => {
        const anchorNode = createMockNode('anchor');
        const focusNode = createMockNode('focus');
        const selection = createMockSelection(anchorNode, focusNode, false);

        vi.mocked($isAtNodeEnd).mockReturnValue(false);

        const result = getSelectedNode(selection);

        expect(result).toBe(anchorNode);
      });
    });

    describe('backward selection (isBackward = true)', () => {
      it('should return anchorNode when at node end', () => {
        const anchorNode = createMockNode('anchor');
        const focusNode = createMockNode('focus');
        const selection = createMockSelection(anchorNode, focusNode, true);

        vi.mocked($isAtNodeEnd).mockReturnValue(true);

        const result = getSelectedNode(selection);

        expect(result).toBe(anchorNode);
      });

      it('should return focusNode when not at node end', () => {
        const anchorNode = createMockNode('anchor');
        const focusNode = createMockNode('focus');
        const selection = createMockSelection(anchorNode, focusNode, true);

        vi.mocked($isAtNodeEnd).mockReturnValue(false);

        const result = getSelectedNode(selection);

        expect(result).toBe(focusNode);
      });
    });
  });

  describe('$isAtNodeEnd integration', () => {
    it('should call $isAtNodeEnd with the focus point', () => {
      const anchorNode = createMockNode('anchor');
      const focusNode = createMockNode('focus');
      const selection = createMockSelection(anchorNode, focusNode, false);

      vi.mocked($isAtNodeEnd).mockReturnValue(false);

      getSelectedNode(selection);

      expect($isAtNodeEnd).toHaveBeenCalledWith(selection.focus);
    });

    it('should not call $isAtNodeEnd when anchor equals focus', () => {
      const node = createMockNode('same-node');
      const selection = createMockSelection(node, node);

      vi.mocked($isAtNodeEnd).mockClear();

      getSelectedNode(selection);

      expect($isAtNodeEnd).not.toHaveBeenCalled();
    });
  });
});
