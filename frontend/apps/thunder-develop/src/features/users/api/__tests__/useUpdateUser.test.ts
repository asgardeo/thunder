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
import useUpdateUser, {type UpdateUserRequest} from '../useUpdateUser';
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

describe('useUpdateUser', () => {
  beforeEach(() => {
    mockHttpRequest.mockReset();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should initialize with correct default values', () => {
    const {result} = renderHook(() => useUpdateUser());

    expect(result.current.data).toBeNull();
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeNull();
    expect(typeof result.current.updateUser).toBe('function');
    expect(typeof result.current.reset).toBe('function');
  });

  it('should update a user successfully', async () => {
    const mockRequest: UpdateUserRequest = {
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Updated',
        email: 'john.updated@example.com',
      },
    };

    const mockResponse: ApiUser = {
      id: 'user-123',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Updated',
        email: 'john.updated@example.com',
      },
    };

    mockHttpRequest.mockResolvedValueOnce({data: mockResponse});

    const {result} = renderHook(() => useUpdateUser());

    await result.current.updateUser('user-123', mockRequest);

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.data).toEqual(mockResponse);
      expect(result.current.error).toBeNull();
    });

    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/users/user-123',
        method: 'PUT',
        data: mockRequest,
      }),
    );
  });

  it('should handle API error with JSON response', async () => {
    const mockRequest: UpdateUserRequest = {
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Updated',
        email: 'john.updated@example.com',
      },
    };

    mockHttpRequest.mockRejectedValueOnce(new Error('Validation failed'));

    const {result} = renderHook(() => useUpdateUser());

    try {
      await result.current.updateUser('user-123', mockRequest);
    } catch {
      // Expected to throw
    }

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toEqual({
        code: 'UPDATE_USER_ERROR',
        message: 'Validation failed',
        description: 'Failed to update user',
      });
      expect(result.current.data).toBeNull();
    });
  });

  it('should handle API error without JSON response', async () => {
    const mockRequest: UpdateUserRequest = {
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Updated',
        email: 'john.updated@example.com',
      },
    };

    mockHttpRequest.mockRejectedValueOnce(new Error('Internal Server Error'));

    const {result} = renderHook(() => useUpdateUser());

    try {
      await result.current.updateUser('user-123', mockRequest);
    } catch {
      // Expected to throw
    }

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toEqual({
        code: 'UPDATE_USER_ERROR',
        message: 'Internal Server Error',
        description: 'Failed to update user',
      });
      expect(result.current.data).toBeNull();
    });
  });

  it('should handle network error', async () => {
    const mockRequest: UpdateUserRequest = {
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Updated',
        email: 'john.updated@example.com',
      },
    };

    mockHttpRequest.mockRejectedValueOnce(new Error('Network error'));

    const {result} = renderHook(() => useUpdateUser());

    try {
      await result.current.updateUser('user-123', mockRequest);
    } catch {
      // Expected to throw
    }

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toEqual({
        code: 'UPDATE_USER_ERROR',
        message: 'Network error',
        description: 'Failed to update user',
      });
    });
  });

  it('should set loading state correctly during request', async () => {
    const mockRequest: UpdateUserRequest = {
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Updated',
        email: 'john.updated@example.com',
      },
    };

    const mockResponse: ApiUser = {
      id: 'user-123',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Updated',
        email: 'john.updated@example.com',
      },
    };

    // Create a promise we can control
    let resolveRequest: (value: {data: ApiUser}) => void;
    const requestPromise = new Promise<{data: ApiUser}>((resolve) => {
      resolveRequest = resolve;
    });

    mockHttpRequest.mockReturnValueOnce(requestPromise);

    const {result} = renderHook(() => useUpdateUser());

    expect(result.current.loading).toBe(false);

    const promise = result.current.updateUser('user-123', mockRequest);

    // Loading should become true
    await waitFor(() => {
      expect(result.current.loading).toBe(true);
    });

    // Now resolve the request
    resolveRequest!({data: mockResponse});

    await promise;

    // Loading should become false after completion
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
  });

  it('should reset state correctly', async () => {
    const mockRequest: UpdateUserRequest = {
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Updated',
        email: 'john.updated@example.com',
      },
    };

    const mockResponse: ApiUser = {
      id: 'user-123',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Updated',
        email: 'john.updated@example.com',
      },
    };

    mockHttpRequest.mockResolvedValueOnce({data: mockResponse});

    const {result} = renderHook(() => useUpdateUser());

    await result.current.updateUser('user-123', mockRequest);

    await waitFor(() => {
      expect(result.current.data).toEqual(mockResponse);
    });

    await waitFor(() => {
      result.current.reset();
    });

    await waitFor(() => {
      expect(result.current.data).toBeNull();
      expect(result.current.error).toBeNull();
    });
  });

  it('should clear previous data when updating', async () => {
    const mockRequest1: UpdateUserRequest = {
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Updated',
        email: 'john.updated@example.com',
      },
    };

    const mockResponse1: ApiUser = {
      id: 'user-123',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Updated',
        email: 'john.updated@example.com',
      },
    };

    const mockRequest2: UpdateUserRequest = {
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'Jane Updated',
        email: 'jane.updated@example.com',
      },
    };

    const mockResponse2: ApiUser = {
      id: 'user-789',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'Jane Updated',
        email: 'jane.updated@example.com',
      },
    };

    mockHttpRequest.mockResolvedValueOnce({data: mockResponse1}).mockResolvedValueOnce({data: mockResponse2});

    const {result} = renderHook(() => useUpdateUser());

    await result.current.updateUser('user-123', mockRequest1);

    await waitFor(() => {
      expect(result.current.data).toEqual(mockResponse1);
    });

    await result.current.updateUser('user-789', mockRequest2);

    await waitFor(() => {
      expect(result.current.data).toEqual(mockResponse2);
    });
  });

  it('should handle partial user attribute updates', async () => {
    const mockRequest: UpdateUserRequest = {
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        email: 'newemail@example.com',
      },
    };

    const mockResponse: ApiUser = {
      id: 'user-123',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Doe',
        email: 'newemail@example.com',
      },
    };

    mockHttpRequest.mockResolvedValueOnce({data: mockResponse});

    const {result} = renderHook(() => useUpdateUser());

    await result.current.updateUser('user-123', mockRequest);

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    await waitFor(() => {
      expect(result.current.data).toEqual(mockResponse);
    });
    expect(result.current.error).toBeNull();
  });

  it('should handle non-Error rejection', async () => {
    const mockRequest: UpdateUserRequest = {
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Updated',
        email: 'john.updated@example.com',
      },
    };

    mockHttpRequest.mockRejectedValueOnce('String error');

    const {result} = renderHook(() => useUpdateUser());

    await expect(result.current.updateUser('user-123', mockRequest)).rejects.toBe('String error');

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'UPDATE_USER_ERROR',
        message: 'An unknown error occurred',
        description: 'Failed to update user',
      });
      expect(result.current.data).toBeNull();
      expect(result.current.loading).toBe(false);
    });
  });
});
