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

// String Operations
export {default as generateRandomHumanReadableIdentifiers} from './string/generateRandomHumanReadableIdentifiers';

// Path Operations
export {default as isAbsoluteUrl} from './path/isAbsoluteUrl';
export {default as isRelativeUrl} from './path/isRelativeUrl';

// Object Operations
export {default as isEmpty} from './object/isEmpty';

// Template Pattern Utilities
export {default as isI18nTemplatePattern, I18N_PATTERN, I18N_KEY_PATTERN} from './template/isI18nTemplatePattern';
export {default as isMetaTemplatePattern, META_PATTERN, META_KEY_PATTERN} from './template/isMetaTemplatePattern';
export {
  default as parseTemplateLiteral,
  TEMPLATE_LITERAL_REGEX,
  FUNCTION_CALL_REGEX,
  TemplateLiteralType,
} from './template/parseTemplateLiteral';
export type {TemplateLiteralResult, TemplateLiteralHandlers} from './template/parseTemplateLiteral';
