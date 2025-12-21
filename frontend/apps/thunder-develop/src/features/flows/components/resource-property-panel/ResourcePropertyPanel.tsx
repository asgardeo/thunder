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

import {memo, useCallback, useState, type HTMLAttributes, type ReactElement} from 'react';
import {Box, Button, Drawer, IconButton, type DrawerProps} from '@wso2/oxygen-ui';
import {useReactFlow} from '@xyflow/react';
import classNames from 'classnames';
import {X, TrashIcon} from '@wso2/oxygen-ui-icons-react';
import useFlowBuilderCore from '../../hooks/useFlowBuilderCore';
import ResourceProperties from './ResourceProperties';
import {ResourceTypes} from '../../models/resources';
import {type Element} from '../../models/elements';
import './ResourcePropertyPanel.scss';

/**
 * Props interface of {@link ResourcePropertyPanel}
 */
export interface ResourcePropertyPanelPropsInterface extends DrawerProps, HTMLAttributes<HTMLDivElement> {
  onComponentDelete: (stepId: string, component: Element) => void;
}

/**
 * Component to render the resource property panel.
 *
 * @param props - Props injected to the component.
 * @returns The ResourcePropertyPanel component.
 */
function ResourcePropertyPanel({
  children,
  open,
  anchor = 'right',
  onComponentDelete,
  className,
  ...rest
}: ResourcePropertyPanelPropsInterface): ReactElement {
  const {deleteElements} = useReactFlow();
  const [containerEl, setContainerEl] = useState<HTMLDivElement | null>(null);

  const {
    resourcePropertiesPanelHeading,
    setIsOpenResourcePropertiesPanel,
    lastInteractedStepId,
    lastInteractedResource,
  } = useFlowBuilderCore();

  const handleClose = useCallback(() => {
    setIsOpenResourcePropertiesPanel(false);
  }, [setIsOpenResourcePropertiesPanel]);

  const handleDelete = useCallback(() => {
    if (!lastInteractedResource) return;

    if (lastInteractedResource.resourceType === ResourceTypes.Step) {
      deleteElements({nodes: [{id: lastInteractedResource.id}]}).catch(() => {
        // Deletion may fail silently if the node doesn't exist or is protected
      });
    } else {
      onComponentDelete(lastInteractedStepId, lastInteractedResource as Element);
    }
    setIsOpenResourcePropertiesPanel(false);
  }, [
    deleteElements,
    lastInteractedResource,
    lastInteractedStepId,
    onComponentDelete,
    setIsOpenResourcePropertiesPanel,
  ]);

  return (
    <Box
      ref={setContainerEl}
      width="100%"
      height="100%"
      id="drawer-container"
      position="relative"
      component="div"
      {...rest}
    >
      {children}
      <Drawer
        open={open}
        anchor={anchor}
        onClose={handleClose}
        elevation={5}
        slotProps={{
          paper: {
            className: classNames('flow-builder-right-panel base', className),
            style: {position: 'absolute'},
          },
          backdrop: {
            style: {position: 'absolute'},
          },
        }}
        ModalProps={{
          container: containerEl,
          keepMounted: true,
          style: {pointerEvents: 'none'},
        }}
        sx={{
          pointerEvents: 'none',
          '& .MuiDrawer-paper': {
            pointerEvents: 'auto',
          },
        }}
        hideBackdrop
        className="flow-builder-right-panel"
        variant="temporary"
      >
        <Box
          display="flex"
          justifyContent="space-between"
          alignItems="center"
          className="flow-builder-right-panel header"
        >
          {resourcePropertiesPanelHeading}
          <IconButton onClick={handleClose}>
            <X height={16} width={16} />
          </IconButton>
        </Box>
        <div className="flow-builder-right-panel content full-height">
          <ResourceProperties />
        </div>
        {lastInteractedResource?.deletable !== false && (
          <Box display="flex" justifyContent="flex-end" alignItems="right" className="flow-builder-right-panel footer">
            <Button
              variant="outlined"
              onClick={handleDelete}
              color="error"
              startIcon={<TrashIcon size={16} />}
              fullWidth
            >
              Delete Element
            </Button>
          </Box>
        )}
      </Drawer>
    </Box>
  );
}

export default memo(ResourcePropertyPanel);
