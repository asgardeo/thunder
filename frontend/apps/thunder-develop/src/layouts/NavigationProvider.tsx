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
import type {NavigationItem} from './models/layouts';
import NavigationContext from './contexts/NavigationContext';

export default function NavigationProvider({children}: {children: ReactNode}) {
  const [currentPage, setCurrentPage] = useState<NavigationItem>({
    id: 'home',
    text: 'Home',
    category: 'Dashboard',
  });
  const [sidebarOpen, setSidebarOpen] = useState(true);

  useEffect(() => {
    const handleLocationChange = () => {
      const path = window.location.pathname;

      // Parse the pathname to determine the current page
      // This is a basic implementation - you may want to customize this logic
      const pathSegments = path.split('/').filter(Boolean);

      if (pathSegments.length === 0 || path === '/') {
        setCurrentPage({
          id: 'home',
          text: 'Home',
          category: 'Dashboard',
        });
      } else {
        const pageId = pathSegments[0];
        const pageText = pageId.charAt(0).toUpperCase() + pageId.slice(1);

        setCurrentPage({
          id: pageId,
          text: pageText,
          category: 'Dashboard',
        });
      }
    };

    // Handle initial load
    handleLocationChange();

    // Listen to popstate events (browser back/forward)
    window.addEventListener('popstate', handleLocationChange);

    // Listen to custom navigation events from react-router
    const originalPushState = window.history.pushState.bind(window.history);
    const originalReplaceState = window.history.replaceState.bind(window.history);

    window.history.pushState = function pushStateWrapper(...args) {
      originalPushState(...args);
      handleLocationChange();
    };

    window.history.replaceState = function replaceStateWrapper(...args) {
      originalReplaceState(...args);
      handleLocationChange();
    };

    return () => {
      window.removeEventListener('popstate', handleLocationChange);
      window.history.pushState = originalPushState;
      window.history.replaceState = originalReplaceState;
    };
  }, []);

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
