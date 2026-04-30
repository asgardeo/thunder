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

import {render, screen} from '@thunder/test-utils';
import type {ReactNode} from 'react';
import {describe, it, expect, vi} from 'vitest';
import UserAddPage from '../UserAddPage';

const mockLoggerWarn = vi.fn();

vi.mock('@thunder/logger/react', () => ({
  useLogger: () => ({
    info: vi.fn(),
    warn: mockLoggerWarn,
    error: vi.fn(),
    debug: vi.fn(),
  }),
}));

vi.mock('../UserInvitePage', () => ({
  default: ({onFlowUnavailable}: {onFlowUnavailable?: (error: unknown) => void}) => (
    <button type="button" onClick={() => onFlowUnavailable?.({code: 'FLM-1003'})}>
      Trigger Flow Fallback
    </button>
  ),
}));

vi.mock('../UserCreatePage', () => ({
  default: () => <div>Manual Create User Page</div>,
}));

vi.mock('../../contexts/UserCreate/UserCreateProvider', () => ({
  default: ({children}: {children: ReactNode}) => <div data-testid="user-create-provider">{children}</div>,
}));

describe('UserAddPage', () => {
  it('renders the flow-based add user page by default', () => {
    render(<UserAddPage />);

    expect(screen.getByRole('button', {name: 'Trigger Flow Fallback'})).toBeInTheDocument();
    expect(screen.queryByText('Manual Create User Page')).not.toBeInTheDocument();
  });

  it('falls back to the manual create page when onboarding flow is unavailable', async () => {
    render(<UserAddPage />);

    screen.getByRole('button', {name: 'Trigger Flow Fallback'}).click();

    expect(await screen.findByText('Manual Create User Page')).toBeInTheDocument();
    expect(screen.getByTestId('user-create-provider')).toBeInTheDocument();
    expect(mockLoggerWarn).toHaveBeenCalledWith(
      'User onboarding flow unavailable. Falling back to manual create user flow.',
      expect.objectContaining({error: {code: 'FLM-1003'}}),
    );
  });
});
