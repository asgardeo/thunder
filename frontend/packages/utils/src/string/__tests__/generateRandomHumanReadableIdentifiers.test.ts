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

import humanId from 'human-id';
import {describe, expect, it, vi, beforeEach, afterEach} from 'vitest';
import generateRandomHumanReadableIdentifiers from '../generateRandomHumanReadableIdentifiers';

vi.mock('human-id', () => ({
  default: vi.fn(() => 'blue falcon'),
}));

describe('generateRandomHumanReadableIdentifiers', () => {
  beforeEach(() => {
    vi.mocked(humanId).mockImplementation(() => 'blue falcon');
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('Basic Functionality', () => {
    it('should return an array of 5 suggestions by default', () => {
      const suggestions = generateRandomHumanReadableIdentifiers();

      expect(Array.isArray(suggestions)).toBe(true);
      expect(suggestions).toHaveLength(5);
    });

    it('should return an array of strings', () => {
      const suggestions = generateRandomHumanReadableIdentifiers();

      suggestions.forEach((suggestion) => {
        expect(typeof suggestion).toBe('string');
      });
    });

    it('should return non-empty strings', () => {
      const suggestions = generateRandomHumanReadableIdentifiers();

      suggestions.forEach((suggestion) => {
        expect(suggestion.length).toBeGreaterThan(0);
      });
    });
  });

  describe('humanId Integration', () => {
    it('should call humanId with correct configuration', () => {
      generateRandomHumanReadableIdentifiers();

      expect(humanId).toHaveBeenCalledWith({
        separator: ' ',
        capitalize: true,
        adjectiveCount: 1,
        addAdverb: false,
      });
    });

    it('should call humanId 5 times by default', () => {
      generateRandomHumanReadableIdentifiers();

      expect(humanId).toHaveBeenCalledTimes(5);
    });

    it('should call humanId with space separator', () => {
      generateRandomHumanReadableIdentifiers();

      const callArgs = vi.mocked(humanId).mock.calls[0][0] as Record<string, unknown>;
      expect(callArgs['separator']).toBe(' ');
    });

    it('should call humanId with capitalize enabled', () => {
      generateRandomHumanReadableIdentifiers();

      const callArgs = vi.mocked(humanId).mock.calls[0][0] as Record<string, unknown>;
      expect(callArgs['capitalize']).toBe(true);
    });

    it('should call humanId with one adjective', () => {
      generateRandomHumanReadableIdentifiers();

      const callArgs = vi.mocked(humanId).mock.calls[0][0] as Record<string, unknown>;
      expect(callArgs['adjectiveCount']).toBe(1);
    });

    it('should call humanId with no adverbs', () => {
      generateRandomHumanReadableIdentifiers();

      const callArgs = vi.mocked(humanId).mock.calls[0][0] as Record<string, unknown>;
      expect(callArgs['addAdverb']).toBe(false);
    });
  });

  describe('Capitalization', () => {
    it('should capitalize first letter of each word', () => {
      vi.mocked(humanId).mockReturnValue('blue falcon');

      const suggestions = generateRandomHumanReadableIdentifiers();

      suggestions.forEach((suggestion) => {
        expect(suggestion).toBe('Blue Falcon');
      });
    });

    it('should handle single word names', () => {
      vi.mocked(humanId).mockReturnValue('phoenix');

      const suggestions = generateRandomHumanReadableIdentifiers();

      suggestions.forEach((suggestion) => {
        expect(suggestion).toBe('Phoenix');
      });
    });

    it('should handle multiple word names', () => {
      vi.mocked(humanId).mockReturnValue('red dragon master');

      const suggestions = generateRandomHumanReadableIdentifiers();

      suggestions.forEach((suggestion) => {
        expect(suggestion).toBe('Red Dragon Master');
      });
    });

    it('should properly capitalize mixed case input', () => {
      vi.mocked(humanId).mockReturnValue('BLUE falcon');

      const suggestions = generateRandomHumanReadableIdentifiers();

      suggestions.forEach((suggestion) => {
        expect(suggestion).toBe('Blue Falcon');
      });
    });

    it('should handle lowercase input', () => {
      vi.mocked(humanId).mockReturnValue('green turtle');

      const suggestions = generateRandomHumanReadableIdentifiers();

      suggestions.forEach((suggestion) => {
        const words = suggestion.split(' ');
        words.forEach((word) => {
          expect(word[0]).toBe(word[0].toUpperCase());
          expect(word.slice(1)).toBe(word.slice(1).toLowerCase());
        });
      });
    });

    it('should handle uppercase input', () => {
      vi.mocked(humanId).mockReturnValue('PURPLE ELEPHANT');

      const suggestions = generateRandomHumanReadableIdentifiers();

      suggestions.forEach((suggestion) => {
        expect(suggestion).toBe('Purple Elephant');
      });
    });
  });

  describe('Randomness and Uniqueness', () => {
    it('should generate different names when humanId returns different values', () => {
      vi.mocked(humanId)
        .mockReturnValueOnce('blue falcon')
        .mockReturnValueOnce('red dragon')
        .mockReturnValueOnce('green phoenix')
        .mockReturnValueOnce('yellow tiger')
        .mockReturnValueOnce('purple wolf');

      const suggestions = generateRandomHumanReadableIdentifiers();

      expect(suggestions).toEqual(['Blue Falcon', 'Red Dragon', 'Green Phoenix', 'Yellow Tiger', 'Purple Wolf']);
    });

    it('should handle duplicate names from humanId', () => {
      vi.mocked(humanId).mockReturnValue('blue falcon');

      const suggestions = generateRandomHumanReadableIdentifiers();

      expect(suggestions).toHaveLength(5);
      suggestions.forEach((suggestion) => {
        expect(suggestion).toBe('Blue Falcon');
      });
    });
  });

  describe('Edge Cases', () => {
    it('should handle empty string from humanId', () => {
      vi.mocked(humanId).mockReturnValue('');

      const suggestions = generateRandomHumanReadableIdentifiers();

      expect(suggestions).toHaveLength(5);
      suggestions.forEach((suggestion) => {
        expect(suggestion).toBe('');
      });
    });

    it('should handle names with extra spaces', () => {
      vi.mocked(humanId).mockReturnValue('blue  falcon');

      const suggestions = generateRandomHumanReadableIdentifiers();

      suggestions.forEach((suggestion) => {
        expect(suggestion).toContain('Blue');
        expect(suggestion).toContain('Falcon');
      });
    });

    it('should handle names with leading spaces', () => {
      vi.mocked(humanId).mockReturnValue(' blue falcon');

      const suggestions = generateRandomHumanReadableIdentifiers();

      suggestions.forEach((suggestion) => {
        expect(suggestion).toContain('Blue');
        expect(suggestion).toContain('Falcon');
      });
    });

    it('should handle names with trailing spaces', () => {
      vi.mocked(humanId).mockReturnValue('blue falcon ');

      const suggestions = generateRandomHumanReadableIdentifiers();

      suggestions.forEach((suggestion) => {
        expect(suggestion).toContain('Blue');
        expect(suggestion).toContain('Falcon');
      });
    });

    it('should handle special characters in names', () => {
      vi.mocked(humanId).mockReturnValue('blue-falcon');

      const suggestions = generateRandomHumanReadableIdentifiers();

      suggestions.forEach((suggestion) => {
        expect(suggestion).toBe('Blue-falcon');
      });
    });
  });

  describe('Return Value Format', () => {
    it('should use space as word separator', () => {
      vi.mocked(humanId).mockReturnValue('blue falcon');

      const suggestions = generateRandomHumanReadableIdentifiers();

      suggestions.forEach((suggestion) => {
        expect(suggestion).toContain(' ');
      });
    });

    it('should not have leading or trailing spaces after processing', () => {
      vi.mocked(humanId).mockReturnValue('blue falcon');

      const suggestions = generateRandomHumanReadableIdentifiers();

      suggestions.forEach((suggestion) => {
        expect(suggestion).toBe(suggestion.trim());
      });
    });

    it('should maintain proper title case format', () => {
      vi.mocked(humanId).mockReturnValue('blue falcon master');

      const suggestions = generateRandomHumanReadableIdentifiers();

      suggestions.forEach((suggestion) => {
        const words = suggestion.split(' ').filter((word) => word.length > 0);
        words.forEach((word) => {
          expect(word[0]).toBe(word[0].toUpperCase());
          expect(word.slice(1)).toBe(word.slice(1).toLowerCase());
        });
      });
    });
  });

  describe('Integration', () => {
    it('should work with the real human-id module', async () => {
      const actual = await vi.importActual<typeof import('human-id')>('human-id');
      vi.mocked(humanId).mockImplementation(actual.default);

      const suggestions = generateRandomHumanReadableIdentifiers();

      expect(suggestions).toHaveLength(5);
      suggestions.forEach((suggestion) => {
        expect(typeof suggestion).toBe('string');
        expect(suggestion.length).toBeGreaterThan(0);
      });
    });
  });
});
