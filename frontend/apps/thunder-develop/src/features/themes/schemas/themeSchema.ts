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

import {z} from 'zod';

/**
 * Regex pattern for validating hex color codes.
 * Matches: #000000, #FFFFFF, #aabbcc, etc.
 */
const hexColorRegex = /^#[0-9A-F]{6}$/i;

/**
 * Regex pattern for validating border radius format.
 * Matches: 8px, 12px, 0px, etc.
 */
const borderRadiusRegex = /^\d+px$/;

/**
 * Schema for validating hex color values.
 */
const hexColorSchema = z
  .string()
  .regex(hexColorRegex, {message: 'Invalid hex color format. Use format: #RRGGBB'})
  .toUpperCase();

/**
 * Schema for brand color configuration (primary/secondary).
 */
const brandColorSchema = z.object({
  main: hexColorSchema,
  dark: hexColorSchema,
  contrastText: hexColorSchema,
});

/**
 * Schema for background color configuration.
 */
const backgroundColorSchema = z.object({
  default: hexColorSchema,
  paper: hexColorSchema,
});

/**
 * Schema for text color configuration.
 */
const textColorSchema = z.object({
  primary: hexColorSchema,
  secondary: hexColorSchema,
});

/**
 * Schema for all color properties.
 */
const colorsSchema = z.object({
  primary: brandColorSchema,
  secondary: brandColorSchema,
  background: backgroundColorSchema,
  text: textColorSchema,
});

/**
 * Schema for a single color scheme (light or dark).
 */
const colorSchemeSchema = z.object({
  colors: colorsSchema,
});

/**
 * Schema for typography configuration.
 */
const typographySchema = z.object({
  fontFamily: z
    .string()
    .min(1, {message: 'Font family is required'})
    .max(200, {message: 'Font family must be 200 characters or less'}),
});

/**
 * Schema for shape configuration.
 */
const shapeSchema = z.object({
  borderRadius: z.string().regex(borderRadiusRegex, {message: 'Border radius must be in format: 8px'}),
});

/**
 * Complete theme configuration schema.
 */
const themeConfigSchema = z.object({
  direction: z.enum(['ltr', 'rtl'], {
    errorMap: () => ({message: 'Direction must be either "ltr" or "rtl"'}),
  }),
  defaultColorScheme: z.enum(['light', 'dark'], {
    errorMap: () => ({message: 'Default color scheme must be either "light" or "dark"'}),
  }),
  colorSchemes: z.object({
    light: colorSchemeSchema,
    dark: colorSchemeSchema,
  }),
  shape: shapeSchema,
  typography: typographySchema,
});

/**
 * Form data schema for theme builder.
 * This is used for react-hook-form validation.
 */
export const themeFormSchema = z.object({
  displayName: z
    .string()
    .min(1, {message: 'Display name is required'})
    .max(100, {message: 'Display name must be 100 characters or less'}),
  theme: themeConfigSchema,
});

/**
 * Type inference for the theme form data.
 * Use this type with react-hook-form.
 */
export type ThemeFormData = z.infer<typeof themeFormSchema>;

/**
 * Default theme configuration for new themes.
 * Provides sensible defaults for all required fields.
 */
export const defaultThemeFormData: ThemeFormData = {
  displayName: '',
  theme: {
    direction: 'ltr',
    defaultColorScheme: 'light',
    colorSchemes: {
      light: {
        colors: {
          primary: {
            main: '#FF7300',
            dark: '#E66800',
            contrastText: '#FFFFFF',
          },
          secondary: {
            main: '#4A4A4A',
            dark: '#3A3A3A',
            contrastText: '#FFFFFF',
          },
          background: {
            default: '#FFFFFF',
            paper: '#F5F5F5',
          },
          text: {
            primary: '#000000',
            secondary: '#666666',
          },
        },
      },
      dark: {
        colors: {
          primary: {
            main: '#FF7300',
            dark: '#E66800',
            contrastText: '#FFFFFF',
          },
          secondary: {
            main: '#AAAAAA',
            dark: '#999999',
            contrastText: '#000000',
          },
          background: {
            default: '#1A1A1A',
            paper: '#2A2A2A',
          },
          text: {
            primary: '#FFFFFF',
            secondary: '#AAAAAA',
          },
        },
      },
    },
    shape: {
      borderRadius: '8px',
    },
    typography: {
      fontFamily: 'Inter, sans-serif',
    },
  },
};
