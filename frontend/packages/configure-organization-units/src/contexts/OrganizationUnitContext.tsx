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
import type {OrganizationUnitTreeItem} from '../models/organization-unit-tree';

/**
 * Organization Unit Context Type
 *
 * Defines the shape of the organization unit context value.
 * Provides tree state management for the organization unit sidebar tree view.
 *
 * @internal
 * @remarks
 * This context is provided by {@link OrganizationUnitProvider} and consumed
 * via the {@link useOrganizationUnit} hook.
 */
export interface OrganizationUnitContextType {
  /** Current tree items displayed in the sidebar */
  treeItems: OrganizationUnitTreeItem[];
  /** Setter for tree items */
  setTreeItems: Dispatch<SetStateAction<OrganizationUnitTreeItem[]>>;
  /** IDs of currently expanded tree nodes */
  expandedItems: string[];
  /** Setter for expanded items */
  setExpandedItems: Dispatch<SetStateAction<string[]>>;
  /** Set of OU IDs whose children have been loaded */
  loadedItems: Set<string>;
  /** Setter for loaded items */
  setLoadedItems: Dispatch<SetStateAction<Set<string>>>;
  /** Clears treeItems and loadedItems, forcing a re-fetch. Preserves expandedItems so the tree re-expands after rebuild. */
  resetTreeState: () => void;
}

const OrganizationUnitContext: Context<OrganizationUnitContextType | null> =
  createContext<OrganizationUnitContextType | null>(null);

export default OrganizationUnitContext;
