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

import {useMemo} from 'react';
// eslint-disable-next-line import/extensions
import useLogger from './useLogger';
import type Logger from '../../../core/Logger';

/**
 * React hook for accessing a logger scoped to a specific component.
 *
 * This hook creates a scoped logger instance that automatically includes
 * the component name in all log entries. It's useful for tracking which
 * component generated specific log messages.
 *
 * The scoped logger is memoized based on the component name, so it won't
 * be recreated on every render unless the component name changes.
 *
 * @param componentName - The name of the component for scoping logs
 *
 * @returns A logger instance scoped to the component
 *
 * @throws {Error} Throws an error if used outside of LoggerProvider
 *
 * @example
 * Basic usage:
 * ```tsx
 * import useLoggerWithComponent from './useLoggerWithComponent';
 *
 * function UserProfile({ userId }: { userId: string }) {
 *   const logger = useLoggerWithComponent('UserProfile');
 *
 *   useEffect(() => {
 *     logger.info('Profile loaded', { userId });
 *     // Logs: { level: 'info', message: 'Profile loaded', component: 'UserProfile', context: { userId } }
 *   }, [userId]);
 *
 *   return <div>User Profile</div>;
 * }
 * ```
 *
 * @example
 * Multiple components with scoped loggers:
 * ```tsx
 * function Navbar() {
 *   const logger = useLoggerWithComponent('Navbar');
 *
 *   const handleMenuClick = () => {
 *     logger.debug('Menu clicked');
 *     // Logs with component: 'Navbar'
 *   };
 *
 *   return <nav onClick={handleMenuClick}>Menu</nav>;
 * }
 *
 * function Sidebar() {
 *   const logger = useLoggerWithComponent('Sidebar');
 *
 *   useEffect(() => {
 *     logger.info('Sidebar rendered');
 *     // Logs with component: 'Sidebar'
 *   }, []);
 *
 *   return <aside>Sidebar</aside>;
 * }
 * ```
 *
 * @public
 */
export default function useLoggerWithComponent(componentName: string): Logger {
  const logger = useLogger();

  return useMemo(() => logger.withComponent(componentName), [logger, componentName]);
}
