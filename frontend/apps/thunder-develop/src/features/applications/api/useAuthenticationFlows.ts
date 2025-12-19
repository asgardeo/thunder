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

import {useMemo, useCallback} from 'react';
import useGetFlows from '@/features/flows/api/useGetFlows';
import type {BasicFlowDefinition} from '@/features/flows/models/responses';

/**
 * Maximum number of flows to fetch. This limit should be sufficient for most
 * deployments.
 */
const FLOW_LIMIT = 100;

/**
 * Return type for the useAuthenticationFlows hook.
 */
export interface UseAuthenticationFlowsResult {
  /**
   * Array of all authentication flows
   */
  flows: BasicFlowDefinition[];

  /**
   * Whether the flows are currently being loaded
   */
  isLoading: boolean;

  /**
   * Error that occurred while fetching flows, if any
   */
  error: Error | null;

  /**
   * Get the flow UUID by its handle
   * @param handle - The flow handle (e.g., 'default-basic-flow')
   * @returns The flow UUID if found, undefined otherwise
   */
  getFlowIdByHandle: (handle: string) => string | undefined;

  /**
   * Check if a flow with the given handle exists
   * @param handle - The flow handle to check
   * @returns true if the flow exists, false otherwise
   */
  isFlowAvailable: (handle: string) => boolean;
}

/**
 * Custom hook to fetch authentication flows and provide handle-to-ID lookup.
 *
 * This hook fetches authentication flows from the API and provides utility
 * functions to look up flow UUIDs by their handles and check flow availability.
 *
 * @returns Object containing flows, loading state, error, and utility functions
 *
 * @example
 * ```tsx
 * function ApplicationCreate() {
 *   const { getFlowIdByHandle, isFlowAvailable, isLoading } = useAuthenticationFlows();
 *
 *   if (isLoading) return <div>Loading...</div>;
 *
 *   // Check if a flow is available
 *   if (!isFlowAvailable('default-basic-flow')) {
 *     console.log('Basic flow not available');
 *   }
 *
 *   // Get the UUID for a flow handle
 *   const flowId = getFlowIdByHandle('default-basic-flow');
 *   console.log('Flow UUID:', flowId);
 * }
 * ```
 */
export default function useAuthenticationFlows(): UseAuthenticationFlowsResult {
  const {data, isLoading, error} = useGetFlows({
    flowType: 'AUTHENTICATION',
    limit: FLOW_LIMIT,
  });

  /**
   * Map of flow handles to their UUIDs for quick lookup
   */
  const flowMap = useMemo((): Map<string, string> => {
    const map = new Map<string, string>();
    data?.flows?.forEach((flow: BasicFlowDefinition) => {
      map.set(flow.handle, flow.id);
    });
    return map;
  }, [data]);

  /**
   * Get the flow UUID by its handle
   */
  const getFlowIdByHandle = useCallback(
    (handle: string): string | undefined => flowMap.get(handle),
    [flowMap],
  );

  /**
   * Check if a flow with the given handle exists
   */
  const isFlowAvailable = useCallback((handle: string): boolean => flowMap.has(handle), [flowMap]);

  return {
    flows: data?.flows ?? [],
    isLoading,
    error: error ?? null,
    getFlowIdByHandle,
    isFlowAvailable,
  };
}
