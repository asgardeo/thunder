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

import {useCallback, useState} from 'react';

/**
 * Return type for the useMultiSelect hook.
 */
export interface UseMultiSelectReturn {
  /** Set of selected item IDs */
  selectedIds: Set<string>;
  /** Check if an item is selected */
  isSelected: (id: string) => boolean;
  /** Toggle selection for a single item */
  toggleSelection: (id: string) => void;
  /** Select all items */
  selectAll: (ids: string[]) => void;
  /** Clear all selections */
  clearSelection: () => void;
  /** Get count of selected items */
  selectedCount: number;
}

/**
 * Hook to manage multi-select state for lists with support for shift-click range selection.
 *
 * @returns Multi-select state and handlers
 *
 * @example
 * ```tsx
 * const { selectedIds, isSelected, toggleSelection, selectAll, clearSelection } = useMultiSelect();
 *
 * // Check if item is selected
 * {isSelected(item.id) && <CheckIcon />}
 *
 * // Toggle selection on click
 * <Checkbox
 *   checked={isSelected(item.id)}
 *   onChange={() => toggleSelection(item.id)}
 * />
 *
 * // Select all
 * <Button onClick={() => selectAll(items.map(i => i.id))}>
 *   Select All
 * </Button>
 * ```
 */
export default function useMultiSelect(): UseMultiSelectReturn {
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());

  const isSelected = useCallback(
    (id: string): boolean => {
      return selectedIds.has(id);
    },
    [selectedIds]
  );

  const toggleSelection = useCallback((id: string) => {
    setSelectedIds((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(id)) {
        newSet.delete(id);
      } else {
        newSet.add(id);
      }
      return newSet;
    });
  }, []);

  const selectAll = useCallback((ids: string[]) => {
    setSelectedIds(new Set(ids));
  }, []);

  const clearSelection = useCallback(() => {
    setSelectedIds(new Set());
  }, []);

  return {
    selectedIds,
    isSelected,
    toggleSelection,
    selectAll,
    clearSelection,
    selectedCount: selectedIds.size,
  };
}
