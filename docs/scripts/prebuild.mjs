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

import { execSync } from 'child_process';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';
import { createLogger } from '@thunder/logger';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const logger = createLogger('prebuild');

/**
 * Execute a command and handle errors
 */
function executeScript(scriptName, scriptPath) {
    logger.info(`\n🔄 Running ${scriptName}...`);
    try {
        execSync(`node ${scriptPath}`, {
            stdio: 'inherit',
            cwd: join(__dirname, '..'),
            env: process.env
        });
        logger.info(`✅ ${scriptName} completed successfully\n`);
    } catch (error) {
        logger.error(`❌ ${scriptName} failed: ${error.message}`);
        process.exit(1);
    }
}

/**
 * Main function to generate all documentation artifacts
 */
async function generateDocs() {
    logger.info('⚡️ Thunder Documentation Generator\n');
    logger.info('Generating documentation artifacts...\n');

    // Generate OpenAPI specs
    executeScript('API Specs Generator', join(__dirname, 'merge-openapi-specs.mjs'));

    // Generate changelog
    executeScript('Changelog Generator', join(__dirname, 'generate-changelog.mjs'));

    logger.info('🎉 All documentation artifacts generated successfully!\n');
}

generateDocs();
