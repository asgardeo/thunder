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
 * TypeScript type definitions for i18n
 */

import type enUS from '../locales/en-US';

/**
 * Translation resources type derived from English translations
 */
export type TranslationResources = {
  [K in keyof typeof enUS]: (typeof enUS)[K] extends Record<string, unknown>
    ? {[P in keyof (typeof enUS)[K]]: unknown}
    : unknown;
};

/**
 * Supported languages in applications
 */
export type SupportedLanguage = 'en-US';

/**
 * Language configuration
 */
export interface LanguageConfig {
  code: SupportedLanguage;
  name: string;
  nativeName: string;
  direction: 'ltr' | 'rtl';
}

/**
 * Available translations by language
 */
export interface Translations {
  'en-US': typeof enUS;
}

/**
 * Language metadata
 */
export const LANGUAGE_CONFIGS: Record<SupportedLanguage, LanguageConfig> = {
  'en-US': {
    code: 'en-US',
    name: 'English (US)',
    nativeName: 'English (US)',
    direction: 'ltr',
  },
};

/**
 * Type guard to check if a string is a supported language
 */
export function isSupportedLanguage(lang: string): lang is SupportedLanguage {
  return lang in LANGUAGE_CONFIGS;
}

/**
 * Recursive type for translation resource values
 */
export type ResourceValue = string | ((...args: unknown[]) => string) | {[key: string]: ResourceValue};

/**
 * Type for namespace resources mapping
 */
export type NamespaceResources = Record<string, ResourceValue>;
