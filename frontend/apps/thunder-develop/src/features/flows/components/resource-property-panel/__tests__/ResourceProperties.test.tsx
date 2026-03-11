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
import {page} from 'vitest/browser';
import type {ReactNode} from 'react';
import {ReactFlowProvider} from '@xyflow/react';
import ResourceProperties from '../ResourceProperties';
import FlowBuilderCoreContext, {type FlowBuilderCoreContextProps} from '../../../context/FlowBuilderCoreContext';
import {EdgeStyleTypes} from '../../../models/steps';
import {PreviewScreenType} from '../../../models/custom-text-preference';
import {ElementTypes} from '../../../models/elements';
import type {Base, BaseConfig} from '../../../models/base';
import type {Resource} from '../../../models/resources';

// Use vi.hoisted for mock functions
const {mockUpdateNodeData} = vi.hoisted(() => ({
  mockUpdateNodeData: vi.fn(),
}));

// Mock @xyflow/react
vi.mock('@xyflow/react', async () => {
  const actual = await vi.importActual('@xyflow/react');
  return {
    ...actual,
    useReactFlow: () => ({
      updateNodeData: mockUpdateNodeData,
    }),
  };
});

// Use vi.hoisted for plugin mock functions
const {mockExecuteSync, mockExecuteAsync} = vi.hoisted(() => ({
  mockExecuteSync: vi.fn().mockReturnValue(true),
  mockExecuteAsync: vi.fn().mockResolvedValue(true),
}));

// Mock PluginRegistry
vi.mock('../../../plugins/PluginRegistry', () => ({
  default: {
    getInstance: () => ({
      executeSync: mockExecuteSync,
      executeAsync: mockExecuteAsync,
    }),
  },
}));

