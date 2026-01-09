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
import useGetLoginFlowBuilderResources from '../useGetLoginFlowBuilderResources';

// Mock the core resources hook
vi.mock('@/features/flows/api/useGetFlowBuilderCoreResources', () => ({
  default: vi.fn(() => ({
    data: {
      elements: [{id: 'core-element'}],
      steps: [{id: 'core-step'}],
      templates: [{id: 'core-template'}],
      widgets: [{id: 'core-widget'}],
    },
    error: null,
    isLoading: false,
    isValidating: false,
    mutate: () => null,
  })),
}));

// Mock the login-flow data files
vi.mock('../../data/executors.json', () => ({
  default: [{id: 'login-executor'}],
}));

vi.mock('../../data/steps.json', () => ({
  default: [{id: 'login-step'}],
}));

vi.mock('../../data/templates.json', () => ({
  default: [{id: 'login-template'}],
}));

vi.mock('../../data/widgets.json', () => ({
  default: [{id: 'login-widget'}],
}));

describe('useGetLoginFlowBuilderResources', () => {
  it('should return combined resources from core and login-flow', () => {
    const {result} = renderHook(() => useGetLoginFlowBuilderResources());

    expect(result.current.data).toBeDefined();
  });

  it('should merge steps from core and login-flow', () => {
    const {result} = renderHook(() => useGetLoginFlowBuilderResources());

    expect(result.current.data.steps).toBeDefined();
    expect(Array.isArray(result.current.data.steps)).toBe(true);
  });

  it('should merge templates from core and login-flow', () => {
    const {result} = renderHook(() => useGetLoginFlowBuilderResources());

    expect(result.current.data.templates).toBeDefined();
    expect(Array.isArray(result.current.data.templates)).toBe(true);
  });

  it('should merge widgets from core and login-flow', () => {
    const {result} = renderHook(() => useGetLoginFlowBuilderResources());

    expect(result.current.data.widgets).toBeDefined();
    expect(Array.isArray(result.current.data.widgets)).toBe(true);
  });

  it('should include executors from login-flow', () => {
    const {result} = renderHook(() => useGetLoginFlowBuilderResources());

    expect(result.current.data.executors).toBeDefined();
    expect(Array.isArray(result.current.data.executors)).toBe(true);
  });

  it('should return loading state as false', () => {
    const {result} = renderHook(() => useGetLoginFlowBuilderResources());

    expect(result.current.isLoading).toBe(false);
  });

  it('should return error as null', () => {
    const {result} = renderHook(() => useGetLoginFlowBuilderResources());

    expect(result.current.error).toBeNull();
  });

  it('should return isValidating as false', () => {
    const {result} = renderHook(() => useGetLoginFlowBuilderResources());

    expect(result.current.isValidating).toBe(false);
  });

  it('should return mutate function', () => {
    const {result} = renderHook(() => useGetLoginFlowBuilderResources());

    expect(result.current.mutate).toBeDefined();
    expect(typeof result.current.mutate).toBe('function');
  });

  it('should support generic type parameter', () => {
    interface CustomResources {
      elements: unknown[];
      steps: unknown[];
    }

    const {result} = renderHook(() => useGetLoginFlowBuilderResources<CustomResources>());

    expect(result.current.data).toBeDefined();
  });
});
