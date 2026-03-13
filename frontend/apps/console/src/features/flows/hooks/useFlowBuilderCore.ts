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

import {useContext} from 'react';
import FlowBuilderCoreContext, {type FlowBuilderCoreContextProps} from '../context/FlowBuilderCoreContext';

/**
 * Props interface of {@link useFlowBuilderCore}
 */
export type UseFlowBuilderCoreInterface = FlowBuilderCoreContextProps;

/**
 * Hook that provides access to the Flow Builder Core context.
 *
 * This hook allows components to access flow builder core-related data and functions
 * provided by the {@link FlowBuilderCoreProvider}. It returns an object containing
 * the context values defined in {@link FlowBuilderCoreContext}.
 *
 * @returns An object containing the context values of {@link FlowBuilderCoreContext}.
 *
 * @throws Will throw an error if the hook is used outside of a FlowBuilderCoreProvider.
 *
 * @example
 * ```tsx
 * const { setIsResourcePanelOpen } = useFlowBuilderCore();
 * ```
 */
const useFlowBuilderCore = (): UseFlowBuilderCoreInterface => {
  const context: FlowBuilderCoreContextProps = useContext(FlowBuilderCoreContext);

  if (context === undefined) {
    throw new Error('useFlowBuilderCore must be used within a FlowBuilderCoreProvider');
  }

  return context;
};

export default useFlowBuilderCore;
