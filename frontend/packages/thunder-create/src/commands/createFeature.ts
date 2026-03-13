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

import {text, select, spinner, cancel} from '@clack/prompts';
import colors from 'picocolors';
import {join} from 'path';
import type {Dirent} from 'fs';
import {createLogger} from '@thunder/logger';
import {readdirSync, existsSync} from 'fs';
import kebabCase from 'lodash-es/kebabCase';
import ensureDir from '../utils/ensureDir';
import createFileFromTemplate from '../utils/createFileFromTemplate';
import getTemplateDir from '../utils/getTemplateDir';
import registerHandlebarsHelpers from '../utils/registerHandlebarsHelpers';
import renderTemplate from '../utils/renderTemplate';
import getWorkspaceInfo from '../utils/getWorkspaceInfo';
import validateName from '../utils/validateName';

const logger = createLogger();

async function createFeature(): Promise<void> {
  const workspaceInfo = getWorkspaceInfo();
  if (!workspaceInfo.packagePath) {
    cancel(colors.red('Missing required folder: frontend/packages. Please create it before running this command.'));
    process.exit(1);
  }

  // Get feature type
  const featureType = await select({
    message: 'Feature type:',
    options: [
      {value: 'admin', label: 'Admin feature (thunder-admin-xxx)'},
      {value: 'gate', label: 'Gate feature (gate-xxx)'},
    ],
  });

  if (typeof featureType !== 'string') {
    cancel(colors.red('Operation cancelled.'));
    process.exit(1);
  }

  // Get feature name
  const name = await text({
    message: 'Feature name:',
    placeholder: 'user-management',
    validate: (value) => {
      try {
        validateName(value, 'Feature');
        return undefined;
      } catch (error) {
        return error instanceof Error ? error.message : 'Invalid feature name';
      }
    },
  });

  if (typeof name !== 'string') {
    cancel(colors.red('Operation cancelled.'));
    process.exit(1);
  }

  const featureName = kebabCase(name);
  const packageName = `thunder-${featureType}-${featureName}`;
  const featureDir = join(workspaceInfo.packagePath, packageName);

  // Check if feature already exists
  if (existsSync(featureDir)) {
    cancel(colors.red(`Feature '${featureName}' already exists at ${featureDir}`));
    process.exit(1);
  }

  const s = spinner();
  s.start(colors.cyan(`Creating feature ${featureName}...`));

  try {
    registerHandlebarsHelpers();

    const context = {
      featureName,
      packageName,
      featureType,
    };

    // Create feature directory structure under src/
    const sourceDirectories = [
      'src/api',
      'src/components',
      'src/config',
      'src/constants',
      'src/contexts',
      'src/data',
      'src/hooks',
      'src/models',
      'src/pages',
      'src/utils',
    ];

    sourceDirectories.forEach((dir) => ensureDir(join(featureDir, dir)));

    // Copy all template files from the template directory
    const copyTemplateFiles = (templateDir: string, targetDir: string, ctx: Record<string, unknown>) => {
      const entries = readdirSync(templateDir, {withFileTypes: true}) as unknown as Dirent[];

      entries.forEach((entry: Dirent) => {
        const templatePath = join(templateDir, entry.name);

        // Process directory names through Handlebars as well
        const processedDirName = renderTemplate(entry.name, ctx as Record<string, string>);
        const targetPath = join(targetDir, processedDirName);

        if (entry.isDirectory()) {
          // Recursively copy directories
          ensureDir(targetPath);
          copyTemplateFiles(templatePath, targetPath, ctx);
        } else if (entry.isFile() && entry.name.endsWith('.hbs')) {
          // Process Handlebars template files
          const outputFileName = entry.name.replace(/\.hbs$/, '');

          // Apply Handlebars processing to the file name as well
          const processedFileName = renderTemplate(outputFileName, ctx as Record<string, string>);
          const processedOutputPath = join(targetDir, processedFileName);

          createFileFromTemplate(templatePath, processedOutputPath, ctx as Record<string, string>);
        }
      });
    };

    // Copy all files from src/ template directory
    const srcTemplateDir = join(getTemplateDir(), 'feature', 'src');
    const srcTargetDir = join(featureDir, 'src');
    copyTemplateFiles(srcTemplateDir, srcTargetDir, context);

    // Create package.json (like thunder-shared-contexts but as feature)
    createFileFromTemplate(
      join(getTemplateDir(), 'feature', 'package.json.hbs'),
      join(featureDir, 'package.json'),
      context,
    );

    // Create other config files
    const configFiles = [
      'tsconfig.json',
      'tsconfig.lib.json',
      'tsconfig.spec.json',
      'tsconfig.eslint.json',
      'eslint.config.js',
      'vitest.config.ts',
      'rolldown.config.js',
      'prettier.config.js',
      '.editorconfig',
      '.gitignore',
      '.prettierignore',
    ];

    configFiles.forEach((file) =>
      createFileFromTemplate(join(getTemplateDir(), 'feature', `${file}.hbs`), join(featureDir, file), context),
    );

    // src/index.ts is already generated by copyTemplateFiles above; do not re-create it here.

    s.stop(colors.green(`✅ Feature '${featureName}' created successfully!`));

    logger.info(`Feature '${featureName}' created at ${featureDir}`);

    // Ask if user wants to install and build
    const shouldInstallAndBuild = await select({
      message: 'Would you like to install dependencies and build the feature now?',
      options: [
        {value: true, label: 'Yes, install and build'},
        {value: false, label: 'No, I will do it later'},
      ],
    });

    if (shouldInstallAndBuild === true) {
      const installBuildSpinner = spinner();
      installBuildSpinner.start(colors.cyan('Installing dependencies and building feature...'));

      try {
        const {execSync} = await import('child_process');

        // Install dependencies and build
        execSync('pnpm install && pnpm build', {
          cwd: featureDir,
          stdio: 'inherit',
        });

        installBuildSpinner.stop(colors.green('✅ Dependencies installed and feature built successfully!'));
      } catch (error) {
        installBuildSpinner.stop(colors.red('❌ Failed to install dependencies or build'));
        logger.error('Install/build failed:', {error: error instanceof Error ? error.message : String(error)});
        // eslint-disable-next-line no-console
        console.log();
        // eslint-disable-next-line no-console
        console.log(colors.yellow(`You can manually run: cd ${featureDir} && pnpm install && pnpm build`));
      }
    } else {
      // eslint-disable-next-line no-console
      console.log();
      // eslint-disable-next-line no-console
      console.log(colors.cyan('Next steps:'));
      // eslint-disable-next-line no-console
      console.log(colors.gray(`1. cd ${featureDir}`));
      // eslint-disable-next-line no-console
      console.log(colors.gray('2. pnpm install'));
      // eslint-disable-next-line no-console
      console.log(colors.gray('3. pnpm build'));
    }
  } catch (error) {
    s.stop(colors.red('❌ Failed to create feature'));
    logger.error('Feature creation failed:', {error: error instanceof Error ? error.message : String(error)});
    process.exit(1);
  }
}

export default createFeature;
