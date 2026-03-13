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
import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import SettingsCard from '../SettingsCard';

describe('SettingsCard', () => {
  describe('Rendering', () => {
    it('should render with title and children', () => {
      render(
        <SettingsCard title="Test Settings">
          <div>Test Content</div>
        </SettingsCard>,
      );

      expect(screen.getByText('Test Settings')).toBeInTheDocument();
      expect(screen.getByText('Test Content')).toBeInTheDocument();
    });

    it('should render with title and description', () => {
      render(
        <SettingsCard title="Test Settings" description="This is a description">
          <div>Content</div>
        </SettingsCard>,
      );

      expect(screen.getByText('Test Settings')).toBeInTheDocument();
      expect(screen.getByText('This is a description')).toBeInTheDocument();
    });

    it('should not render description when not provided', () => {
      render(
        <SettingsCard title="Test Settings">
          <div>Content</div>
        </SettingsCard>,
      );

      expect(screen.getByText('Test Settings')).toBeInTheDocument();
      expect(screen.queryByText('This is a description')).not.toBeInTheDocument();
    });

    it('should not render toggle switch by default', () => {
      render(
        <SettingsCard title="Test Settings">
          <div>Content</div>
        </SettingsCard>,
      );

      const toggleSwitch = screen.queryByRole('switch');
      expect(toggleSwitch).not.toBeInTheDocument();
    });

    it('should render toggle switch when enabled and onToggle are provided', () => {
      const mockOnToggle = vi.fn();
      render(
        <SettingsCard title="Test Settings" enabled onToggle={mockOnToggle}>
          <div>Content</div>
        </SettingsCard>,
      );

      const toggleSwitch = screen.getByRole('switch');
      expect(toggleSwitch).toBeInTheDocument();
      expect(toggleSwitch).toBeChecked();
    });

    it('should render toggle switch as unchecked when enabled is false', () => {
      const mockOnToggle = vi.fn();
      render(
        <SettingsCard title="Test Settings" enabled={false} onToggle={mockOnToggle}>
          <div>Content</div>
        </SettingsCard>,
      );

      const toggleSwitch = screen.getByRole('switch');
      expect(toggleSwitch).not.toBeChecked();
    });

    it('should show children when toggle is enabled', () => {
      const mockOnToggle = vi.fn();
      render(
        <SettingsCard title="Test Settings" enabled onToggle={mockOnToggle}>
          <div>Visible Content</div>
        </SettingsCard>,
      );

      expect(screen.getByText('Visible Content')).toBeInTheDocument();
    });

    it('should hide children when toggle is disabled', () => {
      const mockOnToggle = vi.fn();
      render(
        <SettingsCard title="Test Settings" enabled={false} onToggle={mockOnToggle}>
          <div>Hidden Content</div>
        </SettingsCard>,
      );

      expect(screen.queryByText('Hidden Content')).not.toBeInTheDocument();
    });

    it('should show children when no toggle is provided', () => {
      render(
        <SettingsCard title="Test Settings">
          <div>Always Visible</div>
        </SettingsCard>,
      );

      expect(screen.getByText('Always Visible')).toBeInTheDocument();
    });
  });

  describe('User Interactions', () => {
    it('should call onToggle when switch is clicked', async () => {
      const user = userEvent.setup();
      const mockOnToggle = vi.fn();

      render(
        <SettingsCard title="Test Settings" enabled={false} onToggle={mockOnToggle}>
          <div>Content</div>
        </SettingsCard>,
      );

      const toggleSwitch = screen.getByRole('switch');
      await user.click(toggleSwitch);

      expect(mockOnToggle).toHaveBeenCalledWith(true);
      expect(mockOnToggle).toHaveBeenCalledTimes(1);
    });

    it('should call onToggle with false when toggling off', async () => {
      const user = userEvent.setup();
      const mockOnToggle = vi.fn();

      render(
        <SettingsCard title="Test Settings" enabled onToggle={mockOnToggle}>
          <div>Content</div>
        </SettingsCard>,
      );

      const toggleSwitch = screen.getByRole('switch');
      await user.click(toggleSwitch);

      expect(mockOnToggle).toHaveBeenCalledWith(false);
      expect(mockOnToggle).toHaveBeenCalledTimes(1);
    });
  });

  describe('Accessibility', () => {
    it('should render toggle switch with proper structure', () => {
      const mockOnToggle = vi.fn();

      render(
        <SettingsCard title="Registration Flow" enabled onToggle={mockOnToggle}>
          <div>Content</div>
        </SettingsCard>,
      );

      // Verify the switch element exists and is accessible
      const toggleSwitch = screen.getByRole('switch');
      expect(toggleSwitch).toBeInTheDocument();
      expect(toggleSwitch).toHaveAttribute('type', 'checkbox');
    });
  });

  describe('Edge Cases', () => {
    it('should handle only enabled prop without onToggle', () => {
      render(
        <SettingsCard title="Test Settings" enabled>
          <div>Content</div>
        </SettingsCard>,
      );

      expect(screen.queryByRole('switch')).not.toBeInTheDocument();
      expect(screen.getByText('Content')).toBeInTheDocument();
    });

    it('should handle only onToggle prop without enabled', () => {
      const mockOnToggle = vi.fn();

      render(
        <SettingsCard title="Test Settings" onToggle={mockOnToggle}>
          <div>Content</div>
        </SettingsCard>,
      );

      expect(screen.queryByRole('switch')).not.toBeInTheDocument();
      expect(screen.getByText('Content')).toBeInTheDocument();
    });

    it('should render complex children elements', () => {
      render(
        <SettingsCard title="Complex Settings">
          <div>
            <input type="text" placeholder="Username" />
            <button type="button">Submit</button>
          </div>
        </SettingsCard>,
      );

      expect(screen.getByPlaceholderText('Username')).toBeInTheDocument();
      expect(screen.getByRole('button', {name: 'Submit'})).toBeInTheDocument();
    });
  });

  it('should render headerAction if provided', () => {
    render(
      <SettingsCard title="Test Settings" headerAction={<button type="button">Action</button>}>
        <div>Content</div>
      </SettingsCard>,
    );

    expect(screen.getByText('Action')).toBeInTheDocument();
  });

  it('should not render content wrapper if valid children are not present', () => {
    const {container} = render(<SettingsCard title="Test Settings">{false && <div>Content</div>}</SettingsCard>);

    // Should only have the outer paper, title box, but NO inner content paper
    // The outer paper renders the title box div + potential content div
    // We expect only the title box div to be present inside the outer Paper
    const papers = container.querySelectorAll('.MuiPaper-root');
    expect(papers.length).toBe(1); // Only the outer card, no inner content card
  });

  it('should render content wrapper if valid children are present', () => {
    const {container} = render(
      <SettingsCard title="Test Settings">
        <div>Content</div>
      </SettingsCard>,
    );

    const papers = container.querySelectorAll('.MuiPaper-root');
    expect(papers.length).toBe(2); // Outer card + Inner content card
  });
});
