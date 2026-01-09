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
import type LogEntry from './log-entry';

/**
 * Interface that all transport implementations must satisfy.
 * Transports are responsible for outputting log entries to their destination.
 */
interface Transport {
  /**
   * Get the name/type of this transport.
   */
  getName(): string;

  /**
   * Write a log entry to the transport's destination.
   * @param entry - The log entry to write
   * @returns Promise that resolves when the write is complete
   */
  write(entry: LogEntry): Promise<void>;

  /**
   * Flush any buffered log entries.
   * @returns Promise that resolves when all buffered entries are written
   */
  flush(): Promise<void>;

  /**
   * Close the transport and release any resources.
   * @returns Promise that resolves when the transport is closed
   */
  close(): Promise<void>;

  /**
   * Get the minimum log level this transport handles.
   */
  getLevel(): LogLevel;

  /**
   * Check if this transport should handle a given log level.
   * @param level - The log level to check
   */
  shouldLog(level: LogLevel): boolean;
}

export default Transport;
