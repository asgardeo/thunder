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

import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest';
import {waitFor, renderHook} from '@thunder/test-utils';
import useDeleteUserType from '../useDeleteUserType';

const mockHttpRequest = vi.fn();
vi.mock('@asgardeo/react', () => ({
  useAsgardeo: () => ({
    http: {
      request: mockHttpRequest,
    },
  }),
}));

// Mock useConfig
const mockGetServerUrl = vi.fn<() => string | undefined>(() => 'https://localhost:8090');
vi.mock('@thunder/shared-contexts', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@thunder/shared-contexts')>();
  return {
    ...actual,
    useConfig: () => ({
      getServerUrl: mockGetServerUrl,
    }),
  };
});

describe('useDeleteUserType', () => {
  const mockUserTypeId = '123';

  beforeEach(() => {
    mockHttpRequest.mockReset();
    mockGetServerUrl.mockReturnValue('https://localhost:8090');
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should initialize with default state', () => {
    const {result} = renderHook(() => useDeleteUserType());

    expect(result.current.error).toBeNull();
    expect(result.current.loading).toBe(false);
    expect(typeof result.current.deleteUserType).toBe('function');
    expect(typeof result.current.reset).toBe('function');
  });

  it('should successfully delete a user type', async () => {
    mockHttpRequest.mockResolvedValueOnce({data: null});

    const {result} = renderHook(() => useDeleteUserType());

    const deleteResult = await result.current.deleteUserType(mockUserTypeId);

    expect(deleteResult).toBe(true);

    await waitFor(() => {
      expect(result.current.error).toBeNull();
      expect(result.current.loading).toBe(false);
    });

    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: `https://localhost:8090/user-schemas/${mockUserTypeId}`,
        method: 'DELETE',
      }),
    );
  });

  it('should set loading state during deletion', async () => {
    // Create a promise we can control
    let resolveRequest: (value: {data: null}) => void;
    const requestPromise = new Promise<{data: null}>((resolve) => {
      resolveRequest = resolve;
    });

    mockHttpRequest.mockReturnValueOnce(requestPromise);

    const {result} = renderHook(() => useDeleteUserType());

    const promise = result.current.deleteUserType(mockUserTypeId);

    await waitFor(() => {
      expect(result.current.loading).toBe(true);
    });

    // Now resolve the request
    resolveRequest!({data: null});

    await promise;

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
  });

  it('should handle API error', async () => {
    mockHttpRequest.mockRejectedValueOnce(new Error('User type not found'));

    const {result} = renderHook(() => useDeleteUserType());

    await expect(result.current.deleteUserType(mockUserTypeId)).rejects.toThrow('User type not found');

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'DELETE_USER_TYPE_ERROR',
        message: 'User type not found',
        description: 'Failed to delete user type',
      });
      expect(result.current.loading).toBe(false);
    });
  });

  it('should handle network error', async () => {
    mockHttpRequest.mockRejectedValueOnce(new Error('Network error'));

    const {result} = renderHook(() => useDeleteUserType());

    await expect(result.current.deleteUserType(mockUserTypeId)).rejects.toThrow('Network error');

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'DELETE_USER_TYPE_ERROR',
        message: 'Network error',
        description: 'Failed to delete user type',
      });
      expect(result.current.loading).toBe(false);
    });
  });

  it('should reset error state when reset is called', async () => {
    mockHttpRequest.mockRejectedValueOnce(new Error('User type not found'));

    const {result} = renderHook(() => useDeleteUserType());

    await expect(result.current.deleteUserType(mockUserTypeId)).rejects.toThrow('User type not found');

    await waitFor(() => {
      expect(result.current.error).not.toBeNull();
    });

    result.current.reset();

    await waitFor(() => {
      expect(result.current.error).toBeNull();
    });
  });

  it('should clear previous error when starting new request', async () => {
    mockHttpRequest.mockRejectedValueOnce(new Error('Previous error')).mockResolvedValueOnce({data: null});

    const {result} = renderHook(() => useDeleteUserType());

    await expect(result.current.deleteUserType(mockUserTypeId)).rejects.toThrow('Previous error');

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'DELETE_USER_TYPE_ERROR',
        message: 'Previous error',
        description: 'Failed to delete user type',
      });
    });

    await result.current.deleteUserType(mockUserTypeId);

    await waitFor(() => {
      expect(result.current.error).toBeNull();
    });
  });

  it('should handle non-Error rejection', async () => {
    mockHttpRequest.mockRejectedValueOnce('String error');

    const {result} = renderHook(() => useDeleteUserType());

    await expect(result.current.deleteUserType(mockUserTypeId)).rejects.toBe('String error');

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'DELETE_USER_TYPE_ERROR',
        message: 'An unknown error occurred',
        description: 'Failed to delete user type',
      });
      expect(result.current.loading).toBe(false);
    });
  });

  it('should fallback to env variable when getServerUrl returns undefined', async () => {
    mockGetServerUrl.mockReturnValue(undefined);
    mockHttpRequest.mockResolvedValueOnce({data: null});

    const {result} = renderHook(() => useDeleteUserType());

    const deleteResult = await result.current.deleteUserType(mockUserTypeId);

    expect(deleteResult).toBe(true);

    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: expect.stringContaining('/user-schemas/') as string,
        method: 'DELETE',
      }),
    );
  });
});
