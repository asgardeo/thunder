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
 * The BCP 47 locale code used as the fallback language across all applications.
 */
const I18nDefaultConstants = {
  /**
   * The default fallback language for i18n translations. This is used when a translation for the active locale is unavailable.
   */
  FALLBACK_LANGUAGE: 'en-US',
} as const;

export default I18nDefaultConstants;
