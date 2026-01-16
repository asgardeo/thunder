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

import {
  Avatar,
  ColorSchemeImage,
  Drawer,
  drawerClasses,
  Box,
  Divider,
  Stack,
  Typography,
  IconButton,
  useTheme,
} from '@wso2/oxygen-ui';
import {User} from '@asgardeo/react';
import {Menu} from '@wso2/oxygen-ui-icons-react';
import {useState, useEffect, useMemo, type JSX} from 'react';
import MenuContent from './MenuContent';
import OptionsMenu from './OptionsMenu';
import SidebarContext from './context/SidebarContext';
import {DRAWER_WIDTH, MINI_DRAWER_WIDTH} from './constants';

export interface SideMenuProps {
  defaultExpanded?: boolean;
  onExpandedChange?: (expanded: boolean) => void;
  disableCollapsible?: boolean;
}

export default function SideMenu({
  defaultExpanded = true,
  onExpandedChange,
  disableCollapsible = false,
}: SideMenuProps = {}): JSX.Element {
  const theme = useTheme();
  const [expanded, setExpanded] = useState(defaultExpanded);

  // Sync expanded state when defaultExpanded changes (e.g., navigating between routes)
  useEffect(() => {
    setExpanded(defaultExpanded);
  }, [defaultExpanded]);

  const [isFullyExpanded, setIsFullyExpanded] = useState(expanded);
  const [isFullyCollapsed, setIsFullyCollapsed] = useState(!expanded);

  useEffect(() => {
    if (expanded) {
      const drawerWidthTransitionTimeout = setTimeout(() => {
        setIsFullyExpanded(true);
      }, theme.transitions.duration.enteringScreen);

      return () => clearTimeout(drawerWidthTransitionTimeout);
    }

    setIsFullyExpanded(false);

    return () => {};
  }, [expanded, theme.transitions.duration.enteringScreen]);

  useEffect(() => {
    if (!expanded) {
      const drawerWidthTransitionTimeout = setTimeout(() => {
        setIsFullyCollapsed(true);
      }, theme.transitions.duration.leavingScreen);

      return () => clearTimeout(drawerWidthTransitionTimeout);
    }

    setIsFullyCollapsed(false);

    return () => {};
  }, [expanded, theme.transitions.duration.leavingScreen]);

  const mini = !disableCollapsible && !expanded;

  const handleToggle = () => {
    const newExpanded = !expanded;
    setExpanded(newExpanded);
    onExpandedChange?.(newExpanded);
  };

  const sidebarContextValue = useMemo(
    () => ({
      mini,
      fullyExpanded: isFullyExpanded,
      fullyCollapsed: isFullyCollapsed,
      hasDrawerTransitions: true,
    }),
    [mini, isFullyExpanded, isFullyCollapsed],
  );

  const drawerWidth = mini ? MINI_DRAWER_WIDTH : DRAWER_WIDTH;

  return (
    <SidebarContext.Provider value={sidebarContextValue}>
      <Drawer
        variant="permanent"
        open
        sx={{
          display: {xs: 'none', md: 'block'},
          width: drawerWidth,
          flexShrink: 0,
          boxSizing: 'border-box',
          whiteSpace: 'nowrap',
          transition: (t) =>
            t.transitions.create('width', {
              easing: t.transitions.easing.sharp,
              duration: expanded ? t.transitions.duration.enteringScreen : t.transitions.duration.leavingScreen,
            }),
          [`& .${drawerClasses.paper}`]: {
            width: drawerWidth,
            boxSizing: 'border-box',
            overflowX: 'hidden',
            borderRadius: '0 !important',
            transition: (t) =>
              t.transitions.create('width', {
                easing: t.transitions.easing.sharp,
                duration: expanded ? t.transitions.duration.enteringScreen : t.transitions.duration.leavingScreen,
              }),
          },
        }}
      >
        <Box
          sx={{
            display: 'flex',
            height: 55,
            px: 2,
            justifyContent: mini ? 'center' : 'flex-start',
            alignItems: 'center',
            overflow: 'hidden',
            gap: 2,
          }}
        >
          <IconButton onClick={handleToggle} size="small" aria-label="Expand/Collapse sidebar">
            <Menu size={20} />
          </IconButton>
          {!mini && (
            <Box
              sx={{
                display: 'flex',
                alignItems: 'center',
              }}
            >
              <ColorSchemeImage
                src={{
                  light: `${import.meta.env.BASE_URL}/assets/images/logo.svg`,
                  dark: `${import.meta.env.BASE_URL}/assets/images/logo-inverted.svg`,
                }}
                alt={{light: 'Logo (Light)', dark: 'Logo (Dark)'}}
                height={20}
                width="auto"
                alignItems="center"
                marginBottom="3px"
              />
              <Typography variant="subtitle1" sx={{ml: 1, alignSelf: 'center', fontWeight: 400}}>
                Developer
              </Typography>
            </Box>
          )}
        </Box>
        <Divider />
        <Box
          sx={{
            overflow: 'auto',
            height: '100%',
            display: 'flex',
            flexDirection: 'column',
            scrollbarGutter: mini ? 'stable' : 'auto',
            overflowX: 'hidden',
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
            justifyContent: mini ? 'center' : 'flex-start',
          }}
        >
          <User>
            {(user) => (
              <>
                <Avatar sizes="small" alt={user?.name as string} sx={{width: 36, height: 36}}>
                  {(user?.name as string)?.charAt(0).toUpperCase()}
                </Avatar>
                {!mini && (
                  <>
                    <Box sx={{mr: 'auto', minWidth: 0, flex: 1, overflow: 'hidden'}}>
                      <Typography
                        variant="body2"
                        sx={{
                          fontWeight: 500,
                          lineHeight: '16px',
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          whiteSpace: 'nowrap',
                        }}
                      >
                        {user?.name}
                      </Typography>
                      <Typography
                        variant="caption"
                        sx={{
                          color: 'text.secondary',
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          whiteSpace: 'nowrap',
                        }}
                      >
                        {user?.email}
                      </Typography>
                    </Box>
                    <Box sx={{flexShrink: 0}}>
                      <OptionsMenu />
                    </Box>
                  </>
                )}
              </>
            )}
          </User>
        </Stack>
      </Drawer>
    </SidebarContext.Provider>
  );
}
