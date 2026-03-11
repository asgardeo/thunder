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
import CertificateSection from '../CertificateSection';
import type {Application} from '../../../../models/application';
import CertificateTypes from '../../../../constants/certificate-types';

describe('CertificateSection', () => {
  const mockApplication: Application = {
    id: 'test-app-id',
    name: 'Test Application',
    description: 'Test Description',
    template: 'custom',
    certificate: {
      type: CertificateTypes.NONE,
      value: '',
    },
  } as Application;

  const mockOnFieldChange = vi.fn();

  beforeEach(() => {
    mockOnFieldChange.mockClear();
  });

  describe('Rendering', () => {
    it('should render the certificate section', async () => {
      await render(<CertificateSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      await expect.element(page.getByText('Certificate')).toBeInTheDocument();
      await expect.element(page.getByText('Configure certificates for client authentication and token encryption.')).toBeInTheDocument();
    });

    it('should render certificate type dropdown', async () => {
      await render(<CertificateSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      await expect.element(page.getByLabelText('Certificate Type')).toBeInTheDocument();
    });

    it('should not show value field when certificate type is NONE', async () => {
      await render(<CertificateSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      await expect.element(
        page.getByPlaceholder('Enter JWKS JSON'),
      ).not.toBeInTheDocument();
      await expect.element(
        page.getByPlaceholder('https://example.com/.well-known/jwks'),
      ).not.toBeInTheDocument();
    });

    it('should show JWKS value field when certificate type is JWKS', async () => {
      const appWithJwks = {
        ...mockApplication,
        certificate: {type: CertificateTypes.JWKS, value: ''},
      };

      await render(<CertificateSection application={appWithJwks} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      expect(
        page.getByPlaceholder('Enter JWKS JSON'),
      ).toBeInTheDocument();
      await expect.element(page.getByText('JSON Web Key Set')).toBeInTheDocument();
    });

    it('should show JWKS URI value field when certificate type is JWKS_URI', async () => {
      const appWithJwksUri = {
        ...mockApplication,
        certificate: {type: CertificateTypes.JWKS_URI, value: ''},
      };

      await render(<CertificateSection application={appWithJwksUri} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      expect(
        page.getByPlaceholder('https://example.com/.well-known/jwks'),
      ).toBeInTheDocument();
      await expect.element(page.getByText('URL to the JWKS endpoint')).toBeInTheDocument();
    });
  });

  describe('Certificate Type Selection', () => {
    it('should display current certificate type from application', async () => {
      const appWithJwks = {
        ...mockApplication,
        certificate: {type: CertificateTypes.JWKS, value: ''},
      };

      await render(<CertificateSection application={appWithJwks} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const input = page.getByRole('combobox');
      expect(input).toHaveValue('JWKS (Inline JSON Web Key Set)');
    });

    it('should prioritize editedApp certificate type over application', async () => {
      const appWithNone = {
        ...mockApplication,
        certificate: {type: CertificateTypes.NONE, value: ''},
      };

      await render(
        <CertificateSection
          application={appWithNone}
          editedApp={{certificate: {type: CertificateTypes.JWKS, value: ''}}}
          onFieldChange={mockOnFieldChange}
        />,
      );

      const input = page.getByRole('combobox');
      expect(input).toHaveValue('JWKS (Inline JSON Web Key Set)');
    });

    it('should call onFieldChange when certificate type changes', async () => {
            await render(<CertificateSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const autocomplete = page.getByRole('combobox');
      await userEvent.click(autocomplete);

      const listbox = page.getByRole('listbox');
      const jwksOption = listbox.getByText('JWKS (Inline JSON Web Key Set)');
      await userEvent.click(jwksOption);

      expect(mockOnFieldChange).toHaveBeenCalledWith('certificate', {
        type: CertificateTypes.JWKS,
        value: '',
      });
    });

    it('should preserve certificate value when changing type', async () => {
            const appWithValue = {
        ...mockApplication,
        certificate: {type: CertificateTypes.JWKS, value: 'existing-jwks'},
      };

      await render(<CertificateSection application={appWithValue} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const autocomplete = page.getByRole('combobox');
      await userEvent.click(autocomplete);

      const listbox = page.getByRole('listbox');
      const jwksUriOption = listbox.getByText('JWKS URI (URL to JWKS endpoint)');
      await userEvent.click(jwksUriOption);

      expect(mockOnFieldChange).toHaveBeenCalledWith('certificate', {
        type: CertificateTypes.JWKS_URI,
        value: 'existing-jwks',
      });
    });
  });

  describe('Certificate Value Input', () => {
    it('should display current certificate value from application', async () => {
      const appWithValue = {
        ...mockApplication,
        certificate: {type: CertificateTypes.JWKS, value: 'test-jwks-value'},
      };

      await render(<CertificateSection application={appWithValue} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const valueInput = page.getByPlaceholder('Enter JWKS JSON');
      expect(valueInput).toHaveValue('test-jwks-value');
    });

    it('should prioritize editedApp certificate value over application', async () => {
      const appWithValue = {
        ...mockApplication,
        certificate: {type: CertificateTypes.JWKS, value: 'original-value'},
      };

      await render(
        <CertificateSection
          application={appWithValue}
          editedApp={{certificate: {type: CertificateTypes.JWKS, value: 'edited-value'}}}
          onFieldChange={mockOnFieldChange}
        />,
      );

      const valueInput = page.getByPlaceholder('Enter JWKS JSON');
      expect(valueInput).toHaveValue('edited-value');
    });

    it('should call onFieldChange when certificate value changes', async () => {
            const appWithJwks = {
        ...mockApplication,
        certificate: {type: CertificateTypes.JWKS, value: ''},
      };

      await render(<CertificateSection application={appWithJwks} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const valueInput = page.getByPlaceholder('Enter JWKS JSON');
      await userEvent.type(valueInput, 'new-value');

      expect(mockOnFieldChange).toHaveBeenCalled();
      // Check that the field was updated with the correct type
      expect(mockOnFieldChange).toHaveBeenCalledWith(
        'certificate',
        expect.objectContaining({
          type: CertificateTypes.JWKS,
        }),
      );
    });

    it('should preserve certificate type when changing value', async () => {
            const appWithJwksUri = {
        ...mockApplication,
        certificate: {type: CertificateTypes.JWKS_URI, value: 'https://example.com'},
      };

      await render(<CertificateSection application={appWithJwksUri} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const valueInput = page.getByPlaceholder('https://example.com/.well-known/jwks');
      await userEvent.clear(valueInput);
      await userEvent.type(valueInput, 'https://new-url.com');

      // Check that the field was updated with the correct type
      expect(mockOnFieldChange).toHaveBeenCalledWith(
        'certificate',
        expect.objectContaining({
          type: CertificateTypes.JWKS_URI,
        }),
      );
    });
  });

  describe('Edge Cases', () => {
    it('should handle missing certificate in application', async () => {
      const appWithoutCert = {...mockApplication};
      delete (appWithoutCert as Partial<Application>).certificate;

      await render(<CertificateSection application={appWithoutCert} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const input = page.getByRole('combobox');
      expect(input).toHaveValue('None');
    });

    it('should handle multiline JWKS input', async () => {
      const appWithJwks = {
        ...mockApplication,
        certificate: {type: CertificateTypes.JWKS, value: ''},
      };

      await render(<CertificateSection application={appWithJwks} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const valueInput = page.getByPlaceholder('Enter JWKS JSON');
      expect(valueInput).toHaveAttribute('rows', '3');
    });
  });
});
