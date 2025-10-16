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
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import Stack from '@mui/material/Stack';
import HomeRoundedIcon from '@mui/icons-material/HomeRounded';
import AnalyticsRoundedIcon from '@mui/icons-material/AnalyticsRounded';
import PeopleRoundedIcon from '@mui/icons-material/PeopleRounded';
import AssignmentRoundedIcon from '@mui/icons-material/AssignmentRounded';
import SettingsRoundedIcon from '@mui/icons-material/SettingsRounded';
import InfoRoundedIcon from '@mui/icons-material/InfoRounded';
import HelpRoundedIcon from '@mui/icons-material/HelpRounded';
import useNavigation from '@/layouts/contexts/useNavigation';

const mainListItems = [
  {
    id: 'home',
    text: 'Home',
    icon: <HomeRoundedIcon />,
    category: 'Dashboard',
    path: '/',
  },
  {
    id: 'users',
    text: 'Users',
    icon: <PeopleRoundedIcon />,
    category: 'Dashboard',
    path: '/users',
  },
  {
    id: 'analytics',
    text: 'Integrations',
    icon: <AnalyticsRoundedIcon />,
    category: 'Dashboard',
    path: '/',
  },
  {
    id: 'tasks',
    text: 'Applications',
    icon: <AssignmentRoundedIcon />,
    category: 'Dashboard',
    path: '/',
  },
];

const secondaryListItems = [
  {
    id: 'settings',
    text: 'Settings',
    icon: <SettingsRoundedIcon />,
    category: 'Settings',
    path: '/settings',
  },
  {
    id: 'about',
    text: 'About',
    icon: <InfoRoundedIcon />,
    category: 'Settings',
    path: '/about',
  },
  {
    id: 'feedback',
    text: 'Feedback',
    icon: <HelpRoundedIcon />,
    category: 'Settings',
    path: '/feedback',
  },
];

export default function MenuContent() {
  const {currentPage, setCurrentPage} = useNavigation();

  const handleListItemClick = (item: {id: string; text: string; category: string}) => {
    setCurrentPage({
      id: item.id,
      text: item.text,
      category: item.category,
    });
  };

  return (
    <Stack sx={{flexGrow: 1, p: 1, justifyContent: 'space-between'}}>
      <List dense>
        {mainListItems.map((item) => (
          <ListItem key={item.id} disablePadding sx={{display: 'block'}}>
            <ListItemButton
              component={NavLink}
              to={item.path}
              selected={currentPage.id === item.id}
              onClick={() => handleListItemClick(item)}
            >
              <ListItemIcon>{item.icon}</ListItemIcon>
              <ListItemText primary={item.text} />
            </ListItemButton>
          </ListItem>
        ))}
      </List>
      <List dense>
        {secondaryListItems.map((item) => (
          <ListItem key={item.id} disablePadding sx={{display: 'block'}}>
            <ListItemButton
              component={NavLink}
              to={item.path}
              selected={currentPage.id === item.id}
              onClick={() => handleListItemClick(item)}
            >
              <ListItemIcon>{item.icon}</ListItemIcon>
              <ListItemText primary={item.text} />
            </ListItemButton>
          </ListItem>
        ))}
      </List>
    </Stack>
  );
}
