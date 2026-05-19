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

import {render, screen, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import CertificateTypes from '../../../../../applications/constants/certificate-types';
import type {Agent} from '../../../../models/agent';
import CertificateSection from '../CertificateSection';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

describe('CertificateSection (agent)', () => {
  const mockAgent: Agent = {
    id: 'agent-1',
    ouId: 'ou-1',
    type: 'default',
    name: 'Test Agent',
    certificate: {type: CertificateTypes.NONE, value: ''},
  };

  const mockOnFieldChange = vi.fn();

  beforeEach(() => {
    mockOnFieldChange.mockClear();
  });

  describe('Rendering', () => {
    it('renders the certificate section', () => {
      render(<CertificateSection agent={mockAgent} editedAgent={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.getByText('applications:edit.advanced.labels.certificate')).toBeInTheDocument();
      expect(screen.getByText('applications:edit.advanced.certificate.intro')).toBeInTheDocument();
    });

    it('renders the certificate type dropdown', () => {
      render(<CertificateSection agent={mockAgent} editedAgent={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.getByLabelText('applications:edit.advanced.labels.certificateType')).toBeInTheDocument();
    });

    it('does not render the value field when type is NONE', () => {
      render(<CertificateSection agent={mockAgent} editedAgent={{}} onFieldChange={mockOnFieldChange} />);

      expect(
        screen.queryByPlaceholderText('applications:edit.advanced.certificate.placeholder.jwks'),
      ).not.toBeInTheDocument();
    });

    it('renders the JWKS value field when type is JWKS', () => {
      const agent = {...mockAgent, certificate: {type: CertificateTypes.JWKS, value: ''}};
      render(<CertificateSection agent={agent} editedAgent={{}} onFieldChange={mockOnFieldChange} />);

      expect(
        screen.getByPlaceholderText('applications:edit.advanced.certificate.placeholder.jwks'),
      ).toBeInTheDocument();
    });

    it('renders the JWKS URI value field when type is JWKS_URI', () => {
      const agent = {...mockAgent, certificate: {type: CertificateTypes.JWKS_URI, value: ''}};
      render(<CertificateSection agent={agent} editedAgent={{}} onFieldChange={mockOnFieldChange} />);

      expect(
        screen.getByPlaceholderText('applications:edit.advanced.certificate.placeholder.jwksUri'),
      ).toBeInTheDocument();
    });
  });

  describe('Certificate Type Selection', () => {
    it('reflects the certificate type from the agent', () => {
      const agent = {...mockAgent, certificate: {type: CertificateTypes.JWKS, value: ''}};
      render(<CertificateSection agent={agent} editedAgent={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.getByRole('combobox')).toHaveValue('applications:edit.advanced.certificate.type.jwks');
    });

    it('prioritizes editedAgent.certificate.type over agent.certificate.type', () => {
      const agent = {...mockAgent, certificate: {type: CertificateTypes.NONE, value: ''}};
      render(
        <CertificateSection
          agent={agent}
          editedAgent={{certificate: {type: CertificateTypes.JWKS, value: ''}}}
          onFieldChange={mockOnFieldChange}
        />,
      );

      expect(screen.getByRole('combobox')).toHaveValue('applications:edit.advanced.certificate.type.jwks');
    });

    it('calls onFieldChange when the certificate type changes', async () => {
      const user = userEvent.setup();
      render(<CertificateSection agent={mockAgent} editedAgent={{}} onFieldChange={mockOnFieldChange} />);

      const autocomplete = screen.getByRole('combobox');
      await user.click(autocomplete);

      const listbox = screen.getByRole('listbox');
      await user.click(within(listbox).getByText('applications:edit.advanced.certificate.type.jwks'));

      expect(mockOnFieldChange).toHaveBeenCalledWith('certificate', {
        type: CertificateTypes.JWKS,
        value: '',
      });
    });

    it('preserves certificate value when type changes', async () => {
      const user = userEvent.setup();
      const agent = {...mockAgent, certificate: {type: CertificateTypes.JWKS, value: 'existing'}};
      render(<CertificateSection agent={agent} editedAgent={{}} onFieldChange={mockOnFieldChange} />);

      await user.click(screen.getByRole('combobox'));
      const listbox = screen.getByRole('listbox');
      await user.click(within(listbox).getByText('applications:edit.advanced.certificate.type.jwksUri'));

      expect(mockOnFieldChange).toHaveBeenCalledWith('certificate', {
        type: CertificateTypes.JWKS_URI,
        value: 'existing',
      });
    });
  });

  describe('Certificate Value Input', () => {
    it('reflects certificate value from the agent', () => {
      const agent = {...mockAgent, certificate: {type: CertificateTypes.JWKS, value: 'jwks-value'}};
      render(<CertificateSection agent={agent} editedAgent={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.getByPlaceholderText('applications:edit.advanced.certificate.placeholder.jwks')).toHaveValue(
        'jwks-value',
      );
    });

    it('prioritizes editedAgent.certificate.value over agent value', () => {
      const agent = {...mockAgent, certificate: {type: CertificateTypes.JWKS, value: 'old'}};
      render(
        <CertificateSection
          agent={agent}
          editedAgent={{certificate: {type: CertificateTypes.JWKS, value: 'new'}}}
          onFieldChange={mockOnFieldChange}
        />,
      );

      expect(screen.getByPlaceholderText('applications:edit.advanced.certificate.placeholder.jwks')).toHaveValue('new');
    });

    it('calls onFieldChange when the value changes', async () => {
      const user = userEvent.setup({delay: null});
      const agent = {...mockAgent, certificate: {type: CertificateTypes.JWKS, value: ''}};
      render(<CertificateSection agent={agent} editedAgent={{}} onFieldChange={mockOnFieldChange} />);

      const valueInput = screen.getByPlaceholderText('applications:edit.advanced.certificate.placeholder.jwks');
      await user.type(valueInput, 'X');

      expect(mockOnFieldChange).toHaveBeenCalledWith(
        'certificate',
        expect.objectContaining({type: CertificateTypes.JWKS}),
      );
    });
  });

  describe('Edge Cases', () => {
    it('handles missing certificate on the agent', () => {
      const agent = {...mockAgent};
      delete (agent as Partial<Agent>).certificate;

      render(<CertificateSection agent={agent} editedAgent={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.getByRole('combobox')).toHaveValue('applications:edit.advanced.certificate.type.none');
    });

    it('renders the JWKS value input as a multiline textarea with 3 rows', () => {
      const agent = {...mockAgent, certificate: {type: CertificateTypes.JWKS, value: ''}};
      render(<CertificateSection agent={agent} editedAgent={{}} onFieldChange={mockOnFieldChange} />);

      expect(screen.getByPlaceholderText('applications:edit.advanced.certificate.placeholder.jwks')).toHaveAttribute(
        'rows',
        '3',
      );
    });
  });
});
