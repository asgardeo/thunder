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

/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable @typescript-eslint/no-unsafe-assignment */
/* eslint-disable @typescript-eslint/no-unsafe-call */

import {describe, it, expect, vi} from 'vitest';
import {render, screen} from '@testing-library/react';
import TimerAdapter from '../TimerAdapter';

vi.mock('@wso2/oxygen-ui', () => ({
  Alert: ({children, severity}: any) => (
    <div data-testid="timer-alert" data-severity={severity}>
      {children}
    </div>
  ),
  Typography: ({children, color}: any) => (
    <p data-testid="timer-text" data-color={color}>
      {children}
    </p>
  ),
}));

vi.mock('@asgardeo/react', () => ({
  FlowTimer: ({expiresIn, children}: any) => {
    const isExpired = expiresIn <= 0;
    const formattedTime = isExpired ? 'Timed out' : '1:15';

    return (
      <div data-testid="sdk-flow-timer" data-expires-in={String(expiresIn)}>
        {children?.({formattedTime, isExpired, remaining: isExpired ? 0 : 75})}
      </div>
    );
  },
}));

describe('TimerAdapter', () => {
  it('passes expiresIn to FlowTimer', () => {
    render(<TimerAdapter expiresIn={45} />);
    expect(screen.getByTestId('sdk-flow-timer')).toHaveAttribute('data-expires-in', '45');
  });

  it('renders active countdown text with default template when not expired', () => {
    render(<TimerAdapter expiresIn={45} />);

    expect(screen.getByTestId('timer-text')).toHaveTextContent('Time remaining: 1:15');
    expect(screen.getByTestId('timer-text')).toHaveAttribute('data-color', 'warning.main');
    expect(screen.queryByTestId('timer-alert')).toBeNull();
  });

  it('renders active countdown text with custom textTemplate', () => {
    render(<TimerAdapter expiresIn={45} textTemplate="Expires in {time}" />);

    expect(screen.getByTestId('timer-text')).toHaveTextContent('Expires in 1:15');
  });

  it('renders timeout alert when expired', () => {
    render(<TimerAdapter expiresIn={0} />);

    expect(screen.getByTestId('timer-alert')).toHaveAttribute('data-severity', 'warning');
    expect(screen.getByText('Timed out')).toBeInTheDocument();
    expect(screen.queryByText(/Time remaining/i)).toBeNull();
  });
});
