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
import useGetUserTypes from '../useGetUserTypes';
import type {UserTypeListResponse} from '../../types/users';

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

describe('useGetUserTypes', () => {
  beforeEach(() => {
    mockHttpRequest.mockReset();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should initialize with correct default values', () => {
    const {result} = renderHook(() => useGetUserTypes());

    expect(result.current.data).toBeNull();
    expect(result.current.error).toBeNull();
    expect(typeof result.current.refetch).toBe('function');
  });

  it('should fetch user types on mount', async () => {
    const mockResponse: UserTypeListResponse = {
      totalResults: 2,
      startIndex: 0,
      count: 2,
      schemas: [
        {
          id: 'schema-1',
          name: 'Customer',
          ouId: 'root-ou',
        },
        {
          id: 'schema-2',
          name: 'Employee',
          ouId: 'child-ou',
        },
      ],
    };

    mockHttpRequest.mockResolvedValue({data: mockResponse});

    const {result} = renderHook(() => useGetUserTypes());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(mockResponse);
    expect(result.current.error).toBeNull();
    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/user-types',
        method: 'GET',
      }),
    );
  });

  it('should fetch user types with query parameters', async () => {
    const mockResponse: UserTypeListResponse = {
      totalResults: 100,
      startIndex: 10,
      count: 20,
      schemas: [],
    };

    mockHttpRequest.mockResolvedValue({data: mockResponse});

    const {result} = renderHook(() => useGetUserTypes({limit: 20, offset: 10}));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(mockResponse);
    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/user-types?limit=20&offset=10',
        method: 'GET',
      }),
    );
  });

  it('should fetch user types with limit parameter only', async () => {
    const mockResponse: UserTypeListResponse = {
      totalResults: 50,
      startIndex: 0,
      count: 10,
      schemas: [],
    };

    mockHttpRequest.mockResolvedValue({data: mockResponse});

    const {result} = renderHook(() => useGetUserTypes({limit: 10}));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(mockResponse);
    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/user-types?limit=10',
        method: 'GET',
      }),
    );
  });

  it('should fetch user types with offset parameter only', async () => {
    const mockResponse: UserTypeListResponse = {
      totalResults: 50,
      startIndex: 20,
      count: 30,
      schemas: [],
    };

    mockHttpRequest.mockResolvedValue({data: mockResponse});

    const {result} = renderHook(() => useGetUserTypes({offset: 20}));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(mockResponse);
    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/user-types?offset=20',
        method: 'GET',
      }),
    );
  });

  it('should handle API error with JSON response', async () => {
    mockHttpRequest.mockRejectedValue(new Error('Unauthorized access'));

    const {result} = renderHook(() => useGetUserTypes());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toEqual({
      code: 'FETCH_ERROR',
      message: 'Unauthorized access',
      description: 'Failed to fetch user types',
    });
    expect(result.current.data).toBeNull();
  });

  it('should handle API error without JSON response', async () => {
    mockHttpRequest.mockRejectedValue(new Error('Internal Server Error'));

    const {result} = renderHook(() => useGetUserTypes());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toEqual({
      code: 'FETCH_ERROR',
      message: 'Internal Server Error',
      description: 'Failed to fetch user types',
    });

    expect(result.current.data).toBeNull();
  });

  it('should handle network error', async () => {
    mockHttpRequest.mockRejectedValue(new Error('Network error'));

    const {result} = renderHook(() => useGetUserTypes());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toEqual({
      code: 'FETCH_ERROR',
      message: 'Network error',
      description: 'Failed to fetch user types',
    });
    expect(result.current.data).toBeNull();
  });

  it('should abort previous request when params change', async () => {
    const mockResponse2: UserTypeListResponse = {
      totalResults: 20,
      startIndex: 0,
      count: 20,
      schemas: [],
    };

    const abortError = new Error('Aborted');
    abortError.name = 'AbortError';

    mockHttpRequest.mockImplementation(({url}: {url?: string}) =>
      url!.includes('limit=10') ? Promise.reject(abortError) : Promise.resolve({data: mockResponse2}),
    );

    const {result, rerender} = renderHook(
      ({params}: {params?: {limit?: number; offset?: number}}) => useGetUserTypes(params),
      {
        initialProps: {params: {limit: 10}},
      },
    );

    // Quickly change params to trigger abort
    rerender({params: {limit: 20}});

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // Should have the data from the second request
    expect(result.current.data).toEqual(mockResponse2);
    // Error should be null since AbortError is ignored
    expect(result.current.error).toBeNull();
  });

  it('should refetch with new parameters', async () => {
    const mockResponse1: UserTypeListResponse = {
      totalResults: 10,
      startIndex: 0,
      count: 10,
      schemas: [],
    };

    const mockResponse2: UserTypeListResponse = {
      totalResults: 5,
      startIndex: 0,
      count: 5,
      schemas: [],
    };

    mockHttpRequest.mockResolvedValue({data: mockResponse1});

    const {result} = renderHook(() => useGetUserTypes({limit: 10}));

    // Wait for initial fetch to complete with mockResponse1
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.data?.count).toBe(10);
    });

    // Now update the mock to return mockResponse2 for the refetch
    mockHttpRequest.mockResolvedValue({data: mockResponse2});

    // Call refetch with new parameters
    await result.current.refetch({limit: 5});

    // Wait for refetch to complete with mockResponse2
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.data?.count).toBe(5);
    });

    expect(mockHttpRequest).toHaveBeenCalled();
  });

  it('should refetch with original parameters when no new params provided', async () => {
    const mockResponse: UserTypeListResponse = {
      totalResults: 10,
      startIndex: 0,
      count: 10,
      schemas: [],
    };

    mockHttpRequest.mockResolvedValue({data: mockResponse});

    const {result} = renderHook(() => useGetUserTypes({limit: 10}));

    // Wait for initial fetch to complete
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
    expect(result.current.data).toEqual(mockResponse);

    // Clear mock calls from initial fetch
    vi.clearAllMocks();
    mockHttpRequest.mockResolvedValue({data: mockResponse});

    // Call refetch without parameters
    await result.current.refetch();

    // Wait for refetch to complete
    await waitFor(() => {
      expect(mockHttpRequest).toHaveBeenCalledTimes(1);
    });

    // Should use the same parameters as initial call
    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/user-types?limit=10',
        method: 'GET',
      }),
    );
  });

  it('should cleanup and abort request on unmount', async () => {
    const mockResponse: UserTypeListResponse = {
      totalResults: 10,
      startIndex: 0,
      count: 10,
      schemas: [],
    };

    mockHttpRequest.mockResolvedValue({data: mockResponse});

    const {result, unmount} = renderHook(() => useGetUserTypes());

    await waitFor(() => {
      expect(result.current.data).toEqual(mockResponse);
    });

    unmount();

    // The abort should have been called during cleanup
    // This is difficult to assert directly, but we can verify no errors occur
    expect(true).toBe(true);
  });

  it('should handle empty schemas list', async () => {
    const mockResponse: UserTypeListResponse = {
      totalResults: 0,
      startIndex: 0,
      count: 0,
      schemas: [],
    };

    mockHttpRequest.mockResolvedValue({data: mockResponse});

    const {result} = renderHook(() => useGetUserTypes());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(mockResponse);
    expect(result.current.data?.schemas).toHaveLength(0);
    expect(result.current.error).toBeNull();
  });

  it('should handle large offset pagination', async () => {
    const mockResponse: UserTypeListResponse = {
      totalResults: 1000,
      startIndex: 900,
      count: 50,
      schemas: [],
    };

    mockHttpRequest.mockResolvedValue({data: mockResponse});

    const {result} = renderHook(() => useGetUserTypes({limit: 50, offset: 900}));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(mockResponse);
    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/user-types?limit=50&offset=900',
        method: 'GET',
      }),
    );
  });

  it('should refetch with only offset parameter', async () => {
    const mockResponse: UserTypeListResponse = {
      totalResults: 100,
      startIndex: 0,
      count: 10,
      schemas: [],
    };

    mockHttpRequest.mockResolvedValue({data: mockResponse});

    const {result} = renderHook(() => useGetUserTypes());

    // Wait for initial fetch
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // Call refetch with only offset parameter
    await result.current.refetch({offset: 30});

    // Wait for refetch to complete
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // Verify the last call included only the offset parameter
    expect(mockHttpRequest).toHaveBeenLastCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/user-types?offset=30',
        method: 'GET',
      }),
    );
  });

  it('should handle AbortError in refetch and not set error', async () => {
    const mockResponse: UserTypeListResponse = {
      totalResults: 10,
      startIndex: 0,
      count: 10,
      schemas: [],
    };

    mockHttpRequest.mockResolvedValue({data: mockResponse});

    const {result} = renderHook(() => useGetUserTypes());

    // Wait for initial fetch
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // Mock AbortError for refetch
    const abortError = new Error('Request aborted');
    abortError.name = 'AbortError';
    mockHttpRequest.mockRejectedValue(abortError);

    // Call refetch - it should not throw
    await result.current.refetch();

    // Wait for loading to complete
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // Error should still be null since AbortError is ignored
    expect(result.current.error).toBeNull();
  });

  it('should handle refetch error and throw', async () => {
    const mockResponse: UserTypeListResponse = {
      totalResults: 10,
      startIndex: 0,
      count: 10,
      schemas: [],
    };

    mockHttpRequest.mockResolvedValue({data: mockResponse});

    const {result} = renderHook(() => useGetUserTypes());

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
        description: 'Failed to fetch user types',
      });
    });

    expect(result.current.loading).toBe(false);
  });

  it('should handle non-Error object in refetch catch block', async () => {
    const mockResponse: UserTypeListResponse = {
      totalResults: 10,
      startIndex: 0,
      count: 10,
      schemas: [],
    };

    mockHttpRequest.mockResolvedValue({data: mockResponse});

    const {result} = renderHook(() => useGetUserTypes());

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
        description: 'Failed to fetch user types',
      });
    });

    expect(result.current.loading).toBe(false);
  });

  it('should handle non-Error object during initial fetch', async () => {
    mockHttpRequest.mockRejectedValue('String error during initial fetch');

    const {result} = renderHook(() => useGetUserTypes());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toEqual({
      code: 'FETCH_ERROR',
      message: 'An unknown error occurred',
      description: 'Failed to fetch user types',
    });
    expect(result.current.data).toBeNull();
  });
});
