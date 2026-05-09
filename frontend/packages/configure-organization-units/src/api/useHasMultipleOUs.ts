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

import useGetChildOrganizationUnits from './useGetChildOrganizationUnits';
import useGetOrganizationUnits from './useGetOrganizationUnits';
import type {OrganizationUnit} from '../models/organization-unit';

interface UseHasMultipleOUsResult {
  hasMultipleOUs: boolean;
  isLoading: boolean;
  ouList: OrganizationUnit[];
}

export default function useHasMultipleOUs(): UseHasMultipleOUsResult {
  const {data: ouData, isLoading: isOuLoading} = useGetOrganizationUnits({limit: 2, offset: 0});
  const ouList = ouData?.organizationUnits ?? [];
  const rootCount = ouData?.totalResults ?? 0;
  const singleRootId = rootCount === 1 ? ouList[0]?.id : undefined;

  const {data: childData, isLoading: isChildLoading} = useGetChildOrganizationUnits(singleRootId, {
    limit: 1,
    offset: 0,
  });

  const hasMultipleRoots = rootCount > 1;
  const singleRootHasChildren = rootCount === 1 && (childData?.totalResults ?? 0) > 0;

  return {
    hasMultipleOUs: hasMultipleRoots || singleRootHasChildren,
    isLoading: isOuLoading || (rootCount === 1 && isChildLoading),
    ouList,
  };
}
