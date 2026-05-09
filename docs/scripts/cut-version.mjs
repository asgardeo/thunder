#!/usr/bin/env node

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

/* eslint-disable @thunderid/copyright-header, import/no-extraneous-dependencies, no-underscore-dangle */

/**
 * Cuts a new Docusaurus documentation version and snapshots the API spec for it.
 *
 * Usage:
 *   node scripts/cut-version.mjs <version>
 *
 * Example:
 *   node scripts/cut-version.mjs 1.1.0
 *
 * This script:
 *   1. Generates a versioned combined OpenAPI spec at static/api/<version>/combined.yaml
 *   2. Runs `docusaurus docs:version <version>` to snapshot the content/ and sidebars
 *
 * After running, commit all newly created files and update docusaurus.config.ts
 * to add the new version to the `versions` map if you need a custom label or path.
 */

import {execFileSync} from 'child_process';
import {existsSync} from 'fs';
import {join, dirname} from 'path';
import {createLogger} from '@thunderid/logger';
import {fileURLToPath} from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const logger = createLogger('cut-version');

const version = process.argv[2];

if (!version) {
  logger.error('Usage: node scripts/cut-version.mjs <version>');
  logger.error('Example: node scripts/cut-version.mjs 1.1.0');
  process.exit(1);
}

if (!/^\d+\.\d+\.\d+/.test(version)) {
  logger.error(`Invalid version format: "${version}". Expected semver (e.g. 1.1.0).`);
  process.exit(1);
}

const mergeScript = join(__dirname, 'merge-openapi-specs.mjs');

if (!existsSync(mergeScript)) {
  logger.error(`merge-openapi-specs.mjs not found at ${mergeScript}`);
  process.exit(1);
}

try {
  // Step 1: Generate the versioned combined API spec.
  logger.info(`📦 Generating combined API spec for version ${version}...`);
  execFileSync('node', [mergeScript, '--version-path', version], {stdio: 'inherit'});

  // Step 2: Cut the Docusaurus doc version (snapshots content/ and versioned_sidebars/).
  logger.info(`📸 Cutting Docusaurus doc version ${version}...`);
  execFileSync('pnpm', ['docusaurus', 'docs:version', version], {stdio: 'inherit'});

  logger.info(`✅ Version ${version} cut successfully.`);
  logger.info(`   → API spec: static/api/${version}/combined.yaml`);
  logger.info(`   → Docs snapshot: versioned_docs/version-${version}/`);
  logger.info('');
  logger.info('Next steps:');
  logger.info(
    `  1. git add docs/static/api/${version}/ docs/versioned_docs/version-${version}/ docs/versioned_sidebars/ docs/versions.json`,
  );
  logger.info('  2. Update docs/docusaurus.config.ts → add the new version to the versions map if needed');
  logger.info('  3. git commit -m "chore: cut docs version ' + version + '"');
} catch (error) {
  logger.error('❌ Failed to cut version:', error.message);
  process.exit(1);
}
