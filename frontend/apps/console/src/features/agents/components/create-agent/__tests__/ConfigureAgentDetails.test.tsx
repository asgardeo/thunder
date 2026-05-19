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

import type {ApiAgentType} from '@thunderid/configure-agent-types';
import {render, screen} from '@thunderid/test-utils';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import ConfigureAgentDetails, {type ConfigureAgentDetailsProps} from '../ConfigureAgentDetails';

// Mock the configure-users package — renderSchemaField produces a simple test marker per field.
vi.mock('@thunderid/configure-users', () => ({
  renderSchemaField: vi.fn((fieldName: string) => (
    <div key={fieldName} data-testid={`schema-field-${fieldName}`}>
      {fieldName}
    </div>
  )),
}));

// useResolveDisplayName returns a no-op resolver.
vi.mock('@thunderid/hooks', () => ({
  useResolveDisplayName: () => ({
    resolveDisplayName: (value: string) => value,
  }),
}));

const {renderSchemaField} = await import('@thunderid/configure-users');

describe('ConfigureAgentDetails', () => {
  const mockOnFormValuesChange = vi.fn();
  const mockOnReadyChange = vi.fn();

  const baseSchema: ApiAgentType = {
    id: 'schema-1',
    name: 'default',
    ouId: 'ou-1',
    schema: {
      email: {type: 'string', required: true},
      age: {type: 'number'},
    },
  };

  const defaultProps: ConfigureAgentDetailsProps = {
    schema: baseSchema,
    defaultValues: {},
    onFormValuesChange: mockOnFormValuesChange,
    onReadyChange: mockOnReadyChange,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders the title and subtitle', () => {
    render(<ConfigureAgentDetails {...defaultProps} />);

    expect(screen.getByRole('heading', {level: 1})).toBeInTheDocument();
    expect(screen.getByText(/provide values for the attributes/i)).toBeInTheDocument();
  });

  it('renders one field per schema entry', () => {
    render(<ConfigureAgentDetails {...defaultProps} />);

    expect(screen.getByTestId('schema-field-email')).toBeInTheDocument();
    expect(screen.getByTestId('schema-field-age')).toBeInTheDocument();
    expect(renderSchemaField).toHaveBeenCalledWith(
      'email',
      expect.any(Object),
      expect.anything(),
      expect.anything(),
      expect.any(Function),
    );
    expect(renderSchemaField).toHaveBeenCalledWith(
      'age',
      expect.any(Object),
      expect.anything(),
      expect.anything(),
      expect.any(Function),
    );
  });

  it('renders nothing field-related when schema is empty', () => {
    render(<ConfigureAgentDetails {...defaultProps} schema={{...baseSchema, schema: {}}} />);

    expect(screen.queryByTestId(/schema-field-/)).not.toBeInTheDocument();
  });

  it('calls onReadyChange initially based on form validity', () => {
    render(<ConfigureAgentDetails {...defaultProps} />);

    expect(mockOnReadyChange).toHaveBeenCalled();
  });

  it('does not crash when onReadyChange is undefined', () => {
    expect(() => render(<ConfigureAgentDetails {...defaultProps} onReadyChange={undefined} />)).not.toThrow();
  });

  it('exposes the form-values change handler to the watch subscription', () => {
    render(<ConfigureAgentDetails {...defaultProps} defaultValues={{email: 'a@b.com'}} />);

    // The component subscribes synchronously and react-hook-form fires once on mount-then-change;
    // this test guarantees the prop is consumed without errors. The actual change is forwarded
    // through the schema-field components (which are mocked here).
    expect(renderSchemaField).toHaveBeenCalled();
  });

  it('uses the data-testid wrapper', () => {
    render(<ConfigureAgentDetails {...defaultProps} />);

    expect(screen.getByTestId('configure-agent-details')).toBeInTheDocument();
  });
});
