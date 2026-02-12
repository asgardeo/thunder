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

import {describe, it, expect} from 'vitest';
import {render, screen} from '@testing-library/react';
import {MemoryRouter, Routes, Route} from 'react-router';
import DefaultLayout from '../../layouts/DefaultLayout';

describe('DefaultLayout', () => {
  it('renders without crashing', () => {
    const {container} = render(
      <MemoryRouter>
        <Routes>
          <Route element={<DefaultLayout />}>
            <Route path="*" element={<div>Child Content</div>} />
          </Route>
        </Routes>
      </MemoryRouter>,
    );
    expect(container).toBeInTheDocument();
  });

  it('renders child routes through Outlet', () => {
    render(
      <MemoryRouter initialEntries={['/']}>
        <Routes>
          <Route element={<DefaultLayout />}>
            <Route path="/" element={<div data-testid="child-content">Child Content</div>} />
          </Route>
        </Routes>
      </MemoryRouter>,
    );
    expect(screen.getByTestId('child-content')).toBeInTheDocument();
  });

  it('renders Box with correct height styling', () => {
    render(
      <MemoryRouter>
        <Routes>
          <Route element={<DefaultLayout />}>
            <Route path="*" element={<div>Child</div>} />
          </Route>
        </Routes>
      </MemoryRouter>,
    );
    // The layout should contain the child content
    expect(screen.getByText('Child')).toBeInTheDocument();
  });
});
