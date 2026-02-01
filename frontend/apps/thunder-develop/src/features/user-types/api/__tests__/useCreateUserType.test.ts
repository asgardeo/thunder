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
import useCreateUserType from '../useCreateUserType';
import type {ApiUserSchema, CreateUserSchemaRequest} from '../../types/user-types';

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

describe('useCreateUserType', () => {
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
    },
  };

  const mockRequest: CreateUserSchemaRequest = {
    name: 'TestUserType',
    ouId: 'root-ou',
    allowSelfRegistration: true,
    schema: {
      username: {
        type: 'string',
        required: true,
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

  it('should initialize with default state', () => {
    const {result} = renderHook(() => useCreateUserType());

    expect(result.current.data).toBeNull();
    expect(result.current.error).toBeNull();
    expect(result.current.loading).toBe(false);
    expect(typeof result.current.createUserType).toBe('function');
    expect(typeof result.current.reset).toBe('function');
  });

  it('should successfully create a user type', async () => {
    mockHttpRequest.mockResolvedValueOnce({data: mockUserSchema});

    const {result} = renderHook(() => useCreateUserType());

    await result.current.createUserType(mockRequest);

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchema);
      expect(result.current.error).toBeNull();
      expect(result.current.loading).toBe(false);
    });

    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/user-schemas',
        method: 'POST',
        data: mockRequest,
      }),
    );
  });

  it('should set loading state during creation', async () => {
    // Create a promise we can control
    let resolveRequest: (value: {data: ApiUserSchema}) => void;
    const requestPromise = new Promise<{data: ApiUserSchema}>((resolve) => {
      resolveRequest = resolve;
    });

    mockHttpRequest.mockReturnValueOnce(requestPromise);

    const {result} = renderHook(() => useCreateUserType());

    const promise = result.current.createUserType(mockRequest);

    await waitFor(() => {
      expect(result.current.loading).toBe(true);
    });

    // Now resolve the request
    resolveRequest!({data: mockUserSchema});

    await promise;

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
  });

  it('should handle API error', async () => {
    mockHttpRequest.mockRejectedValueOnce(new Error('Validation failed'));

    const {result} = renderHook(() => useCreateUserType());

    await expect(result.current.createUserType(mockRequest)).rejects.toThrow('Validation failed');

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'CREATE_USER_TYPE_ERROR',
        message: 'Validation failed',
        description: 'Failed to create user type',
      });
      expect(result.current.data).toBeNull();
      expect(result.current.loading).toBe(false);
    });
  });

  it('should handle network error', async () => {
    mockHttpRequest.mockRejectedValueOnce(new Error('Network error'));

    const {result} = renderHook(() => useCreateUserType());

    await expect(result.current.createUserType(mockRequest)).rejects.toThrow('Network error');

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'CREATE_USER_TYPE_ERROR',
        message: 'Network error',
        description: 'Failed to create user type',
      });
      expect(result.current.data).toBeNull();
      expect(result.current.loading).toBe(false);
    });
  });

  it('should reset state when reset is called', async () => {
    mockHttpRequest.mockResolvedValueOnce({data: mockUserSchema});

    const {result} = renderHook(() => useCreateUserType());

    await result.current.createUserType(mockRequest);

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchema);
    });

    result.current.reset();

    await waitFor(() => {
      expect(result.current.data).toBeNull();
      expect(result.current.error).toBeNull();
    });
  });

  it('should clear previous data and error when starting new request', async () => {
    mockHttpRequest
      .mockResolvedValueOnce({data: mockUserSchema})
      .mockResolvedValueOnce({data: {...mockUserSchema, id: '456'}});

    const {result} = renderHook(() => useCreateUserType());

    await result.current.createUserType(mockRequest);

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchema);
    });

    await result.current.createUserType(mockRequest);

    await waitFor(() => {
      expect(result.current.data).toEqual({...mockUserSchema, id: '456'});
      expect(result.current.error).toBeNull();
    });
  });

  it('should handle non-Error rejection', async () => {
    mockHttpRequest.mockRejectedValueOnce('String error');

    const {result} = renderHook(() => useCreateUserType());

    await expect(result.current.createUserType(mockRequest)).rejects.toBe('String error');

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'CREATE_USER_TYPE_ERROR',
        message: 'An unknown error occurred',
        description: 'Failed to create user type',
      });
      expect(result.current.data).toBeNull();
      expect(result.current.loading).toBe(false);
    });
  });

  it('should fallback to env variable when getServerUrl returns undefined', async () => {
    mockGetServerUrl.mockReturnValue(undefined);
    mockHttpRequest.mockResolvedValueOnce({data: mockUserSchema});

    const {result} = renderHook(() => useCreateUserType());

    await result.current.createUserType(mockRequest);

    await waitFor(() => {
      expect(result.current.data).toEqual(mockUserSchema);
    });

    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: expect.stringContaining('/user-schemas') as string,
        method: 'POST',
      }),
    );
  });
});
