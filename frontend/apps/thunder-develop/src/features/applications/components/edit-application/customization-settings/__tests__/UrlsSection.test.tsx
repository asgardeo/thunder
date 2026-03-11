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
import {render} from '@thunder/test-utils/browser';
import {page, userEvent} from 'vitest/browser';
import UrlsSection from '../UrlsSection';
import type {Application} from '../../../../models/application';

describe('UrlsSection', () => {
  const mockApplication: Application = {
    id: 'test-app-id',
    name: 'Test Application',
    description: 'Test Description',
    template: 'custom',
    tos_uri: 'https://example.com/terms',
    policy_uri: 'https://example.com/privacy',
  } as Application;

  const mockOnFieldChange = vi.fn();

  beforeEach(() => {
    mockOnFieldChange.mockClear();
  });

  describe('Rendering', () => {
    it('should render the URLs section', async () => {
      await render(<UrlsSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      await expect.element(page.getByText('URLs')).toBeInTheDocument();
      await expect.element(page.getByText('Configure legal and policy URLs for your application.')).toBeInTheDocument();
    });

    it('should render Terms of Service URL field', async () => {
      await render(<UrlsSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      await expect.element(page.getByText('Terms of Service URI')).toBeInTheDocument();
      await expect.element(page.getByPlaceholder('https://example.com/terms')).toBeInTheDocument();
    });

    it('should render Privacy Policy URL field', async () => {
      await render(<UrlsSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      await expect.element(page.getByText('Privacy Policy URI')).toBeInTheDocument();
      await expect.element(page.getByPlaceholder('https://example.com/privacy')).toBeInTheDocument();
    });

    it('should display helper text for both fields', async () => {
      await render(<UrlsSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      await expect.element(page.getByText("URL to your application's Terms of Service. May be displayed to users during consent or user sign-in, sign-up or recovery flows.")).toBeInTheDocument();
      await expect.element(page.getByText("URL to your application's Privacy Policy. May be displayed to users during consent or user sign-in, sign-up or recovery flows.")).toBeInTheDocument();
    });
  });

  describe('Initial Values', () => {
    it('should display URLs from application', async () => {
      await render(<UrlsSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const tosField = page.getByPlaceholder('https://example.com/terms');
      const policyField = page.getByPlaceholder('https://example.com/privacy');

      expect(tosField).toHaveValue('https://example.com/terms');
      expect(policyField).toHaveValue('https://example.com/privacy');
    });

    it('should prioritize editedApp URLs over application', async () => {
      const editedApp = {
        tos_uri: 'https://edited.com/terms',
        policy_uri: 'https://edited.com/privacy',
      };

      await render(<UrlsSection application={mockApplication} editedApp={editedApp} onFieldChange={mockOnFieldChange} />);

      const tosField = page.getByPlaceholder('https://example.com/terms');
      const policyField = page.getByPlaceholder('https://example.com/privacy');

      expect(tosField).toHaveValue('https://edited.com/terms');
      expect(policyField).toHaveValue('https://edited.com/privacy');
    });

    it('should display empty strings when URLs are not provided', async () => {
      const appWithoutUrls = {...mockApplication};
      delete (appWithoutUrls as Partial<Application>).tos_uri;
      delete (appWithoutUrls as Partial<Application>).policy_uri;

      await render(<UrlsSection application={appWithoutUrls} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const tosField = page.getByPlaceholder('https://example.com/terms');
      const policyField = page.getByPlaceholder('https://example.com/privacy');

      expect(tosField).toHaveValue('');
      expect(policyField).toHaveValue('');
    });
  });

  describe('URL Validation', () => {
    it('should show error for invalid ToS URL', async () => {
            const appWithoutUrls = {...mockApplication, tos_uri: '', policy_uri: ''};

      await render(<UrlsSection application={appWithoutUrls} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const tosField = page.getByPlaceholder('https://example.com/terms');
      await userEvent.type(tosField, 'invalid-url');
      await userEvent.tab();

      await vi.waitFor(async () => {
        await expect.element(page.getByText('Please enter a valid URL')).toBeInTheDocument();
      });
    });

    it('should show error for invalid Policy URL', async () => {
            const appWithoutUrls = {...mockApplication, tos_uri: '', policy_uri: ''};

      await render(<UrlsSection application={appWithoutUrls} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const policyField = page.getByPlaceholder('https://example.com/privacy');
      await userEvent.type(policyField, 'not-a-url');
      await userEvent.tab();

      await vi.waitFor(async () => {
        await expect.element(page.getByText('Please enter a valid URL')).toBeInTheDocument();
      });
    });

    it('should not show error for valid ToS URL', async () => {
            const appWithoutUrls = {...mockApplication, tos_uri: '', policy_uri: ''};

      await render(<UrlsSection application={appWithoutUrls} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const tosField = page.getByPlaceholder('https://example.com/terms');
      await userEvent.type(tosField, 'https://example.com/terms');
      await userEvent.tab();

      await vi.waitFor(async () => {
        await expect.element(page.getByText('Please enter a valid URL')).not.toBeInTheDocument();
        await expect.element(page.getByText("URL to your application's Terms of Service. May be displayed to users during consent or user sign-in, sign-up or recovery flows.")).toBeInTheDocument();
      });
    });

    it('should accept empty string as valid', async () => {
      
      await render(<UrlsSection application={mockApplication} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const tosField = page.getByPlaceholder('https://example.com/terms');
      await userEvent.clear(tosField);
      await userEvent.tab();

      await vi.waitFor(async () => {
        await expect.element(page.getByText('Please enter a valid URL')).not.toBeInTheDocument();
      });
    });
  });

  describe('User Input', () => {
    it('should accept valid ToS URL input', async () => {
            const appWithoutUrls = {...mockApplication, tos_uri: '', policy_uri: ''};

      await render(<UrlsSection application={appWithoutUrls} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const tosField = page.getByPlaceholder('https://example.com/terms');
      await userEvent.type(tosField, 'https://new-url.com/terms');

      // Verify the field accepts input
      expect(tosField).toHaveValue('https://new-url.com/terms');
    });

    it('should accept valid Policy URL input', async () => {
            const appWithoutUrls = {...mockApplication, tos_uri: '', policy_uri: ''};

      await render(<UrlsSection application={appWithoutUrls} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const policyField = page.getByPlaceholder('https://example.com/privacy');
      await userEvent.type(policyField, 'https://new-url.com/privacy');

      // Verify the field accepts input
      expect(policyField).toHaveValue('https://new-url.com/privacy');
    });
  });

  describe('Edge Cases', () => {
    it('should handle missing URLs in application', async () => {
      const appWithoutUrls = {...mockApplication};
      delete (appWithoutUrls as Partial<Application>).tos_uri;
      delete (appWithoutUrls as Partial<Application>).policy_uri;

      await render(<UrlsSection application={appWithoutUrls} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      await expect.element(page.getByPlaceholder('https://example.com/terms')).toHaveValue('');
      await expect.element(page.getByPlaceholder('https://example.com/privacy')).toHaveValue('');
    });

    it('should validate URLs with different protocols', async () => {
            const appWithoutUrls = {...mockApplication, tos_uri: '', policy_uri: ''};

      await render(<UrlsSection application={appWithoutUrls} editedApp={{}} onFieldChange={mockOnFieldChange} />);

      const tosField = page.getByPlaceholder('https://example.com/terms');
      await userEvent.type(tosField, 'http://example.com/terms');
      await userEvent.tab();

      await vi.waitFor(async () => {
        await expect.element(page.getByText('Please enter a valid URL')).not.toBeInTheDocument();
      });
    });
  });
});
