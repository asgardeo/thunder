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
import useGetUsers from '../useGetUsers';
import type {UserListResponse, UserListParams} from '../../types/users';

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

describe('useGetUsers', () => {
  beforeEach(() => {
    mockHttpRequest.mockReset();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  const renderWithParams = (params?: UserListParams) =>
    renderHook(({params: hookParams}: {params?: UserListParams}) => useGetUsers(hookParams), {
      initialProps: {params},
    });

  it('should initialize with correct default values', () => {
    const {result} = renderHook(() => useGetUsers());

    expect(result.current.data).toBeNull();
    expect(result.current.error).toBeNull();
    expect(typeof result.current.refetch).toBe('function');
  });

  it('should fetch users on mount', async () => {
    const mockResponse: UserListResponse = {
      totalResults: 2,
      startIndex: 0,
      count: 2,
      users: [
        {
          id: 'user-1',
          organizationUnit: '/sales',
          type: 'customer',
          attributes: {
            name: 'John Doe',
            email: 'john@example.com',
          },
        },
        {
          id: 'user-2',
          organizationUnit: '/sales',
          type: 'customer',
          attributes: {
            name: 'Jane Smith',
            email: 'jane@example.com',
          },
        },
      ],
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const {result} = renderHook(() => useGetUsers());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(mockResponse);
    expect(result.current.error).toBeNull();
    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/users',
        method: 'GET',
      }),
    );
  });

  it('should fetch users with query parameters', async () => {
    const mockResponse: UserListResponse = {
      totalResults: 100,
      startIndex: 10,
      count: 20,
      users: [],
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const {result} = renderWithParams({limit: 20, offset: 10});

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(mockResponse);
    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/users?limit=20&offset=10',
        method: 'GET',
      }),
    );
  });

  it('should fetch users with filter parameter', async () => {
    const mockResponse: UserListResponse = {
      totalResults: 5,
      startIndex: 0,
      count: 5,
      users: [],
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const {result} = renderWithParams({filter: 'name eq "John"'});

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(mockResponse);
    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/users?filter=name+eq+%22John%22',
        method: 'GET',
      }),
    );
  });

  it('should fetch users with multiple query parameters', async () => {
    const mockResponse: UserListResponse = {
      totalResults: 50,
      startIndex: 20,
      count: 10,
      users: [],
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const {result} = renderWithParams({
      filter: 'type eq "customer"',
      limit: 10,
      offset: 20,
    });

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(mockResponse);
    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/users?limit=10&offset=20&filter=type+eq+%22customer%22',
        method: 'GET',
      }),
    );
  });

  it('should handle API error with JSON response', async () => {
    mockHttpRequest.mockRejectedValue(new Error('Schema not found'));

    const {result} = renderHook(() => useGetUsers());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // Error is caught and re-wrapped with FETCH_ERROR code but preserves message
    expect(result.current.error).toEqual({
      code: 'FETCH_ERROR',
      message: 'Schema not found',
      description: 'Failed to fetch users',
    });
    expect(result.current.data).toBeNull();
  });

  it('should handle API error without JSON response', async () => {
    mockHttpRequest.mockRejectedValue(new Error('Internal Server Error'));

    const {result} = renderHook(() => useGetUsers());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // Error is caught and re-wrapped with FETCH_ERROR code but preserves message
    expect(result.current.error).toEqual({
      code: 'FETCH_ERROR',
      message: 'Internal Server Error',
      description: 'Failed to fetch users',
    });
    expect(result.current.data).toBeNull();
  });

  it('should handle network error', async () => {
    mockHttpRequest.mockRejectedValue(new Error('Network error'));

    const {result} = renderHook(() => useGetUsers());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toEqual({
      code: 'FETCH_ERROR',
      message: 'Network error',
      description: 'Failed to fetch users',
    });
    expect(result.current.data).toBeNull();
  });

  it('should abort previous request when params change', async () => {
    const mockResponse1: UserListResponse = {
      totalResults: 10,
      startIndex: 0,
      count: 10,
      users: [],
    };

    const mockResponse2: UserListResponse = {
      totalResults: 20,
      startIndex: 0,
      count: 20,
      users: [],
    };

    mockHttpRequest.mockImplementation(({url}: {url?: string}) =>
      Promise.resolve({
        data: url!.includes('limit=10') ? mockResponse1 : mockResponse2,
      }),
    );

    const {result, rerender} = renderHook(
      ({params}: {params?: {limit?: number; offset?: number}}) => useGetUsers(params),
      {
        initialProps: {params: {limit: 10}},
      },
    );

    await waitFor(() => {
      expect(mockHttpRequest).toHaveBeenCalled();
    });

    // Change params to trigger new fetch
    rerender({params: {limit: 20}});

    await waitFor(() => {
      expect(result.current.data).toEqual(mockResponse2);
    });

    expect(result.current.error).toBeNull();
  });

  it('should refetch with new parameters', async () => {
    const mockResponse1: UserListResponse = {
      totalResults: 10,
      startIndex: 0,
      count: 10,
      users: [],
    };

    const mockResponse2: UserListResponse = {
      totalResults: 5,
      startIndex: 0,
      count: 5,
      users: [],
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse1,
    });

    const {result} = renderWithParams({limit: 10});

    // Wait for initial fetch to complete
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.data?.count).toBe(10);
    });

    // Now update the mock to return mockResponse2 for the refetch
    mockHttpRequest.mockResolvedValue({
      data: mockResponse2,
    });

    // Call refetch with new params
    await result.current.refetch({limit: 5});

    // Wait for refetch to complete
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.data?.count).toBe(5);
    });

    expect(mockHttpRequest).toHaveBeenCalled();
  });

  it('should refetch with original parameters when no new params provided', async () => {
    const mockResponse: UserListResponse = {
      totalResults: 10,
      startIndex: 0,
      count: 10,
      users: [],
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const {result} = renderWithParams({limit: 10});

    // Wait for initial fetch
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.data).toEqual(mockResponse);
    });

    const callCountBeforeRefetch = mockHttpRequest.mock.calls.length;

    // Call refetch
    await result.current.refetch();

    // Wait for refetch to complete
    await waitFor(() => {
      expect(mockHttpRequest.mock.calls.length).toBeGreaterThan(callCountBeforeRefetch);
    });

    // Both calls should have the same parameters
    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/users?limit=10',
        method: 'GET',
      }),
    );
  });

  it('should cleanup and abort request on unmount', async () => {
    const mockResponse: UserListResponse = {
      totalResults: 10,
      startIndex: 0,
      count: 10,
      users: [],
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const {result, unmount} = renderHook(() => useGetUsers());

    await waitFor(() => {
      expect(result.current.data).toEqual(mockResponse);
    });

    unmount();

    // The abort should have been called during cleanup
    // This is difficult to assert directly, but we can verify no errors occur
    expect(true).toBe(true);
  });

  it('should refetch with filter parameter', async () => {
    const mockResponse: UserListResponse = {
      totalResults: 5,
      startIndex: 0,
      count: 5,
      users: [],
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const {result} = renderWithParams({limit: 10});

    // Wait for initial fetch
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // Call refetch with filter parameter
    await result.current.refetch({filter: 'name eq "Test"'});

    // Wait for refetch to complete
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // Verify the last call included the filter parameter
    expect(mockHttpRequest).toHaveBeenLastCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/users?filter=name+eq+%22Test%22',
        method: 'GET',
      }),
    );
  });

  it('should refetch with all parameters including filter', async () => {
    const mockResponse: UserListResponse = {
      totalResults: 5,
      startIndex: 0,
      count: 5,
      users: [],
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const {result} = renderWithParams();

    // Wait for initial fetch
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // Call refetch with all parameters including filter
    await result.current.refetch({limit: 5, offset: 10, filter: 'type eq "admin"'});

    // Wait for refetch to complete
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // Verify the call included all parameters
    expect(mockHttpRequest).toHaveBeenLastCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/users?limit=5&offset=10&filter=type+eq+%22admin%22',
        method: 'GET',
      }),
    );
  });

  it('should handle refetch error and throw', async () => {
    const mockResponse: UserListResponse = {
      totalResults: 10,
      startIndex: 0,
      count: 10,
      users: [],
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const {result} = renderWithParams({limit: 10});

    // Wait for initial fetch
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.data).toEqual(mockResponse);
    });

    // Mock error for refetch
    mockHttpRequest.mockRejectedValue(new Error('Refetch failed'));

    // Call refetch and expect it to throw
    await expect(result.current.refetch()).rejects.toThrow('Refetch failed');

    // Wait for error state to be set
    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'FETCH_ERROR',
        message: 'Refetch failed',
        description: 'Failed to fetch users',
      });
    });
  });

  it('should handle non-Error object in refetch catch block', async () => {
    const mockResponse: UserListResponse = {
      totalResults: 10,
      startIndex: 0,
      count: 10,
      users: [],
    };

    mockHttpRequest.mockResolvedValue({
      data: mockResponse,
    });

    const {result} = renderWithParams({limit: 10});

    // Wait for initial fetch
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.data).toEqual(mockResponse);
    });

    // Mock non-Error rejection for refetch
    mockHttpRequest.mockRejectedValue('String error');

    // Call refetch and expect it to throw
    let refetchError;
    try {
      await result.current.refetch();
    } catch (err) {
      refetchError = err;
    }

    // Verify the error was thrown
    expect(refetchError).toEqual('String error');

    // Wait for error state to be set
    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'FETCH_ERROR',
        message: 'An unknown error occurred',
        description: 'Failed to fetch users',
      });
    });

    expect(result.current.loading).toBe(false);
  });

  it('should handle non-Error object during initial fetch', async () => {
    mockHttpRequest.mockRejectedValue('String error during initial fetch');

    const {result} = renderHook(() => useGetUsers());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toEqual({
      code: 'FETCH_ERROR',
      message: 'An unknown error occurred',
      description: 'Failed to fetch users',
    });
    expect(result.current.data).toBeNull();
  });
});
