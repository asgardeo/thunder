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
import {page, userEvent, renderWithProviders} from '@thunder/test-utils/browser';
import TranslationEditorHeader from '../TranslationEditorHeader';

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
    it('shows page title when no language is selected', async () => {
      await renderWithProviders(<TranslationEditorHeader {...defaultProps} selectedLanguage={null} />);

      await expect.element(page.getByText('Translations')).toBeInTheDocument();
    });

    it('shows flag and display name when a language is selected', async () => {
      await renderWithProviders(<TranslationEditorHeader {...defaultProps} selectedLanguage="fr-FR" />);

      await expect.element(page.getByText('Flag(fr-FR)')).toBeInTheDocument();
      await expect.element(page.getByText('Language(fr-FR)')).toBeInTheDocument();
    });

    it('renders discard, save, and reset-to-default action buttons', async () => {
      await renderWithProviders(<TranslationEditorHeader {...defaultProps} isEnglish={false} />);

      await expect.element(page.getByText('Discard Changes')).toBeInTheDocument();
      await expect.element(page.getByText('Reset to Default')).toBeInTheDocument();
      await expect.element(page.getByText('Save Changes')).toBeInTheDocument();
    });

    it('hides Reset to Default button when isEnglish is true', async () => {
      await renderWithProviders(<TranslationEditorHeader {...defaultProps} isEnglish />);

      await expect.element(page.getByText('Reset to Default')).not.toBeInTheDocument();
    });

    it('shows Reset to Default button when isEnglish is false', async () => {
      await renderWithProviders(<TranslationEditorHeader {...defaultProps} isEnglish={false} />);

      await expect.element(page.getByText('Reset to Default')).toBeInTheDocument();
    });
  });

  describe('Dirty-changes indicator', () => {
    it('does not show unsaved count when there are no dirty changes', async () => {
      await renderWithProviders(
        <TranslationEditorHeader {...defaultProps} hasDirtyChanges={false} dirtyCount={0} />,
      );

      // The unsaved count uses interpolation: '{{count}} unsaved change'
      await expect.element(page.getByText(/unsaved change/)).not.toBeInTheDocument();
    });

    it('shows unsaved count label when there are dirty changes', async () => {
      await renderWithProviders(<TranslationEditorHeader {...defaultProps} hasDirtyChanges dirtyCount={3} />);

      await expect.element(page.getByText(/unsaved change/)).toBeInTheDocument();
    });
  });

  describe('Button disabled states', () => {
    it('disables Discard when no dirty changes', async () => {
      await renderWithProviders(<TranslationEditorHeader {...defaultProps} hasDirtyChanges={false} />);

      await expect.element(page.getByRole('button', {name: 'Discard Changes'})).toBeDisabled();
    });

    it('enables Discard when dirty changes exist', async () => {
      await renderWithProviders(<TranslationEditorHeader {...defaultProps} hasDirtyChanges dirtyCount={1} />);

      await expect.element(page.getByRole('button', {name: 'Discard Changes'})).not.toBeDisabled();
    });

    it('disables Save when no dirty changes', async () => {
      await renderWithProviders(<TranslationEditorHeader {...defaultProps} hasDirtyChanges={false} />);

      await expect.element(page.getByRole('button', {name: 'Save Changes'})).toBeDisabled();
    });

    it('enables Save when dirty changes exist', async () => {
      await renderWithProviders(<TranslationEditorHeader {...defaultProps} hasDirtyChanges dirtyCount={2} />);

      await expect.element(page.getByRole('button', {name: 'Save Changes'})).not.toBeDisabled();
    });

    it('disables all action buttons while saving', async () => {
      await renderWithProviders(
        <TranslationEditorHeader {...defaultProps} hasDirtyChanges dirtyCount={1} isSaving />,
      );

      await expect.element(page.getByRole('button', {name: 'Discard Changes'})).toBeDisabled();
      await expect.element(page.getByRole('button', {name: 'Save Changes'})).toBeDisabled();
    });

    it('disables Reset to Default when hasNamespace is false', async () => {
      await renderWithProviders(
        <TranslationEditorHeader {...defaultProps} isEnglish={false} hasNamespace={false} />,
      );

      await expect.element(page.getByRole('button', {name: 'Reset to Default'})).toBeDisabled();
    });

    it('enables Reset to Default when hasNamespace is true and not saving', async () => {
      await renderWithProviders(
        <TranslationEditorHeader {...defaultProps} isEnglish={false} hasNamespace isSaving={false} />,
      );

      await expect.element(page.getByRole('button', {name: 'Reset to Default'})).not.toBeDisabled();
    });
  });

  describe('Callbacks', () => {
    it('calls onBack when the back button is clicked', async () => {
      const onBack = vi.fn();

      await renderWithProviders(<TranslationEditorHeader {...defaultProps} onBack={onBack} />);

      // The back button is an IconButton (first button rendered)
      await userEvent.click(page.getByRole('button').all()[0]);

      expect(onBack).toHaveBeenCalledTimes(1);
    });

    it('calls onDiscard when Discard button is clicked', async () => {
      const onDiscard = vi.fn();

      await renderWithProviders(
        <TranslationEditorHeader {...defaultProps} hasDirtyChanges dirtyCount={1} onDiscard={onDiscard} />,
      );

      await userEvent.click(page.getByRole('button', {name: 'Discard Changes'}));

      expect(onDiscard).toHaveBeenCalledTimes(1);
    });

    it('calls onSave when Save button is clicked', async () => {
      const onSave = vi.fn();

      await renderWithProviders(
        <TranslationEditorHeader {...defaultProps} hasDirtyChanges dirtyCount={1} onSave={onSave} />,
      );

      await userEvent.click(page.getByRole('button', {name: 'Save Changes'}));

      expect(onSave).toHaveBeenCalledTimes(1);
    });

    it('calls onResetToDefault when Reset to Default button is clicked', async () => {
      const onResetToDefault = vi.fn();

      await renderWithProviders(
        <TranslationEditorHeader
          {...defaultProps}
          isEnglish={false}
          hasNamespace
          onResetToDefault={onResetToDefault}
        />,
      );

      await userEvent.click(page.getByRole('button', {name: 'Reset to Default'}));

      expect(onResetToDefault).toHaveBeenCalledTimes(1);
    });
  });
});
