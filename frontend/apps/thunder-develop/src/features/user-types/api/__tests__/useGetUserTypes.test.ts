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
import useGetUserTypes from '../useGetUserTypes';
import type {UserSchemaListResponse} from '../../types/user-types';

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

describe('useGetUserTypes', () => {
  const mockUserSchemaList: UserSchemaListResponse = {
    totalResults: 2,
    startIndex: 1,
    count: 2,
    schemas: [
      {id: '123', name: 'UserType1', ouId: 'root-ou', allowSelfRegistration: false},
      {id: '456', name: 'UserType2', ouId: 'child-ou', allowSelfRegistration: true},
    ],
    links: [{rel: 'self', href: 'https://localhost:8090/user-types'}],
  };

  beforeEach(() => {
    mockHttpRequest.mockReset();
    mockGetServerUrl.mockReturnValue('https://localhost:8090');
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should initialize and start fetching', async () => {
    mockHttpRequest.mockResolvedValue({data: mockUserSchemaList});

    const {result} = renderHook(() => useGetUserTypes());

    expect(result.current.data).toBeNull();
    expect(result.current.error).toBeNull();
    expect(typeof result.current.refetch).toBe('function');

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
  });

  it('should fetch user types on mount', async () => {
    mockHttpRequest.mockResolvedValue({data: mockUserSchemaList});

    const {result} = renderHook(() => useGetUserTypes());

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchemaList);
      expect(result.current.error).toBeNull();
      expect(result.current.loading).toBe(false);
    });

    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({url: 'https://localhost:8090/user-types', method: 'GET'}),
    );
  });

  it('should fetch user types with limit parameter', async () => {
    mockHttpRequest.mockResolvedValue({data: mockUserSchemaList});

    renderHook(() => useGetUserTypes({limit: 10}));

    await waitFor(() => {
      expect(
        mockHttpRequest.mock.calls.some(
          (call: unknown[]) => (call[0] as {url?: string})?.url === 'https://localhost:8090/user-types?limit=10',
        ),
      ).toBe(true);
    });
  });

  it('should fetch user types with offset parameter', async () => {
    mockHttpRequest.mockResolvedValue({data: mockUserSchemaList});

    renderHook(() => useGetUserTypes({offset: 5}));

    await waitFor(() => {
      expect(mockHttpRequest).toHaveBeenCalledWith(
        expect.objectContaining({
          url: 'https://localhost:8090/user-types?offset=5',
          method: 'GET',
        }),
      );
    });
  });

  it('should fetch user types with both limit and offset parameters', async () => {
    mockHttpRequest.mockResolvedValue({data: mockUserSchemaList});

    renderHook(() => useGetUserTypes({limit: 10, offset: 5}));

    await waitFor(() => {
      expect(mockHttpRequest).toHaveBeenCalledWith(
        expect.objectContaining({
          url: 'https://localhost:8090/user-types?limit=10&offset=5',
          method: 'GET',
        }),
      );
    });
  });

  it('should set loading state during fetch', async () => {
    mockHttpRequest.mockImplementation(
      () =>
        new Promise(() => {
          // Never resolve to keep loading in true state for assertion
        }),
    );

    const {result, unmount} = renderHook(() => useGetUserTypes());

    await waitFor(() => {
      expect(result.current.loading).toBe(true);
    });

    unmount();
  });

  it('should handle API error with JSON response', async () => {
    mockHttpRequest.mockRejectedValue(new Error('Failed to fetch user types'));

    const {result} = renderHook(() => useGetUserTypes());

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'FETCH_USER_TYPES_ERROR',
        message: 'Failed to fetch user types',
        description: 'Failed to fetch user types',
      });
      expect(result.current.data).toBeNull();
      expect(result.current.loading).toBe(false);
    });
  });

  it('should handle API error without JSON response', async () => {
    mockHttpRequest.mockRejectedValue(new Error('HTTP error! status: 500'));

    const {result} = renderHook(() => useGetUserTypes());

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'FETCH_USER_TYPES_ERROR',
        message: 'HTTP error! status: 500',
        description: 'Failed to fetch user types',
      });
      expect(result.current.data).toBeNull();
      expect(result.current.loading).toBe(false);
    });
  });

  it('should handle network error', async () => {
    mockHttpRequest.mockRejectedValue(new Error('Network error'));

    const {result} = renderHook(() => useGetUserTypes());

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'FETCH_USER_TYPES_ERROR',
        message: 'Network error',
        description: 'Failed to fetch user types',
      });
      expect(result.current.data).toBeNull();
      expect(result.current.loading).toBe(false);
    });
  });

  it('should refetch when refetch is called', async () => {
    mockHttpRequest.mockResolvedValue({data: mockUserSchemaList});

    const {result} = renderHook(() => useGetUserTypes());

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchemaList);
    });

    const updatedList = {...mockUserSchemaList, totalResults: 3};
    mockHttpRequest.mockResolvedValue({data: updatedList});
    const callsBeforeRefetch = mockHttpRequest.mock.calls.length;

    await result.current.refetch();

    await waitFor(() => {
      expect(result.current.data).toEqual(updatedList);
    });

    expect(mockHttpRequest.mock.calls.length).toBeGreaterThan(callsBeforeRefetch);
  });

  it('should abort previous request when params change', async () => {
    let abortSignal: AbortSignal | undefined;

    mockHttpRequest.mockImplementation((_config: unknown) => {
      abortSignal = (_config as {signal?: AbortSignal})?.signal ?? undefined;
      return new Promise((resolve) => {
        setTimeout(
          () =>
            resolve({
              data: mockUserSchemaList,
            }),
          100,
        );
      });
    });

    const {rerender} = renderHook(({params}) => useGetUserTypes(params), {
      initialProps: {params: {limit: 10}},
    });

    // Wait a bit for the first request to start
    await new Promise((resolve) => {
      setTimeout(resolve, 10);
    });

    const firstAbortSignal = abortSignal;

    // Change params to trigger a new request
    rerender({params: {limit: 20}});

    await waitFor(() => {
      expect(firstAbortSignal?.aborted).toBe(true);
    });
  });

  it('should not set error for aborted requests', async () => {
    const abortError = new Error('Aborted');
    abortError.name = 'AbortError';

    mockHttpRequest.mockRejectedValue(abortError);

    const {result} = renderHook(() => useGetUserTypes());

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toBeNull();
  });

  it('should abort request on unmount', async () => {
    let abortSignal: AbortSignal | undefined;

    mockHttpRequest.mockImplementation((_config: unknown) => {
      abortSignal = (_config as {signal?: AbortSignal})?.signal ?? undefined;
      return new Promise((resolve) => {
        setTimeout(
          () =>
            resolve({
              data: mockUserSchemaList,
            }),
          100,
        );
      });
    });

    const {unmount} = renderHook(() => useGetUserTypes());

    // Wait a bit for the request to start
    await new Promise((resolve) => {
      setTimeout(resolve, 10);
    });

    unmount();

    await waitFor(() => {
      expect(abortSignal?.aborted).toBe(true);
    });
  });

  it('should refetch when params change', async () => {
    mockHttpRequest.mockResolvedValue({data: mockUserSchemaList});

    const {rerender} = renderHook(({params}) => useGetUserTypes(params), {
      initialProps: {params: {limit: 10}},
    });

    await waitFor(() => {
      expect(mockHttpRequest).toHaveBeenCalledWith(
        expect.objectContaining({
          url: 'https://localhost:8090/user-types?limit=10',
          method: 'GET',
        }),
      );
    });

    rerender({params: {limit: 20}});

    await waitFor(() => {
      expect(
        mockHttpRequest.mock.calls.some(
          (call: unknown[]) => (call[0] as {url?: string})?.url === 'https://localhost:8090/user-types?limit=20',
        ),
      ).toBe(true);
    });
  });

  it('should refetch with new params when provided to refetch', async () => {
    mockHttpRequest.mockResolvedValue({data: mockUserSchemaList});

    const {result} = renderHook(() => useGetUserTypes({limit: 10}));

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchemaList);
    });

    const updatedList = {...mockUserSchemaList, totalResults: 5};
    mockHttpRequest.mockResolvedValue({data: updatedList});

    await result.current.refetch({limit: 20, offset: 10});

    await waitFor(() => {
      expect(result.current.data).toEqual(updatedList);
    });

    expect(
      mockHttpRequest.mock.calls.some(
        (call: unknown[]) =>
          (call[0] as {url?: string})?.url === 'https://localhost:8090/user-types?limit=20&offset=10',
      ),
    ).toBe(true);
  });

  it('should refetch with only limit param', async () => {
    mockHttpRequest.mockResolvedValue({data: mockUserSchemaList});

    const {result} = renderHook(() => useGetUserTypes());

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchemaList);
    });

    mockHttpRequest.mockResolvedValue({data: mockUserSchemaList});

    await result.current.refetch({limit: 15});

    await waitFor(() => {
      expect(
        mockHttpRequest.mock.calls.some(
          (call: unknown[]) => (call[0] as {url?: string})?.url === 'https://localhost:8090/user-types?limit=15',
        ),
      ).toBe(true);
    });
  });

  it('should refetch with only offset param', async () => {
    mockHttpRequest.mockResolvedValue({data: mockUserSchemaList});

    const {result} = renderHook(() => useGetUserTypes());

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchemaList);
    });

    mockHttpRequest.mockResolvedValue({data: mockUserSchemaList});

    await result.current.refetch({offset: 5});

    await waitFor(() => {
      expect(
        mockHttpRequest.mock.calls.some(
          (call: unknown[]) => (call[0] as {url?: string})?.url === 'https://localhost:8090/user-types?offset=5',
        ),
      ).toBe(true);
    });
  });

  it('should throw error when refetch fails', async () => {
    mockHttpRequest.mockResolvedValue({data: mockUserSchemaList});

    const {result} = renderHook(() => useGetUserTypes());

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchemaList);
    });

    const error = new Error('Refetch failed');
    mockHttpRequest.mockRejectedValue(error);

    await expect(result.current.refetch()).rejects.toThrow('Refetch failed');

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'FETCH_USER_TYPES_ERROR',
        message: 'Refetch failed',
        description: 'Failed to fetch user types',
      });
      expect(result.current.data).toBeNull();
    });
  });

  it('should handle non-Error thrown in refetch', async () => {
    mockHttpRequest.mockResolvedValue({data: mockUserSchemaList});

    const {result} = renderHook(() => useGetUserTypes());

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchemaList);
    });

    mockHttpRequest.mockRejectedValue('String error');

    await expect(result.current.refetch()).rejects.toBe('String error');

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'FETCH_USER_TYPES_ERROR',
        message: 'An unknown error occurred',
        description: 'Failed to fetch user types',
      });
      expect(result.current.data).toBeNull();
    });
  });

  it('should not set error for aborted requests in refetch', async () => {
    mockHttpRequest.mockResolvedValue({data: mockUserSchemaList});

    const {result} = renderHook(() => useGetUserTypes());

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchemaList);
    });

    const abortError = new Error('Aborted');
    abortError.name = 'AbortError';
    mockHttpRequest.mockRejectedValue(abortError);

    await result.current.refetch();

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // Error should remain null for aborted requests
    expect(result.current.error).toBeNull();
    // Data should remain from previous successful fetch
    expect(result.current.data).toEqual(mockUserSchemaList);
  });

  it('should handle non-Error thrown in initial fetch', async () => {
    mockHttpRequest.mockRejectedValueOnce('String error in fetch');

    const {result} = renderHook(() => useGetUserTypes());

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'FETCH_USER_TYPES_ERROR',
        message: 'An unknown error occurred',
        description: 'Failed to fetch user types',
      });
      expect(result.current.data).toBeNull();
      expect(result.current.loading).toBe(false);
    });
  });

  it('should fallback to env variable when getServerUrl returns undefined', async () => {
    mockGetServerUrl.mockReturnValue(undefined);
    mockHttpRequest.mockResolvedValueOnce({data: mockUserSchemaList});

    const {result} = renderHook(() => useGetUserTypes());

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchemaList);
    });

    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: expect.stringContaining('/user-types') as string,
        method: 'GET',
      }),
    );
  });
});
