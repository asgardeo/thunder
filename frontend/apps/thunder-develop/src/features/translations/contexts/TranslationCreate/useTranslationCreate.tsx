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

import {useContext} from 'react';
import TranslationCreateContext, {type TranslationCreateContextType} from './TranslationCreateContext';

/**
 * React hook for accessing translation creation state throughout the wizard.
 *
 * This hook provides access to all the state needed for the multi-step language
 * creation flow. It must be used within a component tree wrapped by
 * `TranslationCreateProvider`, otherwise it will throw an error.
 *
 * @returns The translation creation context containing state data and utility methods
 *
 * @throws {Error} Throws an error if used outside of TranslationCreateProvider
 *
 * @example
 * ```tsx
 * import useTranslationCreate from './useTranslationCreate';
 *
 * function MyComponent() {
 *   const { selectedCountry, currentStep, localeCode } = useTranslationCreate();
 *
 *   return (
 *     <div>
 *       <p>Current step: {currentStep}</p>
 *       <p>Locale: {localeCode}</p>
 *     </div>
 *   );
 * }
 * ```
 *
 * @public
 */
export default function useTranslationCreate(): TranslationCreateContextType {
  const context = useContext(TranslationCreateContext);

  if (context === undefined) {
    throw new Error('useTranslationCreate must be used within TranslationCreateProvider');
  }

  return context;
}
