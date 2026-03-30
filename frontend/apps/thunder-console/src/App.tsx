/**
 * Copyright (c) 2025-2026, WSO2 LLC. (https://www.wso2.com).
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

import {ProtectedRoute} from '@asgardeo/react-router';
import {ToastProvider} from '@thunder/shared-contexts';
import type {JSX} from 'react';
import {BrowserRouter, Route, Routes} from 'react-router';
import ApplicationCreateProvider from './features/applications/contexts/ApplicationCreate/ApplicationCreateProvider';
import ApplicationCreatePage from './features/applications/pages/ApplicationCreatePage';
import ApplicationEditPage from './features/applications/pages/ApplicationEditPage';
import ApplicationsListPage from './features/applications/pages/ApplicationsListPage';
import LayoutBuilderProvider from './features/design/contexts/LayoutBuilder/LayoutBuilderProvider';
import ThemeBuilderProvider from './features/design/contexts/ThemeBuilder/ThemeBuilderProvider';
import DesignPage from './features/design/pages/DesignPage';
import LayoutBuilderPage from './features/design/pages/LayoutBuilderPage';
import ThemeBuilderPage from './features/design/pages/ThemeBuilderPage';
import ThemeCreatePage from './features/design/pages/ThemeCreatePage';
import FlowsListPage from './features/flows/pages/FlowsListPage';
import GroupCreateProvider from './features/groups/contexts/GroupCreate/GroupCreateProvider';
import CreateGroupPage from './features/groups/pages/CreateGroupPage';
import GroupEditPage from './features/groups/pages/GroupEditPage';
import GroupsListPage from './features/groups/pages/GroupsListPage';
import HomePage from './features/home/pages/HomePage';
import IntegrationsPage from './features/integrations/pages/IntegrationsPage';
import LoginFlowBuilderPage from './features/login-flow/pages/LoginFlowPage';
import OrganizationUnitProvider from './features/organization-units/contexts/OrganizationUnitProvider';
import CreateOrganizationUnitPage from './features/organization-units/pages/CreateOrganizationUnitPage';
import OrganizationUnitEditPage from './features/organization-units/pages/OrganizationUnitEditPage';
import OrganizationUnitsListPage from './features/organization-units/pages/OrganizationUnitsListPage';
import RoleCreateProvider from './features/roles/contexts/RoleCreate/RoleCreateProvider';
import CreateRolePage from './features/roles/pages/CreateRolePage';
import RoleEditPage from './features/roles/pages/RoleEditPage';
import RolesListPage from './features/roles/pages/RolesListPage';
import TranslationCreateProvider from './features/translations/contexts/TranslationCreate/TranslationCreateProvider';
import TranslationCreatePage from './features/translations/pages/TranslationCreatePage';
import TranslationsEditPage from './features/translations/pages/TranslationsEditPage';
import TranslationsListPage from './features/translations/pages/TranslationsListPage';
import UserTypeCreateProvider from './features/user-types/contexts/UserTypeCreate/UserTypeCreateProvider';
import CreateUserTypePage from './features/user-types/pages/CreateUserTypePage';
import UserTypesListPage from './features/user-types/pages/UserTypesListPage';
import ViewUserTypePage from './features/user-types/pages/ViewUserTypePage';
import UserCreateProvider from './features/users/contexts/UserCreate/UserCreateProvider';
import CreateUserPage from './features/users/pages/CreateUserPage';
import InviteUserPage from './features/users/pages/InviteUserPage';
import UsersListPage from './features/users/pages/UsersListPage';
import ViewUserPage from './features/users/pages/ViewUserPage';
import DashboardLayout from './layouts/DashboardLayout';
import FullScreenLayout from './layouts/FullScreenLayout';

export default function App(): JSX.Element {
  return (
    <BrowserRouter basename={import.meta.env.BASE_URL}>
      <ToastProvider>
        <Routes>
          <Route
            path="/"
            element={
              <ProtectedRoute>
                <DashboardLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<HomePage />} />
            <Route path="home" element={<HomePage />} />
            <Route path="users" element={<UsersListPage />} />
            <Route path="users/:userId" element={<ViewUserPage />} />
            <Route path="user-types" element={<UserTypesListPage />} />
            <Route path="user-types/:id" element={<ViewUserTypePage />} />
            <Route path="integrations" element={<IntegrationsPage />} />
            <Route path="groups" element={<GroupsListPage />} />
            <Route path="groups/:groupId" element={<GroupEditPage />} />
            <Route path="roles" element={<RolesListPage />} />
            <Route path="roles/:roleId" element={<RoleEditPage />} />
            <Route path="applications" element={<ApplicationsListPage />} />
            <Route path="applications/:applicationId" element={<ApplicationEditPage />} />
            <Route path="flows" element={<FlowsListPage />} />
          </Route>
          {/* Organization Units - wrapped in OrganizationUnitProvider to preserve tree state across navigation */}
          <Route
            path="/organization-units"
            element={
              <ProtectedRoute>
                <OrganizationUnitProvider />
              </ProtectedRoute>
            }
          >
            <Route element={<DashboardLayout />}>
              <Route index element={<OrganizationUnitsListPage />} />
              <Route path=":id" element={<OrganizationUnitEditPage />} />
            </Route>
            <Route path="create" element={<FullScreenLayout />}>
              <Route index element={<CreateOrganizationUnitPage />} />
            </Route>
          </Route>
          <Route
            path="/groups/create"
            element={
              <ProtectedRoute>
                <GroupCreateProvider>
                  <FullScreenLayout />
                </GroupCreateProvider>
              </ProtectedRoute>
            }
          >
            <Route index element={<CreateGroupPage />} />
          </Route>
          <Route
            path="/roles/create"
            element={
              <ProtectedRoute>
                <RoleCreateProvider>
                  <FullScreenLayout />
                </RoleCreateProvider>
              </ProtectedRoute>
            }
          >
            <Route index element={<CreateRolePage />} />
          </Route>
          <Route
            path="/users/create"
            element={
              <ProtectedRoute>
                <UserCreateProvider>
                  <FullScreenLayout />
                </UserCreateProvider>
              </ProtectedRoute>
            }
          >
            <Route index element={<CreateUserPage />} />
          </Route>
          <Route
            path="/users/invite"
            element={
              <ProtectedRoute>
                <FullScreenLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<InviteUserPage />} />
          </Route>
          <Route
            path="/user-types/create"
            element={
              <ProtectedRoute>
                <UserTypeCreateProvider>
                  <FullScreenLayout />
                </UserTypeCreateProvider>
              </ProtectedRoute>
            }
          >
            <Route index element={<CreateUserTypePage />} />
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
                <DashboardLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<LoginFlowBuilderPage />} />
          </Route>
          <Route
            path="/flows/signin/:flowId"
            element={
              <ProtectedRoute>
                <DashboardLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<LoginFlowBuilderPage />} />
          </Route>
          <Route
            path="/design"
            element={
              <ProtectedRoute>
                <DashboardLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<DesignPage />} />
          </Route>
          <Route
            path="/design/themes/create"
            element={
              <ProtectedRoute>
                <FullScreenLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<ThemeCreatePage />} />
          </Route>
          <Route
            path="/design/themes/:themeId"
            element={
              <ProtectedRoute>
                <ThemeBuilderProvider>
                  <DashboardLayout />
                </ThemeBuilderProvider>
              </ProtectedRoute>
            }
          >
            <Route index element={<ThemeBuilderPage />} />
          </Route>
          <Route
            path="/design/layouts/:layoutId"
            element={
              <ProtectedRoute>
                <LayoutBuilderProvider>
                  <DashboardLayout />
                </LayoutBuilderProvider>
              </ProtectedRoute>
            }
          >
            <Route index element={<LayoutBuilderPage />} />
          </Route>
          <Route
            path="/translations/create"
            element={
              <ProtectedRoute>
                <TranslationCreateProvider>
                  <FullScreenLayout />
                </TranslationCreateProvider>
              </ProtectedRoute>
            }
          >
            <Route index element={<TranslationCreatePage />} />
          </Route>
          <Route
            path="/translations"
            element={
              <ProtectedRoute>
                <DashboardLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<TranslationsListPage />} />
            <Route path=":language" element={<TranslationsEditPage />} />
          </Route>
        </Routes>
      </ToastProvider>
    </BrowserRouter>
  );
}