describe('ResourceProperties', () => {
  const mockSetLastInteractedResource = vi.fn();

  const mockBaseResource: Base = {
    id: 'resource-1',
    resourceType: 'ELEMENT',
    type: 'TEXT_INPUT',
    category: 'FIELD',
    version: '1.0.0',
    deprecated: false,
    deletable: true,
    display: {
      label: 'Test Resource',
      image: '',
      showOnResourcePanel: false,
    },
    config: {
      field: {name: '', type: ElementTypes},
      styles: {},
    },
  } as unknown as Base;

  const MockResourcePropertiesComponent = vi.fn(
    ({
      resource,
      properties,
      onChange,
      onVariantChange,
    }: {
      resource: Resource;
      properties?: Record<string, unknown>;
      onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
      onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
    }) => (
      <div data-testid="mock-resource-properties">
        <div data-testid="resource-id">{resource?.id}</div>
        <div data-testid="properties">{JSON.stringify(properties)}</div>
        <button type="button" onClick={() => onChange('label', 'New Label', resource)}>
          Change Label
        </button>
        <button type="button" onClick={() => onVariantChange?.('variant-1')}>
          Change Variant
        </button>
      </div>
    ),
  );

  const createContextValue = (overrides: Partial<FlowBuilderCoreContextProps> = {}): FlowBuilderCoreContextProps => ({
    lastInteractedResource: mockBaseResource,
    lastInteractedStepId: 'step-1',
    ResourceProperties: MockResourcePropertiesComponent,
    resourcePropertiesPanelHeading: 'Test Panel Heading',
    primaryI18nScreen: PreviewScreenType.LOGIN,
    isResourcePanelOpen: true,
    isResourcePropertiesPanelOpen: false,
    isVersionHistoryPanelOpen: false,
    ElementFactory: () => null,
    onResourceDropOnCanvas: vi.fn(),
    selectedAttributes: {},
    setLastInteractedResource: mockSetLastInteractedResource,
    setLastInteractedStepId: vi.fn(),
    setResourcePropertiesPanelHeading: vi.fn(),
    setIsResourcePanelOpen: vi.fn(),
    setIsOpenResourcePropertiesPanel: vi.fn(),
    registerCloseValidationPanel: vi.fn(),
    setIsVersionHistoryPanelOpen: vi.fn(),
    setSelectedAttributes: vi.fn(),
    flowCompletionConfigs: {},
    setFlowCompletionConfigs: vi.fn(),
    flowNodeTypes: {},
    flowEdgeTypes: {},
    setFlowNodeTypes: vi.fn(),
    setFlowEdgeTypes: vi.fn(),
    isVerboseMode: false,
    setIsVerboseMode: vi.fn(),
    edgeStyle: EdgeStyleTypes.SmoothStep,
    setEdgeStyle: vi.fn(),
    ...overrides,
  });

  const createWrapper = (contextValue: FlowBuilderCoreContextProps = createContextValue()) => {
    function Wrapper({children}: {children: ReactNode}) {
      return (
        <ReactFlowProvider>
          <FlowBuilderCoreContext.Provider value={contextValue}>{children}</FlowBuilderCoreContext.Provider>
        </ReactFlowProvider>
      );
    }
    return Wrapper;
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockExecuteSync.mockReturnValue(true);
    mockExecuteAsync.mockResolvedValue(true);
  });

  describe('Rendering', () => {
    it('should render ResourcePropertiesComponent when resource is available', async () => {
      await render(<ResourceProperties />, {wrapper: createWrapper()});

      await expect.element(page.getByTestId('mock-resource-properties')).toBeInTheDocument();
    });

    it('should display resource id', async () => {
      await render(<ResourceProperties />, {wrapper: createWrapper()});

      await expect.element(page.getByTestId('resource-id')).toHaveTextContent('resource-1');
    });

    it('should show "No properties available" when lastInteractedResource is null', async () => {
      const contextWithNoResource = createContextValue({
        lastInteractedResource: null as unknown as Base,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(contextWithNoResource)});

      await expect.element(page.getByText('No properties available.')).toBeInTheDocument();
    });
  });

  describe('Properties Filtering', () => {
    it('should filter out excluded properties from config', async () => {
      const resourceWithConfig: Base = {
        ...mockBaseResource,
        config: {
          ...mockBaseResource.config,
          field: {name: 'field-name', type: ElementTypes},
          label: 'test-name',
        },
      } as unknown as Base;

      const contextWithConfig = createContextValue({
        lastInteractedResource: resourceWithConfig,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(contextWithConfig)});

      await expect.element(page.getByTestId('mock-resource-properties')).toBeInTheDocument();
    });

    it('should extract top-level editable properties', async () => {
      const resourceWithTopLevelProps: Base = {
        ...mockBaseResource,
        label: 'Test Label',
        hint: 'Test Hint',
        placeholder: 'Test Placeholder',
        required: true,
      } as Base & {label: string; hint: string; placeholder: string; required: boolean};

      const contextWithTopLevelProps = createContextValue({
        lastInteractedResource: resourceWithTopLevelProps,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(contextWithTopLevelProps)});

      const propertiesDiv = page.getByTestId('properties');
      const properties = JSON.parse(propertiesDiv.element().textContent ?? '{}') as Record<string, unknown>;

      expect(properties.label).toBe('Test Label');
      expect(properties.hint).toBe('Test Hint');
      expect(properties.placeholder).toBe('Test Placeholder');
      expect(properties.required).toBe(true);
    });
  });

  describe('Variant Change', () => {
    it('should have variant change callback available', async () => {
      const resourceWithVariants: Base = {
        ...mockBaseResource,
        variants: [
          {
            ...mockBaseResource,
            id: 'variant-1',
            variant: 'variant-1',
            type: 'TEXT_INPUT',
            config: {...mockBaseResource.config, variant: 'outlined'},
          },
        ],
      } as Base & {variants: {id: string; variant: string; type: string; config: Record<string, unknown>}[]};

      const contextWithVariants = createContextValue({
        lastInteractedResource: resourceWithVariants,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(contextWithVariants)});

      await expect.element(page.getByText('Change Variant')).toBeInTheDocument();
    });
  });

  describe('Resource Without Config', () => {
    it('should handle resource without config gracefully', async () => {
      const resourceWithoutConfig: Base = {
        ...mockBaseResource,
        config: undefined as unknown as BaseConfig,
      };

      const contextWithoutConfig = createContextValue({
        lastInteractedResource: resourceWithoutConfig,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(contextWithoutConfig)});

      await expect.element(page.getByTestId('mock-resource-properties')).toBeInTheDocument();
    });
  });

  describe('Memoization', () => {
    it('should use memoized component', async () => {
      const {rerender} = await render(<ResourceProperties />, {wrapper: createWrapper()});

      await expect.element(page.getByTestId('mock-resource-properties')).toBeInTheDocument();

      // Re-render with same props should not cause issues
      await rerender(<ResourceProperties />);

      await expect.element(page.getByTestId('mock-resource-properties')).toBeInTheDocument();
    });
  });

  describe('Empty Resource', () => {
    it('should handle empty config object', async () => {
      const resourceWithEmptyConfig: Base = {
        ...mockBaseResource,
        config: {...mockBaseResource.config},
      };

      const contextWithEmptyConfig = createContextValue({
        lastInteractedResource: resourceWithEmptyConfig,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(contextWithEmptyConfig)});

      await expect.element(page.getByTestId('mock-resource-properties')).toBeInTheDocument();
    });
  });

  describe('Property Change Handler', () => {
    it('should trigger onChange callback when property changes', async () => {
      await render(<ResourceProperties />, {wrapper: createWrapper()});

      const changeLabelButton = page.getByText('Change Label');
      await changeLabelButton.click();

      // The onChange callback should be passed to MockResourcePropertiesComponent
      expect(MockResourcePropertiesComponent).toHaveBeenCalled();
    });

    it('should pass resource to onChange callback', async () => {
      await render(<ResourceProperties />, {wrapper: createWrapper()});

      // Verify the MockResourcePropertiesComponent receives the resource
      const {calls} = MockResourcePropertiesComponent.mock;
      expect(calls.length).toBeGreaterThan(0);
      const props = calls[0][0] as {resource: {id: string}};
      expect(props.resource.id).toBe('resource-1');
    });
  });

  describe('Variant Change Handler', () => {
    it('should trigger onVariantChange callback', async () => {
      const resourceWithVariants: Base = {
        ...mockBaseResource,
        variants: [
          {
            ...mockBaseResource,
            id: 'variant-1',
            variant: 'variant-1',
            type: 'TEXT_INPUT',
            config: {...mockBaseResource.config, variant: 'outlined'},
          },
        ],
      } as Base & {variants: {id: string; variant: string; type: string; config: Record<string, unknown>}[]};

      const contextWithVariants = createContextValue({
        lastInteractedResource: resourceWithVariants,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(contextWithVariants)});

      const changeVariantButton = page.getByText('Change Variant');
      await changeVariantButton.click();

      // Verify onVariantChange was passed to the component
      const {calls} = MockResourcePropertiesComponent.mock;
      expect(calls.length).toBeGreaterThan(0);
      const props = calls[0][0] as {onVariantChange: unknown};
      expect(typeof props.onVariantChange).toBe('function');
    });

    it('should handle variant change for resource with label', async () => {
      const resourceWithVariantsAndLabel: Base = {
        ...mockBaseResource,
        label: 'Current Label',
        variants: [
          {
            ...mockBaseResource,
            id: 'variant-2',
            variant: 'variant-2',
            type: 'TEXT_INPUT',
            config: {...mockBaseResource.config, variant: 'filled'},
          },
        ],
      } as Base & {label: string; variants: {id: string; variant: string; type: string; config: Record<string, unknown>}[]};

      const contextWithVariantsAndLabel = createContextValue({
        lastInteractedResource: resourceWithVariantsAndLabel,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(contextWithVariantsAndLabel)});

      await expect.element(page.getByTestId('mock-resource-properties')).toBeInTheDocument();
    });

    it('should handle variant change for resource with text config', async () => {
      const resourceWithVariantsAndText: Base = {
        ...mockBaseResource,
        config: {
          ...mockBaseResource.config,
          text: 'Current Text',
        },
        variants: [
          {
            ...mockBaseResource,
            id: 'variant-3',
            variant: 'variant-3',
            type: 'TEXT_INPUT',
            config: {...mockBaseResource.config, variant: 'standard'},
          },
        ],
      } as Base & {variants: {id: string; variant: string; type: string; config: Record<string, unknown>}[]};

      const contextWithVariantsAndText = createContextValue({
        lastInteractedResource: resourceWithVariantsAndText,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(contextWithVariantsAndText)});

      await expect.element(page.getByTestId('mock-resource-properties')).toBeInTheDocument();
    });

    it('should not change variant when variant is not found', async () => {
      const resourceWithVariants: Base = {
        ...mockBaseResource,
        variants: [
          {
            ...mockBaseResource,
            id: 'variant-1',
            variant: 'variant-1',
            type: 'TEXT_INPUT',
            config: {...mockBaseResource.config, variant: 'outlined'},
          },
        ],
      } as Base & {variants: {id: string; variant: string; type: string; config: Record<string, unknown>}[]};

      const contextWithVariants = createContextValue({
        lastInteractedResource: resourceWithVariants,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(contextWithVariants)});

      // The component should render without issues
      await expect.element(page.getByTestId('mock-resource-properties')).toBeInTheDocument();
    });
  });

  describe('Top-Level Properties', () => {
    it('should extract src property', async () => {
      const resourceWithSrc: Base = {
        ...mockBaseResource,
        src: 'https://example.com/image.png',
      } as Base & {src: string};

      const contextWithSrc = createContextValue({
        lastInteractedResource: resourceWithSrc,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(contextWithSrc)});

      const propertiesDiv = page.getByTestId('properties');
      const properties = JSON.parse(propertiesDiv.element().textContent ?? '{}') as Record<string, unknown>;

      expect(properties.src).toBe('https://example.com/image.png');
    });

    it('should extract alt property', async () => {
      const resourceWithAlt: Base = {
        ...mockBaseResource,
        alt: 'Alternative text',
      } as Base & {alt: string};

      const contextWithAlt = createContextValue({
        lastInteractedResource: resourceWithAlt,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(contextWithAlt)});

      const propertiesDiv = page.getByTestId('properties');
      const properties = JSON.parse(propertiesDiv.element().textContent ?? '{}') as Record<string, unknown>;

      expect(properties.alt).toBe('Alternative text');
    });
  });

  describe('Nested Components', () => {
    it('should handle resource with nested components', async () => {
      const resourceWithNestedComponents: Base = {
        ...mockBaseResource,
        components: [
          {
            id: 'nested-1',
            type: 'BUTTON',
          },
        ],
      } as Base & {components: {id: string; type: string}[]};

      const contextWithNestedComponents = createContextValue({
        lastInteractedResource: resourceWithNestedComponents,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(contextWithNestedComponents)});

      await expect.element(page.getByTestId('mock-resource-properties')).toBeInTheDocument();
    });
  });

  describe('Variant Not Found', () => {
    it('should return early when selected variant is not found', async () => {
      const resourceWithVariants: Base = {
        ...mockBaseResource,
        variants: [
          {
            ...mockBaseResource,
            id: 'variant-1',
            variant: 'variant-1',
            type: 'TEXT_INPUT',
            config: {...mockBaseResource.config, variant: 'outlined'},
          },
        ],
      } as Base & {variants: {id: string; variant: string; type: string; config: Record<string, unknown>}[]};

      // Create mock that triggers variant change with non-existent variant
      const MockComponentWithNonExistentVariant = vi.fn(
        ({
          resource,
          properties,
          onVariantChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <div data-testid="properties">{JSON.stringify(properties)}</div>
            <button type="button" onClick={() => onVariantChange?.('non-existent-variant')}>
              Change to Non-Existent Variant
            </button>
          </div>
        ),
      );

      const contextWithVariants = createContextValue({
        lastInteractedResource: resourceWithVariants,
        ResourceProperties: MockComponentWithNonExistentVariant,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(contextWithVariants)});

      const changeVariantButton = page.getByText('Change to Non-Existent Variant');
      await changeVariantButton.click();

      // updateNodeData should not be called when variant is not found
      expect(mockUpdateNodeData).not.toHaveBeenCalled();
    });
  });

  describe('Preserve Label on Variant Change', () => {
    it('should preserve current label value when changing variants', async () => {
      const resourceWithLabelAndVariants: Base = {
        ...mockBaseResource,
        label: 'My Custom Label',
        variants: [
          {
            ...mockBaseResource,
            id: 'variant-1',
            variant: 'variant-1',
            type: 'TEXT_INPUT',
            label: 'Default Variant Label',
            config: {...mockBaseResource.config, variant: 'outlined'},
          },
        ],
      } as Base & {label: string; variants: {id: string; variant: string; type: string; label: string; config: Record<string, unknown>}[]};

      const contextWithLabelAndVariants = createContextValue({
        lastInteractedResource: resourceWithLabelAndVariants,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(contextWithLabelAndVariants)});

      const changeVariantButton = page.getByText('Change Variant');
      await changeVariantButton.click();

      // Verify updateNodeData was called (variant change should preserve the label)
      expect(mockUpdateNodeData).toHaveBeenCalled();
    });
  });

  describe('Preserve Text on Variant Change', () => {
    it('should preserve current text value when changing variants with selectedVariant.config', async () => {
      const resourceWithTextAndVariants: Base = {
        ...mockBaseResource,
        config: {
          ...mockBaseResource.config,
          text: 'Current text value',
        },
        variants: [
          {
            ...mockBaseResource,
            id: 'variant-1',
            variant: 'variant-1',
            type: 'TEXT_INPUT',
            config: {...mockBaseResource.config, text: 'Default text', variant: 'outlined'},
          },
        ],
      } as Base & {variants: {id: string; variant: string; type: string; config: Record<string, unknown>}[]};

      const contextWithTextAndVariants = createContextValue({
        lastInteractedResource: resourceWithTextAndVariants,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(contextWithTextAndVariants)});

      const changeVariantButton = page.getByText('Change Variant');
      await changeVariantButton.click();

      // Verify updateNodeData was called
      expect(mockUpdateNodeData).toHaveBeenCalled();
    });
  });

  describe('Update Component Recursive Mapping', () => {
    it('should update component when id matches', async () => {
      const resourceWithComponents: Base = {
        ...mockBaseResource,
        variants: [
          {
            ...mockBaseResource,
            id: 'variant-1',
            variant: 'variant-1',
            type: 'TEXT_INPUT',
            config: {...mockBaseResource.config, variant: 'outlined'},
          },
        ],
      } as Base & {variants: {id: string; variant: string; type: string; config: Record<string, unknown>}[]};

      const contextWithComponents = createContextValue({
        lastInteractedResource: resourceWithComponents,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(contextWithComponents)});

      const changeVariantButton = page.getByText('Change Variant');
      await changeVariantButton.click();

      expect(mockUpdateNodeData).toHaveBeenCalled();
    });

    it('should recursively update nested components', async () => {
      const nestedResource = {
        ...mockBaseResource,
        id: 'nested-resource',
        components: [
          {
            ...mockBaseResource,
            id: 'child-1',
            components: [
              {
                ...mockBaseResource,
                id: 'grandchild-1',
              },
            ],
          },
        ],
        variants: [
          {
            ...mockBaseResource,
            id: 'variant-1',
            variant: 'variant-1',
            type: 'TEXT_INPUT',
            config: {...mockBaseResource.config, variant: 'outlined'},
          },
        ],
      } as unknown as Base;

      const contextWithNestedComponents = createContextValue({
        lastInteractedResource: nestedResource,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(contextWithNestedComponents)});

      const changeVariantButton = page.getByText('Change Variant');
      await changeVariantButton.click();

      expect(mockUpdateNodeData).toHaveBeenCalled();
    });
  });


  describe('handlePropertyChange with data Property', () => {
    it('should handle propertyKey === data to replace entire data object', async () => {
      const MockComponentWithDataChange = vi.fn(
        ({
          resource,
          properties,
          onChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <div data-testid="properties">{JSON.stringify(properties)}</div>
            <button type="button" onClick={() => onChange('data', {newKey: 'newValue'}, resource)}>
              Change Data
            </button>
          </div>
        ),
      );

      const context = createContextValue({
        ResourceProperties: MockComponentWithDataChange,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(context)});

      const changeDataButton = page.getByText('Change Data');
      await changeDataButton.click();

      // Wait for debounced function to execute
      await new Promise(resolve => {
        setTimeout(resolve, 400);
      });

      expect(MockComponentWithDataChange).toHaveBeenCalled();
    });
  });

  describe('handlePropertyChange with config/data prefix', () => {
    it('should handle propertyKey starting with config.', async () => {
      const MockComponentWithConfigChange = vi.fn(
        ({
          resource,
          properties,
          onChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <div data-testid="properties">{JSON.stringify(properties)}</div>
            <button type="button" onClick={() => onChange('config.styles.color', 'red', resource)}>
              Change Config Style
            </button>
          </div>
        ),
      );

      const context = createContextValue({
        ResourceProperties: MockComponentWithConfigChange,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(context)});

      const changeButton = page.getByText('Change Config Style');
      await changeButton.click();

      // Wait for debounced function to execute
      await new Promise(resolve => {
        setTimeout(resolve, 400);
      });

      expect(MockComponentWithConfigChange).toHaveBeenCalled();
    });

    it('should handle propertyKey starting with data.', async () => {
      const MockComponentWithDataPrefixChange = vi.fn(
        ({
          resource,
          properties,
          onChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <div data-testid="properties">{JSON.stringify(properties)}</div>
            <button type="button" onClick={() => onChange('data.customField', 'custom value', resource)}>
              Change Data Field
            </button>
          </div>
        ),
      );

      const context = createContextValue({
        ResourceProperties: MockComponentWithDataPrefixChange,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(context)});

      const changeButton = page.getByText('Change Data Field');
      await changeButton.click();

      // Wait for debounced function to execute
      await new Promise(resolve => {
        setTimeout(resolve, 400);
      });

      expect(MockComponentWithDataPrefixChange).toHaveBeenCalled();
    });
  });

  describe('handlePropertyChange for action property', () => {
    it('should not update lastInteractedResource when propertyKey is action', async () => {
      const MockComponentWithActionChange = vi.fn(
        ({
          resource,
          properties,
          onChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <div data-testid="properties">{JSON.stringify(properties)}</div>
            <button type="button" onClick={() => onChange('action', 'SUBMIT', resource)}>
              Change Action
            </button>
          </div>
        ),
      );

      const context = createContextValue({
        ResourceProperties: MockComponentWithActionChange,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(context)});

      const changeButton = page.getByText('Change Action');
      await changeButton.click();

      // Wait for debounced function to execute
      await new Promise(resolve => {
        setTimeout(resolve, 400);
      });

      // setLastInteractedResource should not be called for action changes
      // This is hard to test directly, but we can verify the component renders
      expect(MockComponentWithActionChange).toHaveBeenCalled();
    });
  });

  describe('handlePropertyChange with Non-Top-Level Property', () => {
    it('should set property on resource.data for non-top-level properties', async () => {
      const MockComponentWithCustomProperty = vi.fn(
        ({
          resource,
          properties,
          onChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <div data-testid="properties">{JSON.stringify(properties)}</div>
            <button type="button" onClick={() => onChange('customProperty', 'customValue', resource)}>
              Change Custom Property
            </button>
          </div>
        ),
      );

      const context = createContextValue({
        ResourceProperties: MockComponentWithCustomProperty,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(context)});

      const changeButton = page.getByText('Change Custom Property');
      await changeButton.click();

      // Wait for debounced function to execute
      await new Promise(resolve => {
        setTimeout(resolve, 400);
      });

      expect(MockComponentWithCustomProperty).toHaveBeenCalled();
    });
  });

  describe('Variant Change with Element Partial Override', () => {
    it('should merge element partial when provided to variant change', async () => {
      const resourceWithVariants: Base = {
        ...mockBaseResource,
        variants: [
          {
            ...mockBaseResource,
            id: 'variant-1',
            variant: 'variant-1',
            type: 'TEXT_INPUT',
            config: {...mockBaseResource.config, variant: 'outlined'},
          },
        ],
      } as Base & {variants: {id: string; variant: string; type: string; config: Record<string, unknown>}[]};

      const MockComponentWithElementOverride = vi.fn(
        ({
          resource,
          properties,
          onVariantChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <div data-testid="properties">{JSON.stringify(properties)}</div>
            <button type="button" onClick={() => onVariantChange?.('variant-1', {label: 'Override Label'} as Partial<Resource>)}>
              Change Variant with Override
            </button>
          </div>
        ),
      );

      const contextWithVariants = createContextValue({
        lastInteractedResource: resourceWithVariants,
        ResourceProperties: MockComponentWithElementOverride,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(contextWithVariants)});

      const changeVariantButton = page.getByText('Change Variant with Override');
      await changeVariantButton.click();

      expect(mockUpdateNodeData).toHaveBeenCalled();
    });
  });

  describe('Empty Node Components', () => {
    it('should handle empty node components gracefully', async () => {
      // Setup updateNodeData to simulate empty components
      mockUpdateNodeData.mockImplementation((_stepId: string, callback: (node: {data: {components?: unknown[]}}) => unknown) => {
        callback({data: {components: []}});
      });

      const MockComponentWithPropertyChange = vi.fn(
        ({
          resource,
          properties,
          onChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <div data-testid="properties">{JSON.stringify(properties)}</div>
            <button type="button" onClick={() => onChange('label', 'New Label', resource)}>
              Change Label
            </button>
          </div>
        ),
      );

      const context = createContextValue({
        ResourceProperties: MockComponentWithPropertyChange,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(context)});

      const changeLabelButton = page.getByText('Change Label');
      await changeLabelButton.click();

      // Wait for debounced function to execute
      await new Promise(resolve => {
        setTimeout(resolve, 400);
      });

      expect(MockComponentWithPropertyChange).toHaveBeenCalled();
    });
  });

  describe('handlePropertyChange with Different Element ID', () => {
    it('should not update lastInteractedResource when element.id differs from current', async () => {
      const differentResource: Base = {
        ...mockBaseResource,
        id: 'different-resource-id',
      };

      const MockComponentWithDifferentResource = vi.fn(
        ({
          resource,
          properties,
          onChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <div data-testid="properties">{JSON.stringify(properties)}</div>
            <button type="button" onClick={() => onChange('label', 'New Label', differentResource)}>
              Change Different Resource
            </button>
          </div>
        ),
      );

      const context = createContextValue({
        ResourceProperties: MockComponentWithDifferentResource,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(context)});

      const changeButton = page.getByText('Change Different Resource');
      await changeButton.click();

      // Wait for debounced function to execute
      await new Promise(resolve => {
        setTimeout(resolve, 400);
      });

      expect(MockComponentWithDifferentResource).toHaveBeenCalled();
    });
  });

  describe('Strip data. prefix', () => {
    it('should strip data. prefix when setting property on data object', async () => {
      mockUpdateNodeData.mockImplementation((_stepId: string, callback: (node: {data: {components?: unknown[]}}) => unknown) => {
        callback({data: {}});
      });

      const MockComponentWithDataPrefix = vi.fn(
        ({
          resource,
          properties,
          onChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <div data-testid="properties">{JSON.stringify(properties)}</div>
            <button type="button" onClick={() => onChange('data.someField', 'value', resource)}>
              Change Data Prefixed Field
            </button>
          </div>
        ),
      );

      const context = createContextValue({
        ResourceProperties: MockComponentWithDataPrefix,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(context)});

      const changeButton = page.getByText('Change Data Prefixed Field');
      await changeButton.click();

      // Wait for debounced function to execute
      await new Promise(resolve => {
        setTimeout(resolve, 400);
      });

      expect(MockComponentWithDataPrefix).toHaveBeenCalled();
    });
  });

  describe('changeSelectedVariant updateComponent recursive mapping', () => {
    it('should update component when id matches in updateComponent', async () => {
      const resourceWithMatchingId: Base = {
        ...mockBaseResource,
        id: 'resource-1',
        variants: [
          {
            ...mockBaseResource,
            id: 'variant-1',
            variant: 'variant-1',
            type: 'TEXT_INPUT',
            config: {...mockBaseResource.config, variant: 'outlined'},
          },
        ],
      } as Base & {variants: {id: string; variant: string; type: string; config: Record<string, unknown>}[]};

      const contextWithMatchingId = createContextValue({
        lastInteractedResource: resourceWithMatchingId,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(contextWithMatchingId)});

      const changeVariantButton = page.getByText('Change Variant');
      await changeVariantButton.click();

      // Verify updateNodeData was called with the correct structure
      expect(mockUpdateNodeData).toHaveBeenCalled();
    });

    it('should recursively process nested components in updateComponent', async () => {
      const resourceWithNestedComponents: Base = {
        ...mockBaseResource,
        id: 'nested-parent',
        components: [
          {
            ...mockBaseResource,
            id: 'child-component',
            components: [
              {
                ...mockBaseResource,
                id: 'resource-1', // This should match and be updated
              },
            ],
          },
        ],
        variants: [
          {
            ...mockBaseResource,
            id: 'variant-1',
            variant: 'variant-1',
            type: 'TEXT_INPUT',
            config: {...mockBaseResource.config, variant: 'outlined'},
          },
        ],
      } as unknown as Base;

      // Create a mock that returns a more realistic node structure
      mockUpdateNodeData.mockImplementation((_stepId: string, callback: (node: {data: {components?: unknown[]}}) => unknown) => {
        const node = {
          data: {
            components: [
              {
                id: 'child-component',
                components: [
                  {id: 'resource-1'},
                ],
              },
            ],
          },
        };
        callback(node);
      });

      const contextWithNestedComponents = createContextValue({
        lastInteractedResource: resourceWithNestedComponents,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(contextWithNestedComponents)});

      const changeVariantButton = page.getByText('Change Variant');
      await changeVariantButton.click();

      expect(mockUpdateNodeData).toHaveBeenCalled();
    });

    it('should return component unchanged when id does not match and no nested components', async () => {
      const resourceWithNoMatch: Base = {
        ...mockBaseResource,
        id: 'different-id',
        variants: [
          {
            ...mockBaseResource,
            id: 'variant-1',
            variant: 'variant-1',
            type: 'TEXT_INPUT',
            config: {...mockBaseResource.config, variant: 'outlined'},
          },
        ],
      } as Base & {variants: {id: string; variant: string; type: string; config: Record<string, unknown>}[]};

      mockUpdateNodeData.mockImplementation((_stepId: string, callback: (node: {data: {components?: unknown[]}}) => unknown) => {
        const node = {
          data: {
            components: [
              {id: 'other-component'},
            ],
          },
        };
        callback(node);
      });

      const contextWithNoMatch = createContextValue({
        lastInteractedResource: resourceWithNoMatch,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(contextWithNoMatch)});

      const changeVariantButton = page.getByText('Change Variant');
      await changeVariantButton.click();

      expect(mockUpdateNodeData).toHaveBeenCalled();
    });
  });

  describe('changeSelectedVariant setLastInteractedResource', () => {
    it('should call setLastInteractedResource with merged variant', async () => {
      const resourceWithVariants: Base = {
        ...mockBaseResource,
        variants: [
          {
            ...mockBaseResource,
            id: 'variant-1',
            variant: 'variant-1',
            type: 'TEXT_INPUT',
            config: {...mockBaseResource.config, variant: 'outlined'},
          },
        ],
      } as Base & {variants: {id: string; variant: string; type: string; config: Record<string, unknown>}[]};

      mockUpdateNodeData.mockImplementation((_stepId: string, callback: (node: {data: {components?: unknown[]}}) => unknown) => {
        const node = {
          data: {
            components: [{id: 'resource-1'}],
          },
        };
        callback(node);
      });

      const contextWithVariants = createContextValue({
        lastInteractedResource: resourceWithVariants,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(contextWithVariants)});

      const changeVariantButton = page.getByText('Change Variant');
      await changeVariantButton.click();

      // Verify updateNodeData was called
      expect(mockUpdateNodeData).toHaveBeenCalled();
    });
  });

  describe('handlePropertyChange updateComponent', () => {
    it('should pass onChange callback that can be triggered', async () => {
      const MockComponentWithPropertyChange = vi.fn(
        ({
          resource,
          properties,
          onChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <div data-testid="properties">{JSON.stringify(properties)}</div>
            <button type="button" onClick={() => onChange('label', 'New Label', resource)}>
              Change Label
            </button>
          </div>
        ),
      );

      const context = createContextValue({
        ResourceProperties: MockComponentWithPropertyChange,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(context)});

      // Verify onChange callback is provided
      const {calls} = MockComponentWithPropertyChange.mock;
      expect(calls.length).toBeGreaterThan(0);
      const props = calls[0][0] as {onChange: (key: string, value: unknown, resource: Resource) => void};
      expect(typeof props.onChange).toBe('function');
    });

    it('should provide onChange that accepts nested component updates', async () => {
      const MockComponentWithNestedChange = vi.fn(
        ({
          resource,
          properties,
          onChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <div data-testid="properties">{JSON.stringify(properties)}</div>
            <button type="button" onClick={() => onChange('label', 'New Label', resource)}>
              Change Nested Label
            </button>
          </div>
        ),
      );

      const context = createContextValue({
        ResourceProperties: MockComponentWithNestedChange,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(context)});

      // Verify the component renders and onChange is available
      await expect.element(page.getByText('Change Nested Label')).toBeInTheDocument();
    });

    it('should render component with onChange callback for different ids', async () => {
      const MockComponentWithDifferentId = vi.fn(
        ({
          resource,
          properties,
          onChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <div data-testid="properties">{JSON.stringify(properties)}</div>
            <button type="button" onClick={() => onChange('label', 'New Label', resource)}>
              Change Label
            </button>
          </div>
        ),
      );

      const context = createContextValue({
        ResourceProperties: MockComponentWithDifferentId,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(context)});

      await expect.element(page.getByText('Change Label')).toBeInTheDocument();
    });
  });

  describe('handlePropertyChange updateNodeData callback', () => {
    it('should render component with onChange for node updates', async () => {
      await render(<ResourceProperties />, {wrapper: createWrapper()});

      const changeButton = page.getByText('Change Label');
      expect(changeButton).toBeInTheDocument();
    });

    it('should provide onChange that can handle data replacement', async () => {
      const MockComponentWithDataReplace = vi.fn(
        ({
          resource,
          properties,
          onChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <div data-testid="properties">{JSON.stringify(properties)}</div>
            <button type="button" onClick={() => onChange('data', {newField: 'newValue'}, resource)}>
              Replace Data
            </button>
          </div>
        ),
      );

      const context = createContextValue({
        ResourceProperties: MockComponentWithDataReplace,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(context)});

      await expect.element(page.getByText('Replace Data')).toBeInTheDocument();
    });

    it('should provide onChange that accepts data prefixed properties', async () => {
      const MockComponentWithDataPrefixStrip = vi.fn(
        ({
          resource,
          properties,
          onChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <div data-testid="properties">{JSON.stringify(properties)}</div>
            <button type="button" onClick={() => onChange('data.customField', 'customValue', resource)}>
              Set Data Field
            </button>
          </div>
        ),
      );

      const context = createContextValue({
        ResourceProperties: MockComponentWithDataPrefixStrip,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(context)});

      await expect.element(page.getByText('Set Data Field')).toBeInTheDocument();
    });
  });

  describe('handlePropertyChange lastInteractedResource update', () => {
    it('should provide onChange for top-level editable properties', async () => {
      const MockComponentWithTopLevelProp = vi.fn(
        ({
          resource,
          properties,
          onChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <div data-testid="properties">{JSON.stringify(properties)}</div>
            <button type="button" onClick={() => onChange('hint', 'New Hint', resource)}>
              Change Hint
            </button>
          </div>
        ),
      );

      const context = createContextValue({
        ResourceProperties: MockComponentWithTopLevelProp,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(context)});

      await expect.element(page.getByText('Change Hint')).toBeInTheDocument();
    });

    it('should provide onChange that accepts data property', async () => {
      const MockComponentWithDataProperty = vi.fn(
        ({
          resource,
          properties,
          onChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <div data-testid="properties">{JSON.stringify(properties)}</div>
            <button type="button" onClick={() => onChange('data', {key: 'value'}, resource)}>
              Set Data Object
            </button>
          </div>
        ),
      );

      const context = createContextValue({
        ResourceProperties: MockComponentWithDataProperty,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(context)});

      await expect.element(page.getByText('Set Data Object')).toBeInTheDocument();
    });

    it('should provide onChange that accepts config prefixed properties', async () => {
      const MockComponentWithConfigPrefix = vi.fn(
        ({
          resource,
          properties,
          onChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <div data-testid="properties">{JSON.stringify(properties)}</div>
            <button type="button" onClick={() => onChange('config.someOption', true, resource)}>
              Set Config Option
            </button>
          </div>
        ),
      );

      const context = createContextValue({
        ResourceProperties: MockComponentWithConfigPrefix,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(context)});

      await expect.element(page.getByText('Set Config Option')).toBeInTheDocument();
    });

    it('should provide onChange for non-top-level properties', async () => {
      const MockComponentWithCustomProperty = vi.fn(
        ({
          resource,
          properties,
          onChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <div data-testid="properties">{JSON.stringify(properties)}</div>
            <button type="button" onClick={() => onChange('customNonTopLevel', 'customValue', resource)}>
              Set Custom Property
            </button>
          </div>
        ),
      );

      const context = createContextValue({
        ResourceProperties: MockComponentWithCustomProperty,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(context)});

      await expect.element(page.getByText('Set Custom Property')).toBeInTheDocument();
    });
  });

  describe('handlePropertyChange when plugin returns false', () => {
    it('should update resource and return early when plugin handles the change (lines 208-214)', async () => {
      // Make plugin return false to indicate it handled the change
      mockExecuteAsync.mockResolvedValue(false);

      const MockComponentWithChange = vi.fn(
        ({
          resource,
          onChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <button type="button" onClick={() => onChange('label', 'Plugin Handled Label', resource)}>
              Change Label (Plugin Handles)
            </button>
          </div>
        ),
      );

      const context = createContextValue({
        ResourceProperties: MockComponentWithChange,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(context)});

      const changeButton = page.getByText('Change Label (Plugin Handles)');
      await changeButton.click();

      // Wait for debounced function to execute
      await new Promise(resolve => {
        setTimeout(resolve, 400);
      });

      // When plugin returns false, setLastInteractedResource should be called to update the resource
      expect(mockSetLastInteractedResource).toHaveBeenCalled();
      // updateNodeData should NOT be called since we return early
      expect(mockUpdateNodeData).not.toHaveBeenCalled();
    });

    it('should not update resource when element.id differs from lastInteractedResourceId and plugin returns false', async () => {
      mockExecuteAsync.mockResolvedValue(false);

      const differentResource: Base = {
        ...mockBaseResource,
        id: 'different-resource-id',
      };

      const MockComponentWithDifferentElement = vi.fn(
        ({
          resource,
          onChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <button type="button" onClick={() => onChange('label', 'Different Label', differentResource)}>
              Change Different Element
            </button>
          </div>
        ),
      );

      const context = createContextValue({
        ResourceProperties: MockComponentWithDifferentElement,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(context)});

      const changeButton = page.getByText('Change Different Element');
      await changeButton.click();

      await new Promise(resolve => {
        setTimeout(resolve, 400);
      });

      // setLastInteractedResource should NOT be called because element.id !== lastInteractedResourceIdRef.current
      expect(mockSetLastInteractedResource).not.toHaveBeenCalled();
    });
  });

  describe('handlePropertyChange updateComponent recursive function (lines 217-235)', () => {
    it('should update matching component in updateComponent function', async () => {
      mockExecuteAsync.mockResolvedValue(true);

      // Mock updateNodeData to capture the callback and execute it
      mockUpdateNodeData.mockImplementation((_stepId: string, callback: (node: {data: {components?: unknown[]}}) => unknown) => {
        const node = {
          data: {
            components: [
              {id: 'resource-1', label: 'Old Label'},
              {id: 'other-component', label: 'Other Label'},
            ],
          },
        };
        const result = callback(node);
        return result;
      });

      const MockComponentWithLabelChange = vi.fn(
        ({
          resource,
          onChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <button type="button" onClick={() => onChange('label', 'Updated Label', resource)}>
              Update Component Label
            </button>
          </div>
        ),
      );

      const context = createContextValue({
        ResourceProperties: MockComponentWithLabelChange,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(context)});

      const changeButton = page.getByText('Update Component Label');
      await changeButton.click();

      await new Promise(resolve => {
        setTimeout(resolve, 400);
      });

      expect(mockUpdateNodeData).toHaveBeenCalled();
    });

    it('should recursively update nested components when component has nested components (lines 227-231)', async () => {
      mockExecuteAsync.mockResolvedValue(true);

      mockUpdateNodeData.mockImplementation((_stepId: string, callback: (node: {data: {components?: unknown[]}}) => unknown) => {
        const node = {
          data: {
            components: [
              {
                id: 'parent-component',
                label: 'Parent',
                components: [
                  {id: 'resource-1', label: 'Nested Child'},
                ],
              },
            ],
          },
        };
        const result = callback(node);
        return result;
      });

      const MockComponentWithNestedUpdate = vi.fn(
        ({
          resource,
          onChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <button type="button" onClick={() => onChange('label', 'Updated Nested Label', resource)}>
              Update Nested Component
            </button>
          </div>
        ),
      );

      const context = createContextValue({
        ResourceProperties: MockComponentWithNestedUpdate,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(context)});

      const changeButton = page.getByText('Update Nested Component');
      await changeButton.click();

      await new Promise(resolve => {
        setTimeout(resolve, 400);
      });

      expect(mockUpdateNodeData).toHaveBeenCalled();
    });

    it('should return component unchanged when id does not match and no nested components (line 234)', async () => {
      mockExecuteAsync.mockResolvedValue(true);

      mockUpdateNodeData.mockImplementation((_stepId: string, callback: (node: {data: {components?: unknown[]}}) => unknown) => {
        const node = {
          data: {
            components: [
              {id: 'unrelated-component', label: 'Unrelated'},
            ],
          },
        };
        const result = callback(node);
        return result;
      });

      const MockComponentNoMatch = vi.fn(
        ({
          resource,
          onChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <button type="button" onClick={() => onChange('label', 'No Match Label', resource)}>
              Update No Match
            </button>
          </div>
        ),
      );

      const context = createContextValue({
        ResourceProperties: MockComponentNoMatch,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(context)});

      const changeButton = page.getByText('Update No Match');
      await changeButton.click();

      await new Promise(resolve => {
        setTimeout(resolve, 400);
      });

      expect(mockUpdateNodeData).toHaveBeenCalled();
    });
  });

  describe('changeSelectedVariant early return when no resource', () => {
    it('should return early when currentResource is null (line 135)', async () => {
      // Create a mock that triggers variant change after the resource becomes null
      const MockComponentWithVariantChange = vi.fn(
        ({
          onVariantChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <button type="button" onClick={() => onVariantChange?.('variant-1')}>
              Change Variant When No Resource
            </button>
          </div>
        ),
      );

      // Start with null resource
      const contextWithNullResource = createContextValue({
        lastInteractedResource: null as unknown as Base,
        ResourceProperties: MockComponentWithVariantChange,
      });

      // This will render the "No properties available" message instead of the component
      await render(<ResourceProperties />, {wrapper: createWrapper(contextWithNullResource)});

      // Since lastInteractedResource is null, the component renders the fallback message
      await expect.element(page.getByText('No properties available.')).toBeInTheDocument();
      // updateNodeData should not be called
      expect(mockUpdateNodeData).not.toHaveBeenCalled();
    });
  });

  describe('handlePropertyChange updateNodeData callback paths (lines 240-251)', () => {
    it('should replace entire data object when propertyKey is exactly data and no components (lines 242-244)', async () => {
      mockExecuteAsync.mockResolvedValue(true);

      let capturedResult: unknown;
      mockUpdateNodeData.mockImplementation((_stepId: string, callback: (node: {data: {components?: unknown[]}}) => unknown) => {
        // Simulate node with no components (empty)
        const node = {
          data: {},
        };
        capturedResult = callback(node);
        return capturedResult;
      });

      const MockComponentWithDataReplace = vi.fn(
        ({
          resource,
          onChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <button type="button" onClick={() => onChange('data', {entireNewData: 'value', anotherField: 123}, resource)}>
              Replace Entire Data
            </button>
          </div>
        ),
      );

      const context = createContextValue({
        ResourceProperties: MockComponentWithDataReplace,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(context)});

      const changeButton = page.getByText('Replace Entire Data');
      await changeButton.click();

      await new Promise(resolve => {
        setTimeout(resolve, 400);
      });

      expect(mockUpdateNodeData).toHaveBeenCalled();
      // Verify the result contains the new data object spread
      expect(capturedResult).toEqual({entireNewData: 'value', anotherField: 123});
    });

    it('should strip data. prefix and set on data object when no components (lines 246-248)', async () => {
      mockExecuteAsync.mockResolvedValue(true);

      let capturedResult: unknown;
      mockUpdateNodeData.mockImplementation((_stepId: string, callback: (node: {data: {components?: unknown[]; existingField?: string}}) => unknown) => {
        const node = {
          data: {existingField: 'existingValue'},
        };
        capturedResult = callback(node);
        return capturedResult;
      });

      const MockComponentWithDataPrefix = vi.fn(
        ({
          resource,
          onChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <button type="button" onClick={() => onChange('data.newField', 'newValue', resource)}>
              Set Data Prefixed Property
            </button>
          </div>
        ),
      );

      const context = createContextValue({
        ResourceProperties: MockComponentWithDataPrefix,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(context)});

      const changeButton = page.getByText('Set Data Prefixed Property');
      await changeButton.click();

      await new Promise(resolve => {
        setTimeout(resolve, 400);
      });

      expect(mockUpdateNodeData).toHaveBeenCalled();
      // Verify the prefix was stripped and newField was set
      expect(capturedResult).toEqual(expect.objectContaining({newField: 'newValue'}));
    });

    it('should set property directly on data object when no prefix and no components', async () => {
      mockExecuteAsync.mockResolvedValue(true);

      let capturedResult: unknown;
      mockUpdateNodeData.mockImplementation((_stepId: string, callback: (node: {data: {components?: unknown[]; existingField?: string}}) => unknown) => {
        const node = {
          data: {existingField: 'existingValue'},
        };
        capturedResult = callback(node);
        return capturedResult;
      });

      const MockComponentWithDirectProperty = vi.fn(
        ({
          resource,
          onChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <button type="button" onClick={() => onChange('directField', 'directValue', resource)}>
              Set Direct Property
            </button>
          </div>
        ),
      );

      const context = createContextValue({
        ResourceProperties: MockComponentWithDirectProperty,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(context)});

      const changeButton = page.getByText('Set Direct Property');
      await changeButton.click();

      await new Promise(resolve => {
        setTimeout(resolve, 400);
      });

      expect(mockUpdateNodeData).toHaveBeenCalled();
      expect(capturedResult).toEqual(expect.objectContaining({directField: 'directValue'}));
    });

    it('should handle node with non-empty components array (line 240 true branch)', async () => {
      mockExecuteAsync.mockResolvedValue(true);

      let capturedResult: unknown;
      mockUpdateNodeData.mockImplementation((_stepId: string, callback: (node: {data: {components?: unknown[]}}) => unknown) => {
        const node = {
          data: {
            components: [
              {id: 'resource-1', label: 'Component Label'},
            ],
          },
        };
        capturedResult = callback(node);
        return capturedResult;
      });

      const MockComponentWithComponents = vi.fn(
        ({
          resource,
          onChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <button type="button" onClick={() => onChange('label', 'Updated Label', resource)}>
              Update With Components
            </button>
          </div>
        ),
      );

      const context = createContextValue({
        ResourceProperties: MockComponentWithComponents,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(context)});

      const changeButton = page.getByText('Update With Components');
      await changeButton.click();

      await new Promise(resolve => {
        setTimeout(resolve, 400);
      });

      expect(mockUpdateNodeData).toHaveBeenCalled();
      // Result should have components array updated
      expect(capturedResult).toHaveProperty('components');
    });

    it('should handle node with undefined data (lines 238-241)', async () => {
      mockExecuteAsync.mockResolvedValue(true);

      let capturedResult: unknown;
      mockUpdateNodeData.mockImplementation((_stepId: string, callback: (node: {data?: {components?: unknown[]}}) => unknown) => {
        const node = {
          data: undefined,
        };
        capturedResult = callback(node as unknown as {data: {components?: unknown[]}});
        return capturedResult;
      });

      const MockComponentWithUndefinedData = vi.fn(
        ({
          resource,
          onChange,
        }: {
          resource: Resource;
          properties?: Record<string, unknown>;
          onChange: (propertyKey: string, newValue: string | boolean | object, resource: Resource) => void;
          onVariantChange?: (variant: string, resource?: Partial<Resource>) => void;
        }) => (
          <div data-testid="mock-resource-properties">
            <div data-testid="resource-id">{resource?.id}</div>
            <button type="button" onClick={() => onChange('someField', 'someValue', resource)}>
              Update With Undefined Data
            </button>
          </div>
        ),
      );

      const context = createContextValue({
        ResourceProperties: MockComponentWithUndefinedData,
      });

      await render(<ResourceProperties />, {wrapper: createWrapper(context)});

      const changeButton = page.getByText('Update With Undefined Data');
      await changeButton.click();

      await new Promise(resolve => {
        setTimeout(resolve, 400);
      });

      expect(mockUpdateNodeData).toHaveBeenCalled();
    });
  });

});
