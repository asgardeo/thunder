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

import {existsSync, mkdirSync, writeFileSync, readFileSync, rmSync, statSync, chmodSync} from 'fs';
import {tmpdir} from 'os';
import {join} from 'path';
import {describe, it, expect, beforeAll, afterEach} from 'vitest';
import createFileFromTemplate from '../createFileFromTemplate';
import registerHandlebarsHelpers from '../registerHandlebarsHelpers';

describe('createFileFromTemplate', () => {
  const testDir = join(tmpdir(), 'create-test-create-file');

  beforeAll(() => {
    registerHandlebarsHelpers();
  });

  afterEach(() => {
    if (existsSync(testDir)) {
      rmSync(testDir, {recursive: true, force: true});
    }
  });

  it('should create a file from template', () => {
    mkdirSync(testDir, {recursive: true});
    const templatePath = join(testDir, 'template.hbs');
    const outputPath = join(testDir, 'output.ts');

    writeFileSync(templatePath, 'export const {{name}} = "{{value}}";');

    createFileFromTemplate(templatePath, outputPath, {
      name: 'myConstant',
      value: 'hello',
    });

    expect(existsSync(outputPath)).toBe(true);
    const content = readFileSync(outputPath, 'utf8');
    expect(content).toBe('export const myConstant = "hello";');
  });

  it('should create directory structure if it does not exist', () => {
    mkdirSync(testDir, {recursive: true});
    const templatePath = join(testDir, 'template.hbs');
    const outputPath = join(testDir, 'nested', 'deep', 'output.ts');

    writeFileSync(templatePath, 'content');

    createFileFromTemplate(templatePath, outputPath, {});

    expect(existsSync(outputPath)).toBe(true);
    expect(existsSync(join(testDir, 'nested'))).toBe(true);
    expect(existsSync(join(testDir, 'nested', 'deep'))).toBe(true);
  });

  it('should preserve file permissions from template', () => {
    mkdirSync(testDir, {recursive: true});
    const templatePath = join(testDir, 'template.sh');
    const outputPath = join(testDir, 'output.sh');

    writeFileSync(templatePath, '#!/bin/bash\necho "{{message}}"');

    chmodSync(templatePath, 0o755); // Make executable

    createFileFromTemplate(templatePath, outputPath, {message: 'Hello'});

    expect(existsSync(outputPath)).toBe(true);
    const outputStats = statSync(outputPath);
    const templateStats = statSync(templatePath);

    expect(outputStats.mode & 0o777).toBe(templateStats.mode & 0o777);
  });

  it('should handle complex templates with multiple variables', () => {
    mkdirSync(testDir, {recursive: true});
    const templatePath = join(testDir, 'component.hbs');
    const outputPath = join(testDir, 'Component.tsx');

    const templateContent = `
/**
 * {{description}}
 */
import React from 'react';

export interface {{pascalCase name}}Props {
  {{#each props}}
  {{this.name}}: {{this.type}};
  {{/each}}
}

export const {{pascalCase name}}: React.FC<{{pascalCase name}}Props> = (props) => {
  return <div>{{pascalCase name}}</div>;
};
`;

    writeFileSync(templatePath, templateContent);

    createFileFromTemplate(templatePath, outputPath, {
      name: 'my-component',
      description: 'A sample component',
      props: [
        {name: 'title', type: 'string'},
        {name: 'count', type: 'number'},
      ],
    });

    expect(existsSync(outputPath)).toBe(true);
    const content = readFileSync(outputPath, 'utf8');
    expect(content).toContain('A sample component');
    expect(content).toContain('export interface MyComponentProps');
    expect(content).toContain('title: string;');
    expect(content).toContain('count: number;');
    expect(content).toContain('export const MyComponent: React.FC<MyComponentProps>');
  });

  it('should overwrite existing files', () => {
    mkdirSync(testDir, {recursive: true});
    const templatePath = join(testDir, 'template.hbs');
    const outputPath = join(testDir, 'output.txt');

    writeFileSync(templatePath, 'New content: {{value}}');
    writeFileSync(outputPath, 'Old content');

    createFileFromTemplate(templatePath, outputPath, {value: 'updated'});

    const content = readFileSync(outputPath, 'utf8');
    expect(content).toBe('New content: updated');
    expect(content).not.toContain('Old content');
  });

  it('should handle empty context', () => {
    mkdirSync(testDir, {recursive: true});
    const templatePath = join(testDir, 'template.hbs');
    const outputPath = join(testDir, 'output.txt');

    writeFileSync(templatePath, 'Static content without variables');

    createFileFromTemplate(templatePath, outputPath, {});

    expect(existsSync(outputPath)).toBe(true);
    const content = readFileSync(outputPath, 'utf8');
    expect(content).toBe('Static content without variables');
  });
});
