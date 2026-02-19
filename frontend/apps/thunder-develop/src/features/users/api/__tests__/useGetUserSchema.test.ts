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
import useGetUserType from '../useGetUserType';
import type {ApiUserType} from '../../types/users';

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

describe('useGetUserType', () => {
  beforeEach(() => {
    mockHttpRequest.mockReset();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should initialize with correct default values', () => {
    const {result} = renderHook(() => useGetUserType());

    expect(result.current.data).toBeNull();
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeNull();
    expect(typeof result.current.refetch).toBe('function');
  });

  it('should fetch user type successfully when id is provided', async () => {
    const mockSchema: ApiUserType = {
      id: 'schema-123',
      name: 'Customer',
      schema: {
        name: {
          type: 'string',
          required: true,
        },
        email: {
          type: 'string',
          required: true,
        },
      },
    };

    mockHttpRequest.mockResolvedValueOnce({data: mockSchema});

    const {result} = renderHook(() => useGetUserType('schema-123'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(mockSchema);
    expect(result.current.error).toBeNull();
    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/user-types/schema-123',
        method: 'GET',
      }),
    );
  });

  it('should not fetch when id is not provided', () => {
    const {result} = renderHook(() => useGetUserType());

    expect(result.current.data).toBeNull();
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeNull();
    expect(mockHttpRequest).not.toHaveBeenCalled();
  });

  it('should handle API error with JSON response', async () => {
    mockHttpRequest.mockRejectedValue(new Error('Schema not found'));

    const {result} = renderHook(() => useGetUserType('schema-123'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toEqual({
      code: 'FETCH_ERROR',
      message: 'Schema not found',
      description: 'Failed to fetch user type',
    });
    expect(result.current.data).toBeNull();
  });

  it('should handle API error without JSON response', async () => {
    mockHttpRequest.mockRejectedValue(new Error('Internal Server Error'));

    const {result} = renderHook(() => useGetUserType('schema-123'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toEqual({
      code: 'FETCH_ERROR',
      message: 'Internal Server Error',
      description: 'Failed to fetch user type',
    });
    expect(result.current.data).toBeNull();
  });

  it('should handle network error', async () => {
    mockHttpRequest.mockRejectedValue(new Error('Network error'));

    const {result} = renderHook(() => useGetUserType('schema-123'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toEqual({
      code: 'FETCH_ERROR',
      message: 'Network error',
      description: 'Failed to fetch user type',
    });
    expect(result.current.data).toBeNull();
  });

  it('should refetch user type with the same id', async () => {
    const mockSchema: ApiUserType = {
      id: 'schema-123',
      name: 'Customer',
      schema: {
        name: {
          type: 'string',
          required: true,
        },
      },
    };

    mockHttpRequest.mockResolvedValue({data: mockSchema});

    const {result} = renderHook(() => useGetUserType('schema-123'));

    await waitFor(() => {
      expect(result.current.data).toEqual(mockSchema);
    });

    await result.current.refetch();

    await waitFor(() => {
      expect(mockHttpRequest).toHaveBeenCalledTimes(2);
    });
  });

  it('should refetch user type with a new id', async () => {
    const mockSchema1: ApiUserType = {
      id: 'schema-123',
      name: 'Customer',
      schema: {
        name: {
          type: 'string',
          required: true,
        },
      },
    };

    const mockSchema2: ApiUserType = {
      id: 'schema-456',
      name: 'Employee',
      schema: {
        employeeId: {
          type: 'string',
          required: true,
        },
      },
    };

    mockHttpRequest.mockResolvedValueOnce({data: mockSchema1}).mockResolvedValueOnce({data: mockSchema2});

    const {result} = renderHook(() => useGetUserType('schema-123'));

    await waitFor(() => {
      expect(result.current.data).toEqual(mockSchema1);
    });

    await result.current.refetch('schema-456');

    await waitFor(() => {
      expect(result.current.data).toEqual(mockSchema2);
    });
  });

  it('should set error when refetch is called without id', async () => {
    const {result} = renderHook(() => useGetUserType());

    await result.current.refetch();

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'INVALID_ID',
        message: 'Invalid user type ID',
        description: 'User type ID is required',
      });
    });
  });

  it('should prevent double fetch in strict mode', async () => {
    const mockSchema: ApiUserType = {
      id: 'schema-123',
      name: 'Customer',
      schema: {
        name: {
          type: 'string',
          required: true,
        },
      },
    };

    mockHttpRequest.mockResolvedValue({data: mockSchema});

    const {result} = renderHook(() => useGetUserType('schema-123'));

    await waitFor(() => {
      expect(result.current.data).toEqual(mockSchema);
    });

    // Should only fetch once despite strict mode
    expect(mockHttpRequest).toHaveBeenCalledTimes(1);
  });

  it('should fetch when id changes', async () => {
    const mockSchema1: ApiUserType = {
      id: 'schema-123',
      name: 'Customer',
      schema: {
        name: {
          type: 'string',
          required: true,
        },
      },
    };

    const mockSchema2: ApiUserType = {
      id: 'schema-456',
      name: 'Employee',
      schema: {
        employeeId: {
          type: 'string',
          required: true,
        },
      },
    };

    mockHttpRequest.mockResolvedValueOnce({data: mockSchema1}).mockResolvedValueOnce({data: mockSchema2});

    const {result, rerender} = renderHook(({id}: {id?: string}) => useGetUserType(id), {
      initialProps: {id: 'schema-123'},
    });

    await waitFor(() => {
      expect(result.current.data).toEqual(mockSchema1);
    });

    rerender({id: 'schema-456'});

    await waitFor(() => {
      expect(result.current.data).toEqual(mockSchema2);
    });

    expect(mockHttpRequest).toHaveBeenCalledTimes(2);
  });

  it('should handle schema with complex properties', async () => {
    const mockSchema: ApiUserType = {
      id: 'schema-789',
      name: 'ComplexUser',
      schema: {
        name: {
          type: 'string',
          required: true,
        },
        age: {
          type: 'number',
          required: false,
        },
        isActive: {
          type: 'boolean',
          required: true,
        },
        roles: {
          type: 'array',
          items: {
            type: 'string',
          },
          required: false,
        },
        address: {
          type: 'object',
          properties: {
            street: {
              type: 'string',
              required: true,
            },
            city: {
              type: 'string',
              required: true,
            },
          },
          required: false,
        },
      },
    };

    mockHttpRequest.mockResolvedValueOnce({data: mockSchema});

    const {result} = renderHook(() => useGetUserType('schema-789'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(mockSchema);
    expect(result.current.error).toBeNull();
  });

  it('should handle refetch error and throw', async () => {
    const mockSchema: ApiUserType = {
      id: 'schema-123',
      name: 'Customer',
      schema: {
        name: {
          type: 'string',
          required: true,
        },
      },
    };

    mockHttpRequest.mockResolvedValue({data: mockSchema});

    const {result} = renderHook(() => useGetUserType('schema-123'));

    await waitFor(() => {
      expect(result.current.data).toEqual(mockSchema);
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
        description: 'Failed to fetch user type',
      });
    });

    expect(result.current.loading).toBe(false);
  });

  it('should handle non-Error object in refetch catch block', async () => {
    const mockSchema: ApiUserType = {
      id: 'schema-123',
      name: 'Customer',
      schema: {
        name: {
          type: 'string',
          required: true,
        },
      },
    };

    mockHttpRequest.mockResolvedValue({data: mockSchema});

    const {result} = renderHook(() => useGetUserType('schema-123'));

    await waitFor(() => {
      expect(result.current.data).toEqual(mockSchema);
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
        description: 'Failed to fetch user type',
      });
    });

    expect(result.current.loading).toBe(false);
  });

  it('should handle non-Error object during initial fetch', async () => {
    mockHttpRequest.mockRejectedValue('String error during initial fetch');

    const {result} = renderHook(() => useGetUserType('schema-123'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toEqual({
      code: 'FETCH_ERROR',
      message: 'An unknown error occurred',
      description: 'Failed to fetch user type',
    });
    expect(result.current.data).toBeNull();
  });
});
