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
import {page, userEvent, renderWithProviders, getByDisplayValue} from '@thunder/test-utils/browser';
import TranslationFieldsView from '../TranslationFieldsView';

const sampleValues = {
  'actions.save': 'Save',
  'actions.cancel': 'Cancel',
  'page.title': 'My Page',
};

const defaultProps = {
  localValues: sampleValues,
  serverValues: sampleValues,
  search: '',
  isCustomNamespace: false,
  onChange: vi.fn(),
  onResetField: vi.fn(),
};

describe('TranslationFieldsView', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('renders a text field for each translation key', async () => {
      await renderWithProviders(<TranslationFieldsView {...defaultProps} />);

      await expect.element(getByDisplayValue('Save')).toBeInTheDocument();
      await expect.element(getByDisplayValue('Cancel')).toBeInTheDocument();
      await expect.element(getByDisplayValue('My Page')).toBeInTheDocument();
    });

    it('renders the translation key as a label above each field', async () => {
      await renderWithProviders(<TranslationFieldsView {...defaultProps} />);

      await expect.element(page.getByText('actions.save')).toBeInTheDocument();
      await expect.element(page.getByText('actions.cancel')).toBeInTheDocument();
    });

    it('shows no-keys message when localValues is empty', async () => {
      await renderWithProviders(<TranslationFieldsView {...defaultProps} localValues={{}} serverValues={{}} />);

      await expect.element(page.getByText('No translatable keys in this namespace.')).toBeInTheDocument();
    });
  });

  describe('Search filtering', () => {
    it('shows only keys matching the search query', async () => {
      await renderWithProviders(<TranslationFieldsView {...defaultProps} search="save" />);

      await expect.element(getByDisplayValue('Save')).toBeInTheDocument();
      await expect.element(getByDisplayValue('Cancel')).not.toBeInTheDocument();
    });

    it('matches search against key names (case-insensitive)', async () => {
      await renderWithProviders(<TranslationFieldsView {...defaultProps} search="PAGE" />);

      await expect.element(getByDisplayValue('My Page')).toBeInTheDocument();
      await expect.element(getByDisplayValue('Save')).not.toBeInTheDocument();
    });

    it('matches search against field values', async () => {
      await renderWithProviders(<TranslationFieldsView {...defaultProps} search="Cancel" />);

      await expect.element(getByDisplayValue('Cancel')).toBeInTheDocument();
      await expect.element(getByDisplayValue('Save')).not.toBeInTheDocument();
    });

    it('shows no-results message when search matches nothing', async () => {
      await renderWithProviders(<TranslationFieldsView {...defaultProps} search="nonexistent" />);

      await expect.element(page.getByText('No matching translations.')).toBeInTheDocument();
    });
  });

  describe('Dirty field state', () => {
    it('does not show reset button for a clean field', async () => {
      await renderWithProviders(<TranslationFieldsView {...defaultProps} />);

      await expect.element(page.getByRole('button')).not.toBeInTheDocument();
    });

    it('shows a reset button when a field has a local change', async () => {
      await renderWithProviders(
        <TranslationFieldsView
          {...defaultProps}
          localValues={{'actions.save': 'Enregistrer', 'actions.cancel': 'Cancel', 'page.title': 'My Page'}}
        />,
      );

      await expect.element(page.getByRole('button')).toBeInTheDocument();
    });

    it('shows reset buttons only for dirty fields', async () => {
      await renderWithProviders(
        <TranslationFieldsView
          {...defaultProps}
          localValues={{
            'actions.save': 'Enregistrer',
            'actions.cancel': 'Annuler',
            'page.title': 'My Page',
          }}
        />,
      );

      // Two fields are dirty (save, cancel), page.title is clean
      expect(page.getByRole('button').all()).toHaveLength(2);
    });
  });

  describe('Interaction', () => {
    it('calls onChange with the key and new value when a field is edited', async () => {
      const onChange = vi.fn();

      await renderWithProviders(<TranslationFieldsView {...defaultProps} onChange={onChange} />);

      // Use userEvent.fill to replace the value of the controlled "Save" input
      const saveInput = getByDisplayValue('Save');
      await userEvent.fill(saveInput, 'Enregistrer');

      expect(onChange).toHaveBeenCalledWith('actions.save', 'Enregistrer');
    });

    it('calls onResetField with the key when the reset button is clicked', async () => {
      const onResetField = vi.fn();

      await renderWithProviders(
        <TranslationFieldsView
          {...defaultProps}
          localValues={{'actions.save': 'Enregistrer', 'actions.cancel': 'Cancel', 'page.title': 'My Page'}}
          onResetField={onResetField}
        />,
      );

      await userEvent.click(page.getByRole('button'));

      expect(onResetField).toHaveBeenCalledWith('actions.save');
    });
  });

  describe('Add Key (custom namespace)', () => {
    it('shows the Add Key button when isCustomNamespace is true', async () => {
      await renderWithProviders(<TranslationFieldsView {...defaultProps} isCustomNamespace />);

      await expect.element(page.getByText('Add Key')).toBeInTheDocument();
    });

    it('does not show the Add Key button when isCustomNamespace is false', async () => {
      await renderWithProviders(<TranslationFieldsView {...defaultProps} isCustomNamespace={false} />);

      await expect.element(page.getByText('Add Key')).not.toBeInTheDocument();
    });

    it('shows the add key form when the Add Key button is clicked', async () => {
      await renderWithProviders(<TranslationFieldsView {...defaultProps} isCustomNamespace />);

      await userEvent.click(page.getByText('Add Key'));

      await expect.element(page.getByLabelText('Key')).toBeInTheDocument();
      await expect.element(page.getByLabelText('Value')).toBeInTheDocument();
    });

    it('calls onChange and closes the form when a new key is submitted', async () => {
      const onChange = vi.fn();
      await renderWithProviders(<TranslationFieldsView {...defaultProps} isCustomNamespace onChange={onChange} />);

      await userEvent.click(page.getByText('Add Key'));

      const keyInput = page.getByPlaceholder('e.g. my.translation.key');
      const valueInput = page.getByPlaceholder('Translation value');

      await userEvent.type(keyInput, 'new.key');
      await userEvent.type(valueInput, 'New Value');

      await userEvent.click(page.getByRole('button', {name: 'Add'}));

      expect(onChange).toHaveBeenCalledWith('new.key', 'New Value');
      // Form should be closed, Add Key button visible again
      await expect.element(page.getByText('Add Key')).toBeInTheDocument();
    });

    it('closes the form and clears inputs when Cancel is clicked', async () => {
      await renderWithProviders(<TranslationFieldsView {...defaultProps} isCustomNamespace />);

      await userEvent.click(page.getByText('Add Key'));

      await userEvent.type(page.getByPlaceholder('e.g. my.translation.key'), 'some.key');

      await userEvent.click(page.getByRole('button', {name: 'Cancel'}));

      // Form should be closed, Add Key button visible again
      await expect.element(page.getByText('Add Key')).toBeInTheDocument();
    });

    it('shows a duplicate key error when the entered key already exists', async () => {
      await renderWithProviders(<TranslationFieldsView {...defaultProps} isCustomNamespace />);

      await userEvent.click(page.getByText('Add Key'));

      await userEvent.type(page.getByPlaceholder('e.g. my.translation.key'), 'actions.save');

      await expect.element(page.getByText('This key already exists.')).toBeInTheDocument();
    });

    it('disables the submit button when the key is empty', async () => {
      await renderWithProviders(<TranslationFieldsView {...defaultProps} isCustomNamespace />);

      await userEvent.click(page.getByText('Add Key'));

      await expect.element(page.getByRole('button', {name: 'Add'})).toBeDisabled();
    });

    it('disables the submit button when the key is a duplicate', async () => {
      await renderWithProviders(<TranslationFieldsView {...defaultProps} isCustomNamespace />);

      await userEvent.click(page.getByText('Add Key'));

      await userEvent.type(page.getByPlaceholder('e.g. my.translation.key'), 'actions.save');

      await expect.element(page.getByRole('button', {name: 'Add'})).toBeDisabled();
    });
  });
});
