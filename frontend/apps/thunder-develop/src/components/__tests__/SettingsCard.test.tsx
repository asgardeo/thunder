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
import {render} from '@thunder/test-utils/browser';
import {page, userEvent} from 'vitest/browser';
import SettingsCard from '../SettingsCard';

describe('SettingsCard', () => {
  describe('Rendering', () => {
    it('should render with title and children', async () => {
      await render(
        <SettingsCard title="Test Settings">
          <div>Test Content</div>
        </SettingsCard>,
      );

      await expect.element(page.getByText('Test Settings')).toBeInTheDocument();
      await expect.element(page.getByText('Test Content')).toBeInTheDocument();
    });

    it('should render with title and description', async () => {
      await render(
        <SettingsCard title="Test Settings" description="This is a description">
          <div>Content</div>
        </SettingsCard>,
      );

      await expect.element(page.getByText('Test Settings')).toBeInTheDocument();
      await expect.element(page.getByText('This is a description')).toBeInTheDocument();
    });

    it('should not render description when not provided', async () => {
      await render(
        <SettingsCard title="Test Settings">
          <div>Content</div>
        </SettingsCard>,
      );

      await expect.element(page.getByText('Test Settings')).toBeInTheDocument();
      await expect.element(page.getByText('This is a description')).not.toBeInTheDocument();
    });

    it('should not render toggle switch by default', async () => {
      await render(
        <SettingsCard title="Test Settings">
          <div>Content</div>
        </SettingsCard>,
      );

      await expect.element(page.getByRole('switch')).not.toBeInTheDocument();
    });

    it('should render toggle switch when enabled and onToggle are provided', async () => {
      const mockOnToggle = vi.fn();
      await render(
        <SettingsCard title="Test Settings" enabled onToggle={mockOnToggle}>
          <div>Content</div>
        </SettingsCard>,
      );

      const toggleSwitch = page.getByRole('switch');
      await expect.element(toggleSwitch).toBeInTheDocument();
      await expect.element(toggleSwitch).toBeChecked();
    });

    it('should render toggle switch as unchecked when enabled is false', async () => {
      const mockOnToggle = vi.fn();
      await render(
        <SettingsCard title="Test Settings" enabled={false} onToggle={mockOnToggle}>
          <div>Content</div>
        </SettingsCard>,
      );

      const toggleSwitch = page.getByRole('switch');
      await expect.element(toggleSwitch).not.toBeChecked();
    });

    it('should show children when toggle is enabled', async () => {
      const mockOnToggle = vi.fn();
      await render(
        <SettingsCard title="Test Settings" enabled onToggle={mockOnToggle}>
          <div>Visible Content</div>
        </SettingsCard>,
      );

      await expect.element(page.getByText('Visible Content')).toBeInTheDocument();
    });

    it('should hide children when toggle is disabled', async () => {
      const mockOnToggle = vi.fn();
      await render(
        <SettingsCard title="Test Settings" enabled={false} onToggle={mockOnToggle}>
          <div>Hidden Content</div>
        </SettingsCard>,
      );

      await expect.element(page.getByText('Hidden Content')).not.toBeInTheDocument();
    });

    it('should show children when no toggle is provided', async () => {
      await render(
        <SettingsCard title="Test Settings">
          <div>Always Visible</div>
        </SettingsCard>,
      );

      await expect.element(page.getByText('Always Visible')).toBeInTheDocument();
    });
  });

  describe('User Interactions', () => {
    it('should call onToggle when switch is clicked', async () => {
      const mockOnToggle = vi.fn();

      await render(
        <SettingsCard title="Test Settings" enabled={false} onToggle={mockOnToggle}>
          <div>Content</div>
        </SettingsCard>,
      );

      await userEvent.click(page.getByRole('switch'));

      await vi.waitFor(() => {
        expect(mockOnToggle).toHaveBeenCalledWith(true);
        expect(mockOnToggle).toHaveBeenCalledTimes(1);
      });
    });

    it('should call onToggle with false when toggling off', async () => {
      const mockOnToggle = vi.fn();

      await render(
        <SettingsCard title="Test Settings" enabled onToggle={mockOnToggle}>
          <div>Content</div>
        </SettingsCard>,
      );

      await userEvent.click(page.getByRole('switch'));

      await vi.waitFor(() => {
        expect(mockOnToggle).toHaveBeenCalledWith(false);
        expect(mockOnToggle).toHaveBeenCalledTimes(1);
      });
    });
  });

  describe('Accessibility', () => {
    it('should render toggle switch with proper structure', async () => {
      const mockOnToggle = vi.fn();

      await render(
        <SettingsCard title="Registration Flow" enabled onToggle={mockOnToggle}>
          <div>Content</div>
        </SettingsCard>,
      );

      // Verify the switch element exists and is accessible
      const toggleSwitch = page.getByRole('switch');
      await expect.element(toggleSwitch).toBeInTheDocument();
      await expect.element(toggleSwitch).toHaveAttribute('type', 'checkbox');
    });
  });

  describe('Edge Cases', () => {
    it('should handle only enabled prop without onToggle', async () => {
      await render(
        <SettingsCard title="Test Settings" enabled>
          <div>Content</div>
        </SettingsCard>,
      );

      await expect.element(page.getByRole('switch')).not.toBeInTheDocument();
      await expect.element(page.getByText('Content')).toBeInTheDocument();
    });

    it('should handle only onToggle prop without enabled', async () => {
      const mockOnToggle = vi.fn();

      await render(
        <SettingsCard title="Test Settings" onToggle={mockOnToggle}>
          <div>Content</div>
        </SettingsCard>,
      );

      await expect.element(page.getByRole('switch')).not.toBeInTheDocument();
      await expect.element(page.getByText('Content')).toBeInTheDocument();
    });

    it('should render complex children elements', async () => {
      await render(
        <SettingsCard title="Complex Settings">
          <div>
            <input type="text" placeholder="Username" />
            <button type="button">Submit</button>
          </div>
        </SettingsCard>,
      );

      await expect.element(page.getByPlaceholder('Username')).toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: 'Submit'})).toBeInTheDocument();
    });
  });

  it('should render headerAction if provided', async () => {
    await render(
      <SettingsCard title="Test Settings" headerAction={<button type="button">Action</button>}>
        <div>Content</div>
      </SettingsCard>,
    );

    await expect.element(page.getByText('Action')).toBeInTheDocument();
  });

  it('should not render content wrapper if valid children are not present', async () => {
    const {container} = await render(<SettingsCard title="Test Settings">{false && <div>Content</div>}</SettingsCard>);

    // Should only have the outer paper, title box, but NO inner content paper
    // The outer paper renders the title box div + potential content div
    // We expect only the title box div to be present inside the outer Paper
    const papers = container.querySelectorAll('.MuiPaper-root');
    expect(papers.length).toBe(1); // Only the outer card, no inner content card
  });

  it('should render content wrapper if valid children are present', async () => {
    const {container} = await render(
      <SettingsCard title="Test Settings">
        <div>Content</div>
      </SettingsCard>,
    );

    const papers = container.querySelectorAll('.MuiPaper-root');
    expect(papers.length).toBe(2); // Outer card + Inner content card
  });
});
