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

import {describe, it, expect, vi} from 'vitest';
import {renderHook} from '@testing-library/react';
import type {ReactNode} from 'react';
import FlowBuilderCoreContext, {type FlowBuilderCoreContextProps} from '../../context/FlowBuilderCoreContext';
import useFlowBuilderCore from '../useFlowBuilderCore';
import {PreviewScreenType} from '../../models/custom-text-preference';
import {EdgeStyleTypes} from '../../models/steps';
import type {Base} from '../../models/base';

describe('useFlowBuilderCore', () => {
  const mockContextValue: FlowBuilderCoreContextProps = {
    lastInteractedResource: {id: 'resource-1'} as Base,
    lastInteractedStepId: 'step-1',
    ResourceProperties: () => null,
    resourcePropertiesPanelHeading: 'Properties',
    primaryI18nScreen: PreviewScreenType.LOGIN,
    isResourcePanelOpen: true,
    isResourcePropertiesPanelOpen: false,
    isVersionHistoryPanelOpen: false,
    ElementFactory: () => null,
    onResourceDropOnCanvas: vi.fn(),
    selectedAttributes: {},
    setLastInteractedResource: vi.fn(),
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
  };

  const createWrapper = (contextValue: FlowBuilderCoreContextProps) => {
    function Wrapper({children}: {children: ReactNode}) {
      return <FlowBuilderCoreContext.Provider value={contextValue}>{children}</FlowBuilderCoreContext.Provider>;
    }
    return Wrapper;
  };

  it('should return context values when used within provider', () => {
    const {result} = renderHook(() => useFlowBuilderCore(), {
      wrapper: createWrapper(mockContextValue),
    });

    expect(result.current.lastInteractedStepId).toBe('step-1');
    expect(result.current.isResourcePanelOpen).toBe(true);
    expect(result.current.isResourcePropertiesPanelOpen).toBe(false);
    expect(result.current.isVerboseMode).toBe(false);
    expect(result.current.edgeStyle).toBe(EdgeStyleTypes.SmoothStep);
  });

  it('should return default context values when used without explicit provider', () => {
    // When no provider is present, React returns the default context value
    // The hook checks for undefined context, but createContext provides defaults
    const {result} = renderHook(() => useFlowBuilderCore());

    // Default context values should be returned
    expect(result.current.isResourcePanelOpen).toBe(true);
    expect(result.current.isResourcePropertiesPanelOpen).toBe(false);
  });

  it('should return setLastInteractedResource function', () => {
    const {result} = renderHook(() => useFlowBuilderCore(), {
      wrapper: createWrapper(mockContextValue),
    });

    expect(typeof result.current.setLastInteractedResource).toBe('function');
  });

  it('should return setIsResourcePanelOpen function', () => {
    const {result} = renderHook(() => useFlowBuilderCore(), {
      wrapper: createWrapper(mockContextValue),
    });

    expect(typeof result.current.setIsResourcePanelOpen).toBe('function');
  });

  it('should return onResourceDropOnCanvas function', () => {
    const {result} = renderHook(() => useFlowBuilderCore(), {
      wrapper: createWrapper(mockContextValue),
    });

    expect(typeof result.current.onResourceDropOnCanvas).toBe('function');
  });

  it('should return ElementFactory component', () => {
    const {result} = renderHook(() => useFlowBuilderCore(), {
      wrapper: createWrapper(mockContextValue),
    });

    expect(result.current.ElementFactory).toBeDefined();
  });

  it('should return ResourceProperties component', () => {
    const {result} = renderHook(() => useFlowBuilderCore(), {
      wrapper: createWrapper(mockContextValue),
    });

    expect(result.current.ResourceProperties).toBeDefined();
  });

  it('should return flow node and edge types', () => {
    const contextWithTypes: FlowBuilderCoreContextProps = {
      ...mockContextValue,
      flowNodeTypes: {custom: () => null},
      flowEdgeTypes: {custom: () => null},
    };

    const {result} = renderHook(() => useFlowBuilderCore(), {
      wrapper: createWrapper(contextWithTypes),
    });

    expect(result.current.flowNodeTypes).toHaveProperty('custom');
    expect(result.current.flowEdgeTypes).toHaveProperty('custom');
  });

  it('should return metadata when provided', () => {
    const contextWithMetadata: FlowBuilderCoreContextProps = {
      ...mockContextValue,
      metadata: {
        flowType: 'LOGIN',
        supportedExecutors: [],
        connectorConfigs: {
          multiAttributeLoginEnabled: false,
          accountVerificationEnabled: false,
        },
        attributeProfile: 'default',
        attributeMetadata: [],
        executorConnections: [],
      },
    };

    const {result} = renderHook(() => useFlowBuilderCore(), {
      wrapper: createWrapper(contextWithMetadata),
    });

    expect(result.current.metadata).toBeDefined();
    expect(result.current.metadata?.flowType).toEqual('LOGIN');
  });

  it('should return i18n text when provided', () => {
    const contextWithI18n: FlowBuilderCoreContextProps = {
      ...mockContextValue,
      i18nText: {
        [PreviewScreenType.LOGIN]: {
          'login.title': 'Welcome',
        },
      },
      language: 'en',
    };

    const {result} = renderHook(() => useFlowBuilderCore(), {
      wrapper: createWrapper(contextWithI18n),
    });

    expect(result.current.i18nText).toBeDefined();
    expect(result.current.language).toBe('en');
  });

  it('should return publishFlow function when provided', () => {
    const mockPublishFlow = vi.fn().mockResolvedValue(true);
    const contextWithPublish: FlowBuilderCoreContextProps = {
      ...mockContextValue,
      publishFlow: mockPublishFlow,
    };

    const {result} = renderHook(() => useFlowBuilderCore(), {
      wrapper: createWrapper(contextWithPublish),
    });

    expect(result.current.publishFlow).toBe(mockPublishFlow);
  });

  it('should throw an error when used outside of FlowBuilderCoreProvider with undefined context', () => {
    // Create a wrapper that provides undefined as the context value
    function UndefinedContextWrapper({children}: {children: ReactNode}) {
      return (
        <FlowBuilderCoreContext.Provider value={undefined as unknown as FlowBuilderCoreContextProps}>
          {children}
        </FlowBuilderCoreContext.Provider>
      );
    }

    expect(() => {
      renderHook(() => useFlowBuilderCore(), {
        wrapper: UndefinedContextWrapper,
      });
    }).toThrow('useFlowBuilderCore must be used within a FlowBuilderCoreProvider');
  });
});
