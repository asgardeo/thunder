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

/* eslint-disable @typescript-eslint/naming-convention */
/* eslint-disable no-underscore-dangle */

import {describe, it, expect, beforeEach, afterEach} from 'vitest';
import {mkdir, rm, writeFile} from 'fs/promises';
import {join} from 'path';
import {tmpdir} from 'os';
import {exec} from 'child_process';
import {promisify} from 'util';
import kebabCase from 'lodash-es/kebabCase';
import validateName from '../utils/validateName';

const execAsync = promisify(exec);

describe('CLI Integration', () => {
  let testWorkspaceDir: string;

  beforeEach(async () => {
    // Create a temporary workspace
    testWorkspaceDir = join(tmpdir(), `thunder-test-workspace-${Date.now()}`);
    await mkdir(testWorkspaceDir, {recursive: true});

    // Create a basic package.json to simulate a frontend workspace
    const packageJson = {
      name: 'test-workspace',
      workspaces: ['packages/*'],
      devDependencies: {
        '@thunder/create': 'workspace:^',
      },
    };

    await writeFile(join(testWorkspaceDir, 'package.json'), JSON.stringify(packageJson, null, 2));

    // Create packages directory
    await mkdir(join(testWorkspaceDir, 'packages'), {recursive: true});
  });

  afterEach(async () => {
    // Clean up test workspace
    try {
      await rm(testWorkspaceDir, {recursive: true, force: true});
    } catch {
      // Ignore cleanup errors
    }
  });

  const runCLI = async (command: string, input?: string): Promise<{stdout: string; stderr: string}> => {
    const cliPath = join(__dirname, '../../dist/cli.js');
    const fullCommand = input ? `echo "${input}" | node "${cliPath}" ${command}` : `node "${cliPath}" ${command}`;

    try {
      const {stdout, stderr} = await execAsync(fullCommand, {
        cwd: testWorkspaceDir,
        timeout: 30000,
      });
      return {stdout, stderr};
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : String(error);
      throw new Error(`CLI execution failed: ${errorMessage}`);
    }
  };

  it.skip('should show help when no command is provided', async () => {
    // Skipped: Requires full Thunder workspace setup
    const {stdout} = await runCLI('--help');

    expect(stdout).toContain('⚡ Thunder Create');
    expect(stdout).toContain('Commands:');
    expect(stdout).toContain('feature');
    expect(stdout).toContain('package');
  });

  it.skip('should show feature command help', async () => {
    // Skipped: Requires full Thunder workspace setup
    const {stdout} = await runCLI('feature --help');

    expect(stdout).toContain('Create a new feature module');
  });

  // Note: Interactive CLI tests are complex due to stdin handling
  // For now, we'll focus on template validation and structure testing

  it('should validate feature name input', () => {
    // Test invalid names - names that should fail validation
    const invalidNames = [
      {name: '', reason: 'empty'},
      {name: '123invalid', reason: 'starts with number'},
      {name: 'invalid-name-', reason: 'ends with hyphen'},
      {name: '-invalid', reason: 'starts with hyphen'},
      {name: 'invalid name', reason: 'contains space'},
      {name: 'invalid@name', reason: 'contains @'},
    ];

    invalidNames.forEach((testCase) => {
      expect(
        () => validateName(testCase.name, 'Feature'),
        `${testCase.name} (${testCase.reason}) should throw`,
      ).toThrow();
    });
  });

  it('should accept valid feature names', () => {
    const validNames = ['mytest', 'testFeature', 'user-management', 'oauth2-integration'];

    validNames.forEach((name) => {
      expect(() => validateName(name, 'Feature')).not.toThrow();
    });
  });

  it('should convert feature names to proper formats', () => {
    const testCases = [
      {input: 'TestFeature', expected: 'test-feature'},
      {input: 'userManagement', expected: 'user-management'},
      {input: 'OAuth2Integration', expected: 'o-auth-2-integration'},
    ];

    testCases.forEach((testCase) => {
      expect(kebabCase(testCase.input)).toBe(testCase.expected);
    });
  });
});
