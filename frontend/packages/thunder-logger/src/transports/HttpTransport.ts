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
import formatTimestamp from '../utils/formatTimestamp';
import serializeError from '../utils/serializeError';

/**
 * Serialized version of LogEntry where Error objects are converted to plain objects.
 */
type SerializedLogEntry = Omit<LogEntry, 'error'> & {
  error?: Record<string, unknown>;
};

/**
 * Options for configuring the HTTP transport.
 */
export interface HttpTransportOptions {
  /**
   * The endpoint URL to send logs to.
   */
  endpoint: string;

  /**
   * HTTP method to use.
   * @default 'POST'
   */
  method?: 'POST' | 'PUT';

  /**
   * Additional headers to include in requests.
   */
  headers?: Record<string, string>;

  /**
   * Maximum number of log entries to buffer before sending.
   * @default 1
   */
  batchSize?: number;

  /**
   * Maximum time (in milliseconds) to wait before sending buffered logs.
   * @default 5000
   */
  flushInterval?: number;

  /**
   * Request timeout in milliseconds.
   * @default 10000
   */
  timeout?: number;

  /**
   * Whether to retry failed requests.
   * @default true
   */
  retry?: boolean;

  /**
   * Maximum number of retry attempts.
   * @default 3
   */
  maxRetries?: number;

  /**
   * Callback invoked when logs fail to send after all retries are exhausted.
   * Use this to implement fallback logging strategies (e.g., store to localStorage, send to alternative endpoint).
   * @param entries - The log entries that failed to send
   * @param error - The error that caused the failure
   */
  onDroppedLogs?: (entries: LogEntry[], error: Error) => void;
}

/**
 * HTTP transport for sending logs to a remote endpoint.
 * Supports batching and automatic flushing.
 *
 * **Important:** When batching is enabled (batchSize > 1), this transport starts
 * a periodic flush timer. To prevent memory leaks, you MUST call `close()` when
 * the transport is no longer needed. For Logger instances, call `logger.close()`
 * during application cleanup (e.g., in cleanup hooks, beforeunload handlers, or
 * component unmount).
 *
 * @example
 * ```typescript
 * const logger = new Logger({
 *   transports: [new HttpTransport({ level: 'error', endpoint: '...', batchSize: 10 })]
 * });
 *
 * // Cleanup when done
 * window.addEventListener('beforeunload', () => {
 *   logger.close(); // Stops timers and flushes remaining logs
 * });
 * ```
 */
export default class HttpTransport extends BaseTransport {
  private options: Required<Omit<HttpTransportOptions, 'onDroppedLogs'>> & {
    onDroppedLogs?: (entries: LogEntry[], error: Error) => void;
  };

  private buffer: SerializedLogEntry[] = [];

  private originalEntries: LogEntry[] = [];

  private flushTimer: ReturnType<typeof setTimeout> | null = null;

  private isFlushing = false;

  private isClosed = false;

  constructor(level?: LogLevel, options?: HttpTransportOptions) {
    super('http', level);

    if (!options?.endpoint) {
      throw new Error('HttpTransport requires an endpoint URL');
    }

    this.options = {
      endpoint: options.endpoint,
      method: options.method ?? 'POST',
      headers: options.headers ?? {},
      batchSize: options.batchSize ?? 1,
      flushInterval: options.flushInterval ?? 5000,
      timeout: options.timeout ?? 10000,
      retry: options.retry ?? true,
      maxRetries: options.maxRetries ?? 3,
      onDroppedLogs: options.onDroppedLogs,
    };

    // Set default headers
    if (!this.options.headers['Content-Type']) {
      this.options.headers['Content-Type'] = 'application/json';
    }

    // Start flush timer if batching is enabled
    if (this.options.batchSize > 1) {
      this.startFlushTimer();
    }
  }

