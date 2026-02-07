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
 * Validation schema for screen background
 */
const screenBackgroundSchema = z.object({
  type: z.enum(['gradient', 'solid', 'image'], {
    required_error: 'Background type is required',
  }),
  value: z.string().min(1, 'Background value is required'),
});

/**
 * Validation schema for screen container
 */
const screenContainerSchema = z.object({
  maxWidth: z
    .number()
    .min(300, 'Maximum width must be at least 300px')
    .max(1200, 'Maximum width cannot exceed 1200px'),
  padding: z.number().min(0, 'Padding cannot be negative').max(100, 'Padding cannot exceed 100px'),
  borderRadius: z
    .number()
    .min(0, 'Border radius cannot be negative')
    .max(50, 'Border radius cannot exceed 50px'),
  elevation: z.number().min(0, 'Elevation cannot be negative').max(24, 'Elevation cannot exceed 24'),
  background: z.string().regex(/^#[0-9A-F]{6}$/i, 'Background must be a valid hex color'),
});

/**
 * Validation schema for screen spacing
 */
const screenSpacingSchema = z.object({
  componentGap: z
    .number()
    .min(0, 'Component gap cannot be negative')
    .max(100, 'Component gap cannot exceed 100px'),
  sectionGap: z.number().min(0, 'Section gap cannot be negative').max(100, 'Section gap cannot exceed 100px'),
});

/**
 * Validation schema for screen header
 */
const screenHeaderSchema = z.object({
  enabled: z.boolean(),
  height: z.number().min(0, 'Height cannot be negative').max(200, 'Height cannot exceed 200px'),
  padding: z.number().min(0, 'Padding cannot be negative').max(50, 'Padding cannot exceed 50px'),
  showLogo: z.boolean().optional(),
  showLanguageSelector: z.boolean().optional(),
  showBackButton: z.boolean().optional(),
});

/**
 * Validation schema for screen footer
 */
const screenFooterSchema = z.object({
  enabled: z.boolean(),
  height: z.number().min(0, 'Height cannot be negative').max(200, 'Height cannot exceed 200px'),
  padding: z.number().min(0, 'Padding cannot be negative').max(50, 'Padding cannot exceed 50px'),
  showLinks: z.boolean().optional(),
});

/**
 * Validation schema for screen configuration
 */
const screenConfigSchema = z.object({
  template: z.string().min(1, 'Template is required'),
  background: screenBackgroundSchema.optional(),
  container: screenContainerSchema.optional(),
  spacing: screenSpacingSchema.optional(),
  header: screenHeaderSchema.optional(),
  footer: screenFooterSchema.optional(),
  extends: z.string().optional(),
});

/**
 * Validation schema for layout configuration
 */
const layoutConfigSchema = z.object({
  screens: z.object({
    auth: screenConfigSchema.optional(),
    signin: screenConfigSchema.optional(),
    signup: screenConfigSchema.optional(),
    mfa: screenConfigSchema.optional(),
    'forgot-password': screenConfigSchema.optional(),
    recovery: screenConfigSchema.optional(),
    'password-reset': screenConfigSchema.optional(),
  }),
});

/**
 * Main validation schema for layout form
 */
export const layoutFormSchema = z.object({
  displayName: z
    .string()
    .min(1, 'Display name is required')
    .max(100, 'Display name cannot exceed 100 characters'),
  layout: layoutConfigSchema,
});

/**
 * Type inference from the schema
 */
export type LayoutFormData = z.infer<typeof layoutFormSchema>;

/**
 * Default form values
 */
export const defaultLayoutFormData: LayoutFormData = {
  displayName: '',
  layout: {
    screens: {
      signin: {
        template: 'centered',
        background: {
          type: 'solid',
          value: '#FFFFFF',
        },
        container: {
          maxWidth: 450,
          padding: 32,
          borderRadius: 8,
          elevation: 2,
          background: '#FFFFFF',
        },
        spacing: {
          componentGap: 16,
          sectionGap: 24,
        },
        header: {
          enabled: true,
          height: 64,
          padding: 16,
          showLogo: true,
          showLanguageSelector: true,
          showBackButton: false,
        },
        footer: {
          enabled: true,
          height: 48,
          padding: 16,
          showLinks: true,
        },
      },
    },
  },
};
