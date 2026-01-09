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

import LogLevel, {LOG_LEVEL_PRIORITY} from '../models/log-level';
import type Transport from '../models/transport';
import type LogEntry from '../models/log-entry';

/**
 * Abstract base class for transport implementations.
 * Provides common functionality for log level filtering.
 */
export default abstract class BaseTransport implements Transport {
  protected level: LogLevel;

  protected name: string;

  constructor(name: string, level?: LogLevel) {
    this.name = name;
    this.level = level ?? LogLevel.DEBUG;
  }

  getName(): string {
    return this.name;
  }

  getLevel(): LogLevel {
    return this.level;
  }

  shouldLog(level: LogLevel): boolean {
    return LOG_LEVEL_PRIORITY[level] >= LOG_LEVEL_PRIORITY[this.level];
  }

  abstract write(entry: LogEntry): Promise<void>;

  // eslint-disable-next-line @typescript-eslint/class-methods-use-this
  async flush(): Promise<void> {
    // Default implementation does nothing
    // Subclasses can override for buffering support
  }

  // eslint-disable-next-line @typescript-eslint/class-methods-use-this
  async close(): Promise<void> {
    // Default implementation does nothing
    // Subclasses can override for cleanup
  }
}
