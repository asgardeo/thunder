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

import cloneDeep from 'lodash-es/cloneDeep';
import get from 'lodash-es/get';
import {type ReactElement, useCallback, useEffect} from 'react';
import Notification, {NotificationType} from '../models/notification';
import type {Resource} from '../models/resources';
import useValidationStatus from './useValidationStatus';
import ValidationConstants from '../constants/ValidationConstants';

const IDP_NAME_PLACEHOLDER = '{{IDP_NAME}}';
const IDP_ID_PLACEHOLDER = '{{IDP_ID}}';

/**
 * Interface for the required field.
 */
export interface RequiredFieldInterface {
  /**
   * The name of the required field.
   */
  name: string;
  /**
   * The error message for the required field.
   */
  errorMessage: string;
}

/**
 * Custom hook to manage required fields validation.
 */
const useRequiredFields = (
  resource: Resource,
  generalMessage: string | ReactElement,
  fields: RequiredFieldInterface[],
) => {
  const {addNotification, removeNotification, getNotification} = useValidationStatus();

  /**
   * Builds the error ID for a required field.
   * @param fieldName - The name of the field.
   * @returns The error ID.
   */
  const buildErrorId = useCallback(
    (): string => `${resource.id}_${ValidationConstants.REQUIRED_FIELD_ERROR_CODE}`,
    [resource.id],
  );

  /**
   * Builds the error ID for a field.
   * @param fieldName - The name of the field.
   * @returns The error ID.
   */
  const buildFieldErrorId = useCallback((fieldName: string): string => `${resource.id}_${fieldName}`, [resource.id]);

  /**
   * Checks if a nested property exists in an object.
   *
   * @param obj - The object to check.
   * @param path - The path to the nested property.
   * @returns True if the property exists, false otherwise.
   */
  const getNestedProperty = useCallback((obj: Resource, path: string): string => {
    if (!obj || !path.includes('.')) {
      return '';
    }

    const value: string | null = get(obj, path, null);

    return value === IDP_NAME_PLACEHOLDER || value === IDP_ID_PLACEHOLDER ? '' : (value ?? '');
  }, []);

  useEffect(() => {
    if (!resource || !fields || fields.length === 0) {
      return;
    }

    fields.forEach((field: RequiredFieldInterface) => {
      const errorId: string = buildErrorId();

      if (
        !resource?.config?.[field.name as keyof typeof resource.config] &&
        !resource?.[field.name as keyof Resource] &&
        !getNestedProperty(resource, field.name)
      ) {
        const notification = getNotification(errorId);

        if (!notification) {
          const error: Notification = new Notification(errorId, generalMessage, NotificationType.ERROR);

          error.addResource(resource);
          error.addResourceFieldNotification(buildFieldErrorId(field.name), field.errorMessage);

          addNotification?.(error);
        } else if (!notification.hasResourceFieldNotification(buildFieldErrorId(field.name))) {
          const existingError: Notification = cloneDeep(notification);

          existingError.addResource(resource);
          existingError.addResourceFieldNotification(buildFieldErrorId(field.name), field.errorMessage);
          addNotification?.(existingError);
        }
      } else {
        const notification = getNotification(errorId);

        if (notification?.hasResourceFieldNotification(buildFieldErrorId(field.name))) {
          if (notification.getResourceFieldNotifications().size === 1) {
            removeNotification?.(errorId);
          } else {
            const existingError: Notification = cloneDeep(notification);

            existingError.addResource(resource);
            existingError.removeResourceFieldNotification(buildFieldErrorId(field.name));
            addNotification?.(existingError);
          }
        }
      }
    });
  }, [
    resource,
    fields,
    generalMessage,
    getNotification,
    addNotification,
    removeNotification,
    buildErrorId,
    buildFieldErrorId,
    getNestedProperty,
  ]);

  /**
   * Cleanup function to remove notifications on unmount.
   */
  useEffect(
    () => () => {
      removeNotification?.(buildErrorId());
    },
    [buildErrorId, removeNotification],
  );
};

export default useRequiredFields;
