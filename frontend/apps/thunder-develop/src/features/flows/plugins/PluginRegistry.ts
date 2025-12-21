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

import debounce from 'lodash-es/debounce';
import VisualFlowConstants from '../constants/VisualFlowConstants';

type AsyncPluginHandler = (...args: unknown[]) => Promise<boolean>;
type SyncPluginHandler = (...args: unknown[]) => boolean;

/**
 * PluginRegistry is a singleton class that manages the registration and execution of plugins.
 */
class PluginRegistry {
  private static instance: PluginRegistry;

  private asyncPlugins = new Map<string, Map<string, AsyncPluginHandler>>();

  private syncPlugins = new Map<string, Map<string, SyncPluginHandler>>();

  private constructor() {
    // Private constructor to prevent instantiation.
  }

  public static getInstance(): PluginRegistry {
    if (!PluginRegistry.instance) {
      PluginRegistry.instance = new PluginRegistry();
    }

    return PluginRegistry.instance;
  }

  /**
   * Register an async plugin for the given event.
   *
   * @param eventName - The name of the event to register the plugin for.
   * @param handler - The async handler function to be executed when the event is triggered.
   */
  public registerAsync(eventName: string, handler: AsyncPluginHandler): void {
    const handlerWithId = handler as AsyncPluginHandler & Record<string, unknown>;
    const identifier = handlerWithId[VisualFlowConstants.FLOW_BUILDER_PLUGIN_FUNCTION_IDENTIFIER];

    if (!identifier || typeof identifier !== 'string') {
      throw new Error('Handler function must have the `uniqueName` property.');
    }

    if (!this.asyncPlugins.has(eventName)) {
      this.asyncPlugins.set(eventName, new Map());
    }
    this.asyncPlugins.get(eventName)?.set(identifier, handler);
  }

  /**
   * Register an sync plugin for the given event.
   *
   * @param eventName - The name of the event to register the plugin for.
   * @param handler - The sync handler function to be executed when the event is triggered.
   */
  public registerSync(eventName: string, handler: SyncPluginHandler): void {
    const handlerWithId = handler as SyncPluginHandler & Record<string, unknown>;
    const identifier = handlerWithId[VisualFlowConstants.FLOW_BUILDER_PLUGIN_FUNCTION_IDENTIFIER];

    if (!identifier || typeof identifier !== 'string') {
      throw new Error('Handler function must have the `uniqueName` property.');
    }

    if (!this.syncPlugins.has(eventName)) {
      this.syncPlugins.set(eventName, new Map());
    }
    this.syncPlugins.get(eventName)?.set(identifier, handler);
  }

  /**
   * Unregister a plugin with the given name and event.
   *
   * @param eventName - The name of the event to unregister the plugin from.
   * @param handlerName - The name of the handler function to be removed.
   */
  public unregister(eventName: string, handlerName: string): void {
    const removeHandler = <T extends AsyncPluginHandler | SyncPluginHandler>(map: Map<string, Map<string, T>>) => {
      const handlers = map.get(eventName);

      if (handlers?.has(handlerName)) {
        handlers.delete(handlerName);
        if (handlers.size === 0) {
          map.delete(eventName); // Remove the event if no handlers are left.
        }
      }
    };

    removeHandler(this.asyncPlugins);
    removeHandler(this.syncPlugins);
  }

  /**
   * Executes all registered plugins for the given event with the provided arguments.
   *
   * @param eventName - The name of the event to execute plugins for.
   * @param args - The arguments to pass to the plugin handlers.
   * @returns True if all plugins returned true, false otherwise.
   */
  public executeSync(eventName: string, ...args: unknown[]): boolean {
    const handlersMap = this.syncPlugins.get(eventName);

    if (!handlersMap) {
      return true; // No plugins registered, consider it a success.
    }

    const handlers = Array.from(handlersMap.values());

    return handlers.every((handler) => handler(...args));
  }

  /**
   * Executes all registered plugins for the given event with the provided arguments.
   *
   * @param eventName - The name of the event to execute plugins for.
   * @param args - The arguments to pass to the plugin handlers.
   * @returns True if all plugins returned true, false otherwise.
   */
  public executeAsync(eventName: string, ...args: unknown[]): Promise<boolean>;

  /**
   * Executes all registered plugins for the given event with the provided arguments and debounces the execution.
   *
   * @param eventName - The name of the event to execute plugins for.
   * @param debounceTime - The time in milliseconds to debounce the execution of plugins.
   * @param args - The arguments to pass to the plugin handlers.
   * @returns True if all plugins returned true, false otherwise.
   */
  public executeAsync(eventName: string, debounceTime: number, ...args: unknown[]): Promise<boolean>;

  public executeAsync(eventName: string, ...args: unknown[]): Promise<boolean> {
    // Check if the first argument is a number (debounce time).
    if (args.length > 1 && typeof args[0] === 'number') {
      const allArgs = [...args];
      const debounceTime: number = allArgs.shift() as number;

      return debounce(() => this.executeAllAsync(eventName, ...allArgs), debounceTime)() ?? Promise.resolve(true); // Debounce execution to avoid rapid calls.
    }
    // If no debounce time is provided, execute the plugins immediately.
    return this.executeAllAsync(eventName, ...args);
  }

  /**
   * Executes all registered plugins for the given event with the provided arguments.
   *
   * @param eventName - The name of the event to execute plugins for.
   * @param args - The arguments to pass to the plugin handlers.
   * @returns True if all plugins returned true, false otherwise.
   */
  private async executeAllAsync(eventName: string, ...args: unknown[]): Promise<boolean> {
    const handlersMap = this.asyncPlugins.get(eventName);

    if (!handlersMap) {
      return true; // No plugins registered, consider it a success.
    }

    const handlers = Array.from(handlersMap.values());
    const results = await Promise.all(handlers.map((handler) => handler(...args)));

    return results.every((result) => result);
  }
}

export default PluginRegistry;
