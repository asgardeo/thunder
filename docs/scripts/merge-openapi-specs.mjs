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

/* eslint-disable @thunderid/copyright-header, import/no-extraneous-dependencies */

import {readFileSync, writeFileSync, readdirSync, existsSync, mkdirSync} from 'fs';
import {join, dirname} from 'path';
import {fileURLToPath} from 'url';
import {createLogger} from '@thunderid/logger';
import {parse, stringify} from 'yaml';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const logger = createLogger('merge-openapi-specs');

const API_DIR = join(__dirname, '..', '..', 'api');
const STATIC_DIR = join(__dirname, '..', 'static', 'api');
const PRODUCT_CONFIG_PATH = join(__dirname, '..', 'docusaurus.product.config.ts');

function readProductConfig(configPath) {
  const content = readFileSync(configPath, 'utf8');
  const nameMatch = content.match(/project\s*:\s*\{[^}]*?name\s*:\s*['"]([^'"]+)['"]/s);
  return nameMatch ? nameMatch[1] : 'Unknown Project';
}

const projectName = readProductConfig(PRODUCT_CONFIG_PATH);

// Resolve version path from --version-path <path> CLI arg, defaulting to 'next'
const versionPathArgIndex = process.argv.indexOf('--version-path');
const versionPath = versionPathArgIndex !== -1 ? process.argv[versionPathArgIndex + 1] : 'next';

const OUTPUT_FILE = join(STATIC_DIR, versionPath, 'combined.yaml');

function mergeOpenAPISpecs() {
  logger.info(`🔄 Merging OpenAPI specifications (version path: ${versionPath})...`);

  // Dynamically read all YAML files from the API directory
  const API_FILES = readdirSync(API_DIR)
    .filter((file) => file.endsWith('.yaml') && file !== 'combined.yaml')
    .sort();

  if (API_FILES.length === 0) {
    throw new Error('No API specification files found in the directory.');
  }

  logger.info(`📁 Found ${API_FILES.length} API specification files`);

  // Base structure from the first spec
  const firstSpec = parse(readFileSync(join(API_DIR, API_FILES[0]), 'utf8'));

  const combined = {
    openapi: firstSpec.openapi || '3.0.3',
    info: {
      title: `${projectName} API Reference`,
      version: '1.0',
      description: `Complete API reference for ${projectName} identity and access management.`,
      license: firstSpec.info?.license || {
        name: 'Apache 2.0',
        url: 'https://www.apache.org/licenses/LICENSE-2.0.html',
      },
    },
    servers: firstSpec.servers || [],
    security: firstSpec.security || [],
    tags: [],
    paths: {},
    components: {
      schemas: {},
      securitySchemes: {},
      responses: {},
      parameters: {},
    },
  };

  // Process each API spec
  API_FILES.forEach((file) => {
    logger.info(`  ➜ Processing ${file}...`);
    const specPath = join(API_DIR, file);
    const spec = parse(readFileSync(specPath, 'utf8'));

    // Merge tags
    if (spec.tags) {
      spec.tags.forEach((tag) => {
        if (!combined.tags.find((t) => t.name === tag.name)) {
          combined.tags.push(tag);
        }
      });
    }

    // Merge paths
    if (spec.paths) {
      Object.entries(spec.paths).forEach(([path, pathItem]) => {
        if (combined.paths[path]) {
          // Merge methods if path exists
          combined.paths[path] = {...combined.paths[path], ...pathItem};
        } else {
          combined.paths[path] = pathItem;
        }
      });
    }

    // Merge components
    if (spec.components) {
      if (spec.components.schemas) {
        combined.components.schemas = {
          ...combined.components.schemas,
          ...spec.components.schemas,
        };
      }
      if (spec.components.securitySchemes) {
        combined.components.securitySchemes = {
          ...combined.components.securitySchemes,
          ...spec.components.securitySchemes,
        };
      }
      if (spec.components.responses) {
        combined.components.responses = {
          ...combined.components.responses,
          ...spec.components.responses,
        };
      }
      if (spec.components.parameters) {
        combined.components.parameters = {
          ...combined.components.parameters,
          ...spec.components.parameters,
        };
      }
    }
  });

  // Sort tags alphabetically
  combined.tags.sort((a, b) => a.name.localeCompare(b.name));

  // Ensure the output directory exists (including any version subdirectory)
  const outputDir = dirname(OUTPUT_FILE);

  if (!existsSync(outputDir)) {
    mkdirSync(outputDir, {recursive: true});
    logger.info(`📁 Created output directory: ${outputDir}`);
  }

  // Write the combined spec
  writeFileSync(OUTPUT_FILE, stringify(combined), 'utf8');
  logger.info(`✅ Combined API spec written to ${OUTPUT_FILE}`);
  logger.info(`📊 Stats: ${combined.tags.length} tags, ${Object.keys(combined.paths).length} paths`);
}

try {
  mergeOpenAPISpecs();
} catch (error) {
  logger.error('❌ Error merging API specs:', error);
  process.exit(1);
}
