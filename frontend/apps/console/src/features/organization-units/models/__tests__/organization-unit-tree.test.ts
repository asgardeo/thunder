/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

import {describe, expect, it} from 'vitest';
import type {OrganizationUnitTreeItem} from '../organization-unit-tree';

describe('Organization Unit Tree Models', () => {
  describe('OrganizationUnitTreeItem', () => {
    it('should have required id, label, and handle properties', () => {
      const item: OrganizationUnitTreeItem = {
        id: '550e8400-e29b-41d4-a716-446655440000',
        label: 'Engineering',
        handle: 'engineering',
      };

      expect(item.id).toBe('550e8400-e29b-41d4-a716-446655440000');
      expect(item.label).toBe('Engineering');
      expect(item.handle).toBe('engineering');
    });

    it('should accept optional properties', () => {
      const item: OrganizationUnitTreeItem = {
        id: '1',
        label: 'Engineering',
        handle: 'engineering',
        description: 'Software engineering team',
        isPlaceholder: false,
        logo_url: 'https://example.com/logo.png',
      };

      expect(item.description).toBe('Software engineering team');
      expect(item.isPlaceholder).toBe(false);
      expect(item.logo_url).toBe('https://example.com/logo.png');
    });

    it('should accept nested children', () => {
      const item: OrganizationUnitTreeItem = {
        id: 'parent-id',
        label: 'Engineering',
        handle: 'engineering',
        children: [
          {id: 'child-1', label: 'Frontend', handle: 'frontend'},
          {id: 'child-2', label: 'Backend', handle: 'backend'},
        ],
      };

      expect(item.children).toHaveLength(2);
      expect(item.children![0].label).toBe('Frontend');
      expect(item.children![1].label).toBe('Backend');
    });

    it('should support deeply nested tree structure', () => {
      const item: OrganizationUnitTreeItem = {
        id: 'root',
        label: 'Root',
        handle: 'root',
        children: [
          {
            id: 'level-1',
            label: 'Level 1',
            handle: 'level-1',
            children: [
              {id: 'level-2', label: 'Level 2', handle: 'level-2'},
            ],
          },
        ],
      };

      expect(item.children![0].children![0].label).toBe('Level 2');
    });

    it('should support placeholder items for lazy loading', () => {
      const placeholder: OrganizationUnitTreeItem = {
        id: 'placeholder-id',
        label: 'Loading...',
        handle: '',
        isPlaceholder: true,
      };

      expect(placeholder.isPlaceholder).toBe(true);
    });
  });
});
