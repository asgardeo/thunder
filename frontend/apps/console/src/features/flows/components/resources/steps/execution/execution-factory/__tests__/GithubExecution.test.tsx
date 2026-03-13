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
import GithubExecution from '../GithubExecution';

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
    id: 'github-execution-1',
    type: 'TASK_EXECUTION',
    position: {x: 0, y: 0},
    size: {width: 200, height: 100},
    display: {
      label: 'GitHub',
    },
    data: {
      action: {
        executor: {
          name: 'GithubOAuthExecutor',
        },
      },
      properties: {
        idpId: 'github-idp-123',
      },
    },
    config: {},
    ...overrides,
  }) as Step;

describe('GithubExecution', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render the GitHub execution component', () => {
      const resource = createMockResource();
      render(<GithubExecution resource={resource} />);

      // Resource has display.label = 'GitHub', so it shows that instead of the translation key
      expect(screen.getByText('GitHub')).toBeInTheDocument();
    });

    it('should render GitHub icon', () => {
      const resource = createMockResource();
      render(<GithubExecution resource={resource} />);

      const img = screen.getByRole('img');
      expect(img).toHaveAttribute('src', '/static/assets/images/icons/github.svg');
      expect(img).toHaveAttribute('alt', 'github-icon');
      expect(img).toHaveAttribute('height', '20');
    });

    it('should have the correct CSS class', () => {
      const resource = createMockResource();
      render(<GithubExecution resource={resource} />);

      // Resource has display.label = 'GitHub'
      const container = screen.getByText('GitHub').parentElement;
      expect(container).toHaveClass('flow-builder-execution');
      expect(container).toHaveClass('github');
    });
  });

  describe('Required Fields Validation', () => {
    it('should call useRequiredFields with resource and idpId field', () => {
      const resource = createMockResource();
      render(<GithubExecution resource={resource} />);

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
      const resource = createMockResource({id: 'test-github-id'});
      render(<GithubExecution resource={resource} />);

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
      const {rerender} = render(<GithubExecution resource={resource} />);

      const firstCallFields = mockUseRequiredFields.mock.calls[0][2] as unknown[];

      rerender(<GithubExecution resource={resource} />);

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
              name: 'GithubOAuthExecutor',
            },
          },
          properties: {
            idpId: 'configured-github-idp',
          },
        },
      });
      render(<GithubExecution resource={resource} />);

      // Resource has display.label = 'GitHub'
      expect(screen.getByText('GitHub')).toBeInTheDocument();
    });

    it('should handle resource without idpId', () => {
      const resource = createMockResource({
        data: {
          action: {
            executor: {
              name: 'GithubOAuthExecutor',
            },
          },
          properties: {},
        },
      });
      render(<GithubExecution resource={resource} />);

      // Resource has display.label = 'GitHub'
      expect(screen.getByText('GitHub')).toBeInTheDocument();
    });

    it('should handle resource with undefined properties', () => {
      const resource = createMockResource({
        data: {
          action: {
            executor: {
              name: 'GithubOAuthExecutor',
            },
          },
        },
      });
      render(<GithubExecution resource={resource} />);

      // Resource has display.label = 'GitHub'
      expect(screen.getByText('GitHub')).toBeInTheDocument();
    });
  });

  describe('Memoization', () => {
    it('should memoize generalMessage based on resource.id', () => {
      const resource = createMockResource({id: 'github-1'});
      const {rerender} = render(<GithubExecution resource={resource} />);

      const firstCallMessage = mockUseRequiredFields.mock.calls[0][1] as unknown;

      rerender(<GithubExecution resource={resource} />);

      const secondCallMessage = mockUseRequiredFields.mock.calls[1][1] as unknown;

      // Message should be the same reference due to memoization
      expect(firstCallMessage).toBe(secondCallMessage);
    });

    it('should update generalMessage when resource.id changes', () => {
      const resource1 = createMockResource({id: 'github-1'});
      const resource2 = createMockResource({id: 'github-2'});

      const {rerender} = render(<GithubExecution resource={resource1} />);

      const firstCallMessage = mockUseRequiredFields.mock.calls[0][1] as unknown;

      rerender(<GithubExecution resource={resource2} />);

      const secondCallMessage = mockUseRequiredFields.mock.calls[1][1] as unknown;

      // Message should be different due to different resource.id
      expect(firstCallMessage).not.toBe(secondCallMessage);
    });
  });
});
