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

import {describe, expect, it, vi, beforeEach} from 'vitest';
import {render, screen} from '@thunder/test-utils';
import userEvent from '@testing-library/user-event';
import TranslationEditorHeader from '../TranslationEditorHeader';

vi.mock('react-i18next', async () => {
  const actual = await vi.importActual<typeof import('react-i18next')>('react-i18next');
  return {
    ...actual,
    useTranslation: () => ({t: (key: string) => key}),
  };
});

vi.mock('@thunder/i18n', () => ({
  getDisplayNameForCode: (code: string) => `Language(${code})`,
  toFlagEmoji: (code: string) => `Flag(${code})`,
}));

const defaultProps = {
  selectedLanguage: null,
  hasDirtyChanges: false,
  dirtyCount: 0,
  isSaving: false,
  isEnglish: false,
  hasNamespace: true,
  onBack: vi.fn(),
  onDiscard: vi.fn(),
  onResetToDefault: vi.fn(),
  onSave: vi.fn(),
};

describe('TranslationEditorHeader', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('shows page title key when no language is selected', () => {
      render(<TranslationEditorHeader {...defaultProps} selectedLanguage={null} />);

      expect(screen.getByText('page.title')).toBeInTheDocument();
    });

    it('shows flag and display name when a language is selected', () => {
      render(<TranslationEditorHeader {...defaultProps} selectedLanguage="fr-FR" />);

      expect(screen.getByText('Flag(fr-FR)')).toBeInTheDocument();
      expect(screen.getByText('Language(fr-FR)')).toBeInTheDocument();
    });

    it('renders discard, save, and reset-to-default action buttons', () => {
      render(<TranslationEditorHeader {...defaultProps} isEnglish={false} />);

      expect(screen.getByText('actions.discardChanges')).toBeInTheDocument();
      expect(screen.getByText('actions.resetToDefault')).toBeInTheDocument();
      expect(screen.getByText('actions.saveChanges')).toBeInTheDocument();
    });

    it('hides Reset to Default button when isEnglish is true', () => {
      render(<TranslationEditorHeader {...defaultProps} isEnglish />);

      expect(screen.queryByText('actions.resetToDefault')).not.toBeInTheDocument();
    });

    it('shows Reset to Default button when isEnglish is false', () => {
      render(<TranslationEditorHeader {...defaultProps} isEnglish={false} />);

      expect(screen.getByText('actions.resetToDefault')).toBeInTheDocument();
    });
  });

  describe('Dirty-changes indicator', () => {
    it('does not show unsaved count when there are no dirty changes', () => {
      render(<TranslationEditorHeader {...defaultProps} hasDirtyChanges={false} dirtyCount={0} />);

      expect(screen.queryByText('editor.unsavedCount')).not.toBeInTheDocument();
    });

    it('shows unsaved count label when there are dirty changes', () => {
      render(<TranslationEditorHeader {...defaultProps} hasDirtyChanges dirtyCount={3} />);

      expect(screen.getByText('editor.unsavedCount')).toBeInTheDocument();
    });
  });

  describe('Button disabled states', () => {
    it('disables Discard when no dirty changes', () => {
      render(<TranslationEditorHeader {...defaultProps} hasDirtyChanges={false} />);

      expect(screen.getByText('actions.discardChanges').closest('button')).toBeDisabled();
    });

    it('enables Discard when dirty changes exist', () => {
      render(<TranslationEditorHeader {...defaultProps} hasDirtyChanges dirtyCount={1} />);

      expect(screen.getByText('actions.discardChanges').closest('button')).not.toBeDisabled();
    });

    it('disables Save when no dirty changes', () => {
      render(<TranslationEditorHeader {...defaultProps} hasDirtyChanges={false} />);

      expect(screen.getByText('actions.saveChanges').closest('button')).toBeDisabled();
    });

    it('enables Save when dirty changes exist', () => {
      render(<TranslationEditorHeader {...defaultProps} hasDirtyChanges dirtyCount={2} />);

      expect(screen.getByText('actions.saveChanges').closest('button')).not.toBeDisabled();
    });

    it('disables all action buttons while saving', () => {
      render(<TranslationEditorHeader {...defaultProps} hasDirtyChanges dirtyCount={1} isSaving />);

      expect(screen.getByText('actions.discardChanges').closest('button')).toBeDisabled();
      expect(screen.getByText('actions.saveChanges').closest('button')).toBeDisabled();
    });

    it('disables Reset to Default when hasNamespace is false', () => {
      render(<TranslationEditorHeader {...defaultProps} isEnglish={false} hasNamespace={false} />);

      expect(screen.getByText('actions.resetToDefault').closest('button')).toBeDisabled();
    });

    it('enables Reset to Default when hasNamespace is true and not saving', () => {
      render(<TranslationEditorHeader {...defaultProps} isEnglish={false} hasNamespace isSaving={false} />);

      expect(screen.getByText('actions.resetToDefault').closest('button')).not.toBeDisabled();
    });
  });

  describe('Callbacks', () => {
    it('calls onBack when the back button is clicked', async () => {
      const onBack = vi.fn();
      const user = userEvent.setup();

      render(<TranslationEditorHeader {...defaultProps} onBack={onBack} />);

      // The back button is an IconButton (first button rendered)
      await user.click(screen.getAllByRole('button')[0]);

      expect(onBack).toHaveBeenCalledTimes(1);
    });

    it('calls onDiscard when Discard button is clicked', async () => {
      const onDiscard = vi.fn();
      const user = userEvent.setup();

      render(<TranslationEditorHeader {...defaultProps} hasDirtyChanges dirtyCount={1} onDiscard={onDiscard} />);

      await user.click(screen.getByText('actions.discardChanges'));

      expect(onDiscard).toHaveBeenCalledTimes(1);
    });

    it('calls onSave when Save button is clicked', async () => {
      const onSave = vi.fn();
      const user = userEvent.setup();

      render(<TranslationEditorHeader {...defaultProps} hasDirtyChanges dirtyCount={1} onSave={onSave} />);

      await user.click(screen.getByText('actions.saveChanges'));

      expect(onSave).toHaveBeenCalledTimes(1);
    });

    it('calls onResetToDefault when Reset to Default button is clicked', async () => {
      const onResetToDefault = vi.fn();
      const user = userEvent.setup();

      render(
        <TranslationEditorHeader
          {...defaultProps}
          isEnglish={false}
          hasNamespace
          onResetToDefault={onResetToDefault}
        />,
      );

      await user.click(screen.getByText('actions.resetToDefault'));

      expect(onResetToDefault).toHaveBeenCalledTimes(1);
    });
  });
});
