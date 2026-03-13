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
import {render} from '@testing-library/react';
import CommonWidgetPropertyFactory from '../CommonWidgetPropertyFactory';
import type {Resource} from '../../../models/resources';
import {WidgetTypes} from '../../../models/widget';
import {ElementTypes} from '../../../models/elements';

describe('CommonWidgetPropertyFactory', () => {
  describe('Default Behavior', () => {
    it('should return null for IdentifierPassword widget', () => {
      const resource: Resource = {
        id: 'widget-1',
        type: WidgetTypes.IdentifierPassword,
        config: {},
      } as Resource;

      const {container} = render(<CommonWidgetPropertyFactory resource={resource} />);

      expect(container.firstChild).toBeNull();
    });

    it('should return null for EmailOTP widget', () => {
      const resource: Resource = {
        id: 'widget-2',
        type: WidgetTypes.EmailOTP,
        config: {},
      } as Resource;

      const {container} = render(<CommonWidgetPropertyFactory resource={resource} />);

      expect(container.firstChild).toBeNull();
    });

    it('should return null for GoogleFederation widget', () => {
      const resource: Resource = {
        id: 'widget-3',
        type: WidgetTypes.GoogleFederation,
        config: {},
      } as Resource;

      const {container} = render(<CommonWidgetPropertyFactory resource={resource} />);

      expect(container.firstChild).toBeNull();
    });

    it('should return null for SMSOTP widget', () => {
      const resource: Resource = {
        id: 'widget-4',
        type: WidgetTypes.SMSOTP,
        config: {},
      } as Resource;

      const {container} = render(<CommonWidgetPropertyFactory resource={resource} />);

      expect(container.firstChild).toBeNull();
    });

    it('should return null for AppleFederation widget', () => {
      const resource: Resource = {
        id: 'widget-5',
        type: WidgetTypes.AppleFederation,
        config: {},
      } as Resource;

      const {container} = render(<CommonWidgetPropertyFactory resource={resource} />);

      expect(container.firstChild).toBeNull();
    });

    it('should return null for FacebookFederation widget', () => {
      const resource: Resource = {
        id: 'widget-6',
        type: WidgetTypes.FacebookFederation,
        config: {},
      } as Resource;

      const {container} = render(<CommonWidgetPropertyFactory resource={resource} />);

      expect(container.firstChild).toBeNull();
    });

    it('should return null for MicrosoftFederation widget', () => {
      const resource: Resource = {
        id: 'widget-7',
        type: WidgetTypes.MicrosoftFederation,
        config: {},
      } as Resource;

      const {container} = render(<CommonWidgetPropertyFactory resource={resource} />);

      expect(container.firstChild).toBeNull();
    });

    it('should return null for GithubFederation widget', () => {
      const resource: Resource = {
        id: 'widget-8',
        type: WidgetTypes.GithubFederation,
        config: {},
      } as Resource;

      const {container} = render(<CommonWidgetPropertyFactory resource={resource} />);

      expect(container.firstChild).toBeNull();
    });

    it('should return null for PasskeyAuthentication widget', () => {
      const resource: Resource = {
        id: 'widget-9',
        type: WidgetTypes.PasskeyAuthentication,
        config: {},
      } as Resource;

      const {container} = render(<CommonWidgetPropertyFactory resource={resource} />);

      expect(container.firstChild).toBeNull();
    });

    it('should return null for MagicLink widget', () => {
      const resource: Resource = {
        id: 'widget-10',
        type: WidgetTypes.MagicLink,
        config: {},
      } as Resource;

      const {container} = render(<CommonWidgetPropertyFactory resource={resource} />);

      expect(container.firstChild).toBeNull();
    });

    it('should return null for unknown widget type', () => {
      const resource: Resource = {
        id: 'widget-unknown',
        type: 'UNKNOWN_WIDGET',
        config: {},
      } as Resource;

      const {container} = render(<CommonWidgetPropertyFactory resource={resource} />);

      expect(container.firstChild).toBeNull();
    });
  });

  describe('Props Handling', () => {
    it('should accept additional props without errors', () => {
      const resource: Resource = {
        id: 'widget-1',
        type: WidgetTypes.IdentifierPassword,
        config: {},
      } as Resource;

      const {container} = render(
        <CommonWidgetPropertyFactory resource={resource} customProp="value" anotherProp={123} />,
      );

      expect(container.firstChild).toBeNull();
    });

    it('should handle resource with complex config', () => {
      const resource: Resource = {
        id: 'widget-1',
        resourceType: 'WIDGET',
        type: WidgetTypes.IdentifierPassword,
        category: 'WIDGET',
        version: '1.0.0',
        deprecated: false,
        deletable: true,
        display: {
          label: 'Test Widget',
          image: '',
          showOnResourcePanel: false,
        },
        config: {
          field: {name: '', type: ElementTypes},
          styles: {},
          nested: {
            deep: {
              value: 'test',
            },
          },
          array: [1, 2, 3],
        },
      } as unknown as Resource;

      const {container} = render(<CommonWidgetPropertyFactory resource={resource} />);

      expect(container.firstChild).toBeNull();
    });
  });
});
