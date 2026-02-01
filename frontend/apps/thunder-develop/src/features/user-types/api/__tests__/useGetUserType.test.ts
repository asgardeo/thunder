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
import useGetUserType from '../useGetUserType';
import type {ApiUserSchema} from '../../types/user-types';

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

describe('useGetUserType', () => {
  const mockUserSchema: ApiUserSchema = {
    id: '123',
    name: 'TestUserType',
    ouId: 'root-ou',
    allowSelfRegistration: true,
    schema: {
      username: {
        type: 'string',
        required: true,
      },
      email: {
        type: 'string',
        required: false,
      },
    },
  };

  beforeEach(() => {
    mockHttpRequest.mockReset();
    mockGetServerUrl.mockReturnValue('https://localhost:8090');
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should initialize with default state when no id is provided', () => {
    const {result} = renderHook(() => useGetUserType());

    expect(result.current.data).toBeNull();
    expect(result.current.error).toBeNull();
    expect(result.current.loading).toBe(false);
    expect(typeof result.current.refetch).toBe('function');
  });

  it('should fetch user type when id is provided', async () => {
    mockHttpRequest.mockResolvedValueOnce({data: mockUserSchema});

    const {result} = renderHook(() => useGetUserType('123'));

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchema);
      expect(result.current.error).toBeNull();
      expect(result.current.loading).toBe(false);
    });

    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({url: 'https://localhost:8090/user-schemas/123', method: 'GET'}),
    );
  });

  it('should set loading state during fetch', async () => {
    // Create a promise we can control
    let resolveRequest: (value: {data: ApiUserSchema}) => void;
    const requestPromise = new Promise<{data: ApiUserSchema}>((resolve) => {
      resolveRequest = resolve;
    });

    mockHttpRequest.mockReturnValueOnce(requestPromise);

    const {result} = renderHook(() => useGetUserType('123'));

    await waitFor(() => {
      expect(result.current.loading).toBe(true);
    });

    // Now resolve the request
    resolveRequest!({data: mockUserSchema});

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
  });

  it('should handle API error with JSON response', async () => {
    mockHttpRequest.mockRejectedValueOnce(new Error('User type not found'));

    const {result} = renderHook(() => useGetUserType('123'));

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'FETCH_USER_TYPE_ERROR',
        message: 'User type not found',
        description: 'Failed to fetch user type',
      });
      expect(result.current.data).toBeNull();
      expect(result.current.loading).toBe(false);
    });
  });

  it('should handle API error without JSON response', async () => {
    mockHttpRequest.mockRejectedValueOnce(new Error('Internal Server Error'));

    const {result} = renderHook(() => useGetUserType('123'));

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'FETCH_USER_TYPE_ERROR',
        message: 'Internal Server Error',
        description: 'Failed to fetch user type',
      });
      expect(result.current.data).toBeNull();
      expect(result.current.loading).toBe(false);
    });
  });

  it('should handle network error', async () => {
    mockHttpRequest.mockRejectedValueOnce(new Error('Network error'));

    const {result} = renderHook(() => useGetUserType('123'));

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'FETCH_USER_TYPE_ERROR',
        message: 'Network error',
        description: 'Failed to fetch user type',
      });
      expect(result.current.data).toBeNull();
      expect(result.current.loading).toBe(false);
    });
  });

  it('should refetch when refetch is called', async () => {
    mockHttpRequest.mockResolvedValueOnce({data: mockUserSchema});

    const {result} = renderHook(() => useGetUserType('123'));

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchema);
    });

    const callsBeforeRefetch = mockHttpRequest.mock.calls.length;
    mockHttpRequest
      .mockResolvedValueOnce({data: {...mockUserSchema, name: 'UpdatedUserType'}})
      .mockResolvedValueOnce({data: {...mockUserSchema, name: 'UpdatedUserType'}});

    await result.current.refetch();

    await waitFor(() => {
      expect(result.current.data).toEqual({...mockUserSchema, name: 'UpdatedUserType'});
    });

    expect(mockHttpRequest.mock.calls.length).toBeGreaterThan(callsBeforeRefetch);
  });

  it('should refetch with new id when provided to refetch', async () => {
    mockHttpRequest.mockResolvedValueOnce({data: mockUserSchema});

    const {result} = renderHook(() => useGetUserType('123'));

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchema);
    });

    const newMockUserSchema = {...mockUserSchema, id: '456', name: 'NewUserType'};

    mockHttpRequest.mockResolvedValueOnce({data: newMockUserSchema}).mockResolvedValueOnce({data: newMockUserSchema});

    await result.current.refetch('456');

    await waitFor(() => {
      expect(result.current.data).toEqual(newMockUserSchema);
    });

    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({url: 'https://localhost:8090/user-schemas/456', method: 'GET'}),
    );
  });

  it('should not fetch if refetch is called without id when no id is provided', async () => {
    const {result} = renderHook(() => useGetUserType());

    await result.current.refetch();

    expect(mockHttpRequest).not.toHaveBeenCalled();
  });

  it('should prevent double-fetch in React Strict Mode', async () => {
    mockHttpRequest.mockResolvedValue({data: mockUserSchema});

    const {result, rerender} = renderHook(() => useGetUserType('123'));

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchema);
    });

    // Simulate React Strict Mode double render
    rerender();

    await waitFor(() => {
      // Should only fetch once, not twice
      expect(mockHttpRequest).toHaveBeenCalledTimes(1);
    });
  });

  it('should fetch again when id changes', async () => {
    mockHttpRequest.mockResolvedValueOnce({data: mockUserSchema});

    const {result, rerender} = renderHook(({id}) => useGetUserType(id), {
      initialProps: {id: '123'},
    });

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchema);
    });

    const newMockUserSchema = {...mockUserSchema, id: '456', name: 'NewUserType'};

    mockHttpRequest.mockResolvedValueOnce({data: newMockUserSchema});

    rerender({id: '456'});

    await waitFor(() => {
      expect(result.current.data).toEqual(newMockUserSchema);
    });

    expect(mockHttpRequest).toHaveBeenCalledTimes(2);
    expect(mockHttpRequest).toHaveBeenNthCalledWith(
      1,
      expect.objectContaining({url: 'https://localhost:8090/user-schemas/123', method: 'GET'}),
    );
    expect(mockHttpRequest).toHaveBeenNthCalledWith(
      2,
      expect.objectContaining({url: 'https://localhost:8090/user-schemas/456', method: 'GET'}),
    );
  });

  it('should clear data when id changes to undefined', async () => {
    mockHttpRequest.mockResolvedValueOnce({data: mockUserSchema});

    const {result, rerender} = renderHook(({id}) => useGetUserType(id), {
      initialProps: {id: '123' as string | undefined},
    });

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchema);
    });

    rerender({id: undefined});

    expect(result.current.data).toBeNull();
    expect(result.current.error).toBeNull();
  });

  it('should handle error in refetch without id', async () => {
    const {result} = renderHook(() => useGetUserType());

    await result.current.refetch();

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'INVALID_ID',
        message: 'Invalid schema ID',
        description: 'Schema ID is required',
      });
    });

    expect(mockHttpRequest).not.toHaveBeenCalled();
  });

  it('should throw error when refetch fails', async () => {
    mockHttpRequest.mockResolvedValueOnce({data: mockUserSchema});

    const {result} = renderHook(() => useGetUserType('123'));

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchema);
    });

    const error = new Error('Refetch failed');
    mockHttpRequest.mockRejectedValueOnce(error);

    await expect(result.current.refetch()).rejects.toThrow('Refetch failed');

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'FETCH_USER_TYPE_ERROR',
        message: 'Refetch failed',
        description: 'Failed to fetch user type',
      });
      expect(result.current.data).toBeNull();
    });
  });

  it('should handle non-Error thrown in refetch', async () => {
    mockHttpRequest.mockResolvedValueOnce({data: mockUserSchema});

    const {result} = renderHook(() => useGetUserType('123'));

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchema);
    });

    mockHttpRequest.mockRejectedValueOnce('String error');

    await expect(result.current.refetch()).rejects.toBe('String error');

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'FETCH_USER_TYPE_ERROR',
        message: 'An unknown error occurred',
        description: 'Failed to fetch user type',
      });
      expect(result.current.data).toBeNull();
    });
  });

  it('should handle non-Error thrown in initial fetch', async () => {
    mockHttpRequest.mockRejectedValueOnce('String error in fetch');

    const {result} = renderHook(() => useGetUserType('123'));

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'FETCH_USER_TYPE_ERROR',
        message: 'An unknown error occurred',
        description: 'Failed to fetch user type',
      });
      expect(result.current.data).toBeNull();
      expect(result.current.loading).toBe(false);
    });
  });

  it('should fallback to env variable when getServerUrl returns undefined', async () => {
    mockGetServerUrl.mockReturnValue(undefined);
    mockHttpRequest.mockResolvedValueOnce({data: mockUserSchema});

    const {result} = renderHook(() => useGetUserType('123'));

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchema);
    });

    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: expect.stringContaining('/user-schemas/123') as string,
        method: 'GET',
      }),
    );
  });

  it('should return early when same id is fetched again (double-fetch prevention)', async () => {
    mockHttpRequest.mockResolvedValue({data: mockUserSchema});

    const {result, rerender} = renderHook(({id}) => useGetUserType(id), {
      initialProps: {id: '123'},
    });

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchema);
    });

    // First call made
    expect(mockHttpRequest).toHaveBeenCalledTimes(1);

    // Rerender with same id (simulating Strict Mode or re-render)
    rerender({id: '123'});

    // Wait a bit for any potential additional calls
    await new Promise((resolve) => {
      setTimeout(resolve, 50);
    });

    // Should still only have one call due to double-fetch prevention
    expect(mockHttpRequest).toHaveBeenCalledTimes(1);
  });
});
