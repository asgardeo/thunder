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
import {render, screen, fireEvent} from '@thunder/test-utils';
import userEvent from '@testing-library/user-event';
import TranslationFieldsView from '../TranslationFieldsView';

vi.mock('react-i18next', async () => {
  const actual = await vi.importActual<typeof import('react-i18next')>('react-i18next');
  return {
    ...actual,
    useTranslation: () => ({t: (key: string) => key}),
  };
});

const sampleValues = {
  'actions.save': 'Save',
  'actions.cancel': 'Cancel',
  'page.title': 'My Page',
};

const defaultProps = {
  localValues: sampleValues,
  serverValues: sampleValues,
  search: '',
  onChange: vi.fn(),
  onResetField: vi.fn(),
};

describe('TranslationFieldsView', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('renders a text field for each translation key', () => {
      render(<TranslationFieldsView {...defaultProps} />);

      expect(screen.getByDisplayValue('Save')).toBeInTheDocument();
      expect(screen.getByDisplayValue('Cancel')).toBeInTheDocument();
      expect(screen.getByDisplayValue('My Page')).toBeInTheDocument();
    });

    it('renders the translation key as a label above each field', () => {
      render(<TranslationFieldsView {...defaultProps} />);

      expect(screen.getByText('actions.save')).toBeInTheDocument();
      expect(screen.getByText('actions.cancel')).toBeInTheDocument();
    });

    it('shows no-keys message when localValues is empty', () => {
      render(<TranslationFieldsView {...defaultProps} localValues={{}} serverValues={{}} />);

      expect(screen.getByText('editor.noKeys')).toBeInTheDocument();
    });
  });

  describe('Search filtering', () => {
    it('shows only keys matching the search query', () => {
      render(<TranslationFieldsView {...defaultProps} search="save" />);

      expect(screen.getByDisplayValue('Save')).toBeInTheDocument();
      expect(screen.queryByDisplayValue('Cancel')).not.toBeInTheDocument();
    });

    it('matches search against key names (case-insensitive)', () => {
      render(<TranslationFieldsView {...defaultProps} search="PAGE" />);

      expect(screen.getByDisplayValue('My Page')).toBeInTheDocument();
      expect(screen.queryByDisplayValue('Save')).not.toBeInTheDocument();
    });

    it('matches search against field values', () => {
      render(<TranslationFieldsView {...defaultProps} search="Cancel" />);

      expect(screen.getByDisplayValue('Cancel')).toBeInTheDocument();
      expect(screen.queryByDisplayValue('Save')).not.toBeInTheDocument();
    });

    it('shows no-results message when search matches nothing', () => {
      render(<TranslationFieldsView {...defaultProps} search="nonexistent" />);

      expect(screen.getByText('editor.noResults')).toBeInTheDocument();
    });
  });

  describe('Dirty field state', () => {
    it('does not show reset button for a clean field', () => {
      render(<TranslationFieldsView {...defaultProps} />);

      expect(screen.queryByRole('button')).not.toBeInTheDocument();
    });

    it('shows a reset button when a field has a local change', () => {
      render(
        <TranslationFieldsView
          {...defaultProps}
          localValues={{'actions.save': 'Enregistrer', 'actions.cancel': 'Cancel', 'page.title': 'My Page'}}
        />,
      );

      expect(screen.getByRole('button')).toBeInTheDocument();
    });

    it('shows reset buttons only for dirty fields', () => {
      render(
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
      expect(screen.getAllByRole('button')).toHaveLength(2);
    });
  });

  describe('Interaction', () => {
    it('calls onChange with the key and new value when a field is edited', () => {
      const onChange = vi.fn();

      render(<TranslationFieldsView {...defaultProps} onChange={onChange} />);

      // The field is a controlled input (value driven by localValues prop), so
      // userEvent.type accumulates against the re-rendered prop value on each
      // keystroke. Use fireEvent.change to set an exact target value instead.
      fireEvent.change(screen.getByDisplayValue('Save'), {target: {value: 'Enregistrer'}});

      expect(onChange).toHaveBeenCalledWith('actions.save', 'Enregistrer');
    });

    it('calls onResetField with the key when the reset button is clicked', async () => {
      const onResetField = vi.fn();
      const user = userEvent.setup();

      render(
        <TranslationFieldsView
          {...defaultProps}
          localValues={{'actions.save': 'Enregistrer', 'actions.cancel': 'Cancel', 'page.title': 'My Page'}}
          onResetField={onResetField}
        />,
      );

      await user.click(screen.getByRole('button'));

      expect(onResetField).toHaveBeenCalledWith('actions.save');
    });
  });
});
