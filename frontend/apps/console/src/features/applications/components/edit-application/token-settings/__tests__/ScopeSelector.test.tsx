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

import {render, screen, within, act, fireEvent} from '@testing-library/react';
import {describe, expect, it, vi} from 'vitest';
import ScopeSelector from '../ScopeSelector';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({t: (key: string, fallback?: string) => fallback ?? key}),
}));

const KNOWN_SCOPES = ['openid', 'profile', 'email', 'phone', 'address', 'groups', 'roles'];

describe('ScopeSelector', () => {
  it('renders active scopes as chips', () => {
    render(<ScopeSelector scopes={['openid', 'profile']} onScopesChange={vi.fn()} />);

    expect(screen.getByText('openid')).toBeInTheDocument();
    expect(screen.getByText('profile')).toBeInTheDocument();
  });

  it('renders inactive known scopes as suggested chips', () => {
    render(<ScopeSelector scopes={['openid']} onScopesChange={vi.fn()} />);

    // openid is active, remaining known scopes should appear as suggestions
    KNOWN_SCOPES.filter((s) => s !== 'openid').forEach((scope) => {
      expect(screen.getByText(scope)).toBeInTheDocument();
    });
  });

  it('clicking a suggested scope chip calls onScopesChange with the scope added', () => {
    const onScopesChange = vi.fn();

    render(<ScopeSelector scopes={[]} onScopesChange={onScopesChange} />);

    fireEvent.click(screen.getByText('email'));

    expect(onScopesChange).toHaveBeenCalledWith(['email']);
  });

  it('clicking the delete button on an active scope calls onScopesChange without that scope after delay', () => {
    vi.useFakeTimers();
    const onScopesChange = vi.fn();

    render(<ScopeSelector scopes={['openid', 'profile']} onScopesChange={onScopesChange} />);

    // Find the active section and locate the openid chip root
    const activeSection = screen.getByText('Active').closest('div')!.parentElement!;
    const openidChip = within(activeSection).getByText('openid').closest('.MuiChip-root')!;

    // Click the delete icon SVG directly (has class MuiChip-deleteIcon and its own onClick → onDelete)
    const deleteIcon = openidChip.querySelector('.MuiChip-deleteIcon')!;
    fireEvent.click(deleteIcon);

    act(() => {
      vi.advanceTimersByTime(600);
    });

    expect(onScopesChange).toHaveBeenCalledWith(['profile']);
    vi.useRealTimers();
  });

  it('adding a custom scope via text input and Add button calls onScopesChange after delay', () => {
    vi.useFakeTimers();
    const onScopesChange = vi.fn();

    render(<ScopeSelector scopes={[]} onScopesChange={onScopesChange} />);

    const input = screen.getByPlaceholderText('e.g. custom:read');
    fireEvent.change(input, {target: {value: 'custom:read'}});

    fireEvent.click(screen.getByRole('button', {name: 'Add'}));

    act(() => {
      vi.advanceTimersByTime(600);
    });

    expect(onScopesChange).toHaveBeenCalledWith(['custom:read']);
    vi.useRealTimers();
  });

  it('shows an error when adding an empty scope name', () => {
    render(<ScopeSelector scopes={[]} onScopesChange={vi.fn()} />);

    fireEvent.click(screen.getByRole('button', {name: 'Add'}));

    expect(screen.getByText('Scope name cannot be empty')).toBeInTheDocument();
  });

  it('shows an error when adding a scope name with spaces', () => {
    render(<ScopeSelector scopes={[]} onScopesChange={vi.fn()} />);

    const input = screen.getByPlaceholderText('e.g. custom:read');
    fireEvent.change(input, {target: {value: 'has space'}});

    fireEvent.click(screen.getByRole('button', {name: 'Add'}));

    expect(screen.getByText('Scope name must not contain spaces')).toBeInTheDocument();
  });

  it('shows an error when adding a duplicate scope', () => {
    render(<ScopeSelector scopes={['openid']} onScopesChange={vi.fn()} />);

    const input = screen.getByPlaceholderText('e.g. custom:read');
    fireEvent.change(input, {target: {value: 'openid'}});

    fireEvent.click(screen.getByRole('button', {name: 'Add'}));

    expect(screen.getByText('This scope is already added')).toBeInTheDocument();
  });
});
