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

import {useContext, useMemo} from 'react';
// eslint-disable-next-line import/extensions
import LoggerContext from './LoggerContext';
import type Logger from '../../../core/Logger';

/**
 * React hook for accessing the logger instance throughout the application.
 *
 * This hook provides access to the logger configured via LoggerProvider.
 * It must be used within a component tree wrapped by `LoggerProvider`,
 * otherwise it will throw an error.
 *
 * Optionally accepts a component name to automatically scope all logs
 * to that component. This is useful for tracking which component generated
 * specific log messages.
 *
 * @param componentName - Optional component name for scoping logs
 *
 * @returns The logger instance (scoped if component name provided)
 *
 * @throws {Error} Throws an error if used outside of LoggerProvider
 *
 * @example
 * Basic usage without scoping:
 * ```tsx
 * import {useLogger} from '@thunder/logger/react';
 *
 * function MyComponent() {
 *   const logger = useLogger();
 *
 *   useEffect(() => {
 *     logger.info('Component mounted');
 *     return () => logger.debug('Component unmounting');
 *   }, []);
 *
 *   const handleClick = () => {
 *     logger.info('Button clicked', { buttonId: 'submit' });
 *   };
 *
 *   return <button onClick={handleClick}>Submit</button>;
 * }
 * ```
 *
 * @example
 * Usage with component scoping:
 * ```tsx
 * import {useLogger} from '@thunder/logger/react';
 *
 * function UserProfile({ userId }: { userId: string }) {
 *   const logger = useLogger('UserProfile');
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
 * Logging with error handling:
 * ```tsx
 * function DataFetcher() {
 *   const logger = useLogger('DataFetcher');
 *
 *   const fetchData = async () => {
 *     try {
 *       logger.debug('Fetching data...');
 *       const response = await fetch('/api/data');
 *       logger.info('Data fetched successfully', { status: response.status });
 *     } catch (error) {
 *       logger.error('Failed to fetch data', { error });
 *     }
 *   };
 *
 *   return <button onClick={fetchData}>Fetch</button>;
 * }
 * ```
 *
 * @public
 */
export default function useLogger(componentName?: string): Logger {
  const logger = useContext(LoggerContext);
  if (!logger) {
    throw new Error('useLogger must be used within a LoggerProvider');
  }

  return useMemo(() => {
    if (componentName) {
      return logger.withComponent(componentName);
    }
    return logger;
  }, [logger, componentName]);
}
