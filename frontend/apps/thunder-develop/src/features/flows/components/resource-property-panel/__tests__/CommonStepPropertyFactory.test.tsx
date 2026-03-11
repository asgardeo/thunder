/**
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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
import CommonStepPropertyFactory from '../CommonStepPropertyFactory';
import type {Resource} from '../../../models/resources';
import {ElementTypes} from '../../../models/elements';

// Mock RichTextWithTranslation component
vi.mock('../rich-text/RichTextWithTranslation', () => ({
  default: ({onChange}: {onChange: (html: string) => void}) => (
    <div data-testid="rich-text-with-translation">
      <button type="button" onClick={() => onChange('<p>Updated content</p>')}>
        Rich Text Editor
      </button>
    </div>
  ),
}));

describe('CommonStepPropertyFactory', () => {
  const mockOnChange = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('RichText Element with text property', () => {
    it('should render RichTextWithTranslation for text property when resource is RichText', async () => {
      const richTextResource: Resource = {
        id: 'resource-1',
        type: ElementTypes.RichText,
        config: {},
      } as Resource;

      await render(
        <CommonStepPropertyFactory
          resource={richTextResource}
          propertyKey="text"
          propertyValue="<p>Test content</p>"
          onChange={mockOnChange}
        />,
      );

      await expect.element(page.getByTestId('rich-text-with-translation')).toBeInTheDocument();
    });

    it('should call onChange when RichText content changes', async () => {
      const richTextResource: Resource = {
        id: 'resource-1',
        type: ElementTypes.RichText,
        config: {},
      } as Resource;

      await render(
        <CommonStepPropertyFactory
          resource={richTextResource}
          propertyKey="text"
          propertyValue="<p>Test</p>"
          onChange={mockOnChange}
        />,
      );

      const button = page.getByText('Rich Text Editor');
      await userEvent.click(button);

      expect(mockOnChange).toHaveBeenCalledWith('text', '<p>Updated content</p>', richTextResource);
    });

    it('should not render RichTextWithTranslation for non-text properties on RichText', async () => {
      const richTextResource: Resource = {
        id: 'resource-1',
        type: ElementTypes.RichText,
        config: {},
      } as Resource;

      await render(
        <CommonStepPropertyFactory
          resource={richTextResource}
          propertyKey="other"
          propertyValue="test"
          onChange={mockOnChange}
        />,
      );

      await expect.element(page.getByTestId('rich-text-with-translation')).not.toBeInTheDocument();
    });
  });

  describe('Boolean Properties', () => {
    it('should render FormControlLabel with Checkbox for boolean true value', async () => {
      const resource: Resource = {
        id: 'resource-1',
        type: 'VIEW',
        config: {},
      } as Resource;

      await render(
        <CommonStepPropertyFactory
          resource={resource}
          propertyKey="isEnabled"
          propertyValue
          onChange={mockOnChange}
        />,
      );

      const checkbox = page.getByRole('checkbox');
      expect(checkbox).toBeInTheDocument();
      expect(checkbox).toBeChecked();
    });

    it('should render FormControlLabel with Checkbox for boolean false value', async () => {
      const resource: Resource = {
        id: 'resource-1',
        type: 'VIEW',
        config: {},
      } as Resource;

      await render(
        <CommonStepPropertyFactory
          resource={resource}
          propertyKey="isDisabled"
          propertyValue={false}
          onChange={mockOnChange}
        />,
      );

      const checkbox = page.getByRole('checkbox');
      expect(checkbox).toBeInTheDocument();
      expect(checkbox).not.toBeChecked();
    });

    it('should convert camelCase propertyKey to Start Case label for checkbox', async () => {
      const resource: Resource = {
        id: 'resource-1',
        type: 'VIEW',
        config: {},
      } as Resource;

      await render(
        <CommonStepPropertyFactory
          resource={resource}
          propertyKey="showHeader"
          propertyValue
          onChange={mockOnChange}
        />,
      );

      await expect.element(page.getByText('Show Header')).toBeInTheDocument();
    });

    it('should call onChange with correct parameters when checkbox is toggled', async () => {
      const resource: Resource = {
        id: 'resource-1',
        type: 'VIEW',
        config: {},
      } as Resource;

      await render(
        <CommonStepPropertyFactory
          resource={resource}
          propertyKey="active"
          propertyValue={false}
          onChange={mockOnChange}
        />,
      );

      const checkbox = page.getByRole('checkbox');
      await userEvent.click(checkbox);

      expect(mockOnChange).toHaveBeenCalledWith('active', true, resource);
    });
  });

  describe('String Properties', () => {
    it('should render FormControl with TextField for string value', async () => {
      const resource: Resource = {
        id: 'resource-1',
        type: 'VIEW',
        config: {},
      } as Resource;

      await render(
        <CommonStepPropertyFactory
          resource={resource}
          propertyKey="title"
          propertyValue="My Title"
          onChange={mockOnChange}
        />,
      );

      const textField = page.getByRole('textbox');
      expect(textField).toBeInTheDocument();
      expect(textField).toHaveValue('My Title');
    });

    it('should render TextField for empty string value', async () => {
      const resource: Resource = {
        id: 'resource-1',
        type: 'VIEW',
        config: {},
      } as Resource;

      await render(
        <CommonStepPropertyFactory
          resource={resource}
          propertyKey="description"
          propertyValue=""
          onChange={mockOnChange}
        />,
      );

      const textField = page.getByRole('textbox');
      expect(textField).toBeInTheDocument();
      expect(textField).toHaveValue('');
    });

    it('should convert camelCase propertyKey to Start Case label for text field', async () => {
      const resource: Resource = {
        id: 'resource-1',
        type: 'VIEW',
        config: {},
      } as Resource;

      await render(
        <CommonStepPropertyFactory
          resource={resource}
          propertyKey="pageTitle"
          propertyValue="Test"
          onChange={mockOnChange}
        />,
      );

      await expect.element(page.getByText('Page Title')).toBeInTheDocument();
    });

    it('should show placeholder with Start Case property key', async () => {
      const resource: Resource = {
        id: 'resource-1',
        type: 'VIEW',
        config: {},
      } as Resource;

      await render(
        <CommonStepPropertyFactory
          resource={resource}
          propertyKey="userName"
          propertyValue=""
          onChange={mockOnChange}
        />,
      );

      const textField = page.getByPlaceholder('Enter User Name');
      expect(textField).toBeInTheDocument();
    });

    it('should call onChange when text field value changes', async () => {
      const resource: Resource = {
        id: 'resource-1',
        type: 'VIEW',
        config: {},
      } as Resource;

      await render(
        <CommonStepPropertyFactory
          resource={resource}
          propertyKey="name"
          propertyValue="Initial"
          onChange={mockOnChange}
        />,
      );

      const textField = page.getByRole('textbox');
      await userEvent.fill(textField, 'Updated Value');

      expect(mockOnChange).toHaveBeenCalledWith('name', 'Updated Value', resource);
    });
  });

  describe('Null Cases', () => {
    it('should return null for object property values', async () => {
      const resource: Resource = {
        id: 'resource-1',
        type: 'VIEW',
        config: {},
      } as Resource;

      const {container} = await render(
        <CommonStepPropertyFactory
          resource={resource}
          propertyKey="config"
          propertyValue={{nested: 'value'}}
          onChange={mockOnChange}
        />,
      );

      expect(container.firstChild).toBeNull();
    });

    it('should return null for number property values', async () => {
      const resource: Resource = {
        id: 'resource-1',
        type: 'VIEW',
        config: {},
      } as Resource;

      const {container} = await render(
        <CommonStepPropertyFactory
          resource={resource}
          propertyKey="order"
          propertyValue={123}
          onChange={mockOnChange}
        />,
      );

      expect(container.firstChild).toBeNull();
    });

    it('should return null for array property values', async () => {
      const resource: Resource = {
        id: 'resource-1',
        type: 'VIEW',
        config: {},
      } as Resource;

      const {container} = await render(
        <CommonStepPropertyFactory
          resource={resource}
          propertyKey="items"
          propertyValue={['a', 'b', 'c']}
          onChange={mockOnChange}
        />,
      );

      expect(container.firstChild).toBeNull();
    });

    it('should return null for undefined property values', async () => {
      const resource: Resource = {
        id: 'resource-1',
        type: 'VIEW',
        config: {},
      } as Resource;

      const {container} = await render(
        <CommonStepPropertyFactory
          resource={resource}
          propertyKey="undefinedProp"
          propertyValue={undefined}
          onChange={mockOnChange}
        />,
      );

      expect(container.firstChild).toBeNull();
    });

    it('should return null for null property values', async () => {
      const resource: Resource = {
        id: 'resource-1',
        type: 'VIEW',
        config: {},
      } as Resource;

      const {container} = await render(
        <CommonStepPropertyFactory
          resource={resource}
          propertyKey="nullProp"
          propertyValue={null}
          onChange={mockOnChange}
        />,
      );

      expect(container.firstChild).toBeNull();
    });
  });

  describe('Additional Props', () => {
    it('should pass additional props to FormControlLabel for boolean values', async () => {
      const resource: Resource = {
        id: 'resource-1',
        type: 'VIEW',
        config: {},
      } as Resource;

      await render(
        <CommonStepPropertyFactory
          resource={resource}
          propertyKey="enabled"
          propertyValue
          onChange={mockOnChange}
          data-testid="custom-checkbox"
        />,
      );

      await expect.element(page.getByRole('checkbox')).toBeInTheDocument();
    });

    it('should pass additional props to TextField for string values', async () => {
      const resource: Resource = {
        id: 'resource-1',
        type: 'VIEW',
        config: {},
      } as Resource;

      await render(
        <CommonStepPropertyFactory
          resource={resource}
          propertyKey="name"
          propertyValue="Test"
          onChange={mockOnChange}
          data-testid="custom-textfield"
        />,
      );

      await expect.element(page.getByRole('textbox')).toBeInTheDocument();
    });
  });
});
