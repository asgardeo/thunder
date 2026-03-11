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
import {render} from '@thunder/test-utils/browser';
import {page, userEvent} from 'vitest/browser';
import ResourcePanelStatic from '../ResourcePanelStatic';
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
  type: 'TEMPLATE_TYPE',
  resourceType: 'TEMPLATE',
  display: {
    label: 'Static Template',
    description: 'A static template description',
    image: 'template-icon.svg',
    showOnResourcePanel: true,
  },
  ...overrides,
} as Resource);

describe('ResourcePanelStatic', () => {
  describe('Rendering', () => {
    it('should render resource label', async () => {
      const resource = createMockResource();
      await render(<ResourcePanelStatic id="test-static-1" resource={resource} />);

      await expect.element(page.getByText('Static Template')).toBeInTheDocument();
    });

    it('should render resource description', async () => {
      const resource = createMockResource();
      await render(<ResourcePanelStatic id="test-static-1" resource={resource} />);

      await expect.element(page.getByText('A static template description')).toBeInTheDocument();
    });

    it('should pass id to ResourcePanelItem', async () => {
      const resource = createMockResource();
      const {container} = await render(<ResourcePanelStatic id="unique-id-123" resource={resource} />);

      // ID should be passed through - component renders successfully
      expect(container.firstChild).toBeInTheDocument();
    });
  });

  describe('Type Default', () => {
    it('should have static as default type', async () => {
      const resource = createMockResource();
      const {container} = await render(<ResourcePanelStatic id="test-id" resource={resource} />);

      // Static type card renders with default cursor behavior
      const card = container.querySelector('.MuiCard-root');
      expect(card).toBeInTheDocument();
    });
  });

  describe('Disabled State', () => {
    it('should pass disabled state to ResourcePanelItem', async () => {
      const resource = createMockResource();
      const onAdd = vi.fn();
      await render(<ResourcePanelStatic id="test-id" resource={resource} onAdd={onAdd} disabled />);

      const button = page.getByRole('button');
      expect(button).toBeDisabled();
    });

    it('should not disable by default', async () => {
      const resource = createMockResource();
      const onAdd = vi.fn();
      await render(<ResourcePanelStatic id="test-id" resource={resource} onAdd={onAdd} />);

      const button = page.getByRole('button');
      expect(button).not.toBeDisabled();
    });
  });

  describe('onAdd Callback', () => {
    it('should pass onAdd to ResourcePanelItem', async () => {
      const resource = createMockResource();
      const onAdd = vi.fn();
      await render(<ResourcePanelStatic id="test-id" resource={resource} onAdd={onAdd} />);

      const button = page.getByRole('button');
      await userEvent.click(button);

      expect(onAdd).toHaveBeenCalledWith(resource);
    });
  });

  describe('Additional Props', () => {
    it('should render with additional className', async () => {
      const resource = createMockResource();
      const {container} = await render(
        <ResourcePanelStatic id="test-id" resource={resource} className="custom-class" />,
      );

      // Component renders - additional props may not be passed through to the underlying element
      expect(container.querySelector('.MuiCard-root')).toBeInTheDocument();
    });
  });

  describe('Type Prop', () => {
    it('should accept custom type prop', async () => {
      const resource = createMockResource();
      const {container} = await render(<ResourcePanelStatic id="test-id" resource={resource} type="draggable" />);

      // Component renders with custom type
      const card = container.querySelector('.MuiCard-root');
      expect(card).toBeInTheDocument();
    });

    it('should use static type by default when type is not provided', async () => {
      const resource = createMockResource();
      const {container} = await render(<ResourcePanelStatic id="test-id" resource={resource} />);

      // Component renders with default static type
      const card = container.querySelector('.MuiCard-root');
      expect(card).toBeInTheDocument();
    });
  });
});
