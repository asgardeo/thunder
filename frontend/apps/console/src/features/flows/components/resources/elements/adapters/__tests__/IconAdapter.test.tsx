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

import {describe, it, expect, vi} from 'vitest';
import {render, screen} from '@testing-library/react';
import type {Element as FlowElement} from '@/features/flows/models/elements';
import IconAdapter from '../IconAdapter';

vi.mock('@wso2/oxygen-ui-icons-react', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@wso2/oxygen-ui-icons-react')>();
  return {
    ...actual,
    User: ({size = 14, color = 'currentColor'}: {size?: number; color?: string}) => (
      <svg data-testid="user-icon" data-size={size} data-color={color} />
    ),
  };
});

describe('IconAdapter', () => {
  const createMockElement = (overrides: Record<string, unknown> = {}): FlowElement =>
    ({
      id: 'icon-1',
      type: 'ICON',
      category: 'DISPLAY',
      config: {},
      ...overrides,
    }) as FlowElement;

  describe('Known Icon Rendering', () => {
    it('should render a known icon by name', () => {
      const resource = createMockElement({name: 'User'});

      render(<IconAdapter resource={resource} />);

      expect(screen.getByTestId('user-icon')).toBeInTheDocument();
    });

    it('should use default User icon when name is not provided', () => {
      const resource = createMockElement();

      render(<IconAdapter resource={resource} />);

      expect(screen.getByTestId('user-icon')).toBeInTheDocument();
    });

    it('should pass size to icon component', () => {
      const resource = createMockElement({name: 'User', size: 32});

      render(<IconAdapter resource={resource} />);

      expect(screen.getByTestId('user-icon')).toHaveAttribute('data-size', '32');
    });

    it('should use default size 24 when size is not provided', () => {
      const resource = createMockElement({name: 'User'});

      render(<IconAdapter resource={resource} />);

      expect(screen.getByTestId('user-icon')).toHaveAttribute('data-size', '24');
    });

    it('should pass color to icon component', () => {
      const resource = createMockElement({name: 'User', color: 'red'});

      render(<IconAdapter resource={resource} />);

      expect(screen.getByTestId('user-icon')).toHaveAttribute('data-color', 'red');
    });

    it('should use default color currentColor when color is not provided', () => {
      const resource = createMockElement({name: 'User'});

      render(<IconAdapter resource={resource} />);

      expect(screen.getByTestId('user-icon')).toHaveAttribute('data-color', 'currentColor');
    });
  });

  describe('Unknown Icon Placeholder', () => {
    it('should render placeholder when icon name is not found', () => {
      const resource = createMockElement({name: 'NonExistentIconXyz'});

      const {container} = render(<IconAdapter resource={resource} />);

      expect(container.textContent).toContain('?');
      expect(screen.queryByTestId('user-icon')).not.toBeInTheDocument();
    });

    it('should render placeholder box when icon is unknown', () => {
      const resource = createMockElement({name: 'UnknownIcon123'});

      const {container} = render(<IconAdapter resource={resource} />);

      expect(container.firstChild).toBeInTheDocument();
      expect(container.textContent).toContain('?');
    });

    it('should use provided size for placeholder dimensions', () => {
      const resource = createMockElement({name: 'NonExistentIcon', size: 48});

      const {container} = render(<IconAdapter resource={resource} />);

      expect(container.textContent).toContain('?');
    });
  });

  describe('Resource Properties', () => {
    it('should use resource id as part of element identity', () => {
      const resource = createMockElement({id: 'my-icon', name: 'User'});

      render(<IconAdapter resource={resource} />);

      expect(screen.getByTestId('user-icon')).toBeInTheDocument();
    });
  });
});
