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

import {BrowserRouter, Route, Routes} from 'react-router';
import type {JSX} from 'react';
import {ProtectedRoute} from '@asgardeo/react-router';
import UsersListPage from './features/users/pages/UsersListPage';
import CreateUserPage from './features/users/pages/CreateUserPage';
import ViewUserPage from './features/users/pages/ViewUserPage';
import UserTypesListPage from './features/user-types/pages/UserTypesListPage';
import CreateUserTypePage from './features/user-types/pages/CreateUserTypePage';
import ViewUserTypePage from './features/user-types/pages/ViewUserTypePage';
import IntegrationsPage from './features/integrations/pages/IntegrationsPage';
import ApplicationsListPage from './features/applications/pages/ApplicationsListPage';
import ApplicationCreatePage from './features/applications/pages/ApplicationCreatePage';
import ApplicationEditPage from './features/applications/pages/ApplicationEditPage';
import DashboardLayout from './layouts/DashboardLayout';
import FullScreenLayout from './layouts/FullScreenLayout';
import ApplicationCreateProvider from './features/applications/contexts/ApplicationCreate/ApplicationCreateProvider';
import FlowsListPage from './features/flows/pages/FlowsListPage';
import LoginFlowBuilderPage from './features/login-flow/pages/LoginFlowPage';
import OrganizationUnitsListPage from './features/organization-units/pages/OrganizationUnitsListPage';
import CreateOrganizationUnitPage from './features/organization-units/pages/CreateOrganizationUnitPage';
import OrganizationUnitEditPage from './features/organization-units/pages/OrganizationUnitEditPage';

export default function App(): JSX.Element {
  return (
    <BrowserRouter basename={import.meta.env.BASE_URL}>
      <Routes>
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <DashboardLayout />
            </ProtectedRoute>
          }
        >
          <Route index element={<UsersListPage />} />
          <Route path="users" element={<UsersListPage />} />
          <Route path="users/create" element={<CreateUserPage />} />
          <Route path="users/:userId" element={<ViewUserPage />} />
          <Route path="user-types" element={<UserTypesListPage />} />
          <Route path="user-types/create" element={<CreateUserTypePage />} />
          <Route path="user-types/:id" element={<ViewUserTypePage />} />
          <Route path="integrations" element={<IntegrationsPage />} />
          <Route path="applications" element={<ApplicationsListPage />} />
          <Route path="applications/:applicationId" element={<ApplicationEditPage />} />
          <Route path="flows" element={<FlowsListPage />} />
          <Route path="organization-units" element={<OrganizationUnitsListPage />} />
          <Route path="organization-units/:id" element={<OrganizationUnitEditPage />} />
        </Route>
        <Route
          path="/organization-units/create"
          element={
            <ProtectedRoute>
              <FullScreenLayout />
            </ProtectedRoute>
          }
        >
          <Route index element={<CreateOrganizationUnitPage />} />
        </Route>
        <Route
          path="/applications/create"
          element={
            <ProtectedRoute>
              <ApplicationCreateProvider>
                <FullScreenLayout />
              </ApplicationCreateProvider>
            </ProtectedRoute>
          }
        >
          <Route index element={<ApplicationCreatePage />} />
        </Route>
        <Route
          path="/flows/signin"
          element={
            <ProtectedRoute>
              <DashboardLayout dense />
            </ProtectedRoute>
          }
        >
          <Route index element={<LoginFlowBuilderPage />} />
        </Route>
        <Route
          path="/flows/signin/:flowId"
          element={
            <ProtectedRoute>
              <DashboardLayout dense />
            </ProtectedRoute>
          }
        >
          <Route index element={<LoginFlowBuilderPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
