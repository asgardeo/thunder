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

import BaseTransport from '../core/BaseTransport';
import type LogEntry from '../models/log-entry';
import type LogLevel from '../models/log-level';
import {hasProcess} from '../utils/detectEnvironment';
import formatTimestamp from '../utils/formatTimestamp';
import serializeError from '../utils/serializeError';

/**
 * Options for configuring the stdout transport.
 */
export interface StdoutTransportOptions {
  /**
   * Whether to output as JSON lines.
   * @default true
   */
  json?: boolean;

  /**
   * Whether to include timestamps.
   * @default true
   */
  timestamps?: boolean;
}

/**
 * Stdout transport for Node.js environments.
 * Writes structured JSON logs to stdout.
 */
export default class StdoutTransport extends BaseTransport {
  private options: Required<StdoutTransportOptions>;

  constructor(level?: LogLevel, options: StdoutTransportOptions = {}) {
    super('stdout', level);
    this.options = {
      json: options.json ?? true,
      timestamps: options.timestamps ?? true,
    };
  }

  write(entry: LogEntry): Promise<void> {
    if (!hasProcess()) {
      return Promise.resolve();
    }

    const {level, message, timestamp, context, component, error} = entry;

    if (this.options.json) {
      // Structured JSON output
      const logObject: Record<string, unknown> = {
        level,
        message,
      };

      if (this.options.timestamps) {
        logObject['timestamp'] = formatTimestamp(timestamp);
      }

      if (component) {
        logObject['component'] = component;
      }

      if (context && Object.keys(context).length > 0) {
        logObject['context'] = context;
      }

      if (error) {
        logObject['error'] = serializeError(error);
      }

      // Write to stdout
      process.stdout.write(`${JSON.stringify(logObject)}\n`);
    } else {
      // Plain text output
      const parts: string[] = [];

      if (this.options.timestamps) {
        parts.push(`[${formatTimestamp(timestamp)}]`);
      }

      parts.push(`[${level.toUpperCase()}]`);

      if (component) {
        parts.push(`[${component}]`);
      }

      parts.push(message);

      let output = parts.join(' ');

      if (context && Object.keys(context).length > 0) {
        output += ` ${JSON.stringify(context)}`;
      }

      if (error) {
        output += ` Error: ${JSON.stringify(serializeError(error))}`;
      }

      process.stdout.write(`${output}\n`);
    }

    return Promise.resolve();
  }
}
