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

import type {Base} from './base';

/**
 * Interface for a component.
 */
export interface Element<T = unknown> extends Base<T> {
  /**
   * Nested set of elements.
   * @remarks Some elements like `Form` can have nested elements.
   */
  components?: Element[];
  action?: {
    next?: string;
    [key: string]: unknown;
  };
}

export const ElementCategories = {
  Action: 'ACTION',
  Block: 'BLOCK',
  Display: 'DISPLAY',
  Field: 'FIELD',
} as const;

export const ElementTypes = {
  Input: 'INPUT',
  Button: 'BUTTON',
  Captcha: 'CAPTCHA',
  Divider: 'DIVIDER',
  Choice: 'CHOICE',
  Image: 'IMAGE',
  RichText: 'RICH_TEXT',
  Typography: 'TYPOGRAPHY',
  Resend: 'RESEND',
} as const;

export const BlockTypes = {
  Form: 'FORM',
} as const;

export const InputVariants = {
  Text: 'TEXT',
  Password: 'PASSWORD',
  Email: 'EMAIL',
  Telephone: 'TELEPHONE',
  Number: 'NUMBER',
  Checkbox: 'CHECKBOX',
  OTP: 'OTP',
} as const;

export const ButtonVariants = {
  Primary: 'PRIMARY',
  Secondary: 'SECONDARY',
  Social: 'SOCIAL',
  Text: 'TEXT',
} as const;

export const ButtonTypes = {
  Submit: 'submit',
  Button: 'button',
} as const;

export const TypographyVariants = {
  H1: 'H1',
  H2: 'H2',
  H3: 'H3',
  H4: 'H4',
  H5: 'H5',
  H6: 'H6',
  Body1: 'BODY1',
  Body2: 'BODY2',
} as const;

export const DividerVariants = {
  Horizontal: 'HORIZONTAL',
  Vertical: 'VERTICAL',
} as const;

/**
 * Event types for ACTION components.
 * Defines the interaction semantics for buttons and actions.
 */
export const ActionEventTypes = {
  Trigger: 'TRIGGER',
  Submit: 'SUBMIT',
  Navigate: 'NAVIGATE',
  Cancel: 'CANCEL',
  Reset: 'RESET',
  Back: 'BACK',
} as const;

export type ActionEventTypes = (typeof ActionEventTypes)[keyof typeof ActionEventTypes];
