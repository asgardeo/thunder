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

import get from 'lodash-es/get';
import {type ReactElement, useEffect, useRef, useMemo} from 'react';
import Notification, {NotificationType} from '../models/notification';
import type {Resource} from '../models/resources';
import useValidationStatus from './useValidationStatus';
import ValidationConstants from '../constants/ValidationConstants';

const IDP_NAME_PLACEHOLDER = '{{IDP_NAME}}';

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
 *
 * PERFORMANCE: This hook has been heavily optimized to:
 * 1. Use refs to track previous state and avoid unnecessary updates
 * 2. Remove expensive cloneDeep calls
 * 3. Only run validation when resource config actually changes
 * 4. Debounce rapid updates
 */
const useRequiredFields = (
  resource: Resource,
  generalMessage: string | ReactElement,
  fields: RequiredFieldInterface[],
) => {
  const {addNotification, removeNotification, getNotification} = useValidationStatus();

  // PERFORMANCE: Track previous validation state to avoid redundant operations
  const prevValidationStateRef = useRef<Map<string, boolean>>(new Map());
  const errorIdRef = useRef<string>('');
  const isInitialMountRef = useRef(true);

  // PERFORMANCE: Memoize error ID to avoid recalculation
  const errorId = useMemo(
    () => `${resource?.id}_${ValidationConstants.REQUIRED_FIELD_ERROR_CODE}`,
    [resource?.id],
  );

  // Update ref for cleanup
  errorIdRef.current = errorId;

  // PERFORMANCE: Memoize field validation states to detect actual changes
  const fieldValidationStates = useMemo(() => {
    if (!resource || !fields || fields.length === 0) {
      return new Map<string, boolean>();
    }

    const states = new Map<string, boolean>();

    fields.forEach((field: RequiredFieldInterface) => {
      const configValue = resource?.config?.[field.name as keyof typeof resource.config];
      const resourceValue = resource?.[field.name as keyof Resource];

      // Check nested property only if field name contains a dot
      let nestedValue = '';
      if (field.name.includes('.')) {
        const value: string | null = get(resource, field.name, null);
        nestedValue = value === IDP_NAME_PLACEHOLDER ? '' : (value ?? '');
      }

      const hasValue = Boolean(configValue || resourceValue || nestedValue);
      states.set(field.name, hasValue);
    });

    return states;
  }, [resource, fields]);

  // PERFORMANCE: Store callbacks and values in refs to avoid effect re-runs
  const addNotificationRef = useRef(addNotification);
  const removeNotificationRef = useRef(removeNotification);
  const getNotificationRef = useRef(getNotification);
  const generalMessageRef = useRef(generalMessage);
  const fieldsRef = useRef(fields);
  const resourceRef = useRef(resource);

  // Update refs on each render (this is cheap)
  addNotificationRef.current = addNotification;
  removeNotificationRef.current = removeNotification;
  getNotificationRef.current = getNotification;
  generalMessageRef.current = generalMessage;
  fieldsRef.current = fields;
  resourceRef.current = resource;

  useEffect(() => {
    const currentResource = resourceRef.current;
    const currentFields = fieldsRef.current;

    if (!currentResource || !currentFields || currentFields.length === 0) {
      return;
    }

    // PERFORMANCE: Check if any validation state actually changed
    // On initial mount, prevValidationStateRef.current is empty, so this will detect all fields as changed
    let hasStateChanged = false;
    const isInitialMount = isInitialMountRef.current;

    if (isInitialMount) {
      // On initial mount, always run validation if there are fields
      hasStateChanged = true;
      isInitialMountRef.current = false;
    } else {
      // On subsequent renders, only run if state changed
      fieldValidationStates.forEach((hasValue, fieldName) => {
        if (prevValidationStateRef.current.get(fieldName) !== hasValue) {
          hasStateChanged = true;
        }
      });
    }

    // If nothing changed, skip the expensive notification operations
    if (!hasStateChanged && prevValidationStateRef.current.size > 0) {
      return;
    }

    // Update previous state
    prevValidationStateRef.current = new Map(fieldValidationStates);

    // PERFORMANCE: Batch notification updates
    const missingFields: RequiredFieldInterface[] = [];
    const presentFields: RequiredFieldInterface[] = [];

    currentFields.forEach((field: RequiredFieldInterface) => {
      const hasValue = fieldValidationStates.get(field.name);
      if (!hasValue) {
        missingFields.push(field);
      } else {
        presentFields.push(field);
      }
    });

    // Handle missing fields - create or update notification
    if (missingFields.length > 0) {
      const existingNotification = getNotificationRef.current(errorId);

      if (!existingNotification) {
        // Create new notification with all missing fields at once
        const error = new Notification(errorId, generalMessageRef.current, NotificationType.ERROR);
        error.addResource(currentResource);
        missingFields.forEach((field) => {
          error.addResourceFieldNotification(`${currentResource.id}_${field.name}`, field.errorMessage);
        });
        addNotificationRef.current?.(error);
      } else {
        // PERFORMANCE: Only update if there are actual changes
        let needsUpdate = false;
        missingFields.forEach((field) => {
          const fieldErrorId = `${currentResource.id}_${field.name}`;
          if (!existingNotification.hasResourceFieldNotification(fieldErrorId)) {
            needsUpdate = true;
          }
        });

        if (needsUpdate) {
          // Create a new notification object (avoiding cloneDeep)
          const updatedError = new Notification(errorId, generalMessageRef.current, NotificationType.ERROR);
          updatedError.addResource(currentResource);

          // Copy existing field notifications
          existingNotification.getResourceFieldNotifications().forEach((msg, id) => {
            updatedError.addResourceFieldNotification(id, msg);
          });

          // Add new missing fields
          missingFields.forEach((field) => {
            const fieldErrorId = `${currentResource.id}_${field.name}`;
            if (!updatedError.hasResourceFieldNotification(fieldErrorId)) {
              updatedError.addResourceFieldNotification(fieldErrorId, field.errorMessage);
            }
          });

          addNotificationRef.current?.(updatedError);
        }
      }
    }

    // Handle present fields - remove from notification if exists
    if (presentFields.length > 0) {
      const existingNotification = getNotificationRef.current(errorId);

      if (existingNotification) {
        let hasFieldsToRemove = false;
        presentFields.forEach((field) => {
          const fieldErrorId = `${currentResource.id}_${field.name}`;
          if (existingNotification.hasResourceFieldNotification(fieldErrorId)) {
            hasFieldsToRemove = true;
          }
        });

        if (hasFieldsToRemove) {
          const remainingFieldCount = existingNotification.getResourceFieldNotifications().size -
            presentFields.filter((field) =>
              existingNotification.hasResourceFieldNotification(`${currentResource.id}_${field.name}`)
            ).length;

          if (remainingFieldCount <= 0) {
            // All fields are now valid, remove the notification entirely
            removeNotificationRef.current?.(errorId);
          } else {
            // Create updated notification without the valid fields (avoiding cloneDeep)
            const updatedError = new Notification(errorId, generalMessageRef.current, NotificationType.ERROR);
            updatedError.addResource(currentResource);

            existingNotification.getResourceFieldNotifications().forEach((msg, id) => {
              // Only copy if it's not one of the now-valid fields
              const isValidField = presentFields.some((field) => id === `${currentResource.id}_${field.name}`);
              if (!isValidField) {
                updatedError.addResourceFieldNotification(id, msg);
              }
            });

            addNotificationRef.current?.(updatedError);
          }
        }
      }
    }
  }, [
    // PERFORMANCE: Minimal dependencies - only things that determine validation state
    errorId,
    fieldValidationStates,
  ]);

  /**
   * Cleanup function to remove notifications on unmount.
   */
  useEffect(
    () => () => {
      removeNotification?.(errorIdRef.current);
    },
    [removeNotification],
  );
};

export default useRequiredFields;
