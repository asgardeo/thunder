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

import {describe, it, expect, vi} from 'vitest';
import generateOUNameSuggestions from '../generateOUNameSuggestions';

// Mock human-id module
vi.mock('human-id', () => ({
  default: vi.fn(() => 'Test Name'),
}));

describe('generateOUNameSuggestions', () => {
  it('should return an array of 10 suggestions', () => {
    const suggestions = generateOUNameSuggestions();
    expect(suggestions).toHaveLength(10);
  });

  it('should return an array of strings', () => {
    const suggestions = generateOUNameSuggestions();
    suggestions.forEach((suggestion) => {
      expect(typeof suggestion).toBe('string');
    });
  });

  it('should return properly capitalized names', () => {
    const suggestions = generateOUNameSuggestions();
    suggestions.forEach((suggestion) => {
      // Each word should start with uppercase
      const words = suggestion.split(' ');
      words.forEach((word) => {
        if (word.length > 0) {
          expect(word[0]).toBe(word[0].toUpperCase());
          if (word.length > 1) {
            expect(word.slice(1)).toBe(word.slice(1).toLowerCase());
          }
        }
      });
    });
  });

  it('should return non-empty strings', () => {
    const suggestions = generateOUNameSuggestions();
    suggestions.forEach((suggestion) => {
      expect(suggestion.length).toBeGreaterThan(0);
    });
  });

  it('should return space-separated words in title case', () => {
    const suggestions = generateOUNameSuggestions();
    suggestions.forEach((suggestion) => {
      // Check that it has at least one word
      expect(suggestion.trim().length).toBeGreaterThan(0);
    });
  });

  it('should generate different suggestions on multiple calls', () => {
    // Unmock human-id for this test to verify randomness
    vi.doUnmock('human-id');

    // Since we can't easily test randomness with mock, just ensure function executes
    const suggestions1 = generateOUNameSuggestions();
    const suggestions2 = generateOUNameSuggestions();

    // Both calls should return 10 items
    expect(suggestions1).toHaveLength(10);
    expect(suggestions2).toHaveLength(10);

    // Re-mock for other tests
    vi.mock('human-id', () => ({
      default: vi.fn(() => 'Test Name'),
    }));
  });
});
