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

import {describe, it, expect, vi} from 'vitest';
import {render, screen} from '@thunder/test-utils';
import SliderRow from '../SliderRow';

describe('SliderRow', () => {
  describe('Rendering', () => {
    it('renders the label', () => {
      render(<SliderRow label="Border Radius" value={8} min={0} max={24} onChange={vi.fn()} />);
      expect(screen.getByText('Border Radius')).toBeInTheDocument();
    });

    it('renders the current value with default unit "px"', () => {
      render(<SliderRow label="Spacing" value={16} min={0} max={48} onChange={vi.fn()} />);
      expect(screen.getByText('16px')).toBeInTheDocument();
    });

    it('renders the current value with a custom unit', () => {
      render(<SliderRow label="Opacity" value={50} min={0} max={100} unit="%" onChange={vi.fn()} />);
      expect(screen.getByText('50%')).toBeInTheDocument();
    });

    it('renders the value 0 correctly', () => {
      render(<SliderRow label="Radius" value={0} min={0} max={24} onChange={vi.fn()} />);
      expect(screen.getByText('0px')).toBeInTheDocument();
    });

    it('renders a slider input element', () => {
      render(<SliderRow label="Radius" value={8} min={0} max={24} onChange={vi.fn()} />);
      expect(screen.getByRole('slider')).toBeInTheDocument();
    });

    it('slider reflects the min attribute', () => {
      render(<SliderRow label="Radius" value={4} min={2} max={24} onChange={vi.fn()} />);
      expect(screen.getByRole('slider')).toHaveAttribute('aria-valuemin', '2');
    });

    it('slider reflects the max attribute', () => {
      render(<SliderRow label="Radius" value={4} min={0} max={32} onChange={vi.fn()} />);
      expect(screen.getByRole('slider')).toHaveAttribute('aria-valuemax', '32');
    });

    it('slider reflects the current value', () => {
      render(<SliderRow label="Radius" value={12} min={0} max={24} onChange={vi.fn()} />);
      expect(screen.getByRole('slider')).toHaveAttribute('aria-valuenow', '12');
    });
  });
});
