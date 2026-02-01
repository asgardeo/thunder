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

import {NavLink} from 'react-router';
import {List, ListItem, ListItemButton, ListItemIcon, ListItemText, Stack, Tooltip} from '@wso2/oxygen-ui';
import {Blocks, Building, LayoutGrid, User, UsersRound, Workflow} from '@wso2/oxygen-ui-icons-react';
import {useContext, useMemo, type JSX} from 'react';
import {useTranslation} from 'react-i18next';
import useNavigation from '@/layouts/contexts/useNavigation';
import SidebarContext from './context/SidebarContext';

export default function MenuContent(): JSX.Element {
  const {currentPage, setCurrentPage} = useNavigation();
  const {t} = useTranslation();
  const {mini} = useContext(SidebarContext);

  const mainListItems = useMemo(
    () => [
      {
        id: 'users',
        text: t('navigation:pages.users'),
        icon: <UsersRound size={16} />,
        category: 'Dashboard',
        path: '/users',
      },
      {
        id: 'user-types',
        text: t('navigation:pages.userTypes'),
        icon: <User size={16} />,
        category: 'Dashboard',
        path: '/user-types',
      },
      {
        id: 'organization-units',
        text: t('navigation:pages.organizationUnits'),
        icon: <Building size={16} />,
        category: 'Dashboard',
        path: '/organization-units',
      },
      {
        id: 'integrations',
        text: t('navigation:pages.integrations'),
        icon: <Blocks size={16} />,
        category: 'Dashboard',
        path: '/integrations',
      },
      {
        id: 'applications',
        text: t('navigation:pages.applications'),
        icon: <LayoutGrid size={16} />,
        category: 'Dashboard',
        path: '/applications',
      },
      {
        id: 'flows',
        text: t('navigation:pages.flows'),
        icon: <Workflow size={16} />,
        category: 'Dashboard',
        path: '/flows',
      },
    ],
    [t],
  );

  const handleListItemClick = (item: {id: string; text: string; category: string}) => {
    setCurrentPage(item.id);
  };

  return (
    <Stack sx={{flexGrow: 1, p: 1, justifyContent: 'space-between'}}>
      <List dense>
        {mainListItems.map((item) => (
          <ListItem key={item.id} disablePadding sx={{display: 'block'}}>
            <Tooltip title={mini ? item.text : ''} placement="right">
              <ListItemButton
                component={NavLink}
                to={item.path}
                selected={currentPage === item.id}
                onClick={() => handleListItemClick(item)}
                sx={{
                  minHeight: 38,
                  justifyContent: mini ? 'center' : 'initial',
                  px: 2.5,
                }}
              >
                <ListItemIcon
                  sx={{
                    minWidth: 0,
                    mr: mini ? 'auto' : 3,
                    justifyContent: 'center',
                  }}
                >
                  {item.icon}
                </ListItemIcon>
                <ListItemText
                  primary={item.text}
                  sx={{
                    opacity: mini ? 0 : 1,
                    display: mini ? 'none' : 'block',
                  }}
                />
              </ListItemButton>
            </Tooltip>
          </ListItem>
        ))}
      </List>
    </Stack>
  );
}
