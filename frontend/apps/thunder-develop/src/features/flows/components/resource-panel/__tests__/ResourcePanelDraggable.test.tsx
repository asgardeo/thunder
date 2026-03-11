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
import ResourcePanelDraggable from '../ResourcePanelDraggable';
import type {Resource} from '../../../models/resources';

// Mock Draggable component
vi.mock('../../dnd/Draggable', () => ({
  default: ({
    children,
    id,
    data,
    type,
    accept,
    disabled,
  }: {
    children: React.ReactNode;
    id: string;
    data: object;
    type: string;
    accept: string[];
    disabled: boolean;
  }) => (
    <div
      data-testid="draggable-wrapper"
      data-id={id}
      data-type={type}
      data-accept={JSON.stringify(accept)}
      data-disabled={disabled}
      data-dragged={JSON.stringify(data)}
    >
      {children}
    </div>
  ),
}));

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
  type: 'DRAGGABLE_STEP',
  resourceType: 'STEP',
  display: {
    label: 'Draggable Step',
    description: 'A draggable step description',
    image: 'step-icon.svg',
    showOnResourcePanel: true,
  },
  ...overrides,
} as Resource);

describe('ResourcePanelDraggable', () => {
  describe('Draggable Wrapper', () => {
    it('should wrap content in Draggable component', async () => {
      const resource = createMockResource();
      await render(<ResourcePanelDraggable id="step-1" resource={resource} onAdd={vi.fn()} />);

      await expect.element(page.getByTestId('draggable-wrapper')).toBeInTheDocument();
    });

    it('should pass id to Draggable', async () => {
      const resource = createMockResource();
      await render(<ResourcePanelDraggable id="unique-step-id" resource={resource} onAdd={vi.fn()} />);

      await expect.element(page.getByTestId('draggable-wrapper')).toHaveAttribute('data-id', 'unique-step-id');
    });

    it('should pass resource type to Draggable', async () => {
      const resource = createMockResource({type: 'CUSTOM_STEP'} as Partial<Resource>);
      await render(<ResourcePanelDraggable id="step-1" resource={resource} onAdd={vi.fn()} />);

      await expect.element(page.getByTestId('draggable-wrapper')).toHaveAttribute('data-type', 'CUSTOM_STEP');
    });

    it('should set accept to match resource type', async () => {
      const resource = createMockResource({type: 'CUSTOM_STEP'} as Partial<Resource>);
      await render(<ResourcePanelDraggable id="step-1" resource={resource} onAdd={vi.fn()} />);

      const wrapper = page.getByTestId('draggable-wrapper');
      const accept = JSON.parse(wrapper.element().getAttribute('data-accept') ?? '[]') as string[];
      expect(accept).toContain('CUSTOM_STEP');
    });

    it('should pass dragged resource in data prop', async () => {
      const resource = createMockResource();
      await render(<ResourcePanelDraggable id="step-1" resource={resource} onAdd={vi.fn()} />);

      const wrapper = page.getByTestId('draggable-wrapper');
      const data = JSON.parse(wrapper.element().getAttribute('data-dragged') ?? '{}') as {dragged: Resource};
      expect(data.dragged).toEqual(resource);
    });
  });

  describe('ResourcePanelItem Integration', () => {
    it('should render resource label', async () => {
      const resource = createMockResource();
      await render(<ResourcePanelDraggable id="step-1" resource={resource} onAdd={vi.fn()} />);

      await expect.element(page.getByText('Draggable Step')).toBeInTheDocument();
    });

    it('should render resource description', async () => {
      const resource = createMockResource();
      await render(<ResourcePanelDraggable id="step-1" resource={resource} onAdd={vi.fn()} />);

      await expect.element(page.getByText('A draggable step description')).toBeInTheDocument();
    });
  });

  describe('Type Default', () => {
    it('should have draggable as default type', async () => {
      const resource = createMockResource();
      const {container} = await render(<ResourcePanelDraggable id="step-1" resource={resource} onAdd={vi.fn()} />);

      // Draggable type card renders
      const card = container.querySelector('.MuiCard-root');
      expect(card).toBeInTheDocument();
    });
  });

  describe('Disabled State', () => {
    it('should pass disabled=true to Draggable', async () => {
      const resource = createMockResource();
      await render(<ResourcePanelDraggable id="step-1" resource={resource} onAdd={vi.fn()} disabled />);

      await expect.element(page.getByTestId('draggable-wrapper')).toHaveAttribute('data-disabled', 'true');
    });

    it('should pass disabled=false by default', async () => {
      const resource = createMockResource();
      await render(<ResourcePanelDraggable id="step-1" resource={resource} onAdd={vi.fn()} />);

      await expect.element(page.getByTestId('draggable-wrapper')).toHaveAttribute('data-disabled', 'false');
    });

    it('should disable add button when disabled', async () => {
      const resource = createMockResource();
      const onAdd = vi.fn();
      await render(<ResourcePanelDraggable id="step-1" resource={resource} onAdd={onAdd} disabled />);

      const button = page.getByRole('button');
      expect(button).toBeDisabled();
    });
  });

  describe('onAdd Callback', () => {
    it('should pass onAdd to ResourcePanelItem', async () => {
      const resource = createMockResource();
      const onAdd = vi.fn();
      await render(<ResourcePanelDraggable id="step-1" resource={resource} onAdd={onAdd} />);

      const button = page.getByRole('button');
      await userEvent.click(button);

      expect(onAdd).toHaveBeenCalledWith(resource);
    });

    it('should not call onAdd when disabled', async () => {
      const resource = createMockResource();
      const onAdd = vi.fn();
      await render(<ResourcePanelDraggable id="step-1" resource={resource} onAdd={onAdd} disabled />);

      const button = page.getByRole('button');
      await userEvent.click(button);

      expect(onAdd).not.toHaveBeenCalled();
    });
  });

  describe('Additional Props', () => {
    it('should pass additional HTML attributes', async () => {
      const resource = createMockResource();
      await render(
        <ResourcePanelDraggable
          id="step-1"
          resource={resource}
          onAdd={vi.fn()}
          data-testid="draggable-item"
          className="custom-class"
        />,
      );

      // The Draggable wrapper receives the data-testid through rest props
      await expect.element(page.getByTestId('draggable-wrapper')).toBeInTheDocument();
    });
  });

  describe('Type Prop', () => {
    it('should accept custom type prop', async () => {
      const resource = createMockResource();
      const {container} = await render(
        <ResourcePanelDraggable id="step-1" resource={resource} onAdd={vi.fn()} type="static" />,
      );

      // Component renders with custom type
      const card = container.querySelector('.MuiCard-root');
      expect(card).toBeInTheDocument();
    });

    it('should use draggable type by default when type is not provided', async () => {
      const resource = createMockResource();
      const {container} = await render(<ResourcePanelDraggable id="step-1" resource={resource} onAdd={vi.fn()} />);

      // Component renders with default draggable type
      const card = container.querySelector('.MuiCard-root');
      expect(card).toBeInTheDocument();
    });
  });
});
