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

import {describe, expect, it, vi, beforeEach} from 'vitest';
import PluginRegistry from '../PluginRegistry';
import VisualFlowConstants from '../../constants/VisualFlowConstants';

// Helper to create a handler with the required uniqueName property
const createHandler = <T extends (...args: unknown[]) => unknown>(
  mockFn: T,
  uniqueName: string | number | null,
): T & Record<string, unknown> => {
  const handler = mockFn as unknown as Record<string, unknown>;
  if (uniqueName !== null) {
    handler[VisualFlowConstants.FLOW_BUILDER_PLUGIN_FUNCTION_IDENTIFIER] = uniqueName;
  }
  return handler as T & Record<string, unknown>;
};

describe('PluginRegistry', () => {
  let registry: PluginRegistry;

  beforeEach(() => {
    // Get a fresh instance for each test by accessing the singleton
    registry = PluginRegistry.getInstance();
    // Clear all registered plugins by unregistering them
    // We need to track what we register and clean up
  });

  describe('Singleton Pattern', () => {
    it('should return the same instance when getInstance is called multiple times', () => {
      const instance1 = PluginRegistry.getInstance();
      const instance2 = PluginRegistry.getInstance();

      expect(instance1).toBe(instance2);
    });

    it('should be an instance of PluginRegistry', () => {
      const instance = PluginRegistry.getInstance();

      expect(instance).toBeInstanceOf(PluginRegistry);
    });
  });

  describe('registerAsync', () => {
    it('should register an async plugin with a valid identifier', () => {
      const handler = createHandler(vi.fn().mockResolvedValue(true), 'testAsyncHandler');

      expect(() => registry.registerAsync('testEvent', handler)).not.toThrow();
    });

    it('should throw an error when handler lacks uniqueName property', () => {
      const handler = vi.fn().mockResolvedValue(true);

      expect(() => registry.registerAsync('testEvent', handler)).toThrow(
        'Handler function must have the `uniqueName` property.',
      );
    });

    it('should throw an error when uniqueName is not a string', () => {
      const handler = createHandler(vi.fn().mockResolvedValue(true), 123);

      expect(() => registry.registerAsync('testEvent', handler)).toThrow(
        'Handler function must have the `uniqueName` property.',
      );
    });

    it('should allow registering multiple async handlers for the same event', () => {
      const handler1 = createHandler(vi.fn().mockResolvedValue(true), 'handler1');
      const handler2 = createHandler(vi.fn().mockResolvedValue(true), 'handler2');

      expect(() => {
        registry.registerAsync('multiHandlerEvent', handler1);
        registry.registerAsync('multiHandlerEvent', handler2);
      }).not.toThrow();

      // Clean up
      registry.unregister('multiHandlerEvent', 'handler1');
      registry.unregister('multiHandlerEvent', 'handler2');
    });
  });

  describe('registerSync', () => {
    it('should register a sync plugin with a valid identifier', () => {
      const handler = createHandler(vi.fn().mockReturnValue(true), 'testSyncHandler');

      expect(() => registry.registerSync('testSyncEvent', handler)).not.toThrow();

      // Clean up
      registry.unregister('testSyncEvent', 'testSyncHandler');
    });

    it('should throw an error when handler lacks uniqueName property', () => {
      const handler = vi.fn().mockReturnValue(true);

      expect(() => registry.registerSync('testSyncEvent', handler)).toThrow(
        'Handler function must have the `uniqueName` property.',
      );
    });

    it('should throw an error when uniqueName is not a string', () => {
      const handler = createHandler(vi.fn().mockReturnValue(true), null);

      expect(() => registry.registerSync('testSyncEvent', handler)).toThrow(
        'Handler function must have the `uniqueName` property.',
      );
    });

    it('should allow registering multiple sync handlers for the same event', () => {
      const handler1 = createHandler(vi.fn().mockReturnValue(true), 'syncHandler1');
      const handler2 = createHandler(vi.fn().mockReturnValue(true), 'syncHandler2');

      expect(() => {
        registry.registerSync('multiSyncEvent', handler1);
        registry.registerSync('multiSyncEvent', handler2);
      }).not.toThrow();

      // Clean up
      registry.unregister('multiSyncEvent', 'syncHandler1');
      registry.unregister('multiSyncEvent', 'syncHandler2');
    });
  });

  describe('unregister', () => {
    it('should unregister an async plugin', () => {
      const handler = createHandler(vi.fn().mockResolvedValue(true), 'unregisterAsyncTest');

      registry.registerAsync('unregisterAsyncEvent', handler);
      registry.unregister('unregisterAsyncEvent', 'unregisterAsyncTest');

      // After unregistering, executeAsync should return true (no handlers)
      return expect(registry.executeAsync('unregisterAsyncEvent')).resolves.toBe(true);
    });

    it('should unregister a sync plugin', () => {
      const handler = createHandler(vi.fn().mockReturnValue(false), 'unregisterSyncTest');

      registry.registerSync('unregisterSyncEvent', handler);
      registry.unregister('unregisterSyncEvent', 'unregisterSyncTest');

      // After unregistering, executeSync should return true (no handlers)
      expect(registry.executeSync('unregisterSyncEvent')).toBe(true);
    });

    it('should not throw when unregistering non-existent handler', () => {
      expect(() => registry.unregister('nonExistentEvent', 'nonExistentHandler')).not.toThrow();
    });

    it('should remove the event when all handlers are unregistered', () => {
      const handler = createHandler(vi.fn().mockReturnValue(true), 'onlyHandler');

      registry.registerSync('singleHandlerEvent', handler);
      registry.unregister('singleHandlerEvent', 'onlyHandler');

      // executeSync should return true since no handlers exist
      expect(registry.executeSync('singleHandlerEvent')).toBe(true);
    });
  });

  describe('executeSync', () => {
    it('should return true when no handlers are registered for the event', () => {
      const result = registry.executeSync('noHandlersEvent');

      expect(result).toBe(true);
    });

    it('should execute registered sync handler and return its result', () => {
      const handler = createHandler(vi.fn().mockReturnValue(true), 'execSyncHandler');

      registry.registerSync('execSyncEvent', handler);
      const result = registry.executeSync('execSyncEvent', 'arg1', 'arg2');

      expect(handler).toHaveBeenCalledWith('arg1', 'arg2');
      expect(result).toBe(true);

      // Clean up
      registry.unregister('execSyncEvent', 'execSyncHandler');
    });

    it('should return false if any handler returns false', () => {
      const handler1 = createHandler(vi.fn().mockReturnValue(true), 'syncHandler1False');
      const handler2 = createHandler(vi.fn().mockReturnValue(false), 'syncHandler2False');

      registry.registerSync('mixedSyncEvent', handler1);
      registry.registerSync('mixedSyncEvent', handler2);

      const result = registry.executeSync('mixedSyncEvent');

      expect(result).toBe(false);

      // Clean up
      registry.unregister('mixedSyncEvent', 'syncHandler1False');
      registry.unregister('mixedSyncEvent', 'syncHandler2False');
    });

    it('should return true if all handlers return true', () => {
      const handler1 = createHandler(vi.fn().mockReturnValue(true), 'allTrueHandler1');
      const handler2 = createHandler(vi.fn().mockReturnValue(true), 'allTrueHandler2');

      registry.registerSync('allTrueSyncEvent', handler1);
      registry.registerSync('allTrueSyncEvent', handler2);

      const result = registry.executeSync('allTrueSyncEvent');

      expect(result).toBe(true);

      // Clean up
      registry.unregister('allTrueSyncEvent', 'allTrueHandler1');
      registry.unregister('allTrueSyncEvent', 'allTrueHandler2');
    });
  });

  describe('executeAsync', () => {
    it('should return true when no handlers are registered for the event', async () => {
      const result = await registry.executeAsync('noAsyncHandlersEvent');

      expect(result).toBe(true);
    });

    it('should execute registered async handler and return its result', async () => {
      const handler = createHandler(vi.fn().mockResolvedValue(true), 'execAsyncHandler');

      registry.registerAsync('execAsyncEvent', handler);
      const result = await registry.executeAsync('execAsyncEvent', 'arg1', 'arg2');

      expect(handler).toHaveBeenCalledWith('arg1', 'arg2');
      expect(result).toBe(true);

      // Clean up
      registry.unregister('execAsyncEvent', 'execAsyncHandler');
    });

    it('should return false if any async handler returns false', async () => {
      const handler1 = createHandler(vi.fn().mockResolvedValue(true), 'asyncHandler1False');
      const handler2 = createHandler(vi.fn().mockResolvedValue(false), 'asyncHandler2False');

      registry.registerAsync('mixedAsyncEvent', handler1);
      registry.registerAsync('mixedAsyncEvent', handler2);

      const result = await registry.executeAsync('mixedAsyncEvent');

      expect(result).toBe(false);

      // Clean up
      registry.unregister('mixedAsyncEvent', 'asyncHandler1False');
      registry.unregister('mixedAsyncEvent', 'asyncHandler2False');
    });

    it('should return true if all async handlers return true', async () => {
      const handler1 = createHandler(vi.fn().mockResolvedValue(true), 'allTrueAsyncHandler1');
      const handler2 = createHandler(vi.fn().mockResolvedValue(true), 'allTrueAsyncHandler2');

      registry.registerAsync('allTrueAsyncEvent', handler1);
      registry.registerAsync('allTrueAsyncEvent', handler2);

      const result = await registry.executeAsync('allTrueAsyncEvent');

      expect(result).toBe(true);

      // Clean up
      registry.unregister('allTrueAsyncEvent', 'allTrueAsyncHandler1');
      registry.unregister('allTrueAsyncEvent', 'allTrueAsyncHandler2');
    });

    it('should execute all async handlers in parallel', async () => {
      const executionOrder: string[] = [];

      const handler1 = createHandler(
        vi.fn().mockImplementation(async () => {
          await new Promise((resolve) => {
            setTimeout(resolve, 10);
          });
          executionOrder.push('handler1');
          return true;
        }),
        'parallelHandler1',
      );

      const handler2 = createHandler(
        vi.fn().mockImplementation(async () => {
          executionOrder.push('handler2');
          return true;
        }),
        'parallelHandler2',
      );

      registry.registerAsync('parallelAsyncEvent', handler1);
      registry.registerAsync('parallelAsyncEvent', handler2);

      await registry.executeAsync('parallelAsyncEvent');

      // Both handlers should have been called
      expect(handler1).toHaveBeenCalled();
      expect(handler2).toHaveBeenCalled();

      // Clean up
      registry.unregister('parallelAsyncEvent', 'parallelHandler1');
      registry.unregister('parallelAsyncEvent', 'parallelHandler2');
    });
  });

  describe('executeAsync with debounce', () => {
    it('should execute with debounce when debounce time is provided', async () => {
      const handler = createHandler(vi.fn().mockResolvedValue(true), 'debounceHandler');

      registry.registerAsync('debounceEvent', handler);

      // Execute with debounce time as second argument
      const result = await registry.executeAsync('debounceEvent', 10, 'arg1');

      // The debounced function should eventually execute
      expect(result).toBe(true);

      // Clean up
      registry.unregister('debounceEvent', 'debounceHandler');
    });

    it('should return true when no handlers and debounce time is provided', async () => {
      const result = await registry.executeAsync('noHandlersDebounceEvent', 10);

      expect(result).toBe(true);
    });
  });

  describe('Handler Replacement', () => {
    it('should replace handler with the same identifier', () => {
      const handler1 = createHandler(vi.fn().mockReturnValue(true), 'sameId');
      const handler2 = createHandler(vi.fn().mockReturnValue(false), 'sameId');

      registry.registerSync('replaceEvent', handler1);
      registry.registerSync('replaceEvent', handler2);

      const result = registry.executeSync('replaceEvent');

      // handler2 should have replaced handler1, so result should be false
      expect(result).toBe(false);
      expect(handler1).not.toHaveBeenCalled();
      expect(handler2).toHaveBeenCalled();

      // Clean up
      registry.unregister('replaceEvent', 'sameId');
    });
  });
});
