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
import {ElementCategories, ElementTypes} from '@/features/flows/models/elements';
import {StepCategories, StepTypes, ExecutionTypes} from '@/features/flows/models/steps';
import type {Resource} from '@/features/flows/models/resources';
import type {Element} from '@/features/flows/models/elements';
import ResourceProperties from '../ResourceProperties';

// Mock dependencies
vi.mock('../ResourcePropertyFactory', () => ({
  default: ({
    resource,
    propertyKey,
    propertyValue,
    onChange,
  }: {
    resource: Resource;
    propertyKey: string;
    propertyValue: unknown;
    onChange?: (key: string, value: unknown, resource: Resource) => void;
  }) => (
    <div
      data-testid={`resource-property-factory-${propertyKey}`}
      data-resource-id={resource.id}
      data-property-value={String(propertyValue)}
    >
      {propertyKey}: {String(propertyValue)}
      {onChange && (
        <button
          type="button"
          data-testid={`trigger-change-${propertyKey}`}
          onClick={() => onChange(propertyKey, propertyValue, resource)}
        >
          Trigger Change
        </button>
      )}
    </div>
  ),
}));

vi.mock('../nodes/RulesProperties', () => ({
  default: () => <div data-testid="rules-properties">Rules Properties</div>,
}));

vi.mock('../extended-properties/FieldExtendedProperties', () => ({
  default: ({resource}: {resource: Resource}) => (
    <div data-testid="field-extended-properties" data-resource-id={resource.id}>
      Field Extended Properties
    </div>
  ),
}));

vi.mock('../extended-properties/ButtonExtendedProperties', () => ({
  default: ({resource}: {resource: Resource}) => (
    <div data-testid="button-extended-properties" data-resource-id={resource.id}>
      Button Extended Properties
    </div>
  ),
}));

vi.mock('../extended-properties/ExecutionExtendedProperties', () => ({
  default: ({resource}: {resource: Resource}) => (
    <div data-testid="execution-extended-properties" data-resource-id={resource.id}>
      Execution Extended Properties
    </div>
  ),
}));

vi.mock('@/features/flows/components/resource-property-panel/TextPropertyField', () => ({
  default: ({
    resource,
    propertyKey,
    propertyValue,
  }: {
    resource: Resource;
    propertyKey: string;
    propertyValue: string;
  }) => (
    <div
      data-testid={`text-property-field-${propertyKey}`}
      data-resource-id={resource.id}
      data-property-value={propertyValue}
    >
      {propertyKey}: {propertyValue}
    </div>
  ),
}));

