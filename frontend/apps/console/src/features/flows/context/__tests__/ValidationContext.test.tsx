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

import {describe, it, expect} from 'vitest';
import {useContext} from 'react';
import {renderHook} from '@testing-library/react';
import {ValidationContext} from '../ValidationContext';

describe('ValidationContext', () => {
  describe('default context values', () => {
    it('should have isValid set to true by default', () => {
      const {result} = renderHook(() => useContext(ValidationContext));

      expect(result.current.isValid).toBe(true);
    });

    it('should have empty notifications array by default', () => {
      const {result} = renderHook(() => useContext(ValidationContext));

      expect(result.current.notifications).toEqual([]);
    });

    it('should have currentActiveTab set to 0 by default', () => {
      const {result} = renderHook(() => useContext(ValidationContext));

      expect(result.current.currentActiveTab).toBe(0);
    });

    it('should have openValidationPanel set to false by default', () => {
      const {result} = renderHook(() => useContext(ValidationContext));

      expect(result.current.openValidationPanel).toBe(false);
    });

    it('should have undefined optional functions by default', () => {
      const {result} = renderHook(() => useContext(ValidationContext));

      expect(result.current.addNotification).toBeUndefined();
      expect(result.current.removeNotification).toBeUndefined();
      expect(result.current.setCurrentActiveTab).toBeUndefined();
      expect(result.current.setOpenValidationPanel).toBeUndefined();
    });

    it('should have undefined selectedNotification by default', () => {
      const {result} = renderHook(() => useContext(ValidationContext));

      expect(result.current.selectedNotification).toBeUndefined();
    });

    it('should have default validationConfig with all options disabled', () => {
      const {result} = renderHook(() => useContext(ValidationContext));

      expect(result.current.validationConfig).toEqual({
        isOTPValidationEnabled: false,
        isPasswordExecutorValidationEnabled: false,
        isRecoveryFactorValidationEnabled: false,
      });
    });

    it('should have getNotification function that returns undefined', () => {
      const {result} = renderHook(() => useContext(ValidationContext));

      expect(result.current.getNotification('any-id')).toBeUndefined();
    });
  });
});
