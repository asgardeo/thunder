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

import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest';
import {waitFor, renderHook} from '@thunder/test-utils';
import useDeleteUser from '../useDeleteUser';

// Mock useAsgardeo
const mockHttpRequest = vi.fn();
vi.mock('@asgardeo/react', () => ({
  useAsgardeo: () => ({
    http: {
      request: mockHttpRequest,
    },
  }),
}));

// Mock useConfig
vi.mock('@thunder/shared-contexts', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@thunder/shared-contexts')>();
  return {
    ...actual,
    useConfig: () => ({
      getServerUrl: () => 'https://localhost:8090',
    }),
  };
});

describe('useDeleteUser', () => {
  beforeEach(() => {
    mockHttpRequest.mockReset();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should initialize with correct default values', () => {
    const {result} = renderHook(() => useDeleteUser());

    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeNull();
    expect(typeof result.current.deleteUser).toBe('function');
  });

  it('should delete a user successfully', async () => {
    mockHttpRequest.mockResolvedValueOnce({data: null});

    const {result} = renderHook(() => useDeleteUser());

    const deleteResult = await result.current.deleteUser('user-123');

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(deleteResult).toBe(true);
    expect(result.current.error).toBeNull();
    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/users/user-123',
        method: 'DELETE',
      }),
    );
  });

  it('should handle API error with JSON response', async () => {
    mockHttpRequest.mockRejectedValueOnce(new Error('User not found'));

    const {result} = renderHook(() => useDeleteUser());

    try {
      await result.current.deleteUser('user-123');
    } catch {
      // Expected to throw
    }

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toEqual({
        code: 'DELETE_USER_ERROR',
        message: 'User not found',
        description: 'Failed to delete user',
      });
    });
  });

  it('should handle API error without JSON response', async () => {
    mockHttpRequest.mockRejectedValueOnce(new Error('Internal Server Error'));

    const {result} = renderHook(() => useDeleteUser());

    try {
      await result.current.deleteUser('user-123');
    } catch {
      // Expected to throw
    }

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toEqual({
        code: 'DELETE_USER_ERROR',
        message: 'Internal Server Error',
        description: 'Failed to delete user',
      });
    });
  });

  it('should handle network error', async () => {
    mockHttpRequest.mockRejectedValueOnce(new Error('Network error'));

    const {result} = renderHook(() => useDeleteUser());

    try {
      await result.current.deleteUser('user-123');
    } catch {
      // Expected to throw
    }

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toEqual({
        code: 'DELETE_USER_ERROR',
        message: 'Network error',
        description: 'Failed to delete user',
      });
    });
  });

  it('should set loading state correctly during request', async () => {
    let resolveRequest: () => void;
    const requestPromise = new Promise<void>((resolve) => {
      resolveRequest = resolve;
    });

    mockHttpRequest.mockImplementationOnce(() => requestPromise.then(() => ({data: null})));

    const {result} = renderHook(() => useDeleteUser());

    const promise = result.current.deleteUser('user-123');

    await waitFor(() => {
      expect(result.current.loading).toBe(true);
    });

    resolveRequest!();
    await promise;

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
  });

  it('should clear previous error on new delete attempt', async () => {
    mockHttpRequest.mockRejectedValueOnce(new Error('User not found')).mockResolvedValueOnce({data: null});

    const {result} = renderHook(() => useDeleteUser());

    try {
      await result.current.deleteUser('user-123');
    } catch {
      // Expected to throw
    }

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'DELETE_USER_ERROR',
        message: 'User not found',
        description: 'Failed to delete user',
      });
    });

    await result.current.deleteUser('user-456');

    await waitFor(() => {
      expect(result.current.error).toBeNull();
    });
  });

  it('should handle multiple delete operations', async () => {
    mockHttpRequest.mockResolvedValueOnce({data: null}).mockResolvedValueOnce({data: null});

    const {result} = renderHook(() => useDeleteUser());

    const deleteResult1 = await result.current.deleteUser('user-123');
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
    expect(deleteResult1).toBe(true);

    const deleteResult2 = await result.current.deleteUser('user-456');
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
    expect(deleteResult2).toBe(true);

    expect(mockHttpRequest).toHaveBeenCalledTimes(2);
  });

  it('should reset error state when reset is called', async () => {
    mockHttpRequest.mockRejectedValueOnce(new Error('User not found'));

    const {result} = renderHook(() => useDeleteUser());

    try {
      await result.current.deleteUser('user-123');
    } catch {
      // Expected to throw
    }

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'DELETE_USER_ERROR',
        message: 'User not found',
        description: 'Failed to delete user',
      });
    });

    result.current.reset?.();

    await waitFor(() => {
      expect(result.current.error).toBeNull();
    });
  });

  it('should handle non-Error rejection', async () => {
    mockHttpRequest.mockRejectedValueOnce('String error');

    const {result} = renderHook(() => useDeleteUser());

    await expect(result.current.deleteUser('user-123')).rejects.toBe('String error');

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'DELETE_USER_ERROR',
        message: 'An unknown error occurred',
        description: 'Failed to delete user',
      });
      expect(result.current.loading).toBe(false);
    });
  });
});
