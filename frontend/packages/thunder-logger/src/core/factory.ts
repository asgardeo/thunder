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

import type Transport from '../models/transport';
import type {TransportConfig, TransportFactory} from '../models/transport-config';
import LogLevel from '../models/log-level';
import ConsoleTransport from '../transports/ConsoleTransport';
import StdoutTransport from '../transports/StdoutTransport';
import HttpTransport, {type HttpTransportOptions} from '../transports/HttpTransport';

/**
 * Registry of transport factories.
 * Allows registration of custom transport types.
 */
class TransportFactoryRegistry {
  private factories: Map<string, TransportFactory>;

  constructor() {
    this.factories = new Map();

    // Register built-in transports
    this.register('console', (config) => {
      const level = config.level ?? LogLevel.DEBUG;
      const options = config.options ?? {};
      return new ConsoleTransport(level, options);
    });

    this.register('stdout', (config) => {
      const level = config.level ?? LogLevel.DEBUG;
      const options = config.options ?? {};
      return new StdoutTransport(level, options);
    });

    this.register('http', (config) => {
      const level = config.level ?? LogLevel.DEBUG;
      const options = config.options ?? {};

      if (!('endpoint' in options) || typeof options['endpoint'] !== 'string') {
        throw new Error('HTTP transport requires an endpoint option');
      }

      return new HttpTransport(level, options as unknown as HttpTransportOptions);
    });
  }

  /**
   * Register a custom transport factory.
   * @param type - The transport type identifier
   * @param factory - Factory function that creates the transport
   *
   * @example
   * ```typescript
   * // Register a custom Sentry transport
   * registerTransport('sentry', (config) => {
   *   return new SentryTransport(config.level, config.options);
   * });
   * ```
   */
  register(type: string, factory: TransportFactory): void {
    this.factories.set(type, factory);
  }

  /**
   * Create a transport instance from configuration.
   * @param config - Transport configuration
   * @returns Transport instance
   * @throws Error if transport type is not registered
   */
  create(config: TransportConfig): Transport {
    const factory = this.factories.get(config.type);

    if (!factory) {
      throw new Error(
        `Unknown transport type: ${config.type}. Available types: ${Array.from(this.factories.keys()).join(', ')}`,
      );
    }

    return factory(config);
  }

  /**
   * Check if a transport type is registered.
   * @param type - The transport type to check
   * @returns True if the type is registered
   */
  has(type: string): boolean {
    return this.factories.has(type);
  }

  /**
   * Get all registered transport types.
   * @returns Array of registered type names
   */
  getTypes(): string[] {
    return Array.from(this.factories.keys());
  }

  /**
   * Unregister a transport type.
   * @param type - The transport type to unregister
   * @returns True if the type was unregistered
   */
  unregister(type: string): boolean {
    return this.factories.delete(type);
  }
}

/**
 * Global transport factory registry instance.
 */
const globalRegistry = new TransportFactoryRegistry();

/**
 * Create a transport from configuration using the global registry.
 * @param config - Transport configuration
 * @returns Transport instance
 *
 * @example
 * ```typescript
 * // Create a console transport
 * const transport = createTransport({
 *   type: 'console',
 *   level: LogLevel.DEBUG,
 *   options: { colors: true },
 * });
 *
 * // Create an HTTP transport
 * const httpTransport = createTransport({
 *   type: 'http',
 *   level: LogLevel.ERROR,
 *   options: {
 *     endpoint: 'https://logs.example.com/api/logs',
 *     batchSize: 10,
 *   },
 * });
 * ```
 */
export function createTransport(config: TransportConfig): Transport {
  return globalRegistry.create(config);
}

/**
 * Create multiple transports from an array of configurations.
 * @param configs - Array of transport configurations
 * @returns Array of transport instances
 *
 * @example
 * ```typescript
 * const transports = createTransports([
 *   { type: 'console', level: LogLevel.DEBUG },
 *   { type: 'http', level: LogLevel.ERROR, options: { endpoint: '...' } },
 * ]);
 * ```
 */
export function createTransports(configs: TransportConfig[]): Transport[] {
  return configs.map((config) => createTransport(config));
}

/**
 * Register a custom transport factory globally.
 * @param type - The transport type identifier
 * @param factory - Factory function that creates the transport
 *
 * @example
 * ```typescript
 * import { registerTransport } from '@thunder/logger';
 * import * as Sentry from '@sentry/browser';
 *
 * // Define a custom Sentry transport
 * class SentryTransport extends BaseTransport {
 *   constructor(level: LogLevel, options: { dsn: string }) {
 *     super('sentry', level);
 *     Sentry.init({ dsn: options.dsn });
 *   }
 *
 *   async write(entry: LogEntry): Promise<void> {
 *     if (entry.level === 'error') {
 *       Sentry.captureException(entry.error || new Error(entry.message), {
 *         level: 'error',
 *         extra: entry.context,
 *       });
 *     }
 *   }
 * }
 *
 * // Register the transport
 * registerTransport('sentry', (config) => {
 *   return new SentryTransport(config.level, config.options);
 * });
 *
 * // Use it in a logger
 * const logger = new Logger({
 *   transports: [
 *     createTransport({
 *       type: 'sentry',
 *       level: LogLevel.ERROR,
 *       options: { dsn: 'your-sentry-dsn' },
 *     }),
 *   ],
 * });
 * ```
 */
export function registerTransport(type: string, factory: TransportFactory): void {
  globalRegistry.register(type, factory);
}

/**
 * Check if a transport type is registered.
 * @param type - The transport type to check
 * @returns True if the type is registered
 */
export function hasTransport(type: string): boolean {
  return globalRegistry.has(type);
}

/**
 * Get all registered transport types.
 * @returns Array of registered type names
 */
export function getTransportTypes(): string[] {
  return globalRegistry.getTypes();
}

/**
 * Unregister a transport type.
 * @param type - The transport type to unregister
 * @returns True if the type was unregistered
 */
export function unregisterTransport(type: string): boolean {
  return globalRegistry.unregister(type);
}
