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
import {Box, Button, IconButton, type DrawerProps} from '@wso2/oxygen-ui';
import {useReactFlow} from '@xyflow/react';
import {X, TrashIcon} from '@wso2/oxygen-ui-icons-react';
import BuilderFloatingPanel from '../../../../components/BuilderLayout/BuilderFloatingPanel';
import useFlowBuilderCore from '../../hooks/useFlowBuilderCore';
import ResourceProperties from './ResourceProperties';
import {ResourceTypes} from '../../models/resources';
import {type Element} from '../../models/elements';

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
      <BuilderFloatingPanel open={open ?? false} onClose={handleClose} container={containerEl} anchor={anchor}>
        <Box display="flex" justifyContent="space-between" alignItems="center" paddingBottom={2} flexShrink={0}>
          {resourcePropertiesPanelHeading}
          <IconButton onClick={handleClose}>
            <X height={16} width={16} />
          </IconButton>
        </Box>
        <Box
          sx={{
            flex: 1,
            minHeight: 0,
            overflowY: 'auto',
            overflowX: 'hidden',
            '&::-webkit-scrollbar': {
              width: '6px',
            },
            '&::-webkit-scrollbar-track': {
              background: 'transparent',
            },
            '&::-webkit-scrollbar-thumb': {
              background: 'rgba(0, 0, 0, 0.2)',
              borderRadius: '3px',
              '&:hover': {
                background: 'rgba(0, 0, 0, 0.3)',
              },
            },
          }}
        >
          <ResourceProperties />
        </Box>
        {lastInteractedResource?.deletable !== false && (
          <Box display="flex" justifyContent="flex-end" alignItems="right" flexShrink={0}>
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
      </BuilderFloatingPanel>
    </Box>
  );
}

export default memo(ResourcePropertyPanel);
