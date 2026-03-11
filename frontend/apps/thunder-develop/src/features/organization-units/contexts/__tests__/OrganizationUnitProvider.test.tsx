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

import {describe, it, expect} from 'vitest';
import {renderHook} from '@thunder/test-utils/browser';
import {useContext} from 'react';
import type {ReactNode} from 'react';
import OrganizationUnitProvider from '../OrganizationUnitProvider';
import OrganizationUnitContext from '../OrganizationUnitContext';

function useOrganizationUnitContext() {
  const context = useContext(OrganizationUnitContext);
  if (!context) {
    throw new Error('useOrganizationUnitContext must be used within OrganizationUnitProvider');
  }
  return context;
}

describe('OrganizationUnitProvider', () => {
  it('should provide default context values', async () => {
    const {result} = await renderHook(() => useOrganizationUnitContext(), {
      wrapper: ({children}: {children: ReactNode}) => (
        <OrganizationUnitProvider>{children}</OrganizationUnitProvider>
      ),
    });

    expect(result.current.treeItems).toEqual([]);
    expect(result.current.expandedItems).toEqual([]);
    expect(result.current.loadedItems).toEqual(new Set());
    expect(typeof result.current.setTreeItems).toBe('function');
    expect(typeof result.current.setExpandedItems).toBe('function');
    expect(typeof result.current.setLoadedItems).toBe('function');
    expect(typeof result.current.resetTreeState).toBe('function');
  });

  it('should allow setting treeItems', async () => {
    const {result} = await renderHook(() => useOrganizationUnitContext(), {
      wrapper: ({children}: {children: ReactNode}) => (
        <OrganizationUnitProvider>{children}</OrganizationUnitProvider>
      ),
    });

    const items = [{id: 'ou-1', label: 'Test OU', handle: 'test'}];

    result.current.setTreeItems(items);

    await vi.waitFor(() => {
      expect(result.current.treeItems).toEqual(items);
    });
  });

  it('should allow setting expandedItems', async () => {
    const {result} = await renderHook(() => useOrganizationUnitContext(), {
      wrapper: ({children}: {children: ReactNode}) => (
        <OrganizationUnitProvider>{children}</OrganizationUnitProvider>
      ),
    });

    result.current.setExpandedItems(['ou-1', 'ou-2']);

    await vi.waitFor(() => {
      expect(result.current.expandedItems).toEqual(['ou-1', 'ou-2']);
    });
  });

  it('should allow setting loadedItems', async () => {
    const {result} = await renderHook(() => useOrganizationUnitContext(), {
      wrapper: ({children}: {children: ReactNode}) => (
        <OrganizationUnitProvider>{children}</OrganizationUnitProvider>
      ),
    });

    result.current.setLoadedItems(new Set(['ou-1']));

    await vi.waitFor(() => {
      expect(result.current.loadedItems).toEqual(new Set(['ou-1']));
    });
  });

  it('should reset treeItems and loadedItems but preserve expandedItems on resetTreeState', async () => {
    const {result} = await renderHook(() => useOrganizationUnitContext(), {
      wrapper: ({children}: {children: ReactNode}) => (
        <OrganizationUnitProvider>{children}</OrganizationUnitProvider>
      ),
    });

    // Set some state first
    result.current.setTreeItems([{id: 'ou-1', label: 'Test', handle: 'test'}]);
    result.current.setExpandedItems(['ou-1']);
    result.current.setLoadedItems(new Set(['ou-1']));

    await vi.waitFor(() => {
      expect(result.current.treeItems).toHaveLength(1);
      expect(result.current.expandedItems).toEqual(['ou-1']);
      expect(result.current.loadedItems.size).toBe(1);
    });

    // Reset tree state
    result.current.resetTreeState();

    await vi.waitFor(() => {
      expect(result.current.treeItems).toEqual([]);
      expect(result.current.loadedItems).toEqual(new Set());
      // expandedItems should be preserved
      expect(result.current.expandedItems).toEqual(['ou-1']);
    });
  });
});
