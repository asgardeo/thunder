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

import {useTranslation} from 'react-i18next';
import type {SupportedLanguage, LanguageConfig} from '../models';
import {LANGUAGE_CONFIGS} from '../models';

export interface UseLanguageReturn {
  currentLanguage: SupportedLanguage;
  availableLanguages: LanguageConfig[];
  setLanguage: (language: SupportedLanguage) => Promise<void>;
}

/**
 * Hook to manage language switching in the application.
 * Uses react-i18next internally for language management.
 *
 * @returns {UseLanguageReturn} Language management utilities
 *
 * @example
 * ```tsx
 * function LanguageSwitcher() {
 *   const { currentLanguage, availableLanguages, setLanguage } = useLanguage();
 *
 *   return (
 *     <select
 *       value={currentLanguage}
 *       onChange={(e) => setLanguage(e.target.value as SupportedLanguage)}
 *     >
 *       {availableLanguages.map((lang) => (
 *         <option key={lang.code} value={lang.code}>
 *           {lang.nativeName}
 *         </option>
 *       ))}
 *     </select>
 *   );
 * }
 * ```
 */
export default function useLanguage(): UseLanguageReturn {
  const {i18n} = useTranslation();

  const currentLanguage = i18n.language as SupportedLanguage;
  const availableLanguages = Object.values(LANGUAGE_CONFIGS);

  const setLanguage = async (language: SupportedLanguage): Promise<void> => {
    await i18n.changeLanguage(language);
  };

  return {
    currentLanguage,
    availableLanguages,
    setLanguage,
  };
}
