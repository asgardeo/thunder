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

import {describe, it, expect, vi} from 'vitest';
import {render, screen} from '@thunder/test-utils';
import ErrorPage from '../ErrorPage';

// Mock the Error component
vi.mock('../../components/Error/Error', () => ({
  default: () => <div data-testid="error-component">Error Component</div>,
}));

describe('ErrorPage', () => {
  it('renders without crashing', () => {
    const {container} = render(<ErrorPage />);
    expect(container).toBeInTheDocument();
  });

  it('renders Error component', () => {
    render(<ErrorPage />);
    expect(screen.getByTestId('error-component')).toBeInTheDocument();
  });
});
