/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

import {type Context, createContext} from 'react';

/**
 * Severity level for a toast notification.
 *
 * @public
 */
export type ToastSeverity = 'success' | 'error' | 'warning' | 'info';

/**
 * Toast context interface that provides access to the global toast notification system.
 *
 * @public
 */
export interface ToastContextType {
  /**
   * Displays a toast notification with the given message and severity.
   *
   * @param message - The text to display inside the toast
   * @param severity - The visual style of the toast. Defaults to `'success'`
   */
  showToast: (message: string, severity?: ToastSeverity) => void;
}

/**
 * React context for triggering toast notifications from anywhere in the component tree.
 *
 * This context provides a `showToast` function that renders a temporary snackbar message
 * at the bottom-right of the screen. It should be consumed via the `useToast` hook
 * inside a component tree wrapped by `ToastProvider`.
 *
 * @example
 * ```tsx
 * import ToastContext from './ToastContext';
 * import { useContext } from 'react';
 *
 * const MyComponent = () => {
 *   const context = useContext(ToastContext);
 *   if (!context) {
 *     throw new Error('Component must be used within ToastProvider');
 *   }
 *
 *   return <button onClick={() => context.showToast('Done!', 'success')}>Save</button>;
 * };
 * ```
 *
 * @public
 */
const ToastContext: Context<ToastContextType | undefined> = createContext<ToastContextType | undefined>(undefined);

export default ToastContext;
