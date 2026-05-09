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

import {renderHook} from '@testing-library/react';
import {describe, it, expect} from 'vitest';
import actions from '../../data/actions.json';
import elements from '../../data/elements.json';
import steps from '../../data/steps.json';
import rawTemplates from '../../data/templates.json';
import widgets from '../../data/widgets.json';
import type {FlowTemplate} from '../../models/templates';
import useGetFlowsMeta from '../useGetFlowsMeta';

describe('useGetFlowsMeta', () => {
  describe('Return Structure', () => {
    it('should return an object with data, error, and isLoading', () => {
      const {result} = renderHook(() => useGetFlowsMeta());

      expect(result.current).toHaveProperty('data');
      expect(result.current).toHaveProperty('error');
      expect(result.current).toHaveProperty('isLoading');
    });

    it('should return error as null', () => {
      const {result} = renderHook(() => useGetFlowsMeta());

      expect(result.current.error).toBeNull();
    });

    it('should return isLoading as false', () => {
      const {result} = renderHook(() => useGetFlowsMeta());

      expect(result.current.isLoading).toBe(false);
    });
  });

  describe('Data Content', () => {
    it('should return all templates when no flowType filter is provided', () => {
      const {result} = renderHook(() => useGetFlowsMeta());

      expect(result.current.data.templates).toEqual(rawTemplates);
    });

    it('should return actions from JSON file', () => {
      const {result} = renderHook(() => useGetFlowsMeta());

      expect(result.current.data.actions).toEqual(actions);
    });

    it('should return elements from JSON file', () => {
      const {result} = renderHook(() => useGetFlowsMeta());

      expect(result.current.data.elements).toEqual(elements);
    });

    it('should return steps from JSON file', () => {
      const {result} = renderHook(() => useGetFlowsMeta());

      expect(result.current.data.steps).toEqual(steps);
    });

    it('should return widgets from JSON file', () => {
      const {result} = renderHook(() => useGetFlowsMeta());

      expect(result.current.data.widgets).toEqual(widgets);
    });

    it('should return an empty executors array', () => {
      const {result} = renderHook(() => useGetFlowsMeta());

      expect(result.current.data.executors).toEqual([]);
    });
  });

  describe('Filtering by flowType', () => {
    it('should filter templates by AUTHENTICATION flowType', () => {
      const {result} = renderHook(() => useGetFlowsMeta({flowType: 'AUTHENTICATION'}));

      const templates = result.current.data.templates;
      expect(templates.length).toBeGreaterThan(0);
      expect(templates.every((t: FlowTemplate) => t.flowType === 'AUTHENTICATION')).toBe(true);
    });

    it('should filter templates by REGISTRATION flowType', () => {
      const {result} = renderHook(() => useGetFlowsMeta({flowType: 'REGISTRATION'}));

      const templates = result.current.data.templates;
      expect(templates.length).toBeGreaterThan(0);
      expect(templates.every((t: FlowTemplate) => t.flowType === 'REGISTRATION')).toBe(true);
    });

    it('should not affect non-template resources when filtering by flowType', () => {
      const {result} = renderHook(() => useGetFlowsMeta({flowType: 'AUTHENTICATION'}));

      expect(result.current.data.actions).toEqual(actions);
      expect(result.current.data.elements).toEqual(elements);
      expect(result.current.data.steps).toEqual(steps);
      expect(result.current.data.widgets).toEqual(widgets);
    });
  });

  describe('Memoization', () => {
    it('should return memoized data on subsequent renders', () => {
      const {result, rerender} = renderHook(() => useGetFlowsMeta());

      const initialData = result.current.data;
      rerender();

      expect(result.current.data).toBe(initialData);
    });

    it('should return new templates reference when flowType changes', () => {
      const {result, rerender} = renderHook<
        ReturnType<typeof useGetFlowsMeta>,
        {flowType: 'AUTHENTICATION' | 'REGISTRATION'}
      >(({flowType}) => useGetFlowsMeta({flowType}), {
        initialProps: {flowType: 'AUTHENTICATION'},
      });

      const firstTemplates = result.current.data.templates;
      rerender({flowType: 'REGISTRATION'});

      expect(result.current.data.templates).not.toBe(firstTemplates);
    });
  });
});
