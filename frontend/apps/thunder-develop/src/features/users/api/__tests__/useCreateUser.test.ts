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
import useCreateUser, {type CreateUserRequest} from '../useCreateUser';
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

describe('useCreateUser', () => {
  beforeEach(() => {
    mockHttpRequest.mockReset();
    mockGetServerUrl.mockReturnValue('https://localhost:8090');
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should initialize with correct default values', () => {
    const {result} = renderHook(() => useCreateUser());

    expect(result.current.data).toBeNull();
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeNull();
    expect(typeof result.current.createUser).toBe('function');
    expect(typeof result.current.reset).toBe('function');
  });

  it('should create a user successfully', async () => {
    const mockRequest: CreateUserRequest = {
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Doe',
        email: 'john@example.com',
      },
    };

    const mockResponse: ApiUser = {
      id: 'user-456',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Doe',
        email: 'john@example.com',
      },
    };

    mockHttpRequest.mockResolvedValueOnce({data: mockResponse});

    const {result} = renderHook(() => useCreateUser());

    await result.current.createUser(mockRequest);

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.data).toEqual(mockResponse);
      expect(result.current.error).toBeNull();
    });
    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://localhost:8090/users',
        method: 'POST',
        data: mockRequest,
      }),
    );
  });

  it('should handle API error with JSON response', async () => {
    const mockRequest: CreateUserRequest = {
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Doe',
        email: 'john@example.com',
      },
    };

    mockHttpRequest.mockRejectedValueOnce(new Error('Validation failed'));

    const {result} = renderHook(() => useCreateUser());

    await expect(result.current.createUser(mockRequest)).rejects.toThrow('Validation failed');

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      // Error is caught and re-wrapped with CREATE_USER_ERROR code but preserves message
      expect(result.current.error).toEqual({
        code: 'CREATE_USER_ERROR',
        message: 'Validation failed',
        description: 'Failed to create user',
      });
      expect(result.current.data).toBeNull();
    });
  });

  it('should handle API error without JSON response', async () => {
    const mockRequest: CreateUserRequest = {
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Doe',
        email: 'john@example.com',
      },
    };

    mockHttpRequest.mockRejectedValueOnce(new Error('Internal Server Error'));

    const {result} = renderHook(() => useCreateUser());

    try {
      await result.current.createUser(mockRequest);
    } catch {
      // Expected to throw
    }

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toEqual({
        code: 'CREATE_USER_ERROR',
        message: 'Internal Server Error',
        description: 'Failed to create user',
      });
      expect(result.current.data).toBeNull();
    });
  });

  it('should handle network error', async () => {
    const mockRequest: CreateUserRequest = {
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Doe',
        email: 'john@example.com',
      },
    };

    mockHttpRequest.mockRejectedValueOnce(new Error('Network error'));

    const {result} = renderHook(() => useCreateUser());

    await expect(result.current.createUser(mockRequest)).rejects.toThrow('Network error');

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toEqual({
        code: 'CREATE_USER_ERROR',
        message: 'Network error',
        description: 'Failed to create user',
      });
    });
  });

  it('should set loading state correctly during request', async () => {
    const mockRequest: CreateUserRequest = {
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Doe',
        email: 'john@example.com',
      },
    };

    const mockResponse: ApiUser = {
      id: 'user-456',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Doe',
        email: 'john@example.com',
      },
    };

    // Create a promise we can control
    let resolveRequest: (value: {data: ApiUser}) => void;
    const requestPromise = new Promise<{data: ApiUser}>((resolve) => {
      resolveRequest = resolve;
    });

    mockHttpRequest.mockReturnValueOnce(requestPromise);

    const {result} = renderHook(() => useCreateUser());

    expect(result.current.loading).toBe(false);

    // Start the request without awaiting
    const createUserPromise = result.current.createUser(mockRequest);

    // Loading should become true immediately after calling createUser
    await waitFor(() => {
      expect(result.current.loading).toBe(true);
    });

    // Now resolve the request
    resolveRequest!({data: mockResponse});

    // Wait for the request to complete
    await createUserPromise;

    // Loading should become false after completion
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
      expect(result.current.data).toEqual(mockResponse);
    });
  });

  it('should reset state correctly', async () => {
    const mockRequest: CreateUserRequest = {
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Doe',
        email: 'john@example.com',
      },
    };

    const mockResponse: ApiUser = {
      id: 'user-456',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Doe',
        email: 'john@example.com',
      },
    };

    mockHttpRequest.mockResolvedValueOnce({data: mockResponse});

    const {result} = renderHook(() => useCreateUser());

    await result.current.createUser(mockRequest);

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

  it('should clear previous data when creating new user', async () => {
    const mockRequest1: CreateUserRequest = {
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Doe',
        email: 'john@example.com',
      },
    };

    const mockResponse1: ApiUser = {
      id: 'user-456',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Doe',
        email: 'john@example.com',
      },
    };

    const mockRequest2: CreateUserRequest = {
      organizationUnit: '/marketing',
      type: 'employee',
      attributes: {
        name: 'Jane Smith',
        email: 'jane@example.com',
      },
    };

    const mockResponse2: ApiUser = {
      id: 'user-999',
      organizationUnit: '/marketing',
      type: 'employee',
      attributes: {
        name: 'Jane Smith',
        email: 'jane@example.com',
      },
    };

    mockHttpRequest.mockResolvedValueOnce({data: mockResponse1}).mockResolvedValueOnce({data: mockResponse2});

    const {result} = renderHook(() => useCreateUser());

    await result.current.createUser(mockRequest1);

    await waitFor(() => {
      expect(result.current.data).toEqual(mockResponse1);
    });

    await result.current.createUser(mockRequest2);

    await waitFor(() => {
      expect(result.current.data).toEqual(mockResponse2);
    });
  });

  it('should handle non-Error rejection', async () => {
    const mockRequest: CreateUserRequest = {
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Doe',
        email: 'john@example.com',
      },
    };

    mockHttpRequest.mockRejectedValueOnce('String error');

    const {result} = renderHook(() => useCreateUser());

    await expect(result.current.createUser(mockRequest)).rejects.toBe('String error');

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'CREATE_USER_ERROR',
        message: 'An unknown error occurred',
        description: 'Failed to create user',
      });
      expect(result.current.data).toBeNull();
      expect(result.current.loading).toBe(false);
    });
  });
});

describe('useCreateUser with fallback URL', () => {
  const FALLBACK_BASE_URL = 'https://fallback-api.example.com';

  beforeEach(() => {
    mockHttpRequest.mockReset();
    mockGetServerUrl.mockReturnValue(undefined);
    import.meta.env.VITE_ASGARDEO_BASE_URL = FALLBACK_BASE_URL;
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should use VITE_ASGARDEO_BASE_URL when getServerUrl returns undefined', async () => {
    const mockRequest: CreateUserRequest = {
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Doe',
        email: 'john@example.com',
      },
    };

    const mockResponse: ApiUser = {
      id: 'user-456',
      organizationUnit: '/sales',
      type: 'customer',
      attributes: {
        name: 'John Doe',
        email: 'john@example.com',
      },
    };

    mockHttpRequest.mockResolvedValueOnce({data: mockResponse});

    const {result} = renderHook(() => useCreateUser());

    await result.current.createUser(mockRequest);

    await waitFor(() => {
      expect(result.current.data).toEqual(mockResponse);
    });

    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: `${FALLBACK_BASE_URL}/users`,
        method: 'POST',
      }),
    );
  });
});
