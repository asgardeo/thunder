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

import {existsSync} from 'fs';
import {join, dirname} from 'path';
import {fileURLToPath} from 'url';

/**
 * Returns the absolute path to the template directory used for scaffolding feature and package modules.
 *
 * @returns The absolute path to the template directory
 *
 * @example
 * const templateDir = getTemplateDir();
 * // Use templateDir to locate scaffolding templates
 *
 * @public
 */
export default function getTemplateDir(): string {
  // For the linked global package, find the package root by looking for package.json
  let currentDir = dirname(fileURLToPath(import.meta.url));

  // Go up directories until we find package.json. Use a platform-neutral termination
  // condition to avoid infinite loops on Windows where the root isn't '/'.

  while (true) {
    if (existsSync(join(currentDir, 'package.json'))) {
      break;
    }
    const parentDir = dirname(currentDir);
    if (parentDir === currentDir) {
      // Reached filesystem root
      currentDir = '';
      break;
    }
    currentDir = parentDir;
  }

  // If we found package.json, templates should be in dist/templates or src/templates
  if (currentDir && existsSync(join(currentDir, 'package.json'))) {
    const distTemplates = join(currentDir, 'dist', 'templates');
    const srcTemplates = join(currentDir, 'src', 'templates');

    if (existsSync(distTemplates)) {
      return distTemplates;
    }
    if (existsSync(srcTemplates)) {
      return srcTemplates;
    }
  }

  // Fallback to current directory relative
  return join(dirname(fileURLToPath(import.meta.url)), '..', 'templates');
}
