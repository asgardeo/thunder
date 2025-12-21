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

import {CircleAlertIcon} from '@wso2/oxygen-ui-icons-react';
import classNames from 'classnames';
import {useMemo, useState, type PropsWithChildren, type ReactElement} from 'react';
import type {Resource} from '../../models/resources';
import './ValidationErrorBoundary.scss';
import useValidationStatus from '../../hooks/useValidationStatus';
import Notification, {NotificationType} from '../../models/notification';

/**
 * Props interface of {@link ValidationErrorBoundary}
 */
export interface ValidationErrorBoundaryPropsInterface {
  /**
   * The resource to check for validation errors.
   */
  resource: Resource;
  /**
   * Whether to disable the error boundary on hover.
   */
  disableErrorBoundaryOnHover?: boolean;
}

/**
 * Validation error boundary component that wraps components and shows error indicators.
 * TEST 11: Restore full validation logic.
 *
 * @param props - Props injected to the component.
 * @returns ValidationErrorBoundary component.
 */
function ValidationErrorBoundary({
  resource,
  children = null,
  disableErrorBoundaryOnHover = false,
}: PropsWithChildren<ValidationErrorBoundaryPropsInterface>): ReactElement {
  const {notifications} = useValidationStatus();
  const [active, setActive] = useState<boolean>(false);

  /**
   * Finds the notification for this resource (if any).
   * Prioritizes error notifications over warnings and info.
   */
  const resourceNotification: Notification | null = useMemo(() => {
    // First check for error notifications
    const errorNotification = notifications.find(
      (n: Notification) => n.hasResource(resource.id) && n.getType() === NotificationType.ERROR,
    );

    if (errorNotification) {
      return errorNotification;
    }

    // Then check for warning notifications
    const warningNotification = notifications.find(
      (n: Notification) => n.hasResource(resource.id) && n.getType() === NotificationType.WARNING,
    );

    if (warningNotification) {
      return warningNotification;
    }

    // Finally check for info notifications
    const infoNotification = notifications.find(
      (n: Notification) => n.hasResource(resource.id) && n.getType() === NotificationType.INFO,
    );

    return infoNotification ?? null;
  }, [resource.id, notifications]);

  /**
   * Checks if the resource has any notifications.
   */
  const hasNotification: boolean = resourceNotification !== null;

  /**
   * Gets the notification type for styling.
   */
  const notificationType: NotificationType | null = resourceNotification?.getType() ?? null;

  return (
    <div
      className={classNames({
        active: hasNotification && active && disableErrorBoundaryOnHover,
        [String(notificationType)]: hasNotification && !!notificationType,
        padded: hasNotification && !disableErrorBoundaryOnHover,
        'validation-error-boundary': hasNotification,
      })}
      onMouseOver={() => hasNotification && disableErrorBoundaryOnHover && setActive(true)}
      onFocus={() => hasNotification && disableErrorBoundaryOnHover && setActive(true)}
      onMouseOut={() => hasNotification && disableErrorBoundaryOnHover && setActive(false)}
      onBlur={() => hasNotification && disableErrorBoundaryOnHover && setActive(false)}
    >
      {hasNotification && !(active && disableErrorBoundaryOnHover) && (
        <CircleAlertIcon className="circle-alert-icon" size={24} />
      )}
      {children}
    </div>
  );
}

export default ValidationErrorBoundary;
