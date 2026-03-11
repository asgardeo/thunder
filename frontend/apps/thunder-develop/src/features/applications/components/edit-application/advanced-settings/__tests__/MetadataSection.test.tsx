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
import {render} from '@thunder/test-utils/browser';
import {page} from 'vitest/browser';
import MetadataSection from '../MetadataSection';
import type {Application} from '../../../../models/application';

describe('MetadataSection', () => {
  describe('Rendering', () => {
    it('should render metadata section with created and updated timestamps', async () => {
      const application: Application = {
        id: 'test-app-id',
        name: 'Test Application',
        template: 'custom',
        created_at: '2025-01-01T00:00:00Z',
        updated_at: '2025-01-15T12:30:00Z',
      } as Application;

      await render(<MetadataSection application={application} />);

      await expect.element(page.getByText('Metadata')).toBeInTheDocument();
      await expect.element(page.getByText('Created At')).toBeInTheDocument();
      await expect.element(page.getByText('Updated At')).toBeInTheDocument();
    });

    it('should render only created_at when updated_at is missing', async () => {
      const application: Application = {
        id: 'test-app-id',
        name: 'Test Application',
        template: 'custom',
        created_at: '2025-01-01T00:00:00Z',
      } as Application;

      await render(<MetadataSection application={application} />);

      await expect.element(page.getByText('Metadata')).toBeInTheDocument();
      await expect.element(page.getByText('Created At')).toBeInTheDocument();
      await expect.element(page.getByText('Updated At')).not.toBeInTheDocument();
    });

    it('should render only updated_at when created_at is missing', async () => {
      const application: Application = {
        id: 'test-app-id',
        name: 'Test Application',
        template: 'custom',
        updated_at: '2025-01-15T12:30:00Z',
      } as Application;

      await render(<MetadataSection application={application} />);

      await expect.element(page.getByText('Metadata')).toBeInTheDocument();
      await expect.element(page.getByText('Created At')).not.toBeInTheDocument();
      await expect.element(page.getByText('Updated At')).toBeInTheDocument();
    });

    it('should return null when both timestamps are missing', async () => {
      const application: Application = {
        id: 'test-app-id',
        name: 'Test Application',
        template: 'custom',
      } as Application;

      const {container} = await render(<MetadataSection application={application} />);

      expect(container.firstChild).toBeNull();
    });
  });

  describe('Timestamp Formatting', () => {
    it('should format created_at timestamp as locale string', async () => {
      const createdDate = '2025-01-01T10:30:45Z';
      const application: Application = {
        id: 'test-app-id',
        name: 'Test Application',
        template: 'custom',
        created_at: createdDate,
      } as Application;

      await render(<MetadataSection application={application} />);

      const expectedFormattedDate = new Date(createdDate).toLocaleString();
      await expect.element(page.getByText(expectedFormattedDate)).toBeInTheDocument();
    });

    it('should format updated_at timestamp as locale string', async () => {
      const updatedDate = '2025-01-15T14:20:30Z';
      const application: Application = {
        id: 'test-app-id',
        name: 'Test Application',
        template: 'custom',
        updated_at: updatedDate,
      } as Application;

      await render(<MetadataSection application={application} />);

      const expectedFormattedDate = new Date(updatedDate).toLocaleString();
      await expect.element(page.getByText(expectedFormattedDate)).toBeInTheDocument();
    });

    it('should format both timestamps correctly', async () => {
      const createdDate = '2025-01-01T08:00:00Z';
      const updatedDate = '2025-01-15T16:45:00Z';
      const application: Application = {
        id: 'test-app-id',
        name: 'Test Application',
        template: 'custom',
        created_at: createdDate,
        updated_at: updatedDate,
      } as Application;

      await render(<MetadataSection application={application} />);

      const expectedCreatedDate = new Date(createdDate).toLocaleString();
      const expectedUpdatedDate = new Date(updatedDate).toLocaleString();

      await expect.element(page.getByText(expectedCreatedDate)).toBeInTheDocument();
      await expect.element(page.getByText(expectedUpdatedDate)).toBeInTheDocument();
    });
  });

  describe('Layout and Styling', () => {
    it('should render timestamps with correct typography variants', async () => {
      const application: Application = {
        id: 'test-app-id',
        name: 'Test Application',
        template: 'custom',
        created_at: '2025-01-01T00:00:00Z',
        updated_at: '2025-01-15T00:00:00Z',
      } as Application;

      await render(<MetadataSection application={application} />);

      const createdLabel = page.getByText('Created At');
      const updatedLabel = page.getByText('Updated At');

      expect(createdLabel).toHaveClass('MuiTypography-subtitle2');
      expect(updatedLabel).toHaveClass('MuiTypography-subtitle2');
    });

    it('should render in a Stack with proper spacing', async () => {
      const application: Application = {
        id: 'test-app-id',
        name: 'Test Application',
        template: 'custom',
        created_at: '2025-01-01T00:00:00Z',
        updated_at: '2025-01-15T00:00:00Z',
      } as Application;

      const {container} = await render(<MetadataSection application={application} />);

      const stack = container.querySelector('.MuiStack-root');
      expect(stack).toBeInTheDocument();
    });
  });

  describe('Edge Cases', () => {
    it('should handle invalid date strings gracefully', async () => {
      const application: Application = {
        id: 'test-app-id',
        name: 'Test Application',
        template: 'custom',
        created_at: 'invalid-date',
      } as Application;

      await render(<MetadataSection application={application} />);

      await expect.element(page.getByText('Created At')).toBeInTheDocument();
      await expect.element(page.getByText('Invalid Date')).toBeInTheDocument();
    });

    it('should handle undefined timestamp values', async () => {
      const application: Application = {
        id: 'test-app-id',
        name: 'Test Application',
        template: 'custom',
        created_at: undefined,
        updated_at: undefined,
      } as Application;

      const {container} = await render(<MetadataSection application={application} />);

      expect(container.firstChild).toBeNull();
    });

    it('should handle empty string timestamps', async () => {
      const application: Application = {
        id: 'test-app-id',
        name: 'Test Application',
        template: 'custom',
        created_at: '',
        updated_at: '',
      } as Application;

      const {container} = await render(<MetadataSection application={application} />);

      expect(container.firstChild).toBeNull();
    });
  });
});
