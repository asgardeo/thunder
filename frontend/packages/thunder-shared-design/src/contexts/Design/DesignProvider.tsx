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

import {useMemo, PropsWithChildren} from 'react';
import {useConfig} from '@thunder/shared-contexts';
import DesignContext, {type DesignContextType} from './DesignContext';
import useGetDesignResolve from '../../api/useGetDesignResolve';
import {DesignResolveType} from '../../models/design';

/**
 * Props for the DesignProvider component.
 *
 * @public
 */
export type DesignProviderProps = PropsWithChildren;

/**
 * React context provider component that provides Thunder design configuration
 * to all child components.
 *
 * This component loads design data from the server using the client UUID
 * and provides it through React context. Theme transformation is handled
 * at the hook level via useDesign().
 *
 * @param props - The component props
 * @param props.children - React children to be wrapped with the design context
 *
 * @returns JSX element that provides design context to children
 *
 * @public
 */
export default function DesignProvider({children}: DesignProviderProps) {
  const {getClientUuid} = useConfig();
  const clientUuid = getClientUuid();

  // Skip design resolution when no client UUID is available
  const shouldLoadDesign = Boolean(clientUuid && clientUuid.trim().length > 0);

  const {
    data: design,
    isLoading,
    error,
  } = useGetDesignResolve(
    {
      id: clientUuid ?? '',
      type: DesignResolveType.APP,
    },
    {
      enabled: shouldLoadDesign,
    },
  );

  const contextValue: DesignContextType = useMemo(
    () => ({
      design,
      isDesignEnabled: Boolean(design),
      isLoading,
      error,
      transformedTheme: undefined,
      theme: design?.theme,
      layout: design?.layout,
    }),
    [design, isLoading, error],
  );

  return <DesignContext.Provider value={contextValue}>{children}</DesignContext.Provider>;
}
