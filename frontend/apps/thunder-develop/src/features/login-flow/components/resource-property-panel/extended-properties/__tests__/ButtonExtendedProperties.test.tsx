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
import {page, userEvent} from 'vitest/browser';
import type {Resource} from '@/features/flows/models/resources';
import ButtonExtendedProperties from '../ButtonExtendedProperties';

describe('ButtonExtendedProperties', () => {
  const mockOnChange = vi.fn();

  const createMockResource = (overrides: Partial<Resource> = {}): Resource =>
    ({
      id: 'button-1',
      type: 'ACTION',
      category: 'ACTION',
      resourceType: 'ELEMENT',
      ...overrides,
    }) as Resource;

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render the start icon label', async () => {
      const resource = createMockResource();

      await render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      await expect.element(page.getByText('Start Icon')).toBeInTheDocument();
    });

    it('should render the end icon label', async () => {
      const resource = createMockResource();

      await render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      await expect.element(page.getByText('End Icon')).toBeInTheDocument();
    });

    it('should render start icon input field', async () => {
      const resource = createMockResource();

      await render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const startIconInput = page.getByPlaceholder('Enter icon path (e.g., assets/images/icons/icon.svg)');
      expect(startIconInput).toBeInTheDocument();
    });

    it('should render end icon input field', async () => {
      const resource = createMockResource();

      await render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const endIconInput = page.getByPlaceholder('Enter icon path (e.g., assets/images/icons/icon.svg)');
      expect(endIconInput).toBeInTheDocument();
    });

    it('should render hint text for start icon', async () => {
      const resource = createMockResource();

      await render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      await expect.element(page.getByText('Optional icon displayed before the button label')).toBeInTheDocument();
    });

    it('should render hint text for end icon', async () => {
      const resource = createMockResource();

      await render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      await expect.element(page.getByText('Optional icon displayed after the button label')).toBeInTheDocument();
    });

    it('should render dividers', async () => {
      const resource = createMockResource();

      const {container} = await render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const dividers = container.querySelectorAll('.MuiDivider-root');
      expect(dividers.length).toBe(2);
    });
  });

  describe('Initial Values', () => {
    it('should display existing startIcon value', async () => {
      const resource = createMockResource({
        startIcon: '/assets/icons/test-start.svg',
      } as Partial<Resource>);

      await render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const startIconInput = page.getByPlaceholder(
        'Enter icon path (e.g., assets/images/icons/icon.svg)',
      );
      expect((startIconInput.element() as HTMLInputElement).value).toBe('/assets/icons/test-start.svg');
    });

    it('should display existing endIcon value', async () => {
      const resource = createMockResource({
        endIcon: '/assets/icons/test-end.svg',
      } as Partial<Resource>);

      await render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const endIconInput = page.getByPlaceholder(
        'Enter icon path (e.g., assets/images/icons/icon.svg)',
      );
      expect((endIconInput.element() as HTMLInputElement).value).toBe('/assets/icons/test-end.svg');
    });

    it('should display empty value when startIcon is not set', async () => {
      const resource = createMockResource();

      await render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const startIconInput = page.getByPlaceholder(
        'Enter icon path (e.g., assets/images/icons/icon.svg)',
      );
      expect((startIconInput.element() as HTMLInputElement).value).toBe('');
    });

    it('should display empty value when endIcon is not set', async () => {
      const resource = createMockResource();

      await render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const endIconInput = page.getByPlaceholder(
        'Enter icon path (e.g., assets/images/icons/icon.svg)',
      );
      expect((endIconInput.element() as HTMLInputElement).value).toBe('');
    });
  });

  describe('Change Handlers', () => {
    it('should call onChange when start icon value changes', async () => {
      const resource = createMockResource();

      await render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const startIconInput = page.getByPlaceholder('Enter icon path (e.g., assets/images/icons/icon.svg)');
      await userEvent.fill(startIconInput, '/new/icon/path.svg');

      expect(mockOnChange).toHaveBeenCalledWith('startIcon', '/new/icon/path.svg', resource);
    });

    it('should call onChange when end icon value changes', async () => {
      const resource = createMockResource();

      await render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const endIconInput = page.getByPlaceholder('Enter icon path (e.g., assets/images/icons/icon.svg)');
      await userEvent.fill(endIconInput, '/new/end/icon.svg');

      expect(mockOnChange).toHaveBeenCalledWith('endIcon', '/new/end/icon.svg', resource);
    });

    it('should call onChange with empty string when clearing start icon', async () => {
      const resource = createMockResource({
        startIcon: '/existing/icon.svg',
      } as Partial<Resource>);

      await render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const startIconInput = page.getByPlaceholder('Enter icon path (e.g., assets/images/icons/icon.svg)');
      await userEvent.fill(startIconInput, '');

      expect(mockOnChange).toHaveBeenCalledWith('startIcon', '', resource);
    });

    it('should call onChange with empty string when clearing end icon', async () => {
      const resource = createMockResource({
        endIcon: '/existing/icon.svg',
      } as Partial<Resource>);

      await render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const endIconInput = page.getByPlaceholder('Enter icon path (e.g., assets/images/icons/icon.svg)');
      await userEvent.fill(endIconInput, '');

      expect(mockOnChange).toHaveBeenCalledWith('endIcon', '', resource);
    });
  });

  describe('Input Attributes', () => {
    it('should have correct id for start icon input', async () => {
      const resource = createMockResource();

      await render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const startIconInput = page.getByPlaceholder('Enter icon path (e.g., assets/images/icons/icon.svg)');
      expect(startIconInput).toHaveAttribute('id', 'start-icon-input');
    });

    it('should have correct id for end icon input', async () => {
      const resource = createMockResource();

      await render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const endIconInput = page.getByPlaceholder('Enter icon path (e.g., assets/images/icons/icon.svg)');
      expect(endIconInput).toHaveAttribute('id', 'end-icon-input');
    });
  });
});
