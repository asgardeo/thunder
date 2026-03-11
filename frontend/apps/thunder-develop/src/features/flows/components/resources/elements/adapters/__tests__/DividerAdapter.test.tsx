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
import {render} from '@thunder/test-utils/browser';
import {page} from 'vitest/browser';
import type {ReactNode} from 'react';
import {DividerVariants, type Element as FlowElement} from '@/features/flows/models/elements';
import DividerAdapter from '../DividerAdapter';

// Mock dependencies
vi.mock('@/features/flows/hooks/useRequiredFields', () => ({
  default: vi.fn(),
}));

describe('DividerAdapter', () => {
  const createMockElement = (overrides: Partial<FlowElement> & Record<string, unknown> = {}): FlowElement =>
    ({
      id: 'divider-1',
      type: 'DIVIDER',
      category: 'DISPLAY',
      config: {},
      variant: DividerVariants.Horizontal,
      ...overrides,
    }) as FlowElement;

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render a Divider component', async () => {
      const resource = createMockElement();

      const {container} = await render(<DividerAdapter resource={resource} />);

      expect(container.querySelector('hr')).toBeInTheDocument();
    });

    it('should render label text when provided', async () => {
      const resource = createMockElement({label: 'OR'});

      await render(<DividerAdapter resource={resource} />);

      await expect.element(page.getByText('OR')).toBeInTheDocument();
    });

    it('should handle empty label', async () => {
      const resource = createMockElement({label: ''});

      const {container} = await render(<DividerAdapter resource={resource} />);

      expect(container.querySelector('hr')).toBeInTheDocument();
    });

    it('should handle undefined label', async () => {
      const resource = createMockElement({label: undefined});

      const {container} = await render(<DividerAdapter resource={resource} />);

      expect(container.querySelector('hr')).toBeInTheDocument();
    });
  });

  describe('Divider Variants', () => {
    it('should render horizontal divider', async () => {
      const resource = createMockElement({variant: DividerVariants.Horizontal});

      const {container} = await render(<DividerAdapter resource={resource} />);

      const divider = container.querySelector('hr');
      expect(divider).toBeInTheDocument();
    });

    it('should render vertical divider', async () => {
      const resource = createMockElement({variant: DividerVariants.Vertical});

      const {container} = await render(<DividerAdapter resource={resource} />);

      // Vertical dividers in MUI may render as div instead of hr
      const divider = container.querySelector('.MuiDivider-root');
      expect(divider).toBeInTheDocument();
    });

    it('should handle other variant values as MUI variant', async () => {
      const resource = createMockElement({variant: 'fullWidth' as typeof DividerVariants.Horizontal});

      const {container} = await render(<DividerAdapter resource={resource} />);

      const divider = container.querySelector('hr');
      expect(divider).toBeInTheDocument();
    });

    it('should handle undefined variant', async () => {
      const resource = createMockElement({variant: undefined});

      const {container} = await render(<DividerAdapter resource={resource} />);

      const divider = container.querySelector('hr');
      expect(divider).toBeInTheDocument();
    });
  });

  describe('Validation', () => {
    it('should call useRequiredFields with resource', async () => {
      const useRequiredFields = await import('@/features/flows/hooks/useRequiredFields');
      const mockUseRequiredFields = vi.mocked(useRequiredFields.default);

      const resource = createMockElement();

      await render(<DividerAdapter resource={resource} />);

      expect(mockUseRequiredFields).toHaveBeenCalled();
    });
  });

  describe('Different Resource IDs', () => {
    it('should render with different resource IDs', async () => {
      const resource1 = createMockElement({id: 'divider-1', label: 'First'});
      const resource2 = createMockElement({id: 'divider-2', label: 'Second'});

      await render(<DividerAdapter resource={resource1} />);
      await render(<DividerAdapter resource={resource2} />);

      await expect.element(page.getByText('First')).toBeInTheDocument();
      await expect.element(page.getByText('Second')).toBeInTheDocument();
    });
  });
});
