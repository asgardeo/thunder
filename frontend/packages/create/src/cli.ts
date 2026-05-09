#!/usr/bin/env node

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

/* eslint-disable @thunderid/copyright-header */

import {intro, outro, cancel} from '@clack/prompts';
import {createLogger} from '@thunderid/logger';
import {Command} from 'commander';
import colors from 'picocolors';
import createFeature from './commands/createFeature';
import createPackage from './commands/createPackage';
import getWorkspaceInfo from './utils/getWorkspaceInfo';

const logger = createLogger();

const program: Command = new Command();

async function main(): Promise<void> {
  // eslint-disable-next-line no-console
  console.clear();

  intro(
    '\n' +
      [
        colors.blueBright('████████╗██╗  ██╗██╗   ██╗███╗   ██╗██████╗ ███████╗██████╗ ') +
          colors.magentaBright('██╗██████╗'),
        colors.blueBright('╚══██╔══╝██║  ██║██║   ██║████╗  ██║██╔══██╗██╔════╝██╔══██╗') +
          colors.magentaBright('██║██╔══██╗'),
        colors.blueBright('   ██║   ███████║██║   ██║██╔██╗ ██║██║  ██║█████╗  ██████╔╝') +
          colors.magentaBright('██║██║  ██║'),
        colors.cyanBright('   ██║   ██╔══██║██║   ██║██║╚██╗██║██║  ██║██╔══╝  ██╔══██╗') +
          colors.magentaBright('██║██║  ██║'),
        colors.cyanBright('   ██║   ██║  ██║╚██████╔╝██║ ╚████║██████╔╝███████╗██║  ██║') +
          colors.magentaBright('██║██████╔╝'),
        colors.cyanBright('   ╚═╝   ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝╚═════╝ ╚══════╝╚═╝  ╚═╝') +
          colors.magentaBright('╚═╝╚═════╝'),
      ].join('\n') +
      '\n\n' +
      `          ${colors.yellow('⚡')} ${colors.bold(colors.white('ThunderID'))}${colors.dim(
        colors.gray(' · Frontend Scaffolding Tool'),
      )}\n`,
  );

  // Check if we're in a project workspace

  const workspaceInfo = getWorkspaceInfo();
  if (!workspaceInfo.isThunderWorkspace) {
    cancel(colors.red('This command must be run from a project workspace.'));
    process.exit(1);
  }

  // Validate that required workspace paths exist
  if (!workspaceInfo.packagePath) {
    cancel(colors.red('Missing required folder: frontend/packages. Please create it before running this command.'));
    process.exit(1);
  }
  if (!workspaceInfo.appsPath) {
    cancel(colors.red('Missing required folder: frontend/apps. Please create it before running this command.'));
    process.exit(1);
  }

  program.name('create').description('CLI scaffolding tool for ⚡ ThunderID frontends').version('0.0.0');

  program.command('feature').description('Create a new feature module').action(createFeature);

  program.command('package').description('Create a new shared package').action(createPackage);

  await program.parseAsync();

  outro(colors.green('✅ Done! Happy coding!'));
}

export default main;

if (import.meta.url === `file://${process.argv[1]}`) {
  main().catch((error: unknown) => {
    logger.error('CLI execution failed:', {error});
    process.exit(1);
  });
}
