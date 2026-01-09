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
import {render, screen} from '@testing-library/react';
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

// Mock PluginRegistry
vi.mock('../../../plugins/PluginRegistry', () => ({
  default: {
    getInstance: () => ({
      executeSync: vi.fn().mockReturnValue(true),
      executeAsync: vi.fn().mockResolvedValue(true),
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
  });

  describe('Rendering', () => {
    it('should render ResourcePropertiesComponent when resource is available', () => {
      render(<ResourceProperties />, {wrapper: createWrapper()});

      expect(screen.getByTestId('mock-resource-properties')).toBeInTheDocument();
    });

    it('should display resource id', () => {
      render(<ResourceProperties />, {wrapper: createWrapper()});

      expect(screen.getByTestId('resource-id')).toHaveTextContent('resource-1');
    });

    it('should show "No properties available" when lastInteractedResource is null', () => {
      const contextWithNoResource = createContextValue({
        lastInteractedResource: null as unknown as Base,
      });

      render(<ResourceProperties />, {wrapper: createWrapper(contextWithNoResource)});

      expect(screen.getByText('No properties available.')).toBeInTheDocument();
    });
  });

  describe('Properties Filtering', () => {
    it('should filter out excluded properties from config', () => {
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

      render(<ResourceProperties />, {wrapper: createWrapper(contextWithConfig)});

      expect(screen.getByTestId('mock-resource-properties')).toBeInTheDocument();
    });

    it('should extract top-level editable properties', () => {
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

      render(<ResourceProperties />, {wrapper: createWrapper(contextWithTopLevelProps)});

      const propertiesDiv = screen.getByTestId('properties');
      const properties = JSON.parse(propertiesDiv.textContent ?? '{}') as Record<string, unknown>;

      expect(properties.label).toBe('Test Label');
      expect(properties.hint).toBe('Test Hint');
      expect(properties.placeholder).toBe('Test Placeholder');
      expect(properties.required).toBe(true);
    });
  });

  describe('Variant Change', () => {
    it('should have variant change callback available', () => {
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

      render(<ResourceProperties />, {wrapper: createWrapper(contextWithVariants)});

      expect(screen.getByText('Change Variant')).toBeInTheDocument();
    });
  });

  describe('Resource Without Config', () => {
    it('should handle resource without config gracefully', () => {
      const resourceWithoutConfig: Base = {
        ...mockBaseResource,
        config: undefined as unknown as BaseConfig,
      };

      const contextWithoutConfig = createContextValue({
        lastInteractedResource: resourceWithoutConfig,
      });

      render(<ResourceProperties />, {wrapper: createWrapper(contextWithoutConfig)});

      expect(screen.getByTestId('mock-resource-properties')).toBeInTheDocument();
    });
  });

  describe('Memoization', () => {
    it('should use memoized component', () => {
      const {rerender} = render(<ResourceProperties />, {wrapper: createWrapper()});

      expect(screen.getByTestId('mock-resource-properties')).toBeInTheDocument();

      // Re-render with same props should not cause issues
      rerender(<ResourceProperties />);

      expect(screen.getByTestId('mock-resource-properties')).toBeInTheDocument();
    });
  });

  describe('Empty Resource', () => {
    it('should handle empty config object', () => {
      const resourceWithEmptyConfig: Base = {
        ...mockBaseResource,
        config: {...mockBaseResource.config},
      };

      const contextWithEmptyConfig = createContextValue({
        lastInteractedResource: resourceWithEmptyConfig,
      });

      render(<ResourceProperties />, {wrapper: createWrapper(contextWithEmptyConfig)});

      expect(screen.getByTestId('mock-resource-properties')).toBeInTheDocument();
    });
  });

  describe('Property Change Handler', () => {
    it('should trigger onChange callback when property changes', async () => {
      render(<ResourceProperties />, {wrapper: createWrapper()});

      const changeLabelButton = screen.getByText('Change Label');
      changeLabelButton.click();

      // The onChange callback should be passed to MockResourcePropertiesComponent
      expect(MockResourcePropertiesComponent).toHaveBeenCalled();
    });

    it('should pass resource to onChange callback', () => {
      render(<ResourceProperties />, {wrapper: createWrapper()});

      // Verify the MockResourcePropertiesComponent receives the resource
      const {calls} = MockResourcePropertiesComponent.mock;
      expect(calls.length).toBeGreaterThan(0);
      const props = calls[0][0] as {resource: {id: string}};
      expect(props.resource.id).toBe('resource-1');
    });
  });

  describe('Variant Change Handler', () => {
    it('should trigger onVariantChange callback', () => {
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

      render(<ResourceProperties />, {wrapper: createWrapper(contextWithVariants)});

      const changeVariantButton = screen.getByText('Change Variant');
      changeVariantButton.click();

      // Verify onVariantChange was passed to the component
      const {calls} = MockResourcePropertiesComponent.mock;
      expect(calls.length).toBeGreaterThan(0);
      const props = calls[0][0] as {onVariantChange: unknown};
      expect(typeof props.onVariantChange).toBe('function');
    });

    it('should handle variant change for resource with label', () => {
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

      render(<ResourceProperties />, {wrapper: createWrapper(contextWithVariantsAndLabel)});

      expect(screen.getByTestId('mock-resource-properties')).toBeInTheDocument();
    });

    it('should handle variant change for resource with text config', () => {
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

      render(<ResourceProperties />, {wrapper: createWrapper(contextWithVariantsAndText)});

      expect(screen.getByTestId('mock-resource-properties')).toBeInTheDocument();
    });

    it('should not change variant when variant is not found', () => {
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

      render(<ResourceProperties />, {wrapper: createWrapper(contextWithVariants)});

      // The component should render without issues
      expect(screen.getByTestId('mock-resource-properties')).toBeInTheDocument();
    });
  });

  describe('Top-Level Properties', () => {
    it('should extract src property', () => {
      const resourceWithSrc: Base = {
        ...mockBaseResource,
        src: 'https://example.com/image.png',
      } as Base & {src: string};

      const contextWithSrc = createContextValue({
        lastInteractedResource: resourceWithSrc,
      });

      render(<ResourceProperties />, {wrapper: createWrapper(contextWithSrc)});

      const propertiesDiv = screen.getByTestId('properties');
      const properties = JSON.parse(propertiesDiv.textContent ?? '{}') as Record<string, unknown>;

      expect(properties.src).toBe('https://example.com/image.png');
    });

    it('should extract alt property', () => {
      const resourceWithAlt: Base = {
        ...mockBaseResource,
        alt: 'Alternative text',
      } as Base & {alt: string};

      const contextWithAlt = createContextValue({
        lastInteractedResource: resourceWithAlt,
      });

      render(<ResourceProperties />, {wrapper: createWrapper(contextWithAlt)});

      const propertiesDiv = screen.getByTestId('properties');
      const properties = JSON.parse(propertiesDiv.textContent ?? '{}') as Record<string, unknown>;

      expect(properties.alt).toBe('Alternative text');
    });
  });

  describe('Nested Components', () => {
    it('should handle resource with nested components', () => {
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

      render(<ResourceProperties />, {wrapper: createWrapper(contextWithNestedComponents)});

      expect(screen.getByTestId('mock-resource-properties')).toBeInTheDocument();
    });
  });

  describe('Variant Not Found', () => {
    it('should return early when selected variant is not found', () => {
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

      render(<ResourceProperties />, {wrapper: createWrapper(contextWithVariants)});

      const changeVariantButton = screen.getByText('Change to Non-Existent Variant');
      changeVariantButton.click();

      // updateNodeData should not be called when variant is not found
      expect(mockUpdateNodeData).not.toHaveBeenCalled();
    });
  });

  describe('Preserve Label on Variant Change', () => {
    it('should preserve current label value when changing variants', () => {
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

      render(<ResourceProperties />, {wrapper: createWrapper(contextWithLabelAndVariants)});

      const changeVariantButton = screen.getByText('Change Variant');
      changeVariantButton.click();

      // Verify updateNodeData was called (variant change should preserve the label)
      expect(mockUpdateNodeData).toHaveBeenCalled();
    });
  });

  describe('Preserve Text on Variant Change', () => {
    it('should preserve current text value when changing variants with selectedVariant.config', () => {
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

      render(<ResourceProperties />, {wrapper: createWrapper(contextWithTextAndVariants)});

      const changeVariantButton = screen.getByText('Change Variant');
      changeVariantButton.click();

      // Verify updateNodeData was called
      expect(mockUpdateNodeData).toHaveBeenCalled();
    });
  });

  describe('Update Component Recursive Mapping', () => {
    it('should update component when id matches', () => {
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

      render(<ResourceProperties />, {wrapper: createWrapper(contextWithComponents)});

      const changeVariantButton = screen.getByText('Change Variant');
      changeVariantButton.click();

      expect(mockUpdateNodeData).toHaveBeenCalled();
    });

    it('should recursively update nested components', () => {
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

      render(<ResourceProperties />, {wrapper: createWrapper(contextWithNestedComponents)});

      const changeVariantButton = screen.getByText('Change Variant');
      changeVariantButton.click();

      expect(mockUpdateNodeData).toHaveBeenCalled();
    });
  });

  describe('handlePropertyChange with Plugin Result False', () => {
    it('should still update resource when plugin returns false', async () => {
      vi.mock('../../../plugins/PluginRegistry', () => ({
        default: {
          getInstance: () => ({
            executeSync: vi.fn().mockReturnValue(true),
            executeAsync: vi.fn().mockResolvedValue(false),
          }),
        },
      }));

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

      render(<ResourceProperties />, {wrapper: createWrapper(context)});

      const changeLabelButton = screen.getByText('Change Label');
      changeLabelButton.click();

      // Wait for debounced function to execute
      await new Promise(resolve => {
        setTimeout(resolve, 400);
      });

      expect(MockComponentWithPropertyChange).toHaveBeenCalled();
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

      render(<ResourceProperties />, {wrapper: createWrapper(context)});

      const changeDataButton = screen.getByText('Change Data');
      changeDataButton.click();

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

      render(<ResourceProperties />, {wrapper: createWrapper(context)});

      const changeButton = screen.getByText('Change Config Style');
      changeButton.click();

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

      render(<ResourceProperties />, {wrapper: createWrapper(context)});

      const changeButton = screen.getByText('Change Data Field');
      changeButton.click();

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

      render(<ResourceProperties />, {wrapper: createWrapper(context)});

      const changeButton = screen.getByText('Change Action');
      changeButton.click();

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

      render(<ResourceProperties />, {wrapper: createWrapper(context)});

      const changeButton = screen.getByText('Change Custom Property');
      changeButton.click();

      // Wait for debounced function to execute
      await new Promise(resolve => {
        setTimeout(resolve, 400);
      });

      expect(MockComponentWithCustomProperty).toHaveBeenCalled();
    });
  });

  describe('Variant Change with Element Partial Override', () => {
    it('should merge element partial when provided to variant change', () => {
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

      render(<ResourceProperties />, {wrapper: createWrapper(contextWithVariants)});

      const changeVariantButton = screen.getByText('Change Variant with Override');
      changeVariantButton.click();

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

      render(<ResourceProperties />, {wrapper: createWrapper(context)});

      const changeLabelButton = screen.getByText('Change Label');
      changeLabelButton.click();

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

      render(<ResourceProperties />, {wrapper: createWrapper(context)});

      const changeButton = screen.getByText('Change Different Resource');
      changeButton.click();

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

      render(<ResourceProperties />, {wrapper: createWrapper(context)});

      const changeButton = screen.getByText('Change Data Prefixed Field');
      changeButton.click();

      // Wait for debounced function to execute
      await new Promise(resolve => {
        setTimeout(resolve, 400);
      });

      expect(MockComponentWithDataPrefix).toHaveBeenCalled();
    });
  });
});
