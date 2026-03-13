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

import type {OrganizationUnitTreeItem} from '../models/organization-unit-tree';
import OrganizationUnitTreeConstants from '../constants/organization-unit-tree-constants';

export default function appendTreeItemChildren(
  items: OrganizationUnitTreeItem[],
  parentId: string,
  newChildren: OrganizationUnitTreeItem[],
): OrganizationUnitTreeItem[] {
  return items.map((item) => {
    if (item.id === parentId) {
      const existing = (item.children ?? []).filter((c) => !c.id.endsWith(OrganizationUnitTreeConstants.LOAD_MORE_SUFFIX));

      return {...item, children: [...existing, ...newChildren]};
    }

    if (item.children && item.children.length > 0) {
      return {...item, children: appendTreeItemChildren(item.children, parentId, newChildren)};
    }

    return item;
  });
}
