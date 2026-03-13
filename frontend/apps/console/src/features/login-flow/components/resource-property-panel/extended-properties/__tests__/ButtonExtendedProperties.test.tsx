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
import type {Resource} from '@/features/flows/models/resources';
import ButtonExtendedProperties from '../ButtonExtendedProperties';

// Mock dependencies
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

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
    it('should render the start icon label', () => {
      const resource = createMockResource();

      render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      expect(screen.getByText('flows:core.buttonExtendedProperties.startIcon.label')).toBeInTheDocument();
    });

    it('should render the end icon label', () => {
      const resource = createMockResource();

      render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      expect(screen.getByText('flows:core.buttonExtendedProperties.endIcon.label')).toBeInTheDocument();
    });

    it('should render start icon input field', () => {
      const resource = createMockResource();

      render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const startIconInput = screen.getByPlaceholderText('flows:core.buttonExtendedProperties.startIcon.placeholder');
      expect(startIconInput).toBeInTheDocument();
    });

    it('should render end icon input field', () => {
      const resource = createMockResource();

      render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const endIconInput = screen.getByPlaceholderText('flows:core.buttonExtendedProperties.endIcon.placeholder');
      expect(endIconInput).toBeInTheDocument();
    });

    it('should render hint text for start icon', () => {
      const resource = createMockResource();

      render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      expect(screen.getByText('flows:core.buttonExtendedProperties.startIcon.hint')).toBeInTheDocument();
    });

    it('should render hint text for end icon', () => {
      const resource = createMockResource();

      render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      expect(screen.getByText('flows:core.buttonExtendedProperties.endIcon.hint')).toBeInTheDocument();
    });

    it('should render dividers', () => {
      const resource = createMockResource();

      const {container} = render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const dividers = container.querySelectorAll('.MuiDivider-root');
      expect(dividers.length).toBe(2);
    });
  });

  describe('Initial Values', () => {
    it('should display existing startIcon value', () => {
      const resource = createMockResource({
        startIcon: '/assets/icons/test-start.svg',
      } as Partial<Resource>);

      render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const startIconInput = screen.getByPlaceholderText<HTMLInputElement>(
        'flows:core.buttonExtendedProperties.startIcon.placeholder',
      );
      expect(startIconInput.value).toBe('/assets/icons/test-start.svg');
    });

    it('should display existing endIcon value', () => {
      const resource = createMockResource({
        endIcon: '/assets/icons/test-end.svg',
      } as Partial<Resource>);

      render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const endIconInput = screen.getByPlaceholderText<HTMLInputElement>(
        'flows:core.buttonExtendedProperties.endIcon.placeholder',
      );
      expect(endIconInput.value).toBe('/assets/icons/test-end.svg');
    });

    it('should display empty value when startIcon is not set', () => {
      const resource = createMockResource();

      render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const startIconInput = screen.getByPlaceholderText<HTMLInputElement>(
        'flows:core.buttonExtendedProperties.startIcon.placeholder',
      );
      expect(startIconInput.value).toBe('');
    });

    it('should display empty value when endIcon is not set', () => {
      const resource = createMockResource();

      render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const endIconInput = screen.getByPlaceholderText<HTMLInputElement>(
        'flows:core.buttonExtendedProperties.endIcon.placeholder',
      );
      expect(endIconInput.value).toBe('');
    });
  });

  describe('Change Handlers', () => {
    it('should call onChange when start icon value changes', () => {
      const resource = createMockResource();

      render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const startIconInput = screen.getByPlaceholderText('flows:core.buttonExtendedProperties.startIcon.placeholder');
      fireEvent.change(startIconInput, {target: {value: '/new/icon/path.svg'}});

      expect(mockOnChange).toHaveBeenCalledWith('startIcon', '/new/icon/path.svg', resource);
    });

    it('should call onChange when end icon value changes', () => {
      const resource = createMockResource();

      render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const endIconInput = screen.getByPlaceholderText('flows:core.buttonExtendedProperties.endIcon.placeholder');
      fireEvent.change(endIconInput, {target: {value: '/new/end/icon.svg'}});

      expect(mockOnChange).toHaveBeenCalledWith('endIcon', '/new/end/icon.svg', resource);
    });

    it('should call onChange with empty string when clearing start icon', () => {
      const resource = createMockResource({
        startIcon: '/existing/icon.svg',
      } as Partial<Resource>);

      render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const startIconInput = screen.getByPlaceholderText('flows:core.buttonExtendedProperties.startIcon.placeholder');
      fireEvent.change(startIconInput, {target: {value: ''}});

      expect(mockOnChange).toHaveBeenCalledWith('startIcon', '', resource);
    });

    it('should call onChange with empty string when clearing end icon', () => {
      const resource = createMockResource({
        endIcon: '/existing/icon.svg',
      } as Partial<Resource>);

      render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const endIconInput = screen.getByPlaceholderText('flows:core.buttonExtendedProperties.endIcon.placeholder');
      fireEvent.change(endIconInput, {target: {value: ''}});

      expect(mockOnChange).toHaveBeenCalledWith('endIcon', '', resource);
    });
  });

  describe('Input Attributes', () => {
    it('should have correct id for start icon input', () => {
      const resource = createMockResource();

      render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const startIconInput = screen.getByPlaceholderText('flows:core.buttonExtendedProperties.startIcon.placeholder');
      expect(startIconInput).toHaveAttribute('id', 'start-icon-input');
    });

    it('should have correct id for end icon input', () => {
      const resource = createMockResource();

      render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const endIconInput = screen.getByPlaceholderText('flows:core.buttonExtendedProperties.endIcon.placeholder');
      expect(endIconInput).toHaveAttribute('id', 'end-icon-input');
    });
  });
});
