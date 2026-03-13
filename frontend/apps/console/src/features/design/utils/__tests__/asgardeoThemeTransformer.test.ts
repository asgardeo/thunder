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
import toAsgardeoTheme from '../asgardeoThemeTransformer';

const lightPalette = {
  primary: {main: '#4f46e5', contrastText: '#ffffff', light: '#7c73ea', dark: '#3730a3'},
  secondary: {main: '#9333ea', contrastText: '#ffffff', dark: '#7e22ce'},
  background: {default: '#f5f5f5', paper: '#ffffff'},
  text: {primary: '#000000de', secondary: '#00000099'},
};

const darkPalette = {
  primary: {main: '#818cf8', contrastText: '#000000', light: '#a5b4fc', dark: '#4f46e5'},
  secondary: {main: '#c084fc', contrastText: '#000000', dark: '#a855f7'},
  background: {default: '#121212', paper: '#1e1e1e'},
  text: {primary: '#ffffffde', secondary: '#ffffff99'},
};

function makeTheme(overrides: Partial<Record<string, unknown>> = {}): Theme {
  return {
    defaultColorScheme: 'light',
    colorSchemes: {
      light: {palette: lightPalette},
      dark: {palette: darkPalette},
    },
    shape: {borderRadius: 8},
    typography: {fontFamily: 'Inter, sans-serif'},
    direction: 'ltr',
    ...overrides,
  } as unknown as Theme;
}

