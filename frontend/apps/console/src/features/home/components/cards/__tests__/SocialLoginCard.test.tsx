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
import {render, screen} from '@thunder/test-utils';
import SocialLoginCard from '../SocialLoginCard';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, fallback?: string | object) => (typeof fallback === 'string' ? fallback : key),
  }),
}));

vi.mock('react-router', async () => {
  const actual = await vi.importActual<typeof import('react-router')>('react-router');
  return {
    ...actual,
    useNavigate: () => vi.fn(),
  };
});

describe('SocialLoginCard', () => {
  it('renders the card title', () => {
    render(<SocialLoginCard />);

    expect(screen.getByText('Social Integrations')).toBeInTheDocument();
  });

  it('renders the card description', () => {
    render(<SocialLoginCard />);

    expect(
      screen.getByText('Let users sign in with their favourite identity providers — Google, GitHub, and more.'),
    ).toBeInTheDocument();
  });

  it('renders the "Coming Soon" status badge', () => {
    render(<SocialLoginCard />);

    expect(screen.getByText('Coming Soon')).toBeInTheDocument();
  });

  it('does not render any action buttons (card is disabled)', () => {
    render(<SocialLoginCard />);

    expect(screen.queryByRole('button')).not.toBeInTheDocument();
  });
});
