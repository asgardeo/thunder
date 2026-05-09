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

import {existsSync, mkdirSync, writeFileSync, readFileSync, rmSync} from 'fs';
import {tmpdir} from 'os';
import {join} from 'path';
import {describe, it, expect, beforeAll, afterEach} from 'vitest';
import createFilesFromTemplates from '../createFilesFromTemplates';
import registerHandlebarsHelpers from '../registerHandlebarsHelpers';

describe('createFilesFromTemplates', () => {
  const testDir = join(tmpdir(), 'create-test-create-files');

  beforeAll(() => {
    registerHandlebarsHelpers();
  });

  afterEach(() => {
    if (existsSync(testDir)) {
      rmSync(testDir, {recursive: true, force: true});
    }
  });

  it('should create multiple files from templates', () => {
    mkdirSync(testDir, {recursive: true});

    const template1Path = join(testDir, 'template1.hbs');
    const template2Path = join(testDir, 'template2.hbs');
    const output1Path = join(testDir, 'output1.ts');
    const output2Path = join(testDir, 'output2.ts');

    writeFileSync(template1Path, 'export const {{name1}} = "{{value1}}";');
    writeFileSync(template2Path, 'export const {{name2}} = "{{value2}}";');

    createFilesFromTemplates([
      {
        templatePath: template1Path,
        outputPath: output1Path,
        context: {name1: 'constant1', value1: 'hello'},
      },
      {
        templatePath: template2Path,
        outputPath: output2Path,
        context: {name2: 'constant2', value2: 'world'},
      },
    ]);

    expect(existsSync(output1Path)).toBe(true);
    expect(existsSync(output2Path)).toBe(true);

    const content1 = readFileSync(output1Path, 'utf8');
    const content2 = readFileSync(output2Path, 'utf8');

    expect(content1).toBe('export const constant1 = "hello";');
    expect(content2).toBe('export const constant2 = "world";');
  });

  it('should handle empty template array', () => {
    expect(() => createFilesFromTemplates([])).not.toThrow();
  });

  it('should create files in different directories', () => {
    mkdirSync(testDir, {recursive: true});

    const templatePath = join(testDir, 'template.hbs');
    const output1Path = join(testDir, 'dir1', 'output1.ts');
    const output2Path = join(testDir, 'dir2', 'subdir', 'output2.ts');

    writeFileSync(templatePath, 'File: {{fileName}}');

    createFilesFromTemplates([
      {
        templatePath,
        outputPath: output1Path,
        context: {fileName: 'output1'},
      },
      {
        templatePath,
        outputPath: output2Path,
        context: {fileName: 'output2'},
      },
    ]);

    expect(existsSync(output1Path)).toBe(true);
    expect(existsSync(output2Path)).toBe(true);
    expect(readFileSync(output1Path, 'utf8')).toBe('File: output1');
    expect(readFileSync(output2Path, 'utf8')).toBe('File: output2');
  });

  it('should handle templates with different contexts', () => {
    mkdirSync(testDir, {recursive: true});

    const templatePath = join(testDir, 'template.hbs');
    writeFileSync(
      templatePath,
      `export const {{camelCase name}} = {
  value: '{{value}}',
  count: {{count}}
};`,
    );

    const outputs = [
      {
        path: join(testDir, 'config1.ts'),
        context: {name: 'first-config', value: 'alpha', count: 1},
      },
      {
        path: join(testDir, 'config2.ts'),
        context: {name: 'second-config', value: 'beta', count: 2},
      },
      {
        path: join(testDir, 'config3.ts'),
        context: {name: 'third-config', value: 'gamma', count: 3},
      },
    ];

    createFilesFromTemplates(
      outputs.map((output) => ({
        templatePath,
        outputPath: output.path,
        context: output.context,
      })),
    );

    outputs.forEach((output) => {
      expect(existsSync(output.path)).toBe(true);
      const content = readFileSync(output.path, 'utf8');
      expect(content).toContain(`value: '${output.context.value}'`);
      expect(content).toContain(`count: ${output.context.count}`);
    });
  });

  it('should handle complex template structures', () => {
    mkdirSync(testDir, {recursive: true});

    const indexTemplate = join(testDir, 'index.hbs');
    const componentTemplate = join(testDir, 'component.hbs');
    const testTemplate = join(testDir, 'test.hbs');

    writeFileSync(indexTemplate, "export * from './{{name}}';");
    writeFileSync(
      componentTemplate,
      `export const {{pascalCase name}} = () => {
  return <div>{{description}}</div>;
};`,
    );
    writeFileSync(
      testTemplate,
      `import {describe, it} from 'vitest';
import {{pascalCase name}} from './{{name}}';

describe('{{pascalCase name}}', () => {
  it('should render', () => {
    // test
  });
});`,
    );

    const featureName = 'my-feature';
    const outputDir = join(testDir, 'output');

    createFilesFromTemplates([
      {
        templatePath: indexTemplate,
        outputPath: join(outputDir, 'index.ts'),
        context: {name: featureName},
      },
      {
        templatePath: componentTemplate,
        outputPath: join(outputDir, `${featureName}.tsx`),
        context: {name: featureName, description: 'My Feature Component'},
      },
      {
        templatePath: testTemplate,
        outputPath: join(outputDir, `${featureName}.test.tsx`),
        context: {name: featureName},
      },
    ]);

    expect(existsSync(join(outputDir, 'index.ts'))).toBe(true);
    expect(existsSync(join(outputDir, `${featureName}.tsx`))).toBe(true);
    expect(existsSync(join(outputDir, `${featureName}.test.tsx`))).toBe(true);

    const indexContent = readFileSync(join(outputDir, 'index.ts'), 'utf8');
    expect(indexContent).toContain("export * from './my-feature'");

    const componentContent = readFileSync(join(outputDir, `${featureName}.tsx`), 'utf8');
    expect(componentContent).toContain('export const MyFeature');
    expect(componentContent).toContain('My Feature Component');

    const testContent = readFileSync(join(outputDir, `${featureName}.test.tsx`), 'utf8');
    expect(testContent).toContain("import MyFeature from './my-feature'");
    expect(testContent).toContain("describe('MyFeature'");
  });
});
