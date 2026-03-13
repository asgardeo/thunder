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
import {render, screen, fireEvent} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import UnsavedChangesBar from '../UnsavedChangesBar';

const defaultProps = {
  message: 'You have unsaved changes',
  resetLabel: 'Reset',
  saveLabel: 'Save',
  savingLabel: 'Saving...',
  isSaving: false,
  onReset: vi.fn(),
  onSave: vi.fn(),
};

describe('UnsavedChangesBar', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render the message text', () => {
      render(<UnsavedChangesBar {...defaultProps} />);

      expect(screen.getByText('You have unsaved changes')).toBeInTheDocument();
    });

    it('should render the reset button with the provided label', () => {
      render(<UnsavedChangesBar {...defaultProps} />);

      expect(screen.getByRole('button', {name: 'Reset'})).toBeInTheDocument();
    });

    it('should render the save button with saveLabel when not saving', () => {
      render(<UnsavedChangesBar {...defaultProps} isSaving={false} />);

      expect(screen.getByRole('button', {name: 'Save'})).toBeInTheDocument();
    });

    it('should render the save button with savingLabel when saving', () => {
      render(<UnsavedChangesBar {...defaultProps} isSaving />);

      expect(screen.getByRole('button', {name: 'Saving...'})).toBeInTheDocument();
    });

    it('should render a warning indicator', () => {
      render(<UnsavedChangesBar {...defaultProps} />);

      expect(screen.getByText('!')).toBeInTheDocument();
    });
  });

  describe('Save Button State', () => {
    it('should enable the save button when not saving', () => {
      render(<UnsavedChangesBar {...defaultProps} isSaving={false} />);

      expect(screen.getByRole('button', {name: 'Save'})).not.toBeDisabled();
    });

    it('should disable the save button when saving is in progress', () => {
      render(<UnsavedChangesBar {...defaultProps} isSaving />);

      expect(screen.getByRole('button', {name: 'Saving...'})).toBeDisabled();
    });
  });

  describe('Interactions', () => {
    it('should call onReset when the reset button is clicked', async () => {
      const user = userEvent.setup();
      const onReset = vi.fn();
      render(<UnsavedChangesBar {...defaultProps} onReset={onReset} />);

      await user.click(screen.getByRole('button', {name: 'Reset'}));

      expect(onReset).toHaveBeenCalledTimes(1);
    });

    it('should call onSave when the save button is clicked', async () => {
      const user = userEvent.setup();
      const onSave = vi.fn();
      render(<UnsavedChangesBar {...defaultProps} onSave={onSave} />);

      await user.click(screen.getByRole('button', {name: 'Save'}));

      expect(onSave).toHaveBeenCalledTimes(1);
    });

    it('should not call onSave when the save button is disabled (isSaving)', () => {
      const onSave = vi.fn();
      render(<UnsavedChangesBar {...defaultProps} isSaving onSave={onSave} />);

      fireEvent.click(screen.getByRole('button', {name: 'Saving...'}));

      expect(onSave).not.toHaveBeenCalled();
    });

    it('should not call onReset when save is clicked', async () => {
      const user = userEvent.setup();
      const onReset = vi.fn();
      render(<UnsavedChangesBar {...defaultProps} onReset={onReset} />);

      await user.click(screen.getByRole('button', {name: 'Save'}));

      expect(onReset).not.toHaveBeenCalled();
    });
  });

  describe('Custom Labels', () => {
    it('should render custom message, resetLabel, saveLabel, and savingLabel', () => {
      render(
        <UnsavedChangesBar
          message="Pending changes detected"
          resetLabel="Discard"
          saveLabel="Apply"
          savingLabel="Applying..."
          isSaving={false}
          onReset={vi.fn()}
          onSave={vi.fn()}
        />,
      );

      expect(screen.getByText('Pending changes detected')).toBeInTheDocument();
      expect(screen.getByRole('button', {name: 'Discard'})).toBeInTheDocument();
      expect(screen.getByRole('button', {name: 'Apply'})).toBeInTheDocument();
    });

    it('should render custom savingLabel when isSaving is true', () => {
      render(
        <UnsavedChangesBar
          message="Pending changes"
          resetLabel="Discard"
          saveLabel="Apply"
          savingLabel="Applying..."
          isSaving
          onReset={vi.fn()}
          onSave={vi.fn()}
        />,
      );

      expect(screen.getByRole('button', {name: 'Applying...'})).toBeInTheDocument();
    });
  });
});
