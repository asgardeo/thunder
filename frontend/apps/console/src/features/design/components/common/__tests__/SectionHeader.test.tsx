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
import {Palette} from '@wso2/oxygen-ui-icons-react';
import SectionHeader from '../SectionHeader';

vi.mock('react-i18next', async () => {
  const actual = await vi.importActual<typeof import('react-i18next')>('react-i18next');
  return {
    ...actual,
    useTranslation: () => ({t: (key: string, fallback?: string | Record<string, unknown>) => (typeof fallback === 'string' ? fallback : key)}),
  };
});

describe('SectionHeader', () => {
  describe('Rendering', () => {
    it('renders the title', () => {
      render(<SectionHeader title="Themes" count={5} icon={<Palette />} />);
      expect(screen.getByText('Themes')).toBeInTheDocument();
    });

    it('renders the count', () => {
      render(<SectionHeader title="Themes" count={5} icon={<Palette />} />);
      expect(screen.getByText('5')).toBeInTheDocument();
    });

    it('renders zero count', () => {
      render(<SectionHeader title="Themes" count={0} icon={<Palette />} />);
      expect(screen.getByText('0')).toBeInTheDocument();
    });

    it('renders icon', () => {
      const {container} = render(<SectionHeader title="Themes" count={3} icon={<Palette />} />);
      expect(container.querySelector('svg')).toBeInTheDocument();
    });
  });

  describe('Optional action prop', () => {
    it('renders action element when provided', () => {
      const action = <button type="button">Add</button>;
      render(<SectionHeader title="Themes" count={3} icon={<Palette />} action={action} />);
      expect(screen.getByText('Add')).toBeInTheDocument();
    });

    it('does not render action area when not provided', () => {
      render(<SectionHeader title="Themes" count={3} icon={<Palette />} />);
      expect(screen.queryByRole('button')).not.toBeInTheDocument();
    });
  });

  describe('comingSoon prop', () => {
    it('renders "Coming Soon" badge when comingSoon is true', () => {
      render(<SectionHeader title="Themes" count={3} icon={<Palette />} comingSoon />);
      expect(screen.getByText(/coming soon/i)).toBeInTheDocument();
    });

    it('does not render "Coming Soon" badge by default', () => {
      render(<SectionHeader title="Themes" count={3} icon={<Palette />} />);
      expect(screen.queryByText(/coming soon/i)).not.toBeInTheDocument();
    });

    it('does not render "Coming Soon" badge when comingSoon is false', () => {
      render(<SectionHeader title="Themes" count={3} icon={<Palette />} comingSoon={false} />);
      expect(screen.queryByText(/coming soon/i)).not.toBeInTheDocument();
    });
  });
});
