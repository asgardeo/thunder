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

import type LogLevel from './log-level';
import type LogContext from './log-context';
import type Transport from './transport';

/**
 * Configuration options for a logger instance.
 */
interface LoggerConfig {
  /**
   * Minimum log level to process. Logs below this level will be ignored.
   * @default LogLevel.INFO
   */
  level?: LogLevel;

  /**
   * Whether logging is enabled.
   * @default true
   */
  enabled?: boolean;

  /**
   * Array of transports to use for log output.
   * If not provided, a default transport will be automatically selected
   * based on the runtime environment (Console for browser, Stdout for Node.js).
   */
  transports?: Transport[];

  /**
   * Default context to include in all log entries from this logger.
   */
  context?: LogContext;

  /**
   * Component name to identify the source of logs.
   */
  component?: string;

  /**
   * Whether to mask sensitive data in logs.
   * @default false
   */
  maskSensitiveData?: boolean;
}

export default LoggerConfig;
