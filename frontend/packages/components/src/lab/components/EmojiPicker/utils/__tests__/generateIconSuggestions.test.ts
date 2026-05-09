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

import {describe, expect, it, vi, beforeEach, afterEach} from 'vitest';
import generateIconSuggestions, {EMOJI_CATEGORIES} from '../generateIconSuggestions';

describe('generateIconSuggestions', () => {
  beforeEach(() => {
    vi.spyOn(Math, 'random');
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('Basic Functionality', () => {
    it('should return an array of emoji strings', () => {
      const icons = generateIconSuggestions(5);

      expect(Array.isArray(icons)).toBe(true);
      expect(icons).toHaveLength(5);
    });

    it('should return the correct number of icons', () => {
      const counts = [1, 3, 5, 8, 12];

      counts.forEach((count) => {
        const icons = generateIconSuggestions(count);
        expect(icons).toHaveLength(count);
      });
    });

    it('should return non-empty strings', () => {
      const icons = generateIconSuggestions(5);

      icons.forEach((icon) => {
        expect(typeof icon).toBe('string');
        expect(icon.length).toBeGreaterThan(0);
      });
    });

    it('should not return URL strings', () => {
      const icons = generateIconSuggestions(10);

      icons.forEach((icon) => {
        expect(icon.startsWith('http')).toBe(false);
      });
    });
  });

  describe('Edge Cases', () => {
    it('should handle count of 0', () => {
      const icons = generateIconSuggestions(0);

      expect(icons).toHaveLength(0);
      expect(icons).toEqual([]);
    });

    it('should handle count of 1', () => {
      const icons = generateIconSuggestions(1);

      expect(icons).toHaveLength(1);
      expect(typeof icons[0]).toBe('string');
    });

    it('should handle counts larger than available icons', () => {
      const totalEmojis = EMOJI_CATEGORIES.flatMap((c) => c.emojis).length;
      const icons = generateIconSuggestions(totalEmojis + 100);

      expect(icons.length).toBeLessThanOrEqual(totalEmojis);
    });
  });

  describe('Randomization', () => {
    it('should use Math.random for shuffling', () => {
      generateIconSuggestions(5);

      expect(Math.random).toHaveBeenCalled();
    });

    it('should produce different orderings with different random values', () => {
      vi.mocked(Math.random).mockReturnValueOnce(0.1).mockReturnValueOnce(0.9).mockReturnValueOnce(0.3);
      const icons1 = generateIconSuggestions(3);

      vi.mocked(Math.random).mockReturnValueOnce(0.9).mockReturnValueOnce(0.1).mockReturnValueOnce(0.5);
      const icons2 = generateIconSuggestions(3);

      expect(icons1).not.toEqual(icons2);
    });
  });

  describe('Uniqueness', () => {
    it('should not return duplicate icons in a single call', () => {
      const icons = generateIconSuggestions(15);
      const uniqueIcons = new Set(icons);

      expect(uniqueIcons.size).toBe(icons.length);
    });
  });

  describe('EMOJI_CATEGORIES', () => {
    it('should have at least one category', () => {
      expect(EMOJI_CATEGORIES.length).toBeGreaterThan(0);
    });

    it('each category should have a label and non-empty emojis', () => {
      EMOJI_CATEGORIES.forEach((category) => {
        expect(typeof category.label).toBe('string');
        expect(category.label.length).toBeGreaterThan(0);
        expect(category.emojis.length).toBeGreaterThan(0);
      });
    });

    it('each emoji entry should have char and keywords', () => {
      EMOJI_CATEGORIES.forEach((category) => {
        category.emojis.forEach((emoji) => {
          expect(typeof emoji.char).toBe('string');
          expect(emoji.char.length).toBeGreaterThan(0);
          expect(typeof emoji.keywords).toBe('string');
          expect(emoji.keywords.length).toBeGreaterThan(0);
        });
      });
    });
  });
});
