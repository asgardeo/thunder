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
import {render, screen, fireEvent} from '@testing-library/react';
import type {Resource} from '@/features/flows/models/resources';
import ButtonExtendedProperties from '../ButtonExtendedProperties';

// Mock dependencies
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

vi.mock('../ButtonExtendedProperties.scss', () => ({}));

const mockSetLastInteractedResource = vi.fn();

vi.mock('@/features/flows/hooks/useFlowBuilderCore', () => ({
  default: () => ({
    lastInteractedResource: {action: {type: 'SUBMIT'}},
    setLastInteractedResource: mockSetLastInteractedResource,
  }),
}));

vi.mock('@/features/flows/hooks/useValidationStatus', () => ({
  default: () => ({
    selectedNotification: {
      hasResourceFieldNotification: vi.fn().mockReturnValue(false),
      getResourceFieldNotification: vi.fn().mockReturnValue(''),
    },
  }),
}));

vi.mock('@wso2/oxygen-ui', async () => {
  const actual = await vi.importActual('@wso2/oxygen-ui');
  return {
    ...actual,
    useColorScheme: () => ({mode: 'light', systemMode: 'light'}),
  };
});

vi.mock('@/features/login-flow/api/useGetLoginFlowBuilderActions', () => ({
  default: () => ({
    data: [
      {
        id: 'action-group-1',
        type: 'ACTION_GROUP',
        display: {label: 'Primary Actions'},
        types: [
          {
            id: 'submit-action',
            type: 'ACTION',
            display: {label: 'Submit', image: '/submit.png'},
            action: {type: 'SUBMIT'},
          },
          {
            id: 'cancel-action',
            type: 'ACTION',
            display: {label: 'Cancel', image: '/cancel.png'},
            action: {type: 'CANCEL'},
          },
        ],
      },
    ],
    isLoading: false,
  }),
}));

vi.mock('@/features/flows/utils/resolveStaticResourcePath', () => ({
  default: (path: string) => path,
}));

describe('ButtonExtendedProperties', () => {
  const mockOnChange = vi.fn();

  const createMockResource = (overrides: Partial<Resource> = {}): Resource =>
    ({
      id: 'button-1',
      type: 'ACTION',
      category: 'ACTION',
      resourceType: 'ELEMENT',
      action: {type: 'SUBMIT'},
      ...overrides,
    }) as Resource;

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render the component', () => {
      const resource = createMockResource();

      render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      expect(screen.getByText('flows:core.buttonExtendedProperties.type')).toBeInTheDocument();
    });

    it('should render action group labels', () => {
      const resource = createMockResource();

      render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      expect(screen.getByText('Primary Actions')).toBeInTheDocument();
    });

    it('should render action type options', () => {
      const resource = createMockResource();

      render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      expect(screen.getByText('Submit')).toBeInTheDocument();
      expect(screen.getByText('Cancel')).toBeInTheDocument();
    });

    it('should render dividers', () => {
      const resource = createMockResource();

      const {container} = render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const dividers = container.querySelectorAll('.MuiDivider-root');
      expect(dividers.length).toBeGreaterThan(0);
    });
  });

  describe('Action Selection', () => {
    it('should call onChange when an action type is clicked', () => {
      const resource = createMockResource();

      render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const cancelAction = screen.getByText('Cancel').closest('.MuiCard-root');
      if (cancelAction) {
        fireEvent.click(cancelAction);
      }

      expect(mockOnChange).toHaveBeenCalled();
    });

    it('should call setLastInteractedResource when action is selected', () => {
      const resource = createMockResource();

      render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const submitAction = screen.getByText('Submit').closest('.MuiCard-root');
      if (submitAction) {
        fireEvent.click(submitAction);
      }

      expect(mockSetLastInteractedResource).toHaveBeenCalled();
    });
  });

  describe('Avatar Rendering', () => {
    it('should render avatars for action types', () => {
      const resource = createMockResource();

      const {container} = render(<ButtonExtendedProperties resource={resource} onChange={mockOnChange} />);

      const avatars = container.querySelectorAll('.MuiAvatar-root');
      expect(avatars.length).toBeGreaterThan(0);
    });
  });
});
