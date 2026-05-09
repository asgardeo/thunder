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

import {statSync, chmodSync, existsSync, mkdirSync, writeFileSync} from 'fs';
import {dirname} from 'path';
import {createLogger} from '@thunderid/logger';
import renderTemplateFile from './renderTemplateFile';
import type {TemplateContext} from '../models/templates';

const logger = createLogger();

/**
 * Renders a single Handlebars template file and writes the output to the specified target file.
 *
 * @param templatePath - Path to the Handlebars template file
 * @param targetPath - Path to write the rendered output file
 * @param context - Data context for template rendering
 *
 * @example
 * createFileFromTemplate('template.hbs', 'output.ts', { name: 'Feature' });
 *
 * @public
 */
export default function createFileFromTemplate(
  templatePath: string,
  outputPath: string,
  context: TemplateContext,
): void {
  const content = renderTemplateFile(templatePath, context);

  // Ensure the directory exists
  const dir = dirname(outputPath);
  if (!existsSync(dir)) {
    mkdirSync(dir, {recursive: true});
  }

  writeFileSync(outputPath, content, 'utf8');
  try {
    const templateStats = statSync(templatePath);
    const mode = templateStats.mode & 0o777;
    chmodSync(outputPath, mode);
  } catch (err) {
    logger.warn(`Could not preserve file permissions for ${outputPath}: ${(err as Error).message}`);
  }
}
