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

/* eslint-disable @thunderid/copyright-header */

/**
 * Fetches the Unicode emoji test file and regenerates
 * src/components/EmojiPicker/emojis.json.
 *
 * Usage:
 *   node scripts/update-emojis.mjs
 *
 * The generated JSON is an array of EmojiCategory objects:
 *   [ { label: string, emojis: Array<{ char: string, keywords: string }> } ]
 *
 * Only `fully-qualified` emojis are included. The "Component" group
 * (skin-tone modifiers) is skipped entirely.
 */

import {writeFileSync} from 'node:fs';
import {dirname, resolve} from 'node:path';
import {fileURLToPath} from 'node:url';
import {createLogger} from '@thunderid/logger';

const logger = createLogger({level: 'info'});

const EMOJI_TEST_URL = 'https://unicode.org/Public/emoji/latest/emoji-test.txt';

/** Groups to skip — skin-tone components add noise without value in a picker. */
const SKIP_GROUPS = new Set(['Component']);

/**
 * Maximum Emoji version to include.
 * Emoji 15.0 ships with macOS 13, iOS 16.4, Android 13, Windows 11 22H2.
 * Anything newer shows "NO GLYPH" on older systems.
 */
const MAX_EMOJI_VERSION = 15.0;

const __dirname = dirname(fileURLToPath(import.meta.url));
const OUTPUT_PATH = resolve(__dirname, '../src/lab/EmojiPicker/emojis.json');

/**
 * @param {string} text
 * @returns {Array<{ label: string, emojis: Array<{ char: string, keywords: string }> }>}
 */
function parse(text) {
  const {categories} = text
    .split('\n')
    .map((raw) => raw.trim())
    .reduce(
      ({categories: cats, currentGroup}, line) => {
        // Track current group
        if (line.startsWith('# group:')) {
          const label = line.slice('# group:'.length).trim();
          if (SKIP_GROUPS.has(label)) {
            return {categories: cats, currentGroup: null};
          }
          const newGroup = {label, emojis: []};
          cats.push(newGroup);
          return {categories: cats, currentGroup: newGroup};
        }

        // Skip comments, empty lines, lines in a skipped group, or non-fully-qualified entries
        if (!currentGroup || !line || line.startsWith('#') || !line.includes('; fully-qualified')) {
          return {categories: cats, currentGroup};
        }

        // Line format example:
        //   1F600  ; fully-qualified  # 😀 E1.0 grinning face
        const hashIndex = line.indexOf('#');
        if (hashIndex === -1) return {categories: cats, currentGroup};

        const afterHash = line.slice(hashIndex + 1).trim();

        // afterHash: "😀 E1.0 grinning face"
        // The emoji char is the first grapheme cluster (may be multiple code points)
        const spaceAfterEmoji = afterHash.indexOf(' ');
        if (spaceAfterEmoji === -1) return {categories: cats, currentGroup};

        const char = afterHash.slice(0, spaceAfterEmoji).trim();
        const rest = afterHash.slice(spaceAfterEmoji + 1).trim();

        // rest: "E1.0 grinning face"
        // Parse and gate the version token (E<number>.<number>)
        const versionMatch = rest.match(/^E(\d+\.\d+)\s*/);
        if (!versionMatch || parseFloat(versionMatch[1]) > MAX_EMOJI_VERSION) {
          return {categories: cats, currentGroup};
        }

        const keywords = rest.slice(versionMatch[0].length).trim();
        if (char && keywords) {
          currentGroup.emojis.push({char, keywords});
        }

        return {categories: cats, currentGroup};
      },
      {categories: [], currentGroup: null},
    );

  return categories;
}

async function main() {
  logger.info(`Fetching ${EMOJI_TEST_URL} …`);
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 15000);
  const response = await fetch(EMOJI_TEST_URL, {signal: controller.signal});

  clearTimeout(timeout);

  if (!response.ok) {
    throw new Error(`Failed to fetch emoji data: ${response.status} ${response.statusText}`);
  }

  const text = await response.text();
  logger.info(`Parsing …`);
  const categories = parse(text);

  const totalEmojis = categories.reduce((sum, c) => sum + c.emojis.length, 0);
  logger.info(`Found ${categories.length} categories, ${totalEmojis} fully-qualified emojis.`);

  const json = JSON.stringify(categories, null, 2);
  writeFileSync(OUTPUT_PATH, json, 'utf8');
  logger.info(`Written to ${OUTPUT_PATH}`);
}

main().catch((err) => {
  logger.error(err);
  // eslint-disable-next-line no-undef
  process.exit(1);
});
