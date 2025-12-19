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
import {renderHook} from '@testing-library/react';
import useAuthenticationFlows from '../useAuthenticationFlows';

// Mock the useGetFlows hook
vi.mock('@/features/flows/api/useGetFlows');

const {default: useGetFlows} = await import('@/features/flows/api/useGetFlows');

describe('useAuthenticationFlows', () => {
  const mockFlows = [
    {
      id: 'flow-uuid-1',
      handle: 'default-basic-flow',
      name: 'Basic Flow',
      flowType: 'AUTHENTICATION',
      activeVersion: 1,
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2024-01-01T00:00:00Z',
    },
    {
      id: 'flow-uuid-2',
      handle: 'default-google-flow',
      name: 'Google Flow',
      flowType: 'AUTHENTICATION',
      activeVersion: 1,
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2024-01-01T00:00:00Z',
    },
    {
      id: 'flow-uuid-3',
      handle: 'default-github-flow',
      name: 'GitHub Flow',
      flowType: 'AUTHENTICATION',
      activeVersion: 1,
      createdAt: '2024-01-01T00:00:00Z',
      updatedAt: '2024-01-01T00:00:00Z',
    },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Flow fetching and mapping', () => {
    it('should return flows from useGetFlows', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {
          totalResults: 3,
          startIndex: 1,
          count: 3,
          flows: mockFlows,
        },
        isLoading: false,
        error: null,
      } as ReturnType<typeof useGetFlows>);

      const {result} = renderHook(() => useAuthenticationFlows());

      expect(result.current.flows).toHaveLength(3);
      expect(result.current.isLoading).toBe(false);
      expect(result.current.error).toBeNull();
    });

    it('should call useGetFlows with correct parameters', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: []},
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useGetFlows>);

      renderHook(() => useAuthenticationFlows());

      expect(useGetFlows).toHaveBeenCalledWith({
        flowType: 'AUTHENTICATION',
        limit: 100,
      });
    });

    it('should handle empty flows array', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: []},
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useGetFlows>);

      const {result} = renderHook(() => useAuthenticationFlows());

      expect(result.current.flows).toHaveLength(0);
      expect(result.current.error).toBeNull();
    });

    it('should handle undefined data', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: undefined,
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useGetFlows>);

      const {result} = renderHook(() => useAuthenticationFlows());

      expect(result.current.flows).toHaveLength(0);
    });
  });

  describe('getFlowIdByHandle', () => {
    it('should return flow UUID for valid handle', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockFlows},
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useGetFlows>);

      const {result} = renderHook(() => useAuthenticationFlows());

      expect(result.current.getFlowIdByHandle('default-basic-flow')).toBe('flow-uuid-1');
      expect(result.current.getFlowIdByHandle('default-google-flow')).toBe('flow-uuid-2');
      expect(result.current.getFlowIdByHandle('default-github-flow')).toBe('flow-uuid-3');
    });

    it('should return undefined for invalid handle', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockFlows},
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useGetFlows>);

      const {result} = renderHook(() => useAuthenticationFlows());

      expect(result.current.getFlowIdByHandle('non-existent-flow')).toBeUndefined();
    });

    it('should return undefined when flows are empty', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: []},
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useGetFlows>);

      const {result} = renderHook(() => useAuthenticationFlows());

      expect(result.current.getFlowIdByHandle('default-basic-flow')).toBeUndefined();
    });
  });

  describe('isFlowAvailable', () => {
    it('should return true for existing flow handles', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockFlows},
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useGetFlows>);

      const {result} = renderHook(() => useAuthenticationFlows());

      expect(result.current.isFlowAvailable('default-basic-flow')).toBe(true);
      expect(result.current.isFlowAvailable('default-google-flow')).toBe(true);
    });

    it('should return false for non-existing flow handles', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockFlows},
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useGetFlows>);

      const {result} = renderHook(() => useAuthenticationFlows());

      expect(result.current.isFlowAvailable('non-existent-flow')).toBe(false);
      expect(result.current.isFlowAvailable('another-missing-flow')).toBe(false);
    });

    it('should return false when flows are empty', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: []},
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useGetFlows>);

      const {result} = renderHook(() => useAuthenticationFlows());

      expect(result.current.isFlowAvailable('default-basic-flow')).toBe(false);
    });
  });

  describe('Loading and error states', () => {
    it('should reflect loading state', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: undefined,
        isLoading: true,
        error: null,
      } as unknown as ReturnType<typeof useGetFlows>);

      const {result} = renderHook(() => useAuthenticationFlows());

      expect(result.current.isLoading).toBe(true);
      expect(result.current.flows).toEqual([]);
    });

    it('should reflect error state', () => {
      const testError = new Error('Network error');
      vi.mocked(useGetFlows).mockReturnValue({
        data: undefined,
        isLoading: false,
        error: testError,
      } as unknown as ReturnType<typeof useGetFlows>);

      const {result} = renderHook(() => useAuthenticationFlows());

      expect(result.current.error).toBe(testError);
    });

    it('should return null error when no error', () => {
      vi.mocked(useGetFlows).mockReturnValue({
        data: {flows: mockFlows},
        isLoading: false,
        error: null,
      } as unknown as ReturnType<typeof useGetFlows>);

      const {result} = renderHook(() => useAuthenticationFlows());

      expect(result.current.error).toBeNull();
    });
  });
});
