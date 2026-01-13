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
import {hasConsole} from '../utils/detectEnvironment';
import formatTimestamp from '../utils/formatTimestamp';
import serializeError from '../utils/serializeError';

/* eslint-disable no-console */

/**
 * Options for configuring the console transport.
 */
export interface ConsoleTransportOptions {
  /**
   * Whether to use colored output (browser only).
   * @default true
   */
  colors?: boolean;

  /**
   * Whether to include timestamps in the output.
   * @default true
   */
  timestamps?: boolean;

  /**
   * Whether to use pretty formatting for objects.
   * @default true
   */
  prettyPrint?: boolean;
}

/**
 * Console color codes for different log levels (browser).
 */
const LEVEL_COLORS: Record<string, string> = {
  debug: 'color: #6c757d', // Gray
  info: 'color: #0dcaf0', // Cyan
  warn: 'color: #ffc107', // Yellow
  error: 'color: #dc3545', // Red
};

/**
 * Console transport for browser environments.
 * Writes logs to the browser console with styled output.
 */
export default class ConsoleTransport extends BaseTransport {
  private options: Required<ConsoleTransportOptions>;

  constructor(level?: LogLevel, options: ConsoleTransportOptions = {}) {
    super('console', level);
    this.options = {
      colors: options.colors ?? true,
      timestamps: options.timestamps ?? true,
      prettyPrint: options.prettyPrint ?? true,
    };
  }

  write(entry: LogEntry): Promise<void> {
    if (!hasConsole()) {
      return Promise.resolve();
    }

    const {level, message, timestamp, context, component, error} = entry;

    // Build the log message
    const parts: string[] = [];
    const styles: string[] = [];

    // Add timestamp
    if (this.options.timestamps) {
      parts.push(`[${formatTimestamp(timestamp)}]`);
      styles.push('color: #999');
    }

    // Add level with color
    parts.push(`[${level.toUpperCase()}]`);
    styles.push(LEVEL_COLORS[level] || '');

    // Add component if present
    if (component) {
      parts.push(`[${component}]`);
      styles.push('color: #6610f2; font-weight: bold');
    }

    // Add message
    parts.push(message);
    styles.push('');

    // Format the log string
    const logString = parts.join(' ');
    const styleString = this.options.colors ? parts.map(() => '%c').join('') : '';

    // Select console method based on level
    const consoleMethod = this.getConsoleMethod(level);

    // Log with styles if colors are enabled
    if (this.options.colors && typeof consoleMethod === 'function') {
      consoleMethod(styleString + logString, ...styles);
    } else if (typeof consoleMethod === 'function') {
      consoleMethod(logString);
    }

    // Log context if present (filter out null/undefined values)
    if (context && Object.keys(context).length > 0) {
      const filteredContext = Object.entries(context).reduce(
        (acc, [key, value]) => {
          if (value !== null && value !== undefined) {
            acc[key] = value;
          }
          return acc;
        },
        {} as Record<string, unknown>,
      );

      if (Object.keys(filteredContext).length > 0) {
        if (this.options.prettyPrint && typeof console.log === 'function') {
          console.log('  Context:', filteredContext);
        } else if (typeof consoleMethod === 'function') {
          consoleMethod('Context:', filteredContext);
        }
      }
    }

    // Log error if present
    if (error && typeof console.error === 'function') {
      if (this.options.prettyPrint) {
        console.error('Error:', error);
      } else {
        console.error('Error:', serializeError(error));
      }
    }

    return Promise.resolve();
  }

  // eslint-disable-next-line @typescript-eslint/class-methods-use-this
  private getConsoleMethod(level: string): (...args: unknown[]) => void {
    switch (level) {
      case 'debug':
        return console.debug || console.log;
      case 'info':
        return console.info || console.log;
      case 'warn':
        return console.warn || console.log;
      case 'error':
        return console.error || console.log;
      default:
        return console.log;
    }
  }
}
