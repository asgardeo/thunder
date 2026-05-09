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

import {Context, createContext} from 'react';
import type Logger from '../../../core/Logger';

/**
 * Logger context interface that provides access to the logger instance.
 *
 * @public
 */
export type LoggerContextType = Logger | null;

/**
 * React context for accessing the logger instance throughout the application.
 *
 * This context provides access to the logger configured via the LoggerProvider.
 * It should be used within a `LoggerProvider` component.
 *
 * @example
 * ```tsx
 * import LoggerContext from './LoggerContext';
 * import { useContext } from 'react';
 *
 * const MyComponent = () => {
 *   const logger = useContext(LoggerContext);
 *   if (!logger) {
 *     throw new Error('Component must be used within LoggerProvider');
 *   }
 *
 *   logger.info('Component rendered');
 *   return <div>Hello World</div>;
 * };
 * ```
 *
 * @public
 */
const LoggerContext: Context<LoggerContextType> = createContext<LoggerContextType>(null);

export default LoggerContext;
