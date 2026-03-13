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

import type {Theme} from '@thunder/shared-design';
import {describe, it, expect} from 'vitest';
import buildThemeFromPrimaryColor from '../buildThemeFromPrimaryColor';

function makeBaseTheme(overrides: Partial<Record<string, unknown>> = {}): Theme {
  return {
    defaultColorScheme: 'light',
    colorSchemes: {
      light: {
        palette: {
          primary: {main: '#ffffff', contrastText: '#000000', light: '#ffffff', dark: '#cccccc'},
          secondary: {main: '#9333ea', contrastText: '#ffffff'},
          background: {default: '#f5f5f5', paper: '#ffffff'},
        },
      },
      dark: {
        palette: {
          primary: {main: '#ffffff', contrastText: '#000000', light: '#ffffff', dark: '#cccccc'},
          secondary: {main: '#c084fc', contrastText: '#000000'},
          background: {default: '#121212', paper: '#1e1e1e'},
        },
      },
    },
    shape: {borderRadius: 8},
    typography: {fontFamily: 'Inter'},
    ...overrides,
  } as unknown as Theme;
}

describe('buildThemeFromPrimaryColor', () => {
  describe('immutability', () => {
    it('does not mutate the base theme', () => {
      const base = makeBaseTheme();
      const originalJson = JSON.stringify(base);
      buildThemeFromPrimaryColor(base, '#4f46e5');
      expect(JSON.stringify(base)).toBe(originalJson);
    });

    it('returns a new object (deep clone)', () => {
      const base = makeBaseTheme();
      const result = buildThemeFromPrimaryColor(base, '#4f46e5');
      expect(result).not.toBe(base);
    });

    it('does not share palette references with the base theme', () => {
      const base = makeBaseTheme();
      const result = buildThemeFromPrimaryColor(base, '#4f46e5');
      expect(result.colorSchemes?.light?.palette).not.toBe(base.colorSchemes?.light?.palette);
    });
  });

  describe('primary palette replacement', () => {
    it('sets light.palette.primary.main to the provided hex', () => {
      const hex = '#4f46e5';
      const result = buildThemeFromPrimaryColor(makeBaseTheme(), hex);
      expect(result.colorSchemes?.light?.palette?.primary?.main).toBe(hex);
    });

    it('sets dark.palette.primary.main to the provided hex', () => {
      const hex = '#4f46e5';
      const result = buildThemeFromPrimaryColor(makeBaseTheme(), hex);
      expect(result.colorSchemes?.dark?.palette?.primary?.main).toBe(hex);
    });

    it('replaces primaries in both schemes with the same value', () => {
      const hex = '#dc2626';
      const result = buildThemeFromPrimaryColor(makeBaseTheme(), hex);
      expect(result.colorSchemes?.light?.palette?.primary?.main).toBe(hex);
      expect(result.colorSchemes?.dark?.palette?.primary?.main).toBe(hex);
    });

    it('generates a light variant (rgb string) in the primary palette', () => {
      const result = buildThemeFromPrimaryColor(makeBaseTheme(), '#4f46e5');
      expect(result.colorSchemes?.light?.palette?.primary?.light).toMatch(/^rgb\(\d+, \d+, \d+\)$/);
    });

    it('generates a dark variant (rgb string) in the primary palette', () => {
      const result = buildThemeFromPrimaryColor(makeBaseTheme(), '#4f46e5');
      expect(result.colorSchemes?.light?.palette?.primary?.dark).toMatch(/^rgb\(\d+, \d+, \d+\)$/);
    });

    it('generates contrastText as #ffffff for a dark primary color', () => {
      // #4f46e5 is a dark color → white contrast text
      const result = buildThemeFromPrimaryColor(makeBaseTheme(), '#4f46e5');
      expect(result.colorSchemes?.light?.palette?.primary?.contrastText).toBe('#ffffff');
    });

    it('generates contrastText as #000000 for a light primary color', () => {
      // #ffdd00 is a bright color → black contrast text
      const result = buildThemeFromPrimaryColor(makeBaseTheme(), '#ffdd00');
      expect(result.colorSchemes?.light?.palette?.primary?.contrastText).toBe('#000000');
    });

    it('generates mainChannel as space-separated RGB integers', () => {
      // #4f46e5 → r=79, g=70, b=229
      const result = buildThemeFromPrimaryColor(makeBaseTheme(), '#4f46e5');
      expect(result.colorSchemes?.light?.palette?.primary?.mainChannel).toBe('79 70 229');
    });

    it('generates lightChannel as space-separated RGB integers', () => {
      const result = buildThemeFromPrimaryColor(makeBaseTheme(), '#4f46e5');
      const channel = result.colorSchemes?.light?.palette?.primary?.lightChannel;
      expect(channel).toMatch(/^\d+ \d+ \d+$/);
    });

    it('generates darkChannel as space-separated RGB integers', () => {
      const result = buildThemeFromPrimaryColor(makeBaseTheme(), '#4f46e5');
      const channel = result.colorSchemes?.light?.palette?.primary?.darkChannel;
      expect(channel).toMatch(/^\d+ \d+ \d+$/);
    });
  });

  describe('non-primary fields are preserved', () => {
    it('preserves secondary palette in light scheme', () => {
      const base = makeBaseTheme();
      const result = buildThemeFromPrimaryColor(base, '#4f46e5');
      expect(result.colorSchemes?.light?.palette?.secondary).toEqual(base.colorSchemes?.light?.palette?.secondary);
    });

    it('preserves secondary palette in dark scheme', () => {
      const base = makeBaseTheme();
      const result = buildThemeFromPrimaryColor(base, '#4f46e5');
      expect(result.colorSchemes?.dark?.palette?.secondary).toEqual(base.colorSchemes?.dark?.palette?.secondary);
    });

    it('preserves background palette', () => {
      const base = makeBaseTheme();
      const result = buildThemeFromPrimaryColor(base, '#4f46e5');
      expect(result.colorSchemes?.light?.palette?.background).toEqual(base.colorSchemes?.light?.palette?.background);
    });

    it('preserves shape.borderRadius', () => {
      const base = makeBaseTheme({shape: {borderRadius: 16}});
      const result = buildThemeFromPrimaryColor(base, '#22c55e');
      expect(result.shape?.borderRadius).toBe(16);
    });

    it('preserves typography.fontFamily', () => {
      const base = makeBaseTheme({typography: {fontFamily: 'Georgia'}});
      const result = buildThemeFromPrimaryColor(base, '#22c55e');
      expect(result.typography?.fontFamily).toBe('Georgia');
    });

    it('preserves defaultColorScheme', () => {
      const base = makeBaseTheme({defaultColorScheme: 'dark'});
      const result = buildThemeFromPrimaryColor(base, '#22c55e');
      expect(result.defaultColorScheme).toBe('dark');
    });
  });

  describe('different hex values', () => {
    it.each([
      ['#000000', '0 0 0'],
      ['#ff0000', '255 0 0'],
      ['#00ff00', '0 255 0'],
      ['#0000ff', '0 0 255'],
      ['#ffffff', '255 255 255'],
    ])('parses %s correctly into mainChannel %s', (hex, expectedChannel) => {
      const result = buildThemeFromPrimaryColor(makeBaseTheme(), hex);
      expect(result.colorSchemes?.light?.palette?.primary?.mainChannel).toBe(expectedChannel);
    });
  });

  describe('light/dark channel values are bounded 0–255', () => {
    it('light channel values do not exceed 255', () => {
      // Very bright color – lightening shouldn't overflow
      const result = buildThemeFromPrimaryColor(makeBaseTheme(), '#f0f0f0');
      const channel = result.colorSchemes.light!.palette.primary.lightChannel.split(' ').map(Number);
      expect(channel.every((v) => v <= 255)).toBe(true);
    });

    it('dark channel values are not negative', () => {
      // Very dark color – darkening shouldn't underflow
      const result = buildThemeFromPrimaryColor(makeBaseTheme(), '#010101');
      const channel = result.colorSchemes.light!.palette.primary.darkChannel.split(' ').map(Number);
      expect(channel.every((v) => v >= 0)).toBe(true);
    });
  });
});
