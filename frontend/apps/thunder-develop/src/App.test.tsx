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
import {render} from '@testing-library/react';
import App from './App';

// Mock the ProtectedRoute component
vi.mock('@asgardeo/react-router', () => ({
  ProtectedRoute: ({children}: {children: React.ReactNode}) => <div data-testid="protected-route">{children}</div>,
}));

// Mock all the page components
vi.mock('./features/users/pages/UsersListPage', () => ({
  default: () => <div data-testid="users-list-page">Users List Page</div>,
}));

vi.mock('./features/users/pages/CreateUserPage', () => ({
  default: () => <div data-testid="create-user-page">Create User Page</div>,
}));

vi.mock('./features/users/pages/ViewUserPage', () => ({
  default: () => <div data-testid="view-user-page">View User Page</div>,
}));

vi.mock('./features/user-types/pages/UserTypesListPage', () => ({
  default: () => <div data-testid="user-types-list-page">User Types List Page</div>,
}));

vi.mock('./features/user-types/pages/CreateUserTypePage', () => ({
  default: () => <div data-testid="create-user-type-page">Create User Type Page</div>,
}));

vi.mock('./features/user-types/pages/ViewUserTypePage', () => ({
  default: () => <div data-testid="view-user-type-page">View User Type Page</div>,
}));

vi.mock('./features/integrations/pages/IntegrationsPage', () => ({
  default: () => <div data-testid="integrations-page">Integrations Page</div>,
}));

vi.mock('./features/applications/pages/ApplicationsPage', () => ({
  default: () => <div data-testid="applications-page">Applications Page</div>,
}));

vi.mock('./layouts/DashboardLayout', () => ({
  default: () => <div data-testid="dashboard-layout">Dashboard Layout</div>,
}));

describe('App', () => {
  it('renders without crashing', () => {
    const {container} = render(<App />);
    expect(container).toBeInTheDocument();
  });
});
