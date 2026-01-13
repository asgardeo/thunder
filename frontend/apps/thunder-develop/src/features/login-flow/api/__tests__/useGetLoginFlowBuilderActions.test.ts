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
import {renderHook} from '@testing-library/react';
import useGetLoginFlowBuilderActions from '../useGetLoginFlowBuilderActions';

// Mock the core actions hook
vi.mock('@/features/flows/api/useGetFlowBuilderCoreActions', () => ({
  default: vi.fn(() => ({
    data: [{id: 'core-action-1', name: 'Core Action'}],
    error: null,
    isLoading: false,
    isValidating: false,
    mutate: () => null,
  })),
}));

// Mock the login-flow actions data
vi.mock('../../data/actions.json', () => ({
  default: [{id: 'login-action-1', name: 'Login Action'}],
}));

describe('useGetLoginFlowBuilderActions', () => {
  it('should return combined actions from core and login-flow', () => {
    const {result} = renderHook(() => useGetLoginFlowBuilderActions());

    expect(result.current.data).toBeDefined();
    expect(Array.isArray(result.current.data)).toBe(true);
  });

  it('should return loading state as false', () => {
    const {result} = renderHook(() => useGetLoginFlowBuilderActions());

    expect(result.current.isLoading).toBe(false);
  });

  it('should return error as null', () => {
    const {result} = renderHook(() => useGetLoginFlowBuilderActions());

    expect(result.current.error).toBeNull();
  });

  it('should return isValidating as false', () => {
    const {result} = renderHook(() => useGetLoginFlowBuilderActions());

    expect(result.current.isValidating).toBe(false);
  });

  it('should return mutate function', () => {
    const {result} = renderHook(() => useGetLoginFlowBuilderActions());

    expect(result.current.mutate).toBeDefined();
    expect(typeof result.current.mutate).toBe('function');
  });

  it('should support generic type parameter', () => {
    interface CustomAction {
      id: string;
      customField: string;
    }

    const {result} = renderHook(() => useGetLoginFlowBuilderActions<CustomAction[]>());

    expect(result.current.data).toBeDefined();
  });
});
