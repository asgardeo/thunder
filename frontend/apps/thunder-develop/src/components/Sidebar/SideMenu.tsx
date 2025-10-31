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

import {styled} from '@mui/material/styles';
import Avatar from '@mui/material/Avatar';
import MuiDrawer, {drawerClasses} from '@mui/material/Drawer';
import Box from '@mui/material/Box';
import Divider from '@mui/material/Divider';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import {User} from '@asgardeo/react';
import {ThemedIcon} from '@thunder/ui';
import type {JSX} from 'react';
import MenuContent from './MenuContent';
import OptionsMenu from './OptionsMenu';

const drawerWidth = 240;

const Drawer = styled(MuiDrawer)({
  width: drawerWidth,
  flexShrink: 0,
  boxSizing: 'border-box',
  mt: 10,
  [`& .${drawerClasses.paper}`]: {
    width: drawerWidth,
    boxSizing: 'border-box',
  },
});

export default function SideMenu(): JSX.Element {
  return (
    <Drawer
      variant="permanent"
      sx={{
        display: {xs: 'none', md: 'block'},
        [`& .${drawerClasses.paper}`]: {
          backgroundColor: 'background.paper',
        },
      }}
    >
      <Box
        sx={{
          display: 'flex',
          mt: 'calc(var(--template-frame-height, 0px) + 4px)',
          p: 1.5,
          justifyContent: 'center',
          alignItems: 'center',
        }}
      >
        <ThemedIcon
          src={{
            light: `${import.meta.env.BASE_URL}/assets/images/logo.svg`,
            dark: `${import.meta.env.BASE_URL}/assets/images/logo-inverted.svg`,
          }}
          alt={{light: 'Logo (Light)', dark: 'Logo (Dark)'}}
          height={16}
          width="auto"
          alignItems="center"
        />
        <Typography variant="h6" sx={{mt: '4px', ml: 1, alignSelf: 'center', fontWeight: 400}}>
          Develop
        </Typography>
      </Box>
      <Divider />
      <Box
        sx={{
          overflow: 'auto',
          height: '100%',
          display: 'flex',
          flexDirection: 'column',
        }}
      >
        <MenuContent />
      </Box>
      <Stack
        direction="row"
        sx={{
          p: 2,
          gap: 1,
          alignItems: 'center',
          borderTop: '1px solid',
          borderColor: 'divider',
        }}
      >
        <User>
          {(user) => (
            <>
              <Avatar sizes="small" alt={user?.name as string} sx={{width: 36, height: 36}}>
                {(user?.name as string)?.charAt(0).toUpperCase()}
              </Avatar>
              <Box sx={{mr: 'auto'}}>
                <Typography variant="body2" sx={{fontWeight: 500, lineHeight: '16px'}}>
                  {user?.name}
                </Typography>
                <Typography variant="caption" sx={{color: 'text.secondary'}}>
                  {user?.email}
                </Typography>
              </Box>
              <OptionsMenu />
            </>
          )}
        </User>
      </Stack>
    </Drawer>
  );
}
