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

import {existsSync, mkdirSync, writeFileSync, rmSync, realpathSync} from 'fs';
import {tmpdir} from 'os';
import {join} from 'path';
import {describe, it, expect, afterEach, beforeEach} from 'vitest';
import getWorkspaceInfo from '../getWorkspaceInfo';

describe('getWorkspaceInfo', () => {
  const testDir = join(realpathSync(tmpdir()), 'create-test-workspace-info');
  let originalCwd: string;

  beforeEach(() => {
    originalCwd = process.cwd();
  });

  afterEach(() => {
    process.chdir(originalCwd);
    if (existsSync(testDir)) {
      rmSync(testDir, {recursive: true, force: true});
    }
  });

  it('should detect workspace from root directory', () => {
    mkdirSync(join(testDir, 'frontend'), {recursive: true});

    writeFileSync(join(testDir, 'frontend', 'package.json'), JSON.stringify({name: '@thunderid/frontend'}));
    writeFileSync(join(testDir, 'frontend', 'nx.json'), JSON.stringify({}));

    process.chdir(testDir);

    const info = getWorkspaceInfo();

    expect(info.isThunderWorkspace).toBe(true);
    expect(info.frontendPath).toBe(join(testDir, 'frontend'));
    expect(info.currentWorkingDirectory).toBe(testDir);
  });

  it('should detect workspace from frontend directory', () => {
    const frontendDir = join(testDir, 'frontend');
    mkdirSync(frontendDir, {recursive: true});

    writeFileSync(join(frontendDir, 'package.json'), JSON.stringify({name: '@thunderid/frontend'}));
    writeFileSync(join(frontendDir, 'nx.json'), JSON.stringify({}));

    process.chdir(frontendDir);

    const info = getWorkspaceInfo();

    expect(info.isThunderWorkspace).toBe(true);
    expect(info.frontendPath).toBe(frontendDir);
    expect(info.currentWorkingDirectory).toBe(frontendDir);
  });

  it('should detect workspace from nested directory', () => {
    const frontendDir = join(testDir, 'frontend');
    const packagesDir = join(frontendDir, 'packages');
    const packageDir = join(packagesDir, 'my-package');

    mkdirSync(packageDir, {recursive: true});

    writeFileSync(join(frontendDir, 'package.json'), JSON.stringify({name: '@thunderid/frontend'}));
    writeFileSync(join(frontendDir, 'nx.json'), JSON.stringify({}));

    process.chdir(packageDir);

    const info = getWorkspaceInfo();

    expect(info.isThunderWorkspace).toBe(true);
    expect(info.frontendPath).toBe(frontendDir);
    expect(info.currentWorkingDirectory).toBe(packageDir);
  });

  it('should return non-workspace info when not in a workspace', () => {
    mkdirSync(testDir, {recursive: true});
    process.chdir(testDir);

    const info = getWorkspaceInfo();

    expect(info.isThunderWorkspace).toBe(false);
    expect(info.frontendPath).toBeNull();
    expect(info.packagePath).toBeNull();
    expect(info.appsPath).toBeNull();
    expect(info.currentWorkingDirectory).toBe(testDir);
  });

  it('should detect packages directory when it exists', () => {
    const frontendDir = join(testDir, 'frontend');
    const packagesDir = join(frontendDir, 'packages');

    mkdirSync(packagesDir, {recursive: true});

    writeFileSync(join(frontendDir, 'package.json'), JSON.stringify({name: '@thunderid/frontend'}));
    writeFileSync(join(frontendDir, 'nx.json'), JSON.stringify({}));

    process.chdir(testDir);

    const info = getWorkspaceInfo();

    expect(info.isThunderWorkspace).toBe(true);
    expect(info.packagePath).toBe(packagesDir);
  });

  it('should detect apps directory when it exists', () => {
    const frontendDir = join(testDir, 'frontend');
    const appsDir = join(frontendDir, 'apps');

    mkdirSync(appsDir, {recursive: true});

    writeFileSync(join(frontendDir, 'package.json'), JSON.stringify({name: '@thunderid/frontend'}));
    writeFileSync(join(frontendDir, 'nx.json'), JSON.stringify({}));

    process.chdir(testDir);

    const info = getWorkspaceInfo();

    expect(info.isThunderWorkspace).toBe(true);
    expect(info.appsPath).toBe(appsDir);
  });

  it('should handle missing packages and apps directories', () => {
    const frontendDir = join(testDir, 'frontend');

    mkdirSync(frontendDir, {recursive: true});

    writeFileSync(join(frontendDir, 'package.json'), JSON.stringify({name: '@thunderid/frontend'}));
    writeFileSync(join(frontendDir, 'nx.json'), JSON.stringify({}));

    process.chdir(testDir);

    const info = getWorkspaceInfo();

    expect(info.isThunderWorkspace).toBe(true);
    expect(info.packagePath).toBeNull();
    expect(info.appsPath).toBeNull();
  });

  it('should not detect workspace with incomplete setup', () => {
    const frontendDir = join(testDir, 'frontend');

    mkdirSync(frontendDir, {recursive: true});

    // Missing nx.json
    writeFileSync(join(frontendDir, 'package.json'), JSON.stringify({name: '@thunderid/frontend'}));

    process.chdir(testDir);

    const info = getWorkspaceInfo();

    expect(info.isThunderWorkspace).toBe(false);
  });

  it('should not detect workspace when package.json does not have product name in name', () => {
    const frontendDir = join(testDir, 'frontend');

    mkdirSync(frontendDir, {recursive: true});

    writeFileSync(join(frontendDir, 'package.json'), JSON.stringify({name: 'some-other-project'}));
    writeFileSync(join(frontendDir, 'nx.json'), JSON.stringify({}));

    process.chdir(testDir);

    const info = getWorkspaceInfo();

    expect(info.isThunderWorkspace).toBe(false);
  });

  it('should detect workspace with product-related package name', () => {
    const frontendDir = join(testDir, 'frontend');

    mkdirSync(frontendDir, {recursive: true});

    writeFileSync(join(frontendDir, 'package.json'), JSON.stringify({name: '@company/app'}));
    writeFileSync(join(frontendDir, 'nx.json'), JSON.stringify({}));

    process.chdir(testDir);

    const info = getWorkspaceInfo();

    expect(info.isThunderWorkspace).toBe(true);
  });
});
