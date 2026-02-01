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
import type {Step} from '@/features/flows/models/steps';
import GoogleExecution from '../GoogleExecution';

// Use vi.hoisted to define mock function before vi.mock hoisting
const mockUseRequiredFields = vi.hoisted(() => vi.fn());

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
  Trans: ({i18nKey, children = null}: {i18nKey: string; children?: React.ReactNode}) => (
    <span data-i18n-key={i18nKey}>{children}</span>
  ),
}));

// Mock resolveStaticResourcePath
vi.mock('@/features/flows/utils/resolveStaticResourcePath', () => ({
  default: (path: string) => `/static/${path}`,
}));

// Mock useRequiredFields
vi.mock('@/features/flows/hooks/useRequiredFields', () => ({
  default: mockUseRequiredFields,
}));

// Create mock resource
const createMockResource = (overrides: Partial<Step> = {}): Step =>
  ({
    id: 'google-execution-1',
    type: 'TASK_EXECUTION',
    position: {x: 0, y: 0},
    size: {width: 200, height: 100},
    display: {
      label: 'Google',
    },
    data: {
      action: {
        executor: {
          name: 'GoogleOIDCAuthExecutor',
        },
      },
      properties: {
        idpId: 'google-idp-123',
      },
    },
    config: {},
    ...overrides,
  }) as Step;

describe('GoogleExecution', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render the Google execution component', () => {
      const resource = createMockResource();
      render(<GoogleExecution resource={resource} />);

      // Resource has display.label = 'Google', so it shows that instead of the translation key
      expect(screen.getByText('Google')).toBeInTheDocument();
    });

    it('should render Google icon', () => {
      const resource = createMockResource();
      render(<GoogleExecution resource={resource} />);

      const img = screen.getByRole('img');
      expect(img).toHaveAttribute('src', '/static/assets/images/icons/google.svg');
      expect(img).toHaveAttribute('alt', 'google-icon');
      expect(img).toHaveAttribute('height', '20');
    });

    it('should render content in a flex container', () => {
      const resource = createMockResource();
      render(<GoogleExecution resource={resource} />);

      // Resource has display.label = 'Google'
      const container = screen.getByText('Google').parentElement;
      expect(container).toBeInTheDocument();
    });
  });

  describe('Required Fields Validation', () => {
    it('should call useRequiredFields with resource and idpId field', () => {
      const resource = createMockResource();
      render(<GoogleExecution resource={resource} />);

      expect(mockUseRequiredFields).toHaveBeenCalledWith(
        resource,
        expect.anything(),
        expect.arrayContaining([
          expect.objectContaining({
            name: 'data.properties.idpId',
            errorMessage: 'flows:core.validation.fields.input.idpId',
          }),
        ]),
      );
    });

    it('should pass generalMessage as ReactElement to useRequiredFields', () => {
      const resource = createMockResource({id: 'test-google-id'});
      render(<GoogleExecution resource={resource} />);

      expect(mockUseRequiredFields).toHaveBeenCalledWith(
        resource,
        expect.objectContaining({
          props: expect.objectContaining({
            i18nKey: 'flows:core.validation.fields.executor.general',
          }) as Record<string, unknown>,
        }),
        expect.any(Array),
      );
    });

    it('should memoize fields array', () => {
      const resource = createMockResource();
      const {rerender} = render(<GoogleExecution resource={resource} />);

      const firstCallFields = mockUseRequiredFields.mock.calls[0][2] as unknown[];

      rerender(<GoogleExecution resource={resource} />);

      const secondCallFields = mockUseRequiredFields.mock.calls[1][2] as unknown[];

      // Fields should be the same reference due to memoization
      expect(firstCallFields).toBe(secondCallFields);
    });
  });

  describe('Resource Handling', () => {
    it('should handle resource with idpId configured', () => {
      const resource = createMockResource({
        data: {
          action: {
            executor: {
              name: 'GoogleOIDCAuthExecutor',
            },
          },
          properties: {
            idpId: 'configured-google-idp',
          },
        },
      });
      render(<GoogleExecution resource={resource} />);

      // Resource has display.label = 'Google'
      expect(screen.getByText('Google')).toBeInTheDocument();
    });

    it('should handle resource without idpId', () => {
      const resource = createMockResource({
        data: {
          action: {
            executor: {
              name: 'GoogleOIDCAuthExecutor',
            },
          },
          properties: {},
        },
      });
      render(<GoogleExecution resource={resource} />);

      // Resource has display.label = 'Google'
      expect(screen.getByText('Google')).toBeInTheDocument();
    });

    it('should handle resource with undefined properties', () => {
      const resource = createMockResource({
        data: {
          action: {
            executor: {
              name: 'GoogleOIDCAuthExecutor',
            },
          },
        },
      });
      render(<GoogleExecution resource={resource} />);

      // Resource has display.label = 'Google'
      expect(screen.getByText('Google')).toBeInTheDocument();
    });
  });

  describe('Memoization', () => {
    it('should memoize generalMessage based on resource.id', () => {
      const resource = createMockResource({id: 'google-1'});
      const {rerender} = render(<GoogleExecution resource={resource} />);

      const firstCallMessage = mockUseRequiredFields.mock.calls[0][1] as unknown;

      rerender(<GoogleExecution resource={resource} />);

      const secondCallMessage = mockUseRequiredFields.mock.calls[1][1] as unknown;

      // Message should be the same reference due to memoization
      expect(firstCallMessage).toBe(secondCallMessage);
    });

    it('should update generalMessage when resource.id changes', () => {
      const resource1 = createMockResource({id: 'google-1'});
      const resource2 = createMockResource({id: 'google-2'});

      const {rerender} = render(<GoogleExecution resource={resource1} />);

      const firstCallMessage = mockUseRequiredFields.mock.calls[0][1] as unknown;

      rerender(<GoogleExecution resource={resource2} />);

      const secondCallMessage = mockUseRequiredFields.mock.calls[1][1] as unknown;

      // Message should be different due to different resource.id
      expect(firstCallMessage).not.toBe(secondCallMessage);
    });
  });
});
