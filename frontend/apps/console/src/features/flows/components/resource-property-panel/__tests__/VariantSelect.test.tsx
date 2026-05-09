/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

import {render, screen, fireEvent} from '@testing-library/react';
import {describe, it, expect, vi} from 'vitest';
import type {Element} from '../../../models/elements';
import type {Resource} from '../../../models/resources';
import VariantSelect from '../VariantSelect';

describe('VariantSelect', () => {
  const createResource = (variants?: Element[]): Resource =>
    ({
      id: 'resource-1',
      type: 'ACTION',
      category: 'ACTION',
      variants,
    }) as unknown as Resource;

  const mockVariants: Element[] = [
    {variant: 'PRIMARY'} as unknown as Element,
    {variant: 'SECONDARY'} as unknown as Element,
    {variant: 'TEXT'} as unknown as Element,
  ];

  it('should return null when resource has no variants', () => {
    const {container} = render(
      <VariantSelect resource={createResource()} selectedVariant={undefined} onVariantChange={vi.fn()} />,
    );

    expect(container.firstChild).toBeNull();
  });

  it('should return null when variants array is empty', () => {
    const {container} = render(
      <VariantSelect resource={createResource([])} selectedVariant={undefined} onVariantChange={vi.fn()} />,
    );

    expect(container.firstChild).toBeNull();
  });

  it('should render a select with variant options', () => {
    render(
      <VariantSelect
        resource={createResource(mockVariants)}
        selectedVariant={mockVariants[0]}
        onVariantChange={vi.fn()}
      />,
    );

    expect(screen.getByText('Variant')).toBeInTheDocument();
    expect(screen.getByRole('combobox')).toBeInTheDocument();
  });

  it('should display the selected variant value', () => {
    render(
      <VariantSelect
        resource={createResource(mockVariants)}
        selectedVariant={mockVariants[1]}
        onVariantChange={vi.fn()}
      />,
    );

    expect(screen.getByRole('combobox')).toHaveTextContent('SECONDARY');
  });

  it('should call onVariantChange when a variant is selected', () => {
    const mockOnVariantChange = vi.fn();

    render(
      <VariantSelect
        resource={createResource(mockVariants)}
        selectedVariant={mockVariants[0]}
        onVariantChange={mockOnVariantChange}
      />,
    );

    // Open the select
    fireEvent.mouseDown(screen.getByRole('combobox'));
    // Click SECONDARY option
    fireEvent.click(screen.getByText('SECONDARY'));

    expect(mockOnVariantChange).toHaveBeenCalledWith('SECONDARY');
  });
});
