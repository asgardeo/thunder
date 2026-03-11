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
import type {Element as FlowElement} from '@/features/flows/models/elements';
import RichTextAdapter from '../RichTextAdapter';

// Mock dependencies
vi.mock('../RichTextAdapter.scss', () => ({}));

vi.mock('@/features/flows/hooks/useRequiredFields', () => ({
  default: vi.fn(),
}));

vi.mock('../PlaceholderComponent', () => ({
  default: ({value, children}: {value: string; children?: ReactNode}) => (
    <span data-testid="placeholder" data-value={value}>
      {children ?? value}
    </span>
  ),
}));

describe('RichTextAdapter', () => {
  const createMockElement = (overrides: Partial<FlowElement> & Record<string, unknown> = {}): FlowElement =>
    ({
      id: 'richtext-1',
      type: 'RICH_TEXT',
      category: 'DISPLAY',
      config: {},
      label: 'Hello <b>World</b>',
      ...overrides,
    }) as FlowElement;

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render with rich-text-content class', async () => {
      const resource = createMockElement();

      const {container} = await render(<RichTextAdapter resource={resource} />);

      expect(container.querySelector('.rich-text-content')).toBeInTheDocument();
    });

    it('should render PlaceholderComponent', async () => {
      const resource = createMockElement();

      await render(<RichTextAdapter resource={resource} />);

      await expect.element(page.getByTestId('placeholder')).toBeInTheDocument();
    });

    it('should pass label to PlaceholderComponent', async () => {
      const resource = createMockElement({label: 'Test Label'});

      await render(<RichTextAdapter resource={resource} />);

      await expect.element(page.getByTestId('placeholder')).toHaveAttribute('data-value', 'Test Label');
    });
  });

  describe('HTML Sanitization', () => {
    it('should render sanitized HTML content', async () => {
      const resource = createMockElement({label: '<p>Paragraph content</p>'});

      await render(<RichTextAdapter resource={resource} />);

      await expect.element(page.getByTestId('placeholder')).toBeInTheDocument();
    });

    it('should handle plain text content', async () => {
      const resource = createMockElement({label: 'Plain text without HTML'});

      await render(<RichTextAdapter resource={resource} />);

      await expect.element(page.getByTestId('placeholder')).toHaveAttribute('data-value', 'Plain text without HTML');
    });
  });

  describe('Empty Content', () => {
    it('should handle empty label', async () => {
      const resource = createMockElement({label: ''});

      await render(<RichTextAdapter resource={resource} />);

      await expect.element(page.getByTestId('placeholder')).toHaveAttribute('data-value', '');
    });

    it('should handle undefined label', async () => {
      const resource = createMockElement({label: undefined});

      await render(<RichTextAdapter resource={resource} />);

      await expect.element(page.getByTestId('placeholder')).toHaveAttribute('data-value', '');
    });
  });

  describe('Validation', () => {
    it('should call useRequiredFields with resource', async () => {
      const useRequiredFields = await import('@/features/flows/hooks/useRequiredFields');
      const mockUseRequiredFields = vi.mocked(useRequiredFields.default);

      const resource = createMockElement();

      await render(<RichTextAdapter resource={resource} />);

      expect(mockUseRequiredFields).toHaveBeenCalled();
    });
  });

  describe('Different Resource IDs', () => {
    it('should render with different resource IDs', async () => {
      const resource1 = createMockElement({id: 'richtext-1', label: 'First'});
      const resource2 = createMockElement({id: 'richtext-2', label: 'Second'});

      const {container: container1} = await render(<RichTextAdapter resource={resource1} />);
      const {container: container2} = await render(<RichTextAdapter resource={resource2} />);

      expect(container1.querySelector('.rich-text-content')).toBeInTheDocument();
      expect(container2.querySelector('.rich-text-content')).toBeInTheDocument();
    });
  });

  describe('Anchor Tag Security', () => {
    it('should handle anchor tags with target="_blank"', async () => {
      const resource = createMockElement({
        label: '<a href="https://example.com" target="_blank">External Link</a>',
      });

      await render(<RichTextAdapter resource={resource} />);

      // The content should be sanitized and rendered
      await expect.element(page.getByTestId('placeholder')).toBeInTheDocument();
    });

    it('should handle anchor tags without target attribute', async () => {
      const resource = createMockElement({
        label: '<a href="https://example.com">Regular Link</a>',
      });

      await render(<RichTextAdapter resource={resource} />);

      await expect.element(page.getByTestId('placeholder')).toBeInTheDocument();
    });

    it('should handle anchor tags with target="_self"', async () => {
      const resource = createMockElement({
        label: '<a href="https://example.com" target="_self">Same Window Link</a>',
      });

      await render(<RichTextAdapter resource={resource} />);

      await expect.element(page.getByTestId('placeholder')).toBeInTheDocument();
    });

    it('should handle multiple anchor tags', async () => {
      const resource = createMockElement({
        label:
          '<a href="https://link1.com" target="_blank">Link 1</a> and <a href="https://link2.com">Link 2</a>',
      });

      await render(<RichTextAdapter resource={resource} />);

      await expect.element(page.getByTestId('placeholder')).toBeInTheDocument();
    });
  });
});
