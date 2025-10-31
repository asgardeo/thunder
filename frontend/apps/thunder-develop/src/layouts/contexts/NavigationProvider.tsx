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

import {useCallback, useEffect, useMemo, useState, type ReactNode} from 'react';
import {useLocation} from 'react-router';
import NavigationContext from './NavigationContext';

export default function NavigationProvider({children}: {children: ReactNode}) {
  const location = useLocation();
  const [currentPage, setCurrentPage] = useState<string>('home');
  const [sidebarOpen, setSidebarOpen] = useState(true);

  useEffect(() => {
    const path = location.pathname;

    // Parse the pathname to determine the current page
    const pathSegments = path.split('/').filter(Boolean);

    if (pathSegments.length === 0 || path === '/') {
      setCurrentPage('home');
    } else {
      const pageId = pathSegments[0];

      setCurrentPage(pageId);
    }
  }, [location.pathname]);

  const toggleSidebar = useCallback(() => {
    setSidebarOpen((prev) => !prev);
  }, []);

  const value = useMemo(
    () => ({
      currentPage,
      setCurrentPage,
      sidebarOpen,
      setSidebarOpen,
      toggleSidebar,
    }),
    [currentPage, sidebarOpen, toggleSidebar],
  );

  return <NavigationContext.Provider value={value}>{children}</NavigationContext.Provider>;
}
