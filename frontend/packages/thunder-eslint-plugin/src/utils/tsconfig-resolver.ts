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
import {dirname, join, parse, resolve} from 'path';

/**
 * Parser options configuration for ESLint TypeScript integration.
 */
export interface ParserOptions {
  projectService: {
    allowDefaultProject: string[];
  };
  tsconfigRootDir?: string;
}

/**
 * Checks if any of the specified tsconfig files exist in the given directory.
 *
 * @param directory - The directory to check
 * @param configFiles - Array of config filenames to check for
 * @returns True if any config file exists, false otherwise
 */
function hasTsconfigFile(directory: string, configFiles: string[]): boolean {
  return configFiles.some((configFile: string) => {
    const configPath: string = join(directory, configFile);
    return existsSync(configPath);
  });
}

/**
 * Resolves the TypeScript configuration root directory by finding the nearest tsconfig file
 * starting from the current working directory and walking up the directory tree.
 * Prioritizes tsconfig.eslint.json over tsconfig.json for ESLint-specific configurations.
 *
 * @param startDir - The directory to start searching from (defaults to process.cwd())
 * @returns The directory containing a tsconfig file, or undefined if not found
 */
function resolveTsconfigRootDir(startDir: string = process.cwd()): string | undefined {
  let currentDir: string = resolve(startDir);
  const rootDir: string = parse(currentDir).root;

  // List of tsconfig files to check, in order of preference
  const tsconfigFiles: string[] = ['tsconfig.eslint.json', 'tsconfig.json'];

  while (currentDir !== rootDir) {
    if (hasTsconfigFile(currentDir, tsconfigFiles)) {
      return currentDir;
    }

    currentDir = dirname(currentDir);
  }

  // Check root directory as well
  if (hasTsconfigFile(rootDir, tsconfigFiles)) {
    return rootDir;
  }

  return undefined;
}

/**
 * Creates parser options with dynamic tsconfig resolution and appropriate allowDefaultProject patterns.
 *
 * @param additionalPatterns - Additional patterns to include in allowDefaultProject
 * @returns Parser options object for ESLint TypeScript configuration
 */
export default function createParserOptions(additionalPatterns: string[] = []): ParserOptions {
  const defaultPatterns: string[] = [
    'public/*.js',
    '.*.js',
    '.*.cjs',
    '*.js',
    '*.*.js',
    '*.cjs',
    'esbuild.config.js',
    'eslint.config.js',
    'prettier.config.js',
    'webpack.config.js',
    'rollup.config.js',
    'rolldown.config.js',
  ];

  const tsconfigRootDir: string | undefined = resolveTsconfigRootDir();

  return {
    projectService: {
      allowDefaultProject: [...defaultPatterns, ...additionalPatterns],
    },
    ...(tsconfigRootDir && {tsconfigRootDir}),
  };
}
