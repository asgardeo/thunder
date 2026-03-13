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

/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable @typescript-eslint/no-unsafe-assignment */
/* eslint-disable @typescript-eslint/no-unsafe-member-access */
/* eslint-disable @typescript-eslint/no-unsafe-call */

import {describe, it, expect, vi, beforeEach} from 'vitest';
import {fireEvent, render, screen} from '@testing-library/react';
import ConsentAdapter from '../ConsentAdapter';

vi.mock('@wso2/oxygen-ui', () => ({
  Box: ({children}: any) => <div>{children}</div>,
  Checkbox: ({checked, disabled, onChange}: any) => (
    <input type="checkbox" checked={checked} disabled={disabled} onChange={onChange} />
  ),
  Divider: () => <hr data-testid="consent-divider" />,
  FormControlLabel: ({control, label}: any) => (
    <div>
      {control}
      {label}
    </div>
  ),
  Typography: ({children}: any) => <span>{children}</span>,
}));

vi.mock('@asgardeo/react', () => ({
  Consent: ({consentData, formValues, onInputChange, children}: any) => {
    let purposes: any[] = [];
    if (Array.isArray(consentData)) {
      purposes = consentData;
    } else if (consentData && Array.isArray(consentData.purposes)) {
      purposes = consentData.purposes;
    }

    return <div data-testid="sdk-consent">{children?.({purposes, formValues, onInputChange})}</div>;
  },
  ConsentCheckboxList: ({variant, purpose, formValues, onInputChange, children}: any) => {
    const attributes: string[] = variant === 'ESSENTIAL' ? (purpose.essential ?? []) : (purpose.optional ?? []);

    const isChecked = (attrName: string): boolean => {
      if (variant === 'ESSENTIAL') return true;
      const key = `__consent_opt__${purpose.purpose_id}__${attrName}`;
      return formValues[key] !== 'false';
    };

    const handleChange = (attrName: string, checked: boolean): void => {
      const key = `__consent_opt__${purpose.purpose_id}__${attrName}`;
      onInputChange(key, String(checked));
    };

    return <div>{children?.({attributes, isChecked, handleChange, variant})}</div>;
  },
}));

const mockOnInputChange = vi.fn();

const consentData = {
  purposes: [
    {
      description: 'Profile attributes for personalization',
      essential: ['email', 'mobileNumber'],
      optional: ['givenName', 'lastName'],
      purpose_id: 'purpose-1',
      purpose_name: 'Profile',
    },
    {
      description: 'Address details',
      essential: ['country'],
      optional: ['city'],
      purpose_id: 'purpose-2',
      purpose_name: 'Address',
    },
  ],
};

describe('ConsentAdapter', () => {
  beforeEach(() => {
    mockOnInputChange.mockReset();
  });

  it('returns null when consentData is not provided', () => {
    const {container} = render(<ConsentAdapter formValues={{}} onInputChange={mockOnInputChange} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders essential and optional sections with attribute labels', () => {
    render(<ConsentAdapter consentData={consentData} formValues={{}} onInputChange={mockOnInputChange} />);

    expect(screen.getAllByText('Essential Attributes')).toHaveLength(2);
    expect(screen.getAllByText('Optional Attributes')).toHaveLength(2);

    expect(screen.getByText('email')).toBeInTheDocument();
    expect(screen.getByText('mobileNumber')).toBeInTheDocument();
    expect(screen.getByText('givenName')).toBeInTheDocument();
    expect(screen.getByText('lastName')).toBeInTheDocument();

    expect(screen.getAllByTestId('consent-divider')).toHaveLength(1);
  });

  it('renders essential checkboxes as disabled', () => {
    render(<ConsentAdapter consentData={consentData} formValues={{}} onInputChange={mockOnInputChange} />);

    const checkboxes = screen.getAllByRole('checkbox');
    // The first two checkboxes are essential attributes from the first purpose.
    expect(checkboxes[0]).toBeDisabled();
    expect(checkboxes[1]).toBeDisabled();
  });

  it('tracks optional checkbox state from form values and emits changes', () => {
    const formValues = {
      '__consent_opt__purpose-1__givenName': 'false',
      '__consent_opt__purpose-1__lastName': 'true',
    };

    render(<ConsentAdapter consentData={consentData} formValues={formValues} onInputChange={mockOnInputChange} />);

    const checkboxes = screen.getAllByRole('checkbox');
    const givenNameCheckbox = checkboxes[2];
    const lastNameCheckbox = checkboxes[3];

    expect(givenNameCheckbox).not.toBeChecked();
    expect(lastNameCheckbox).toBeChecked();

    fireEvent.click(lastNameCheckbox);
    expect(mockOnInputChange).toHaveBeenCalledWith('__consent_opt__purpose-1__lastName', 'false');
  });
});
