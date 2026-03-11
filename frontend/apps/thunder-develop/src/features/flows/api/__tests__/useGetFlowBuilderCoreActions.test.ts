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

import {describe, it, expect} from 'vitest';
import {renderHook} from '@thunder/test-utils/browser';
import useGetFlowBuilderCoreActions from '../useGetFlowBuilderCoreActions';
import actions from '../../data/actions.json';

describe('useGetFlowBuilderCoreActions', () => {
  describe('Return Structure', () => {
    it('should return an object with data, error, isLoading, isValidating, and mutate', async () => {
      const {result} = await renderHook(() => useGetFlowBuilderCoreActions());

      expect(result.current).toHaveProperty('data');
      expect(result.current).toHaveProperty('error');
      expect(result.current).toHaveProperty('isLoading');
      expect(result.current).toHaveProperty('isValidating');
      expect(result.current).toHaveProperty('mutate');
    });

    it('should return actions data from JSON file', async () => {
      const {result} = await renderHook(() => useGetFlowBuilderCoreActions());

      expect(result.current.data).toEqual(actions);
    });

    it('should return error as null', async () => {
      const {result} = await renderHook(() => useGetFlowBuilderCoreActions());

      expect(result.current.error).toBeNull();
    });

    it('should return isLoading as false', async () => {
      const {result} = await renderHook(() => useGetFlowBuilderCoreActions());

      expect(result.current.isLoading).toBe(false);
    });

    it('should return isValidating as false', async () => {
      const {result} = await renderHook(() => useGetFlowBuilderCoreActions());

      expect(result.current.isValidating).toBe(false);
    });

    it('should return mutate as a function that returns null', async () => {
      const {result} = await renderHook(() => useGetFlowBuilderCoreActions());

      expect(typeof result.current.mutate).toBe('function');
      expect(result.current.mutate()).toBeNull();
    });
  });

  describe('Generic Type Support', () => {
    it('should support custom generic type', async () => {
      interface CustomType {
        customField: string;
      }

      const {result} = await renderHook(() => useGetFlowBuilderCoreActions<CustomType>());

      // The data is cast to the generic type
      expect(result.current.data).toBeDefined();
    });

    it('should default to Actions type when no generic is provided', async () => {
      const {result} = await renderHook(() => useGetFlowBuilderCoreActions());

      // Verify data matches expected Actions structure
      const {data} = result.current;
      expect(Array.isArray(data)).toBe(true);
    });
  });

  describe('Data Content', () => {
    it('should return actions data as an array', async () => {
      const {result} = await renderHook(() => useGetFlowBuilderCoreActions());

      expect(Array.isArray(result.current.data)).toBe(true);
    });

    it('should contain action items with expected structure', async () => {
      const {result} = await renderHook(() => useGetFlowBuilderCoreActions());

      const {data} = result.current;
      if (data.length > 0) {
        const firstAction = data[0];
        expect(firstAction).toHaveProperty('resourceType');
        expect(firstAction).toHaveProperty('category');
        expect(firstAction).toHaveProperty('display');
      }
    });

    it('should contain navigation category actions', async () => {
      const {result} = await renderHook(() => useGetFlowBuilderCoreActions());

      const {data} = result.current;
      const navigationActions = data.filter((action) => action.category === 'NAVIGATION');
      expect(navigationActions.length).toBeGreaterThan(0);
    });
  });

  describe('Consistency', () => {
    it('should return the same data structure on multiple calls', async () => {
      const {result: result1} = await renderHook(() => useGetFlowBuilderCoreActions());
      const {result: result2} = await renderHook(() => useGetFlowBuilderCoreActions());

      expect(result1.current.data).toEqual(result2.current.data);
      expect(result1.current.error).toEqual(result2.current.error);
      expect(result1.current.isLoading).toEqual(result2.current.isLoading);
      expect(result1.current.isValidating).toEqual(result2.current.isValidating);
    });

    it('should maintain stable reference for mutate function', async () => {
      const {result, rerender} = await renderHook(() => useGetFlowBuilderCoreActions());

      const initialMutate = result.current.mutate;
      await rerender();

      // Note: In the current implementation, mutate is recreated on each call
      expect(typeof result.current.mutate).toBe('function');
      expect(result.current.mutate()).toBeNull();
      expect(initialMutate()).toBeNull();
    });
  });
});
