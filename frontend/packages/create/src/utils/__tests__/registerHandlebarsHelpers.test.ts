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

import Handlebars from 'handlebars';
import {describe, it, expect, beforeEach} from 'vitest';
import registerHandlebarsHelpers from '../registerHandlebarsHelpers';

describe('registerHandlebarsHelpers', () => {
  beforeEach(() => {
    // Clear any existing helpers and re-register
    Handlebars.unregisterHelper('pascalCase');
    Handlebars.unregisterHelper('camelCase');
    Handlebars.unregisterHelper('kebabCase');
    Handlebars.unregisterHelper('constantCase');
    Handlebars.unregisterHelper('snakeCase');
    Handlebars.unregisterHelper('if_eq');
    Handlebars.unregisterHelper('if_includes');
    registerHandlebarsHelpers();
  });

  describe('pascalCase helper', () => {
    it('should convert kebab-case to PascalCase', () => {
      const template = Handlebars.compile('{{pascalCase name}}');
      expect(template({name: 'my-feature-name'})).toBe('MyFeatureName');
    });

    it('should convert snake_case to PascalCase', () => {
      const template = Handlebars.compile('{{pascalCase name}}');
      expect(template({name: 'my_feature_name'})).toBe('MyFeatureName');
    });

    it('should convert camelCase to PascalCase', () => {
      const template = Handlebars.compile('{{pascalCase name}}');
      expect(template({name: 'myFeatureName'})).toBe('MyFeatureName');
    });

    it('should handle already PascalCase strings', () => {
      const template = Handlebars.compile('{{pascalCase name}}');
      expect(template({name: 'MyFeatureName'})).toBe('MyFeatureName');
    });

    it('should handle single word', () => {
      const template = Handlebars.compile('{{pascalCase name}}');
      expect(template({name: 'feature'})).toBe('Feature');
    });

    it('should handle mixed separators', () => {
      const template = Handlebars.compile('{{pascalCase name}}');
      expect(template({name: 'my_feature-name'})).toBe('MyFeatureName');
    });
  });

  describe('camelCase helper', () => {
    it('should convert kebab-case to camelCase', () => {
      const template = Handlebars.compile('{{camelCase name}}');
      expect(template({name: 'my-feature-name'})).toBe('myFeatureName');
    });

    it('should convert snake_case to camelCase', () => {
      const template = Handlebars.compile('{{camelCase name}}');
      expect(template({name: 'my_feature_name'})).toBe('myFeatureName');
    });

    it('should convert PascalCase to camelCase', () => {
      const template = Handlebars.compile('{{camelCase name}}');
      expect(template({name: 'MyFeatureName'})).toBe('myFeatureName');
    });

    it('should handle already camelCase strings', () => {
      const template = Handlebars.compile('{{camelCase name}}');
      expect(template({name: 'myFeatureName'})).toBe('myFeatureName');
    });

    it('should handle single word', () => {
      const template = Handlebars.compile('{{camelCase name}}');
      expect(template({name: 'feature'})).toBe('feature');
    });
  });

  describe('kebabCase helper', () => {
    it('should convert PascalCase to kebab-case', () => {
      const template = Handlebars.compile('{{kebabCase name}}');
      expect(template({name: 'MyFeatureName'})).toBe('my-feature-name');
    });

    it('should convert camelCase to kebab-case', () => {
      const template = Handlebars.compile('{{kebabCase name}}');
      expect(template({name: 'myFeatureName'})).toBe('my-feature-name');
    });

    it('should convert snake_case to kebab-case', () => {
      const template = Handlebars.compile('{{kebabCase name}}');
      expect(template({name: 'my_feature_name'})).toBe('my-feature-name');
    });

    it('should handle already kebab-case strings', () => {
      const template = Handlebars.compile('{{kebabCase name}}');
      expect(template({name: 'my-feature-name'})).toBe('my-feature-name');
    });

    it('should handle spaces', () => {
      const template = Handlebars.compile('{{kebabCase name}}');
      expect(template({name: 'my feature name'})).toBe('my-feature-name');
    });

    it('should handle single word', () => {
      const template = Handlebars.compile('{{kebabCase name}}');
      expect(template({name: 'feature'})).toBe('feature');
    });
  });

  describe('constantCase helper', () => {
    it('should convert camelCase to CONSTANT_CASE', () => {
      const template = Handlebars.compile('{{constantCase name}}');
      expect(template({name: 'myFeatureName'})).toBe('MY_FEATURE_NAME');
    });

    it('should convert PascalCase to CONSTANT_CASE', () => {
      const template = Handlebars.compile('{{constantCase name}}');
      expect(template({name: 'MyFeatureName'})).toBe('MY_FEATURE_NAME');
    });

    it('should convert kebab-case to CONSTANT_CASE', () => {
      const template = Handlebars.compile('{{constantCase name}}');
      expect(template({name: 'my-feature-name'})).toBe('MY_FEATURE_NAME');
    });

    it('should handle already CONSTANT_CASE strings', () => {
      const template = Handlebars.compile('{{constantCase name}}');
      expect(template({name: 'MY_FEATURE_NAME'})).toBe('MY_FEATURE_NAME');
    });

    it('should handle spaces', () => {
      const template = Handlebars.compile('{{constantCase name}}');
      expect(template({name: 'my feature name'})).toBe('MY_FEATURE_NAME');
    });

    it('should handle single word', () => {
      const template = Handlebars.compile('{{constantCase name}}');
      expect(template({name: 'feature'})).toBe('FEATURE');
    });
  });

  describe('if_eq helper', () => {
    it('should render content when values are equal', () => {
      const template = Handlebars.compile('{{#if_eq value "test"}}Equal{{/if_eq}}');
      expect(template({value: 'test'})).toBe('Equal');
    });

    it('should not render content when values are not equal', () => {
      const template = Handlebars.compile('{{#if_eq value "test"}}Equal{{/if_eq}}');
      expect(template({value: 'other'})).toBe('');
    });

    it('should render else block when values are not equal', () => {
      const template = Handlebars.compile('{{#if_eq value "test"}}Equal{{else}}Not Equal{{/if_eq}}');
      expect(template({value: 'other'})).toBe('Not Equal');
    });

    it('should handle number comparisons', () => {
      const template = Handlebars.compile('{{#if_eq count 5}}Five{{/if_eq}}');
      expect(template({count: 5})).toBe('Five');
      expect(template({count: 3})).toBe('');
    });

    it('should handle boolean comparisons', () => {
      const template = Handlebars.compile('{{#if_eq flag true}}True{{/if_eq}}');
      expect(template({flag: true})).toBe('True');
      expect(template({flag: false})).toBe('');
    });

    it('should use strict equality', () => {
      const template = Handlebars.compile('{{#if_eq value "5"}}String{{else}}Number{{/if_eq}}');
      expect(template({value: 5})).toBe('Number');
      expect(template({value: '5'})).toBe('String');
    });
  });

  describe('if_includes helper', () => {
    it('should render content when array includes item', () => {
      const template = Handlebars.compile('{{#if_includes items "test"}}Found{{/if_includes}}');
      expect(template({items: ['test', 'other']})).toBe('Found');
    });

    it('should not render content when array does not include item', () => {
      const template = Handlebars.compile('{{#if_includes items "test"}}Found{{/if_includes}}');
      expect(template({items: ['other', 'another']})).toBe('');
    });

    it('should render else block when item is not found', () => {
      const template = Handlebars.compile('{{#if_includes items "test"}}Found{{else}}Not Found{{/if_includes}}');
      expect(template({items: ['other', 'another']})).toBe('Not Found');
    });

    it('should handle empty arrays', () => {
      const template = Handlebars.compile('{{#if_includes items "test"}}Found{{/if_includes}}');
      expect(template({items: []})).toBe('');
    });

    it('should handle non-array values gracefully', () => {
      const template = Handlebars.compile('{{#if_includes items "test"}}Found{{/if_includes}}');
      expect(template({items: 'not-an-array'})).toBe('');
      expect(template({items: null})).toBe('');
      expect(template({items: undefined})).toBe('');
    });

    it('should handle number items', () => {
      const template = Handlebars.compile('{{#if_includes items 5}}Found{{/if_includes}}');
      expect(template({items: [1, 2, 5, 10]})).toBe('Found');
      expect(template({items: [1, 2, 3]})).toBe('');
    });
  });

  describe('helpers in complex templates', () => {
    it('should work with multiple helpers in one template', () => {
      const template = Handlebars.compile(`
export interface {{pascalCase name}}Props {
  id: string;
}

export const {{camelCase name}} = "{{kebabCase name}}";
export const {{constantCase name}}_ENABLED = true;
`);

      const result = template({name: 'my-feature'});

      expect(result).toContain('export interface MyFeatureProps');
      expect(result).toContain('export const myFeature = "my-feature"');
      expect(result).toContain('export const MY_FEATURE_ENABLED = true');
    });

    it('should work with conditionals and case helpers', () => {
      const template = Handlebars.compile(`
{{#if_eq type "component"}}
export const {{pascalCase name}}: React.FC = () => {
  return <div>{{pascalCase name}}</div>;
};
{{else}}
export const {{camelCase name}} = {};
{{/if_eq}}
`);

      const componentResult = template({name: 'my-widget', type: 'component'});
      expect(componentResult).toContain('export const MyWidget: React.FC');

      const otherResult = template({name: 'my-widget', type: 'other'});
      expect(otherResult).toContain('export const myWidget = {}');
    });
  });
});
