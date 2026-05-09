/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

import type {Context, Dispatch, SetStateAction} from 'react';
import {createContext} from 'react';
import {type Base} from '../models/base';
import type {Claim} from '../models/metadata';
import type {Resource} from '../models/resources';

/**
 * Props interface of {@link InteractionContext}
 */
export interface InteractionContextProps {
  /**
   * The properties of the last interacted resource.
   */
  lastInteractedResource: Base;
  /**
   * The ID of the last user interacted resource node.
   */
  lastInteractedStepId: string;
  /**
   * Sets the latest interacted resource inside the canvas.
   * @param resource - The resource that was interacted with.
   * @param openPanel - Whether to open the properties panel. Defaults to true.
   */
  setLastInteractedResource: (resource: Resource, openPanel?: boolean) => void;
  /**
   * Sets the active element node ID.
   */
  setLastInteractedStepId: (stepId: string) => void;
  /**
   * Function to be called when an element is dropped on the canvas.
   * @param element - The element that was dropped on the canvas.
   * @param nodeId - The ID of the node on which the element was dropped.
   */
  onResourceDropOnCanvas: (element: Base, nodeId: string) => void;
  /**
   * The set of attributes that are selected for the flow that are maintained per node.
   */
  selectedAttributes: Record<string, Claim[]>;
  /**
   * Sets the selected attributes for the flow.
   */
  setSelectedAttributes: Dispatch<SetStateAction<Record<string, Claim[]>>>;
}

const InteractionContext: Context<InteractionContextProps | undefined> = createContext<
  InteractionContextProps | undefined
>(undefined);

InteractionContext.displayName = 'InteractionContext';

export default InteractionContext;
