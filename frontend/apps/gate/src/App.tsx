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
import appRoutes, {type AppRoute} from './config/appRoutes';

export default function App(): JSX.Element {
  return (
    <BrowserRouter basename={import.meta.env.BASE_URL}>
      <Routes>
        {appRoutes.map((route: AppRoute) => (
          <Route key={route.path} path={route.path} element={route.element}>
            {route.children?.map((child: AppRoute, index: number) => (
              <Route
                key={child.path ?? `index-${index}`}
                path={child.path}
                index={child.index}
                element={child.element}
              />
            ))}
          </Route>
        ))}
      </Routes>
    </BrowserRouter>
  );
}
