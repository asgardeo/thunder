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

import {describe, it, expect, vi} from 'vitest';
import {render, screen} from '@testing-library/react';
import {CommonStaticStepFactory} from '../CommonStaticStepFactory';
import {StaticStepTypes} from '../../../../models/steps';

// Mock the Start component
vi.mock('../start/Start', () => ({
  default: () => <div data-testid="start-component">Start Component</div>,
}));

describe('CommonStaticStepFactory', () => {
  it('should render Start component when type is StaticStepTypes.Start', () => {
    render(<CommonStaticStepFactory type={StaticStepTypes.Start} />);

    expect(screen.getByTestId('start-component')).toBeInTheDocument();
    expect(screen.getByText('Start Component')).toBeInTheDocument();
  });

  it('should return null when type is StaticStepTypes.UserOnboard', () => {
    const {container} = render(<CommonStaticStepFactory type={StaticStepTypes.UserOnboard} />);

    expect(container.firstChild).toBeNull();
  });

  it('should return null for unknown static step type', () => {
    const {container} = render(<CommonStaticStepFactory type={'UNKNOWN_TYPE' as StaticStepTypes} />);

    expect(container.firstChild).toBeNull();
  });

  it('should render Start component with correct data-testid', () => {
    render(<CommonStaticStepFactory type={StaticStepTypes.Start} />);

    const startComponent = screen.getByTestId('start-component');
    expect(startComponent).toBeInTheDocument();
  });

  describe('StaticStepTypes values', () => {
    it('should have Start type defined', () => {
      expect(StaticStepTypes.Start).toBe('START');
    });

    it('should have UserOnboard type defined', () => {
      expect(StaticStepTypes.UserOnboard).toBe('USER_ONBOARD');
    });
  });
});