describe('toAsgardeoTheme', () => {
  describe('activeScheme resolution', () => {
    it('uses light scheme when defaultColorScheme is "light" and no override given', () => {
      const theme = makeTheme({defaultColorScheme: 'light'});
      const result = toAsgardeoTheme(theme);
      expect(result.colors?.primary?.main).toBe(lightPalette.primary.main);
    });

    it('uses dark scheme when defaultColorScheme is "dark" and no override given', () => {
      const theme = makeTheme({defaultColorScheme: 'dark'});
      const result = toAsgardeoTheme(theme);
      expect(result.colors?.primary?.main).toBe(darkPalette.primary.main);
    });

    it('defaults to light scheme when defaultColorScheme is "system"', () => {
      const theme = makeTheme({defaultColorScheme: 'system'});
      const result = toAsgardeoTheme(theme);
      expect(result.colors?.primary?.main).toBe(lightPalette.primary.main);
    });

    it('uses explicit "light" override regardless of defaultColorScheme', () => {
      const theme = makeTheme({defaultColorScheme: 'dark'});
      const result = toAsgardeoTheme(theme, 'light');
      expect(result.colors?.primary?.main).toBe(lightPalette.primary.main);
    });

    it('uses explicit "dark" override regardless of defaultColorScheme', () => {
      const theme = makeTheme({defaultColorScheme: 'light'});
      const result = toAsgardeoTheme(theme, 'dark');
      expect(result.colors?.primary?.main).toBe(darkPalette.primary.main);
    });
  });

  describe('colors mapping', () => {
    it('maps primary.main', () => {
      const result = toAsgardeoTheme(makeTheme(), 'light');
      expect(result.colors?.primary?.main).toBe(lightPalette.primary.main);
    });

    it('maps primary.contrastText', () => {
      const result = toAsgardeoTheme(makeTheme(), 'light');
      expect(result.colors?.primary?.contrastText).toBe(lightPalette.primary.contrastText);
    });

    it('maps primary.dark when present', () => {
      const result = toAsgardeoTheme(makeTheme(), 'light');
      expect(result.colors?.primary?.dark).toBe(lightPalette.primary.dark);
    });

    it('omits primary.dark when not present in source palette', () => {
      const theme = makeTheme();
      const {dark: omit, ...primaryWithoutDark} = lightPalette.primary;
      (
        theme as unknown as {colorSchemes: {light: {palette: {primary: typeof primaryWithoutDark}}}}
      ).colorSchemes.light.palette.primary = primaryWithoutDark;
      const result = toAsgardeoTheme(theme, 'light');
      expect(result.colors?.primary?.dark).toBeUndefined();
    });

    it('maps secondary.main', () => {
      const result = toAsgardeoTheme(makeTheme(), 'light');
      expect(result.colors?.secondary?.main).toBe(lightPalette.secondary.main);
    });

    it('maps secondary.contrastText', () => {
      const result = toAsgardeoTheme(makeTheme(), 'light');
      expect(result.colors?.secondary?.contrastText).toBe(lightPalette.secondary.contrastText);
    });

    it('maps secondary.dark when present', () => {
      const result = toAsgardeoTheme(makeTheme(), 'light');
      expect(result.colors?.secondary?.dark).toBe(lightPalette.secondary.dark);
    });

    it('maps background.body.main to palette.background.default', () => {
      const result = toAsgardeoTheme(makeTheme(), 'light');
      expect(result.colors?.background?.body?.main).toBe(lightPalette.background.default);
    });

    it('maps background.surface to palette.background.paper', () => {
      const result = toAsgardeoTheme(makeTheme(), 'light');
      expect(result.colors?.background?.surface).toBe(lightPalette.background.paper);
    });

    it('maps text.primary', () => {
      const result = toAsgardeoTheme(makeTheme(), 'light');
      expect(result.colors?.text?.primary).toBe(lightPalette.text.primary);
    });

    it('maps text.secondary', () => {
      const result = toAsgardeoTheme(makeTheme(), 'light');
      expect(result.colors?.text?.secondary).toBe(lightPalette.text.secondary);
    });

    it('omits colors entirely when palette is undefined', () => {
      const theme = makeTheme();
      (theme as unknown as {colorSchemes: object}).colorSchemes = {};
      const result = toAsgardeoTheme(theme, 'light');
      expect(result.colors).toBeUndefined();
    });
  });

  describe('borderRadius mapping', () => {
    it('converts a numeric borderRadius to a px string', () => {
      const theme = makeTheme({shape: {borderRadius: 12}});
      const result = toAsgardeoTheme(theme);
      expect(result.borderRadius?.small).toBe('12px');
      expect(result.borderRadius?.medium).toBe('12px');
      expect(result.borderRadius?.large).toBe('12px');
    });

    it('passes through a string borderRadius unchanged', () => {
      const theme = makeTheme({shape: {borderRadius: '1rem'}});
      const result = toAsgardeoTheme(theme);
      expect(result.borderRadius?.small).toBe('1rem');
      expect(result.borderRadius?.medium).toBe('1rem');
      expect(result.borderRadius?.large).toBe('1rem');
    });

    it('converts borderRadius 0 to "0px"', () => {
      const theme = makeTheme({shape: {borderRadius: 0}});
      const result = toAsgardeoTheme(theme);
      expect(result.borderRadius?.small).toBe('0px');
    });

    it('omits borderRadius when shape is undefined', () => {
      const theme = makeTheme({shape: undefined});
      const result = toAsgardeoTheme(theme);
      expect(result.borderRadius).toBeUndefined();
    });

    it('omits borderRadius when shape.borderRadius is undefined', () => {
      const theme = makeTheme({shape: {}});
      const result = toAsgardeoTheme(theme);
      expect(result.borderRadius).toBeUndefined();
    });
  });

  describe('typography mapping', () => {
    it('maps fontFamily when present', () => {
      const theme = makeTheme({typography: {fontFamily: 'Georgia, serif'}});
      const result = toAsgardeoTheme(theme);
      expect(result.typography?.fontFamily).toBe('Georgia, serif');
    });

    it('omits typography when fontFamily is absent', () => {
      const theme = makeTheme({typography: {}});
      const result = toAsgardeoTheme(theme);
      expect(result.typography).toBeUndefined();
    });

    it('omits typography when typography is undefined', () => {
      const theme = makeTheme({typography: undefined});
      const result = toAsgardeoTheme(theme);
      expect(result.typography).toBeUndefined();
    });
  });

  describe('direction mapping', () => {
    it('maps direction "ltr"', () => {
      const theme = makeTheme({direction: 'ltr'});
      const result = toAsgardeoTheme(theme);
      expect(result.direction).toBe('ltr');
    });

    it('maps direction "rtl"', () => {
      const theme = makeTheme({direction: 'rtl'});
      const result = toAsgardeoTheme(theme);
      expect(result.direction).toBe('rtl');
    });

    it('omits direction when it is neither "ltr" nor "rtl"', () => {
      const theme = makeTheme({direction: 'auto'});
      const result = toAsgardeoTheme(theme);
      expect(result.direction).toBeUndefined();
    });

    it('omits direction when it is undefined', () => {
      const theme = makeTheme({direction: undefined});
      const result = toAsgardeoTheme(theme);
      expect(result.direction).toBeUndefined();
    });
  });

  describe('empty / minimal themes', () => {
    it('returns an empty object for a completely empty theme', () => {
      const result = toAsgardeoTheme({} as Theme);
      expect(result).toEqual({});
    });

    it('returns an empty object when colorSchemes is missing', () => {
      const result = toAsgardeoTheme({defaultColorScheme: 'light'} as Theme);
      expect(result).toEqual({});
    });

    it('does not mutate the original theme object', () => {
      const theme = makeTheme();
      const original = JSON.parse(JSON.stringify(theme)) as Theme;
      toAsgardeoTheme(theme, 'dark');
      expect(theme).toEqual(original);
    });
  });
});
