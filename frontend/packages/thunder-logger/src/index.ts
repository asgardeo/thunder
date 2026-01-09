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

// Core logger exports
export {default as Logger, createLogger} from './core/Logger';
export {default as BaseTransport} from './core/BaseTransport';
export {default as TransportRegistry} from './core/TransportRegistry';

// Transports
export {default as ConsoleTransport} from './transports/ConsoleTransport';
export {default as StdoutTransport} from './transports/StdoutTransport';
export {default as HttpTransport} from './transports/HttpTransport';

// Type exports from models
export {default as LogLevel, LOG_LEVEL_PRIORITY} from './models/log-level';
export type {default as LogContext} from './models/log-context';
export type {default as LogEntry} from './models/log-entry';
export type {default as LoggerConfig} from './models/logger-config';
export type {default as Transport} from './models/transport';
export type {default as TransportConfig, TransportFactory} from './models/transport-config';
export type {default as RuntimeEnvironment} from './models/runtime-environment';

// Transport options type exports
export type {ConsoleTransportOptions} from './transports/ConsoleTransport';
export type {StdoutTransportOptions} from './transports/StdoutTransport';
export type {HttpTransportOptions} from './transports/HttpTransport';

// React context exports
export {default as LoggerContext} from './react/contexts/Logger/LoggerContext';
export type {LoggerContextType} from './react/contexts/Logger/LoggerContext';

export {default as LoggerProvider} from './react/contexts/Logger/LoggerProvider';
export type {LoggerProviderProps} from './react/contexts/Logger/LoggerProvider';

export {default as useLogger} from './react/contexts/Logger/useLogger';

// Utility exports
export {default as detectEnvironment, hasConsole, hasProcess} from './utils/detectEnvironment';
export {default as formatTimestamp} from './utils/formatTimestamp';
export {default as maskSensitiveData, maskString} from './utils/maskSensitiveData';
export {default as serializeError} from './utils/serializeError';
