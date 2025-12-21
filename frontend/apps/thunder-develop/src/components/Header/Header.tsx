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

import {ColorSchemeToggle, Stack, Tooltip} from '@wso2/oxygen-ui';
import type {JSX} from 'react';
import {Bell, Menu} from '@wso2/oxygen-ui-icons-react';
import {useTranslation} from 'react-i18next';
import NavbarBreadcrumbs from '../Navbar/NavbarBreadcrumbs';
import MenuButton from '../Sidebar/MenuButton';
import Search from './Search';
import LanguageSwitcher from './LanguageSwitcher';

export default function Header(): JSX.Element {
  const {t} = useTranslation();
  return (
    <Stack
      direction="row"
      sx={{
        display: 'flex',
        width: '100%',
        alignItems: {xs: 'flex-start', md: 'center'},
        justifyContent: 'space-between',
        maxWidth: {sm: '100%'},
        height: 50,
        px: 3,
      }}
      spacing={2}
    >
      <Stack direction="row">
        <MenuButton aria-label="menu" sx={{display: {md: 'none'}}}>
          <Menu />
        </MenuButton>
        <NavbarBreadcrumbs />
      </Stack>
      <Stack direction="row" sx={{gap: 1}}>
        <Search />
        <Tooltip title={t('common:header.notifications')}>
          <MenuButton aria-label={t('common:header.openNotifications')}>
            <Bell strokeWidth={1} />
          </MenuButton>
        </Tooltip>
        <LanguageSwitcher />
        <ColorSchemeToggle data-testid="theme-toggle" />
      </Stack>
    </Stack>
  );
}
