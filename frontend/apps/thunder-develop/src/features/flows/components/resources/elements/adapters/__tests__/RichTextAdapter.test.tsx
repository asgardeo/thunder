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
import {render, screen} from '@testing-library/react';
import type {ReactNode} from 'react';
import type {Element as FlowElement} from '@/features/flows/models/elements';
import RichTextAdapter from '../RichTextAdapter';

// Mock dependencies
vi.mock('../RichTextAdapter.scss', () => ({}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
  Trans: ({children}: {children: ReactNode}) => children,
}));

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
    it('should render with rich-text-content class', () => {
      const resource = createMockElement();

      const {container} = render(<RichTextAdapter resource={resource} />);

      expect(container.querySelector('.rich-text-content')).toBeInTheDocument();
    });

    it('should render PlaceholderComponent', () => {
      const resource = createMockElement();

      render(<RichTextAdapter resource={resource} />);

      expect(screen.getByTestId('placeholder')).toBeInTheDocument();
    });

    it('should pass label to PlaceholderComponent', () => {
      const resource = createMockElement({label: 'Test Label'});

      render(<RichTextAdapter resource={resource} />);

      expect(screen.getByTestId('placeholder')).toHaveAttribute('data-value', 'Test Label');
    });
  });

  describe('HTML Sanitization', () => {
    it('should render sanitized HTML content', () => {
      const resource = createMockElement({label: '<p>Paragraph content</p>'});

      render(<RichTextAdapter resource={resource} />);

      expect(screen.getByTestId('placeholder')).toBeInTheDocument();
    });

    it('should handle plain text content', () => {
      const resource = createMockElement({label: 'Plain text without HTML'});

      render(<RichTextAdapter resource={resource} />);

      expect(screen.getByTestId('placeholder')).toHaveAttribute('data-value', 'Plain text without HTML');
    });
  });

  describe('Empty Content', () => {
    it('should handle empty label', () => {
      const resource = createMockElement({label: ''});

      render(<RichTextAdapter resource={resource} />);

      expect(screen.getByTestId('placeholder')).toHaveAttribute('data-value', '');
    });

    it('should handle undefined label', () => {
      const resource = createMockElement({label: undefined});

      render(<RichTextAdapter resource={resource} />);

      expect(screen.getByTestId('placeholder')).toHaveAttribute('data-value', '');
    });
  });

  describe('Validation', () => {
    it('should call useRequiredFields with resource', async () => {
      const useRequiredFields = await import('@/features/flows/hooks/useRequiredFields');
      const mockUseRequiredFields = vi.mocked(useRequiredFields.default);

      const resource = createMockElement();

      render(<RichTextAdapter resource={resource} />);

      expect(mockUseRequiredFields).toHaveBeenCalled();
    });
  });

  describe('Different Resource IDs', () => {
    it('should render with different resource IDs', () => {
      const resource1 = createMockElement({id: 'richtext-1', label: 'First'});
      const resource2 = createMockElement({id: 'richtext-2', label: 'Second'});

      const {container: container1} = render(<RichTextAdapter resource={resource1} />);
      const {container: container2} = render(<RichTextAdapter resource={resource2} />);

      expect(container1.querySelector('.rich-text-content')).toBeInTheDocument();
      expect(container2.querySelector('.rich-text-content')).toBeInTheDocument();
    });
  });

  describe('Anchor Tag Security', () => {
    it('should handle anchor tags with target="_blank"', () => {
      const resource = createMockElement({
        label: '<a href="https://example.com" target="_blank">External Link</a>',
      });

      render(<RichTextAdapter resource={resource} />);

      // The content should be sanitized and rendered
      expect(screen.getByTestId('placeholder')).toBeInTheDocument();
    });

    it('should handle anchor tags without target attribute', () => {
      const resource = createMockElement({
        label: '<a href="https://example.com">Regular Link</a>',
      });

      render(<RichTextAdapter resource={resource} />);

      expect(screen.getByTestId('placeholder')).toBeInTheDocument();
    });

    it('should handle anchor tags with target="_self"', () => {
      const resource = createMockElement({
        label: '<a href="https://example.com" target="_self">Same Window Link</a>',
      });

      render(<RichTextAdapter resource={resource} />);

      expect(screen.getByTestId('placeholder')).toBeInTheDocument();
    });

    it('should handle multiple anchor tags', () => {
      const resource = createMockElement({
        label:
          '<a href="https://link1.com" target="_blank">Link 1</a> and <a href="https://link2.com">Link 2</a>',
      });

      render(<RichTextAdapter resource={resource} />);

      expect(screen.getByTestId('placeholder')).toBeInTheDocument();
    });
  });
});
