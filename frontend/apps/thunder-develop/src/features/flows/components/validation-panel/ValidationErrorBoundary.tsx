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
import {memo, useMemo, useState, useCallback, type PropsWithChildren, type ReactElement} from 'react';
import type {Resource} from '../../models/resources';
import './ValidationErrorBoundary.scss';
import useValidationStatus from '../../hooks/useValidationStatus';
import type Notification from '../../models/notification';
import {NotificationType} from '../../models/notification';

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
 *
 * PERFORMANCE: This component has been optimized to:
 * 1. Use useNotificationsOnly instead of useValidationStatus - prevents re-renders
 *    when selectedNotification or openValidationPanel change (which happens on every click)
 * 2. Use a single loop instead of 3 separate .find() calls
 * 3. Memoize event handlers
 * 4. Use memo() to prevent unnecessary re-renders
 *
 * @param props - Props injected to the component.
 * @returns ValidationErrorBoundary component.
 */
function ValidationErrorBoundary({
  resource,
  children = null,
  disableErrorBoundaryOnHover = true,
}: PropsWithChildren<ValidationErrorBoundaryPropsInterface>): ReactElement {
  const {notifications} = useValidationStatus();
  const [active, setActive] = useState<boolean>(false);

  /**
   * PERFORMANCE: Finds the notification for this resource using a single loop.
   * Prioritizes error notifications over warnings and info.
   */
  const resourceNotification: Notification | null = useMemo(() => {
    if (!notifications || notifications.length === 0) {
      return null;
    }

    // PERFORMANCE: Single loop with priority tracking instead of 3 separate .find() calls
    let errorNotification: Notification | null = null;
    let warningNotification: Notification | null = null;
    let infoNotification: Notification | null = null;

    for (const notification of notifications) {
      if (!notification.hasResource(resource.id)) {
        continue;
      }

      const type = notification.getType();

      // Return immediately if we find an error (highest priority)
      if (type === NotificationType.ERROR) {
        return notification;
      }

      // Track other types for fallback
      if (type === NotificationType.WARNING && !warningNotification) {
        warningNotification = notification;
      } else if (type === NotificationType.INFO && !infoNotification) {
        infoNotification = notification;
      }
    }

    // Return in priority order
    return errorNotification ?? warningNotification ?? infoNotification;
  }, [resource.id, notifications]);

  const hasNotification: boolean = resourceNotification !== null;
  const notificationType: NotificationType | null = resourceNotification?.getType() ?? null;

  // PERFORMANCE: Memoize event handlers to prevent inline function recreation
  const handleMouseOver = useCallback(() => {
    if (hasNotification && disableErrorBoundaryOnHover) {
      setActive(true);
    }
  }, [hasNotification, disableErrorBoundaryOnHover]);

  const handleMouseOut = useCallback(() => {
    if (hasNotification && disableErrorBoundaryOnHover) {
      setActive(false);
    }
  }, [hasNotification, disableErrorBoundaryOnHover]);

  // PERFORMANCE: Memoize className to avoid object recreation
  const className = useMemo(() =>
    classNames({
      active: hasNotification && active && disableErrorBoundaryOnHover,
      [String(notificationType)]: hasNotification && !!notificationType,
      padded: hasNotification && !disableErrorBoundaryOnHover,
      'validation-error-boundary': hasNotification,
    }),
    [hasNotification, active, disableErrorBoundaryOnHover, notificationType]
  );

  return (
    <div
      className={className}
      onMouseOver={handleMouseOver}
      onFocus={handleMouseOver}
      onMouseOut={handleMouseOut}
      onBlur={handleMouseOut}
    >
      {hasNotification && !(active && disableErrorBoundaryOnHover) && (
        <CircleAlertIcon className="circle-alert-icon" size={24} />
      )}
      {children}
    </div>
  );
}

// PERFORMANCE: Memoize component to prevent re-renders when props haven't changed
export default memo(ValidationErrorBoundary, (prevProps, nextProps) =>
  prevProps.resource === nextProps.resource &&
  prevProps.resource?.id === nextProps.resource?.id &&
  prevProps.disableErrorBoundaryOnHover === nextProps.disableErrorBoundaryOnHover &&
  prevProps.children === nextProps.children
);
