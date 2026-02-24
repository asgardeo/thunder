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

function hexToRgb(hex: string): [number, number, number] {
  const clean = hex.replace('#', '');
  return [parseInt(clean.slice(0, 2), 16), parseInt(clean.slice(2, 4), 16), parseInt(clean.slice(4, 6), 16)];
}

function lighten(r: number, g: number, b: number, amount: number): [number, number, number] {
  return [
    Math.min(255, Math.round(r + (255 - r) * amount)),
    Math.min(255, Math.round(g + (255 - g) * amount)),
    Math.min(255, Math.round(b + (255 - b) * amount)),
  ];
}

function darken(r: number, g: number, b: number, amount: number): [number, number, number] {
  return [Math.round(r * (1 - amount)), Math.round(g * (1 - amount)), Math.round(b * (1 - amount))];
}

function toLinear(c: number): number {
  const s = c / 255;
  return s <= 0.03928 ? s / 12.92 : Math.pow((s + 0.055) / 1.055, 2.4);
}

function contrastText(r: number, g: number, b: number): string {
  const lum = 0.2126 * toLinear(r) + 0.7152 * toLinear(g) + 0.0722 * toLinear(b);
  return lum > 0.179 ? '#000000' : '#ffffff';
}

function toChannel(r: number, g: number, b: number): string {
  return `${r} ${g} ${b}`;
}

function buildPrimaryPalette(hex: string): Record<string, string> {
  const [r, g, b] = hexToRgb(hex);
  const [lr, lg, lb] = lighten(r, g, b, 0.3);
  const [dr, dg, db] = darken(r, g, b, 0.3);
  const ct = contrastText(r, g, b);
  const [ctr, ctg, ctb] = hexToRgb(ct === '#000000' ? '#000000' : '#ffffff');

  return {
    main: hex,
    light: `rgb(${lr}, ${lg}, ${lb})`,
    dark: `rgb(${dr}, ${dg}, ${db})`,
    contrastText: ct,
    mainChannel: toChannel(r, g, b),
    lightChannel: toChannel(lr, lg, lb),
    darkChannel: toChannel(dr, dg, db),
    contrastTextChannel: toChannel(ctr, ctg, ctb),
  };
}

/**
 * Deep-clones `baseTheme` and replaces the primary palette in both light and
 * dark colorSchemes with values derived from `primaryHex`.
 * Everything else (typography, shape, secondary, background, etc.) is
 * preserved unchanged from the base theme.
 */
export default function buildThemeFromPrimaryColor(baseTheme: Theme, primaryHex: string): Theme {
  const clone = JSON.parse(JSON.stringify(baseTheme)) as Theme;
  const primary = buildPrimaryPalette(primaryHex) as unknown as typeof clone.colorSchemes.light.palette.primary;

  if (clone.colorSchemes?.light?.palette) {
    clone.colorSchemes.light.palette.primary = primary;
  }
  if (clone.colorSchemes?.dark?.palette) {
    clone.colorSchemes.dark.palette.primary = primary;
  }

  return clone;
}
