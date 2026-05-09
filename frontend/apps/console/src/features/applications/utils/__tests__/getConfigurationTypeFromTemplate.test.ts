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

import {describe, expect, it} from 'vitest';
import {ApplicationCreateFlowConfiguration} from '../../models/application-create-flow';
import getConfigurationTypeFromTemplate from '../getConfigurationTypeFromTemplate';

const makeTemplate = (name: string, redirectUris?: string[]) => ({
  defaults: {
    name,
    inboundAuthConfig: [
      {
        type: 'oauth2',
        config: {
          grantTypes: ['authorization_code'],
          responseTypes: ['code'],
          redirectUris: redirectUris ?? [],
        },
      },
    ],
  },
});

describe('getConfigurationTypeFromTemplate', () => {
  it('returns NONE for null template config', () => {
    expect(getConfigurationTypeFromTemplate(null)).toBe(ApplicationCreateFlowConfiguration.NONE);
  });

  it('returns NONE when redirectUris is already populated', () => {
    const template = makeTemplate('Browser App', ['https://example.com/callback']);

    expect(getConfigurationTypeFromTemplate(template)).toBe(ApplicationCreateFlowConfiguration.NONE);
  });

  it('returns DEEPLINK for mobile applications', () => {
    const template = makeTemplate('Mobile App');

    expect(getConfigurationTypeFromTemplate(template)).toBe(ApplicationCreateFlowConfiguration.DEEPLINK);
  });

  it('returns URL for browser applications', () => {
    const template = makeTemplate('Browser Application');

    expect(getConfigurationTypeFromTemplate(template)).toBe(ApplicationCreateFlowConfiguration.URL);
  });

  it('returns URL for server applications', () => {
    const template = makeTemplate('Server Application');

    expect(getConfigurationTypeFromTemplate(template)).toBe(ApplicationCreateFlowConfiguration.URL);
  });

  it('returns NONE for backend applications', () => {
    const template = makeTemplate('Backend Application');

    expect(getConfigurationTypeFromTemplate(template)).toBe(ApplicationCreateFlowConfiguration.NONE);
  });

  it('returns URL as default for unknown application types', () => {
    const template = makeTemplate('Unknown App');

    expect(getConfigurationTypeFromTemplate(template)).toBe(ApplicationCreateFlowConfiguration.URL);
  });

  it('returns URL as default for an empty template with no defaults', () => {
    expect(getConfigurationTypeFromTemplate({})).toBe(ApplicationCreateFlowConfiguration.URL);
  });

  it('handles case-insensitive template names', () => {
    const template = makeTemplate('MOBILE APPLICATION');

    expect(getConfigurationTypeFromTemplate(template)).toBe(ApplicationCreateFlowConfiguration.DEEPLINK);
  });
});
