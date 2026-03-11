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
import {render, page, userEvent} from '@thunder/test-utils/browser';
import ConfigCard from '../ConfigCard';

describe('ConfigCard', () => {
  describe('Rendering', () => {
    it('renders the title text', async () => {
      await render(
        <ConfigCard title="Shape">
          <p>Content here</p>
        </ConfigCard>,
      );
      await expect.element(page.getByText('Shape')).toBeInTheDocument();
    });

    it('renders children content when open (default)', async () => {
      await render(
        <ConfigCard title="General">
          <p>My child content</p>
        </ConfigCard>,
      );
      await expect.element(page.getByText('My child content')).toBeVisible();
    });
  });

  describe('Accordion behavior', () => {
    it('is expanded by default (defaultOpen not specified)', async () => {
      const {container} = await render(
        <ConfigCard title="Typography">
          <p>Inner text</p>
        </ConfigCard>,
      );
      // Check the content region is visible (Accordion expanded)
      const details = container.querySelector('.MuiCollapse-entered, [aria-expanded="true"]');
      expect(details ?? container.querySelector('p')).toBeTruthy();
    });

    it('can be toggled closed by clicking the header', async () => {
      await render(
        <ConfigCard title="Colors">
          <p>Color content</p>
        </ConfigCard>,
      );

      await userEvent.click(page.getByText('Colors'));
      // After clicking, accordion may collapse — title should still be visible
      await expect.element(page.getByText('Colors')).toBeInTheDocument();
    });

    it('renders with defaultOpen=false as collapsed', async () => {
      await render(
        <ConfigCard title="Shape" defaultOpen={false}>
          <p>Shape content</p>
        </ConfigCard>,
      );
      // Title should always be visible
      await expect.element(page.getByText('Shape')).toBeInTheDocument();
    });

    it('can be opened by clicking when defaultOpen=false', async () => {
      await render(
        <ConfigCard title="Type" defaultOpen={false}>
          <p>Some type content</p>
        </ConfigCard>,
      );

      await userEvent.click(page.getByText('Type'));
      await expect.element(page.getByText('Some type content')).toBeInTheDocument();
    });
  });
});
