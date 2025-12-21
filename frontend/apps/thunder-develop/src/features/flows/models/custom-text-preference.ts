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

/**
 * Constants for preview screen types.
 */
export const PreviewScreenType = {
  COMMON: 'common',
  LOGIN: 'login',
  MY_ACCOUNT: 'myaccount',
  EMAIL_LINK_EXPIRY: 'email-link-expiry',
  EMAIL_TEMPLATE: 'email-template',
  SIGN_UP: 'sign-up',
  EMAIL_OTP: 'email-otp',
  PUSH_AUTH: 'push-auth',
  SMS_OTP: 'sms-otp',
  TOTP: 'totp',
  PASSWORD_RECOVERY: 'password-recovery',
  PASSWORD_RESET: 'password-reset',
  PASSWORD_RESET_SUCCESS: 'password-reset-success',
  USERNAME_RECOVERY_CLAIM: 'username-recovery-claim',
  USERNAME_RECOVERY_CHANNEL_SELECTION: 'username-recovery-channel-selection',
  USERNAME_RECOVERY_SUCCESS: 'username-recovery-success',
} as const;

export type PreviewScreenType = (typeof PreviewScreenType)[keyof typeof PreviewScreenType];

/**
 * Interface for the base custom text hook result.
 */
export interface CustomTextPreferenceResult {
  /**
   * The data returned by the hook.
   */
  data?: Partial<Record<PreviewScreenType, Record<string, string>>>;
  /**
   * Error state.
   */
  error?: unknown;
  /**
   * Loading state.
   */
  isLoading: boolean;
  /**
   * Mutate function to revalidate data.
   */
  mutate: () => void;
}

/**
 * Interface for the screen meta hook result.
 */
export interface CustomTextPreferenceScreenMetaResult {
  /**
   * The data returned by the hook.
   */
  data?: Partial<Record<PreviewScreenType, CustomTextPreferenceScreenMetaInterface>>;
  /**
   * Error state.
   */
  error?: unknown;
  /**
   * Loading state.
   */
  isLoading: boolean;
  /**
   * Mutate function to revalidate data.
   */
  mutate: () => void;
}

/**
 * Interface for the custom text preference screen meta.
 */
export type CustomTextPreferenceScreenMetaInterface = Record<
  string,
  {
    /**
     * Is the text preference editable.
     */
    EDITABLE: boolean;
    /**
     * Screen name of the text preference.
     */
    SCREEN: string;
    /**
     * Is the text preference multi-line.
     */
    MULTI_LINE: boolean;
  }
>;
