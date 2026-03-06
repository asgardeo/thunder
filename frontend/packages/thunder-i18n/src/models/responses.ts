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

/**
 * Response from the languages API.
 */
export interface LanguagesResponse {
  /**
   * List of available language codes (e.g., ["en", "fr", "de"])
   */
  languages: string[];
}

/**
 * Response from the translations API.
 */
export interface TranslationsResponse {
  /**
   * The language code for the translations (e.g., "en").
   */
  language: string;
  /**
   * Total number of translation keys (optional).
   */
  totalResults?: number;
  /**
   * Translations object: { namespace: { key: value } }.
   */
  translations: Record<string, Record<string, string>>;
}

/**
 * Response from the translation API.
 */
export interface TranslationResponse {
  /**
   * The language code for the translation (e.g., "en").
   */
  language: string;
  /**
   * The namespace of the translation key.
   */
  namespace: string;
  /**
   * The translation key.
   */
  key: string;
  /**
   * The translation value.
   */
  value: string;
}
