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
import {render, screen, userEvent} from '@thunder/test-utils';
import {Bell} from '@wso2/oxygen-ui-icons-react';
import MenuButton from '../MenuButton';

describe('MenuButton', () => {
  it('renders an icon button', () => {
    render(
      <MenuButton aria-label="test button">
        <Bell />
      </MenuButton>,
    );

    const button = screen.getByRole('button', {name: /test button/i});
    expect(button).toBeInTheDocument();
  });

  it('does not show badge by default', () => {
    const {container} = render(
      <MenuButton aria-label="test button">
        <Bell />
      </MenuButton>,
    );

    const badge = container.querySelector('.MuiBadge-badge');
    expect(badge).toHaveClass('MuiBadge-invisible');
  });

  it('shows badge when showBadge is true', () => {
    const {container} = render(
      <MenuButton showBadge aria-label="test button">
        <Bell />
      </MenuButton>,
    );

    const badge = container.querySelector('.MuiBadge-badge');
    expect(badge).toBeInTheDocument();
    expect(badge).not.toHaveClass('MuiBadge-invisible');
  });

  it('calls onClick handler when clicked', async () => {
    const user = userEvent.setup();
    const handleClick = vi.fn();

    render(
      <MenuButton onClick={handleClick} aria-label="test button">
        <Bell />
      </MenuButton>,
    );

    const button = screen.getByRole('button', {name: /test button/i});
    await user.click(button);

    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  it('is disabled when disabled prop is true', () => {
    render(
      <MenuButton disabled aria-label="test button">
        <Bell />
      </MenuButton>,
    );

    const button = screen.getByRole('button', {name: /test button/i});
    expect(button).toBeDisabled();
  });

  it('passes through additional IconButton props', () => {
    render(
      <MenuButton aria-label="test button" data-testid="custom-button">
        <Bell />
      </MenuButton>,
    );

    const button = screen.getByTestId('custom-button');
    expect(button).toBeInTheDocument();
  });
});
