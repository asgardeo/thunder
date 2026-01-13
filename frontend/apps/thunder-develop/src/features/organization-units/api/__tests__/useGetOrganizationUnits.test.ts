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
import {waitFor} from '@testing-library/react';
import {renderHook} from '../../../../test/test-utils';
import useGetOrganizationUnits from '../useGetOrganizationUnits';
import type {OrganizationUnitListResponse} from '../../types/organization-units';

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
vi.mock('@thunder/commons-contexts', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@thunder/commons-contexts')>();
  return {
    ...actual,
    useConfig: () => ({
      getServerUrl: () => 'https://localhost:8090',
    }),
  };
});

describe('useGetOrganizationUnits', () => {
  const mockOrganizationUnitList: OrganizationUnitListResponse = {
    totalResults: 2,
    startIndex: 1,
    count: 2,
    organizationUnits: [
      {id: 'ou-1', handle: 'root', name: 'Root Organization', description: 'Root OU', parent: null},
      {id: 'ou-2', handle: 'child', name: 'Child Organization', description: 'Child OU', parent: 'ou-1'},
    ],
    links: [{rel: 'self', href: 'https://localhost:8090/organization-units'}],
  };

  beforeEach(() => {
    mockHttpRequest.mockReset();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('should initialize and start fetching', async () => {
    mockHttpRequest.mockResolvedValue({data: mockOrganizationUnitList});

    const {result} = renderHook(() => useGetOrganizationUnits());

    expect(result.current.data).toBeNull();
    expect(result.current.error).toBeNull();
    expect(typeof result.current.refetch).toBe('function');

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });
  });

  it('should fetch organization units on mount', async () => {
    mockHttpRequest.mockResolvedValue({data: mockOrganizationUnitList});

    const {result} = renderHook(() => useGetOrganizationUnits());

    await waitFor(() => {
      expect(result.current.data).toEqual(mockOrganizationUnitList);
      expect(result.current.error).toBeNull();
      expect(result.current.loading).toBe(false);
    });

    expect(mockHttpRequest).toHaveBeenCalledWith(
      expect.objectContaining({url: 'https://localhost:8090/organization-units', method: 'GET'}),
    );
  });

  it('should fetch organization units with limit parameter', async () => {
    mockHttpRequest.mockResolvedValue({data: mockOrganizationUnitList});

    renderHook(() => useGetOrganizationUnits({limit: 10}));

    await waitFor(() => {
      expect(
        mockHttpRequest.mock.calls.some(
          (call: unknown[]) =>
            (call[0] as {url?: string})?.url === 'https://localhost:8090/organization-units?limit=10',
        ),
      ).toBe(true);
    });
  });

  it('should fetch organization units with offset parameter', async () => {
    mockHttpRequest.mockResolvedValue({data: mockOrganizationUnitList});

    renderHook(() => useGetOrganizationUnits({offset: 5}));

    await waitFor(() => {
      expect(mockHttpRequest).toHaveBeenCalledWith(
        expect.objectContaining({
          url: 'https://localhost:8090/organization-units?offset=5',
          method: 'GET',
        }),
      );
    });
  });

  it('should fetch organization units with both limit and offset parameters', async () => {
    mockHttpRequest.mockResolvedValue({data: mockOrganizationUnitList});

    renderHook(() => useGetOrganizationUnits({limit: 10, offset: 5}));

    await waitFor(() => {
      expect(mockHttpRequest).toHaveBeenCalledWith(
        expect.objectContaining({
          url: 'https://localhost:8090/organization-units?limit=10&offset=5',
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

    const {result, unmount} = renderHook(() => useGetOrganizationUnits());

    await waitFor(() => {
      expect(result.current.loading).toBe(true);
    });

    unmount();
  });

  it('should handle API error with Error instance', async () => {
    mockHttpRequest.mockRejectedValue(new Error('Failed to fetch organization units'));

    const {result} = renderHook(() => useGetOrganizationUnits());

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'FETCH_ERROR',
        message: 'Failed to fetch organization units',
        description: 'Failed to fetch organization units',
      });
      expect(result.current.data).toBeNull();
      expect(result.current.loading).toBe(false);
    });
  });

  it('should handle non-Error thrown values', async () => {
    mockHttpRequest.mockRejectedValue('String error');

    const {result} = renderHook(() => useGetOrganizationUnits());

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'FETCH_ERROR',
        message: 'An unknown error occurred',
        description: 'Failed to fetch organization units',
      });
      expect(result.current.data).toBeNull();
      expect(result.current.loading).toBe(false);
    });
  });

  it('should handle network error', async () => {
    mockHttpRequest.mockRejectedValue(new Error('Network error'));

    const {result} = renderHook(() => useGetOrganizationUnits());

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'FETCH_ERROR',
        message: 'Network error',
        description: 'Failed to fetch organization units',
      });
      expect(result.current.data).toBeNull();
      expect(result.current.loading).toBe(false);
    });
  });

  it('should refetch when refetch is called', async () => {
    mockHttpRequest.mockResolvedValue({data: mockOrganizationUnitList});

    const {result} = renderHook(() => useGetOrganizationUnits());

    await waitFor(() => {
      expect(result.current.data).toEqual(mockOrganizationUnitList);
    });

    const updatedList = {...mockOrganizationUnitList, totalResults: 3};
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
              data: mockOrganizationUnitList,
            }),
          100,
        );
      });
    });

    const {rerender} = renderHook(({params}) => useGetOrganizationUnits(params), {
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

    const {result} = renderHook(() => useGetOrganizationUnits());

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
              data: mockOrganizationUnitList,
            }),
          100,
        );
      });
    });

    const {unmount} = renderHook(() => useGetOrganizationUnits());

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
    mockHttpRequest.mockResolvedValue({data: mockOrganizationUnitList});

    const {rerender} = renderHook(({params}) => useGetOrganizationUnits(params), {
      initialProps: {params: {limit: 10}},
    });

    await waitFor(() => {
      expect(mockHttpRequest).toHaveBeenCalledWith(
        expect.objectContaining({
          url: 'https://localhost:8090/organization-units?limit=10',
          method: 'GET',
        }),
      );
    });

    rerender({params: {limit: 20}});

    await waitFor(() => {
      expect(
        mockHttpRequest.mock.calls.some(
          (call: unknown[]) =>
            (call[0] as {url?: string})?.url === 'https://localhost:8090/organization-units?limit=20',
        ),
      ).toBe(true);
    });
  });

  it('should refetch with new params when provided to refetch', async () => {
    mockHttpRequest.mockResolvedValue({data: mockOrganizationUnitList});

    const {result} = renderHook(() => useGetOrganizationUnits({limit: 10}));

    await waitFor(() => {
      expect(result.current.data).toEqual(mockOrganizationUnitList);
    });

    const updatedList = {...mockOrganizationUnitList, totalResults: 5};
    mockHttpRequest.mockResolvedValue({data: updatedList});

    await result.current.refetch({limit: 20, offset: 10});

    await waitFor(() => {
      expect(result.current.data).toEqual(updatedList);
    });

    expect(
      mockHttpRequest.mock.calls.some(
        (call: unknown[]) =>
          (call[0] as {url?: string})?.url === 'https://localhost:8090/organization-units?limit=20&offset=10',
      ),
    ).toBe(true);
  });

  it('should refetch with only limit param', async () => {
    mockHttpRequest.mockResolvedValue({data: mockOrganizationUnitList});

    const {result} = renderHook(() => useGetOrganizationUnits());

    await waitFor(() => {
      expect(result.current.data).toEqual(mockOrganizationUnitList);
    });

    mockHttpRequest.mockResolvedValue({data: mockOrganizationUnitList});

    await result.current.refetch({limit: 15});

    await waitFor(() => {
      expect(
        mockHttpRequest.mock.calls.some(
          (call: unknown[]) =>
            (call[0] as {url?: string})?.url === 'https://localhost:8090/organization-units?limit=15',
        ),
      ).toBe(true);
    });
  });

  it('should refetch with only offset param', async () => {
    mockHttpRequest.mockResolvedValue({data: mockOrganizationUnitList});

    const {result} = renderHook(() => useGetOrganizationUnits());

    await waitFor(() => {
      expect(result.current.data).toEqual(mockOrganizationUnitList);
    });

    mockHttpRequest.mockResolvedValue({data: mockOrganizationUnitList});

    await result.current.refetch({offset: 5});

    await waitFor(() => {
      expect(
        mockHttpRequest.mock.calls.some(
          (call: unknown[]) =>
            (call[0] as {url?: string})?.url === 'https://localhost:8090/organization-units?offset=5',
        ),
      ).toBe(true);
    });
  });

  it('should not throw error for aborted requests in refetch', async () => {
    mockHttpRequest.mockResolvedValue({data: mockOrganizationUnitList});

    const {result} = renderHook(() => useGetOrganizationUnits());

    await waitFor(() => {
      expect(result.current.data).toEqual(mockOrganizationUnitList);
    });

    const abortError = new Error('Aborted');
    abortError.name = 'AbortError';
    mockHttpRequest.mockRejectedValue(abortError);

    // Should not throw for AbortError
    await result.current.refetch();

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // Error should remain null for aborted requests
    expect(result.current.error).toBeNull();
    // Data should remain from previous successful fetch
    expect(result.current.data).toEqual(mockOrganizationUnitList);
  });

  it('should set error state when refetch fails with non-abort error', async () => {
    mockHttpRequest.mockResolvedValue({data: mockOrganizationUnitList});

    const {result} = renderHook(() => useGetOrganizationUnits());

    await waitFor(() => {
      expect(result.current.data).toEqual(mockOrganizationUnitList);
    });

    const error = new Error('Refetch failed');
    mockHttpRequest.mockRejectedValue(error);

    // refetch doesn't throw - it handles errors internally and sets error state
    await result.current.refetch();

    await waitFor(() => {
      expect(result.current.error).toEqual({
        code: 'FETCH_ERROR',
        message: 'Refetch failed',
        description: 'Failed to fetch organization units',
      });
    });
  });

  it('should clear error on successful fetch after error', async () => {
    mockHttpRequest.mockRejectedValueOnce(new Error('Initial error'));

    const {result} = renderHook(() => useGetOrganizationUnits());

    await waitFor(() => {
      expect(result.current.error).not.toBeNull();
    });

    mockHttpRequest.mockResolvedValue({data: mockOrganizationUnitList});

    await result.current.refetch();

    await waitFor(() => {
      expect(result.current.error).toBeNull();
      expect(result.current.data).toEqual(mockOrganizationUnitList);
    });
  });

  it('should use default params when no params provided', async () => {
    mockHttpRequest.mockResolvedValue({data: mockOrganizationUnitList});

    renderHook(() => useGetOrganizationUnits());

    await waitFor(() => {
      expect(mockHttpRequest).toHaveBeenCalledWith(
        expect.objectContaining({
          url: 'https://localhost:8090/organization-units',
          method: 'GET',
        }),
      );
    });
  });

  it('should handle undefined params in refetch', async () => {
    mockHttpRequest.mockResolvedValue({data: mockOrganizationUnitList});

    const {result} = renderHook(() => useGetOrganizationUnits({limit: 10}));

    await waitFor(() => {
      expect(result.current.data).toEqual(mockOrganizationUnitList);
    });

    // Call refetch without params - should use the hook's original params
    await result.current.refetch();

    await waitFor(() => {
      expect(
        mockHttpRequest.mock.calls.some(
          (call: unknown[]) =>
            (call[0] as {url?: string})?.url === 'https://localhost:8090/organization-units?limit=10',
        ),
      ).toBe(true);
    });
  });
});
