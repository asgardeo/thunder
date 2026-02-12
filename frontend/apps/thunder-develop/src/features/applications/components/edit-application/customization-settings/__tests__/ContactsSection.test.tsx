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
import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import ContactsSection from '../ContactsSection';
import type {Application} from '../../../../models/application';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

describe('ContactsSection', () => {
  const mockApplication: Application = {
    id: 'test-app-id',
    name: 'Test Application',
    description: 'Test Description',
    template: 'custom',
    contacts: ['contact1@example.com', 'contact2@example.com'],
  } as Application;

  const mockOnFieldChange = vi.fn();

  beforeEach(() => {
    mockOnFieldChange.mockClear();
  });

  describe('Rendering', () => {
    it('should render the contacts section', () => {
      render(<ContactsSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.getByText('applications:edit.general.sections.contacts')).toBeInTheDocument();
      expect(screen.getByText('applications:edit.general.sections.contacts.description')).toBeInTheDocument();
    });

    it('should render multiline text field', () => {
      render(<ContactsSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const textField = screen.getByPlaceholderText('applications:edit.general.contacts.placeholder');
      expect(textField).toBeInTheDocument();
      expect(textField).toHaveAttribute('rows', '2');
    });

    it('should display helper text', () => {
      render(<ContactsSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.getByText('applications:edit.general.contacts.hint')).toBeInTheDocument();
    });
  });

  describe('Initial Values', () => {
    it('should display contacts from application as comma-separated string', () => {
      render(<ContactsSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const textField = screen.getByPlaceholderText('applications:edit.general.contacts.placeholder');
      expect(textField).toHaveValue('contact1@example.com, contact2@example.com');
    });

    it('should prioritize editedApp contacts over application', () => {
      const editedApp = {
        contacts: ['edited1@example.com', 'edited2@example.com'],
      };

      render(<ContactsSection application={mockApplication} editedApp={editedApp} onFieldChange={mockOnFieldChange} />);

      const textField = screen.getByPlaceholderText('applications:edit.general.contacts.placeholder');
      expect(textField).toHaveValue('edited1@example.com, edited2@example.com');
    });

    it('should display empty string when no contacts are provided', () => {
      const appWithoutContacts = {...mockApplication, contacts: []};

      render(<ContactsSection application={appWithoutContacts} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const textField = screen.getByPlaceholderText('applications:edit.general.contacts.placeholder');
      expect(textField).toHaveValue('');
    });
  });

  describe('User Input', () => {
    it('should render text field that accepts user input', async () => {
      const user = userEvent.setup({delay: null});
      const appWithoutContacts = {...mockApplication, contacts: []};

      render(<ContactsSection application={appWithoutContacts} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const textField = screen.getByPlaceholderText('applications:edit.general.contacts.placeholder');
      await user.type(textField, 'test@example.com');

      // Verify the field accepts input
      expect(textField).toHaveValue('test@example.com');
    });

    it('should handle multiple comma-separated email addresses', async () => {
      const user = userEvent.setup({delay: null});
      const appWithoutContacts = {...mockApplication, contacts: []};

      render(<ContactsSection application={appWithoutContacts} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const textField = screen.getByPlaceholderText('applications:edit.general.contacts.placeholder');
      await user.type(textField, 'test1@example.com, test2@example.com');

      expect(textField).toHaveValue('test1@example.com, test2@example.com');
    });
  });

  describe('Edge Cases', () => {
    it('should handle missing contacts in application', () => {
      const appWithoutContacts = {...mockApplication};
      delete (appWithoutContacts as Partial<Application>).contacts;

      render(<ContactsSection application={appWithoutContacts} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const textField = screen.getByPlaceholderText('applications:edit.general.contacts.placeholder');
      expect(textField).toHaveValue('');
    });

    it('should handle single email address', () => {
      const appWithOneContact = {...mockApplication, contacts: ['single@example.com']};

      render(<ContactsSection application={appWithOneContact} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const textField = screen.getByPlaceholderText('applications:edit.general.contacts.placeholder');
      expect(textField).toHaveValue('single@example.com');
    });

    it('should handle clearing all contacts', async () => {
      const user = userEvent.setup({delay: null});

      render(<ContactsSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const textField = screen.getByPlaceholderText('applications:edit.general.contacts.placeholder');
      await user.clear(textField);

      // Verify field is cleared
      expect(textField).toHaveValue('');
    });
  });

  describe('Contacts Sync Effect', () => {
    it('should not call onFieldChange when contacts match current value', () => {
      // When the form value matches the current contacts, no update should be triggered
      const editedApp = {
        contacts: ['contact1@example.com', 'contact2@example.com'],
      };

      render(<ContactsSection application={mockApplication} editedApp={editedApp} onFieldChange={mockOnFieldChange} />);

      // Initial render should not trigger onFieldChange since values match
      expect(mockOnFieldChange).not.toHaveBeenCalled();
    });

    it('should use editedApp contacts when provided', () => {
      const editedApp = {
        contacts: ['edited@example.com'],
      };

      render(<ContactsSection application={mockApplication} editedApp={editedApp} onFieldChange={mockOnFieldChange} />);

      const textField = screen.getByPlaceholderText('applications:edit.general.contacts.placeholder');
      expect(textField).toHaveValue('edited@example.com');
    });

    it('should fall back to application contacts when editedApp contacts not provided', () => {
      render(<ContactsSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const textField = screen.getByPlaceholderText('applications:edit.general.contacts.placeholder');
      expect(textField).toHaveValue('contact1@example.com, contact2@example.com');
    });

    it('should handle undefined contacts in both editedApp and application gracefully', () => {
      const appWithoutContacts = {...mockApplication};
      delete (appWithoutContacts as Partial<Application>).contacts;

      render(<ContactsSection application={appWithoutContacts} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const textField = screen.getByPlaceholderText('applications:edit.general.contacts.placeholder');
      expect(textField).toHaveValue('');
    });
  });

});
