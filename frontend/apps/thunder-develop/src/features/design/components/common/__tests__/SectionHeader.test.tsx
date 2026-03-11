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
import {render, page} from '@thunder/test-utils/browser';
import {Palette} from '@wso2/oxygen-ui-icons-react';
import SectionHeader from '../SectionHeader';

describe('SectionHeader', () => {
  describe('Rendering', () => {
    it('renders the title', async () => {
      render(<SectionHeader title="Themes" count={5} icon={<Palette />} />);
      await expect.element(page.getByText('Themes')).toBeInTheDocument();
    });

    it('renders the count', async () => {
      render(<SectionHeader title="Themes" count={5} icon={<Palette />} />);
      await expect.element(page.getByText('5')).toBeInTheDocument();
    });

    it('renders zero count', async () => {
      render(<SectionHeader title="Themes" count={0} icon={<Palette />} />);
      await expect.element(page.getByText('0')).toBeInTheDocument();
    });

    it('renders icon', async () => {
      const {container} = render(<SectionHeader title="Themes" count={3} icon={<Palette />} />);
      expect(container.querySelector('svg')).toBeInTheDocument();
    });
  });

  describe('Optional action prop', () => {
    it('renders action element when provided', async () => {
      const action = <button type="button">Add</button>;
      render(<SectionHeader title="Themes" count={3} icon={<Palette />} action={action} />);
      await expect.element(page.getByText('Add')).toBeInTheDocument();
    });

    it('does not render action area when not provided', async () => {
      render(<SectionHeader title="Themes" count={3} icon={<Palette />} />);
      await expect.element(page.getByRole('button')).not.toBeInTheDocument();
    });
  });

  describe('comingSoon prop', () => {
    it('renders "Coming Soon" badge when comingSoon is true', async () => {
      render(<SectionHeader title="Themes" count={3} icon={<Palette />} comingSoon />);
      await expect.element(page.getByText(/coming soon/i)).toBeInTheDocument();
    });

    it('does not render "Coming Soon" badge by default', async () => {
      render(<SectionHeader title="Themes" count={3} icon={<Palette />} />);
      await expect.element(page.getByText(/coming soon/i)).not.toBeInTheDocument();
    });

    it('does not render "Coming Soon" badge when comingSoon is false', async () => {
      render(<SectionHeader title="Themes" count={3} icon={<Palette />} comingSoon={false} />);
      await expect.element(page.getByText(/coming soon/i)).not.toBeInTheDocument();
    });
  });
});
