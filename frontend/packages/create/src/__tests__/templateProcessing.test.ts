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

import {existsSync} from 'fs';
import {mkdir, rm, readFile, stat, writeFile} from 'fs/promises';
import {tmpdir} from 'os';
import {join} from 'path';
import {createLogger} from '@thunderid/logger';
import {describe, it, expect, beforeEach, afterEach} from 'vitest';
import createFileFromTemplate from '../utils/createFileFromTemplate';
import ensureDir from '../utils/ensureDir';
import getTemplateDir from '../utils/getTemplateDir';
import registerHandlebarsHelpers from '../utils/registerHandlebarsHelpers';

const logger = createLogger();

describe('Template Processing', () => {
  let testDir: string;

  beforeEach(async () => {
    // Register Handlebars helpers
    registerHandlebarsHelpers();
    // Create a unique test directory
    testDir = join(tmpdir(), `create-test-${Date.now()}`);
    await mkdir(testDir, {recursive: true});
  });

  afterEach(async () => {
    // Clean up test directory
    try {
      await rm(testDir, {recursive: true, force: true});
    } catch (error: unknown) {
      logger.error(`Failed to clean up test directory: ${error instanceof Error ? error.message : String(error)}`);
    }
  });

  it('should create file from template with correct content', async () => {
    const templateContent = `/**
 * {{featureName}} component
 */
export const {{pascalCase featureName}} = () => {
  return <div>{{featureName}}</div>;
};`;

    const templatePath = join(testDir, 'template.tsx.hbs');
    const outputPath = join(testDir, 'Component.tsx');

    // Create template file
    await writeFile(templatePath, templateContent);

    const data = {featureName: 'TestFeature'};

    createFileFromTemplate(templatePath, outputPath, data);

    const result = await readFile(outputPath, 'utf-8');

    expect(result).toContain('TestFeature component');
    expect(result).toContain('export const TestFeature =');
    expect(result).toContain('<div>TestFeature</div>');
    expect(result).not.toContain('{{');
  });

  it('should handle nested directory creation', () => {
    const nestedPath = join(testDir, 'src', 'components', 'nested');

    ensureDir(nestedPath);

    expect(existsSync(nestedPath)).toBe(true);
  });

  it('should detect file existence correctly', async () => {
    const filePath = join(testDir, 'test-file.txt');

    // File doesn't exist initially
    expect(existsSync(filePath)).toBe(false);

    // Create file
    await writeFile(filePath, 'test content');

    // Now it should exist
    expect(existsSync(filePath)).toBe(true);
  });

  it('should get correct template directory', () => {
    const templateDir = getTemplateDir();
    expect(templateDir).toContain('templates');
  });

  it('should handle complex template with all helpers', async () => {
    const templateContent = `// {{constantCase featureName}} Constants
export const {{constantCase featureName}}_PREFIX = '{{kebabCase featureName}}';

export interface {{pascalCase featureName}} {
  name: string;
}

export const use{{pascalCase featureName}} = () => {
  // Implementation for {{camelCase featureName}}
};`;

    const templatePath = join(testDir, 'template.ts.hbs');
    const outputPath = join(testDir, 'output.ts');

    await writeFile(templatePath, templateContent);

    const data = {featureName: 'userManagement'};

    createFileFromTemplate(templatePath, outputPath, data);

    const result = await readFile(outputPath, 'utf-8');

    expect(result).toContain('USER_MANAGEMENT_PREFIX');
    expect(result).toContain("'user-management'");
    expect(result).toContain('interface UserManagement');
    expect(result).toContain('useUserManagement =');
    expect(result).toContain('// Implementation for userManagement');
  });

  it('should preserve file permissions when creating from template', async () => {
    const templateContent = '#!/bin/bash\necho "{{featureName}}"';
    const templatePath = join(testDir, 'script.sh.hbs');
    const outputPath = join(testDir, 'script.sh');

    await writeFile(templatePath, templateContent);

    const data = {featureName: 'test'};

    createFileFromTemplate(templatePath, outputPath, data);

    const stats = await stat(outputPath);
    expect(stats.isFile()).toBe(true);

    const result = await readFile(outputPath, 'utf-8');
    expect(result).toContain('echo "test"');
  });
});
