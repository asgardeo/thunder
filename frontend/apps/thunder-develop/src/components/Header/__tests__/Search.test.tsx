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
import {render, screen, userEvent} from '@thunder/test-utils';
import Search from '../Search';

describe('Search', () => {
  it('renders search input with placeholder', () => {
    render(<Search />);

    const searchInput = screen.getByPlaceholderText('Search...');
    expect(searchInput).toBeInTheDocument();
  });

  it('has correct aria-label', () => {
    render(<Search />);

    const searchInput = screen.getByLabelText('Search');
    expect(searchInput).toBeInTheDocument();
  });

  it('renders search icon', () => {
    const {container} = render(<Search />);

    const searchIcon = container.querySelector('svg.lucide-search');
    expect(searchIcon).toBeInTheDocument();
  });

  it('allows typing in the search input', async () => {
    const user = userEvent.setup();
    render(<Search />);

    const searchInput = screen.getByPlaceholderText('Search...');
    await user.type(searchInput, 'test query');

    expect(searchInput).toHaveValue('test query');
  });

  it('has correct input id', () => {
    render(<Search />);

    const searchInput = screen.getByPlaceholderText('Search...');
    expect(searchInput).toHaveAttribute('id', 'search');
  });

  it('is not disabled by default', () => {
    render(<Search />);

    const searchInput = screen.getByPlaceholderText('Search...');
    expect(searchInput).not.toBeDisabled();
  });

  it('clears input when value is deleted', async () => {
    const user = userEvent.setup();
    render(<Search />);

    const searchInput = screen.getByPlaceholderText('Search...');
    await user.type(searchInput, 'test');
    expect(searchInput).toHaveValue('test');

    await user.clear(searchInput);
    expect(searchInput).toHaveValue('');
  });

  it('renders with outlined variant', () => {
    const {container} = render(<Search />);

    const outlinedInput = container.querySelector('.MuiOutlinedInput-root');
    expect(outlinedInput).toBeInTheDocument();
  });

  it('renders search icon as start adornment', () => {
    const {container} = render(<Search />);

    const adornment = container.querySelector('.MuiInputAdornment-positionStart');
    expect(adornment).toBeInTheDocument();
  });

  it('renders input with small size', () => {
    const {container} = render(<Search />);

    const input = container.querySelector('.MuiInputBase-sizeSmall');
    expect(input).toBeInTheDocument();
  });

  it('renders FormControl with correct structure', () => {
    const {container} = render(<Search />);

    const formControl = container.querySelector('.MuiFormControl-root');
    expect(formControl).toBeInTheDocument();

    const outlinedInput = formControl?.querySelector('.MuiOutlinedInput-root');
    expect(outlinedInput).toBeInTheDocument();
  });

  it('preserves content after re-render', () => {
    const {rerender} = render(<Search />);

    rerender(<Search />);

    expect(screen.getByPlaceholderText('Search...')).toBeInTheDocument();
    expect(screen.getByLabelText('Search')).toBeInTheDocument();
  });
});
