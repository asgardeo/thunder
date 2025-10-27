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

const config = {
  embeddedLanguageFormatting: 'auto',
  arrowParens: 'always',
  bracketSpacing: false,
  endOfLine: 'lf',
  htmlWhitespaceSensitivity: 'css',
  insertPragma: false,
  jsxBracketSameLine: false,
  jsxSingleQuote: false,
  printWidth: 120,
  proseWrap: 'always',
  quoteProps: 'as-needed',
  requirePragma: false,
  semi: true,
  singleQuote: true,
  tabWidth: 2,
  // Adds trailing commas at the end of all objects.
  // 1. Easier to add an item or re-order items.
  // 2. Removes the need to have one line item be special because it lacks the ,.
  // 3. Allows for cleaner Git diffs. (New properties can be added without modifying the previous last line.)
  // https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Trailing_commas
  trailingComma: 'all',
  useTabs: false,
};

export default config;
