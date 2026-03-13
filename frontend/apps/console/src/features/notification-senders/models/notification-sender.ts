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
 * Supported notification sender providers
 */
export type NotificationSenderProvider = 'twilio' | 'vonage' | 'custom';

/**
 * Property of a notification sender
 */
export interface NotificationSenderProperty {
  name: string;
  value: string;
  is_secret?: boolean;
}

/**
 * Message notification sender model
 */
export interface NotificationSender {
  id: string;
  name: string;
  description?: string;
  provider: NotificationSenderProvider;
  properties?: NotificationSenderProperty[];
}

/**
 * Response type for listing notification senders
 */
export type NotificationSenderListResponse = NotificationSender[];
