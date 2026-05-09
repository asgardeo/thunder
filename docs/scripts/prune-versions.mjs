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
 * Prunes old Docusaurus documentation versions, keeping only the N most recent.
 *
 * Usage:
 *   node scripts/prune-versions.mjs --keep <N>
 *
 * Examples:
 *   node scripts/prune-versions.mjs --keep 1   # only 'next' (no stable versions)
 *   node scripts/prune-versions.mjs --keep 2   # next + 1 most recent stable version
 *   node scripts/prune-versions.mjs --keep 3   # next + 2 most recent stable versions
 *
 * For each pruned version, removes:
 *   - versioned_docs/version-<v>/
 *   - versioned_sidebars/version-<v>-sidebars.json
 *   - static/api/<v>/  (entire directory)
 *   - The version entry from versions.json
 *
 * Note: The 'next' (current) version and its static/api/next/ spec are never pruned.
 *
 * Docusaurus has no native version-delete CLI command — this script handles it directly.
 * The `onlyIncludeVersions` config option can be used in Docusaurus for build-time
 * filtering, but this script does a full filesystem prune for CI deployments.
 */

import {readFileSync, writeFileSync, rmSync, existsSync} from 'fs';
import {join, dirname} from 'path';
import {createLogger} from '@thunderid/logger';
import {fileURLToPath} from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const logger = createLogger('prune-versions');

// Parse --keep <N> from CLI args.
const keepArgIndex = process.argv.indexOf('--keep');
const keepTotal = keepArgIndex !== -1 ? parseInt(process.argv[keepArgIndex + 1], 10) : NaN;

if (isNaN(keepTotal) || keepTotal < 1) {
  logger.error('Usage: node scripts/prune-versions.mjs --keep <N>');
  logger.error('  --keep 1  → only next (no stable versions)');
  logger.error('  --keep 2  → next + 1 most recent stable version');
  process.exit(1);
}

// Maximum number of stable versions to keep (total minus 1 for 'next').
const keepStable = keepTotal - 1;

const docsDir = join(__dirname, '..');
const versionsFile = join(docsDir, 'versions.json');
const versionedDocsDir = join(docsDir, 'versioned_docs');
const versionedSidebarsDir = join(docsDir, 'versioned_sidebars');
const staticApiDir = join(docsDir, 'static', 'api');

if (!existsSync(versionsFile)) {
  logger.info('ℹ️  No versions.json found — nothing to prune.');
  process.exit(0);
}

// versions.json is an array of stable version strings, newest-first (e.g. ["1.2.0", "1.1.0"]).
const versions = JSON.parse(readFileSync(versionsFile, 'utf8'));

if (!Array.isArray(versions)) {
  logger.error('versions.json is not an array. Aborting.');
  process.exit(1);
}

if (versions.length <= keepStable) {
  logger.info(`ℹ️  ${versions.length} stable version(s) present, keeping ${keepStable}. Nothing to prune.`);
  process.exit(0);
}

const toPrune = versions.slice(keepStable); // oldest versions beyond the keep limit
const toKeep = versions.slice(0, keepStable);

logger.info(`📋 Stable versions found: ${versions.join(', ')}`);
logger.info(`✅ Keeping: ${toKeep.length > 0 ? toKeep.join(', ') : '(none)'}`);
logger.info(`🗑️  Pruning: ${toPrune.join(', ')}`);

for (const version of toPrune) {
  const versionedDocsPath = join(versionedDocsDir, `version-${version}`);
  const sidebarPath = join(versionedSidebarsDir, `version-${version}-sidebars.json`);
  const apiSpecDir = join(staticApiDir, version);

  if (existsSync(versionedDocsPath)) {
    rmSync(versionedDocsPath, {recursive: true, force: true});
    logger.info(`  🗑️  Removed ${versionedDocsPath}`);
  }

  if (existsSync(sidebarPath)) {
    rmSync(sidebarPath, {force: true});
    logger.info(`  🗑️  Removed ${sidebarPath}`);
  }

  if (existsSync(apiSpecDir)) {
    rmSync(apiSpecDir, {recursive: true, force: true});
    logger.info(`  🗑️  Removed ${apiSpecDir}`);
  }
}

// Update versions.json to only contain the kept versions.
writeFileSync(versionsFile, JSON.stringify(toKeep, null, 2) + '\n', 'utf8');
logger.info(`✅ Updated versions.json → [${toKeep.join(', ')}]`);
logger.info('✅ Pruning complete.');
