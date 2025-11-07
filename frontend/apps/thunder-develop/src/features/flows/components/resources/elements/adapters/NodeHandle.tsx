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

import {useEffect, useRef, useCallback, type ReactElement} from 'react';
import {Handle, useNodeId, useUpdateNodeInternals, type HandleProps} from '@xyflow/react';

/**
 * Props interface of {@link NodeHandle}
 */
export interface NodeHandlePropsInterface extends HandleProps {
  /**
   * Optional key/index to track position changes for reordering scenarios.
   * When this changes, the component will notify React Flow to update handle positions.
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
 * @param props - Props injected to the component.
 * @returns NodeHandle component.
 */
function NodeHandle({positionKey, ...handleProps}: NodeHandlePropsInterface): ReactElement {
  const nodeId = useNodeId();
  const updateNodeInternals = useUpdateNodeInternals();
  const handleRef = useRef<HTMLDivElement>(null);
  const lastPositionRef = useRef<{top: number; left: number} | null>(null);
  const updateScheduledRef = useRef<boolean>(false);

  // Memoized function to check position and update if changed
  const checkAndUpdatePosition = useCallback(() => {
    if (!handleRef.current || !nodeId || updateScheduledRef.current) return;

    const rect = handleRef.current.getBoundingClientRect();
    const currentPosition = {top: Math.round(rect.top), left: Math.round(rect.left)};

    // Check if position has actually changed (with tolerance for sub-pixel differences)
    const hasPositionChanged =
      lastPositionRef.current !== null &&
      (Math.abs(lastPositionRef.current.top - currentPosition.top) > 1 ||
        Math.abs(lastPositionRef.current.left - currentPosition.left) > 1);

    if (hasPositionChanged) {
      // Prevent multiple updates in quick succession
      updateScheduledRef.current = true;

      // Use setTimeout to batch updates and let DOM settle
      setTimeout(() => {
        updateNodeInternals(nodeId);
        updateScheduledRef.current = false;
      }, 0);
    }

    lastPositionRef.current = currentPosition;
  }, [nodeId, updateNodeInternals]);

  // Use MutationObserver to detect DOM changes that might affect handle position
  useEffect(() => {
    if (!handleRef.current || !nodeId) return;

    // Find the closest React Flow node element to observe
    const nodeElement = handleRef.current.closest('.react-flow__node');
    if (!nodeElement) return;

    // Create observer to watch for childList and subtree changes
    const observer = new MutationObserver(() => {
      // When DOM changes, check if our position changed
      requestAnimationFrame(() => {
        checkAndUpdatePosition();
      });
    });

    // Observe the node element for changes in its subtree
    observer.observe(nodeElement, {
      childList: true,
      subtree: true,
      attributes: true,
      attributeFilter: ['style', 'class'],
    });

    // Initial position capture
    const rect = handleRef.current.getBoundingClientRect();
    lastPositionRef.current = {top: Math.round(rect.top), left: Math.round(rect.left)};

    return () => {
      observer.disconnect();
    };
  }, [nodeId, checkAndUpdatePosition]);

  // Also check position when positionKey changes (explicit trigger)
  useEffect(() => {
    if (!nodeId) return;

    // Use multiple frames to ensure DOM has fully settled
    const timeoutId = setTimeout(() => {
      requestAnimationFrame(() => {
        checkAndUpdatePosition();
      });
    }, 50);

    return () => clearTimeout(timeoutId);
  }, [nodeId, positionKey, checkAndUpdatePosition]);

  return (
    <Handle
      ref={handleRef}
      {...handleProps}
    />
  );
}

export default NodeHandle;
