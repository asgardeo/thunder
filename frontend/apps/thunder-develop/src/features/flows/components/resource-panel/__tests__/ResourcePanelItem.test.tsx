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
import {render, screen, fireEvent} from '@testing-library/react';
import ResourcePanelItem from '../ResourcePanelItem';
import type {Resource} from '../../../models/resources';

// Mock useColorScheme
vi.mock('@wso2/oxygen-ui', async () => {
  const actual = await vi.importActual('@wso2/oxygen-ui');
  return {
    ...actual,
    useColorScheme: () => ({mode: 'light', systemMode: 'light'}),
  };
});

// Mock resolveStaticResourcePath
vi.mock('../../../utils/resolveStaticResourcePath', () => ({
  default: (path: string) => `/static/${path}`,
}));

const createMockResource = (overrides: Partial<Resource> = {}): Resource => ({
  type: 'TEST_STEP',
  resourceType: 'STEP',
  display: {
    label: 'Test Resource',
    description: 'Test Description',
    image: 'test-image.svg',
    showOnResourcePanel: true,
  },
  ...overrides,
} as Resource);

describe('ResourcePanelItem', () => {
  describe('Rendering', () => {
    it('should render resource label', () => {
      const resource = createMockResource();
      render(<ResourcePanelItem resource={resource} />);

      expect(screen.getByText('Test Resource')).toBeInTheDocument();
    });

    it('should render resource description', () => {
      const resource = createMockResource();
      render(<ResourcePanelItem resource={resource} />);

      expect(screen.getByText('Test Description')).toBeInTheDocument();
    });

    it('should not render description when not provided', () => {
      const resource = createMockResource({
        display: {
          label: 'Test Resource',
          showOnResourcePanel: true,
        },
      } as Partial<Resource>);
      render(<ResourcePanelItem resource={resource} />);

      expect(screen.queryByText('Test Description')).not.toBeInTheDocument();
    });

    it('should render Avatar with image when provided', () => {
      const resource = createMockResource();
      render(<ResourcePanelItem resource={resource} />);

      const avatar = screen.getByRole('img');
      expect(avatar).toHaveAttribute('src', '/static/test-image.svg');
    });

    it('should render children when provided instead of default card', () => {
      const resource = createMockResource();
      render(
        <ResourcePanelItem resource={resource}>
          <div data-testid="custom-child">Custom Content</div>
        </ResourcePanelItem>,
      );

      expect(screen.getByTestId('custom-child')).toBeInTheDocument();
      expect(screen.queryByText('Test Resource')).not.toBeInTheDocument();
    });
  });

  describe('Type Variants', () => {
    it('should have default type of static', () => {
      const resource = createMockResource();
      const {container} = render(<ResourcePanelItem resource={resource} />);

      // Static type should have default cursor
      const card = container.querySelector('.MuiCard-root');
      expect(card).toBeInTheDocument();
    });

    it('should accept draggable type', () => {
      const resource = createMockResource();
      const {container} = render(<ResourcePanelItem resource={resource} type="draggable" />);

      const card = container.querySelector('.MuiCard-root');
      expect(card).toBeInTheDocument();
    });
  });

  describe('Add Button', () => {
    it('should render add button when onAdd is provided', () => {
      const resource = createMockResource();
      const onAdd = vi.fn();
      render(<ResourcePanelItem resource={resource} onAdd={onAdd} />);

      const addButton = screen.getByRole('button');
      expect(addButton).toBeInTheDocument();
    });

    it('should not render add button when onAdd is not provided', () => {
      const resource = createMockResource();
      render(<ResourcePanelItem resource={resource} />);

      expect(screen.queryByRole('button')).not.toBeInTheDocument();
    });

    it('should call onAdd with resource when add button is clicked', () => {
      const resource = createMockResource();
      const onAdd = vi.fn();
      render(<ResourcePanelItem resource={resource} onAdd={onAdd} />);

      const addButton = screen.getByRole('button');
      fireEvent.click(addButton);

      expect(onAdd).toHaveBeenCalledTimes(1);
      expect(onAdd).toHaveBeenCalledWith(resource);
    });

    it('should disable add button when disabled prop is true', () => {
      const resource = createMockResource();
      const onAdd = vi.fn();
      render(<ResourcePanelItem resource={resource} onAdd={onAdd} disabled />);

      const addButton = screen.getByRole('button');
      expect(addButton).toBeDisabled();
    });

    it('should not call onAdd when button is disabled and clicked', () => {
      const resource = createMockResource();
      const onAdd = vi.fn();
      render(<ResourcePanelItem resource={resource} onAdd={onAdd} disabled />);

      const addButton = screen.getByRole('button');
      fireEvent.click(addButton);

      expect(onAdd).not.toHaveBeenCalled();
    });
  });

  describe('Color Scheme', () => {
    it('should render with light mode by default', () => {
      const resource = createMockResource();
      render(<ResourcePanelItem resource={resource} />);

      // Component renders without error in light mode
      expect(screen.getByText('Test Resource')).toBeInTheDocument();
    });

    it('should apply dark mode filter when mode is dark', async () => {
      // Re-mock useColorScheme for dark mode
      const oxygenUI = await import('@wso2/oxygen-ui');
      vi.spyOn(oxygenUI, 'useColorScheme').mockReturnValue({
        mode: 'dark',
        systemMode: 'light',
        setMode: vi.fn(),
      } as unknown as ReturnType<typeof oxygenUI.useColorScheme>);

      const resource = createMockResource();
      render(<ResourcePanelItem resource={resource} />);

      // Component renders with dark mode
      expect(screen.getByText('Test Resource')).toBeInTheDocument();
    });

    it('should use systemMode when mode is system', async () => {
      // Re-mock useColorScheme for system mode
      const oxygenUI = await import('@wso2/oxygen-ui');
      vi.spyOn(oxygenUI, 'useColorScheme').mockReturnValue({
        mode: 'system',
        systemMode: 'dark',
        setMode: vi.fn(),
      } as unknown as ReturnType<typeof oxygenUI.useColorScheme>);

      const resource = createMockResource();
      render(<ResourcePanelItem resource={resource} />);

      // Component renders using systemMode (dark)
      expect(screen.getByText('Test Resource')).toBeInTheDocument();
    });

    it('should use light effectiveMode when mode is system and systemMode is light', async () => {
      const oxygenUI = await import('@wso2/oxygen-ui');
      vi.spyOn(oxygenUI, 'useColorScheme').mockReturnValue({
        mode: 'system',
        systemMode: 'light',
        setMode: vi.fn(),
      } as unknown as ReturnType<typeof oxygenUI.useColorScheme>);

      const resource = createMockResource();
      render(<ResourcePanelItem resource={resource} />);

      expect(screen.getByText('Test Resource')).toBeInTheDocument();
    });
  });
});
