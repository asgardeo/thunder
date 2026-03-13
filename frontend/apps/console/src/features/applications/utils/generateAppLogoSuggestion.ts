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

/**
 * A comprehensive list of animal names used for generating logo suggestions.
 *
 * These animal names correspond to available profile images from Google's static assets.
 */
const ANIMALS: string[] = [
  'alligator',
  'anteater',
  'armadillo',
  'axolotl',
  'badger',
  'bat',
  'beaver',
  'buffalo',
  'camel',
  'chinchilla',
  'chupacabra',
  'coyote',
  'crow',
  'dingo',
  'dolphin',
  'dragon',
  'duck',
  'elephant',
  'ferret',
  'fox',
  'giraffe',
  'gopher',
  'iguana',
  'kangaroo',
  'kiwi',
  'koala',
  'lemur',
  'leopard',
  'llama',
  'manatee',
  'narwhal',
  'otter',
  'panda',
  'penguin',
  'platypus',
  'quokka',
  'raccoon',
  'sheep',
  'squirrel',
  'tiger',
  'turtle',
  'walrus',
  'wolf',
  'wombat',
];

/**
 * Generates a specified number of random animal logo URLs.
 *
 * This function randomly selects animal names from the available list and returns
 * their corresponding Google static asset URLs that can be used as application logos.
 *
 * @param count - The number of random logo URLs to generate.
 * @returns An array of logo URLs pointing to Google's static animal profile images.
 *
 * @example
 * ```typescript
 * const logos = getRandomLogos(5);
 * // Returns: [
 * //   'https://ssl.gstatic.com/docs/common/profile/panda_lg.png',
 * //   'https://ssl.gstatic.com/docs/common/profile/fox_lg.png',
 * //   ...
 * // ]
 * ```
 */
export default function generateAppLogoSuggestions(count: number): string[] {
  const shuffled: string[] = [...ANIMALS].sort((): number => Math.random() - 0.5);

  return shuffled
    .slice(0, count)
    .map((animal: string): string => `https://ssl.gstatic.com/docs/common/profile/${animal}_lg.png`);
}
