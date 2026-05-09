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

import {render, screen, fireEvent, waitFor} from '@testing-library/react';
import {AppWindow} from '@wso2/oxygen-ui-icons-react';
import {describe, it, expect, vi} from 'vitest';
import ResourceAvatar from '../ResourceAvatar';

describe('ResourceAvatar', () => {
  describe('Read-only mode (no onSelect)', () => {
    it('should render the fallback icon when no value is provided', () => {
      render(<ResourceAvatar fallback={<AppWindow data-testid="fallback-icon" />} />);

      expect(screen.getByTestId('fallback-icon')).toBeInTheDocument();
    });

    it('should render the emoji character when value is emoji:-prefixed', () => {
      render(<ResourceAvatar value="emoji:🎉" />);

      expect(screen.getByText('🎉')).toBeInTheDocument();
    });

    it('should render the raw emoji character when value has no prefix', () => {
      render(<ResourceAvatar value="🐼" />);

      expect(screen.getByText('🐼')).toBeInTheDocument();
    });

    it('should pass the URL as an img src when value is a URL', () => {
      render(<ResourceAvatar value="https://example.com/logo.png" />);

      const img = screen.getByRole('img');
      expect(img).toHaveAttribute('src', 'https://example.com/logo.png');
    });

    it('should not render an edit button when editable and onSelect are not provided', () => {
      render(<ResourceAvatar value="emoji:🎉" />);

      expect(screen.queryByRole('button')).not.toBeInTheDocument();
    });

    it('should call onClick when avatar is clicked in read-only mode', () => {
      const handleClick = vi.fn();
      render(<ResourceAvatar value="emoji:🎉" onClick={handleClick} />);

      const avatar = screen.getByText('🎉').closest('[class*="Avatar"]') ?? screen.getByText('🎉').parentElement!;
      fireEvent.click(avatar);

      expect(handleClick).toHaveBeenCalledTimes(1);
    });

    it('should render with emoji value', () => {
      render(<ResourceAvatar value="emoji:🎉" />);

      expect(screen.getByText('🎉')).toBeInTheDocument();
    });
  });

  describe('Edit mode (onSelect provided)', () => {
    it('should render an edit (pencil) button when onSelect and editable are provided', () => {
      render(<ResourceAvatar editable value="emoji:🎉" onSelect={vi.fn()} />);

      expect(screen.getByRole('button', {name: 'Change logo'})).toBeInTheDocument();
    });

    it('should have the default aria-label "Change logo" on the edit button', () => {
      render(<ResourceAvatar editable value="emoji:🎉" onSelect={vi.fn()} />);

      expect(screen.getByRole('button', {name: 'Change logo'})).toBeInTheDocument();
    });

    it('should accept a custom editAriaLabel', () => {
      render(<ResourceAvatar editable value="emoji:🎉" onSelect={vi.fn()} editAriaLabel="Update icon" />);

      expect(screen.getByRole('button', {name: 'Update icon'})).toBeInTheDocument();
    });

    it('should open the ResourceLogoDialog when the edit button is clicked', () => {
      render(<ResourceAvatar editable value="emoji:🎉" onSelect={vi.fn()} />);

      expect(screen.queryByRole('dialog')).not.toBeInTheDocument();

      fireEvent.click(screen.getByRole('button', {name: 'Change logo'}));

      expect(screen.getByRole('dialog')).toBeInTheDocument();
    });

    it('should open the ResourceLogoDialog when the avatar itself is clicked', () => {
      render(<ResourceAvatar value="emoji:🎉" onSelect={vi.fn()} />);

      fireEvent.click(screen.getByText('🎉'));

      expect(screen.getByRole('dialog')).toBeInTheDocument();
    });

    it('should close the dialog when the cancel button is clicked', async () => {
      render(<ResourceAvatar editable value="emoji:🎉" onSelect={vi.fn()} />);

      fireEvent.click(screen.getByRole('button', {name: 'Change logo'}));
      expect(screen.getByRole('dialog')).toBeInTheDocument();

      fireEvent.click(screen.getByRole('button', {name: /cancel/i}));
      await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
    });

    it('should call onSelect with the confirmed value and close the dialog', async () => {
      const handleSelect = vi.fn();
      render(<ResourceAvatar editable value="emoji:🎉" onSelect={handleSelect} />);

      // Open the dialog
      fireEvent.click(screen.getByRole('button', {name: 'Change logo'}));

      // The dialog pre-populates with the existing emoji, so Select is enabled
      const selectButton = screen.getByRole('button', {name: /select/i});
      fireEvent.click(selectButton);

      expect(handleSelect).toHaveBeenCalledWith('emoji:🎉');
      await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
    });

    it('should show fallback icon inside avatar when no value provided in edit mode', () => {
      render(<ResourceAvatar editable fallback={<AppWindow data-testid="fallback-icon" />} onSelect={vi.fn()} />);

      expect(screen.getByTestId('fallback-icon')).toBeInTheDocument();
    });
  });
});
