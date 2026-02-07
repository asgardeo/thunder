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

import {useEffect} from 'react';

/**
 * Configuration for a keyboard shortcut.
 */
export interface ShortcutConfig {
  /** The key to listen for (e.g., 's', 'k', 'Escape') */
  key: string;
  /** Whether the Ctrl key must be pressed */
  ctrl?: boolean;
  /** Whether the Meta (Cmd on Mac, Win on Windows) key must be pressed */
  meta?: boolean;
  /** Whether the Shift key must be pressed */
  shift?: boolean;
  /** Whether the Alt key must be pressed */
  alt?: boolean;
  /** Callback to execute when the shortcut is triggered */
  callback: (event: KeyboardEvent) => void;
  /** Description of what this shortcut does (for documentation) */
  description?: string;
  /** Whether to prevent default browser behavior */
  preventDefault?: boolean;
}

/**
 * Hook to register keyboard shortcuts with proper cleanup.
 * Automatically handles event listener registration and cleanup.
 *
 * @param shortcuts - Array of shortcut configurations
 *
 * @example
 * ```tsx
 * useKeyboardShortcuts([
 *   {
 *     key: 's',
 *     meta: true, // Cmd on Mac, Ctrl on Windows
 *     callback: handleSave,
 *     description: 'Save theme',
 *     preventDefault: true,
 *   },
 *   {
 *     key: 'k',
 *     meta: true,
 *     callback: () => searchInputRef.current?.focus(),
 *     description: 'Focus search',
 *     preventDefault: true,
 *   },
 *   {
 *     key: 'Escape',
 *     callback: handleClose,
 *     description: 'Close dialog',
 *   },
 * ]);
 * ```
 */
export default function useKeyboardShortcuts(shortcuts: ShortcutConfig[]): void {
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent): void => {
      for (const shortcut of shortcuts) {
        const keyMatches = event.key.toLowerCase() === shortcut.key.toLowerCase();
        const ctrlMatches = shortcut.ctrl ? event.ctrlKey : true;
        const metaMatches = shortcut.meta ? event.metaKey : true;
        const shiftMatches = shortcut.shift ? event.shiftKey : true;
        const altMatches = shortcut.alt ? event.altKey : true;

        // Check if modifiers that shouldn't be pressed are actually not pressed
        const noExtraCtrl = shortcut.ctrl === undefined ? true : event.ctrlKey === shortcut.ctrl;
        const noExtraMeta = shortcut.meta === undefined ? true : event.metaKey === shortcut.meta;
        const noExtraShift = shortcut.shift === undefined ? true : event.shiftKey === shortcut.shift;
        const noExtraAlt = shortcut.alt === undefined ? true : event.altKey === shortcut.alt;

        if (
          keyMatches &&
          ctrlMatches &&
          metaMatches &&
          shiftMatches &&
          altMatches &&
          noExtraCtrl &&
          noExtraMeta &&
          noExtraShift &&
          noExtraAlt
        ) {
          if (shortcut.preventDefault !== false) {
            event.preventDefault();
          }
          shortcut.callback(event);
          break;
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);

    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [shortcuts]);
}
