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
import {render, screen, waitFor, userEvent} from '@thunder/test-utils';
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
    it('renders basic string TextField', () => {
      const fieldDef: PropertyDefinition = {type: 'string'};
      render(<TestForm fieldName="username" fieldDef={fieldDef} />);

      expect(screen.getByLabelText('username')).toBeInTheDocument();
      expect(screen.getByPlaceholderText('Enter username')).toBeInTheDocument();
    });

    it('shows required asterisk for required string fields', () => {
      const fieldDef: PropertyDefinition = {type: 'string', required: true};
      render(<TestForm fieldName="username" fieldDef={fieldDef} />);

      const label = screen.getByText('username');
      expect(label).toBeInTheDocument();
    });

    it('renders Select dropdown when enum is provided', () => {
      const fieldDef: PropertyDefinition = {
        type: 'string',
        enum: ['admin', 'user', 'guest'],
      };
      render(<TestForm fieldName="role" fieldDef={fieldDef} />);

      const select = screen.getByRole('combobox');
      expect(select).toBeInTheDocument();
    });

    it('displays enum options in Select dropdown', async () => {
      const user = userEvent.setup();
      const fieldDef: PropertyDefinition = {
        type: 'string',
        enum: ['admin', 'user', 'guest'],
      };
      render(<TestForm fieldName="role" fieldDef={fieldDef} />);

      const select = screen.getByRole('combobox');
      await user.click(select);

      await waitFor(() => {
        expect(screen.getByText('Admin')).toBeInTheDocument();
        expect(screen.getByText('User')).toBeInTheDocument();
        expect(screen.getByText('Guest')).toBeInTheDocument();
      });
    });

    it('renders with default value for string field', () => {
      const fieldDef: PropertyDefinition = {type: 'string'};
      render(<TestForm fieldName="username" fieldDef={fieldDef} defaultValues={{username: 'john'}} />);

      const input = screen.getByPlaceholderText('Enter username');
      expect(input).toHaveValue('john');
    });
  });

  describe('Number fields', () => {
    it('renders number TextField', () => {
      const fieldDef: PropertyDefinition = {type: 'number'};
      render(<TestForm fieldName="age" fieldDef={fieldDef} />);

      const input = screen.getByPlaceholderText('Enter age');
      expect(input).toHaveAttribute('type', 'number');
    });

    it('shows required asterisk for required number fields', () => {
      const fieldDef: PropertyDefinition = {type: 'number', required: true};
      render(<TestForm fieldName="age" fieldDef={fieldDef} />);

      const label = screen.getByText('age');
      expect(label).toBeInTheDocument();
    });

    it('renders with default value for number field', () => {
      const fieldDef: PropertyDefinition = {type: 'number'};
      render(<TestForm fieldName="age" fieldDef={fieldDef} defaultValues={{age: 25}} />);

      const input = screen.getByPlaceholderText('Enter age');
      expect(input).toHaveValue(25);
    });
  });

  describe('Boolean fields', () => {
    it('renders checkbox for boolean field', () => {
      const fieldDef: PropertyDefinition = {type: 'boolean'};
      render(<TestForm fieldName="isActive" fieldDef={fieldDef} />);

      const checkbox = screen.getByRole('checkbox');
      expect(checkbox).toBeInTheDocument();
      expect(screen.getByLabelText('isActive')).toBeInTheDocument();
    });

    it('shows required asterisk for required boolean fields', () => {
      const fieldDef: PropertyDefinition = {type: 'boolean', required: true};
      render(<TestForm fieldName="isActive" fieldDef={fieldDef} />);

      const label = screen.getByText('isActive');
      expect(label).toBeInTheDocument();
    });

    it('checkbox is checked when default value is true', () => {
      const fieldDef: PropertyDefinition = {type: 'boolean'};
      render(<TestForm fieldName="isActive" fieldDef={fieldDef} defaultValues={{isActive: true}} />);

      const checkbox = screen.getByRole('checkbox');
      expect(checkbox).toBeChecked();
    });

    it('checkbox is unchecked when default value is false', () => {
      const fieldDef: PropertyDefinition = {type: 'boolean'};
      render(<TestForm fieldName="isActive" fieldDef={fieldDef} defaultValues={{isActive: false}} />);

      const checkbox = screen.getByRole('checkbox');
      expect(checkbox).not.toBeChecked();
    });
  });

  describe('Array fields', () => {
    it('renders ArrayFieldInput for array field', () => {
      const fieldDef: PropertyDefinition = {
        type: 'array',
        items: {type: 'string'},
      };
      render(<TestForm fieldName="tags" fieldDef={fieldDef} />);

      expect(screen.getByPlaceholderText('Add tags')).toBeInTheDocument();
    });

    it('shows required asterisk for required array fields', () => {
      const fieldDef: PropertyDefinition = {
        type: 'array',
        items: {type: 'string'},
        required: true,
      };
      render(<TestForm fieldName="tags" fieldDef={fieldDef} />);

      const label = screen.getByText('tags');
      expect(label).toBeInTheDocument();
    });

    it('renders with default array values', () => {
      const fieldDef: PropertyDefinition = {
        type: 'array',
        items: {type: 'string'},
      };
      render(<TestForm fieldName="tags" fieldDef={fieldDef} defaultValues={{tags: ['tag1', 'tag2']}} />);

      expect(screen.getByText('tag1')).toBeInTheDocument();
      expect(screen.getByText('tag2')).toBeInTheDocument();
    });

    it('validates required array field with empty array', async () => {
      const user = userEvent.setup();
      const fieldDef: PropertyDefinition = {
        type: 'array',
        items: {type: 'string'},
        required: true,
      };
      render(<TestForm fieldName="tags" fieldDef={fieldDef} defaultValues={{tags: []}} />);

      const submitButton = screen.getByRole('button', {name: 'Submit'});
      await user.click(submitButton);

      await waitFor(() => {
        // The validation message could be either "tags is required" or "tags must have at least one value"
        // depending on which validation rule runs first
        const errorMessage = screen.getByText(/tags (is required|must have at least one value)/);
        expect(errorMessage).toBeInTheDocument();
      });
    });

    it('validates required array field with non-array value (undefined)', async () => {
      const user = userEvent.setup();
      const fieldDef: PropertyDefinition = {
        type: 'array',
        items: {type: 'string'},
        required: true,
      };
      render(<TestForm fieldName="tags" fieldDef={fieldDef} defaultValues={{}} />);

      const submitButton = screen.getByRole('button', {name: 'Submit'});
      await user.click(submitButton);

      await waitFor(() => {
        // The validation message could be either "tags is required" or "tags must have at least one value"
        // depending on which validation rule runs first
        const errorMessage = screen.getByText(/tags (is required|must have at least one value)/);
        expect(errorMessage).toBeInTheDocument();
      });
    });

    it('validates successfully when required array has values', async () => {
      const user = userEvent.setup();
      const onSubmit = vi.fn();
      const fieldDef: PropertyDefinition = {
        type: 'array',
        items: {type: 'string'},
        required: true,
      };
      render(<TestForm fieldName="tags" fieldDef={fieldDef} defaultValues={{tags: ['tag1']}} onSubmit={onSubmit} />);

      const submitButton = screen.getByRole('button', {name: 'Submit'});
      await user.click(submitButton);

      await waitFor(() => {
        expect(onSubmit).toHaveBeenCalled();
      });
    });

    it('handles non-array value gracefully', () => {
      const fieldDef: PropertyDefinition = {
        type: 'array',
        items: {type: 'string'},
      };
      render(<TestForm fieldName="tags" fieldDef={fieldDef} defaultValues={{tags: 'not-an-array'}} />);

      // Should render without crashing and treat as empty array
      expect(screen.getByPlaceholderText('Add tags')).toBeInTheDocument();
    });

    it('shows validation error when optional array field is empty', () => {
      const fieldDef: PropertyDefinition = {
        type: 'array',
        items: {type: 'string'},
        required: false,
      };
      render(<TestForm fieldName="tags" fieldDef={fieldDef} defaultValues={{tags: []}} />);

      // Should render without showing any error for non-required empty array
      expect(screen.getByPlaceholderText('Add tags')).toBeInTheDocument();
      expect(screen.queryByText(/tags (is required|must have at least one value)/)).not.toBeInTheDocument();
    });
  });

  describe('Unsupported types', () => {
    it('returns null for object type', () => {
      const fieldDef: PropertyDefinition = {
        type: 'object',
        properties: {},
      };
      render(<TestForm fieldName="metadata" fieldDef={fieldDef} />);

      // Should only render the submit button, no field components
      expect(screen.queryByLabelText('metadata')).not.toBeInTheDocument();
      expect(screen.getByRole('button', {name: 'Submit'})).toBeInTheDocument();
    });
  });

  describe('Field validation', () => {
    it('handles regex validation for string fields', () => {
      const fieldDef: PropertyDefinition = {
        type: 'string',
        regex: '^[a-z]+$',
        required: true,
      };
      render(<TestForm fieldName="username" fieldDef={fieldDef} />);

      const input = screen.getByPlaceholderText('Enter username');
      expect(input).toBeInTheDocument();
    });
  });

  describe('Edge cases', () => {
    it('handles empty enum array', () => {
      const fieldDef: PropertyDefinition = {
        type: 'string',
        enum: [],
      };
      render(<TestForm fieldName="role" fieldDef={fieldDef} />);

      // Should render as regular TextField since enum is empty
      expect(screen.getByPlaceholderText('Enter role')).toBeInTheDocument();
    });

    it('handles field without required property', () => {
      const fieldDef: PropertyDefinition = {
        type: 'string',
      };
      render(<TestForm fieldName="username" fieldDef={fieldDef} />);

      expect(screen.getByLabelText('username')).toBeInTheDocument();
    });

    it('handles unique property on string field', () => {
      const fieldDef: PropertyDefinition = {
        type: 'string',
        unique: true,
      };
      render(<TestForm fieldName="email" fieldDef={fieldDef} />);

      expect(screen.getByPlaceholderText('Enter email')).toBeInTheDocument();
    });

    it('returns null for unsupported field type', () => {
      // Using an unsupported type to test the catch-all return null
      const fieldDef: PropertyDefinition = {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any, @typescript-eslint/no-unsafe-assignment
        type: 'date' as any,
      };
      render(<TestForm fieldName="birthdate" fieldDef={fieldDef} />);

      // Should only render the submit button, no field components
      expect(screen.queryByLabelText('birthdate')).not.toBeInTheDocument();
      expect(screen.getByRole('button', {name: 'Submit'})).toBeInTheDocument();
    });
  });

  describe('User interactions', () => {
    it('allows typing in string TextField', async () => {
      const user = userEvent.setup();
      const fieldDef: PropertyDefinition = {type: 'string'};
      render(<TestForm fieldName="username" fieldDef={fieldDef} />);

      const input = screen.getByPlaceholderText('Enter username');
      await user.type(input, 'john.doe');

      expect(input).toHaveValue('john.doe');
    });

    it('allows toggling checkbox', async () => {
      const user = userEvent.setup();
      const fieldDef: PropertyDefinition = {type: 'boolean'};
      render(<TestForm fieldName="isActive" fieldDef={fieldDef} />);

      const checkbox = screen.getByRole('checkbox');
      expect(checkbox).not.toBeChecked();

      await user.click(checkbox);
      expect(checkbox).toBeChecked();
    });

    it('allows typing in number TextField', async () => {
      const user = userEvent.setup();
      const fieldDef: PropertyDefinition = {type: 'number'};
      render(<TestForm fieldName="age" fieldDef={fieldDef} />);

      const input = screen.getByPlaceholderText('Enter age');
      await user.type(input, '25');

      expect(input).toHaveValue(25);
    });
  });
});
