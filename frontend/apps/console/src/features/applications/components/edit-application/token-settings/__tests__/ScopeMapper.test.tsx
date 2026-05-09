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
import {describe, expect, it, vi, beforeEach} from 'vitest';
import ScopeMapper from '../ScopeMapper';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({t: (key: string, fallback?: string) => fallback ?? key}),
}));

vi.mock('../../../../constants/token-constants', () => ({
  default: {
    DEFAULT_TOKEN_ATTRIBUTES: ['sub', 'iss', 'aud', 'exp', 'iat'],
    ADDITIONAL_USER_ATTRIBUTES: [],
  },
}));

const defaultProps = {
  scopes: ['openid', 'profile'],
  scopeClaims: {},
  userAttributes: ['email', 'username', 'given_name'],
  isLoadingUserAttributes: false,
  onScopeClaimsChange: vi.fn(),
};

describe('ScopeMapper', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders an info alert when scopes list is empty', () => {
    render(
      <ScopeMapper
        scopes={[]}
        scopeClaims={{}}
        userAttributes={[]}
        isLoadingUserAttributes={false}
        onScopeClaimsChange={vi.fn()}
      />,
    );

    expect(screen.getByText('Add at least one scope above to start mapping attributes.')).toBeInTheDocument();
  });

  it('renders scope names in the left panel when scopes exist', () => {
    render(<ScopeMapper {...defaultProps} />);

    expect(screen.getByText('openid')).toBeInTheDocument();
    expect(screen.getByText('profile')).toBeInTheDocument();
  });

  it('clicking a scope in the left panel selects it and shows its mapping panel', async () => {
    const user = userEvent.setup();
    render(<ScopeMapper {...defaultProps} />);

    const profileScope = screen.getByRole('button', {name: /profile/i});
    await user.click(profileScope);

    expect(screen.getByText('Mapped Attributes')).toBeInTheDocument();
    expect(screen.getByText('Available Attributes')).toBeInTheDocument();
  });

  it('shows "Mapped Attributes" and "Available Attributes" labels when a scope is selected', () => {
    render(<ScopeMapper {...defaultProps} />);

    // First scope is auto-selected on mount
    expect(screen.getByText('Mapped Attributes')).toBeInTheDocument();
    expect(screen.getByText('Available Attributes')).toBeInTheDocument();
  });

  it('clicking an available attribute chip calls onScopeClaimsChange with it added to the selected scope', async () => {
    const onScopeClaimsChange = vi.fn();
    const user = userEvent.setup();

    render(
      <ScopeMapper {...defaultProps} scopes={['openid']} scopeClaims={{}} onScopeClaimsChange={onScopeClaimsChange} />,
    );

    // email is an available attribute (not in DEFAULT_TOKEN_ATTRIBUTES mock)
    const emailChip = screen.getByText('email');
    await user.click(emailChip);

    expect(onScopeClaimsChange).toHaveBeenCalledWith({openid: ['email']});
  });

  it('clicking delete on a mapped attribute calls onScopeClaimsChange with it removed', async () => {
    const onScopeClaimsChange = vi.fn();
    const user = userEvent.setup();

    render(
      <ScopeMapper
        {...defaultProps}
        scopes={['openid']}
        scopeClaims={{openid: ['email']}}
        onScopeClaimsChange={onScopeClaimsChange}
      />,
    );

    const mappedSection = screen.getByText('Mapped Attributes').closest('div')!.parentElement!;
    const emailChipRoot = within(mappedSection).getByText('email').closest('.MuiChip-root')!;
    const deleteIcon = emailChipRoot.querySelector('.MuiChip-deleteIcon')!;

    await user.click(deleteIcon);

    expect(onScopeClaimsChange).toHaveBeenCalledWith({openid: []});
  });

  it('shows loading message when isLoadingUserAttributes is true', () => {
    render(<ScopeMapper {...defaultProps} isLoadingUserAttributes />);

    expect(screen.getByText('Loading available attributes...')).toBeInTheDocument();
  });

  it('auto-selects the first scope on mount', () => {
    render(<ScopeMapper {...defaultProps} scopes={['openid', 'profile']} />);

    // Mapped Attributes panel visible means a scope was auto-selected
    expect(screen.getByText('Mapped Attributes')).toBeInTheDocument();
  });

  it('shows "all mapped" message when all available attributes are already mapped to the selected scope', () => {
    render(
      <ScopeMapper
        {...defaultProps}
        scopes={['openid']}
        scopeClaims={{openid: ['email', 'given_name', 'username']}}
        userAttributes={['email', 'given_name', 'username']}
      />,
    );

    expect(screen.getByText('All available attributes are already mapped to this scope')).toBeInTheDocument();
  });
});
