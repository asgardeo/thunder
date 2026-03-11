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

import {describe, it, expect, beforeEach, vi} from 'vitest';
import {renderHook} from '@thunder/test-utils/browser';
import {
  updateLastInteractedStore,
  useLastInteractedResourceOnly,
  updatePropertiesPanelStore,
  usePropertiesPanelStateOnly,
} from '../flowBuilderCoreStore';
import type {Resource} from '../../models/resources';

describe('flowBuilderCoreStore', () => {
  describe('LastInteractedResource Store', () => {
    beforeEach(() => {
      // Reset the store to initial state
      updateLastInteractedStore(null, '');
    });

    it('should return initial state with null resource and empty stepId', async () => {
      const {result} = await renderHook(() => useLastInteractedResourceOnly());

      expect(result.current.resource).toBeNull();
      expect(result.current.stepId).toBe('');
    });

    it('should update store when resource changes', async () => {
      const {result} = await renderHook(() => useLastInteractedResourceOnly());

      const mockResource: Resource = {
        id: 'resource-1',
        type: 'VIEW',
        resourceType: 'STEP',
      } as Resource;

      updateLastInteractedStore(mockResource, 'step-1');

      await vi.waitFor(() => {
        expect(result.current.resource).toEqual(mockResource);
        expect(result.current.stepId).toBe('step-1');
      });
    });

    it('should update store when stepId changes', async () => {
      const {result} = await renderHook(() => useLastInteractedResourceOnly());

      const mockResource: Resource = {
        id: 'resource-1',
        type: 'VIEW',
        resourceType: 'STEP',
      } as Resource;

      updateLastInteractedStore(mockResource, 'step-1');

      await vi.waitFor(() => {
        expect(result.current.stepId).toBe('step-1');
      });

      updateLastInteractedStore(mockResource, 'step-2');

      await vi.waitFor(() => {
        expect(result.current.stepId).toBe('step-2');
      });
    });

    it('should not notify listeners when values do not change', async () => {
      const {result} = await renderHook(() => useLastInteractedResourceOnly());

      const mockResource: Resource = {
        id: 'resource-1',
        type: 'VIEW',
        resourceType: 'STEP',
      } as Resource;

      updateLastInteractedStore(mockResource, 'step-1');

      await vi.waitFor(() => {
        expect(result.current.resource).toEqual(mockResource);
      });

      const initialResource = result.current.resource;
      const initialStepId = result.current.stepId;

      // Update with same values - should not trigger re-render
      updateLastInteractedStore(mockResource, 'step-1');

      expect(result.current.resource).toBe(initialResource);
      expect(result.current.stepId).toBe(initialStepId);
    });

    it('should handle setting resource back to null', async () => {
      const {result} = await renderHook(() => useLastInteractedResourceOnly());

      const mockResource: Resource = {
        id: 'resource-1',
        type: 'VIEW',
        resourceType: 'STEP',
      } as Resource;

      updateLastInteractedStore(mockResource, 'step-1');

      await vi.waitFor(() => {
        expect(result.current.resource).toEqual(mockResource);
      });

      updateLastInteractedStore(null, '');

      await vi.waitFor(() => {
        expect(result.current.resource).toBeNull();
        expect(result.current.stepId).toBe('');
      });
    });

    it('should handle multiple subscribers', async () => {
      const {result: result1} = await renderHook(() => useLastInteractedResourceOnly());
      const {result: result2} = await renderHook(() => useLastInteractedResourceOnly());

      const mockResource: Resource = {
        id: 'resource-1',
        type: 'VIEW',
        resourceType: 'STEP',
      } as Resource;

      updateLastInteractedStore(mockResource, 'step-1');

      await vi.waitFor(() => {
        expect(result1.current.resource).toEqual(mockResource);
        expect(result2.current.resource).toEqual(mockResource);
      });
    });

    it('should clean up listener on unmount', async () => {
      const {unmount} = await renderHook(() => useLastInteractedResourceOnly());

      // Unmount should not throw
      expect(() => unmount()).not.toThrow();
    });
  });

  describe('PropertiesPanel Store', () => {
    beforeEach(() => {
      // Reset the store to initial state
      updatePropertiesPanelStore(false);
    });

    it('should return initial state with isOpen false', async () => {
      const {result} = await renderHook(() => usePropertiesPanelStateOnly());

      expect(result.current.isOpen).toBe(false);
    });

    it('should update store when isOpen changes to true', async () => {
      const {result} = await renderHook(() => usePropertiesPanelStateOnly());

      updatePropertiesPanelStore(true);

      await vi.waitFor(() => {
        expect(result.current.isOpen).toBe(true);
      });
    });

    it('should update store when isOpen changes to false', async () => {
      const {result} = await renderHook(() => usePropertiesPanelStateOnly());

      updatePropertiesPanelStore(true);

      await vi.waitFor(() => {
        expect(result.current.isOpen).toBe(true);
      });

      updatePropertiesPanelStore(false);

      await vi.waitFor(() => {
        expect(result.current.isOpen).toBe(false);
      });
    });

    it('should not notify listeners when value does not change', async () => {
      const {result} = await renderHook(() => usePropertiesPanelStateOnly());

      updatePropertiesPanelStore(true);

      await vi.waitFor(() => {
        expect(result.current.isOpen).toBe(true);
      });

      const initialIsOpen = result.current.isOpen;

      // Update with same value - should not trigger unnecessary updates
      updatePropertiesPanelStore(true);

      expect(result.current.isOpen).toBe(initialIsOpen);
    });

    it('should handle multiple subscribers', async () => {
      const {result: result1} = await renderHook(() => usePropertiesPanelStateOnly());
      const {result: result2} = await renderHook(() => usePropertiesPanelStateOnly());

      updatePropertiesPanelStore(true);

      await vi.waitFor(() => {
        expect(result1.current.isOpen).toBe(true);
        expect(result2.current.isOpen).toBe(true);
      });
    });

    it('should clean up listener on unmount', async () => {
      const {unmount} = await renderHook(() => usePropertiesPanelStateOnly());

      // Unmount should not throw
      expect(() => unmount()).not.toThrow();
    });

    it('should toggle isOpen state correctly', async () => {
      const {result} = await renderHook(() => usePropertiesPanelStateOnly());

      expect(result.current.isOpen).toBe(false);

      updatePropertiesPanelStore(true);
      await vi.waitFor(() => {
        expect(result.current.isOpen).toBe(true);
      });

      updatePropertiesPanelStore(false);
      await vi.waitFor(() => {
        expect(result.current.isOpen).toBe(false);
      });

      updatePropertiesPanelStore(true);
      await vi.waitFor(() => {
        expect(result.current.isOpen).toBe(true);
      });
    });
  });

  describe('Store Isolation', () => {
    beforeEach(() => {
      updateLastInteractedStore(null, '');
      updatePropertiesPanelStore(false);
    });

    it('should have independent stores that do not affect each other', async () => {
      const {result: lastInteracted} = await renderHook(() => useLastInteractedResourceOnly());
      const {result: propertiesPanel} = await renderHook(() => usePropertiesPanelStateOnly());

      const mockResource: Resource = {
        id: 'resource-1',
        type: 'VIEW',
        resourceType: 'STEP',
      } as Resource;

      // Update last interacted store
      updateLastInteractedStore(mockResource, 'step-1');

      await vi.waitFor(() => {
        expect(lastInteracted.current.resource).toEqual(mockResource);
      });

      // Properties panel should remain unchanged
      expect(propertiesPanel.current.isOpen).toBe(false);

      // Update properties panel store
      updatePropertiesPanelStore(true);

      await vi.waitFor(() => {
        expect(propertiesPanel.current.isOpen).toBe(true);
      });

      // Last interacted should remain unchanged
      expect(lastInteracted.current.resource).toEqual(mockResource);
    });
  });
});
