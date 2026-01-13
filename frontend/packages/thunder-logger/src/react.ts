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

/* eslint-disable import/extensions */

// React-specific exports
export {default as LoggerContext} from './react/contexts/Logger/LoggerContext';
export type {LoggerContextType} from './react/contexts/Logger/LoggerContext';

export {default as LoggerProvider} from './react/contexts/Logger/LoggerProvider';
export type {LoggerProviderProps} from './react/contexts/Logger/LoggerProvider';

export {default as useLogger} from './react/contexts/Logger/useLogger';

// Re-export core logger class and types for convenience
export {default as Logger} from './core/Logger';
export {default as LogLevel} from './models/log-level';
export type {default as LoggerConfig} from './models/logger-config';
export type {default as LogContext} from './models/log-context';

// Re-export transports for convenience
export {default as ConsoleTransport} from './transports/ConsoleTransport';
export {default as StdoutTransport} from './transports/StdoutTransport';
export {default as HttpTransport} from './transports/HttpTransport';
