/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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

import {TextField} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';
import {useEffect} from 'react';
import {useForm, Controller} from 'react-hook-form';
import type {Application} from '../../../models/application';
import SettingsCard from '../../../../../components/SettingsCard';

/**
 * Props for the {@link ContactsSection} component.
 */
interface ContactsSectionProps {
  /**
   * The application being edited
   */
  application: Application;
  /**
   * Partial application object containing edited fields
   */
  editedApp: Partial<Application>;
  /**
   * Callback function to handle field value changes
   * @param field - The application field being updated
   * @param value - The new value for the field
   */
  onFieldChange: (field: keyof Application, value: unknown) => void;
}

/**
 * Section component for configuring application contact information.
 *
 * Provides a multiline text field for entering contact email addresses.
 * Multiple emails should be comma-separated.
 * Changes are automatically converted to array format and synced to parent.
 *
 * @param props - Component props
 * @returns Contact information input UI within a SettingsCard
 */
export default function ContactsSection({application, editedApp, onFieldChange}: ContactsSectionProps) {
  const {t} = useTranslation();

  const {control, watch} = useForm({
    mode: 'onChange',
    defaultValues: {
      contacts: (editedApp.contacts ?? application.contacts ?? []).join(', '),
    },
  });

  const contacts = watch('contacts');

  /**
   * Effect to synchronize contact information changes from the form.
   */
  useEffect(() => {
    const currentContacts = (editedApp.contacts ?? application.contacts ?? []).join(', ');
    if (contacts !== currentContacts) {
      const contactsArray =
        contacts
          ?.split(',')
          .map((c) => c.trim())
          .filter((c) => c) ?? [];
      onFieldChange('contacts', contactsArray);
    }
  }, [contacts, editedApp.contacts, application.contacts, onFieldChange]);

  return (
    <SettingsCard
      title={t('applications:edit.general.sections.contacts')}
      description={t('applications:edit.general.sections.contacts.description')}
    >
      <Controller
        name="contacts"
        control={control}
        render={({field}) => (
          <TextField
            {...field}
            fullWidth
            multiline
            rows={2}
            placeholder={t('applications:edit.general.contacts.placeholder')}
            helperText={t('applications:edit.general.contacts.hint')}
          />
        )}
      />
    </SettingsCard>
  );
}
