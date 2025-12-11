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

import type {ReactElement} from 'react';
import type {Resource} from './resources';

const NotificationType = {
  WARNING: 'warning',
  ERROR: 'error',
  INFO: 'info',
} as const;

type NotificationType = (typeof NotificationType)[keyof typeof NotificationType];

export {NotificationType};
export type {NotificationType as NotificationTypeType};

class Notification {
  private readonly id: string;

  private readonly message: string | ReactElement;

  private readonly type: NotificationType;

  private readonly resources: Map<string, Resource>;

  private PanelNotification: ReactElement | undefined;

  private resourceFieldNotifications: Map<string, string>;

  constructor(id: string, message: string | ReactElement, type: NotificationType) {
    this.id = id;
    this.message = message;
    this.type = type;
    this.resources = new Map<string, Resource>();
    this.resourceFieldNotifications = new Map<string, string>();
  }

  getId(): string {
    return this.id;
  }

  getMessage(): string | ReactElement {
    return this.message;
  }

  getType(): NotificationType {
    return this.type;
  }

  addResource(resource: Resource): void {
    this.resources.set(resource.id, resource);
  }

  getResources(): Resource[] {
    return Array.from(this.resources.values());
  }

  hasResources(): boolean {
    return this.resources ? this.resources.size > 0 : false;
  }

  getResource(id: string): Resource | undefined {
    return this.resources.get(id);
  }

  hasResource(id: string): boolean {
    return this.resources ? this.resources.has(id) : false;
  }

  setPanelNotification(notificationEl: ReactElement): void {
    this.PanelNotification = notificationEl;
  }

  getPanelNotification(): ReactElement | undefined {
    return this.PanelNotification;
  }

  hasPanelNotification(): boolean {
    return !!this.PanelNotification;
  }

  addResourceFieldNotification(field: string, msg: string): void {
    this.resourceFieldNotifications.set(field, msg);
  }

  getResourceFieldNotification(field: string): string {
    return this.resourceFieldNotifications.get(field) ?? '';
  }

  hasResourceFieldNotification(field: string): boolean {
    return this.resourceFieldNotifications ? this.resourceFieldNotifications.has(field) : false;
  }

  removeResourceFieldNotification(field: string): void {
    this.resourceFieldNotifications.delete(field);
  }

  getResourceFieldNotifications(): Map<string, string> {
    return this.resourceFieldNotifications;
  }
}

export default Notification;
