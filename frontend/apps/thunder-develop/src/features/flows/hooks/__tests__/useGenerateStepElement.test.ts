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

/* eslint-disable @typescript-eslint/no-explicit-any, @typescript-eslint/no-unsafe-member-access */

import {describe, it, expect, vi} from 'vitest';
import {renderHook} from '@thunder/test-utils/browser';
import useGenerateStepElement from '../useGenerateStepElement';
import type {Element} from '../../models/elements';

// Mock generateResourceId
vi.mock('../../utils/generateResourceId', () => ({
  default: (prefix: string) => `${prefix}-generated-id`,
}));

describe('useGenerateStepElement', () => {
  describe('Hook Interface', () => {
    it('should return generateStepElement function', async () => {
      const {result} = await renderHook(() => useGenerateStepElement());

      expect(typeof result.current.generateStepElement).toBe('function');
    });
  });

  describe('generateStepElement', () => {
    it('should generate element with unique ID based on category', async () => {
      const {result} = await renderHook(() => useGenerateStepElement());

      const element: Element = {
        id: 'original-id',
        type: 'ACTION',
        category: 'ACTION',
        display: {label: 'Test Button', image: '', showOnResourcePanel: true},
      } as Element;

      const generatedElement = result.current.generateStepElement(element);

      expect(generatedElement.id).toBe('action-generated-id');
    });

    it('should preserve original element properties', async () => {
      const {result} = await renderHook(() => useGenerateStepElement());

      const element: Element = {
        id: 'original-id',
        type: 'TEXT_INPUT',
        category: 'INPUT',
        display: {label: 'Email Input', image: '', showOnResourcePanel: true},
        config: {placeholder: 'Enter email'},
      } as unknown as Element;

      const generatedElement = result.current.generateStepElement(element);

      expect(generatedElement.type).toBe('TEXT_INPUT');
      expect(generatedElement.category).toBe('INPUT');
      expect(generatedElement.display?.label).toBe('Email Input');
      expect((generatedElement as any).config?.placeholder).toBe('Enter email');
    });

    it('should convert category to lowercase for ID generation', async () => {
      const {result} = await renderHook(() => useGenerateStepElement());

      const element: Element = {
        id: 'original-id',
        type: 'BUTTON',
        category: 'ACTION',
        display: {label: 'Button', image: '', showOnResourcePanel: true},
      } as Element;

      const generatedElement = result.current.generateStepElement(element);

      expect(generatedElement.id).toBe('action-generated-id');
    });

    it('should apply default variant when variants exist', async () => {
      const {result} = await renderHook(() => useGenerateStepElement());

      const element: Element = {
        id: 'original-id',
        type: 'BUTTON',
        category: 'ACTION',
        display: {label: 'Button', image: '', showOnResourcePanel: true, defaultVariant: 'primary'},
        variants: [
          {variant: 'primary', style: 'contained', color: 'primary'},
          {variant: 'secondary', style: 'outlined', color: 'secondary'},
        ] as unknown as Element[],
      } as unknown as Element;

      const generatedElement = result.current.generateStepElement(element);

      expect((generatedElement as any).variant).toBe('primary');
      expect((generatedElement as any).style).toBe('contained');
      expect((generatedElement as any).color).toBe('primary');
    });

    it('should use first variant when no defaultVariant is specified', async () => {
      const {result} = await renderHook(() => useGenerateStepElement());

      const element: Element = {
        id: 'original-id',
        type: 'BUTTON',
        category: 'ACTION',
        display: {label: 'Button', image: '', showOnResourcePanel: true},
        variants: [
          {variant: 'outlined', style: 'outlined'},
          {variant: 'contained', style: 'contained'},
        ] as unknown as Element[],
      } as unknown as Element;

      const generatedElement = result.current.generateStepElement(element);

      expect((generatedElement as any).variant).toBe('outlined');
      expect((generatedElement as any).style).toBe('outlined');
    });

    it('should not modify element without variants', async () => {
      const {result} = await renderHook(() => useGenerateStepElement());

      const element: Element = {
        id: 'original-id',
        type: 'DIVIDER',
        category: 'LAYOUT',
        display: {label: 'Divider', image: '', showOnResourcePanel: true},
      } as Element;

      const generatedElement = result.current.generateStepElement(element);

      expect(generatedElement.type).toBe('DIVIDER');
      expect((generatedElement as any).variants).toBeUndefined();
    });

    it('should handle empty variants array', async () => {
      const {result} = await renderHook(() => useGenerateStepElement());

      const element: Element = {
        id: 'original-id',
        type: 'TEXT',
        category: 'TYPOGRAPHY',
        display: {label: 'Text', image: '', showOnResourcePanel: true},
        variants: [] as unknown as Element[],
      } as unknown as Element;

      const generatedElement = result.current.generateStepElement(element);

      expect(generatedElement.id).toBe('typography-generated-id');
      expect(generatedElement.type).toBe('TEXT');
    });

    it('should handle element with undefined variants property', async () => {
      const {result} = await renderHook(() => useGenerateStepElement());

      const element: Element = {
        id: 'original-id',
        type: 'IMAGE',
        category: 'MEDIA',
        display: {label: 'Image', image: '', showOnResourcePanel: true},
      } as Element;

      const generatedElement = result.current.generateStepElement(element);

      expect(generatedElement.id).toBe('media-generated-id');
    });
  });

  describe('Function Behavior', () => {
    it('should work correctly after rerender', async () => {
      const {result, rerender} = await renderHook(() => useGenerateStepElement());

      await rerender();

      const element: Element = {
        id: 'original-id',
        type: 'BUTTON',
        category: 'ACTION',
        display: {label: 'Button', image: '', showOnResourcePanel: true},
      } as Element;

      const generatedElement = result.current.generateStepElement(element);

      expect(generatedElement.id).toBe('action-generated-id');
    });
  });
});
