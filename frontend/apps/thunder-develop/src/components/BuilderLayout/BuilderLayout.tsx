/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

import {Box, Drawer, IconButton, Tooltip} from '@wso2/oxygen-ui';
import {ChevronRightIcon} from '@wso2/oxygen-ui-icons-react';
import {memo, type HTMLAttributes, type ReactElement, type ReactNode} from 'react';

/**
 * Props interface of {@link BuilderLayout}
 */
export interface BuilderLayoutProps extends HTMLAttributes<HTMLDivElement> {
  /**
   * Whether the side panel is currently open.
   * @defaultValue true
   */
  open?: boolean;
  /**
   * Callback invoked when the panel expand/collapse toggle is clicked.
   */
  onPanelToggle: () => void;
  /**
   * Width in pixels of the side panel drawer.
   * @defaultValue 350
   */
  panelWidth?: number;
  /**
   * Content rendered inside the sliding drawer panel.
   */
  panelContent?: ReactNode;
  /**
   * Tooltip label shown on the floating expand button (visible only when panel is collapsed).
   * @defaultValue "Show panel"
   */
  expandTooltip?: string;
  /**
   * Additional sx overrides applied to the sliding panel Drawer's paper element.
   * Use to customise padding, overflow, border, display, etc. for specific builder contexts.
   */
  panelPaperSx?: object;
  /**
   * Main canvas / editor content rendered to the right of the panel.
   */
  children?: ReactNode;
}

/**
 * Generic two-column builder layout with a collapsible left panel and a main content area.
 *
 * The left panel is rendered inside a persistent MUI Drawer that slides in/out.
 * A floating expand button appears when the panel is collapsed.
 *
 * @param props - Props injected to the component.
 * @returns The BuilderLayout component.
 */
function BuilderLayout({
  open = true,
  onPanelToggle,
  panelWidth = 350,
  panelContent = undefined,
  expandTooltip = 'Show panel',
  panelPaperSx = undefined,
  children = undefined,
  rightPanel = undefined,
  ...rest
}: BuilderLayoutProps): ReactElement {
  return (
    <Box width="100%" height="100%" display="flex" position="relative" {...rest}>
      {/* Floating expand button — visible only when panel is collapsed */}
      {!open && (
        <Tooltip title={expandTooltip} placement="right">
          <IconButton
            onClick={onPanelToggle}
            size="small"
            sx={{
              position: 'absolute',
              top: 16,
              left: 16,
              zIndex: 10,
              borderRadius: 1,
              '&:hover': {
                backgroundColor: 'action.hover',
              },
            }}
          >
            <ChevronRightIcon size={16} />
          </IconButton>
        </Tooltip>
      )}

      {/* Sliding side panel */}
      <Drawer
        variant="persistent"
        anchor="left"
        open={open ?? false}
        sx={{
          width: panelWidth,
          height: '100%',
          flexShrink: 0,
          mr: 1,
          transition: (theme) =>
            theme.transitions.create('width', {
              easing: open ? theme.transitions.easing.easeOut : theme.transitions.easing.sharp,
              duration: open ? theme.transitions.duration.enteringScreen : theme.transitions.duration.leavingScreen,
            }),
          ...(!open && {
            width: 0,
            mr: 0,
          }),
          '& .MuiDrawer-paper': {
            width: panelWidth,
            position: 'relative',
            border: 'none',
            overflow: 'scroll',
            p: 2,
            gap: 1,
            ...(panelPaperSx ?? {}),
          },
        }}
      >
        {panelContent}
      </Drawer>

      {/* Main content area (canvas / editor) */}
      <Box
        component="main"
        sx={{
          flexGrow: 1,
          height: '100%',
          position: 'relative',
          transition: (theme) =>
            theme.transitions.create(['margin', 'width'], {
              easing: open ? theme.transitions.easing.easeOut : theme.transitions.easing.sharp,
              duration: open ? theme.transitions.duration.enteringScreen : theme.transitions.duration.leavingScreen,
            }),
        }}
      >
        {children}
      </Box>
      {rightPanel}
    </Box>
  );
}

export default memo(BuilderLayout);
