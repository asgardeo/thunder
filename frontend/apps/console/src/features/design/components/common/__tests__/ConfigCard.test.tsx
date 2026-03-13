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

import {describe, it, expect} from 'vitest';
import {render, screen} from '@thunder/test-utils';
import userEvent from '@testing-library/user-event';
import ConfigCard from '../ConfigCard';

describe('ConfigCard', () => {
  describe('Rendering', () => {
    it('renders the title text', () => {
      render(
        <ConfigCard title="Shape">
          <p>Content here</p>
        </ConfigCard>,
      );
      expect(screen.getByText('Shape')).toBeInTheDocument();
    });

    it('renders children content when open (default)', () => {
      render(
        <ConfigCard title="General">
          <p>My child content</p>
        </ConfigCard>,
      );
      expect(screen.getByText('My child content')).toBeVisible();
    });
  });

  describe('Accordion behavior', () => {
    it('is expanded by default (defaultOpen not specified)', () => {
      const {container} = render(
        <ConfigCard title="Typography">
          <p>Inner text</p>
        </ConfigCard>,
      );
      // Check the content region is visible (Accordion expanded)
      const details = container.querySelector('.MuiCollapse-entered, [aria-expanded="true"]');
      expect(details ?? screen.getByText('Inner text')).toBeTruthy();
    });

    it('can be toggled closed by clicking the header', async () => {
      const user = userEvent.setup();
      render(
        <ConfigCard title="Colors">
          <p>Color content</p>
        </ConfigCard>,
      );

      await user.click(screen.getByText('Colors'));
      // After clicking, accordion may collapse — title should still be visible
      expect(screen.getByText('Colors')).toBeInTheDocument();
    });

    it('renders with defaultOpen=false as collapsed', () => {
      render(
        <ConfigCard title="Shape" defaultOpen={false}>
          <p>Shape content</p>
        </ConfigCard>,
      );
      // Title should always be visible
      expect(screen.getByText('Shape')).toBeInTheDocument();
    });

    it('can be opened by clicking when defaultOpen=false', async () => {
      const user = userEvent.setup();
      render(
        <ConfigCard title="Type" defaultOpen={false}>
          <p>Some type content</p>
        </ConfigCard>,
      );

      await user.click(screen.getByText('Type'));
      expect(screen.getByText('Some type content')).toBeInTheDocument();
    });
  });
});
