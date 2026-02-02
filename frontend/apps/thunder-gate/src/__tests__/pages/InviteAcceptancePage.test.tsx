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
import {render} from '@thunder/test-utils/browser';
import {page} from 'vitest/browser';
import InviteAcceptancePage from '../../pages/InviteAcceptancePage';

// Mock the AcceptInviteBox component
vi.mock('../../components/AcceptInvite/AcceptInviteBox', () => ({
  default: () => <div data-testid="accept-invite-box">AcceptInviteBox Component</div>,
}));

describe('InviteAcceptancePage', () => {
  it('renders AcceptInviteBox component', async () => {
    await render(<InviteAcceptancePage />);
    await expect.element(page.getByTestId('accept-invite-box')).toBeInTheDocument();
  });

  it('renders main element', async () => {
    await render(<InviteAcceptancePage />);
    await expect.element(page.getByRole('main')).toBeInTheDocument();
  });
});
