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

import {describe, it, expect, vi} from 'vitest';
import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {useForm, type Control, type FieldErrors} from 'react-hook-form';
import TokenValidationSection from '../TokenValidationSection';

interface FormValues {
  validityPeriod: number;
  accessTokenValidity: number;
  idTokenValidity: number;
}

// Mock the SettingsCard component
vi.mock('../../../../../../components/SettingsCard', () => ({
  default: ({title, description, children}: {title: string; description: string; children: React.ReactNode}) => (
    <div data-testid="settings-card">
      <div data-testid="card-title">{title}</div>
      <div data-testid="card-description">{description}</div>
      {children}
    </div>
  ),
}));

// Wrapper component for testing with react-hook-form
function TestWrapper({
  children,
  defaultValues = {},
}: {
  children: (props: {control: Control<FormValues>; errors: FieldErrors<FormValues>}) => React.ReactNode;
  defaultValues?: Partial<FormValues>;
}) {
  const {control, formState} = useForm<FormValues>({
    defaultValues: {
      validityPeriod: 3600,
      accessTokenValidity: 3600,
      idTokenValidity: 3600,
      ...defaultValues,
    },
  });

  return <>{children({control, errors: formState.errors})}</>;
}

describe('TokenValidationSection', () => {
  describe('Rendering with tokenType="shared"', () => {
    it('should render the settings card with correct title and description', () => {
      render(
        <TestWrapper>
          {({control, errors}) => <TokenValidationSection control={control} errors={errors} tokenType="shared" />}
        </TestWrapper>,
      );

      expect(screen.getByTestId('card-title')).toHaveTextContent('Token Validation');
      expect(screen.getByTestId('card-description')).toHaveTextContent(
        'Configure token validation settings such as audience and validity',
      );
    });

    it('should render the validity period field with default value', () => {
      render(
        <TestWrapper>
          {({control, errors}) => <TokenValidationSection control={control} errors={errors} tokenType="shared" />}
        </TestWrapper>,
      );

      const input = screen.getByLabelText('Token Validity');
      expect(input).toBeInTheDocument();
      expect(input).toHaveValue(3600);
    });

    it('should render number input type', () => {
      render(
        <TestWrapper>
          {({control, errors}) => <TokenValidationSection control={control} errors={errors} tokenType="shared" />}
        </TestWrapper>,
      );

      const input = screen.getByLabelText('Token Validity');
      expect((input as HTMLInputElement).type).toBe('number');
    });

    it('should render with custom initial value', () => {
      render(
        <TestWrapper defaultValues={{validityPeriod: 7200}}>
          {({control, errors}) => <TokenValidationSection control={control} errors={errors} tokenType="shared" />}
        </TestWrapper>,
      );

      const input = screen.getByLabelText('Token Validity');
      expect(input).toHaveValue(7200);
    });

    it('should render helper text when no error', () => {
      render(
        <TestWrapper>
          {({control, errors}) => <TokenValidationSection control={control} errors={errors} tokenType="shared" />}
        </TestWrapper>,
      );

      expect(screen.getByText('Token validity period in seconds (e.g., 3600 for 1 hour)')).toBeInTheDocument();
    });
  });

  describe('Rendering with tokenType="access"', () => {
    it('should render the settings card with access token title and description', () => {
      render(
        <TestWrapper>
          {({control, errors}) => <TokenValidationSection control={control} errors={errors} tokenType="access" />}
        </TestWrapper>,
      );

      expect(screen.getByTestId('card-title')).toHaveTextContent('Access Token Validation');
      expect(screen.getByTestId('card-description')).toHaveTextContent(
        'Configure how long access tokens remain valid before expiration',
      );
    });

    it('should render the access token validity field with default value', () => {
      render(
        <TestWrapper>
          {({control, errors}) => <TokenValidationSection control={control} errors={errors} tokenType="access" />}
        </TestWrapper>,
      );

      const input = screen.getByLabelText('Token Validity');
      expect(input).toHaveValue(3600);
    });

    it('should render with custom initial access token value', () => {
      render(
        <TestWrapper defaultValues={{accessTokenValidity: 1800}}>
          {({control, errors}) => <TokenValidationSection control={control} errors={errors} tokenType="access" />}
        </TestWrapper>,
      );

      const input = screen.getByLabelText('Token Validity');
      expect(input).toHaveValue(1800);
    });
  });

  describe('Rendering with tokenType="id"', () => {
    it('should render the settings card with ID token title and description', () => {
      render(
        <TestWrapper>
          {({control, errors}) => <TokenValidationSection control={control} errors={errors} tokenType="id" />}
        </TestWrapper>,
      );

      expect(screen.getByTestId('card-title')).toHaveTextContent('ID Token Validation');
      expect(screen.getByTestId('card-description')).toHaveTextContent(
        'Configure how long ID tokens remain valid before expiration',
      );
    });

    it('should render the ID token validity field with default value', () => {
      render(
        <TestWrapper>
          {({control, errors}) => <TokenValidationSection control={control} errors={errors} tokenType="id" />}
        </TestWrapper>,
      );

      const input = screen.getByLabelText('Token Validity');
      expect(input).toHaveValue(3600);
    });

    it('should render with custom initial ID token value', () => {
      render(
        <TestWrapper defaultValues={{idTokenValidity: 900}}>
          {({control, errors}) => <TokenValidationSection control={control} errors={errors} tokenType="id" />}
        </TestWrapper>,
      );

      const input = screen.getByLabelText('Token Validity');
      expect(input).toHaveValue(900);
    });
  });

  describe('User Interaction', () => {
    it('should allow user to type a new validity value', async () => {
      const user = userEvent.setup();

      render(
        <TestWrapper>
          {({control, errors}) => <TokenValidationSection control={control} errors={errors} tokenType="shared" />}
        </TestWrapper>,
      );

      const input = screen.getByLabelText('Token Validity');
      await user.clear(input);
      await user.type(input, '7200');

      expect(input).toHaveValue(7200);
    });

    it('should allow user to update existing value', async () => {
      const user = userEvent.setup();

      render(
        <TestWrapper defaultValues={{validityPeriod: 1800}}>
          {({control, errors}) => <TokenValidationSection control={control} errors={errors} tokenType="shared" />}
        </TestWrapper>,
      );

      const input = screen.getByLabelText('Token Validity');
      await user.clear(input);
      await user.type(input, '3600');

      expect(input).toHaveValue(3600);
    });

    it('should handle numeric input for access token', async () => {
      const user = userEvent.setup();

      render(
        <TestWrapper>
          {({control, errors}) => <TokenValidationSection control={control} errors={errors} tokenType="access" />}
        </TestWrapper>,
      );

      const input = screen.getByLabelText('Token Validity');
      await user.clear(input);
      await user.type(input, '5400');

      expect(input).toHaveValue(5400);
    });

    it('should handle numeric input for ID token', async () => {
      const user = userEvent.setup();

      render(
        <TestWrapper>
          {({control, errors}) => <TokenValidationSection control={control} errors={errors} tokenType="id" />}
        </TestWrapper>,
      );

      const input = screen.getByLabelText('Token Validity');
      await user.clear(input);
      await user.type(input, '1200');

      expect(input).toHaveValue(1200);
    });
  });
});
