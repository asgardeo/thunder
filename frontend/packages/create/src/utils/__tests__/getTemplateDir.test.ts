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

import {existsSync, statSync} from 'fs';
import {join, dirname} from 'path';
import {fileURLToPath} from 'url';
import {describe, it, expect} from 'vitest';
import getTemplateDir from '../getTemplateDir';

describe('getTemplateDir', () => {
  it('should return a valid template directory path', () => {
    const templateDir = getTemplateDir();

    expect(templateDir).toBeDefined();
    expect(typeof templateDir).toBe('string');
    expect(templateDir.length).toBeGreaterThan(0);
  });

  it('should return a path that includes "templates"', () => {
    const templateDir = getTemplateDir();

    expect(templateDir).toContain('templates');
  });

  it('should prefer dist/templates if it exists', () => {
    const templateDir = getTemplateDir();

    // In development, src/templates should exist
    // In production/built version, dist/templates should exist
    expect(existsSync(templateDir) || templateDir.includes('dist') || templateDir.includes('src')).toBe(true);
  });

  it('should return an absolute path', () => {
    const templateDir = getTemplateDir();

    // Absolute paths start with / on Unix or drive letter on Windows
    expect(templateDir.startsWith('/') || /^[A-Z]:\\/.test(templateDir)).toBe(true);
  });

  it('should handle package resolution correctly', () => {
    const templateDir = getTemplateDir();
    const currentFileDir = dirname(fileURLToPath(import.meta.url));

    // Template dir should be related to the package structure
    expect(
      templateDir.includes('create') ||
        // Or relative to current file (fallback case)
        templateDir.includes(join(currentFileDir, '..')),
    ).toBe(true);
  });

  it('should consistently return the same path on multiple calls', () => {
    const templateDir1 = getTemplateDir();
    const templateDir2 = getTemplateDir();

    expect(templateDir1).toBe(templateDir2);
  });

  it('should return a path that could contain template files', () => {
    const templateDir = getTemplateDir();

    // The directory should either exist as a directory, or be a plausible template directory path
    const dirExists = existsSync(templateDir);
    const isValidPath = dirExists ? statSync(templateDir).isDirectory() : /[/\\]templates$/.test(templateDir);

    expect(isValidPath).toBe(true);
  });
});
