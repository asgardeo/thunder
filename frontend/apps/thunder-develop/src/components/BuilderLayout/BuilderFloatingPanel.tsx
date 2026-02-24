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
import {Drawer, type DrawerProps} from '@wso2/oxygen-ui';

/**
 * Props interface of {@link BuilderFloatingPanel}
 */
export interface BuilderFloatingPanelProps {
  /**
   * Whether the panel is open.
   */
  open: boolean;
  /**
   * Callback invoked when the panel is closed.
   */
  onClose: () => void;
  /**
   * Container element for the MUI Modal portal.
   * Typically the element with `id="drawer-container"` that wraps the canvas.
   * @defaultValue undefined
   */
  container?: Element | null;
  /**
   * Width of the panel in pixels.
   * @defaultValue 350
   */
  width?: number;
  /**
   * Side from which the panel slides in.
   * @defaultValue 'right'
   */
  anchor?: DrawerProps['anchor'];
  /**
   * Additional sx overrides merged into the Drawer paper element.
   * Use this to customise positioning, colours, or spacing per usage context.
   */
  paperSx?: object;
  /**
   * Content rendered inside the floating panel.
   */
  children?: ReactNode;
}

/**
 * Floating side panel rendered as an absolutely-positioned temporary Drawer within a canvas container.
 *
 * The panel renders portal-ed inside the provided `container` element (e.g. `#drawer-container`)
 * so it appears to float over the canvas rather than the full viewport.
 * Pointer events are disabled on the backdrop so canvas interactions remain active when the panel is open.
 *
 * @param props - Props injected to the component.
 * @returns The BuilderFloatingPanel component.
 */
function BuilderFloatingPanel({
  open,
  onClose,
  container = undefined,
  width = 350,
  anchor = 'right',
  paperSx = undefined,
  children = undefined,
  ...rest
}: BuilderFloatingPanelProps): ReactElement {
  return (
    <Drawer
      open={open}
      anchor={anchor}
      onClose={onClose}
      elevation={5}
      slotProps={{
        paper: {
          sx: {
            width,
            p: 2,
            boxShadow: '-2px 0px 12px 0 rgba(0, 0, 0, 0.08)',
            top: '66px',
            right: '8px',
            bottom: '8px',
            height: 'calc(100% - 72px)',
            backdropFilter: 'blur(10px)',
            display: 'flex',
            flexDirection: 'column',
            pointerEvents: 'auto',
            ...(paperSx ?? {}),
          },
          style: {position: 'absolute'},
        },
        backdrop: {
          style: {position: 'absolute'},
        },
      }}
      ModalProps={{
        container,
        keepMounted: true,
        style: {pointerEvents: 'none'},
      }}
      sx={{pointerEvents: 'none'}}
      hideBackdrop
      variant="temporary"
      {...rest}
    >
      {children}
    </Drawer>
  );
}

export default memo(BuilderFloatingPanel);
