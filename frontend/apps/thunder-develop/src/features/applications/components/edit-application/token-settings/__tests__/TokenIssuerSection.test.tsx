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
import TokenIssuerSection from '../TokenIssuerSection';

interface FormValues {
  validityPeriod: number;
  accessTokenValidity: number;
  idTokenValidity: number;
  issuer?: string;
}

// Mock the SettingsCard component
vi.mock('../../SettingsCard', () => ({
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
      issuer: '',
      ...defaultValues,
    },
  });

  return <div>{children({control, errors: formState.errors})}</div>;
}

describe('TokenIssuerSection', () => {
  describe('Rendering', () => {
    it('should render the settings card with title and description', () => {
      render(
        <TestWrapper>{({control, errors}) => <TokenIssuerSection control={control} errors={errors} />}</TestWrapper>,
      );

      expect(screen.getByTestId('card-title')).toHaveTextContent('Token Issuer');
      expect(screen.getByTestId('card-description')).toHaveTextContent(
        'Configure the issuer URL that will be included in both access and ID tokens',
      );
    });

    it('should render the issuer text field', () => {
      render(
        <TestWrapper>{({control, errors}) => <TokenIssuerSection control={control} errors={errors} />}</TestWrapper>,
      );

      expect(screen.getByLabelText('Issuer URL')).toBeInTheDocument();
    });

    it('should render with placeholder text', () => {
      render(
        <TestWrapper>{({control, errors}) => <TokenIssuerSection control={control} errors={errors} />}</TestWrapper>,
      );

      const input = screen.getByPlaceholderText('https://your-domain.com');
      expect(input).toBeInTheDocument();
    });

    it('should render with default empty value', () => {
      render(
        <TestWrapper>{({control, errors}) => <TokenIssuerSection control={control} errors={errors} />}</TestWrapper>,
      );

      const input = screen.getByLabelText('Issuer URL');
      expect(input).toHaveValue('');
    });

    it('should render with provided initial value', () => {
      render(
        <TestWrapper defaultValues={{issuer: 'https://test.com'}}>
          {({control, errors}) => <TokenIssuerSection control={control} errors={errors} />}
        </TestWrapper>,
      );

      const input = screen.getByLabelText('Issuer URL');
      expect(input).toHaveValue('https://test.com');
    });

    it('should render helper text when no error', () => {
      render(
        <TestWrapper>{({control, errors}) => <TokenIssuerSection control={control} errors={errors} />}</TestWrapper>,
      );

      expect(
        screen.getByText('The issuer URL will be included in the "iss" attribute of issued tokens'),
      ).toBeInTheDocument();
    });
  });

  describe('User Interaction', () => {
    it('should allow user to type in the issuer field', async () => {
      const user = userEvent.setup();

      render(
        <TestWrapper>{({control, errors}) => <TokenIssuerSection control={control} errors={errors} />}</TestWrapper>,
      );

      const input = screen.getByLabelText('Issuer URL');
      await user.type(input, 'https://myissuer.com');

      expect(input).toHaveValue('https://myissuer.com');
    });

    it('should allow user to clear the issuer field', async () => {
      const user = userEvent.setup();

      render(
        <TestWrapper defaultValues={{issuer: 'https://test.com'}}>
          {({control, errors}) => <TokenIssuerSection control={control} errors={errors} />}
        </TestWrapper>,
      );

      const input = screen.getByLabelText('Issuer URL');
      await user.clear(input);

      expect(input).toHaveValue('');
    });

    it('should allow user to update existing issuer value', async () => {
      const user = userEvent.setup();

      render(
        <TestWrapper defaultValues={{issuer: 'https://old.com'}}>
          {({control, errors}) => <TokenIssuerSection control={control} errors={errors} />}
        </TestWrapper>,
      );

      const input = screen.getByLabelText('Issuer URL');
      await user.clear(input);
      await user.type(input, 'https://new.com');

      expect(input).toHaveValue('https://new.com');
    });
  });

  describe('Error State', () => {
    it('should render error message when errors.issuer has message', () => {
      const mockErrors: FieldErrors<FormValues> = {
        issuer: {
          type: 'pattern',
          message: 'Please enter a valid URL',
        },
      };

      render(
        <TestWrapper>
          {({control}) => <TokenIssuerSection control={control} errors={mockErrors} />}
        </TestWrapper>,
      );

      expect(screen.getByText('Please enter a valid URL')).toBeInTheDocument();
    });

    it('should show error styling when error exists', () => {
      const mockErrors: FieldErrors<FormValues> = {
        issuer: {
          type: 'pattern',
          message: 'Invalid URL format',
        },
      };

      render(
        <TestWrapper>
          {({control}) => <TokenIssuerSection control={control} errors={mockErrors} />}
        </TestWrapper>,
      );

      // The error message should be displayed instead of the hint
      expect(screen.getByText('Invalid URL format')).toBeInTheDocument();
      expect(
        screen.queryByText('The issuer URL will be included in the "iss" attribute of issued tokens'),
      ).not.toBeInTheDocument();
    });

    it('should show default hint when no error message', () => {
      render(
        <TestWrapper>{({control, errors}) => <TokenIssuerSection control={control} errors={errors} />}</TestWrapper>,
      );

      expect(
        screen.getByText('The issuer URL will be included in the "iss" attribute of issued tokens'),
      ).toBeInTheDocument();
    });
  });
});
