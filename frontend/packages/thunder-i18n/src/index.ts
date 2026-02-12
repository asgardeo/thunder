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

/**
 * Thunder i18n - Translation resources for Thunder applications
 *
 * This package provides translation resources organized by language and namespace.
 * Applications should use react-i18next directly for i18n functionality.
 *
 * @example Using translations with react-i18next
 * ```tsx
 * import i18n from 'i18next';
 * import { initReactI18next } from 'react-i18next';
 * import enUS from '@thunder/i18n/locales/en-US';
 *
 * i18n
 *   .use(initReactI18next)
 *   .init({
 *     resources: {
 *       'en-US': {
 *         common: enUS.common,
 *         develop: enUS.develop,
 *         gate: enUS.gate,
 *       },
 *     },
 *     lng: 'en-US',
 *     fallbackLng: 'en-US',
 *     defaultNS: 'common',
 *     interpolation: {
 *       escapeValue: false,
 *     },
 *   });
 * ```
 *
 * @example Using in components with react-i18next
 * ```tsx
 * import { useTranslation } from 'react-i18next';
 *
 * function MyComponent() {
 *   const { t } = useTranslation('common');
 *   return <h1>{t('actions.save')}</h1>;
 * }
 * ```
 */

// Export translation resources
export {default as enUS} from './locales/en-US';

// Export types and models
export type {
  TranslationResources,
  SupportedLanguage,
  LanguageConfig,
  ResourceValue,
  NamespaceResources,
} from './models';

export {LANGUAGE_CONFIGS, isSupportedLanguage} from './models';

// Export hooks
export {default as useGetTranslations} from './hooks/useGetTranslations';
export type {TranslationsResponse, UseGetTranslationsOptions} from './hooks/useGetTranslations';

export {default as useGetLanguages} from './hooks/useGetLanguages';
export type {LanguagesResponse, UseGetLanguagesOptions} from './hooks/useGetLanguages';

export {default as useUpdateTranslation} from './hooks/useUpdateTranslation';
export type {
  TranslationResponse,
  UpdateTranslationVariables,
  UseUpdateTranslationOptions,
} from './hooks/useUpdateTranslation';

export {default as useLanguage} from './hooks/useLanguage';
export type {UseLanguageReturn} from './hooks/useLanguage';

// Export constants
export {default as I18nQueryKeys} from './constants/I18nQueryKeys';
