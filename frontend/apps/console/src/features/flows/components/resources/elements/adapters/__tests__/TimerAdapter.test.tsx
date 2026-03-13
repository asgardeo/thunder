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
import type {Resource} from '@/features/flows/models/resources';
import TimerAdapter from '../TimerAdapter';

describe('TimerAdapter', () => {
  const createResource = (label?: string): Resource =>
    ({
      id: 'timer-1',
      ...(label !== undefined ? {label} : {}),
    }) as unknown as Resource;

  it('should replace the dynamic time placeholder with a canvas value', () => {
    render(<TimerAdapter resource={createResource('Try again in {time}')} />);

    expect(screen.getByText('Try again in 05:00')).toBeInTheDocument();
  });

  it('should render default text when resource label is missing', () => {
    render(<TimerAdapter resource={createResource()} />);

    expect(screen.getByText('Time remaining: 05:00')).toBeInTheDocument();
  });

  it('should render default text when resource is undefined', () => {
    render(<TimerAdapter />);

    expect(screen.getByText('Time remaining: 05:00')).toBeInTheDocument();
  });

  it('should render default text when resource label property is missing', () => {
    const resourceWithoutLabel = {
      id: 'timer-no-label',
    } as unknown as Resource;

    render(<TimerAdapter resource={resourceWithoutLabel} />);

    expect(screen.getByText('Time remaining: 05:00')).toBeInTheDocument();
  });
});