  async write(entry: LogEntry): Promise<void> {
    // Guard against writes after close
    if (this.isClosed) {
      // eslint-disable-next-line no-console
      if (typeof console !== 'undefined' && typeof console.warn === 'function') {
        // eslint-disable-next-line no-console
        console.warn('Attempted to write to closed HttpTransport');
      }
      return;
    }

    // Serialize the log entry
    const serializedEntry = this.serializeEntry(entry);

    // Add to buffers
    this.buffer.push(serializedEntry);
    this.originalEntries.push(entry);

    // Flush if batch size is reached
    if (this.buffer.length >= this.options.batchSize) {
      await this.flush();
    }
  }

  override async flush(): Promise<void> {
    // Guard against flush after close or if already flushing
    if (this.isClosed || this.isFlushing || this.buffer.length === 0) {
      return;
    }

    this.isFlushing = true;

    // Take all buffered entries
    const entries = [...this.buffer];
    const originals = [...this.originalEntries];
    this.buffer = [];
    this.originalEntries = [];

    try {
      await this.sendBatch(entries);
    } catch (error) {
      const err = error as Error;

      // Check if this is a retry exhaustion error
      const retriesExhausted = err.message?.includes('retries exhausted');

      // Re-add failed entries to buffer for retry (if retry is enabled and not exhausted)
      if (this.options.retry && !retriesExhausted) {
        this.buffer.unshift(...entries);
        this.originalEntries.unshift(...originals);
      } else if (this.options.onDroppedLogs) {
        // Notify about dropped logs when retry is disabled or exhausted
        this.options.onDroppedLogs(originals, err);
      }

      // Log error to console as fallback
      // eslint-disable-next-line no-console
      if (typeof console !== 'undefined' && typeof console.error === 'function') {
        // eslint-disable-next-line no-console
        console.error('Failed to send logs to HTTP endpoint:', error);
      }
    } finally {
      this.isFlushing = false;
    }
  }

  override async close(): Promise<void> {
    // Mark as closed to prevent new writes
    this.isClosed = true;

    // Stop flush timer
    if (this.flushTimer) {
      clearInterval(this.flushTimer);
      this.flushTimer = null;
    }

    // Flush remaining entries
    await this.flush();
  }

  // eslint-disable-next-line @typescript-eslint/class-methods-use-this
  private serializeEntry(entry: LogEntry): SerializedLogEntry {
    return {
      level: entry.level,
      message: entry.message,
      timestamp: entry.timestamp,
      component: entry.component,
      context: entry.context,
      error: entry.error ? serializeError(entry.error) : undefined,
    };
  }

  private async sendBatch(entries: SerializedLogEntry[], retryCount = 0): Promise<void> {
    const payload = {
      logs: entries.map((entry) => ({
        level: entry.level,
        message: entry.message,
        timestamp: formatTimestamp(entry.timestamp),
        component: entry.component,
        context: entry.context,
        error: entry.error,
      })),
    };

    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), this.options.timeout);

      const response = await fetch(this.options.endpoint, {
        method: this.options.method,
        headers: this.options.headers,
        body: JSON.stringify(payload),
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      return undefined;
    } catch (error) {
      // Retry logic
      if (this.options.retry && retryCount < this.options.maxRetries) {
        const delay = 2 ** retryCount * 1000; // Exponential backoff
        await new Promise<void>((resolve) => {
          setTimeout(() => {
            resolve();
          }, delay);
        });
        return this.sendBatch(entries, retryCount + 1);
      }

      // Mark error as retries exhausted if applicable
      const err = error as Error;
      if (this.options.retry && retryCount >= this.options.maxRetries) {
        throw new Error(`Failed after ${this.options.maxRetries} retries exhausted: ${err.message}`);
      }

      throw error;
    }
  }

  private startFlushTimer(): void {
    this.flushTimer = setInterval(() => {
      // Guard against timer running after close
      if (this.isClosed) {
        if (this.flushTimer) {
          clearInterval(this.flushTimer);
          this.flushTimer = null;
        }
        return;
      }

      this.flush().catch(() => {
        // Ignore flush errors
      });
    }, this.options.flushInterval);
  }
}
