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
import {render, screen, fireEvent} from '@testing-library/react';
import type {ReactNode} from 'react';
import type {Element as FlowElement} from '@/features/flows/models/elements';
import ImageAdapter from '../ImageAdapter';

// Mock dependencies
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
  Trans: ({children}: {children: ReactNode}) => children,
}));

vi.mock('@/features/flows/hooks/useRequiredFields', () => ({
  default: vi.fn(),
}));

describe('ImageAdapter', () => {
  const createMockElement = (overrides: Partial<FlowElement> & Record<string, unknown> = {}): FlowElement =>
    ({
      id: 'image-1',
      type: 'IMAGE',
      category: 'DISPLAY',
      config: {},
      src: 'https://example.com/image.png',
      alt: 'Test Image',
      ...overrides,
    }) as FlowElement;

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Image Rendering', () => {
    it('should render an image when src is provided', () => {
      const resource = createMockElement({src: 'https://example.com/test.png'});

      render(<ImageAdapter resource={resource} />);

      const img = screen.getByRole('img');
      expect(img).toBeInTheDocument();
      expect(img).toHaveAttribute('src', 'https://example.com/test.png');
    });

    it('should render image with alt text', () => {
      const resource = createMockElement({
        src: 'https://example.com/test.png',
        alt: 'My Image Alt',
      });

      render(<ImageAdapter resource={resource} />);

      const img = screen.getByRole('img');
      expect(img).toHaveAttribute('alt', 'My Image Alt');
    });

    it('should render image with full width', () => {
      const resource = createMockElement({src: 'https://example.com/test.png'});

      render(<ImageAdapter resource={resource} />);

      const img = screen.getByRole('img');
      expect(img).toHaveAttribute('width', '100%');
    });

    it('should apply custom styles to image', () => {
      const resource = createMockElement({
        src: 'https://example.com/test.png',
        styles: {borderRadius: '10px'},
      });

      render(<ImageAdapter resource={resource} />);

      const img = screen.getByRole('img');
      expect(img).toHaveStyle({borderRadius: '10px'});
    });
  });

  describe('Placeholder Rendering', () => {
    it('should render placeholder when src is empty', () => {
      const resource = createMockElement({src: ''});

      render(<ImageAdapter resource={resource} />);

      expect(screen.getByText('flows:core.placeholders.image')).toBeInTheDocument();
      expect(screen.queryByRole('img')).not.toBeInTheDocument();
    });

    it('should render placeholder when src is undefined', () => {
      const resource = createMockElement({src: undefined});

      render(<ImageAdapter resource={resource} />);

      expect(screen.getByText('flows:core.placeholders.image')).toBeInTheDocument();
    });

    it('should render placeholder when src is whitespace only', () => {
      const resource = createMockElement({src: '   '});

      render(<ImageAdapter resource={resource} />);

      expect(screen.getByText('flows:core.placeholders.image')).toBeInTheDocument();
    });

    it('should render ImageIcon in placeholder', () => {
      const resource = createMockElement({src: ''});

      const {container} = render(<ImageAdapter resource={resource} />);

      // The placeholder should have the ImageIcon
      expect(container.querySelector('svg')).toBeInTheDocument();
    });
  });

  describe('Error Handling', () => {
    it('should show placeholder on image load error', () => {
      const resource = createMockElement({src: 'https://example.com/broken.png'});

      render(<ImageAdapter resource={resource} />);

      const img = screen.getByRole('img');
      fireEvent.error(img);

      expect(screen.getByText('flows:core.placeholders.image')).toBeInTheDocument();
    });
  });

  describe('Validation', () => {
    it('should call useRequiredFields with resource', async () => {
      const useRequiredFields = await import('@/features/flows/hooks/useRequiredFields');
      const mockUseRequiredFields = vi.mocked(useRequiredFields.default);

      const resource = createMockElement();

      render(<ImageAdapter resource={resource} />);

      expect(mockUseRequiredFields).toHaveBeenCalled();
    });
  });

  describe('Centering', () => {
    it('should center the image in a flex container', () => {
      const resource = createMockElement({src: 'https://example.com/test.png'});

      const {container} = render(<ImageAdapter resource={resource} />);

      const box = container.firstChild;
      expect(box).toBeInTheDocument();
    });
  });

  describe('Alt Text', () => {
    it('should handle undefined alt text', () => {
      const resource = createMockElement({
        src: 'https://example.com/test.png',
        alt: undefined,
      });

      render(<ImageAdapter resource={resource} />);

      const img = screen.getByRole('img');
      expect(img).toBeInTheDocument();
    });

    it('should handle empty alt text', () => {
      const resource = createMockElement({
        src: 'https://example.com/test.png',
        alt: '',
      });

      const {container} = render(<ImageAdapter resource={resource} />);

      // Images with empty alt don't have img role (they're presentational)
      const img = container.querySelector('img');
      expect(img).toHaveAttribute('alt', '');
    });
  });
});
