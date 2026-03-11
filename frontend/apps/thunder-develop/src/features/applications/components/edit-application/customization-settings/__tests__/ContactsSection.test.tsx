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
import ContactsSection from '../ContactsSection';
import type {Application} from '../../../../models/application';

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
    it('should render the contacts section', async () => {
      await render(<ContactsSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      await expect.element(page.getByText('Contacts')).toBeInTheDocument();
      await expect.element(page.getByText('Contact email addresses for application administrators.')).toBeInTheDocument();
    });

    it('should render multiline text field', async () => {
      await render(<ContactsSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const textField = page.getByPlaceholder('admin@example.com, support@example.com');
      expect(textField).toBeInTheDocument();
      expect(textField).toHaveAttribute('rows', '2');
    });

    it('should display helper text', async () => {
      await render(<ContactsSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      await expect.element(page.getByText('Enter email addresses separated by commas')).toBeInTheDocument();
    });
  });

  describe('Initial Values', () => {
    it('should display contacts from application as comma-separated string', async () => {
      await render(<ContactsSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const textField = page.getByPlaceholder('admin@example.com, support@example.com');
      expect(textField).toHaveValue('contact1@example.com, contact2@example.com');
    });

    it('should prioritize editedApp contacts over application', async () => {
      const editedApp = {
        contacts: ['edited1@example.com', 'edited2@example.com'],
      };

      await render(<ContactsSection application={mockApplication} editedApp={editedApp} onFieldChange={mockOnFieldChange} />);

      const textField = page.getByPlaceholder('admin@example.com, support@example.com');
      expect(textField).toHaveValue('edited1@example.com, edited2@example.com');
    });

    it('should display empty string when no contacts are provided', async () => {
      const appWithoutContacts = {...mockApplication, contacts: []};

      await render(<ContactsSection application={appWithoutContacts} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const textField = page.getByPlaceholder('admin@example.com, support@example.com');
      expect(textField).toHaveValue('');
    });
  });

  describe('User Input', () => {
    it('should render text field that accepts user input', async () => {
            const appWithoutContacts = {...mockApplication, contacts: []};

      await render(<ContactsSection application={appWithoutContacts} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const textField = page.getByPlaceholder('admin@example.com, support@example.com');
      await userEvent.type(textField, 'test@example.com');

      // Verify the field accepts input
      expect(textField).toHaveValue('test@example.com');
    });

    it('should handle multiple comma-separated email addresses', async () => {
            const appWithoutContacts = {...mockApplication, contacts: []};

      await render(<ContactsSection application={appWithoutContacts} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const textField = page.getByPlaceholder('admin@example.com, support@example.com');
      await userEvent.type(textField, 'test1@example.com, test2@example.com');

      expect(textField).toHaveValue('test1@example.com, test2@example.com');
    });
  });

  describe('Edge Cases', () => {
    it('should handle missing contacts in application', async () => {
      const appWithoutContacts = {...mockApplication};
      delete (appWithoutContacts as Partial<Application>).contacts;

      await render(<ContactsSection application={appWithoutContacts} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const textField = page.getByPlaceholder('admin@example.com, support@example.com');
      expect(textField).toHaveValue('');
    });

    it('should handle single email address', async () => {
      const appWithOneContact = {...mockApplication, contacts: ['single@example.com']};

      await render(<ContactsSection application={appWithOneContact} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const textField = page.getByPlaceholder('admin@example.com, support@example.com');
      expect(textField).toHaveValue('single@example.com');
    });

    it('should handle clearing all contacts', async () => {
      
      await render(<ContactsSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const textField = page.getByPlaceholder('admin@example.com, support@example.com');
      await userEvent.clear(textField);

      // Verify field is cleared
      expect(textField).toHaveValue('');
    });
  });

  describe('Contacts Sync Effect', () => {
    it('should not call onFieldChange when contacts match current value', async () => {
      // When the form value matches the current contacts, no update should be triggered
      const editedApp = {
        contacts: ['contact1@example.com', 'contact2@example.com'],
      };

      await render(<ContactsSection application={mockApplication} editedApp={editedApp} onFieldChange={mockOnFieldChange} />);

      // Initial render should not trigger onFieldChange since values match
      expect(mockOnFieldChange).not.toHaveBeenCalled();
    });

    it('should use editedApp contacts when provided', async () => {
      const editedApp = {
        contacts: ['edited@example.com'],
      };

      await render(<ContactsSection application={mockApplication} editedApp={editedApp} onFieldChange={mockOnFieldChange} />);

      const textField = page.getByPlaceholder('admin@example.com, support@example.com');
      expect(textField).toHaveValue('edited@example.com');
    });

    it('should fall back to application contacts when editedApp contacts not provided', async () => {
      await render(<ContactsSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const textField = page.getByPlaceholder('admin@example.com, support@example.com');
      expect(textField).toHaveValue('contact1@example.com, contact2@example.com');
    });

    it('should handle undefined contacts in both editedApp and application gracefully', async () => {
      const appWithoutContacts = {...mockApplication};
      delete (appWithoutContacts as Partial<Application>).contacts;

      await render(<ContactsSection application={appWithoutContacts} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const textField = page.getByPlaceholder('admin@example.com, support@example.com');
      expect(textField).toHaveValue('');
    });
  });

});
