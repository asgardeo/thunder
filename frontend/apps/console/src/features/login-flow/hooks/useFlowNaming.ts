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

import {useCallback, useState} from 'react';

/**
 * Props for the useFlowNaming hook.
 */
export interface UseFlowNamingProps {
  /** The existing flow data loaded from the API. */
  existingFlowData?: {
    name?: string;
    handle?: string;
  };
  /** Default name for new flows. */
  defaultName?: string;
  /** Default handle for new flows. */
  defaultHandle?: string;
}

/**
 * Return type for the useFlowNaming hook.
 */
export interface UseFlowNamingReturn {
  /** Current flow name. */
  flowName: string;
  /** Current flow handle (URL-friendly identifier). */
  flowHandle: string;
  /** Whether the loaded flow needs auto-layout. */
  needsAutoLayout: boolean;
  /** Set whether auto-layout is needed. */
  setNeedsAutoLayout: React.Dispatch<React.SetStateAction<boolean>>;
  /** Handler for flow name changes - also updates the handle. */
  handleFlowNameChange: (newName: string) => void;
}

/**
 * Hook to manage flow naming (name and handle) state and logic.
 * Handles synchronization with existing flow data and generates
 * URL-friendly handles from flow names.
 *
 * @param props - Configuration options for the hook.
 * @returns Flow naming state and handlers.
 *
 * @example
 * ```tsx
 * const { flowName, flowHandle, handleFlowNameChange } = useFlowNaming({
 *   existingFlowData,
 *   defaultName: 'Login Flow',
 *   defaultHandle: 'login-flow',
 * });
 * ```
 */
const generateHandleFromName = (name: string): string =>
  name
    .toLowerCase()
    .trim()
    .replace(/\s+/g, '-')
    .replace(/[^a-z0-9-]/g, '')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '');

const useFlowNaming = (props?: UseFlowNamingProps): UseFlowNamingReturn => {
  const {existingFlowData, defaultName = 'Login Flow', defaultHandle = 'login-flow'} = props ?? {};

  const [flowName, setFlowName] = useState<string>(() => existingFlowData?.name ?? defaultName);
  const [flowHandle, setFlowHandle] = useState<string>(() => {
    if (existingFlowData?.handle) return existingFlowData.handle;
    if (existingFlowData?.name) return generateHandleFromName(existingFlowData.name);
    return defaultHandle;
  });
  const [needsAutoLayout, setNeedsAutoLayout] = useState<boolean>(false);

  // Sync flowName and flowHandle when existingFlowData changes after initial render
  const [prevExistingFlowData, setPrevExistingFlowData] = useState(existingFlowData);
  if (existingFlowData !== prevExistingFlowData) {
    setPrevExistingFlowData(existingFlowData);
    if (existingFlowData?.name) {
      setFlowName(existingFlowData.name);
    }
    const handle = existingFlowData?.handle;
    if (handle) {
      setFlowHandle(handle);
    } else if (existingFlowData?.name) {
      setFlowHandle(generateHandleFromName(existingFlowData.name));
    }
  }

  /**
   * Handler for flow name changes.
   * Updates both the name and generates a new handle.
   */
  const isExistingFlow = Boolean(existingFlowData?.handle);

  const handleFlowNameChange = useCallback(
    (newName: string) => {
      setFlowName(newName);
      if (!isExistingFlow) {
        setFlowHandle(generateHandleFromName(newName));
      }
    },
    [isExistingFlow],
  );

  return {
    flowName,
    flowHandle,
    needsAutoLayout,
    setNeedsAutoLayout,
    handleFlowNameChange,
  };
};

export default useFlowNaming;
