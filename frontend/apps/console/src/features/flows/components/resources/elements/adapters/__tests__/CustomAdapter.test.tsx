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

import {render, screen} from '@testing-library/react';
import {describe, it, expect, vi} from 'vitest';
import CustomAdapter from '../CustomAdapter';
import type {Element as FlowElement} from '@/features/flows/models/elements';

vi.mock('@wso2/oxygen-ui-icons-react', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@wso2/oxygen-ui-icons-react')>();
  return {
    ...actual,
    PuzzleIcon: ({size = 24}: {size?: number} = {}) => <svg data-testid="puzzle-icon" data-size={size} />,
  };
});

describe('CustomAdapter', () => {
  const createMockElement = (overrides: Partial<FlowElement> = {}): FlowElement =>
    ({
      id: 'custom-1',
      type: 'CUSTOM',
      category: 'MISCELLANEOUS',
      ...overrides,
    }) as FlowElement;

  it('should render the puzzle icon', () => {
    render(<CustomAdapter resource={createMockElement()} />);

    expect(screen.getByTestId('puzzle-icon')).toBeInTheDocument();
  });

  it('should render the "Custom" label', () => {
    render(<CustomAdapter resource={createMockElement()} />);

    expect(screen.getByText('Custom')).toBeInTheDocument();
  });

  it('should display the resource identifier', () => {
    render(<CustomAdapter resource={createMockElement({id: 'my-custom-element'})} />);

    expect(screen.getByText(/my-custom-element/)).toBeInTheDocument();
  });

  it('should pass size 20 to the puzzle icon', () => {
    render(<CustomAdapter resource={createMockElement()} />);

    expect(screen.getByTestId('puzzle-icon')).toHaveAttribute('data-size', '20');
  });
});
