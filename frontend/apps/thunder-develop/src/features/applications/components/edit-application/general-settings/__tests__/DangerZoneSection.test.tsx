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

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {render} from '@thunder/test-utils/browser';
import {page, userEvent} from 'vitest/browser';
import DangerZoneSection from '../DangerZoneSection';

describe('DangerZoneSection', () => {
  const mockOnRegenerateClick = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render the danger zone section', async () => {
    await render(<DangerZoneSection onRegenerateClick={mockOnRegenerateClick} />);

    await expect.element(page.getByText('Danger Zone')).toBeInTheDocument();
    await expect.element(page.getByText('Actions in this section are irreversible. Proceed with caution.')).toBeInTheDocument();
  });

  it('should render revoke application title', async () => {
    await render(<DangerZoneSection onRegenerateClick={mockOnRegenerateClick} />);

    await expect.element(page.getByRole('heading', {name: 'Regenerate Client Secret', level: 6})).toBeInTheDocument();
  });

  it('should render warning description', async () => {
    await render(<DangerZoneSection onRegenerateClick={mockOnRegenerateClick} />);

    await expect.element(
      page.getByText('Regenerating the client secret will immediately invalidate the current client secret and cannot be undone.'),
    ).toBeInTheDocument();
  });

  it('should render revoke button', async () => {
    await render(<DangerZoneSection onRegenerateClick={mockOnRegenerateClick} />);

    await expect.element(page.getByRole('button', {name: 'Regenerate Client Secret'})).toBeInTheDocument();
  });

  it('should call onRegenerateClick when revoke button is clicked', async () => {
    await render(<DangerZoneSection onRegenerateClick={mockOnRegenerateClick} />);

    await userEvent.click(page.getByRole('button', {name: 'Regenerate Client Secret'}));

    expect(mockOnRegenerateClick).toHaveBeenCalledTimes(1);
  });

  it('should call onRegenerateClick multiple times when clicked multiple times', async () => {
    await render(<DangerZoneSection onRegenerateClick={mockOnRegenerateClick} />);

    const regenerateButton = page.getByRole('button', {name: 'Regenerate Client Secret'});
    await userEvent.click(regenerateButton);
    await userEvent.click(regenerateButton);
    await userEvent.click(regenerateButton);

    expect(mockOnRegenerateClick).toHaveBeenCalledTimes(3);
  });

  it('should render revoke button with error color', async () => {
    await render(<DangerZoneSection onRegenerateClick={mockOnRegenerateClick} />);

    const regenerateButton = page.getByRole('button', {name: 'Regenerate Client Secret'});
    expect(regenerateButton.element()).toHaveClass('MuiButton-colorError');
  });
});
