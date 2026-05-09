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
 * Constants related to the flow builder elements.
 *
 * @remarks
 * This class is not meant to be instantiated. It only provides static constants.
 *
 * @example
 * ```typescript
 * const confirmPasswordIdentifier = FlowBuilderElementConstants.CONFIRM_PASSWORD_IDENTIFIER;
 * ```
 */
class FlowBuilderElementConstants {
  /**
   * Private constructor to avoid object instantiation from outside the class.
   */
  // eslint-disable-next-line @typescript-eslint/no-empty-function
  private constructor() {}

  public static readonly PASSWORD_IDENTIFIER: string = 'password';

  public static readonly CONFIRM_PASSWORD_IDENTIFIER: string = 'confirmPassword';

  public static readonly DEFAULT_CAPTCHA_PROVIDER: string = 'ReCAPTCHA V2';
}

export default FlowBuilderElementConstants;
