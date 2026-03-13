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
import {render, screen, fireEvent} from '@thunder/test-utils';
import MFACard from '../MFACard';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, fallback?: string | object) => (typeof fallback === 'string' ? fallback : key),
  }),
}));

const mockNavigate = vi.fn();
vi.mock('react-router', async () => {
  const actual = await vi.importActual<typeof import('react-router')>('react-router');
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

describe('MFACard', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockNavigate.mockReturnValue(undefined);
  });

  it('renders the card title', () => {
    render(<MFACard />);

    expect(screen.getByText('Multi-factor Authentication')).toBeInTheDocument();
  });

  it('renders the card description', () => {
    render(<MFACard />);

    expect(
      screen.getByText('Protect users by enabling an additional verification factor to the sign-in process.'),
    ).toBeInTheDocument();
  });

  it('renders the primary action button', () => {
    render(<MFACard />);

    expect(screen.getByRole('button', {name: 'Configure Flows'})).toBeInTheDocument();
  });

  it('navigates to /flows when the primary button is clicked', () => {
    render(<MFACard />);

    fireEvent.click(screen.getByRole('button', {name: 'Configure Flows'}));

    expect(mockNavigate).toHaveBeenCalledWith('/flows');
  });

  it('does not render a feature status badge', () => {
    render(<MFACard />);

    expect(screen.queryByText('New')).not.toBeInTheDocument();
    expect(screen.queryByText('Coming Soon')).not.toBeInTheDocument();
  });
});
