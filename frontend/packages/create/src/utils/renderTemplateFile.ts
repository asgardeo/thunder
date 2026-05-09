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

import {readFileSync, existsSync} from 'fs';
import renderTemplate from './renderTemplate';
import type {TemplateContext} from '../models/templates';

/**
 * Renders a Handlebars template file with the provided context and writes the output to the target file.
 *
 * @param templatePath - Path to the Handlebars template file
 * @param targetPath - Path to write the rendered output file
 * @param context - Data context for template rendering
 *
 * @example
 * renderTemplateFile('template.hbs', 'output.ts', { name: 'Feature' });
 *
 * @public
 */
export default function renderTemplateFile(templatePath: string, context: TemplateContext): string {
  if (!existsSync(templatePath)) {
    throw new Error(`Template file not found: ${templatePath}`);
  }

  const templateContent = readFileSync(templatePath, 'utf8');
  return renderTemplate(templateContent, context);
}
