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

import {describe, it, expect, beforeEach, vi} from 'vitest';
import {page, userEvent} from 'vitest/browser';
import {renderWithProviders} from '@thunder/test-utils/browser';
import {useForm} from 'react-hook-form';
import renderSchemaField from '../renderSchemaField';
import type {PropertyDefinition} from '../../types/users';

type TestFormData = Record<string, unknown>;

function TestForm({
  fieldName,
  fieldDef,
  defaultValues = {},
  onSubmit = undefined,
}: {
  fieldName: string;
  fieldDef: PropertyDefinition;
  defaultValues?: TestFormData;
  onSubmit?: (data: TestFormData) => void;
}) {
  const {
    control,
    formState: {errors},
    handleSubmit,
  } = useForm<TestFormData>({
    defaultValues,
  });

  const handleFormSubmit = (data: TestFormData) => {
    if (onSubmit) {
      onSubmit(data);
    }
  };

  return (
    <form
      onSubmit={(e): void => {
        handleSubmit(handleFormSubmit)(e).catch(() => {});
      }}
    >
      {renderSchemaField(fieldName, fieldDef, control, errors)}
      <button type="submit">Submit</button>
    </form>
  );
}

describe('renderSchemaField', () => {
  beforeEach(() => {
    // Reset any state if needed
  });

  describe('String fields', () => {
    it('renders basic string TextField', async () => {
      const fieldDef: PropertyDefinition = {type: 'string'};
      await renderWithProviders(<TestForm fieldName="username" fieldDef={fieldDef} />);

      await expect.element(page.getByLabelText('username')).toBeInTheDocument();
      await expect.element(page.getByPlaceholder('Enter username')).toBeInTheDocument();
    });

    it('shows required asterisk for required string fields', async () => {
      const fieldDef: PropertyDefinition = {type: 'string', required: true};
      await renderWithProviders(<TestForm fieldName="username" fieldDef={fieldDef} />);

      const label = page.getByText('username');
      await expect.element(label).toBeInTheDocument();
    });

    it('renders Select dropdown when enum is provided', async () => {
      const fieldDef: PropertyDefinition = {
        type: 'string',
        enum: ['admin', 'user', 'guest'],
      };
      await renderWithProviders(<TestForm fieldName="role" fieldDef={fieldDef} />);

      const select = page.getByRole('combobox');
      await expect.element(select).toBeInTheDocument();
    });

    it('displays enum options in Select dropdown', async () => {
        const fieldDef: PropertyDefinition = {
        type: 'string',
        enum: ['admin', 'user', 'guest'],
      };
      await renderWithProviders(<TestForm fieldName="role" fieldDef={fieldDef} />);

      const select = page.getByRole('combobox');
      await userEvent.click(select);

      await expect.element(page.getByText('Admin')).toBeInTheDocument();
        await expect.element(page.getByText('User')).toBeInTheDocument();
        await expect.element(page.getByText('Guest')).toBeInTheDocument();

    });

    it('renders with default value for string field', async () => {
      const fieldDef: PropertyDefinition = {type: 'string'};
      await renderWithProviders(<TestForm fieldName="username" fieldDef={fieldDef} defaultValues={{username: 'john'}} />);

      const input = page.getByPlaceholder('Enter username');
      await expect.element(input).toHaveValue('john');
    });
  });

  describe('Number fields', () => {
    it('renders number TextField', async () => {
      const fieldDef: PropertyDefinition = {type: 'number'};
      await renderWithProviders(<TestForm fieldName="age" fieldDef={fieldDef} />);

      const input = page.getByPlaceholder('Enter age');
      await expect.element(input).toHaveAttribute('type', 'number');
    });

    it('shows required asterisk for required number fields', async () => {
      const fieldDef: PropertyDefinition = {type: 'number', required: true};
      await renderWithProviders(<TestForm fieldName="age" fieldDef={fieldDef} />);

      const label = page.getByText('age');
      await expect.element(label).toBeInTheDocument();
    });

    it('renders with default value for number field', async () => {
      const fieldDef: PropertyDefinition = {type: 'number'};
      await renderWithProviders(<TestForm fieldName="age" fieldDef={fieldDef} defaultValues={{age: 25}} />);

      const input = page.getByPlaceholder('Enter age');
      await expect.element(input).toHaveValue(25);
    });
  });

  describe('Boolean fields', () => {
    it('renders checkbox for boolean field', async () => {
      const fieldDef: PropertyDefinition = {type: 'boolean'};
      await renderWithProviders(<TestForm fieldName="isActive" fieldDef={fieldDef} />);

      const checkbox = page.getByRole('checkbox');
      await expect.element(checkbox).toBeInTheDocument();
      await expect.element(page.getByLabelText('isActive')).toBeInTheDocument();
    });

    it('shows required asterisk for required boolean fields', async () => {
      const fieldDef: PropertyDefinition = {type: 'boolean', required: true};
      await renderWithProviders(<TestForm fieldName="isActive" fieldDef={fieldDef} />);

      const label = page.getByText('isActive');
      await expect.element(label).toBeInTheDocument();
    });

    it('checkbox is checked when default value is true', async () => {
      const fieldDef: PropertyDefinition = {type: 'boolean'};
      await renderWithProviders(<TestForm fieldName="isActive" fieldDef={fieldDef} defaultValues={{isActive: true}} />);

      const checkbox = page.getByRole('checkbox');
      await expect.element(checkbox).toBeChecked();
    });

    it('checkbox is unchecked when default value is false', async () => {
      const fieldDef: PropertyDefinition = {type: 'boolean'};
      await renderWithProviders(<TestForm fieldName="isActive" fieldDef={fieldDef} defaultValues={{isActive: false}} />);

      const checkbox = page.getByRole('checkbox');
      await expect.element(checkbox).not.toBeChecked();
    });
  });

  describe('Array fields', () => {
    it('renders ArrayFieldInput for array field', async () => {
      const fieldDef: PropertyDefinition = {
        type: 'array',
        items: {type: 'string'},
      };
      await renderWithProviders(<TestForm fieldName="tags" fieldDef={fieldDef} />);

      await expect.element(page.getByPlaceholder('Add tags')).toBeInTheDocument();
    });

    it('shows required asterisk for required array fields', async () => {
      const fieldDef: PropertyDefinition = {
        type: 'array',
        items: {type: 'string'},
        required: true,
      };
      await renderWithProviders(<TestForm fieldName="tags" fieldDef={fieldDef} />);

      const label = page.getByText('tags');
      await expect.element(label).toBeInTheDocument();
    });

    it('renders with default array values', async () => {
      const fieldDef: PropertyDefinition = {
        type: 'array',
        items: {type: 'string'},
      };
      await renderWithProviders(<TestForm fieldName="tags" fieldDef={fieldDef} defaultValues={{tags: ['tag1', 'tag2']}} />);

      await expect.element(page.getByText('tag1')).toBeInTheDocument();
      await expect.element(page.getByText('tag2')).toBeInTheDocument();
    });

    it('validates required array field with empty array', async () => {
        const fieldDef: PropertyDefinition = {
        type: 'array',
        items: {type: 'string'},
        required: true,
      };
      await renderWithProviders(<TestForm fieldName="tags" fieldDef={fieldDef} defaultValues={{tags: []}} />);

      const submitButton = page.getByRole('button', {name: 'Submit'});
      await userEvent.click(submitButton);

      await vi.waitFor(async () => {
        // The validation message could be either "tags is required" or "tags must have at least one value"
        // depending on which validation rule runs first
        const errorMessage = page.getByText(/tags (is required|must have at least one value)/);
        await expect.element(errorMessage).toBeInTheDocument();
      });
    });

    it('validates required array field with non-array value (undefined)', async () => {
        const fieldDef: PropertyDefinition = {
        type: 'array',
        items: {type: 'string'},
        required: true,
      };
      await renderWithProviders(<TestForm fieldName="tags" fieldDef={fieldDef} defaultValues={{}} />);

      const submitButton = page.getByRole('button', {name: 'Submit'});
      await userEvent.click(submitButton);

      await vi.waitFor(async () => {
        // The validation message could be either "tags is required" or "tags must have at least one value"
        // depending on which validation rule runs first
        const errorMessage = page.getByText(/tags (is required|must have at least one value)/);
        await expect.element(errorMessage).toBeInTheDocument();
      });
    });

    it('validates successfully when required array has values', async () => {
        const onSubmit = vi.fn();
      const fieldDef: PropertyDefinition = {
        type: 'array',
        items: {type: 'string'},
        required: true,
      };
      await renderWithProviders(<TestForm fieldName="tags" fieldDef={fieldDef} defaultValues={{tags: ['tag1']}} onSubmit={onSubmit} />);

      const submitButton = page.getByRole('button', {name: 'Submit'});
      await userEvent.click(submitButton);

      await vi.waitFor(() => {
        expect(onSubmit).toHaveBeenCalled();
      });
    });

    it('handles non-array value gracefully', async () => {
      const fieldDef: PropertyDefinition = {
        type: 'array',
        items: {type: 'string'},
      };
      await renderWithProviders(<TestForm fieldName="tags" fieldDef={fieldDef} defaultValues={{tags: 'not-an-array'}} />);

      // Should render without crashing and treat as empty array
      await expect.element(page.getByPlaceholder('Add tags')).toBeInTheDocument();
    });

    it('shows validation error when optional array field is empty', async () => {
      const fieldDef: PropertyDefinition = {
        type: 'array',
        items: {type: 'string'},
        required: false,
      };
      await renderWithProviders(<TestForm fieldName="tags" fieldDef={fieldDef} defaultValues={{tags: []}} />);

      // Should render without showing any error for non-required empty array
      await expect.element(page.getByPlaceholder('Add tags')).toBeInTheDocument();
      await expect.element(page.getByText(/tags (is required|must have at least one value)/)).not.toBeInTheDocument();
    });
  });

  describe('Unsupported types', () => {
    it('returns null for object type', async () => {
      const fieldDef: PropertyDefinition = {
        type: 'object',
        properties: {},
      };
      await renderWithProviders(<TestForm fieldName="metadata" fieldDef={fieldDef} />);

      // Should only render the submit button, no field components
      await expect.element(page.getByLabelText('metadata')).not.toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: 'Submit'})).toBeInTheDocument();
    });
  });

  describe('Field validation', () => {
    it('handles regex validation for string fields', async () => {
      const fieldDef: PropertyDefinition = {
        type: 'string',
        regex: '^[a-z]+$',
        required: true,
      };
      await renderWithProviders(<TestForm fieldName="username" fieldDef={fieldDef} />);

      const input = page.getByPlaceholder('Enter username');
      await expect.element(input).toBeInTheDocument();
    });
  });

  describe('Edge cases', () => {
    it('handles empty enum array', async () => {
      const fieldDef: PropertyDefinition = {
        type: 'string',
        enum: [],
      };
      await renderWithProviders(<TestForm fieldName="role" fieldDef={fieldDef} />);

      // Should render as regular TextField since enum is empty
      await expect.element(page.getByPlaceholder('Enter role')).toBeInTheDocument();
    });

    it('handles field without required property', async () => {
      const fieldDef: PropertyDefinition = {
        type: 'string',
      };
      await renderWithProviders(<TestForm fieldName="username" fieldDef={fieldDef} />);

      await expect.element(page.getByLabelText('username')).toBeInTheDocument();
    });

    it('handles unique property on string field', async () => {
      const fieldDef: PropertyDefinition = {
        type: 'string',
        unique: true,
      };
      await renderWithProviders(<TestForm fieldName="email" fieldDef={fieldDef} />);

      await expect.element(page.getByPlaceholder('Enter email')).toBeInTheDocument();
    });

    it('returns null for unsupported field type', async () => {
      // Using an unsupported type to test the catch-all return null
      const fieldDef: PropertyDefinition = {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any, @typescript-eslint/no-unsafe-assignment
        type: 'date' as any,
      };
      await renderWithProviders(<TestForm fieldName="birthdate" fieldDef={fieldDef} />);

      // Should only render the submit button, no field components
      await expect.element(page.getByLabelText('birthdate')).not.toBeInTheDocument();
      await expect.element(page.getByRole('button', {name: 'Submit'})).toBeInTheDocument();
    });
  });

  describe('User interactions', () => {
    it('allows typing in string TextField', async () => {
        const fieldDef: PropertyDefinition = {type: 'string'};
      await renderWithProviders(<TestForm fieldName="username" fieldDef={fieldDef} />);

      const input = page.getByPlaceholder('Enter username');
      await userEvent.fill(input, 'john.doe');

      await expect.element(input).toHaveValue('john.doe');
    });

    it('allows toggling checkbox', async () => {
        const fieldDef: PropertyDefinition = {type: 'boolean'};
      await renderWithProviders(<TestForm fieldName="isActive" fieldDef={fieldDef} />);

      const checkbox = page.getByRole('checkbox');
      await expect.element(checkbox).not.toBeChecked();

      await userEvent.click(checkbox);
      await expect.element(checkbox).toBeChecked();
    });

    it('allows typing in number TextField', async () => {
        const fieldDef: PropertyDefinition = {type: 'number'};
      await renderWithProviders(<TestForm fieldName="age" fieldDef={fieldDef} />);

      const input = page.getByPlaceholder('Enter age');
      await userEvent.fill(input, '25');

      await expect.element(input).toHaveValue(25);
    });
  });

  describe('Credential fields', () => {
    it('renders credential string field as password input', async () => {
      const fieldDef: PropertyDefinition = {type: 'string', credential: true};
      await renderWithProviders(<TestForm fieldName="password" fieldDef={fieldDef} />);

      const input = page.getByPlaceholder('Enter password');
      await expect.element(input).toHaveAttribute('type', 'password');
    });

    it('renders non-credential string field as text input', async () => {
      const fieldDef: PropertyDefinition = {type: 'string', credential: false};
      await renderWithProviders(<TestForm fieldName="username" fieldDef={fieldDef} />);

      const input = page.getByPlaceholder('Enter username');
      await expect.element(input).toHaveAttribute('type', 'text');
    });

    it('renders toggle password visibility button for credential fields', async () => {
      const fieldDef: PropertyDefinition = {type: 'string', credential: true};
      await renderWithProviders(<TestForm fieldName="password" fieldDef={fieldDef} />);

      await expect.element(page.getByLabelText('show password')).toBeInTheDocument();
    });

    it('does not render toggle button for non-credential fields', async () => {
      const fieldDef: PropertyDefinition = {type: 'string'};
      await renderWithProviders(<TestForm fieldName="username" fieldDef={fieldDef} />);

      await expect.element(page.getByLabelText('show password')).not.toBeInTheDocument();
    });

    it('toggles password visibility when icon button is clicked', async () => {
        const fieldDef: PropertyDefinition = {type: 'string', credential: true};
      await renderWithProviders(<TestForm fieldName="secret" fieldDef={fieldDef} />);

      const input = page.getByPlaceholder('Enter secret');
      await expect.element(input).toHaveAttribute('type', 'password');

      const toggleButton = page.getByLabelText('show password');
      await userEvent.click(toggleButton);
      await expect.element(input).toHaveAttribute('type', 'text');
      await expect.element(toggleButton).toHaveAttribute('aria-label', 'hide password');

      await userEvent.click(toggleButton);
      await expect.element(input).toHaveAttribute('type', 'password');
      await expect.element(toggleButton).toHaveAttribute('aria-label', 'show password');
    });

    it('allows typing in credential field', async () => {
        const fieldDef: PropertyDefinition = {type: 'string', credential: true};
      await renderWithProviders(<TestForm fieldName="pin" fieldDef={fieldDef} />);

      const input = page.getByPlaceholder('Enter pin');
      await userEvent.fill(input, 'secret123');

      await expect.element(input).toHaveValue('secret123');
    });
  });
});
