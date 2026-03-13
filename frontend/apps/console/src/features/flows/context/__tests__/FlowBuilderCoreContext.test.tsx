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

import {describe, it, expect} from 'vitest';
import {useContext} from 'react';
import {renderHook} from '@testing-library/react';
import FlowBuilderCoreContext from '../FlowBuilderCoreContext';
import {PreviewScreenType} from '../../models/custom-text-preference';
import {EdgeStyleTypes} from '../../models/steps';

describe('FlowBuilderCoreContext', () => {
  describe('default context values', () => {
    it('should have correct default values when accessed without provider', () => {
      const {result} = renderHook(() => useContext(FlowBuilderCoreContext));

      // Verify default boolean values
      expect(result.current.isResourcePanelOpen).toBe(true);
      expect(result.current.isResourcePropertiesPanelOpen).toBe(false);
      expect(result.current.isVersionHistoryPanelOpen).toBe(false);
      expect(result.current.i18nTextLoading).toBe(false);
      expect(result.current.isBrandingEnabled).toBe(false);
      expect(result.current.isFlowMetadataLoading).toBe(false);
      expect(result.current.isVerboseMode).toBe(false);
    });

    it('should have correct default string values', () => {
      const {result} = renderHook(() => useContext(FlowBuilderCoreContext));

      expect(result.current.lastInteractedStepId).toBe('');
      expect(result.current.language).toBe('');
    });

    it('should have correct default object values', () => {
      const {result} = renderHook(() => useContext(FlowBuilderCoreContext));

      expect(result.current.flowCompletionConfigs).toEqual({});
      expect(result.current.flowEdgeTypes).toEqual({});
      expect(result.current.flowNodeTypes).toEqual({});
      expect(result.current.selectedAttributes).toEqual({});
      expect(result.current.supportedLocales).toEqual({});
    });

    it('should have correct default undefined values', () => {
      const {result} = renderHook(() => useContext(FlowBuilderCoreContext));

      expect(result.current.i18nText).toBeUndefined();
      expect(result.current.metadata).toBeUndefined();
      expect(result.current.addResourceToFlow).toBeUndefined();
    });

    it('should have correct default null values', () => {
      const {result} = renderHook(() => useContext(FlowBuilderCoreContext));

      expect(result.current.resourcePropertiesPanelHeading).toBeNull();
    });

    it('should have correct default enum values', () => {
      const {result} = renderHook(() => useContext(FlowBuilderCoreContext));

      expect(result.current.primaryI18nScreen).toBe(PreviewScreenType.LOGIN);
      expect(result.current.edgeStyle).toBe(EdgeStyleTypes.SmoothStep);
    });

    it('should have default function stubs that do not throw', () => {
      const {result} = renderHook(() => useContext(FlowBuilderCoreContext));

      // Test that default functions don't throw when called
      expect(() => result.current.onResourceDropOnCanvas({} as never, 'node-1')).not.toThrow();
      expect(() => result.current.setFlowCompletionConfigs(() => ({}))).not.toThrow();
      expect(() => result.current.setFlowEdgeTypes(() => ({}))).not.toThrow();
      expect(() => result.current.setFlowNodeTypes(() => ({}))).not.toThrow();
      expect(() => result.current.setIsOpenResourcePropertiesPanel(true)).not.toThrow();
      expect(() => result.current.registerCloseValidationPanel(() => {})).not.toThrow();
      expect(() => result.current.setIsResourcePanelOpen(() => true)).not.toThrow();
      expect(() => result.current.setIsVersionHistoryPanelOpen(() => true)).not.toThrow();
      expect(() => result.current.setLanguage?.('en')).not.toThrow();
      expect(() => result.current.setLastInteractedResource({} as never)).not.toThrow();
      expect(() => result.current.setLastInteractedStepId('step-1')).not.toThrow();
      expect(() => result.current.setResourcePropertiesPanelHeading(() => null)).not.toThrow();
      expect(() => result.current.setSelectedAttributes(() => ({}))).not.toThrow();
      expect(() => result.current.setIsVerboseMode(() => true)).not.toThrow();
      expect(() => result.current.setEdgeStyle(() => EdgeStyleTypes.Bezier)).not.toThrow();
    });

    it('should have default isCustomI18nKey function that returns false', () => {
      const {result} = renderHook(() => useContext(FlowBuilderCoreContext));

      expect(result.current.isCustomI18nKey?.('any-key')).toBe(false);
      expect(result.current.isCustomI18nKey?.('any-key', true)).toBe(false);
    });

    it('should have default publishFlow function that resolves to false', async () => {
      const {result} = renderHook(() => useContext(FlowBuilderCoreContext));

      const publishResult = await result.current.publishFlow?.();
      expect(publishResult).toBe(false);
    });

    it('should have default ElementFactory component that returns null', () => {
      const {result} = renderHook(() => useContext(FlowBuilderCoreContext));

      const {ElementFactory} = result.current;
      expect(ElementFactory({stepId: 'step-1'})).toBeNull();
    });

    it('should have default ResourceProperties component that returns null', () => {
      const {result} = renderHook(() => useContext(FlowBuilderCoreContext));

      const {ResourceProperties} = result.current;
      expect(
        ResourceProperties({
          resource: {} as never,
          onChange: () => {},
        }),
      ).toBeNull();
    });

    it('should have lastInteractedResource as empty object', () => {
      const {result} = renderHook(() => useContext(FlowBuilderCoreContext));

      expect(result.current.lastInteractedResource).toEqual({});
    });
  });

  describe('context displayName', () => {
    it('should have displayName set to FlowBuilderCoreContext', () => {
      expect(FlowBuilderCoreContext.displayName).toBe('FlowBuilderCoreContext');
    });
  });
});
