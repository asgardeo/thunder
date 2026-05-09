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

export default function buildItemMap(items: OrganizationUnitTreeItem[]): Map<string, OrganizationUnitTreeItem> {
  const map = new Map<string, OrganizationUnitTreeItem>();
  const visit = (list: OrganizationUnitTreeItem[]): void => {
    list.forEach((item) => {
      map.set(item.id, item);
      if (item.children) visit(item.children);
    });
  };
  visit(items);

  return map;
}
