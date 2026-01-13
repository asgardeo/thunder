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
import type LogContext from '../models/log-context';
import type LoggerConfig from '../models/logger-config';
import type LogEntry from '../models/log-entry';
import type Transport from '../models/transport';
import detectEnvironment from '../utils/detectEnvironment';
import maskSensitiveData from '../utils/maskSensitiveData';
import ConsoleTransport from '../transports/ConsoleTransport';
import StdoutTransport from '../transports/StdoutTransport';

/**
 * Logger class for structured, multi-transport logging.
 *
 * @example
 * ```typescript
 * // Create a logger with default transport
 * const logger = new Logger({ level: LogLevel.INFO });
 *
 * // Log messages
 * logger.info('User logged in', { userId: '123' });
 * logger.error('Failed to fetch data', { error: err });
 *
 * // Create a logger with custom transports
 * const logger = new Logger({
 *   level: LogLevel.DEBUG,
 *   transports: [
 *     new ConsoleTransport(LogLevel.DEBUG),
 *     new HttpTransport(LogLevel.ERROR, { endpoint: 'https://logs.example.com' }),
 *   ],
 * });
 *
 * // Create contextual loggers
 * const userLogger = logger.withContext({ userId: '123' });
 * const componentLogger = logger.withComponent('AuthService');
 * ```
 */
export default class Logger {
  private config: Omit<LoggerConfig, 'transports'> & {
    level: LogLevel;
    enabled: boolean;
    transports: Transport[];
    context: LogContext;
    maskSensitiveData: boolean;
  };

  private context: LogContext;

  constructor(config: LoggerConfig = {}) {
    // Detect environment for defaults
    const env = detectEnvironment();

    // Set up transports
    let transports = config.transports ?? [];

    // If no transports provided, use default based on environment
    if (transports.length === 0) {
      const defaultLevel = (config.level ?? 'info') as LogLevel;
      if (env.isBrowser) {
        transports = [new ConsoleTransport(defaultLevel)];
      } else if (env.isNode) {
        transports = [new StdoutTransport(defaultLevel)];
      } else {
        // Fallback to console
        transports = [new ConsoleTransport(defaultLevel)];
      }
    }

    this.config = {
      level: (config.level ?? 'info') as LogLevel,
      enabled: config.enabled ?? true,
      transports,
      context: config.context ?? {},
      component: config.component,
      maskSensitiveData: config.maskSensitiveData ?? false,
    };

    this.context = {...this.config.context};
  }

  /**
   * Log a debug message.
   * @param message - The log message
   * @param context - Additional contextual data
   */
  debug(message: string, context?: LogContext): void {
    this.log(LogLevel.DEBUG, message, context);
  }

  /**
   * Log an info message.
   * @param message - The log message
   * @param context - Additional contextual data
   */
  info(message: string, context?: LogContext): void {
    this.log(LogLevel.INFO, message, context);
  }

  /**
   * Log a warning message.
   * @param message - The log message
   * @param context - Additional contextual data
   */
  warn(message: string, context?: LogContext): void {
    this.log(LogLevel.WARN, message, context);
  }

  /**
   * Log an error message.
   * @param message - The log message
   * @param errorOrContext - Error object or contextual data
   * @param context - Additional contextual data (if first param is Error)
   */
  error(message: string, errorOrContext?: Error | LogContext, context?: LogContext): void {
    let error: Error | undefined;
    let ctx: LogContext | undefined;

    if (errorOrContext instanceof Error) {
      error = errorOrContext;
      ctx = context;
    } else {
      ctx = errorOrContext;
    }

    this.log(LogLevel.ERROR, message, ctx, error);
  }

  /**
   * Create a new logger instance with additional context.
   * @param context - Context to add to all log entries
   * @returns New logger instance with merged context
   */
  withContext(context: LogContext): Logger {
    const newLogger = new Logger({
      ...this.config,
      context: {...this.context, ...context},
    });
    return newLogger;
  }

  /**
   * Create a new logger instance with a component name.
   * @param component - Component name to identify log source
   * @returns New logger instance with component set
   */
  withComponent(component: string): Logger {
    const newLogger = new Logger({
      ...this.config,
      component,
    });
    newLogger.context = {...this.context};
    return newLogger;
  }

  /**
   * Check if a given log level would be logged.
   * Useful for avoiding expensive operations when logging is disabled.
   * @param level - The log level to check
   * @returns True if the level would be logged
   */
  isLevelEnabled(level: LogLevel): boolean {
    if (!this.config.enabled) {
      return false;
    }
    return LOG_LEVEL_PRIORITY[level] >= LOG_LEVEL_PRIORITY[this.config.level];
  }

  /**
   * Check if debug logging is enabled.
   * @returns True if debug logs would be output
   */
  isDebugEnabled(): boolean {
    return this.isLevelEnabled(LogLevel.DEBUG);
  }

  /**
   * Flush all transports, ensuring buffered logs are written.
   * @returns Promise that resolves when all transports are flushed
   */
  async flush(): Promise<void> {
    await Promise.all(this.config.transports.map((transport) => transport.flush()));
  }

  /**
   * Close all transports and release resources.
   * @returns Promise that resolves when all transports are closed
   */
  async close(): Promise<void> {
    await Promise.all(this.config.transports.map((transport) => transport.close()));
  }

  /**
   * Internal log method that routes to all transports.
   * @param level - Log level
   * @param message - Log message
   * @param context - Additional context
   * @param error - Optional error object
   */
  private log(level: LogLevel, message: string, context?: LogContext, error?: Error): void {
    // Check if logging is enabled
    if (!this.config.enabled) {
      return;
    }

    // Check if log level should be processed
    if (!this.isLevelEnabled(level)) {
      return;
    }

    // Merge contexts
    const mergedContext = {...this.context, ...context};

    // Mask sensitive data if enabled
    const finalContext = this.config.maskSensitiveData
      ? (maskSensitiveData(mergedContext) as LogContext)
      : mergedContext;

    // Create log entry
    const entry: LogEntry = {
      level,
      message,
      timestamp: new Date(),
      context: Object.keys(finalContext).length > 0 ? finalContext : undefined,
      component: this.config.component,
      error,
    };

    // Write to all transports
    this.config.transports.forEach((transport) => {
      // Check if transport should handle this level
      if (transport.shouldLog(level)) {
        // Write asynchronously but don't wait
        // Errors in transports should not affect the application
        transport.write(entry).catch((err) => {
          // Fallback error logging
          // eslint-disable-next-line no-console
          if (typeof console !== 'undefined' && typeof console.error === 'function') {
            // eslint-disable-next-line no-console
            console.error(`Transport ${transport.getName()} failed:`, err);
          }
        });
      }
    });
  }
}

/**
 * Create a new logger instance.
 * @param config - Logger configuration
 * @returns New logger instance
 */
export function createLogger(config?: LoggerConfig): Logger {
  return new Logger(config);
}
