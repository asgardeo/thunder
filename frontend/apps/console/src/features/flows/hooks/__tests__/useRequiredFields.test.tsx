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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {renderHook} from '@testing-library/react';
import type {ReactNode} from 'react';
import {ValidationContext, type ValidationContextProps} from '../../context/ValidationContext';
import useRequiredFields, {type RequiredFieldInterface} from '../useRequiredFields';
import type {Resource} from '../../models/resources';
import Notification from '../../models/notification';

describe('useRequiredFields', () => {
  const mockAddNotification = vi.fn();
  const mockRemoveNotification = vi.fn();
  const mockGetNotification = vi.fn();

  const mockContextValue: ValidationContextProps = {
    isValid: true,
    notifications: [],
    getNotification: mockGetNotification,
    addNotification: mockAddNotification,
    removeNotification: mockRemoveNotification,
    validationConfig: {
      isOTPValidationEnabled: false,
      isRecoveryFactorValidationEnabled: false,
      isPasswordExecutorValidationEnabled: false,
    },
  };

  const createWrapper = (contextValue: ValidationContextProps = mockContextValue) => {
    function Wrapper({children}: {children: ReactNode}) {
      return <ValidationContext.Provider value={contextValue}>{children}</ValidationContext.Provider>;
    }
    return Wrapper;
  };

  const createResource = (overrides: Partial<Resource> = {}): Resource =>
    ({
      id: 'resource-1',
      type: 'BUTTON',
      config: {
        field: {name: '', type: {}},
        styles: {},
      },
      ...overrides,
    }) as Resource;

  beforeEach(() => {
    vi.clearAllMocks();
    mockGetNotification.mockReturnValue(null);
  });

  describe('Basic Functionality', () => {
    it('should not throw when used with provider', () => {
      const resource = createResource();
      const fields: RequiredFieldInterface[] = [{name: 'label', errorMessage: 'Label is required'}];

      expect(() => {
        renderHook(() => useRequiredFields(resource, 'Required field missing', fields), {
          wrapper: createWrapper(),
        });
      }).not.toThrow();
    });

    it('should add notification when required field is missing', () => {
      const resource = createResource();
      const fields: RequiredFieldInterface[] = [{name: 'label', errorMessage: 'Label is required'}];

      renderHook(() => useRequiredFields(resource, 'Required field missing', fields), {
        wrapper: createWrapper(),
      });

      expect(mockAddNotification).toHaveBeenCalled();
    });

    it('should not add notification when required field is present in config', () => {
      const resource = createResource({
        config: {field: {name: '', type: {}}, styles: {}, label: 'Test Label'},
      } as Partial<Resource>);
      const fields: RequiredFieldInterface[] = [{name: 'label', errorMessage: 'Label is required'}];

      renderHook(() => useRequiredFields(resource, 'Required field missing', fields), {
        wrapper: createWrapper(),
      });

      expect(mockAddNotification).not.toHaveBeenCalled();
    });

    it('should not add notification when required field is present on resource', () => {
      const resource = createResource();
      (resource as unknown as Record<string, string>).label = 'Test Label';
      const fields: RequiredFieldInterface[] = [{name: 'label', errorMessage: 'Label is required'}];

      renderHook(() => useRequiredFields(resource, 'Required field missing', fields), {
        wrapper: createWrapper(),
      });

      expect(mockAddNotification).not.toHaveBeenCalled();
    });
  });

  describe('Nested Property Validation', () => {
    it('should validate nested properties', () => {
      const resource = createResource();
      (resource as unknown as Record<string, {property: string}>).nested = {property: 'value'};
      const fields: RequiredFieldInterface[] = [{name: 'nested.property', errorMessage: 'Nested property is required'}];

      renderHook(() => useRequiredFields(resource, 'Required field missing', fields), {
        wrapper: createWrapper(),
      });

      expect(mockAddNotification).not.toHaveBeenCalled();
    });

    it('should add notification for missing nested property', () => {
      const resource = createResource();
      const fields: RequiredFieldInterface[] = [{name: 'nested.property', errorMessage: 'Nested property is required'}];

      renderHook(() => useRequiredFields(resource, 'Required field missing', fields), {
        wrapper: createWrapper(),
      });

      expect(mockAddNotification).toHaveBeenCalled();
    });

    it('should treat IDP placeholder as missing value', () => {
      const resource = createResource();
      (resource as unknown as Record<string, {name: string}>).idp = {name: '{{IDP_NAME}}'};
      const fields: RequiredFieldInterface[] = [{name: 'idp.name', errorMessage: 'IDP name is required'}];

      renderHook(() => useRequiredFields(resource, 'Required field missing', fields), {
        wrapper: createWrapper(),
      });

      expect(mockAddNotification).toHaveBeenCalled();
    });

    it('should treat IDP_ID placeholder as missing value', () => {
      const resource = createResource();
      (resource as unknown as Record<string, {id: string}>).idp = {id: '{{IDP_ID}}'};
      const fields: RequiredFieldInterface[] = [{name: 'idp.id', errorMessage: 'IDP ID is required'}];

      renderHook(() => useRequiredFields(resource, 'Required field missing', fields), {
        wrapper: createWrapper(),
      });

      expect(mockAddNotification).toHaveBeenCalled();
    });
  });

  describe('Multiple Fields', () => {
    it('should validate multiple required fields', () => {
      const resource = createResource();
      const fields: RequiredFieldInterface[] = [
        {name: 'label', errorMessage: 'Label is required'},
        {name: 'value', errorMessage: 'Value is required'},
      ];

      renderHook(() => useRequiredFields(resource, 'Required fields missing', fields), {
        wrapper: createWrapper(),
      });

      expect(mockAddNotification).toHaveBeenCalled();
    });

    it('should handle mix of present and missing fields', () => {
      const resource = createResource({
        config: {field: {name: '', type: {}}, styles: {}, label: 'Test'},
      } as Partial<Resource>);
      const fields: RequiredFieldInterface[] = [
        {name: 'label', errorMessage: 'Label is required'},
        {name: 'value', errorMessage: 'Value is required'},
      ];

      renderHook(() => useRequiredFields(resource, 'Required fields missing', fields), {
        wrapper: createWrapper(),
      });

      expect(mockAddNotification).toHaveBeenCalled();
    });
  });

  describe('Existing Notification Handling', () => {
    it('should add field to existing notification when notification exists', () => {
      const existingNotification = new Notification('resource-1_REQUIRED_FIELD', 'Required field missing', 'error');
      mockGetNotification.mockReturnValue(existingNotification);

      const resource = createResource();
      const fields: RequiredFieldInterface[] = [{name: 'label', errorMessage: 'Label is required'}];

      renderHook(() => useRequiredFields(resource, 'Required field missing', fields), {
        wrapper: createWrapper(),
      });

      expect(mockAddNotification).toHaveBeenCalled();
    });

    it('should not add duplicate field notification', () => {
      const existingNotification = new Notification('resource-1_REQUIRED_FIELD', 'Required field missing', 'error');
      existingNotification.addResourceFieldNotification('resource-1_label', 'Label is required');
      mockGetNotification.mockReturnValue(existingNotification);

      const resource = createResource();
      const fields: RequiredFieldInterface[] = [{name: 'label', errorMessage: 'Label is required'}];

      renderHook(() => useRequiredFields(resource, 'Required field missing', fields), {
        wrapper: createWrapper(),
      });

      // Should not add since field notification already exists
      expect(mockAddNotification).not.toHaveBeenCalled();
    });
  });

  describe('Notification Removal', () => {
    it('should remove field notification when field becomes valid', () => {
      const existingNotification = new Notification('resource-1_REQUIRED_FIELD', 'Required field missing', 'error');
      existingNotification.addResourceFieldNotification('resource-1_label', 'Label is required');
      mockGetNotification.mockReturnValue(existingNotification);

      const resource = createResource({
        config: {field: {name: '', type: {}}, styles: {}, label: 'Test Label'},
      } as Partial<Resource>);
      const fields: RequiredFieldInterface[] = [{name: 'label', errorMessage: 'Label is required'}];

      renderHook(() => useRequiredFields(resource, 'Required field missing', fields), {
        wrapper: createWrapper(),
      });

      // Should remove the entire notification since it's the only field
      expect(mockRemoveNotification).toHaveBeenCalledWith('resource-1_REQUIRED_FIELD_ERROR');
    });

    it('should update notification when one of multiple fields becomes valid', () => {
      const existingNotification = new Notification('resource-1_REQUIRED_FIELD_ERROR', 'Required field missing', 'error');
      existingNotification.addResourceFieldNotification('resource-1_label', 'Label is required');
      existingNotification.addResourceFieldNotification('resource-1_value', 'Value is required');
      mockGetNotification.mockReturnValue(existingNotification);

      const resource = createResource({
        config: {field: {name: '', type: {}}, styles: {}, label: 'Test Label'},
      } as Partial<Resource>);
      const fields: RequiredFieldInterface[] = [{name: 'label', errorMessage: 'Label is required'}];

      renderHook(() => useRequiredFields(resource, 'Required field missing', fields), {
        wrapper: createWrapper(),
      });

      // Should update the notification, not remove it entirely
      expect(mockAddNotification).toHaveBeenCalled();
      expect(mockRemoveNotification).not.toHaveBeenCalled();
    });
  });

  describe('Cleanup on Unmount', () => {
    it('should remove notification on unmount', () => {
      const resource = createResource();
      const fields: RequiredFieldInterface[] = [{name: 'label', errorMessage: 'Label is required'}];

      const {unmount} = renderHook(() => useRequiredFields(resource, 'Required field missing', fields), {
        wrapper: createWrapper(),
      });

      unmount();

      expect(mockRemoveNotification).toHaveBeenCalledWith('resource-1_REQUIRED_FIELD_ERROR');
    });
  });

  describe('Edge Cases', () => {
    it('should handle empty fields array', () => {
      const resource = createResource();
      const fields: RequiredFieldInterface[] = [];

      expect(() => {
        renderHook(() => useRequiredFields(resource, 'Required field missing', fields), {
          wrapper: createWrapper(),
        });
      }).not.toThrow();

      expect(mockAddNotification).not.toHaveBeenCalled();
    });

    it('should handle undefined config', () => {
      // Testing runtime edge case where config might be undefined
      const resource = {id: 'resource-1', type: 'BUTTON', config: undefined} as unknown as Resource;
      const fields: RequiredFieldInterface[] = [{name: 'label', errorMessage: 'Label is required'}];

      expect(() => {
        renderHook(() => useRequiredFields(resource, 'Required field missing', fields), {
          wrapper: createWrapper(),
        });
      }).not.toThrow();
    });

    it('should handle resource without config property', () => {
      const resource = {id: 'resource-1', type: 'BUTTON'} as Resource;
      const fields: RequiredFieldInterface[] = [{name: 'label', errorMessage: 'Label is required'}];

      expect(() => {
        renderHook(() => useRequiredFields(resource, 'Required field missing', fields), {
          wrapper: createWrapper(),
        });
      }).not.toThrow();
    });
  });
});
