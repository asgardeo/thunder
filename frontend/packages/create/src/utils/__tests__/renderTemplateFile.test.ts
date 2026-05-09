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

import {existsSync, mkdirSync, writeFileSync, rmSync} from 'fs';
import {tmpdir} from 'os';
import {join} from 'path';
import {describe, it, expect, beforeAll, afterEach} from 'vitest';
import registerHandlebarsHelpers from '../registerHandlebarsHelpers';
import renderTemplateFile from '../renderTemplateFile';

describe('renderTemplateFile', () => {
  const testDir = join(tmpdir(), 'create-test-render-template-file');

  beforeAll(() => {
    registerHandlebarsHelpers();
  });

  afterEach(() => {
    if (existsSync(testDir)) {
      rmSync(testDir, {recursive: true, force: true});
    }
  });

  it('should render a template file with context', () => {
    mkdirSync(testDir, {recursive: true});
    const templatePath = join(testDir, 'template.hbs');
    writeFileSync(templatePath, 'Hello, {{name}}!');

    const result = renderTemplateFile(templatePath, {name: 'ThunderID'});

    expect(result).toBe('Hello, ThunderID!');
  });

  it('should throw error if template file does not exist', () => {
    const nonExistentPath = join(testDir, 'non-existent.hbs');

    expect(() => renderTemplateFile(nonExistentPath, {})).toThrow(`Template file not found: ${nonExistentPath}`);
  });

  it('should handle complex template files', () => {
    mkdirSync(testDir, {recursive: true});
    const templatePath = join(testDir, 'complex.hbs');
    const templateContent = `
/**
 * Copyright (c) {{year}}, {{company}}
 */

export interface {{pascalCase name}} {
  id: string;
  {{#each fields}}
  {{this.name}}: {{this.type}};
  {{/each}}
}
`;

    writeFileSync(templatePath, templateContent);

    const result = renderTemplateFile(templatePath, {
      year: '2026',
      company: 'WSO2',
      name: 'user-profile',
      fields: [
        {name: 'username', type: 'string'},
        {name: 'email', type: 'string'},
        {name: 'age', type: 'number'},
      ],
    });

    expect(result).toContain('Copyright (c) 2026, WSO2');
    expect(result).toContain('export interface UserProfile {');
    expect(result).toContain('username: string;');
    expect(result).toContain('email: string;');
    expect(result).toContain('age: number;');
  });

  it('should handle template files with conditional logic', () => {
    mkdirSync(testDir, {recursive: true});
    const templatePath = join(testDir, 'conditional.hbs');
    const templateContent = `
{{#if includeTests}}
import {describe, it} from 'vitest';
{{/if}}

export function {{camelCase name}}() {
  // Implementation
}
`;

    writeFileSync(templatePath, templateContent);

    const withTests = renderTemplateFile(templatePath, {
      name: 'my-function',
      includeTests: true,
    });

    const withoutTests = renderTemplateFile(templatePath, {
      name: 'my-function',
      includeTests: false,
    });

    expect(withTests).toContain("import {describe, it} from 'vitest';");
    expect(withTests).toContain('export function myFunction()');
    expect(withoutTests).not.toContain('import');
    expect(withoutTests).toContain('export function myFunction()');
  });

  it('should preserve file encoding and handle multi-line templates', () => {
    mkdirSync(testDir, {recursive: true});
    const templatePath = join(testDir, 'multiline.hbs');
    const templateContent = `Line 1: {{var1}}
Line 2: {{var2}}
Line 3: {{var3}}`;

    writeFileSync(templatePath, templateContent, 'utf8');

    const result = renderTemplateFile(templatePath, {
      var1: 'Value1',
      var2: 'Value2',
      var3: 'Value3',
    });

    expect(result).toBe(`Line 1: Value1
Line 2: Value2
Line 3: Value3`);
  });
});
