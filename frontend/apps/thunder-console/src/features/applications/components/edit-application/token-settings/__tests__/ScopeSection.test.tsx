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

import {describe, expect, it, vi, beforeEach} from 'vitest';
import {render, screen, fireEvent} from '@testing-library/react';
import ScopeSection from '../ScopeSection';
import type {ScopeClaims} from '../../../../models/oauth';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({t: (key: string, fallback?: string) => fallback ?? key}),
}));

vi.mock('../ScopeSelector', () => ({
  default: ({scopes, onScopesChange}: {scopes: string[]; onScopesChange: (s: string[]) => void}) => (
    <div>
      <div data-testid="scope-selector-scopes">{scopes.join(',')}</div>
      <button type="button" data-testid="remove-scope" onClick={() => onScopesChange(scopes.slice(1))}>
        remove-first
      </button>
      <button type="button" data-testid="add-scope" onClick={() => onScopesChange([...scopes, 'newscope'])}>
        add-scope
      </button>
    </div>
  ),
}));

vi.mock('../ScopeMapper', () => ({
  default: ({scopes}: {scopes: string[]}) => (
    <div>
      <div data-testid="scope-mapper-scopes">{scopes.join(',')}</div>
    </div>
  ),
}));

vi.mock('../../../../../../components/SettingsCard', () => ({
  default: ({title, description, children}: {title: string; description: string; children: React.ReactNode}) => (
    <div>
      <h2>{title}</h2>
      <p>{description}</p>
      {children}
    </div>
  ),
}));

const defaultProps = {
  scopes: ['openid', 'profile'],
  scopeClaims: {openid: ['email'], profile: ['given_name']} as ScopeClaims,
  userAttributes: ['email', 'given_name'],
  isLoadingUserAttributes: false,
  onScopesChange: vi.fn(),
  onScopeClaimsChange: vi.fn(),
};

describe('ScopeSection', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders the settings card title and description', () => {
    render(<ScopeSection {...defaultProps} />);

    expect(screen.getByText('Scopes & User Attribute Mappings')).toBeInTheDocument();
    expect(
      screen.getByText('Configure the OAuth2 scopes and the user attributes exposed for each scope'),
    ).toBeInTheDocument();
  });

  it('renders the attribute mapping title and hint', () => {
    render(<ScopeSection {...defaultProps} />);

    expect(screen.getByText('Attribute Mapping')).toBeInTheDocument();
    expect(
      screen.getByText('Select a scope to configure which user attributes are exposed when it is requested.'),
    ).toBeInTheDocument();
  });

  it('passes the scopes prop to ScopeSelector', () => {
    render(<ScopeSection {...defaultProps} />);

    expect(screen.getByTestId('scope-selector-scopes').textContent).toBe('openid,profile');
  });

  it('passes the scopes prop to ScopeMapper', () => {
    render(<ScopeSection {...defaultProps} />);

    expect(screen.getByTestId('scope-mapper-scopes').textContent).toBe('openid,profile');
  });

  it('calls onScopesChange when a scope is added and does NOT call onScopeClaimsChange', () => {
    const onScopesChange = vi.fn();
    const onScopeClaimsChange = vi.fn();

    render(
      <ScopeSection {...defaultProps} onScopesChange={onScopesChange} onScopeClaimsChange={onScopeClaimsChange} />,
    );

    fireEvent.click(screen.getByTestId('add-scope'));

    expect(onScopesChange).toHaveBeenCalledWith(['openid', 'profile', 'newscope']);
    expect(onScopeClaimsChange).not.toHaveBeenCalled();
  });

  it('cleans up scope claims and calls onScopesChange when a scope is removed', () => {
    const onScopesChange = vi.fn();
    const onScopeClaimsChange = vi.fn();

    render(
      <ScopeSection
        {...defaultProps}
        scopes={['openid', 'profile']}
        scopeClaims={{openid: ['email'], profile: ['given_name']}}
        onScopesChange={onScopesChange}
        onScopeClaimsChange={onScopeClaimsChange}
      />,
    );

    // remove-first calls onScopesChange(['profile']), removing 'openid'
    fireEvent.click(screen.getByTestId('remove-scope'));

    expect(onScopesChange).toHaveBeenCalledWith(['profile']);
    expect(onScopeClaimsChange).toHaveBeenCalledWith({profile: ['given_name']});
  });

  it('removes only the deleted scope entry from scopeClaims, keeping others intact', () => {
    const onScopeClaimsChange = vi.fn();

    render(
      <ScopeSection
        {...defaultProps}
        scopes={['openid', 'profile', 'email']}
        scopeClaims={{openid: ['sub'], profile: ['given_name'], email: ['address']}}
        onScopeClaimsChange={onScopeClaimsChange}
      />,
    );

    // Removes 'openid' (first scope)
    fireEvent.click(screen.getByTestId('remove-scope'));

    const updatedClaims = onScopeClaimsChange.mock.calls[0][0] as ScopeClaims;
    expect(updatedClaims).not.toHaveProperty('openid');
    expect(updatedClaims).toHaveProperty('profile', ['given_name']);
    expect(updatedClaims).toHaveProperty('email', ['address']);
  });

  it('calls onScopeClaimsChange even when the removed scope has no existing claims', () => {
    const onScopeClaimsChange = vi.fn();

    render(
      <ScopeSection
        {...defaultProps}
        scopes={['openid', 'profile']}
        scopeClaims={{profile: ['given_name']}} // 'openid' has no claims
        onScopeClaimsChange={onScopeClaimsChange}
      />,
    );

    fireEvent.click(screen.getByTestId('remove-scope'));

    // Removing a scope with no claims still triggers the cleanup callback
    expect(onScopeClaimsChange).toHaveBeenCalledWith({profile: ['given_name']});
  });
});
