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

import CssBaseline from '@mui/material/CssBaseline';
import {Outlet} from 'react-router';
import type {ReactNode} from 'react';
import Box from '@mui/material/Box';
import {Layout} from '@thunder/ui';
import SideMenu from '../components/Sidebar/SideMenu';
import Header from '../components/Header/Header';
import NavigationProvider from './NavigationProvider';

export default function DashboardLayout(): ReactNode {
  return (
    <NavigationProvider>
      <Layout.Provider>
        <CssBaseline enableColorScheme />
        <Layout.Root>
          <Layout.Sidebar>
            <SideMenu />
          </Layout.Sidebar>
          <Layout.Content>
            <Layout.Header>
              <Header />
            </Layout.Header>
            <Box sx={{p: 3}}>
              <Outlet />
            </Box>
          </Layout.Content>
        </Layout.Root>
      </Layout.Provider>
    </NavigationProvider>
  );
}
