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
import UserTypeQueryKeys from '../../constants/userTypeQueryKeys';

// Mock the dependencies
vi.mock('@asgardeo/react', () => ({
  useAsgardeo: vi.fn(),
}));

vi.mock('@thunder/shared-contexts', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@thunder/shared-contexts')>();
  return {
    ...actual,
    useConfig: vi.fn(),
  };
});

const {useAsgardeo} = await import('@asgardeo/react');
const {useConfig} = await import('@thunder/shared-contexts');

describe('useGetUserType', () => {
  let mockHttpRequest: ReturnType<typeof vi.fn>;
  let mockGetServerUrl: ReturnType<typeof vi.fn>;

  const mockUserSchema: ApiUserSchema = {
    id: '123',
    name: 'Person',
    ouId: 'ou-1',
    allowSelfRegistration: true,
    schema: {
      email: {
        type: 'string',
        required: true,
      },
    },
  };

  beforeEach(() => {
    mockHttpRequest = vi.fn();
    mockGetServerUrl = vi.fn().mockReturnValue('https://api.test.com');

    vi.mocked(useAsgardeo).mockReturnValue({
      http: {
        request: mockHttpRequest,
      },
    } as unknown as ReturnType<typeof useAsgardeo>);

    vi.mocked(useConfig).mockReturnValue({
      getServerUrl: mockGetServerUrl,
    } as unknown as ReturnType<typeof useConfig>);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should initialize with loading state when id is provided', () => {
    mockHttpRequest.mockReturnValue(new Promise(() => {})); // Never resolves

    const {result} = renderHook(() => useGetUserType('123'));

    expect(result.current.isLoading).toBe(true);
    expect(result.current.data).toBeUndefined();
    expect(result.current.error).toBeNull();
  });

  it('should not fetch when no id is provided', () => {
    const {result} = renderHook(() => useGetUserType());

    expect(result.current.fetchStatus).toBe('idle');
    expect(mockHttpRequest).not.toHaveBeenCalled();
  });

  it('should not fetch when id is an empty string', () => {
    const {result} = renderHook(() => useGetUserType(''));

    expect(result.current.fetchStatus).toBe('idle');
    expect(mockHttpRequest).not.toHaveBeenCalled();
  });

  it('should successfully fetch a single user type', async () => {
    mockHttpRequest.mockResolvedValueOnce({
      data: mockUserSchema,
    });

    const {result} = renderHook(() => useGetUserType('123'));

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data).toEqual(mockUserSchema);
    expect(result.current.data?.id).toBe('123');
    expect(result.current.data?.name).toBe('Person');
  });

  it('should handle API error', async () => {
    const apiError = new Error('Failed to fetch user type');
    mockHttpRequest.mockRejectedValueOnce(apiError);

    const {result} = renderHook(() => useGetUserType('123'));

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });

    expect(result.current.error).toEqual(apiError);
    expect(result.current.data).toBeUndefined();
  });

  it('should use correct server URL and endpoint', async () => {
    mockHttpRequest.mockResolvedValueOnce({
      data: mockUserSchema,
    });

    renderHook(() => useGetUserType('123'));

    await waitFor(() => {
      expect(mockHttpRequest).toHaveBeenCalledTimes(1);
    });

    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({
        url: 'https://api.test.com/user-schemas/123',
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      }),
    );
  });

  it('should use correct query key', async () => {
    mockHttpRequest.mockResolvedValueOnce({
      data: mockUserSchema,
    });

    const {result, queryClient} = renderHook(() => useGetUserType('123'));

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    const queryKey = [UserTypeQueryKeys.USER_TYPE, '123'];
    const cachedData = queryClient.getQueryData(queryKey);
    expect(cachedData).toEqual(mockUserSchema);
  });

  it('should support refetching', async () => {
    mockHttpRequest.mockResolvedValueOnce({data: mockUserSchema});

    const {result} = renderHook(() => useGetUserType('123'));

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data?.name).toBe('Person');

    const updatedUserSchema: ApiUserSchema = {
      ...mockUserSchema,
      name: 'Updated Person',
    };

    mockHttpRequest.mockResolvedValueOnce({data: updatedUserSchema});

    await result.current.refetch();

    await waitFor(() => {
      expect(result.current.data?.name).toBe('Updated Person');
    });

    expect(mockHttpRequest).toHaveBeenCalledTimes(2);
  });
});
