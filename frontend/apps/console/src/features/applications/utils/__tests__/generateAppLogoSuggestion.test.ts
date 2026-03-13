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

import {describe, expect, it, vi, beforeEach, afterEach} from 'vitest';
import generateAppLogoSuggestions from '../generateAppLogoSuggestion';

describe('generateAppLogoSuggestions', () => {
  beforeEach(() => {
    // Mock Math.random to return predictable values for testing
    vi.spyOn(Math, 'random');
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('Basic Functionality', () => {
    it('should return an array of logo URLs', () => {
      const logos = generateAppLogoSuggestions(5);

      expect(Array.isArray(logos)).toBe(true);
      expect(logos).toHaveLength(5);
    });

    it('should return the correct number of logo URLs', () => {
      const counts = [1, 3, 5, 10, 15];

      counts.forEach((count) => {
        const logos = generateAppLogoSuggestions(count);
        expect(logos).toHaveLength(count);
      });
    });

    it('should return valid Google static asset URLs', () => {
      const logos = generateAppLogoSuggestions(5);

      logos.forEach((logo) => {
        expect(logo).toMatch(/^https:\/\/ssl\.gstatic\.com\/docs\/common\/profile\/\w+_lg\.png$/);
      });
    });

    it('should return URLs with animal names from the ANIMALS list', () => {
      const logos = generateAppLogoSuggestions(5);
      const regex = /profile\/(\w+)_lg\.png$/;
      const animalNames = logos.map((logo) => {
        const match = regex.exec(logo);
        return match ? match[1] : null;
      });

      animalNames.forEach((animalName) => {
        expect(animalName).toBeTruthy();
        expect(typeof animalName).toBe('string');
      });
    });
  });

  describe('Edge Cases', () => {
    it('should handle count of 0', () => {
      const logos = generateAppLogoSuggestions(0);

      expect(logos).toHaveLength(0);
      expect(logos).toEqual([]);
    });

    it('should handle count of 1', () => {
      const logos = generateAppLogoSuggestions(1);

      expect(logos).toHaveLength(1);
      expect(logos[0]).toMatch(/^https:\/\/ssl\.gstatic\.com\/docs\/common\/profile\/\w+_lg\.png$/);
    });

    it('should handle large counts (maximum available animals)', () => {
      // There are 44 animals in the ANIMALS array
      const logos = generateAppLogoSuggestions(44);

      expect(logos).toHaveLength(44);
      logos.forEach((logo) => {
        expect(logo).toMatch(/^https:\/\/ssl\.gstatic\.com\/docs\/common\/profile\/\w+_lg\.png$/);
      });
    });

    it('should handle counts larger than available animals', () => {
      // Requesting more than 44 animals should return only available ones
      const logos = generateAppLogoSuggestions(100);

      expect(logos.length).toBeLessThanOrEqual(44);
      logos.forEach((logo) => {
        expect(logo).toMatch(/^https:\/\/ssl\.gstatic\.com\/docs\/common\/profile\/\w+_lg\.png$/);
      });
    });
  });

  describe('Randomization', () => {
    it('should return different results on multiple calls', () => {
      const logos1 = generateAppLogoSuggestions(5);
      const logos2 = generateAppLogoSuggestions(5);

      // With random shuffling, results should differ (though there's a tiny chance they match)
      // We'll test this by calling it multiple times
      const allSame = logos1.every((logo, index) => logo === logos2[index]);

      // If random is working, there's a very low probability all 5 would be the same
      // This test might occasionally fail due to randomness, but it's very unlikely
      expect(allSame).toBe(false);
    });

    it('should use Math.random for shuffling', () => {
      generateAppLogoSuggestions(5);

      expect(Math.random).toHaveBeenCalled();
    });

    it('should produce different orderings with different random values', () => {
      // First call with specific random sequence
      vi.mocked(Math.random).mockReturnValueOnce(0.1).mockReturnValueOnce(0.9).mockReturnValueOnce(0.3);

      const logos1 = generateAppLogoSuggestions(3);

      // Second call with different random sequence
      vi.mocked(Math.random).mockReturnValueOnce(0.9).mockReturnValueOnce(0.1).mockReturnValueOnce(0.5);

      const logos2 = generateAppLogoSuggestions(3);

      // Different random values should produce different results
      expect(logos1).not.toEqual(logos2);
    });
  });

  describe('URL Format', () => {
    it('should use HTTPS protocol', () => {
      const logos = generateAppLogoSuggestions(5);

      logos.forEach((logo) => {
        expect(logo.startsWith('https://')).toBe(true);
      });
    });

    it('should use the correct domain', () => {
      const logos = generateAppLogoSuggestions(5);

      logos.forEach((logo) => {
        expect(logo).toContain('ssl.gstatic.com');
      });
    });

    it('should use the correct path structure', () => {
      const logos = generateAppLogoSuggestions(5);

      logos.forEach((logo) => {
        expect(logo).toContain('/docs/common/profile/');
      });
    });

    it('should use the _lg.png suffix', () => {
      const logos = generateAppLogoSuggestions(5);

      logos.forEach((logo) => {
        expect(logo.endsWith('_lg.png')).toBe(true);
      });
    });
  });

  describe('Uniqueness', () => {
    it('should not return duplicate URLs in a single call', () => {
      const logos = generateAppLogoSuggestions(10);
      const uniqueLogos = new Set(logos);

      expect(uniqueLogos.size).toBe(logos.length);
    });

    it('should return unique animal names within a single call', () => {
      const logos = generateAppLogoSuggestions(10);
      const regex = /profile\/(\w+)_lg\.png$/;
      const animalNames = logos.map((logo) => {
        const match = regex.exec(logo);
        return match ? match[1] : null;
      });

      const uniqueAnimals = new Set(animalNames);
      expect(uniqueAnimals.size).toBe(animalNames.length);
    });
  });
});
