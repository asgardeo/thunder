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

import {memo, type ReactElement, type ReactNode} from 'react';
import {Box, Drawer, Typography} from '@wso2/oxygen-ui';

/**
 * Props interface of {@link BuilderStaticPanel}
 */
export interface BuilderStaticPanelProps {
  /**
   * Width of the panel in pixels.
   * @defaultValue 350
   */
  width?: number;
  /**
   * Side the panel is anchored to. Controls the direction of the drop shadow.
   * @defaultValue 'right'
   */
  anchor?: 'left' | 'right';
  /**
   * Optional header content.
   * - Pass a `string` to render a standard uppercase label inside the header bar.
   * - Pass a `ReactNode` for a fully custom header.
   * - Omit to render no header.
   */
  header?: ReactNode;
  /**
   * Additional sx overrides merged into the Drawer paper element.
   */
  paperSx?: object;
  /**
   * Main scrollable content rendered inside the panel body.
   */
  children?: ReactNode;
  /**
   * Whether the panel is open. Controls the expand/collapse state and transition.
   * @defaultValue true
   */
  open?: boolean;
}

/**
 * Persistent, in-flow side panel whose visual style matches {@link BuilderFloatingPanel}.
 *
 * Unlike the floating panel this component is part of the normal document flow — it pushes
 * adjacent content rather than overlaying the canvas. Use it for always-visible config or
 * property panels in a builder layout.
 *
 * @param props - Props injected to the component.
 * @returns The BuilderStaticPanel component.
 */
function BuilderStaticPanel({
  width = 350,
  anchor = 'right',
  header = undefined,
  paperSx = undefined,
  children = undefined,
  open = true,
}: BuilderStaticPanelProps): ReactElement {
  return (
    <Drawer
      variant="persistent"
      anchor={anchor}
      open
      elevation={5}
      sx={{
        width,
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
          width,
          position: 'relative',
          border: 'none',
          overflow: 'scroll',
          p: 2,
          gap: 1,
          ...(paperSx ?? {}),
        },
      }}
    >
      {/* Header bar */}
      {header !== undefined && (
        <Box
          sx={{
            height: 40,
            flexShrink: 0,
            px: 2,
            display: 'flex',
            alignItems: 'center',
            borderBottom: '1px solid',
            borderColor: 'divider',
          }}
        >
          {typeof header === 'string' ? <Typography variant="h6">{header}</Typography> : header}
        </Box>
      )}

      {/* Body */}
      <Box sx={{flex: 1, overflow: 'hidden'}}>{children}</Box>
    </Drawer>
  );
}

export default memo(BuilderStaticPanel);
