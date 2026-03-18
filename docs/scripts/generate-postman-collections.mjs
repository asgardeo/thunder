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

/**
 * Generates Postman collections from Thunder OpenAPI specifications.
 *
 * Usage:
 *   node scripts/generate-postman-collections.mjs
 *
 * Output:
 *   samples/api/postman/<spec-name>.json  - one collection per OpenAPI spec
 *   samples/api/postman/thunder.json      - combined collection from all specs
 */

import {readFileSync, writeFileSync, readdirSync, existsSync, mkdirSync} from 'fs';
import {join, dirname, basename} from 'path';
import {fileURLToPath} from 'url';
import Converter from 'openapi-to-postmanv2';
import {promisify} from 'util';

const convert = promisify(Converter.convert.bind(Converter));

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const ROOT_DIR = join(__dirname, '..', '..');
const API_DIR = join(ROOT_DIR, 'api');
const OUTPUT_DIR = join(ROOT_DIR, 'samples', 'api', 'postman');

// Spec files to skip (WIP / not yet stable)
const SKIP_FILES = new Set(['design.yaml']);

const CONVERT_OPTIONS = {
    folderStrategy: 'Tags',
    requestNameSource: 'Fallback',
    indentCharacter: '  ',
    collapseFolders: true,
    optimizeConversion: false,
    strictRequestNames: false,
    includeAuthInfoInExample: true,
    exampleParametersResolution: 'Schema',
    enableOptionalParameters: false,
    disabledParametersValidation: false,
    keepImplicitHeaders: false,
};

/**
 * Replace hardcoded OAuth2 URLs with collection variable references.
 */
function replaceOAuth2Urls(obj) {
    if (obj === null || typeof obj !== 'object') return;

    if (Array.isArray(obj)) {
        for (const item of obj) replaceOAuth2Urls(item);
        return;
    }

    if ((obj.key === 'accessTokenUrl' || obj.key === 'authUrl') && typeof obj.value === 'string') {
        obj.value = obj.value.replace(/https?:\/\/localhost:\d+/g, '{{baseUrl}}');
    }

    for (const value of Object.values(obj)) replaceOAuth2Urls(value);
}

/**
 * Convert a single OpenAPI spec file to a Postman collection.
 */
async function convertSpec(specPath) {
    const specContent = readFileSync(specPath, 'utf8');
    const result = await convert({type: 'string', data: specContent}, CONVERT_OPTIONS);

    if (!result.result) {
        throw new Error(`Conversion failed for ${specPath}: ${result.reason}`);
    }

    const collection = result.output[0].data;
    replaceOAuth2Urls(collection);

    return collection;
}

/**
 * Generate individual Postman collections for each OpenAPI spec.
 */
async function generateIndividualCollections(specFiles) {
    const collections = [];

    for (const file of specFiles) {
        const specPath = join(API_DIR, file);
        const name = basename(file, '.yaml');
        const outputPath = join(OUTPUT_DIR, `${name}.json`);

        console.log(`  Converting ${file}...`);
        const collection = await convertSpec(specPath);

        writeFileSync(outputPath, JSON.stringify(collection, null, 2), 'utf8');
        console.log(`  Written to ${outputPath}`);

        collections.push({name, collection});
    }

    return collections;
}

/**
 * Generate a single combined Postman collection from all specs.
 *
 * Merges all items (folders/requests) from each individual collection
 * into one top-level collection named "Thunder API".
 */
function generateCombinedCollection(collections) {
    const combined = {
        info: {
            name: 'Thunder API',
            description: 'Complete API collection for Thunder identity and access management.',
            schema: 'https://schema.getpostman.com/json/collection/v2.1.0/collection.json',
        },
        item: [],
        variable: [],
    };

    const first = collections[0]?.collection;

    if (first?.variable) {
        combined.variable = first.variable;
    }

    const authSource = collections.find(({collection}) => collection.auth)?.collection;

    if (authSource?.auth) {
        combined.auth = authSource.auth;
    }

    for (const {collection} of collections) {
        if (collection.item) {
            combined.item.push(...collection.item);
        }
    }

    return combined;
}

async function main() {
    console.log('Generating Postman collections from OpenAPI specs...');

    const specFiles = readdirSync(API_DIR)
        .filter((file) => file.endsWith('.yaml') && !SKIP_FILES.has(file))
        .sort();

    if (specFiles.length === 0) {
        throw new Error('No OpenAPI spec files found in the api/ directory.');
    }

    console.log(`Found ${specFiles.length} spec file(s)`);

    if (!existsSync(OUTPUT_DIR)) {
        mkdirSync(OUTPUT_DIR, {recursive: true});
    }

    const collections = await generateIndividualCollections(specFiles);

    console.log('Generating combined collection...');
    const combined = generateCombinedCollection(collections);
    const combinedPath = join(OUTPUT_DIR, 'thunder.json');

    writeFileSync(combinedPath, JSON.stringify(combined, null, 2), 'utf8');
    console.log(`Combined collection written to ${combinedPath}`);

    console.log(`Done. ${collections.length} individual collection(s) + 1 combined collection generated.`);
}

main().catch((error) => {
    console.error('Error generating Postman collections:', error);
    process.exit(1);
});
