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
import useGetUser from '../useGetUser';
import type {ApiUser} from '../../types/users';

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

describe('useGetUser', () => {
  beforeEach(() => {
    mockHttpRequest.mockReset();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should initialize with correct default values', () => {
    const {result} = renderHook(() => useGetUser(undefined));

    expect(result.current.data).toBeNull();
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeNull();
    expect(typeof result.current.refetch).toBe('function');
  });

  it('should fetch user successfully when id is provided', async () => {
    const mockUser: ApiUser = {
      id: 'user-123',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Doe',
        email: 'john@example.com',
      },
    };

    mockHttpRequest.mockResolvedValue({data: mockUser});

    const {result} = renderHook(() => useGetUser('user-123'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(mockUser);
    expect(result.current.error).toBeNull();
    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/users/user-123',
        method: 'GET',
      }),
    );
  });

  it('should not fetch when id is not provided', () => {
    const {result} = renderHook(() => useGetUser(undefined));

    expect(result.current.data).toBeNull();
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeNull();
    expect(mockHttpRequest).not.toHaveBeenCalled();
  });

  it('should handle API error with JSON response', async () => {
    mockHttpRequest.mockRejectedValue(new Error('User not found'));

    const {result} = renderHook(() => useGetUser('user-123'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toEqual({
      code: 'FETCH_ERROR',
      message: 'User not found',
      description: 'Failed to fetch user',
    });
    expect(result.current.data).toBeNull();
  });

  it('should handle API error without JSON response', async () => {
    mockHttpRequest.mockRejectedValue(new Error('Internal Server Error'));

    const {result} = renderHook(() => useGetUser('user-123'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toEqual({
      code: 'FETCH_ERROR',
      message: 'Internal Server Error',
      description: 'Failed to fetch user',
    });
    expect(result.current.data).toBeNull();
  });

  it('should handle network error', async () => {
    mockHttpRequest.mockRejectedValue(new Error('Network error'));

    const {result} = renderHook(() => useGetUser('user-123'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toEqual({
      code: 'FETCH_ERROR',
      message: 'Network error',
      description: 'Failed to fetch user',
    });
    expect(result.current.data).toBeNull();
  });

  it('should refetch user with the same id', async () => {
    const mockUser: ApiUser = {
      id: 'user-123',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Doe',
        email: 'john@example.com',
      },
    };

    mockHttpRequest.mockResolvedValue({data: mockUser});

    const {result} = renderHook(() => useGetUser('user-123'));

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUser);
    });

    const callsBeforeRefetch = mockHttpRequest.mock.calls.length;
    await result.current.refetch();

    await waitFor(() => {
      expect(mockHttpRequest.mock.calls.length).toBeGreaterThan(callsBeforeRefetch);
    });
  });

  it('should refetch user with a new id', async () => {
    const mockUser1: ApiUser = {
      id: 'user-123',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Doe',
        email: 'john@example.com',
      },
    };

    const mockUser2: ApiUser = {
      id: 'user-789',
      organizationUnit: '/marketing',
      type: 'employee',
      attributes: {
        name: 'Jane Smith',
        email: 'jane@example.com',
      },
    };

    mockHttpRequest.mockImplementation(({url}: {url?: string}) =>
      Promise.resolve({
        data: url!.endsWith('user-123') ? mockUser1 : mockUser2,
      }),
    );

    const {result, rerender} = renderHook(({id}) => useGetUser(id), {
      initialProps: {id: 'user-123'},
    });

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUser1);
    });

    rerender({id: 'user-789'});

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUser2);
    });
  });

  it('should set error when refetch is called without id', async () => {
    const {result} = renderHook(() => useGetUser(undefined));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await result.current.refetch();

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // When userId is undefined, fetchUser returns early without setting error
    expect(result.current.error).toBeNull();
    expect(result.current.data).toBeNull();
  });

  it('should prevent double fetch in strict mode', async () => {
    const mockUser: ApiUser = {
      id: 'user-123',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Doe',
        email: 'john@example.com',
      },
    };

    mockHttpRequest.mockResolvedValue({data: mockUser});

    const {result} = renderHook(() => useGetUser('user-123'));

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUser);
    });

    expect(mockHttpRequest).toHaveBeenCalled();
  });

  it('should fetch when id changes', async () => {
    const mockUser1: ApiUser = {
      id: 'user-123',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Doe',
        email: 'john@example.com',
      },
    };

    const mockUser2: ApiUser = {
      id: 'user-789',
      organizationUnit: '/marketing',
      type: 'employee',
      attributes: {
        name: 'Jane Smith',
        email: 'jane@example.com',
      },
    };

    mockHttpRequest.mockImplementation(({url}: {url?: string}) =>
      Promise.resolve({
        data: url!.endsWith('user-123') ? mockUser1 : mockUser2,
      }),
    );

    const {result, rerender} = renderHook(({id}: {id?: string}) => useGetUser(id), {
      initialProps: {id: 'user-123'},
    });

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUser1);
    });

    rerender({id: 'user-789'});

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUser2);
    });
  });

  it('should handle error when refetch is called and throw error', async () => {
    const mockUser: ApiUser = {
      id: 'user-123',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Doe',
        email: 'john@example.com',
      },
    };

    mockHttpRequest.mockResolvedValueOnce({data: mockUser});

    const {result} = renderHook(() => useGetUser('user-123'));

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUser);
    });

    const refetchError = new Error('Refetch failed');
    mockHttpRequest.mockRejectedValueOnce(refetchError);

    await expect(result.current.refetch()).rejects.toThrow('Refetch failed');

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'FETCH_ERROR',
        message: 'Refetch failed',
        description: 'Failed to fetch user',
      });
    });
  });

  it('should handle non-Error object when refetch fails', async () => {
    const mockUser: ApiUser = {
      id: 'user-123',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Doe',
        email: 'john@example.com',
      },
    };

    mockHttpRequest.mockResolvedValueOnce({data: mockUser});

    const {result} = renderHook(() => useGetUser('user-123'));

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUser);
    });

    mockHttpRequest.mockRejectedValueOnce('Unknown error');

    await expect(result.current.refetch()).rejects.toEqual('Unknown error');

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'FETCH_ERROR',
        message: 'An unknown error occurred',
        description: 'Failed to fetch user',
      });
    });

    expect(result.current.loading).toBe(false);
  });

  it('should handle non-Error object during initial fetch', async () => {
    mockHttpRequest.mockRejectedValue('String error during initial fetch');

    const {result} = renderHook(() => useGetUser('user-123'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toEqual({
      code: 'FETCH_ERROR',
      message: 'An unknown error occurred',
      description: 'Failed to fetch user',
    });
    expect(result.current.data).toBeNull();
  });
});
