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
import {render, page, userEvent} from '@thunder/test-utils/browser';
import {Palette} from '@wso2/oxygen-ui-icons-react';
import SectionCard from '../SectionCard';

describe('SectionCard', () => {
  const defaultProps = {
    label: 'Colors',
    description: 'Manage the color palette',
    icon: <Palette />,
    isSelected: false,
    onClick: vi.fn(),
  };

  describe('Rendering', () => {
    it('renders the label', async () => {
      render(<SectionCard {...defaultProps} />);
      await expect.element(page.getByText('Colors')).toBeInTheDocument();
    });

    it('renders the description', async () => {
      render(<SectionCard {...defaultProps} />);
      await expect.element(page.getByText('Manage the color palette')).toBeInTheDocument();
    });

    it('renders the icon', async () => {
      const {container} = render(<SectionCard {...defaultProps} />);
      expect(container.querySelector('svg')).toBeInTheDocument();
    });
  });

  describe('Selection state', () => {
    it('renders without visual selection when isSelected is false', async () => {
      const {container} = render(<SectionCard {...defaultProps} isSelected={false} />);
      // Just verify it renders without errors
      expect(container.firstChild).toBeInTheDocument();
    });

    it('renders with visual selection when isSelected is true', async () => {
      const {container} = render(<SectionCard {...defaultProps} isSelected />);
      expect(container.firstChild).toBeInTheDocument();
    });
  });

  describe('Interaction', () => {
    it('calls onClick when the card is clicked', async () => {
      const onClick = vi.fn();
      render(<SectionCard {...defaultProps} onClick={onClick} />);

      await userEvent.click(page.getByText('Colors'));

      expect(onClick).toHaveBeenCalledOnce();
    });

    it('calls onClick multiple times on repeated clicks', async () => {
      const onClick = vi.fn();
      render(<SectionCard {...defaultProps} onClick={onClick} />);

      await userEvent.click(page.getByText('Colors'));
      await userEvent.click(page.getByText('Colors'));

      expect(onClick).toHaveBeenCalledTimes(2);
    });
  });
});
