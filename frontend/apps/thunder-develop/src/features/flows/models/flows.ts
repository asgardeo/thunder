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

import type {FlowTypes} from './metadata';
import type {BaseConfig} from './base';

/**
 * Interface for Flow completion configurations.
 * Flow completion configs originate from end-step resources, which expose `BaseConfig`
 * metadata in addition to arbitrary backend data.
 */
export type FlowCompletionConfigsInterface = BaseConfig | Record<string, unknown>;

/**
 * Interface for Flow local history.
 */
export interface FlowsHistoryInterface {
  /**
   * Author of the change.
   */
  author: {
    userName: string;
  };
  /**
   * Entire flow as an object.
   */
  flowData: Record<string, unknown>;
  /**
   * Flow saved at timestamp.
   */
  timestamp: number;
}

export interface FlowConfigInterface {
  flowType: FlowTypes;
  isEnabled: boolean;
  flowCompletionConfigs: FlowCompletionConfigsInterface;
}
