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
import FlowConstants from '../FlowConstants';

describe('FlowConstants', () => {
  describe('AUTO_SAVE_INTERVAL', () => {
    it('should have AUTO_SAVE_INTERVAL set to 60000ms (1 minute)', () => {
      expect(FlowConstants.AUTO_SAVE_INTERVAL).toBe(60000);
    });

    it('should be a readonly property', () => {
      expect(typeof FlowConstants.AUTO_SAVE_INTERVAL).toBe('number');
    });
  });

  describe('MAX_HISTORY_ITEMS', () => {
    it('should have MAX_HISTORY_ITEMS set to 20', () => {
      expect(FlowConstants.MAX_HISTORY_ITEMS).toBe(20);
    });

    it('should be a readonly property', () => {
      expect(typeof FlowConstants.MAX_HISTORY_ITEMS).toBe('number');
    });
  });

  describe('class structure', () => {
    it('should have both constants defined', () => {
      expect(FlowConstants).toHaveProperty('AUTO_SAVE_INTERVAL');
      expect(FlowConstants).toHaveProperty('MAX_HISTORY_ITEMS');
    });

    it('should be accessible as static properties', () => {
      // Verify constants are accessible without instantiation
      const autoSaveInterval = FlowConstants.AUTO_SAVE_INTERVAL;
      const maxHistoryItems = FlowConstants.MAX_HISTORY_ITEMS;

      expect(autoSaveInterval).toBeDefined();
      expect(maxHistoryItems).toBeDefined();
    });
  });
});