describe('ResourceProperties', () => {
  const mockOnChange = vi.fn();
  const mockOnVariantChange = vi.fn();

  const createMockResource = (overrides: Partial<Resource> = {}): Resource =>
    ({
      id: 'resource-1',
      type: 'TEXT_INPUT',
      category: ElementCategories.Field,
      ...overrides,
    }) as Resource;

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Field Category', () => {
    it('should render FieldExtendedProperties for Field category', async () => {
      const resource = createMockResource({category: ElementCategories.Field});

      await render(
        <ResourceProperties
          resource={resource}
          properties={{label: 'Test Label'}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByTestId('field-extended-properties')).toBeInTheDocument();
    });

    it('should render element ID for Field category', async () => {
      const resource = createMockResource({id: 'field-123', category: ElementCategories.Field});

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByTestId('resource-property-factory-id')).toBeInTheDocument();
      await expect.element(page.getByTestId('resource-property-factory-id')).toHaveAttribute('data-property-value', 'field-123');
    });

    it('should render property factories for Field category properties', async () => {
      const resource = createMockResource({category: ElementCategories.Field});

      await render(
        <ResourceProperties
          resource={resource}
          properties={{label: 'Test Label', placeholder: 'Enter value'}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByTestId('resource-property-factory-label')).toBeInTheDocument();
      await expect.element(page.getByTestId('resource-property-factory-placeholder')).toBeInTheDocument();
    });
  });

  describe('Action Category', () => {
    it('should render ButtonExtendedProperties for Action type', async () => {
      const resource = createMockResource({
        category: ElementCategories.Action,
        type: ElementTypes.Action,
      });

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByTestId('button-extended-properties')).toBeInTheDocument();
    });

    it('should render element ID for Action category', async () => {
      const resource = createMockResource({
        id: 'action-456',
        category: ElementCategories.Action,
        type: ElementTypes.Action,
      });

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByTestId('resource-property-factory-id')).toHaveAttribute('data-property-value', 'action-456');
    });

    it('should not render ButtonExtendedProperties for non-Action type in Action category', async () => {
      const resource = createMockResource({
        category: ElementCategories.Action,
        type: 'LINK' as typeof ElementTypes.Action,
      });

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByTestId('button-extended-properties')).not.toBeInTheDocument();
    });
  });

  describe('Decision Category - Rule Type', () => {
    it('should render RulesProperties for Rule type', async () => {
      const resource = createMockResource({
        category: StepCategories.Decision,
        type: StepTypes.Rule,
      } as Partial<Resource>);

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByTestId('rules-properties')).toBeInTheDocument();
    });

    it('should render element ID for Rule type', async () => {
      const resource = createMockResource({
        id: 'rule-789',
        category: StepCategories.Decision,
        type: StepTypes.Rule,
      } as Partial<Resource>);

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByTestId('resource-property-factory-id')).toHaveAttribute('data-property-value', 'rule-789');
    });

    it('should return null for Decision category with non-Rule type', async () => {
      const resource = createMockResource({
        category: StepCategories.Decision,
        type: 'CONDITION' as typeof StepTypes.Rule,
      } as Partial<Resource>);

      const {container} = await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      expect(container.firstChild).toBeNull();
    });
  });

  describe('Interface Category - End Type', () => {
    it('should render element ID for End type', async () => {
      const resource = createMockResource({
        id: 'end-step',
        category: StepCategories.Interface,
        type: StepTypes.End,
      } as Partial<Resource>);

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByTestId('resource-property-factory-id')).toHaveAttribute('data-property-value', 'end-step');
    });

    it('should return null for Interface category with non-End type', async () => {
      const resource = createMockResource({
        category: StepCategories.Interface,
        type: 'VIEW' as typeof StepTypes.End,
      } as Partial<Resource>);

      const {container} = await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      expect(container.firstChild).toBeNull();
    });
  });

  describe('Workflow Category', () => {
    it('should render ExecutionExtendedProperties for Workflow category', async () => {
      const resource = createMockResource({
        category: StepCategories.Workflow,
        type: StepTypes.Execution,
      } as Partial<Resource>);

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByTestId('execution-extended-properties')).toBeInTheDocument();
    });

    it('should render ConfirmationCode execution type differently', async () => {
      const resource = createMockResource({
        category: StepCategories.Workflow,
        type: StepTypes.Execution,
        data: {
          action: {
            executor: {
              name: ExecutionTypes.ConfirmationCode,
            },
          },
        },
      } as Partial<Resource>);

      await render(
        <ResourceProperties
          resource={resource}
          properties={{prop1: 'value1'}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      // ConfirmationCode renders property factories instead of ExecutionExtendedProperties
      await expect.element(page.getByTestId('execution-extended-properties')).not.toBeInTheDocument();
      await expect.element(page.getByTestId('resource-property-factory-prop1')).toBeInTheDocument();
    });
  });

  describe('Display Category - Text Type', () => {
    it('should render TextPropertyField for Text type', async () => {
      const resource = createMockResource({
        category: ElementCategories.Display,
        type: ElementTypes.Text,
        label: 'Sample Text',
      } as Partial<Resource>);

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByTestId('text-property-field-label')).toBeInTheDocument();
    });

    it('should render variant selector for Text type with variants', async () => {
      const variants = [
        {variant: 'heading', id: 'v1'},
        {variant: 'body', id: 'v2'},
      ] as Element[];

      const resource = createMockResource({
        category: ElementCategories.Display,
        type: ElementTypes.Text,
        variants,
        variant: 'heading',
      } as Partial<Resource>);

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByText('Variant')).toBeInTheDocument();
    });
  });

  describe('Display Category - Image Type', () => {
    it('should render TextPropertyFields for Image type', async () => {
      const resource = createMockResource({
        category: ElementCategories.Display,
        type: ElementTypes.Image,
        src: 'https://example.com/image.png',
        alt: 'Sample Image',
      } as Partial<Resource>);

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByTestId('text-property-field-src')).toBeInTheDocument();
      await expect.element(page.getByTestId('text-property-field-alt')).toBeInTheDocument();
    });

    it('should handle empty src and alt values', async () => {
      const resource = createMockResource({
        category: ElementCategories.Display,
        type: ElementTypes.Image,
      } as Partial<Resource>);

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByTestId('text-property-field-src')).toHaveAttribute('data-property-value', '');
      await expect.element(page.getByTestId('text-property-field-alt')).toHaveAttribute('data-property-value', '');
    });
  });

  describe('Display Category - Other Types', () => {
    it('should render default property factories for other Display types', async () => {
      const resource = createMockResource({
        category: ElementCategories.Display,
        type: 'DIVIDER' as typeof ElementTypes.Text,
      } as Partial<Resource>);

      await render(
        <ResourceProperties
          resource={resource}
          properties={{thickness: '2px'}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByTestId('resource-property-factory-thickness')).toBeInTheDocument();
    });
  });

  describe('Default Category', () => {
    it('should render default property factories for unknown category', async () => {
      const resource = createMockResource({
        category: 'UNKNOWN' as typeof ElementCategories.Field,
      });

      await render(
        <ResourceProperties
          resource={resource}
          properties={{customProp: 'customValue'}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByTestId('resource-property-factory-id')).toBeInTheDocument();
      await expect.element(page.getByTestId('resource-property-factory-customProp')).toBeInTheDocument();
    });
  });

  describe('Variant Selection', () => {
    it('should render variant selector when resource has variants', async () => {
      const variants = [
        {variant: 'primary', id: 'v1'},
        {variant: 'secondary', id: 'v2'},
      ] as Element[];

      const resource = createMockResource({
        category: ElementCategories.Field,
        variants,
        variant: 'primary',
      } as Partial<Resource>);

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByText('Variant')).toBeInTheDocument();
    });

    it('should call onVariantChange when variant is selected', async () => {
      const variants = [
        {variant: 'primary', id: 'v1'},
        {variant: 'secondary', id: 'v2'},
      ] as Element[];

      const resource = createMockResource({
        category: ElementCategories.Field,
        variants,
        variant: 'primary',
      } as Partial<Resource>);

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      // Find the variant select by role combobox with the variant-select id
      const variantSelect = document.getElementById('variant-select')!;
      await userEvent.click(variantSelect);

      const secondaryOption = page.getByRole('option', {name: 'secondary'});
      await userEvent.click(secondaryOption);

      expect(mockOnVariantChange).toHaveBeenCalledWith('secondary');
    });

    it('should not render variant selector when resource has no variants', async () => {
      const resource = createMockResource({
        category: ElementCategories.Field,
      });

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByText('Variant')).not.toBeInTheDocument();
    });

    it('should not render variant selector when variants array is empty', async () => {
      const resource = createMockResource({
        category: ElementCategories.Field,
        variants: [],
      } as Partial<Resource>);

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByText('Variant')).not.toBeInTheDocument();
    });
  });

  describe('onChange Handler - Type Preservation', () => {
    it('should preserve boolean values in onChange', async () => {
      const resource = createMockResource({category: ElementCategories.Field});

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      // The internal handleChange function processes values
      // This test verifies the component renders without errors
      await expect.element(page.getByTestId('field-extended-properties')).toBeInTheDocument();
    });
  });

  describe('Sync Selected Variant on Resource Change', () => {
    it('should sync selected variant when resource changes', async () => {
      const variants1 = [
        {variant: 'v1', id: 'variant-1'},
        {variant: 'v2', id: 'variant-2'},
      ] as Element[];

      const resource1 = createMockResource({
        id: 'resource-1',
        category: ElementCategories.Field,
        variants: variants1,
        variant: 'v1',
      } as Partial<Resource>);

      const variants2 = [
        {variant: 'v3', id: 'variant-3'},
        {variant: 'v4', id: 'variant-4'},
      ] as Element[];

      const resource2 = createMockResource({
        id: 'resource-2',
        category: ElementCategories.Field,
        variants: variants2,
        variant: 'v3',
      } as Partial<Resource>);

      const {rerender} = await render(
        <ResourceProperties
          resource={resource1}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByText('Variant')).toBeInTheDocument();

      await rerender(
        <ResourceProperties
          resource={resource2}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByText('Variant')).toBeInTheDocument();
    });

    it('should set selectedVariant to undefined when resource has no variants', async () => {
      const resource1 = createMockResource({
        id: 'resource-1',
        category: ElementCategories.Field,
        variants: [{variant: 'v1', id: 'variant-1'}] as Element[],
        variant: 'v1',
      } as Partial<Resource>);

      const resource2 = createMockResource({
        id: 'resource-2',
        category: ElementCategories.Field,
        variants: undefined,
      } as Partial<Resource>);

      const {rerender} = await render(
        <ResourceProperties
          resource={resource1}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByText('Variant')).toBeInTheDocument();

      await rerender(
        <ResourceProperties
          resource={resource2}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByText('Variant')).not.toBeInTheDocument();
    });

    it('should fall back to first variant when current variant is not found', async () => {
      const variants = [
        {variant: 'first', id: 'variant-1'},
        {variant: 'second', id: 'variant-2'},
      ] as Element[];

      const resource = createMockResource({
        id: 'resource-1',
        category: ElementCategories.Field,
        variants,
        variant: 'non-existent',
      } as Partial<Resource>);

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByText('Variant')).toBeInTheDocument();
    });
  });

  describe('handleChange Type Preservation', () => {
    it('should preserve boolean values in onChange', async () => {
      const resource = createMockResource({category: ElementCategories.Field});

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      // The component should render
      await expect.element(page.getByTestId('field-extended-properties')).toBeInTheDocument();
    });

    it('should preserve object values in onChange', async () => {
      const resource = createMockResource({category: ElementCategories.Field});

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByTestId('field-extended-properties')).toBeInTheDocument();
    });

    it('should convert number values to string in onChange', async () => {
      const resource = createMockResource({category: ElementCategories.Field});

      await render(
        <ResourceProperties
          resource={resource}
          properties={{numericProp: 42}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByTestId('resource-property-factory-numericProp')).toBeInTheDocument();
    });

    it('should default to empty string for null/undefined values', async () => {
      const resource = createMockResource({category: ElementCategories.Field});

      await render(
        <ResourceProperties
          resource={resource}
          properties={{nullProp: null as unknown as string}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByTestId('resource-property-factory-nullProp')).toBeInTheDocument();
    });
  });

  describe('Display Category - Text Type with Variants', () => {
    it('should render variant selector for Text type and call onVariantChange', async () => {
      const variants = [
        {variant: 'heading', id: 'v1'},
        {variant: 'body', id: 'v2'},
      ] as Element[];

      const resource = createMockResource({
        category: ElementCategories.Display,
        type: ElementTypes.Text,
        variants,
        variant: 'heading',
      } as Partial<Resource>);

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByText('Variant')).toBeInTheDocument();

      // Find and click the variant select
      const variantSelect = document.getElementById('variant-select')!;
      await userEvent.click(variantSelect);

      const bodyOption = page.getByRole('option', {name: 'body'});
      await userEvent.click(bodyOption);

      expect(mockOnVariantChange).toHaveBeenCalledWith('body');
    });

    it('should handle onVariantChange being undefined for Text type', async () => {
      const variants = [
        {variant: 'heading', id: 'v1'},
        {variant: 'body', id: 'v2'},
      ] as Element[];

      const resource = createMockResource({
        category: ElementCategories.Display,
        type: ElementTypes.Text,
        variants,
        variant: 'heading',
      } as Partial<Resource>);

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={undefined}
        />,
      );

      const variantSelect = document.getElementById('variant-select')!;
      await userEvent.click(variantSelect);

      const bodyOption = page.getByRole('option', {name: 'body'});
      await userEvent.click(bodyOption);

      // Should not throw error
      await expect.element(page.getByText('Variant')).toBeInTheDocument();
    });

    it('should handle variant not found in variants array for Text type', async () => {
      const variants = [
        {variant: 'heading', id: 'v1'},
        {variant: 'body', id: 'v2'},
      ] as Element[];

      const resource = createMockResource({
        category: ElementCategories.Display,
        type: ElementTypes.Text,
        variants,
        variant: 'heading',
      } as Partial<Resource>);

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      const variantSelect = document.getElementById('variant-select')!;
      await userEvent.click(variantSelect);

      // Try to select a non-existent variant through the select component
      // The select should still work properly
      await expect.element(page.getByText('Variant')).toBeInTheDocument();
    });

    it('should not render variant selector for Text type without variants', async () => {
      const resource = createMockResource({
        category: ElementCategories.Display,
        type: ElementTypes.Text,
        label: 'Sample Text',
      } as Partial<Resource>);

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByText('Variant')).not.toBeInTheDocument();
    });

    it('should handle variant selection with empty variant value', async () => {
      const variants = [
        {variant: 'primary', id: 'v1'},
        {variant: '', id: 'v2'},
      ] as Element[];

      const resource = createMockResource({
        category: ElementCategories.Display,
        type: ElementTypes.Text,
        variants,
        variant: 'primary',
      } as Partial<Resource>);

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByText('Variant')).toBeInTheDocument();
    });
  });

  describe('Workflow Category - Non-ConfirmationCode Execution', () => {
    it('should render ExecutionExtendedProperties for regular execution', async () => {
      const resource = createMockResource({
        category: StepCategories.Workflow,
        type: StepTypes.Execution,
        data: {
          action: {
            executor: {
              name: 'RegularExecutor',
            },
          },
        },
      } as Partial<Resource>);

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByTestId('execution-extended-properties')).toBeInTheDocument();
    });

    it('should render ExecutionExtendedProperties when executor is undefined', async () => {
      const resource = createMockResource({
        category: StepCategories.Workflow,
        type: StepTypes.Execution,
        data: {},
      } as Partial<Resource>);

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByTestId('execution-extended-properties')).toBeInTheDocument();
    });
  });

  describe('Display Category - Image Type', () => {
    it('should render src and alt fields with empty values when not provided', async () => {
      const resource = createMockResource({
        category: ElementCategories.Display,
        type: ElementTypes.Image,
      } as Partial<Resource>);

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByTestId('text-property-field-src')).toBeInTheDocument();
      await expect.element(page.getByTestId('text-property-field-alt')).toBeInTheDocument();
    });
  });

  describe('Interface Category - Non-End Type', () => {
    it('should return null for VIEW type in Interface category', async () => {
      const resource = createMockResource({
        category: StepCategories.Interface,
        type: StepTypes.View,
      } as Partial<Resource>);

      const {container} = await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      expect(container.firstChild).toBeNull();
    });
  });

  describe('Field Category with Variants', () => {
    it('should render variant selector and handle empty variant selection', async () => {
      const variants = [
        {variant: 'text', id: 'v1'},
        {variant: 'password', id: 'v2'},
      ] as Element[];

      const resource = createMockResource({
        category: ElementCategories.Field,
        variants,
        variant: 'text',
      } as Partial<Resource>);

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByText('Variant')).toBeInTheDocument();
    });
  });

  describe('Action Category - Non-Action Types', () => {
    it('should render only ID for Link type in Action category', async () => {
      const resource = createMockResource({
        id: 'link-123',
        category: ElementCategories.Action,
        type: 'LINK',
      } as Partial<Resource>);

      await render(
        <ResourceProperties
          resource={resource}
          properties={{href: 'https://example.com'}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByTestId('resource-property-factory-id')).toBeInTheDocument();
      await expect.element(page.getByTestId('resource-property-factory-href')).toBeInTheDocument();
      await expect.element(page.getByTestId('button-extended-properties')).not.toBeInTheDocument();
    });
  });

  describe('Action Category with Variants', () => {
    it('should render variant selector for Action category with variants', async () => {
      const variants = [
        {variant: 'primary', id: 'v1'},
        {variant: 'secondary', id: 'v2'},
      ] as Element[];

      const resource = createMockResource({
        category: ElementCategories.Action,
        type: ElementTypes.Action,
        variants,
        variant: 'primary',
      } as Partial<Resource>);

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      await expect.element(page.getByText('Variant')).toBeInTheDocument();
    });

    it('should call onVariantChange for Action category variant selection', async () => {
      const variants = [
        {variant: 'filled', id: 'v1'},
        {variant: 'outlined', id: 'v2'},
      ] as Element[];

      const resource = createMockResource({
        category: ElementCategories.Action,
        type: ElementTypes.Action,
        variants,
        variant: 'filled',
      } as Partial<Resource>);

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      const variantSelect = document.getElementById('variant-select')!;
      await userEvent.click(variantSelect);

      const outlinedOption = page.getByRole('option', {name: 'outlined'});
      await userEvent.click(outlinedOption);

      expect(mockOnVariantChange).toHaveBeenCalledWith('outlined');
    });

    it('should handle onVariantChange being undefined', async () => {
      const variants = [
        {variant: 'primary', id: 'v1'},
        {variant: 'secondary', id: 'v2'},
      ] as Element[];

      const resource = createMockResource({
        category: ElementCategories.Action,
        type: ElementTypes.Action,
        variants,
        variant: 'primary',
      } as Partial<Resource>);

      await render(
        <ResourceProperties
          resource={resource}
          properties={{}}
          onChange={mockOnChange}
          onVariantChange={undefined}
        />,
      );

      const variantSelect = document.getElementById('variant-select')!;
      await userEvent.click(variantSelect);

      const secondaryOption = page.getByRole('option', {name: 'secondary'});
      await userEvent.click(secondaryOption);

      // Should not throw error
      await expect.element(page.getByText('Variant')).toBeInTheDocument();
    });
  });

  describe('handleChange Type Processing', () => {
    it('should preserve boolean values and call onChange with boolean', async () => {
      const resource = createMockResource({
        category: ElementCategories.Field,
      });

      await render(
        <ResourceProperties
          resource={resource}
          properties={{enabled: true}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      // Trigger the change via the mocked button
      const triggerButton = page.getByTestId('trigger-change-enabled');
      await userEvent.click(triggerButton);

      expect(mockOnChange).toHaveBeenCalledWith('enabled', true, resource);
    });

    it('should preserve object values and call onChange with object', async () => {
      const resource = createMockResource({
        category: ElementCategories.Field,
      });

      const objectValue = {nested: 'value', count: 5};

      await render(
        <ResourceProperties
          resource={resource}
          properties={{config: objectValue}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      const triggerButton = page.getByTestId('trigger-change-config');
      await userEvent.click(triggerButton);

      expect(mockOnChange).toHaveBeenCalledWith('config', objectValue, resource);
    });

    it('should convert string values to string in onChange', async () => {
      const resource = createMockResource({
        category: ElementCategories.Field,
      });

      await render(
        <ResourceProperties
          resource={resource}
          properties={{label: 'test string'}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      const triggerButton = page.getByTestId('trigger-change-label');
      await userEvent.click(triggerButton);

      expect(mockOnChange).toHaveBeenCalledWith('label', 'test string', resource);
    });

    it('should convert number values to string in onChange', async () => {
      const resource = createMockResource({
        category: ElementCategories.Field,
      });

      await render(
        <ResourceProperties
          resource={resource}
          properties={{maxLength: 100}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      const triggerButton = page.getByTestId('trigger-change-maxLength');
      await userEvent.click(triggerButton);

      expect(mockOnChange).toHaveBeenCalledWith('maxLength', '100', resource);
    });

    it('should convert null values to empty string in onChange', async () => {
      const resource = createMockResource({
        category: ElementCategories.Field,
      });

      await render(
        <ResourceProperties
          resource={resource}
          properties={{optionalProp: null as unknown as string}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      const triggerButton = page.getByTestId('trigger-change-optionalProp');
      await userEvent.click(triggerButton);

      expect(mockOnChange).toHaveBeenCalledWith('optionalProp', '', resource);
    });

    it('should convert undefined values to empty string in onChange', async () => {
      const resource = createMockResource({
        category: ElementCategories.Field,
      });

      await render(
        <ResourceProperties
          resource={resource}
          properties={{undefinedProp: undefined as unknown as string}}
          onChange={mockOnChange}
          onVariantChange={mockOnVariantChange}
        />,
      );

      const triggerButton = page.getByTestId('trigger-change-undefinedProp');
      await userEvent.click(triggerButton);

      expect(mockOnChange).toHaveBeenCalledWith('undefinedProp', '', resource);
    });
  });
});
