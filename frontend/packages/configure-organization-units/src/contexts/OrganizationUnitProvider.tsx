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

import {useState, useMemo, useCallback} from 'react';
import type {PropsWithChildren, JSX} from 'react';
import {Outlet} from 'react-router';
import OrganizationUnitContext from './OrganizationUnitContext';
import type {OrganizationUnitContextType} from './OrganizationUnitContext';
import type {OrganizationUnitTreeItem} from '../models/organization-unit-tree';

export default function OrganizationUnitProvider({children}: PropsWithChildren): JSX.Element {
  const [treeItems, setTreeItems] = useState<OrganizationUnitTreeItem[]>([]);
  const [expandedItems, setExpandedItems] = useState<string[]>([]);
  const [loadedItems, setLoadedItems] = useState<Set<string>>(new Set());

  const resetTreeState = useCallback(() => {
    setTreeItems([]);
    setLoadedItems(new Set());
    // expandedItems intentionally preserved so tree re-expands after rebuild
  }, []);

  const contextValue: OrganizationUnitContextType = useMemo(
    () => ({
      treeItems,
      setTreeItems,
      expandedItems,
      setExpandedItems,
      loadedItems,
      setLoadedItems,
      resetTreeState,
    }),
    [treeItems, expandedItems, loadedItems, resetTreeState],
  );

  return (
    <OrganizationUnitContext.Provider value={contextValue}>{children ?? <Outlet />}</OrganizationUnitContext.Provider>
  );
}
