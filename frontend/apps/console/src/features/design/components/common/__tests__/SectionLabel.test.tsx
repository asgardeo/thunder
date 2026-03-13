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
import {render, screen} from '@thunder/test-utils';
import SectionLabel from '../SectionLabel';

describe('SectionLabel', () => {
  it('renders the provided text', () => {
    render(<SectionLabel>Position</SectionLabel>);
    expect(screen.getByText('Position')).toBeInTheDocument();
  });

  it('renders different text correctly', () => {
    render(<SectionLabel>Container</SectionLabel>);
    expect(screen.getByText('Container')).toBeInTheDocument();
  });

  it('renders as a Typography element (not hidden)', () => {
    const {container} = render(<SectionLabel>My Label</SectionLabel>);
    expect(container.textContent).toBe('My Label');
  });
});
