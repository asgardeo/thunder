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

import {readdir, readFile, stat} from 'fs/promises';
import {join} from 'path';
import Handlebars from 'handlebars';
import {describe, it, expect, beforeAll} from 'vitest';
import registerHandlebarsHelpers from '../utils/registerHandlebarsHelpers';

describe('Template Validation', () => {
  beforeAll(() => {
    // Register Handlebars helpers before testing
    registerHandlebarsHelpers();
  });

  const getTemplateFiles = async (dir: string): Promise<string[]> => {
    const entries = await readdir(dir);

    const results = await Promise.all(
      entries.map(async (entry) => {
        const fullPath = join(dir, entry);
        const stats = await stat(fullPath);

        if (stats.isDirectory()) {
          return getTemplateFiles(fullPath);
        }

        return entry.endsWith('.hbs') ? [fullPath] : [];
      }),
    );

    return results.flat();
  };

  it('should have valid Handlebars syntax in all template files', async () => {
    const templatesDir = join(__dirname, '..', 'templates', 'feature');
    const templateFiles = await getTemplateFiles(templatesDir);

    expect(templateFiles.length).toBeGreaterThan(0);

    await Promise.all(
      templateFiles.map(async (filePath) => {
        const content = await readFile(filePath, 'utf-8');

        try {
          // This will throw if there's a syntax error
          Handlebars.compile(content);
        } catch (error: unknown) {
          const msg = error instanceof Error ? error.message : String(error);
          throw new Error(`Handlebars syntax error in ${filePath}: ${msg}`, {cause: error});
        }
      }),
    );
  });

  it('should compile templates with sample data without errors', async () => {
    const templatesDir = join(__dirname, '..', 'templates', 'feature');
    const templateFiles = await getTemplateFiles(templatesDir);

    const sampleData = {
      featureName: 'TestFeature',
      packageName: 'thunder-admin-test-feature',
      description: 'Test feature for validation',
    };

    await Promise.all(
      templateFiles.map(async (filePath) => {
        const content = await readFile(filePath, 'utf-8');

        try {
          const template = Handlebars.compile(content);
          const result = template(sampleData);

          // Ensure the result is a string and not empty for most files
          expect(typeof result).toBe('string');

          // For non-gitkeep files, ensure some content was generated
          expect(filePath.includes('gitkeep') || result.trim().length > 0).toBe(true);
        } catch (error: unknown) {
          const msg = error instanceof Error ? error.message : String(error);
          throw new Error(`Template compilation error in ${filePath}: ${msg}`, {cause: error});
        }
      }),
    );
  });

  it('should not have any unescaped triple braces outside of comments', async () => {
    const templatesDir = join(__dirname, '..', 'templates', 'feature');
    const templateFiles = await getTemplateFiles(templatesDir);

    expect(templateFiles.length).toBeGreaterThan(0);

    await Promise.all(
      templateFiles.map(async (filePath) => {
        const content = await readFile(filePath, 'utf-8');
        const lines = content.split('\n');

        lines.forEach((line, idx) => {
          const lineNumber = idx + 1;

          // Skip comment lines
          if (line.trim().startsWith('//') || line.trim().startsWith('*') || line.trim().startsWith('/*')) {
            return;
          }

          // Check for problematic patterns
          if (line.includes('{{{') && !line.includes('{{{{')) {
            // Allow triple braces only in specific contexts like JSDoc examples in comments
            if (!line.includes('*') && !line.includes('//')) {
              throw new Error(
                `Found unescaped triple braces at ${filePath}:${lineNumber} - "${line.trim()}". ` +
                  'Use double braces {{}} for variable interpolation or escape triple braces.',
              );
            }
          }
        });
      }),
    );
  });

  it('should not have template literal syntax conflicts', async () => {
    const templatesDir = join(__dirname, '..', 'templates', 'feature');
    const templateFiles = await getTemplateFiles(templatesDir);

    expect(templateFiles.length).toBeGreaterThan(0);

    await Promise.all(
      templateFiles.map(async (filePath) => {
        const content = await readFile(filePath, 'utf-8');
        const lines = content.split('\n');

        lines.forEach((line, idx) => {
          const lineNumber = idx + 1;

          // Check for template literal syntax that could conflict with Handlebars
          if (line.includes('`') && line.includes('${')) {
            // Skip comment lines
            if (line.trim().startsWith('//') || line.trim().startsWith('*') || line.trim().startsWith('/*')) {
              return;
            }

            throw new Error(
              `Found template literal syntax at ${filePath}:${lineNumber} - "${line.trim()}". ` +
                'Use string concatenation instead to avoid conflicts with Handlebars.',
            );
          }
        });
      }),
    );
  });

  it('should not have JSX object syntax conflicts', async () => {
    const templatesDir = join(__dirname, '..', 'templates', 'feature');
    const templateFiles = await getTemplateFiles(templatesDir);

    expect(templateFiles.length).toBeGreaterThan(0);

    await Promise.all(
      templateFiles.map(async (filePath) => {
        const content = await readFile(filePath, 'utf-8');
        const lines = content.split('\n');

        lines.forEach((line, idx) => {
          const lineNumber = idx + 1;

          // Skip comment lines and example code
          if (line.trim().startsWith('//') || line.trim().startsWith('*') || line.trim().startsWith('/*')) {
            return;
          }

          // Check for JSX props with double braces that aren't escaped
          // Skip props that are already properly escaped with backslash
          const jsxPropsPattern = /\w+={[^\\]/;
          if (jsxPropsPattern.test(line) && !line.includes('=\\{{')) {
            throw new Error(
              `Found unescaped JSX props at ${filePath}:${lineNumber} - "${line.trim()}". ` +
                'Escape JSX object syntax to avoid Handlebars conflicts.',
            );
          }
        });
      }),
    );
  });

  it('should generate valid TypeScript for all TS/TSX templates', async () => {
    const templatesDir = join(__dirname, '..', 'templates', 'feature');
    const templateFiles = await getTemplateFiles(templatesDir);
    const tsFiles = templateFiles.filter((f) => f.endsWith('.ts.hbs') || f.endsWith('.tsx.hbs'));

    const sampleData = {
      featureName: 'TestFeature',
      packageName: 'thunder-admin-test-feature',
      description: 'Test feature for validation',
    };

    await Promise.all(
      tsFiles.map(async (filePath) => {
        const content = await readFile(filePath, 'utf-8');

        try {
          const template = Handlebars.compile(content);
          const result = template(sampleData);

          // Basic TypeScript syntax checks
          expect(result).toMatch(/import|export|interface|type|const|let|var|function|class/);
          expect(result).not.toMatch(/\{\{\{[^}]/); // No unprocessed triple braces

          // Ensure basic TypeScript keywords are present
          expect(result.length).toBeGreaterThan(50); // Should have substantial content
        } catch (error: unknown) {
          const msg = error instanceof Error ? error.message : String(error);
          throw new Error(`TypeScript template compilation error in ${filePath}: ${msg}`, {cause: error});
        }
      }),
    );
  });

  it('should have all required Handlebars helpers registered', () => {
    const testData = {featureName: 'testFeature'};

    // Test all helper functions
    const helpers = ['pascalCase', 'camelCase', 'kebabCase', 'constantCase'];

    helpers.forEach((helper) => {
      const template = Handlebars.compile(`{{${helper} featureName}}`);
      const result = template(testData);

      expect(typeof result).toBe('string');
      expect(result.length).toBeGreaterThan(0);
      expect(result).not.toEqual(`{{${helper} featureName}}`); // Should be processed
    });
  });

  it('should generate consistent package names', () => {
    const testCases = [
      {input: 'testFeature', expected: 'thunder-admin-test-feature'},
      {input: 'user-management', expected: 'thunder-admin-user-management'},
      {input: 'MyFeature', expected: 'thunder-admin-my-feature'},
    ];

    testCases.forEach((testCase) => {
      const template = Handlebars.compile('thunder-admin-{{kebabCase featureName}}');
      const result = template({featureName: testCase.input});

      expect(result).toBe(testCase.expected);
    });
  });
});
