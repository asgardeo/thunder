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

import {useContext} from 'react';
import ApplicationCreateContext, {
  type ApplicationCreateContextType,
} from '../contexts/ApplicationCreate/ApplicationCreateContext';

/**
 * Custom React hook to access the Application Create context.
 *
 * This hook provides access to all state and actions needed for the application
 * creation flow. It must be used within an ApplicationCreateProvider component.
 *
 * @throws {Error} If used outside of an ApplicationCreateProvider
 * @returns The application creation context containing state and actions
 *
 * @example
 * ```tsx
 * import useApplicationCreateContext from '@/features/applications/hooks/useApplicationCreateContext';
 *
 * function MyComponent() {
 *   const {
 *     appName,
 *     setAppName,
 *     currentStep,
 *     selectedAuthFlow,
 *     setSelectedAuthFlow
 *   } = useApplicationCreateContext();
 *
 *   return (
 *     <div>
 *       <h1>Creating: {appName}</h1>
 *       <p>Step: {currentStep}</p>
 *       {selectedAuthFlow && <p>Selected Flow: {selectedAuthFlow.name}</p>}
 *     </div>
 *   );
 * }
 * ```
 *
 * @public
 */
export default function useApplicationCreateContext(): ApplicationCreateContextType {
  const context = useContext(ApplicationCreateContext);

  if (!context) {
    throw new Error(
      'useApplicationCreateContext must be used within an ApplicationCreateProvider. ' +
        'Make sure your component is wrapped with ApplicationCreateProvider.',
    );
  }

  return context;
}
