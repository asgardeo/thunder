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

import {mkdirSync, existsSync} from 'fs';

/**
 * Ensures that the specified directory exists. If it does not exist, it will be created recursively.
 *
 * @param dirPath - The path of the directory to ensure exists
 *
 * @example
 * ensureDir('/path/to/dir');
 * // Creates the directory if it does not exist
 *
 * @public
 */
export default function ensureDir(path: string): void {
  if (!existsSync(path)) {
    mkdirSync(path, {recursive: true});
  }
}
