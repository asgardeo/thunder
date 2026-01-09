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

import {describe, expect, it} from 'vitest';
import resolveComponentMetadata from '../resolveComponentMetadata';
import type {Resources} from '../../models/resources';
import type {Element} from '../../models/elements';

const createMockElement = (overrides: Partial<Element> = {}): Element =>
  ({
    id: 'element-1',
    type: 'TEXT_INPUT',
    category: 'FIELD',
    version: '1.0.0',
    deprecated: false,
    deletable: true,
    resourceType: 'ELEMENT',
    display: {
      label: 'Text Input',
      image: '',
      showOnResourcePanel: true,
    },
    config: {
      field: {name: 'textField', type: 'TEXT_INPUT' as unknown as typeof import('../../models/elements').ElementTypes},
      styles: {},
    },
    ...overrides,
  }) as Element;

const createMockResources = (overrides: Partial<Resources> = {}): Resources => ({
  elements: [],
  steps: [],
  widgets: [],
  templates: [],
  executors: [],
  ...overrides,
});

describe('resolveComponentMetadata', () => {
  describe('Basic Metadata Resolution', () => {
    it('should return components with merged metadata from resources', () => {
      const components: Element[] = [createMockElement({id: 'comp-1', type: 'TEXT_INPUT'})];

      const resources = createMockResources({
        elements: [
          createMockElement({
            type: 'TEXT_INPUT',
            display: {
              label: 'Text Field',
              image: '/images/text-input.svg',
              showOnResourcePanel: true,
              description: 'A text input field',
            },
          }),
        ],
      });

      const result = resolveComponentMetadata(resources, components);

      expect(result).toHaveLength(1);
      // lodash merge: component values override metadata values
      // But metadata's additional properties (description) are added
      expect(result[0].display.description).toBe('A text input field');
      expect(result[0].type).toBe('TEXT_INPUT');
    });

    it('should preserve original component data while merging metadata', () => {
      const components: Element[] = [
        createMockElement({
          id: 'unique-comp-id',
          type: 'TEXT_INPUT',
          config: {
            field: {name: 'customField', type: 'TEXT_INPUT'} as unknown as Element['config']['field'],
            styles: {width: '100%'},
            label: 'Custom Label',
          },
        }),
      ];

      const resources = createMockResources({
        elements: [
          createMockElement({
            type: 'TEXT_INPUT',
            display: {
              label: 'Generic Text Input',
              image: '/images/text.svg',
              showOnResourcePanel: true,
            },
          }),
        ],
      });

      const result = resolveComponentMetadata(resources, components);

      expect(result[0].id).toBe('unique-comp-id');
      expect(result[0].config.label).toBe('Custom Label');
      expect(result[0].config.styles).toEqual({width: '100%'});
    });

    it('should handle components without matching metadata', () => {
      const components: Element[] = [createMockElement({id: 'comp-1', type: 'CUSTOM_TYPE'})];

      const resources = createMockResources({
        elements: [createMockElement({type: 'TEXT_INPUT'})],
      });

      const result = resolveComponentMetadata(resources, components);

      expect(result).toHaveLength(1);
      expect(result[0].type).toBe('CUSTOM_TYPE');
    });

    it('should handle multiple components and add metadata properties', () => {
      const components: Element[] = [
        createMockElement({id: 'comp-1', type: 'TEXT_INPUT'}),
        createMockElement({id: 'comp-2', type: 'PASSWORD_INPUT'}),
        createMockElement({id: 'comp-3', type: 'ACTION'}),
      ];

      const resources = createMockResources({
        elements: [
          createMockElement({
            type: 'TEXT_INPUT',
            display: {label: 'Text', image: '', showOnResourcePanel: true, description: 'Text field'},
          }),
          createMockElement({
            type: 'PASSWORD_INPUT',
            display: {label: 'Password', image: '', showOnResourcePanel: true, description: 'Password field'},
          }),
          createMockElement({
            type: 'ACTION',
            display: {label: 'Button', image: '', showOnResourcePanel: true, description: 'Action button'},
          }),
        ],
      });

      const result = resolveComponentMetadata(resources, components);

      expect(result).toHaveLength(3);
      // Metadata's additional properties are added
      expect(result[0].display.description).toBe('Text field');
      expect(result[1].display.description).toBe('Password field');
      expect(result[2].display.description).toBe('Action button');
    });
  });

  describe('Variant Handling', () => {
    it('should match component with variant against element with variants array', () => {
      const components: Element[] = [
        createMockElement({
          id: 'comp-1',
          type: 'ACTION',
          variant: 'PRIMARY',
        }),
      ];

      const resources = createMockResources({
        elements: [
          createMockElement({
            type: 'ACTION',
            variants: [
              createMockElement({
                variant: 'PRIMARY',
                display: {label: 'Primary Button', image: '', showOnResourcePanel: true, description: 'Primary variant'},
              }),
              createMockElement({
                variant: 'SECONDARY',
                display: {label: 'Secondary Button', image: '', showOnResourcePanel: true},
              }),
            ],
            display: {
              label: 'Action Button',
              image: '/images/button.svg',
              showOnResourcePanel: true,
              description: 'Generic action',
            },
          }),
        ],
      });

      const result = resolveComponentMetadata(resources, components);

      // Component values are preserved, metadata description is added
      expect(result[0].variant).toBe('PRIMARY');
      expect(result[0].display.description).toBe('Generic action');
    });

    it('should match component with variant against element with high-level variant', () => {
      const components: Element[] = [
        createMockElement({
          id: 'comp-1',
          type: 'ACTION',
          variant: 'SOCIAL',
        }),
      ];

      const resources = createMockResources({
        elements: [
          createMockElement({
            type: 'ACTION',
            variant: 'SOCIAL',
            display: {
              label: 'Social Button',
              image: '/images/social.svg',
              showOnResourcePanel: true,
              description: 'Social login button',
            },
          }),
          createMockElement({
            type: 'ACTION',
            variant: 'PRIMARY',
            display: {
              label: 'Primary Button',
              image: '/images/primary.svg',
              showOnResourcePanel: true,
            },
          }),
        ],
      });

      const result = resolveComponentMetadata(resources, components);

      // Metadata description is added
      expect(result[0].display.description).toBe('Social login button');
    });

    it('should not merge when variant does not match', () => {
      const components: Element[] = [
        createMockElement({
          id: 'comp-1',
          type: 'ACTION',
          variant: 'TEXT',
        }),
      ];

      const resources = createMockResources({
        elements: [
          createMockElement({
            type: 'ACTION',
            variant: 'PRIMARY',
            display: {
              label: 'Primary Button',
              image: '/images/primary.svg',
              showOnResourcePanel: true,
              description: 'Should not appear',
            },
          }),
        ],
      });

      const result = resolveComponentMetadata(resources, components);

      // No match, so original description (undefined) is preserved
      expect(result[0].display.description).toBeUndefined();
    });

    it('should merge when component has no variant and element has no variant', () => {
      const components: Element[] = [
        createMockElement({
          id: 'comp-1',
          type: 'DIVIDER',
        }),
      ];

      const resources = createMockResources({
        elements: [
          createMockElement({
            type: 'DIVIDER',
            display: {
              label: 'Divider',
              image: '/images/divider.svg',
              showOnResourcePanel: true,
              description: 'Divider element',
            },
          }),
        ],
      });

      const result = resolveComponentMetadata(resources, components);

      expect(result[0].display.description).toBe('Divider element');
    });
  });

  describe('Nested Components', () => {
    it('should recursively resolve metadata for nested components', () => {
      const components: Element[] = [
        createMockElement({
          id: 'form-1',
          type: 'BLOCK',
          components: [
            createMockElement({id: 'input-1', type: 'TEXT_INPUT'}),
            createMockElement({id: 'input-2', type: 'PASSWORD_INPUT'}),
          ],
        }),
      ];

      const resources = createMockResources({
        elements: [
          createMockElement({
            type: 'BLOCK',
            display: {label: 'Form Block', image: '', showOnResourcePanel: true, description: 'Block element'},
          }),
          createMockElement({
            type: 'TEXT_INPUT',
            display: {label: 'Text Field', image: '/images/text.svg', showOnResourcePanel: true, description: 'Text input'},
          }),
          createMockElement({
            type: 'PASSWORD_INPUT',
            display: {label: 'Password Field', image: '/images/password.svg', showOnResourcePanel: true, description: 'Password input'},
          }),
        ],
      });

      const result = resolveComponentMetadata(resources, components);

      // Metadata descriptions are added
      expect(result[0].display.description).toBe('Block element');
      expect(result[0].components).toHaveLength(2);
      expect(result[0].components![0].display.description).toBe('Text input');
      expect(result[0].components![1].display.description).toBe('Password input');
    });

    it('should handle deeply nested components', () => {
      const components: Element[] = [
        createMockElement({
          id: 'outer',
          type: 'BLOCK',
          components: [
            createMockElement({
              id: 'inner',
              type: 'BLOCK',
              components: [createMockElement({id: 'deepest', type: 'TEXT_INPUT'})],
            }),
          ],
        }),
      ];

      const resources = createMockResources({
        elements: [
          createMockElement({
            type: 'BLOCK',
            display: {label: 'Block', image: '', showOnResourcePanel: true, description: 'A block'},
          }),
          createMockElement({
            type: 'TEXT_INPUT',
            display: {label: 'Deep Input', image: '', showOnResourcePanel: true, description: 'Deep input element'},
          }),
        ],
      });

      const result = resolveComponentMetadata(resources, components);

      expect(result[0].components![0].components![0].display.description).toBe('Deep input element');
    });
  });

  describe('Resource Type Filtering', () => {
    it('should filter out non-ELEMENT resources', () => {
      const components: Element[] = [
        createMockElement({id: 'comp-1', type: 'TEXT_INPUT', resourceType: 'ELEMENT'}),
        createMockElement({id: 'widget-1', type: 'SOME_WIDGET', resourceType: 'WIDGET'}),
        createMockElement({id: 'step-1', type: 'SOME_STEP', resourceType: 'STEP'}),
      ];

      const resources = createMockResources({
        elements: [
          createMockElement({type: 'TEXT_INPUT', display: {label: 'Text', image: '', showOnResourcePanel: true}}),
        ],
      });

      const result = resolveComponentMetadata(resources, components);

      expect(result).toHaveLength(1);
      expect(result[0].resourceType).toBe('ELEMENT');
    });

    it('should include components without resourceType (template components)', () => {
      const components: Element[] = [
        createMockElement({id: 'comp-1', type: 'TEXT_INPUT', resourceType: undefined as unknown as string}),
      ];

      const resources = createMockResources({
        elements: [
          createMockElement({type: 'TEXT_INPUT', display: {label: 'Text', image: '', showOnResourcePanel: true}}),
        ],
      });

      const result = resolveComponentMetadata(resources, components);

      expect(result).toHaveLength(1);
    });
  });

  describe('Edge Cases', () => {
    it('should return empty array for undefined components', () => {
      const resources = createMockResources();
      const result = resolveComponentMetadata(resources, undefined);

      expect(result).toEqual([]);
    });

    it('should return empty array for empty components array', () => {
      const resources = createMockResources();
      const result = resolveComponentMetadata(resources, []);

      expect(result).toEqual([]);
    });

    it('should handle empty resources', () => {
      const components: Element[] = [createMockElement({id: 'comp-1', type: 'TEXT_INPUT'})];
      const resources = createMockResources();

      const result = resolveComponentMetadata(resources, components);

      expect(result).toHaveLength(1);
      expect(result[0].type).toBe('TEXT_INPUT');
    });

    it('should handle null elements in resources', () => {
      const components: Element[] = [createMockElement({id: 'comp-1', type: 'TEXT_INPUT'})];

      const resources = createMockResources({
        elements: undefined as unknown as Element[],
      });

      const result = resolveComponentMetadata(resources, components);

      expect(result).toHaveLength(1);
    });

    it('should handle components with empty nested components array', () => {
      const components: Element[] = [
        createMockElement({
          id: 'form-1',
          type: 'BLOCK',
          components: [],
        }),
      ];

      const resources = createMockResources({
        elements: [createMockElement({type: 'BLOCK', display: {label: 'Block', image: '', showOnResourcePanel: true}})],
      });

      const result = resolveComponentMetadata(resources, components);

      expect(result[0].components).toEqual([]);
    });
  });
});
