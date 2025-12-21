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

import {useEffect, useRef, type ReactElement} from 'react';
import {Handle, useNodeId, useUpdateNodeInternals, type HandleProps} from '@xyflow/react';

/**
 * Props interface of {@link NodeHandle}
 */
export interface NodeHandlePropsInterface extends HandleProps {
  /**
   * Optional key/index to track position changes for reordering scenarios.
   * When this changes, the component will notify React Flow to update handle positions.
   * @defaultValue undefined
   */
  positionKey?: string | number;
}

/**
 * A wrapper around React Flow's Handle component that automatically calls
 * updateNodeInternals when the handle position changes (e.g., after reordering).
 *
 * This solves the issue where edges don't update their positions when handles
 * move within a custom node (such as when reordering elements in a View).
 *
 * This component has been optimized to remove the expensive MutationObserver.
 * Instead, it only updates when positionKey changes (explicit trigger from parent).
 *
 * @param props - Props injected to the component.
 * @returns NodeHandle component.
 */
function NodeHandle({positionKey = undefined, ...handleProps}: NodeHandlePropsInterface): ReactElement {
  const nodeId = useNodeId();
  const updateNodeInternals = useUpdateNodeInternals();
  const prevPositionKeyRef = useRef<string | number | undefined>(positionKey);

  // Removed expensive MutationObserver that watched entire node subtree
  useEffect(() => {
    if (!nodeId) return;

    // Only trigger update if positionKey actually changed (not on initial mount)
    if (prevPositionKeyRef.current !== undefined && prevPositionKeyRef.current !== positionKey) {
      // Use RAF to batch with other updates and ensure DOM is settled
      requestAnimationFrame(() => {
        updateNodeInternals(nodeId);
      });
    }

    prevPositionKeyRef.current = positionKey;
  }, [nodeId, positionKey, updateNodeInternals]);

  return <Handle {...handleProps} />;
}

export default NodeHandle;
